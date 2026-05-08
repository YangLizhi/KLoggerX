package v1

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
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

// GetDocumentTree godoc
// @Summary 获取文档树
// @Description 获取用户的文档树结构
// @Tags 文档管理
// @Produce json
// @Param parentId query int false "父文档ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /document/tree [get]
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

// GetDocumentDetail godoc
// @Summary 获取文档详情
// @Description 获取指定文档的详细信息
// @Tags 文档管理
// @Produce json
// @Param id path int true "文档ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /document/{id} [get]
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

// CreateDocument godoc
// @Summary 创建文档
// @Description 创建一个新文档
// @Tags 文档管理
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /document/create [post]
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

// UpdateDocument godoc
// @Summary 更新文档标题
// @Description 更新指定文档的标题
// @Tags 文档管理
// @Accept json
// @Produce json
// @Param id path int true "文档ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /document/{id}/update [post]
func UpdateDocument(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	// 权限检查：需要编辑权限
	if err := service.CheckDocumentPermission(uid, uint(id), "edit"); err != nil {
		c.JSON(http.StatusForbidden, model.ErrorMsg("无权限执行此操作"))
		return
	}

	var body struct {
		Title string `json:"title" binding:"max=500"`
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

// SaveDocumentContent godoc
// @Summary 保存文档内容
// @Description 保存文档的内容
// @Tags 文档管理
// @Accept json
// @Produce json
// @Param id path int true "文档ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /document/{id}/content [post]
func SaveDocumentContent(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		Content string `json:"content" binding:"required,max=5000000"`
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

// DeleteDocument godoc
// @Summary 删除文档
// @Description 将文档移入回收站
// @Tags 文档管理
// @Produce json
// @Param id path int true "文档ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /document/{id}/delete [post]
func DeleteDocument(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	// 权限检查：需要删除权限（owner 或 admin）
	if err := service.CheckDocumentPermission(uid, uint(id), "delete"); err != nil {
		c.JSON(http.StatusForbidden, model.ErrorMsg("无权限执行此操作"))
		return
	}

	if err := service.DeleteDocument(uint(id)); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	service.CreateOperationLog(uid, "", "delete", "document", uint(id), "", "", c.ClientIP())
	c.JSON(http.StatusOK, model.Success(nil))
}

// RestoreDocument godoc
// @Summary 恢复文档
// @Description 从回收站恢复文档
// @Tags 文档管理
// @Produce json
// @Param id path int true "文档ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /document/{id}/restore [post]
func RestoreDocument(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.RestoreDocument(uint(id)); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

// PermanentDelete 永久删除文档
// DELETE /document/:id/permanent
func PermanentDelete(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.PermanentDeleteDocument(uid, uint(id)); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	service.CreateOperationLog(uid, "", "permanent_delete", "document", uint(id), "", "", c.ClientIP())
	c.JSON(http.StatusOK, model.Success(nil))
}

// BatchRestore 批量恢复文档
// POST /document/batch-restore  body: { "doc_ids": [1,2,3] }
func BatchRestore(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	var body struct {
		DocIDs []uint `json:"doc_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}
	if err := service.BatchRestoreDocuments(uid, body.DocIDs); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

// CleanupExpired 手动触发清理过期文档（管理员）
// POST /document/cleanup-expired
func CleanupExpired(c *gin.Context) {
	count, err := service.CleanupExpiredDocuments()
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("清理失败: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(map[string]interface{}{"deleted_count": count}))
}

// PermanentDeleteDocument legacy handler (kept for backward compatibility with DELETE /document/:id)
// PermanentDeleteDocument godoc
// @Summary 永久删除文档
// @Description 永久删除文档
// @Tags 文档管理
// @Produce json
// @Param id path int true "文档ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /document/{id} [delete]
func PermanentDeleteDocument(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.PermanentDeleteDocument(uid, uint(id)); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

// MoveDocument godoc
// @Summary 移动文档
// @Description 移动文档到目标目录
// @Tags 文档管理
// @Accept json
// @Produce json
// @Param id path int true "文档ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /document/{id}/move [post]
func MoveDocument(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	// 权限检查：对源文档需要编辑权限
	if err := service.CheckDocumentPermission(uid, uint(id), "edit"); err != nil {
		c.JSON(http.StatusForbidden, model.ErrorMsg("无权限执行此操作"))
		return
	}

	var body struct {
		TargetParentID *uint `json:"targetParentId"`
	}
	c.ShouldBindJSON(&body)

	// 如果目标文件夹非根目录，检查对目标文件夹的权限
	if body.TargetParentID != nil {
		if err := service.CheckDocumentPermission(uid, *body.TargetParentID, "edit"); err != nil {
			c.JSON(http.StatusForbidden, model.ErrorMsg("无权限移动到目标位置"))
			return
		}
	}

	if err := service.MoveDocument(uint(id), body.TargetParentID); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	service.CreateOperationLog(uid, "", "move", "document", uint(id), "", "", c.ClientIP())
	c.JSON(http.StatusOK, model.Success(nil))
}

// CopyDocument godoc
// @Summary 复制文档
// @Description 复制文档
// @Tags 文档管理
// @Produce json
// @Param id path int true "文档ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /document/{id}/copy [post]
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

// PinDocument godoc
// @Summary 置顶文档
// @Description 置顶或取消置顶文档
// @Tags 文档管理
// @Accept json
// @Produce json
// @Param id path int true "文档ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /document/{id}/pin [post]
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

// FavoriteDocument godoc
// @Summary 收藏文档
// @Description 收藏或取消收藏文档
// @Tags 文档管理
// @Accept json
// @Produce json
// @Param id path int true "文档ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /document/{id}/favorite [post]
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

// TransferOwnership godoc
// @Summary 转移所有权
// @Description 转移文档所有权给其他用户
// @Tags 文档管理
// @Accept json
// @Produce json
// @Param id path int true "文档ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /document/{id}/transfer [post]
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

// GetRecycleBin godoc
// @Summary 获取回收站
// @Description 获取回收站中的文档列表
// @Tags 文档管理
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /document/recycle-bin [get]
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

// GetFavorites godoc
// @Summary 获取收藏列表
// @Description 获取用户收藏的文档列表
// @Tags 文档管理
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /document/favorites [get]
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

// GetPinnedDocuments godoc
// @Summary 获取置顶文档
// @Description 获取用户置顶的文档列表
// @Tags 文档管理
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /document/pinned [get]
func GetPinnedDocuments(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	docs, err := service.GetPinnedDocuments(uid)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(docs))
}

// GetRecentDocuments godoc
// @Summary 获取最近文档
// @Description 获取用户最近访问的文档
// @Tags 文档管理
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /document/recent [get]
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

// SearchDocuments godoc
// @Summary 搜索文档
// @Description 根据关键词搜索文档
// @Tags 文档管理
// @Produce json
// @Param keyword query string false "搜索关键词"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /document/search [get]
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

// GetDocumentVersions godoc
// @Summary 获取文档版本历史
// @Description 获取文档的版本历史记录
// @Tags 文档管理
// @Produce json
// @Param id path int true "文档ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /document/{id}/versions [get]
func GetDocumentVersions(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	versions, err := service.GetDocumentVersions(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(versions))
}

// RollbackVersion godoc
// @Summary 回滚版本
// @Description 回滚文档到指定版本
// @Tags 文档管理
// @Accept json
// @Produce json
// @Param id path int true "文档ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /document/{id}/rollback [post]
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

// GetVersionDiff godoc
// @Summary 获取版本差异
// @Description 获取两个版本之间的差异
// @Tags 文档管理
// @Produce json
// @Param id path int true "文档ID"
// @Param v1 query string true "版本1"
// @Param v2 query string true "版本2"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /document/{id}/versions/diff [get]
func GetVersionDiff(c *gin.Context) {
	docID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	v1 := c.Query("v1")
	v2 := c.Query("v2")
	if v1 == "" || v2 == "" {
		c.JSON(http.StatusOK, model.ErrorMsg("请提供v1和v2版本号"))
		return
	}
	diff, err := service.GetDocumentVersionDiff(uint(docID), v1, v2)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(diff))
}

// ImportDocument godoc
// @Summary 导入文档
// @Description 上传文件导入为文档
// @Tags 文档管理
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "文件"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /document/import [post]
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
		Title:        title,
		Type:         docType,
		ParentID:     parentID,
		FileSize:     file.Size,
		FileExt:      ext,
		OriginalName: file.Filename,
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

// GetDocumentFilePreview godoc
// @Summary 获取文档文件预览URL
// @Description 获取文档关联文件的预览URL
// @Tags 文档管理
// @Produce json
// @Param id path int true "文档ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /document/{id}/file-preview [get]
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

// ExportDocument godoc
// @Summary 导出文档
// @Description 导出指定文档
// @Tags 文档管理
// @Produce json
// @Param id path int true "文档ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /document/{id}/export [get]
func ExportDocument(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	uid := utils.GetUserID(c.MustGet("userId"))
	format := c.DefaultQuery("format", "markdown")

	var content, filename string
	var err error

	switch format {
	case "text":
		content, filename, err = service.ExportDocumentAsText(uid, uint(id))
		if err != nil {
			c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
			return
		}
		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
		c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(content))
	case "markdown":
		content, filename, err = service.ExportDocumentAsMarkdown(uid, uint(id))
		if err != nil {
			c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
			return
		}
		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
		c.Data(http.StatusOK, "text/markdown; charset=utf-8", []byte(content))
	default:
		// Fallback: return JSON like before
		doc, err := service.GetDocumentDetail(uint(id), uid)
		if err != nil {
			c.JSON(http.StatusOK, model.ErrorMsg("文档不存在"))
			return
		}
		c.JSON(http.StatusOK, model.Success(gin.H{
			"title":   doc.Title,
			"content": doc.Content,
			"type":    doc.Type,
		}))
	}

	service.CreateOperationLog(uid, "", "export", "document", uint(id), "", format, c.ClientIP())
}

// SaveDocumentAsTemplate godoc
// @Summary 保存为模板
// @Description 将文档保存为模板
// @Tags 文档管理
// @Accept json
// @Produce json
// @Param id path int true "文档ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /document/{id}/save-as-template [post]
func SaveDocumentAsTemplate(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var req struct {
		Name        string `json:"name" binding:"required,max=200"`
		Description string `json:"description" binding:"max=1000"`
		Category    string `json:"category" binding:"max=100"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("请填写模板名称"))
		return
	}

	// Get document detail to verify access and get content
	doc, err := service.GetDocumentDetail(uint(id), uid)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("文档不存在或无权限"))
		return
	}

	// Create template from document
	t := &model.Template{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Type:        doc.Type,
		Content:     doc.Content,
		IsBuiltin:   false, // User-defined template
	}

	if err := service.CreateTemplate(t); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("保存模板失败: "+err.Error()))
		return
	}

	service.CreateOperationLog(uid, "", "save_as_template", "document", uint(id), doc.Title, "template:"+t.Name, c.ClientIP())
	c.JSON(http.StatusOK, model.Success(t))
}

// DownloadDocumentFile godoc
// @Summary 下载文档文件
// @Description 下载文档关联的原始文件
// @Tags 文档管理
// @Produce application/octet-stream
// @Param id path int true "文档ID"
// @Success 200 {file} binary
// @Security BearerAuth
// @Router /document/{id}/download [get]
func DownloadDocumentFile(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var doc model.Document
	if err := mysql.DB.First(&doc, id).Error; err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("文档不存在"))
		return
	}

	// Parse content to get the MinIO object path
	var contentData struct {
		FilePath string `json:"filePath"`
		FileName string `json:"fileName"`
	}
	if err := json.Unmarshal([]byte(doc.Content), &contentData); err != nil || contentData.FilePath == "" {
		c.JSON(http.StatusOK, model.ErrorMsg("该文档没有可下载的文件"))
		return
	}

	// Determine original filename
	filename := contentData.FileName
	if filename == "" {
		filename = doc.OriginalName
	}
	if filename == "" {
		filename = doc.Title
	}

	rc, err := service.GetFileStreamByPath(contentData.FilePath)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("获取文件失败: "+err.Error()))
		return
	}
	defer rc.Close()

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("Content-Type", "application/octet-stream")
	if _, err := io.Copy(c.Writer, rc); err != nil {
		// Connection may have closed, ignore
		_ = err
	}
}

// GetDocumentFileContent godoc
// @Summary 获取文档文件内容（流式）
// @Description 直接流式返回文档关联文件的内容，用于PDF等文件的内联预览。支持通过query参数传递token。
// @Tags 文档管理
// @Produce application/pdf
// @Param id path int true "文档ID"
// @Param token query string true "认证Token"
// @Success 200 {file} binary
// @Router /document/{id}/file-content [get]
func GetDocumentFileContent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var doc model.Document
	if err := mysql.DB.First(&doc, id).Error; err != nil {
		c.JSON(http.StatusNotFound, model.ErrorMsg("文档不存在"))
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
		c.JSON(http.StatusNotFound, model.ErrorMsg("该文档没有关联的文件"))
		return
	}

	// Determine content type based on file type
	contentType := "application/octet-stream"
	switch contentData.FileType {
	case "pdf":
		contentType = "application/pdf"
	case "word":
		contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case "excel":
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "ppt":
		contentType = "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	}

	rc, err := service.GetFileStreamByPath(contentData.FilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorMsg("获取文件失败: "+err.Error()))
		return
	}
	defer rc.Close()

	// Set headers for inline display (not download)
	fileName := contentData.FileName
	if fileName == "" {
		fileName = doc.Title
	}
	c.Header("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, fileName))
	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "private, max-age=3600")

	c.Status(http.StatusOK)
	if _, err := io.Copy(c.Writer, rc); err != nil {
		_ = err
	}
}

// SearchSuggestions godoc
// @Summary 搜索建议
// @Description 根据前缀返回文档标题建议
// @Tags 文档管理
// @Produce json
// @Param q query string true "搜索前缀"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /search/suggestions [get]
func SearchSuggestions(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusOK, model.Success(map[string]interface{}{"suggestions": []string{}}))
		return
	}

	uid := utils.GetUserID(c.MustGet("userId"))
	var titles []string
	mysql.DB.Model(&model.Document{}).
		Where("title LIKE ? AND is_deleted = ? AND user_id = ?", q+"%", false, uid).
		Limit(10).
		Pluck("title", &titles)

	c.JSON(http.StatusOK, model.Success(map[string]interface{}{"suggestions": titles}))
}

// AddDocumentShortcut godoc
// @Summary 创建快捷方式
// @Description 创建指向现有文档的快捷方式
// @Tags 文档管理
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /document/shortcut [post]
func AddDocumentShortcut(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	var req struct {
		DocumentID     uint  `json:"document_id" binding:"required"`
		TargetParentID *uint `json:"target_parent_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorMsg("参数错误"))
		return
	}
	if err := service.CreateDocumentShortcut(uid, req.DocumentID, req.TargetParentID); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessMsg("快捷方式已创建"))
}

// MigrateDocuments godoc
// @Summary 批量迁移文档
// @Description 批量移动文档到目标目录
// @Tags 文档管理
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /document/migrate [post]
func MigrateDocuments(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	var req struct {
		DocumentIDs    []uint `json:"document_ids" binding:"required"`
		TargetParentID *uint  `json:"target_parent_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorMsg("参数错误"))
		return
	}
	if len(req.DocumentIDs) == 0 {
		c.JSON(http.StatusBadRequest, model.ErrorMsg("请选择要迁移的文档"))
		return
	}
	// Check permission on target folder
	if req.TargetParentID != nil {
		if err := service.CheckDocumentPermission(uid, *req.TargetParentID, "edit"); err != nil {
			c.JSON(http.StatusForbidden, model.ErrorMsg("无权限移动到目标位置"))
			return
		}
	}
	// Check permission on each document
	for _, docID := range req.DocumentIDs {
		if err := service.CheckDocumentPermission(uid, docID, "edit"); err != nil {
			c.JSON(http.StatusForbidden, model.ErrorMsg(fmt.Sprintf("无权限操作文档 %d", docID)))
			return
		}
	}
	if err := service.MigrateDocuments(uid, req.DocumentIDs, req.TargetParentID); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.SuccessMsg("文档迁移成功"))
}

// EnhancedSearchDocuments godoc
// @Summary 增强搜索文档
// @Description 多关键词搜索文档，支持标题和内容搜索
// @Tags 文档管理
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /document/search [post]
func EnhancedSearchDocuments(c *gin.Context) {
	var req struct {
		Query    string `json:"query" binding:"max=500"`
		Type     string `json:"type" binding:"max=50"`
		Page     int    `json:"page"`
		PageSize int    `json:"pageSize"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}

	// Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	uid := utils.GetUserID(c.MustGet("userId"))

	// Build base query - only search documents user has access to
	query := mysql.DB.Model(&model.Document{}).Where("is_deleted = ? AND owner_id = ?", false, uid)

	// Apply keyword search if provided
	if req.Query != "" {
		// Split query into keywords by whitespace
		keywords := strings.Fields(req.Query)
		for _, kw := range keywords {
			if kw == "" {
				continue
			}
			like := "%" + kw + "%"
			query = query.Where("title LIKE ? OR content LIKE ?", like, like)
		}
	}

	// Apply type filter
	if req.Type != "" && req.Type != "all" {
		query = query.Where("type = ?", req.Type)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get paginated results
	var docs []model.Document
	offset := (req.Page - 1) * req.PageSize
	if err := query.Order("updated_at DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&docs).Error; err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("搜索失败"))
		return
	}

	// Format results with preview
	results := make([]map[string]interface{}, 0, len(docs))
	for _, doc := range docs {
		preview := doc.Content
		// Truncate preview to 200 characters
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}

		results = append(results, map[string]interface{}{
			"id":         doc.ID,
			"title":      doc.Title,
			"preview":    preview,
			"type":       doc.Type,
			"updatedAt":  doc.UpdatedAt,
			"ownerId":    doc.OwnerID,
		})
	}

	c.JSON(http.StatusOK, model.Success(map[string]interface{}{
		"items":    results,
		"total":    total,
		"page":     req.Page,
		"pageSize": req.PageSize,
	}))
}
