package v1

import (
	"net/http"
	"strconv"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/pkg/utils"
	"kloggerx-server/internal/service"

	"github.com/gin-gonic/gin"
)

// 允许的文件类型白名单
var allowedMimeTypes = map[string]bool{
	"image/jpeg": true, "image/png": true, "image/gif": true, "image/webp": true,
	"application/pdf":  true,
	"application/msword": true, "application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/vnd.ms-excel": true, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
	"application/vnd.ms-powerpoint": true, "application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
	"text/plain": true, "text/csv": true, "text/markdown": true,
	"application/zip": true, "application/x-rar-compressed": true,
	"video/mp4": true, "audio/mpeg": true,
	"application/octet-stream": true, // 允许通用二进制流
}

// 最大文件大小: 100MB
const maxFileSize = 100 * 1024 * 1024

// UploadFile godoc
// @Summary 上传文件
// @Description 上传文件到存储
// @Tags 文件
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "文件"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /file/upload [post]
func UploadFile(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("请选择文件"))
		return
	}

	// 检查文件大小
	if file.Size > maxFileSize {
		c.JSON(http.StatusOK, model.ErrorMsg("文件大小超过限制(100MB)"))
		return
	}

	// 检测实际文件内容类型
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("无法读取文件"))
		return
	}
	defer src.Close()

	// 读取前512字节用于内容类型检测
	buf := make([]byte, 512)
	n, _ := src.Read(buf)
	detectedType := http.DetectContentType(buf[:n])

	// 检查Content-Type是否在白名单中
	contentType := file.Header.Get("Content-Type")
	if !allowedMimeTypes[contentType] && !allowedMimeTypes[detectedType] {
		c.JSON(http.StatusOK, model.ErrorMsg("不支持的文件类型: "+contentType))
		return
	}

	record, url, err := service.UploadFile(file, uid)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("上传失败: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(gin.H{
		"url":  url,
		"name": record.Name,
		"size": record.Size,
		"id":   record.ID,
	}))
}

// GetFilePreviewURL godoc
// @Summary 获取文件预览URL
// @Description 获取文件的预览访问URL
// @Tags 文件
// @Produce json
// @Param id path int true "文件ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /file/{id}/preview [get]
func GetFilePreviewURL(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	url, err := service.GetFilePreviewURL(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(gin.H{"url": url}))
}

// DeleteFile godoc
// @Summary 删除文件
// @Description 删除指定文件
// @Tags 文件
// @Produce json
// @Param id path int true "文件ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /file/{id}/delete [post]
func DeleteFile(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeleteFile(uint(id)); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}
