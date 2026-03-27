package v1

import (
	"net/http"
	"strconv"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/pkg/utils"
	"kloggerx-server/internal/service"

	"github.com/gin-gonic/gin"
)

func GetDocumentPermissions(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	perms, err := service.GetDocumentPermissions(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(perms))
}

func SetPermission(c *gin.Context) {
	var body struct {
		DocumentID uint   `json:"documentId" binding:"required"`
		UserID     uint   `json:"userId" binding:"required"`
		Level      string `json:"level" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}
	if err := service.SetPermission(body.DocumentID, body.UserID, body.Level); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

func RemovePermission(c *gin.Context) {
	var body struct {
		DocumentID uint `json:"documentId" binding:"required"`
		UserID     uint `json:"userId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}
	if err := service.RemovePermission(body.DocumentID, body.UserID); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

func GetShareSetting(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	s, err := service.GetShareSetting(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(s))
}

func UpdateShareSetting(c *gin.Context) {
	var s model.ShareSetting
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}
	if err := service.UpdateShareSetting(&s); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

func CheckPermission(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	level := service.CheckPermission(uint(id), uid)
	c.JSON(http.StatusOK, model.Success(gin.H{"level": level}))
}
