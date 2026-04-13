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

func AddComment(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	var body struct {
		DocumentID uint   `json:"documentId" binding:"required"`
		Content    string `json:"content" binding:"required"`
		Selection  string `json:"selection"`
		ParentID   *uint  `json:"parentId"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}
	comment, err := service.AddComment(body.DocumentID, uid, body.Content, body.Selection, body.ParentID)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(comment))
}

func GetComments(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	comments, err := service.GetComments(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(comments))
}

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

func MarkNotificationRead(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.MarkNotificationRead(uint(id)); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

func MarkAllNotificationsRead(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	if err := service.MarkAllNotificationsRead(uid); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

func CreateSystemNotification(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	var body struct {
		Type    string `json:"type" binding:"required"`
		Title   string `json:"title" binding:"required"`
		Content string `json:"content"`
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

func InviteCollaborators(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	var body struct {
		DocumentID uint   `json:"documentId" binding:"required"`
		UserIDs    []uint `json:"userIds" binding:"required"`
		Message    string `json:"message"`
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
