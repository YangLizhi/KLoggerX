package service

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/repository/mysql"
	rds "kloggerx-server/internal/repository/redis"
)

type CreateDocReq struct {
	Title        string `json:"title" binding:"required,max=500"`
	Type         string `json:"type" binding:"required,oneof=doc sheet slide mind code survey bitable file image"`
	ParentID     *uint  `json:"parentId"`
	FileSize     int64  `json:"fileSize"`
	FileExt      string `json:"fileExt" binding:"max=20"`
	OriginalName string `json:"originalName" binding:"max=500"`
}

func CreateDocument(userID uint, req CreateDocReq) (*model.Document, error) {
	doc := model.Document{
		Title:        req.Title,
		Type:         req.Type,
		ParentID:     req.ParentID,
		OwnerID:      userID,
		Version:      1,
		FileSize:     req.FileSize,
		FileExt:      req.FileExt,
		OriginalName: req.OriginalName,
	}
	if err := mysql.DB.Create(&doc).Error; err != nil {
		return nil, err
	}
	rds.InvalidateDocTreeCache(userID)
	return &doc, nil
}

func GetDocumentTree(userID uint, parentID *uint) ([]model.Document, error) {
	// Try cache for root-level tree only
	if parentID == nil {
		cacheKey := rds.DocTreeCacheKey(userID)
		var cached []model.Document
		if err := rds.GetCache(cacheKey, &cached); err == nil {
			return cached, nil
		}
	}

	var docs []model.Document
	db := mysql.DB.Table("documents").
		Select("documents.*, users.username as owner_name").
		Joins("LEFT JOIN users ON documents.owner_id = users.id").
		Where("documents.owner_id = ? AND documents.is_deleted = ?", userID, false)
	if parentID == nil {
		db = db.Where("documents.parent_id IS NULL")
	} else {
		db = db.Where("documents.parent_id = ?", *parentID)
	}
	err := db.Order("documents.is_pinned DESC, documents.updated_at DESC").Find(&docs).Error
	if err != nil {
		return nil, err
	}

	// Batch check favorites
	if len(docs) > 0 {
		var docIDs []uint
		for _, d := range docs {
			docIDs = append(docIDs, d.ID)
		}
		var favDocIDs []uint
		mysql.DB.Model(&model.Favorite{}).Where("user_id = ? AND document_id IN ?", userID, docIDs).
			Pluck("document_id", &favDocIDs)
		favSet := make(map[uint]bool, len(favDocIDs))
		for _, id := range favDocIDs {
			favSet[id] = true
		}
		for i := range docs {
			if favSet[docs[i].ID] {
				docs[i].IsFavorite = true
			}
		}
	}

	// Cache root-level tree for 5 minutes
	if parentID == nil {
		rds.SetCache(rds.DocTreeCacheKey(userID), docs, 5*time.Minute)
	}

	return docs, nil
}

func GetDocumentDetail(docID, userID uint) (*model.Document, error) {
	var doc model.Document
	if err := mysql.DB.First(&doc, docID).Error; err != nil {
		return nil, err
	}
	var owner model.User
	if mysql.DB.Select("username").First(&owner, doc.OwnerID).Error == nil {
		doc.OwnerName = owner.Username
	}
	var fav model.Favorite
	if mysql.DB.Where("user_id = ? AND document_id = ?", userID, docID).First(&fav).Error == nil {
		doc.IsFavorite = true
	}
	// record recent
	now := time.Now()
	mysql.DB.Where("user_id = ? AND document_id = ?", userID, docID).Delete(&model.RecentDocument{})
	mysql.DB.Create(&model.RecentDocument{UserID: userID, DocumentID: docID, AccessedAt: now})
	return &doc, nil
}

func SaveDocumentContent(docID, userID uint, content string) error {
	var doc model.Document
	if err := mysql.DB.First(&doc, docID).Error; err != nil {
		return err
	}
	// save version
	mysql.DB.Create(&model.DocumentVersion{
		DocumentID: docID,
		Version:    doc.Version,
		Content:    doc.Content,
		EditorID:   userID,
	})
	return mysql.DB.Model(&doc).Updates(map[string]interface{}{
		"content": content,
		"version": doc.Version + 1,
	}).Error
}

func UpdateDocumentTitle(docID uint, title string) error {
	return mysql.DB.Model(&model.Document{}).Where("id = ?", docID).Update("title", title).Error
}

func DeleteDocument(docID uint) error {
	now := time.Now()
	err := mysql.DB.Model(&model.Document{}).Where("id = ?", docID).Updates(map[string]interface{}{
		"is_deleted": true,
		"deleted_at": now,
	}).Error
	if err == nil {
		rds.InvalidateAllDocTreeCache()
	}
	return err
}

func RestoreDocument(docID uint) error {
	err := mysql.DB.Model(&model.Document{}).Where("id = ?", docID).Updates(map[string]interface{}{
		"is_deleted": false,
		"deleted_at": nil,
	}).Error
	if err == nil {
		rds.InvalidateAllDocTreeCache()
	}
	return err
}

// PermanentDeleteDocument 永久删除文档（物理删除）
// 验证文档已在回收站中（is_deleted = true），然后物理删除
func PermanentDeleteDocument(userID uint, docID uint) error {
	var doc model.Document
	if err := mysql.DB.Where("id = ? AND owner_id = ? AND is_deleted = ?", docID, userID, true).First(&doc).Error; err != nil {
		return errors.New("文档不存在或未在回收站中")
	}
	// Delete associated versions
	mysql.DB.Where("document_id = ?", docID).Delete(&model.DocumentVersion{})
	// Delete associated favorites
	mysql.DB.Where("document_id = ?", docID).Delete(&model.Favorite{})
	// Delete associated recent records
	mysql.DB.Where("document_id = ?", docID).Delete(&model.RecentDocument{})
	// Physical delete
	err := mysql.DB.Unscoped().Delete(&model.Document{}, docID).Error
	if err == nil {
		rds.InvalidateAllDocTreeCache()
	}
	return err
}

// BatchRestoreDocuments 批量恢复文档
// 将多个已删除文档恢复（清除 is_deleted 和 deleted_at）
func BatchRestoreDocuments(userID uint, docIDs []uint) error {
	if len(docIDs) == 0 {
		return errors.New("请选择要恢复的文档")
	}
	err := mysql.DB.Model(&model.Document{}).Where("id IN ? AND owner_id = ? AND is_deleted = ?", docIDs, userID, true).Updates(map[string]interface{}{
		"is_deleted": false,
		"deleted_at": nil,
	}).Error
	if err == nil {
		rds.InvalidateAllDocTreeCache()
	}
	return err
}

// CleanupExpiredDocuments 清理过期文档
// 删除 deleted_at 超过30天的文档（定时任务调用）
func CleanupExpiredDocuments() (int64, error) {
	expireTime := time.Now().AddDate(0, 0, -30)
	// Find expired documents
	var expiredDocs []model.Document
	if err := mysql.DB.Where("is_deleted = ? AND deleted_at < ?", true, expireTime).Find(&expiredDocs).Error; err != nil {
		return 0, err
	}
	if len(expiredDocs) == 0 {
		return 0, nil
	}
	var docIDs []uint
	for _, d := range expiredDocs {
		docIDs = append(docIDs, d.ID)
	}
	// Clean associated data
	mysql.DB.Where("document_id IN ?", docIDs).Delete(&model.DocumentVersion{})
	mysql.DB.Where("document_id IN ?", docIDs).Delete(&model.Favorite{})
	mysql.DB.Where("document_id IN ?", docIDs).Delete(&model.RecentDocument{})
	// Physical delete
	result := mysql.DB.Unscoped().Where("id IN ?", docIDs).Delete(&model.Document{})
	if result.Error == nil {
		rds.InvalidateAllDocTreeCache()
	}
	return result.RowsAffected, result.Error
}

func MoveDocument(docID uint, targetParentID *uint) error {
	err := mysql.DB.Model(&model.Document{}).Where("id = ?", docID).Update("parent_id", targetParentID).Error
	if err == nil {
		rds.InvalidateAllDocTreeCache()
	}
	return err
}

func CopyDocument(docID, userID uint) (*model.Document, error) {
	var src model.Document
	if err := mysql.DB.First(&src, docID).Error; err != nil {
		return nil, err
	}
	doc := model.Document{
		Title:        src.Title + " (副本)",
		Type:         src.Type,
		ParentID:     src.ParentID,
		OwnerID:      userID,
		Content:      src.Content,
		Version:      1,
		FileSize:     src.FileSize,
		FileExt:      src.FileExt,
		OriginalName: src.OriginalName,
	}
	if err := mysql.DB.Create(&doc).Error; err != nil {
		return nil, err
	}
	rds.InvalidateDocTreeCache(userID)
	return &doc, nil
}

func PinDocument(docID uint, isPinned bool) error {
	err := mysql.DB.Model(&model.Document{}).Where("id = ?", docID).Update("is_pinned", isPinned).Error
	if err == nil {
		rds.InvalidateAllDocTreeCache()
	}
	return err
}

func FavoriteDocument(userID, docID uint, isFavorite bool) error {
	if isFavorite {
		fav := model.Favorite{UserID: userID, DocumentID: docID}
		return mysql.DB.Where("user_id = ? AND document_id = ?", userID, docID).FirstOrCreate(&fav).Error
	}
	return mysql.DB.Where("user_id = ? AND document_id = ?", userID, docID).Delete(&model.Favorite{}).Error
}

func TransferOwnership(docID, targetUserID uint) error {
	return mysql.DB.Model(&model.Document{}).Where("id = ?", docID).Update("owner_id", targetUserID).Error
}

func GetRecycleBin(userID uint, page, pageSize int) ([]model.Document, int64, error) {
	var docs []model.Document
	var total int64
	db := mysql.DB.Model(&model.Document{}).Where("owner_id = ? AND is_deleted = ?", userID, true)
	db.Count(&total)
	err := db.Offset((page - 1) * pageSize).Limit(pageSize).Order("deleted_at DESC").Find(&docs).Error
	return docs, total, err
}

func GetFavorites(userID uint, page, pageSize int) ([]model.Document, int64, error) {
	var favs []model.Favorite
	var total int64
	db := mysql.DB.Model(&model.Favorite{}).Where("user_id = ?", userID)
	db.Count(&total)
	db.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&favs)
	var docs []model.Document
	for _, f := range favs {
		var d model.Document
		if mysql.DB.First(&d, f.DocumentID).Error == nil {
			d.IsFavorite = true
			docs = append(docs, d)
		}
	}
	return docs, total, nil
}

func GetPinnedDocuments(userID uint) ([]model.Document, error) {
	var docs []model.Document
	err := mysql.DB.Where("owner_id = ? AND is_pinned = ? AND is_deleted = ?", userID, true, false).Find(&docs).Error
	return docs, err
}

func GetRecentDocuments(userID uint, page, pageSize int) ([]model.Document, int64, error) {
	var recents []model.RecentDocument
	var total int64
	db := mysql.DB.Model(&model.RecentDocument{}).Where("user_id = ?", userID)
	db.Count(&total)
	db.Offset((page - 1) * pageSize).Limit(pageSize).Order("accessed_at DESC").Find(&recents)
	var docs []model.Document
	for _, r := range recents {
		var d model.Document
		if mysql.DB.First(&d, r.DocumentID).Error == nil && !d.IsDeleted {
			docs = append(docs, d)
		}
	}
	return docs, total, nil
}

func SearchDocuments(userID uint, keyword string, page, pageSize int) ([]model.Document, int64, error) {
	var docs []model.Document
	var total int64
	db := mysql.DB.Model(&model.Document{}).Where("owner_id = ? AND is_deleted = ?", userID, false).
		Where("title LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	db.Count(&total)
	err := db.Offset((page - 1) * pageSize).Limit(pageSize).Order("updated_at DESC").Find(&docs).Error
	return docs, total, err
}

func GetDocumentVersions(docID uint) ([]model.DocumentVersion, error) {
	var versions []model.DocumentVersion
	err := mysql.DB.Where("document_id = ?", docID).Order("version DESC").Find(&versions).Error
	for i := range versions {
		var u model.User
		if mysql.DB.Select("username").First(&u, versions[i].EditorID).Error == nil {
			versions[i].EditorName = u.Username
		}
	}
	return versions, err
}

func RollbackVersion(docID uint, version int) error {
	var v model.DocumentVersion
	if err := mysql.DB.Where("document_id = ? AND version = ?", docID, version).First(&v).Error; err != nil {
		return errors.New("版本不存在")
	}
	return mysql.DB.Model(&model.Document{}).Where("id = ?", docID).Updates(map[string]interface{}{
		"content": v.Content,
		"version": version,
	}).Error
}

// CreateDocumentShortcut creates a shortcut document pointing to an existing document
func CreateDocumentShortcut(userID uint, documentID uint, targetParentID *uint) error {
	// Verify source document exists
	var src model.Document
	if err := mysql.DB.First(&src, documentID).Error; err != nil {
		return errors.New("源文档不存在")
	}
	// Create shortcut document
	shortcut := model.Document{
		Title:            src.Title,
		Type:             src.Type,
		ParentID:         targetParentID,
		OwnerID:          userID,
		IsShortcut:       true,
		ShortcutTargetID: &documentID,
		Version:          1,
	}
	if err := mysql.DB.Create(&shortcut).Error; err != nil {
		return err
	}
	rds.InvalidateDocTreeCache(userID)
	return nil
}

// MigrateDocuments batch moves documents to a target directory
func MigrateDocuments(userID uint, documentIDs []uint, targetParentID *uint) error {
	for _, docID := range documentIDs {
		if err := mysql.DB.Model(&model.Document{}).Where("id = ?", docID).Update("parent_id", targetParentID).Error; err != nil {
			return err
		}
	}
	rds.InvalidateDocTreeCache(userID)
	return nil
}

// --- Export / Import ---

// ExportDocumentAsMarkdown 导出文档为Markdown格式
func ExportDocumentAsMarkdown(userID uint, docID uint) (string, string, error) {
	doc, err := GetDocumentDetail(docID, userID)
	if err != nil {
		return "", "", errors.New("文档不存在")
	}
	md := htmlToMarkdown(doc.Content)
	filename := doc.Title + ".md"
	return md, filename, nil
}

// ExportDocumentAsText 导出为纯文本
func ExportDocumentAsText(userID uint, docID uint) (string, string, error) {
	doc, err := GetDocumentDetail(docID, userID)
	if err != nil {
		return "", "", errors.New("文档不存在")
	}
	text := stripHTMLTags(doc.Content)
	filename := doc.Title + ".txt"
	return text, filename, nil
}

// ImportDocumentFromMarkdown 从Markdown导入
func ImportDocumentFromMarkdown(userID uint, title string, markdownContent string, folderID uint) (*model.Document, error) {
	htmlContent := markdownToHTML(markdownContent)
	var parentID *uint
	if folderID > 0 {
		parentID = &folderID
	}
	doc := model.Document{
		Title:    title,
		Type:     "doc",
		ParentID: parentID,
		OwnerID:  userID,
		Content:  htmlContent,
		Version:  1,
	}
	if err := mysql.DB.Create(&doc).Error; err != nil {
		return nil, err
	}
	rds.InvalidateDocTreeCache(userID)
	return &doc, nil
}

// ExportKnowledgeBaseAsZip 知识库批量导出为zip
func ExportKnowledgeBaseAsZip(userID uint, kbID uint) ([]byte, string, error) {
	// Get knowledge base info
	kb, err := GetKnowledgeBaseDetail(kbID)
	if err != nil {
		return nil, "", errors.New("知识库不存在")
	}
	// Get all documents in the knowledge base
	docs, err := GetKnowledgeBaseTree(kbID)
	if err != nil {
		return nil, "", err
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, doc := range docs {
		md := htmlToMarkdown(doc.Content)
		filename := doc.Title + ".md"
		w, err := zw.Create(filename)
		if err != nil {
			continue
		}
		w.Write([]byte(md))
	}
	zw.Close()
	zipFilename := kb.Name + ".zip"
	return buf.Bytes(), zipFilename, nil
}

// htmlToMarkdown converts HTML content to Markdown using regex
func htmlToMarkdown(html string) string {
	if html == "" {
		return ""
	}
	s := html
	// Pre/code blocks (must be before inline code)
	rePreCode := regexp.MustCompile(`(?s)<pre><code[^>]*>(.*?)</code></pre>`)
	s = rePreCode.ReplaceAllString(s, "\n```\n$1\n```\n")
	// Headings
	for i := 6; i >= 1; i-- {
		re := regexp.MustCompile(fmt.Sprintf(`<h%d[^>]*>(.*?)</h%d>`, i, i))
		prefix := strings.Repeat("#", i) + " "
		s = re.ReplaceAllString(s, "\n"+prefix+"$1\n")
	}
	// Bold
	reBold := regexp.MustCompile(`<(?:strong|b)>(.*?)</(?:strong|b)>`)
	s = reBold.ReplaceAllString(s, "**$1**")
	// Italic
	reItalic := regexp.MustCompile(`<(?:em|i)>(.*?)</(?:em|i)>`)
	s = reItalic.ReplaceAllString(s, "*$1*")
	// Links
	reLink := regexp.MustCompile(`<a[^>]*href="([^"]*?)"[^>]*>(.*?)</a>`)
	s = reLink.ReplaceAllString(s, "[$2]($1)")
	// Inline code
	reCode := regexp.MustCompile(`<code>(.*?)</code>`)
	s = reCode.ReplaceAllString(s, "`$1`")
	// List items
	reLi := regexp.MustCompile(`<li[^>]*>(.*?)</li>`)
	s = reLi.ReplaceAllString(s, "- $1\n")
	// Paragraphs
	reP := regexp.MustCompile(`<p[^>]*>(.*?)</p>`)
	s = reP.ReplaceAllString(s, "$1\n\n")
	// Line breaks
	reBr := regexp.MustCompile(`<br\s*/?>`) 
	s = reBr.ReplaceAllString(s, "\n")
	// Strip remaining tags
	s = stripHTMLTags(s)
	// Clean up multiple newlines
	reMultiNL := regexp.MustCompile(`\n{3,}`)
	s = reMultiNL.ReplaceAllString(s, "\n\n")
	return strings.TrimSpace(s)
}

// markdownToHTML converts Markdown content to HTML using regex
func markdownToHTML(md string) string {
	if md == "" {
		return ""
	}
	lines := strings.Split(md, "\n")
	var result []string
	inCodeBlock := false
	var codeBlockContent []string

	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			if inCodeBlock {
				result = append(result, "<pre><code>"+strings.Join(codeBlockContent, "\n")+"</code></pre>")
				codeBlockContent = nil
				inCodeBlock = false
			} else {
				inCodeBlock = true
			}
			continue
		}
		if inCodeBlock {
			codeBlockContent = append(codeBlockContent, line)
			continue
		}
		// Headings
		if strings.HasPrefix(line, "###### ") {
			result = append(result, "<h6>"+strings.TrimPrefix(line, "###### ")+"</h6>")
		} else if strings.HasPrefix(line, "##### ") {
			result = append(result, "<h5>"+strings.TrimPrefix(line, "##### ")+"</h5>")
		} else if strings.HasPrefix(line, "#### ") {
			result = append(result, "<h4>"+strings.TrimPrefix(line, "#### ")+"</h4>")
		} else if strings.HasPrefix(line, "### ") {
			result = append(result, "<h3>"+strings.TrimPrefix(line, "### ")+"</h3>")
		} else if strings.HasPrefix(line, "## ") {
			result = append(result, "<h2>"+strings.TrimPrefix(line, "## ")+"</h2>")
		} else if strings.HasPrefix(line, "# ") {
			result = append(result, "<h1>"+strings.TrimPrefix(line, "# ")+"</h1>")
		} else if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			item := strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* ")
			result = append(result, "<li>"+item+"</li>")
		} else if line == "" {
			continue
		} else {
			result = append(result, "<p>"+line+"</p>")
		}
	}
	// Handle unclosed code block
	if inCodeBlock && len(codeBlockContent) > 0 {
		result = append(result, "<pre><code>"+strings.Join(codeBlockContent, "\n")+"</code></pre>")
	}

	html := strings.Join(result, "\n")
	// Inline formatting
	reBold := regexp.MustCompile(`\*\*(.+?)\*\*`)
	html = reBold.ReplaceAllString(html, "<strong>$1</strong>")
	reItalic := regexp.MustCompile(`\*(.+?)\*`)
	html = reItalic.ReplaceAllString(html, "<em>$1</em>")
	reCode := regexp.MustCompile("`([^`]+)`")
	html = reCode.ReplaceAllString(html, "<code>$1</code>")
	reLink := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	html = reLink.ReplaceAllString(html, `<a href="$2">$1</a>`)
	return html
}

// stripHTMLTags removes all HTML tags from a string
func stripHTMLTags(s string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(s, "")
}

// --- Version Diff ---

type DiffLine struct {
	Type    string `json:"type"`              // "add" | "delete" | "equal"
	Content string `json:"content"`
	OldLine int    `json:"old_line,omitempty"`
	NewLine int    `json:"new_line,omitempty"`
}

type VersionDiff struct {
	OldVersion string     `json:"old_version"`
	NewVersion string     `json:"new_version"`
	Lines      []DiffLine `json:"lines"`
	Stats      DiffStats  `json:"stats"`
}

type DiffStats struct {
	Added   int `json:"added"`
	Deleted int `json:"deleted"`
	Changed int `json:"changed"`
}

func GetDocumentVersionDiff(docID uint, v1, v2 string) (*VersionDiff, error) {
	v1Int, err := strconv.Atoi(v1)
	if err != nil {
		return nil, errors.New("版本号v1无效")
	}
	v2Int, err := strconv.Atoi(v2)
	if err != nil {
		return nil, errors.New("版本号v2无效")
	}

	// Get old version content
	var oldContent string
	var ver1 model.DocumentVersion
	if err := mysql.DB.Where("document_id = ? AND version = ?", docID, v1Int).First(&ver1).Error; err != nil {
		// Maybe it's the current version
		var doc model.Document
		if err2 := mysql.DB.First(&doc, docID).Error; err2 != nil {
			return nil, errors.New("文档不存在")
		}
		if doc.Version == v1Int {
			oldContent = doc.Content
		} else {
			return nil, errors.New("版本v1不存在")
		}
	} else {
		oldContent = ver1.Content
	}

	// Get new version content
	var newContent string
	var ver2 model.DocumentVersion
	if err := mysql.DB.Where("document_id = ? AND version = ?", docID, v2Int).First(&ver2).Error; err != nil {
		var doc model.Document
		if err2 := mysql.DB.First(&doc, docID).Error; err2 != nil {
			return nil, errors.New("文档不存在")
		}
		if doc.Version == v2Int {
			newContent = doc.Content
		} else {
			return nil, errors.New("版本v2不存在")
		}
	} else {
		newContent = ver2.Content
	}

	lines := computeDiff(oldContent, newContent)
	stats := DiffStats{}
	for _, l := range lines {
		switch l.Type {
		case "add":
			stats.Added++
		case "delete":
			stats.Deleted++
		}
	}
	stats.Changed = stats.Added + stats.Deleted

	return &VersionDiff{
		OldVersion: v1,
		NewVersion: v2,
		Lines:      lines,
		Stats:      stats,
	}, nil
}

func computeDiff(oldText, newText string) []DiffLine {
	oldLines := strings.Split(oldText, "\n")
	newLines := strings.Split(newText, "\n")

	// LCS DP
	m, n := len(oldLines), len(newLines)
	// Limit size to avoid excessive memory usage
	if m > 5000 || n > 5000 {
		// Fallback: show all old as deleted, all new as added
		var result []DiffLine
		for i, l := range oldLines {
			result = append(result, DiffLine{Type: "delete", Content: l, OldLine: i + 1})
		}
		for i, l := range newLines {
			result = append(result, DiffLine{Type: "add", Content: l, NewLine: i + 1})
		}
		return result
	}

	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if oldLines[i-1] == newLines[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] >= dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}

	// Backtrack to build diff
	var result []DiffLine
	i, j := m, n
	for i > 0 || j > 0 {
		if i > 0 && j > 0 && oldLines[i-1] == newLines[j-1] {
			result = append(result, DiffLine{Type: "equal", Content: oldLines[i-1], OldLine: i, NewLine: j})
			i--
			j--
		} else if j > 0 && (i == 0 || dp[i][j-1] >= dp[i-1][j]) {
			result = append(result, DiffLine{Type: "add", Content: newLines[j-1], NewLine: j})
			j--
		} else {
			result = append(result, DiffLine{Type: "delete", Content: oldLines[i-1], OldLine: i})
			i--
		}
	}

	// Reverse result
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}
