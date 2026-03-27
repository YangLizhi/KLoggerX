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

// GetOnlyOfficeConfig returns the editor configuration for a document
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

// OnlyOfficeCallback handles callbacks from OnlyOffice DocumentServer
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

// DownloadDocumentForOnlyOffice serves document content for OnlyOffice to download
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

// GetOnlyOfficeServerURL returns the OnlyOffice server URL for frontend
func GetOnlyOfficeServerURL(c *gin.Context) {
	c.JSON(http.StatusOK, model.Success(gin.H{
		"serverUrl": config.Cfg.OnlyOffice.ServerURL,
	}))
}
