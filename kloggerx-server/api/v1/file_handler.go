package v1

import (
	"net/http"
	"strconv"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/pkg/utils"
	"kloggerx-server/internal/service"

	"github.com/gin-gonic/gin"
)

func UploadFile(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("请选择文件"))
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

func GetFilePreviewURL(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	url, err := service.GetFilePreviewURL(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(gin.H{"url": url}))
}

func DeleteFile(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeleteFile(uint(id)); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}
