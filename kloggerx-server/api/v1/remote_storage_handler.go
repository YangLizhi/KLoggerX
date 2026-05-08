package v1

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/service"

	"github.com/gin-gonic/gin"
)

// ListRemoteStorageFiles godoc
// @Summary 列出远程存储文件
// @Description 列出远程存储目录中的文件
// @Tags 远程存储
// @Produce json
// @Param id path int true "存储ID"
// @Param path query string false "目录路径"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /remote-storage/{id}/files [get]
func ListRemoteStorageFiles(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("无效的存储ID"))
		return
	}

	// Get path parameter, default to root
	path := c.Query("path")
	if path == "" {
		path = "/"
	}

	log.Printf("[Handler] ListRemoteStorageFiles: id=%d, path=%q", id, path)

	// List files
	rs := service.GetRemoteStorageService()
	files, err := rs.ListFiles(uint(id), path)
	if err != nil {
		log.Printf("[Handler] ListRemoteStorageFiles error: %v", err)
		c.JSON(http.StatusOK, model.ErrorMsg("列出文件失败: "+err.Error()))
		return
	}

	log.Printf("[Handler] ListRemoteStorageFiles success: %d files", len(files))
	c.JSON(http.StatusOK, model.Success(map[string]interface{}{
		"path":  path,
		"files": files,
	}))
}

// DownloadRemoteFile godoc
// @Summary 下载远程文件
// @Description 从远程存储下载文件
// @Tags 远程存储
// @Produce application/octet-stream
// @Param id path int true "存储ID"
// @Param path query string true "文件路径"
// @Success 200 {file} binary
// @Security BearerAuth
// @Router /remote-storage/{id}/download [get]
func DownloadRemoteFile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("无效的存储ID"))
		return
	}

	// Get file path
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusOK, model.ErrorMsg("文件路径不能为空"))
		return
	}

	// Download file
	rs := service.GetRemoteStorageService()
	reader, err := rs.DownloadFile(uint(id), path)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("下载文件失败: "+err.Error()))
		return
	}
	defer reader.Close()

	// Get filename from path
	filename := filepath.Base(path)
	if filename == "" || filename == "." || filename == "/" {
		filename = "download"
	}

	// Set response headers for download
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Transfer-Encoding", "binary")

	// Stream the file content
	_, err = io.Copy(c.Writer, reader)
	if err != nil {
		// Can't send error response as headers are already sent
		return
	}
}

// UploadRemoteFile godoc
// @Summary 上传文件到远程存储
// @Description 上传文件到远程存储
// @Tags 远程存储
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "存储ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /remote-storage/{id}/upload [post]
func UploadRemoteFile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("无效的存储ID"))
		return
	}

	// Get destination path
	path := c.PostForm("path")
	if path == "" {
		path = "/"
	}

	// Get uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("请选择要上传的文件"))
		return
	}
	defer file.Close()

	// Build full path
	destPath := path
	if !strings.HasSuffix(destPath, "/") {
		destPath += "/"
	}
	destPath += header.Filename

	// Upload file
	rs := service.GetRemoteStorageService()
	err = rs.UploadFile(uint(id), destPath, file, header.Size)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("上传文件失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.Success(map[string]interface{}{
		"path": destPath,
		"size": header.Size,
	}))
}

// CreateRemoteFolder godoc
// @Summary 创建远程目录
// @Description 在远程存储中创建目录
// @Tags 远程存储
// @Accept json
// @Produce json
// @Param id path int true "存储ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /remote-storage/{id}/mkdir [post]
func CreateRemoteFolder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("无效的存储ID"))
		return
	}

	var req struct {
		Path string `json:"path" binding:"required,max=1000"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}

	// Create directory
	rs := service.GetRemoteStorageService()
	err = rs.CreateDir(uint(id), req.Path)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("创建目录失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.Success(nil))
}

// DeleteRemoteFile godoc
// @Summary 删除远程文件
// @Description 删除远程存储中的文件或目录
// @Tags 远程存储
// @Accept json
// @Produce json
// @Param id path int true "存储ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /remote-storage/{id}/file [delete]
func DeleteRemoteFile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("无效的存储ID"))
		return
	}

	var req struct {
		Path string `json:"path" binding:"required,max=1000"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}

	// Delete file/directory
	rs := service.GetRemoteStorageService()
	err = rs.DeleteFile(uint(id), req.Path)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("删除失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.Success(nil))
}

// TestRemoteStorageConnect godoc
// @Summary 测试远程存储连接
// @Description 测试与远程存储的连接
// @Tags 远程存储
// @Produce json
// @Param id path int true "存储ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /remote-storage/{id}/test [post]
func TestRemoteStorageConnect(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("无效的存储ID"))
		return
	}

	// Test connection
	rs := service.GetRemoteStorageService()
	err = rs.TestConnection(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("连接失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.Success(map[string]interface{}{
		"connected": true,
		"message":   "连接成功",
	}))
}

// DisconnectRemoteStorageHandler godoc
// @Summary 断开远程存储连接
// @Description 断开与远程存储的连接
// @Tags 远程存储
// @Produce json
// @Param id path int true "存储ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /remote-storage/{id}/disconnect [post]
func DisconnectRemoteStorageHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("无效的存储ID"))
		return
	}

	// Disconnect
	rs := service.GetRemoteStorageService()
	err = rs.Disconnect(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("断开连接失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.Success(nil))
}
