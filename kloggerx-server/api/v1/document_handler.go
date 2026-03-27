package v1

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/pkg/utils"
	miniosvc "kloggerx-server/internal/pkg/minio"
	"kloggerx-server/internal/repository/mysql"
	"kloggerx-server/internal/service"

	"github.com/gin-gonic/gin"
)

func GetDocumentTree(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	pidStr := c.Query("parentId")
	var parentID *uint
	if pidStr != "" {
		pid, _ := strconv.ParseUint(pidStr, 10, 64)
		p := uint(pid)
		parentID = &p
	}
	docs, err := service.GetDocumentTree(uid, parentID)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(docs))
}

func GetDocumentDetail(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	doc, err := service.GetDocumentDetail(uint(id), uid)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("文档不存在"))
		return
	}
	c.JSON(http.StatusOK, model.Success(doc))
}

func CreateDocument(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	var req service.CreateDocReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}
	doc, err := service.CreateDocument(uid, req)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	service.CreateOperationLog(uid, "", "create", "document", doc.ID, doc.Title, "", c.ClientIP())
	c.JSON(http.StatusOK, model.Success(doc))
}

func UpdateDocument(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		Title string `json:"title"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}
	if err := service.UpdateDocumentTitle(uint(id), body.Title); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

func SaveDocumentContent(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}
	if err := service.SaveDocumentContent(uint(id), uid, body.Content); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

func DeleteDocument(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeleteDocument(uint(id)); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	service.CreateOperationLog(uid, "", "delete", "document", uint(id), "", "", c.ClientIP())
	c.JSON(http.StatusOK, model.Success(nil))
}

func RestoreDocument(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.RestoreDocument(uint(id)); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

func PermanentDeleteDocument(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.PermanentDeleteDocument(uint(id)); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

func MoveDocument(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		TargetParentID *uint `json:"targetParentId"`
	}
	c.ShouldBindJSON(&body)
	if err := service.MoveDocument(uint(id), body.TargetParentID); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	service.CreateOperationLog(uid, "", "move", "document", uint(id), "", "", c.ClientIP())
	c.JSON(http.StatusOK, model.Success(nil))
}

func CopyDocument(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	doc, err := service.CopyDocument(uint(id), uid)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(doc))
}

func PinDocument(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		IsPinned bool `json:"isPinned"`
	}
	c.ShouldBindJSON(&body)
	if err := service.PinDocument(uint(id), body.IsPinned); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

func FavoriteDocument(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		IsFavorite bool `json:"isFavorite"`
	}
	c.ShouldBindJSON(&body)
	if err := service.FavoriteDocument(uid, uint(id), body.IsFavorite); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

func TransferOwnership(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		TargetUserID uint `json:"targetUserId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}
	if err := service.TransferOwnership(uint(id), body.TargetUserID); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	service.CreateOperationLog(uid, "", "transfer", "document", uint(id), "", "", c.ClientIP())
	c.JSON(http.StatusOK, model.Success(nil))
}

func GetRecycleBin(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	docs, total, err := service.GetRecycleBin(uid, page, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(model.PaginatedData{List: docs, Total: total, Page: page, PageSize: pageSize}))
}

func GetFavorites(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	docs, total, err := service.GetFavorites(uid, page, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(model.PaginatedData{List: docs, Total: total, Page: page, PageSize: pageSize}))
}

func GetPinnedDocuments(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	docs, err := service.GetPinnedDocuments(uid)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(docs))
}

func GetRecentDocuments(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	docs, total, err := service.GetRecentDocuments(uid, page, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(model.PaginatedData{List: docs, Total: total, Page: page, PageSize: pageSize}))
}

func SearchDocuments(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	docs, total, err := service.SearchDocuments(uid, keyword, page, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(model.PaginatedData{List: docs, Total: total, Page: page, PageSize: pageSize}))
}

func GetDocumentVersions(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	versions, err := service.GetDocumentVersions(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(versions))
}

func RollbackVersion(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		Version int `json:"version" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}
	if err := service.RollbackVersion(uint(id), body.Version); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

func ImportDocument(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("请上传文件"))
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("无法读取文件"))
		return
	}
	defer f.Close()

	buf := make([]byte, file.Size)
	f.Read(buf)
	content := string(buf)

	title := file.Filename
	ext := ""
	for i := len(title) - 1; i >= 0; i-- {
		if title[i] == '.' {
			ext = strings.ToLower(title[i:])
			title = title[:i]
			break
		}
	}

	parentIDStr := c.Query("parentId")
	var parentID *uint
	if parentIDStr != "" {
		pid, _ := strconv.ParseUint(parentIDStr, 10, 64)
		p := uint(pid)
		parentID = &p
	}

	// Determine document type based on file extension
	docType := "doc"
	switch ext {
	case ".xlsx", ".xls":
		docType = "file"
		filePath := uploadImportFileToStorage(file, ext)
		content = fmt.Sprintf(`{"type":"file","fileType":"excel","fileName":"%s","filePath":"%s"}`, file.Filename, filePath)
	case ".pptx", ".ppt":
		docType = "file"
		filePath := uploadImportFileToStorage(file, ext)
		content = fmt.Sprintf(`{"type":"file","fileType":"ppt","fileName":"%s","filePath":"%s"}`, file.Filename, filePath)
	case ".md", ".markdown":
		// Convert markdown to document format
		content = convertMarkdownToDoc(content)
	case ".json":
		// Keep JSON as is
	case ".txt", ".html":
		// Convert text to document format
		content = convertTextToDoc(content)
	case ".pdf":
		docType = "file"
		filePath := uploadImportFileToStorage(file, ext)
		content = fmt.Sprintf(`{"type":"file","fileType":"pdf","fileName":"%s","filePath":"%s"}`, file.Filename, filePath)
	case ".docx", ".doc":
		docType = "file"
		filePath := uploadImportFileToStorage(file, ext)
		content = fmt.Sprintf(`{"type":"file","fileType":"word","fileName":"%s","filePath":"%s"}`, file.Filename, filePath)
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg", ".bmp":
		docType = "image"
		// Store image as base64 data URL
		mimeType := "image/png"
		switch ext {
		case ".jpg", ".jpeg":
			mimeType = "image/jpeg"
		case ".gif":
			mimeType = "image/gif"
		case ".webp":
			mimeType = "image/webp"
		case ".svg":
			mimeType = "image/svg+xml"
		case ".bmp":
			mimeType = "image/bmp"
		}
		base64Data := base64.StdEncoding.EncodeToString(buf)
		content = fmt.Sprintf(`{"type":"image","url":"data:%s;base64,%s","name":"%s"}`, mimeType, base64Data, file.Filename)
	default:
		content = convertTextToDoc(content)
	}

	doc, err := service.CreateDocument(uid, service.CreateDocReq{
		Title:    title,
		Type:     docType,
		ParentID: parentID,
	})
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}

	service.SaveDocumentContent(doc.ID, uid, content)
	service.CreateOperationLog(uid, "", "import", "document", doc.ID, title, "", c.ClientIP())

	c.JSON(http.StatusOK, model.Success(doc))
}

func convertTextToDoc(text string) string {
	// Escape special characters for JSON
	escaped := strings.ReplaceAll(text, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
	escaped = strings.ReplaceAll(escaped, "\n", "\\n")
	escaped = strings.ReplaceAll(escaped, "\r", "\\r")
	escaped = strings.ReplaceAll(escaped, "\t", "\\t")
	return fmt.Sprintf(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"%s"}]}]}`, escaped)
}

func convertMarkdownToDoc(md string) string {
	lines := strings.Split(md, "\n")
	var content []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Escape special characters
		escaped := strings.ReplaceAll(line, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
		if strings.HasPrefix(escaped, "# ") {
			content = append(content, fmt.Sprintf(`{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"%s"}]}`, strings.TrimPrefix(escaped, "# ")))
		} else if strings.HasPrefix(escaped, "## ") {
			content = append(content, fmt.Sprintf(`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"%s"}]}`, strings.TrimPrefix(escaped, "## ")))
		} else if strings.HasPrefix(escaped, "### ") {
			content = append(content, fmt.Sprintf(`{"type":"heading","attrs":{"level":3},"content":[{"type":"text","text":"%s"}]}`, strings.TrimPrefix(escaped, "### ")))
		} else {
			content = append(content, fmt.Sprintf(`{"type":"paragraph","content":[{"type":"text","text":"%s"}]}`, escaped))
		}
	}
	if len(content) == 0 {
		content = append(content, `{"type":"paragraph","content":[{"type":"text","text":""}]}`)
	}
	return fmt.Sprintf(`{"type":"doc","content":[%s]}`, strings.Join(content, ","))
}

// uploadImportFileToStorage uploads the original file to MinIO and returns the object path
func uploadImportFileToStorage(file *multipart.FileHeader, ext string) string {
	src, err := file.Open()
	if err != nil {
		return ""
	}
	defer src.Close()

	objectName := fmt.Sprintf("office/%s/%s%s", time.Now().Format("2006/01/02"), utils.GenerateRandomString(16), ext)
	contentType := "application/octet-stream"
	switch ext {
	case ".pdf":
		contentType = "application/pdf"
	case ".docx":
		contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".doc":
		contentType = "application/msword"
	case ".xlsx":
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".xls":
		contentType = "application/vnd.ms-excel"
	case ".pptx":
		contentType = "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	case ".ppt":
		contentType = "application/vnd.ms-powerpoint"
	}

	if err := miniosvc.Upload(objectName, src, file.Size, contentType); err != nil {
		return ""
	}
	return objectName
}

// GetDocumentFilePreview returns a presigned URL for the original uploaded file
func GetDocumentFilePreview(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var doc model.Document
	if err := mysql.DB.First(&doc, id).Error; err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("文档不存在"))
		return
	}

	// Parse content to get filePath
	var contentData struct {
		Type     string `json:"type"`
		FileType string `json:"fileType"`
		FileName string `json:"fileName"`
		FilePath string `json:"filePath"`
	}
	if err := json.Unmarshal([]byte(doc.Content), &contentData); err != nil || contentData.FilePath == "" {
		c.JSON(http.StatusOK, model.ErrorMsg("该文档没有关联的文件"))
		return
	}

	var fileURL string
	if miniosvc.IsAvailable() {
		// MinIO available - get presigned URL
		url, err := miniosvc.GetPresignedURL(contentData.FilePath)
		if err != nil {
			c.JSON(http.StatusOK, model.ErrorMsg("获取文件预览URL失败"))
			return
		}
		fileURL = url
	} else {
		// Local filesystem - build full URL
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		host := c.Request.Host
		fileURL = fmt.Sprintf("%s://%s/uploads/%s", scheme, host, contentData.FilePath)
	}

	c.JSON(http.StatusOK, model.Success(gin.H{
		"url":      fileURL,
		"fileType": contentData.FileType,
		"fileName": contentData.FileName,
	}))
}

func ExportDocument(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	uid := utils.GetUserID(c.MustGet("userId"))
	doc, err := service.GetDocumentDetail(uint(id), uid)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("文档不存在"))
		return
	}

	service.CreateOperationLog(uid, "", "export", "document", uint(id), doc.Title, "", c.ClientIP())

	c.JSON(http.StatusOK, model.Success(gin.H{
		"title":   doc.Title,
		"content": doc.Content,
		"type":    doc.Type,
	}))
}
