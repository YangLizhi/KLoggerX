package v1

import (
	"net/http"
	"strconv"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/pkg/utils"
	"kloggerx-server/internal/service"

	"github.com/gin-gonic/gin"
)

// GetDocumentPermissions godoc
// @Summary 获取文档权限列表
// @Description 获取指定文档的权限设置
// @Tags 认证
// @Produce json
// @Param id path int true "文档ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /auth/document/{id} [get]
func GetDocumentPermissions(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	perms, err := service.GetDocumentPermissions(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(perms))
}

// SetPermission godoc
// @Summary 设置权限
// @Description 设置用户对文档的权限级别
// @Tags 认证
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /auth/permission/set [post]
func SetPermission(c *gin.Context) {
	var body struct {
		DocumentID uint   `json:"documentId" binding:"required"`
		UserID     uint   `json:"userId" binding:"required"`
		Level      string `json:"level" binding:"required,oneof=read write owner"`
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

// RemovePermission godoc
// @Summary 移除权限
// @Description 移除用户对文档的权限
// @Tags 认证
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /auth/permission/remove [post]
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

// GetShareSetting godoc
// @Summary 获取分享设置
// @Description 获取文档的分享设置
// @Tags 认证
// @Produce json
// @Param id path int true "文档ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /auth/share/{id} [get]
func GetShareSetting(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	s, err := service.GetShareSetting(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(s))
}

// UpdateShareSetting godoc
// @Summary 更新分享设置
// @Description 更新文档的分享设置
// @Tags 认证
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /auth/share/update [post]
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

// CheckPermission godoc
// @Summary 检查权限
// @Description 检查当前用户对文档的权限级别
// @Tags 认证
// @Produce json
// @Param id path int true "文档ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /auth/check/{id} [get]
func CheckPermission(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	level := service.CheckPermission(uint(id), uid)
	c.JSON(http.StatusOK, model.Success(gin.H{"level": level}))
}
