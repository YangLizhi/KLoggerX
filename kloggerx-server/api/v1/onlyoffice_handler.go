package v1

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"kloggerx-server/config"
	"kloggerx-server/internal/model"
	"kloggerx-server/internal/service"

	"github.com/gin-gonic/gin"
)

// OnlyOfficeConfigRequest represents the request for editor config
type OnlyOfficeConfigRequest struct {
	DocumentID uint   `json:"documentId" binding:"required"`
	Mode       string `json:"mode"` // "edit" or "view"
}

// OnlyOfficeCallbackRequest represents the callback from OnlyOffice
type OnlyOfficeCallbackRequest struct {
	Key        string          `json:"key"`
	Status     int             `json:"status"`
	URL        string          `json:"url"`
	ChangesURL string          `json:"changesurl"`
	History    json.RawMessage `json:"history"`
	Changes    json.RawMessage `json:"changes"`
	Users      []string        `json:"users"`
	Actions    json.RawMessage `json:"actions"`
	LastSave   string          `json:"lastsave"`
	NotModified bool           `json:"notmodified"`
	Token      string          `json:"token"` // JWT token from OnlyOffice
}

// GetOnlyOfficeConfig godoc
// @Summary 获取OnlyOffice编辑器配置
// @Description 获取文档的OnlyOffice编辑器配置
// @Tags 文档管理
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /onlyoffice/config [post]
func GetOnlyOfficeConfig(c *gin.Context) {
	var req OnlyOfficeConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.Error(400, "Invalid request: "+err.Error()))
		return
	}

	userID := c.GetUint("userId") // 使用正确的 key（与 middleware 一致）
	user, err := service.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("Failed to get user info"))
		return
	}

	mode := req.Mode
	if mode == "" {
		mode = "edit"
	}

	ooConfig, err := service.GetOnlyOfficeConfig(req.DocumentID, userID, user.Username, mode)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("Failed to generate editor config: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.Success(ooConfig))
}

// OnlyOfficeCallback godoc
// @Summary OnlyOffice回调
// @Description 处理OnlyOffice服务器的回调请求
// @Tags 文档管理
// @Accept json
// @Produce json
// @Param docId path int true "文档ID"
// @Success 200 {object} map[string]interface{}
// @Router /onlyoffice/callback/{docId} [post]
func OnlyOfficeCallback(c *gin.Context) {
	docIDStr := c.Param("docId")
	docID, err := strconv.ParseUint(docIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": 1, "message": "Invalid document ID"})
		return
	}

	// Read raw body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": 1, "message": "Failed to read body"})
		return
	}

	// Parse callback data
	var callbackReq OnlyOfficeCallbackRequest
	if err := json.Unmarshal(body, &callbackReq); err != nil {
		c.JSON(http.StatusOK, gin.H{"error": 1, "message": "Invalid JSON"})
		return
	}

	// Create callback object
	callback := &service.OnlyOfficeCallback{
		Key:         callbackReq.Key,
		Status:      callbackReq.Status,
		URL:         callbackReq.URL,
		ChangesURL:  callbackReq.ChangesURL,
		Users:       callbackReq.Users,
		LastSave:    callbackReq.LastSave,
		NotModified: callbackReq.NotModified,
	}

	// Handle callback
	if err := service.HandleOnlyOfficeCallback(uint(docID), callback); err != nil {
		c.JSON(http.StatusOK, gin.H{"error": 1, "message": err.Error()})
		return
	}

	// Return success (OnlyOffice expects {"error": 0} for success)
	c.JSON(http.StatusOK, gin.H{"error": 0})
}

// DownloadDocumentForOnlyOffice godoc
// @Summary OnlyOffice文档下载
// @Description 为OnlyOffice提供文档内容下载
// @Tags 文档管理
// @Produce application/octet-stream
// @Param docId path int true "文档ID"
// @Success 200 {file} binary
// @Router /onlyoffice/download/{docId} [get]
func DownloadDocumentForOnlyOffice(c *gin.Context) {
	docIDStr := c.Param("docId")
	docID, err := strconv.ParseUint(docIDStr, 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid document ID")
		return
	}

	// Verify signature
	signature := c.Query("sig")
	if !service.VerifyOnlyOfficeSignature(uint(docID), signature) {
		log.Printf("[OnlyOffice][download] invalid signature docID=%d sig=%q", docID, signature)
		c.String(http.StatusForbidden, "Invalid signature")
		return
	}
	log.Printf("[OnlyOffice][download] signature ok docID=%d", docID)

	// Get document content
	content, fileType, err := service.GetDocumentContentForOnlyOffice(uint(docID))
	if err != nil {
		log.Printf("[OnlyOffice][download] get content failed docID=%d err=%v", docID, err)
		c.String(http.StatusNotFound, "Document not found")
		return
	}
	log.Printf("[OnlyOffice][download] success docID=%d fileType=%s size=%d", docID, fileType, len(content))

	// Set headers
	filename := fmt.Sprintf("document.%s", fileType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", strconv.Itoa(len(content)))
	c.Header("Cache-Control", "no-cache")

	// Return content
	c.Data(http.StatusOK, "application/octet-stream", content)
}

// GetOnlyOfficeServerURL godoc
// @Summary 获取OnlyOffice服务器URL
// @Description 返回OnlyOffice服务器地址
// @Tags 文档管理
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /onlyoffice/server-url [get]
func GetOnlyOfficeServerURL(c *gin.Context) {
	c.JSON(http.StatusOK, model.Success(gin.H{
		"serverUrl": config.Cfg.OnlyOffice.ServerURL,
	}))
}
