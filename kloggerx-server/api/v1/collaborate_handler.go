package v1

import (
	"net/http"
	"strconv"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/pkg/utils"
	"kloggerx-server/internal/repository/mysql"
	"kloggerx-server/internal/service"

	"github.com/gin-gonic/gin"
)

// AddComment godoc
// @Summary 添加评论
// @Description 为文档添加评论
// @Tags 协作
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /collaborate/comment/add [post]
func AddComment(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	var body struct {
		DocumentID uint   `json:"documentId" binding:"required"`
		Content    string `json:"content" binding:"required,max=10000"`
		QuotedText string `json:"quoted_text" binding:"max=5000"`
		ParentID   *uint  `json:"parent_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}
	comment, err := service.AddComment(body.DocumentID, uid, body.Content, body.QuotedText, body.ParentID)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(comment))
}

// GetComments godoc
// @Summary 获取评论列表
// @Description 获取文档的评论列表
// @Tags 协作
// @Produce json
// @Param id path int true "文档ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /collaborate/comment/{id} [get]
func GetComments(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	comments, err := service.GetComments(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(comments))
}

// ResolveComment godoc
// @Summary 解决评论
// @Description 标记评论为已解决
// @Tags 协作
// @Produce json
// @Param id path int true "评论ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /collaborate/comment/{id}/resolve [post]
func ResolveComment(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	uid := utils.GetUserID(c.MustGet("userId"))
	role, _ := c.Get("role")

	// Check permission: comment author, document owner, or admin can resolve
	comment, err := service.GetCommentByID(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("评论不存在"))
		return
	}

	// Get document owner
	var doc model.Document
	if err := mysql.DB.Select("owner_id").First(&doc, comment.DocumentID).Error; err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("文档不存在"))
		return
	}

	// Check permission
	if comment.UserID != uid && doc.OwnerID != uid && role != "admin" {
		c.JSON(http.StatusOK, model.ErrorMsg("无权限标记此评论"))
		return
	}

	if err := service.ResolveComment(uint(id)); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

// DeleteComment godoc
// @Summary 删除评论
// @Description 删除指定评论
// @Tags 协作
// @Produce json
// @Param id path int true "评论ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /collaborate/comment/{id}/delete [post]
func DeleteComment(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	uid := utils.GetUserID(c.MustGet("userId"))
	role, _ := c.Get("role")

	// Check permission: only comment author or admin can delete
	comment, err := service.GetCommentByID(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("评论不存在"))
		return
	}

	if comment.UserID != uid && role != "admin" {
		c.JSON(http.StatusOK, model.ErrorMsg("无权限删除此评论"))
		return
	}

	if err := service.DeleteComment(uint(id)); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

// GetNotifications godoc
// @Summary 获取通知列表
// @Description 获取当前用户的通知列表
// @Tags 协作
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /collaborate/notifications [get]
func GetNotifications(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.GetNotifications(uid, page, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(model.PaginatedData{List: list, Total: total, Page: page, PageSize: pageSize}))
}

// MarkNotificationRead godoc
// @Summary 标记通知已读
// @Description 标记指定通知为已读
// @Tags 协作
// @Produce json
// @Param id path int true "通知ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /collaborate/notification/{id}/read [post]
func MarkNotificationRead(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.MarkNotificationRead(uint(id)); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

// MarkAllNotificationsRead godoc
// @Summary 全部标记已读
// @Description 标记所有通知为已读
// @Tags 协作
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /collaborate/notification/read-all [post]
func MarkAllNotificationsRead(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	if err := service.MarkAllNotificationsRead(uid); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

// CreateSystemNotification godoc
// @Summary 创建系统通知
// @Description 创建一条系统通知
// @Tags 协作
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /collaborate/notification/create [post]
func CreateSystemNotification(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	var body struct {
		Type    string `json:"type" binding:"required,max=50"`
		Title   string `json:"title" binding:"required,max=200"`
		Content string `json:"content" binding:"max=5000"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}
	if err := service.CreateNotification(body.Type, body.Title, body.Content, 0, uid, 0); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

// InviteCollaborators godoc
// @Summary 邀请协作者
// @Description 邀请用户协作编辑文档
// @Tags 协作
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /collaborate/invite [post]
func InviteCollaborators(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	var body struct {
		DocumentID uint   `json:"documentId" binding:"required"`
		UserIDs    []uint `json:"userIds" binding:"required"`
		Message    string `json:"message" binding:"max=1000"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}

	// Validate document exists and user has permission
	// Get document to check ownership/permission
	var doc model.Document
	if err := mysql.DB.Select("id", "owner_id").First(&doc, body.DocumentID).Error; err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("文档不存在"))
		return
	}

	// Check if user has permission to invite (owner or has edit permission)
	if doc.OwnerID != uid {
		// Check if user has edit permission
		var perm model.Permission
		err := mysql.DB.Where("document_id = ? AND user_id = ? AND level IN (?)",
			body.DocumentID, uid, []string{"owner", "write"}).First(&perm).Error
		if err != nil {
			c.JSON(http.StatusOK, model.ErrorMsg("没有权限邀请协作者"))
			return
		}
	}

	// Call service to create invitations
	if err := service.InviteCollaborators(body.DocumentID, uid, body.UserIDs, body.Message); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.Success(nil))
}
