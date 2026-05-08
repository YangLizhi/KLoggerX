package v1

import (
	"net/http"
	"strconv"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/pkg/utils"
	"kloggerx-server/internal/service"

	"github.com/gin-gonic/gin"
)

// CreateCommentHandler POST /document/:id/comments
func CreateCommentHandler(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	docID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if docID == 0 {
		c.JSON(http.StatusOK, model.ErrorMsg("文档ID无效"))
		return
	}

	var body struct {
		Content    string `json:"content" binding:"required,max=5000"`
		ParentID   *uint  `json:"parent_id"`
		QuotedText string `json:"quoted_text"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}

	comment, err := service.CreateComment(uid, uint(docID), body.Content, body.ParentID, body.QuotedText)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(comment))
}

// ListCommentsHandler GET /document/:id/comments
func ListCommentsHandler(c *gin.Context) {
	docID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if docID == 0 {
		c.JSON(http.StatusOK, model.ErrorMsg("文档ID无效"))
		return
	}

	comments, err := service.ListComments(uint(docID))
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(comments))
}

// DeleteCommentHandler DELETE /comment/:commentId
func DeleteCommentHandler(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	commentID, _ := strconv.ParseUint(c.Param("commentId"), 10, 64)
	if commentID == 0 {
		c.JSON(http.StatusOK, model.ErrorMsg("评论ID无效"))
		return
	}

	if err := service.DeleteCommentByUser(uid, uint(commentID)); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

// ResolveCommentHandler PUT /comment/:commentId/resolve
func ResolveCommentHandler(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	commentID, _ := strconv.ParseUint(c.Param("commentId"), 10, 64)
	if commentID == 0 {
		c.JSON(http.StatusOK, model.ErrorMsg("评论ID无效"))
		return
	}

	if err := service.ResolveCommentByUser(uid, uint(commentID)); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}
