package v1

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/pkg/utils"
	"kloggerx-server/internal/repository/mysql"
	"kloggerx-server/internal/service"

	"github.com/gin-gonic/gin"
)

// ExportKnowledgeBase 导出知识库为zip
// GET /knowledge/:id/export
func ExportKnowledgeBase(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	data, filename, err := service.ExportKnowledgeBaseAsZip(uid, uint(id))
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, "application/zip", data)
}

// ─── Conversation History ─────────────────────────────────────────────────────

// GetConversationListHandler godoc
// @Summary 获取会话列表
// @Description 获取用户的AI会话历史列表
// @Tags 知识库
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /knowledge/conversations [get]
func GetConversationListHandler(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	kbIDStr := c.Query("knowledgeBaseId")
	var kbID *uint
	if kbIDStr != "" {
		id, _ := strconv.ParseUint(kbIDStr, 10, 64)
		p := uint(id)
		kbID = &p
	}
	list, total, err := service.GetConversationList(uid, kbID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(model.PaginatedData{List: list, Total: total, Page: page, PageSize: pageSize}))
}

// CreateConversationHandler godoc
// @Summary 创建会话
// @Description 创建新的AI会话
// @Tags 知识库
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /knowledge/conversations [post]
func CreateConversationHandler(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	var body struct {
		KnowledgeBaseID *uint  `json:"knowledgeBaseId"`
		Model           string `json:"model" binding:"max=100"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}
	conv, err := service.CreateConversation(uid, body.KnowledgeBaseID, body.Model)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(conv))
}

// GetConversationDetailHandler godoc
// @Summary 获取会话详情
// @Description 获取会话详情及消息列表
// @Tags 知识库
// @Produce json
// @Param id path int true "会话ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /knowledge/conversations/{id} [get]
func GetConversationDetailHandler(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	msgPage, _ := strconv.Atoi(c.DefaultQuery("msgPage", "1"))
	msgPageSize, _ := strconv.Atoi(c.DefaultQuery("msgPageSize", "50"))
	conv, messages, msgTotal, err := service.GetConversationDetail(uint(id), uid, msgPage, msgPageSize)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(map[string]interface{}{
		"conversation": conv,
		"messages":     messages,
		"msgTotal":     msgTotal,
		"msgPage":      msgPage,
		"msgPageSize":  msgPageSize,
	}))
}

// UpdateConversationHandler godoc
// @Summary 更新会话
// @Description 更新会话标题或置顶状态
// @Tags 知识库
// @Accept json
// @Produce json
// @Param id path int true "会话ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /knowledge/conversations/{id} [put]
func UpdateConversationHandler(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		Title    string `json:"title" binding:"max=500"`
		IsPinned *bool  `json:"isPinned"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}
	updates := map[string]interface{}{}
	if body.Title != "" {
		updates["title"] = body.Title
	}
	if body.IsPinned != nil {
		updates["is_pinned"] = *body.IsPinned
	}
	if len(updates) == 0 {
		c.JSON(http.StatusOK, model.ErrorMsg("无更新内容"))
		return
	}
	if err := service.UpdateConversation(uint(id), uid, updates); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

// DeleteConversationHandler godoc
// @Summary 删除会话
// @Description 删除指定会话
// @Tags 知识库
// @Produce json
// @Param id path int true "会话ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /knowledge/conversations/{id} [delete]
func DeleteConversationHandler(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeleteConversation(uint(id), uid); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

// BatchDeleteConversationsHandler godoc
// @Summary 批量删除会话
// @Description 批量删除多个会话
// @Tags 知识库
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /knowledge/conversations [delete]
func BatchDeleteConversationsHandler(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	var body struct {
		IDs []uint `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}
	if err := service.BatchDeleteConversations(body.IDs, uid); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

// ExportConversationHandler GET /api/v1/knowledge/conversations/:id/export
func ExportConversationHandler(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	format := c.DefaultQuery("format", "markdown")

	content, filename, contentType, err := service.ExportConversation(uint(id), uid, format)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, contentType, content)
}

// CreateShareLinkHandler POST /api/v1/knowledge/conversations/:id/share
func CreateShareLinkHandler(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var body struct {
		ExpiresInHours *int `json:"expiresInHours"`
	}
	c.ShouldBindJSON(&body)

	link, err := service.CreateShareLink(uint(id), uid, body.ExpiresInHours)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(link))
}

// RevokeShareLinkHandler DELETE /api/v1/knowledge/conversations/:id/share
func RevokeShareLinkHandler(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	if err := service.RevokeShareLink(uint(id), uid); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

// GetSharedConversationHandler GET /api/v1/knowledge/share/:token (public, no auth)
func GetSharedConversationHandler(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusOK, model.ErrorMsg("无效的分享链接"))
		return
	}

	conv, messages, err := service.GetSharedConversation(token)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(map[string]interface{}{
		"conversation": conv,
		"messages":     messages,
	}))
}

// GetFollowUpSuggestionsHandler GET /api/v1/admin/settings/follow-up-suggestions
func GetFollowUpSuggestionsHandler(c *gin.Context) {
	enabled := service.IsFollowUpSuggestionsEnabled()
	c.JSON(http.StatusOK, model.Success(gin.H{"enabled": enabled}))
}

// UpdateFollowUpSuggestionsHandler PUT /api/v1/admin/settings/follow-up-suggestions
func UpdateFollowUpSuggestionsHandler(c *gin.Context) {
	var body struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}

	if err := service.UpdateFollowUpSuggestionsEnabled(body.Enabled); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("更新设置失败: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

// ─── Knowledge Base ───────────────────────────────────────────────────────────

// ─── Feedback ─────────────────────────────────────────────────────────────────

// SubmitFeedbackHandler POST /api/v1/knowledge/messages/:id/feedback
func SubmitFeedbackHandler(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	msgID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("无效的消息ID"))
		return
	}

	var body struct {
		Rating        int8   `json:"rating" binding:"required"`
		FeedbackType  string `json:"feedbackType" binding:"max=100"`
		Comment       string `json:"comment" binding:"max=2000"`
		CorrectAnswer string `json:"correctAnswer" binding:"max=5000"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}

	// 验证 rating 值
	if body.Rating != 1 && body.Rating != -1 {
		c.JSON(http.StatusOK, model.ErrorMsg("rating 必须为 1 或 -1"))
		return
	}

	// 踩时 feedbackType 必填
	if body.Rating == -1 && body.FeedbackType == "" {
		c.JSON(http.StatusOK, model.ErrorMsg("踩时必须提供 feedbackType"))
		return
	}

	if err := service.SubmitFeedback(uid, uint(msgID), body.Rating, body.FeedbackType, body.Comment, body.CorrectAnswer); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

// GetFeedbackHandler GET /api/v1/knowledge/messages/:id/feedback
func GetFeedbackHandler(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	msgID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("无效的消息ID"))
		return
	}

	feedback, err := service.GetMessageFeedback(uid, uint(msgID))
	if err != nil {
		// 无反馈记录返回 null
		c.JSON(http.StatusOK, model.Success(nil))
		return
	}
	c.JSON(http.StatusOK, model.Success(feedback))
}

// GetKnowledgeBaseList godoc
// @Summary 获取知识库列表
// @Description 分页获取知识库列表
// @Tags 知识库
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /knowledge/list [get]
func GetKnowledgeBaseList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	keyword := c.Query("keyword")
	list, total, err := service.GetKnowledgeBaseList(page, pageSize, keyword)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(model.PaginatedData{List: list, Total: total, Page: page, PageSize: pageSize}))
}

// GetKnowledgeBaseDetail godoc
// @Summary 获取知识库详情
// @Description 获取指定知识库的详细信息
// @Tags 知识库
// @Produce json
// @Param id path int true "知识库ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /knowledge/{id} [get]
func GetKnowledgeBaseDetail(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	// 权限检查：需要是知识库成员
	if err := service.CheckKnowledgePermission(uid, uint(id), "read"); err != nil {
		c.JSON(http.StatusForbidden, model.ErrorMsg("无权限访问该知识库"))
		return
	}

	kb, err := service.GetKnowledgeBaseDetail(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("知识库不存在"))
		return
	}
	c.JSON(http.StatusOK, model.Success(kb))
}

// CreateKnowledgeBase godoc
// @Summary 创建知识库
// @Description 创建新的知识库
// @Tags 知识库
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /knowledge/create [post]
func CreateKnowledgeBase(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	var body struct {
		Name        string `json:"name" binding:"required,min=1,max=100"`
		Description string `json:"description" binding:"max=1000"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}
	kb, err := service.CreateKnowledgeBase(uid, body.Name, body.Description)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(kb))
}

// UpdateKnowledgeBase godoc
// @Summary 更新知识库
// @Description 更新知识库信息
// @Tags 知识库
// @Accept json
// @Produce json
// @Param id path int true "知识库ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /knowledge/{id}/update [post]
func UpdateKnowledgeBase(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	// 权限检查：需要 owner 或 admin 成员
	if err := service.CheckKnowledgePermission(uid, uint(id), "edit"); err != nil {
		c.JSON(http.StatusForbidden, model.ErrorMsg("无权限修改该知识库"))
		return
	}

	var body struct {
		Name        string `json:"name" binding:"max=100"`
		Description string `json:"description" binding:"max=1000"`
	}
	c.ShouldBindJSON(&body)
	updates := map[string]interface{}{}
	if body.Name != "" {
		updates["name"] = body.Name
	}
	if body.Description != "" {
		updates["description"] = body.Description
	}
	if len(updates) > 0 {
		service.UpdateKnowledgeBaseInfo(uint(id), updates)
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

// DeleteKnowledgeBase godoc
// @Summary 删除知识库
// @Description 删除指定知识库
// @Tags 知识库
// @Produce json
// @Param id path int true "知识库ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /knowledge/{id} [delete]
func DeleteKnowledgeBase(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	// 权限检查：只有 owner 才能删除
	if err := service.CheckKnowledgePermission(uid, uint(id), "delete"); err != nil {
		c.JSON(http.StatusForbidden, model.ErrorMsg("无权限删除该知识库"))
		return
	}

	if err := service.DeleteKnowledgeBase(uint(id)); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

func GetKnowledgeBaseTree(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	docs, err := service.GetKnowledgeBaseTree(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(docs))
}

func AddKnowledgeMember(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		UserID uint   `json:"userId" binding:"required"`
		Role   string `json:"role"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}
	if body.Role == "" {
		body.Role = "member"
	}
	if err := service.AddKnowledgeMember(uint(id), body.UserID, body.Role); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

func RemoveKnowledgeMember(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		UserID uint `json:"userId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}
	if err := service.RemoveKnowledgeMember(uint(id), body.UserID); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

func GetKnowledgeMembers(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	members, err := service.GetKnowledgeMembers(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(members))
}

func SearchKnowledge(c *gin.Context) {
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	kbIDStr := c.Query("knowledgeBaseId")
	var kbID *uint
	if kbIDStr != "" {
		id, _ := strconv.ParseUint(kbIDStr, 10, 64)
		p := uint(id)
		kbID = &p
	}
	docs, total, err := service.SearchKnowledge(keyword, kbID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(model.PaginatedData{List: docs, Total: total, Page: page, PageSize: pageSize}))
}

func PublishDocument(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		DocumentID uint `json:"documentId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}
	if err := service.PublishDocument(uint(id), body.DocumentID); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	// Index the newly published document for RAG
	go service.IndexDocumentContent(uint(id), body.DocumentID)
	c.JSON(http.StatusOK, model.Success(nil))
}

// ─── Knowledge Sources ────────────────────────────────────────────────────────

func GetKnowledgeSources(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	sources, err := service.GetKnowledgeSources(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(sources))
}

func AddKnowledgeSource(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		SourceType string `json:"sourceType" binding:"required,oneof=folder document"` // "folder" | "document"
		SourceID   uint   `json:"sourceId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}
	src, err := service.AddKnowledgeSource(uint(id), body.SourceType, body.SourceID)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(src))
}

func RemoveKnowledgeSource(c *gin.Context) {
	kbID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	srcID, _ := strconv.ParseUint(c.Param("srcId"), 10, 64)
	if err := service.RemoveKnowledgeSource(uint(kbID), uint(srcID)); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

func SyncKnowledgeSource(c *gin.Context) {
	srcID, _ := strconv.ParseUint(c.Param("srcId"), 10, 64)
	if err := service.SyncKnowledgeSource(uint(srcID)); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

// ─── Knowledge AI Chat ────────────────────────────────────────────────────────

// KnowledgeChat godoc
// @Summary 知识库对话
// @Description 基于知识库进行AI对话
// @Tags 知识库
// @Accept json
// @Produce json
// @Param id path int true "知识库ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /knowledge/{id}/chat [post]
func KnowledgeChat(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		Question string                 `json:"question" binding:"required,max=10000"`
		History  []service.ChatMessage  `json:"history"`
		Model    string                 `json:"model" binding:"max=100"` // Optional: override model
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}
	result, err := service.ChatWithKnowledgeOpts(uint(id), body.Question, body.History, body.Model)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(result))
}

// KnowledgeChatGlobal handles chat without a specific KB (searches all accessible KBs).
func KnowledgeChatGlobal(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	var body struct {
		Question string                `json:"question" binding:"required,max=10000"`
		History  []service.ChatMessage `json:"history"`
		KBIDs    []uint                `json:"kbIds"`
		Model    string                `json:"model" binding:"max=100"` // Optional: override model
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}

	// If no specific KBs given, use all KBs the user is a member of
	if len(body.KBIDs) == 0 {
		var members []model.KnowledgeMember
		mysql.DB.Where("user_id = ?", uid).Find(&members)
		for _, m := range members {
			body.KBIDs = append(body.KBIDs, m.KnowledgeBaseID)
		}
	}

	if len(body.KBIDs) == 0 {
		c.JSON(http.StatusOK, model.ErrorMsg("您还没有加入任何知识库，请先创建或加入知识库"))
		return
	}

	// Check if any chunks exist in the selected KBs
	var totalChunks int64
	mysql.DB.Model(&model.KnowledgeChunk{}).Where("knowledge_base_id IN ?", body.KBIDs).Count(&totalChunks)
	if totalChunks == 0 {
		c.JSON(http.StatusOK, model.ErrorMsg("当前知识库内没有内容，请先添加文档并同步"))
		return
	}

	// Search chunks across all KBs
	var allChunks []service.ChunkRef
	for _, kbID := range body.KBIDs {
		chunks := service.SearchChunks(kbID, body.Question, 3)
		allChunks = append(allChunks, chunks...)
	}

	// Use first KB for full RAG with model override
	result, err := service.ChatWithKnowledgeOpts(body.KBIDs[0], body.Question, body.History, body.Model)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	result.Sources = allChunks
	c.JSON(http.StatusOK, model.Success(result))
}

// ─── Streaming Chat (SSE) ─────────────────────────────────────────────────────

// StreamChatHandler handles POST /api/v1/knowledge/chat/stream
// It uses Server-Sent Events to stream AI responses back to the client.
func StreamChatHandler(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))

	var body struct {
		Question        string                `json:"question" binding:"required,max=10000"`
		ConversationID  uint                  `json:"conversationId"`
		KnowledgeBaseID *uint                 `json:"knowledgeBaseId"`
		History         []service.ChatMessage `json:"history"`
		Model           string                `json:"model" binding:"max=100"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorMsg("参数错误"))
		return
	}

	// Set SSE response headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Flush()

	ctx := c.Request.Context()

	opts := service.ChatOptions{
		UserID:          uid,
		ConversationID:  body.ConversationID,
		KnowledgeBaseID: body.KnowledgeBaseID,
		Question:        body.Question,
		History:         body.History,
		ModelOverride:   body.Model,
	}

	writer := func(event service.SSEEvent) {
		// Check if client disconnected
		select {
		case <-ctx.Done():
			return
		default:
		}

		data, err := json.Marshal(event)
		if err != nil {
			return
		}
		fmt.Fprintf(c.Writer, "event: message\ndata: %s\n\n", string(data))
		c.Writer.(http.Flusher).Flush()
	}

	_ = service.StreamChatWithKnowledge(ctx, opts, writer)
}

// ─── RAPTOR Tree Management ────────────────────────────────────────────────

// ─── Knowledge Graph ──────────────────────────────────────────────────────────

// BuildKnowledgeGraph godoc
// @Summary 构建知识图谱
// @Description 触发构建知识图谱
// @Tags 知识库
// @Produce json
// @Param id path int true "知识库ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /knowledge/{id}/graph/build [post]
func BuildKnowledgeGraph(c *gin.Context) {
	_ = utils.GetUserID(c.MustGet("userId"))
	kbID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("无效的知识库ID"))
		return
	}

	if err := service.BuildKnowledgeGraph(uint(kbID)); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success("知识图谱构建已启动"))
}

// GetKnowledgeGraph godoc
// @Summary 获取知识图谱数据
// @Description 获取知识图谱数据
// @Tags 知识库
// @Produce json
// @Param id path int true "知识库ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /knowledge/{id}/graph [get]
func GetKnowledgeGraph(c *gin.Context) {
	kbID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("无效的知识库ID"))
		return
	}

	graph, err := service.GetKnowledgeGraphData(uint(kbID))
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(graph))
}

// GetKnowledgeGraphStatus godoc
// @Summary 获取知识图谱状态
// @Description 获取知识图谱构建状态
// @Tags 知识库
// @Produce json
// @Param id path int true "知识库ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /knowledge/{id}/graph/status [get]
func GetKnowledgeGraphStatus(c *gin.Context) {
	kbID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("无效的知识库ID"))
		return
	}

	status, err := service.GetKnowledgeGraphStatus(uint(kbID))
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(status))
}

// ─── RAPTOR Tree Management (continued) ────────────────────────────────────

// BuildRaptorTree handles building RAPTOR tree for a knowledge base
func BuildRaptorTree(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	kbID := uint(id)

	// Check if KB exists
	var kb model.KnowledgeBase
	if err := mysql.DB.First(&kb, kbID).Error; err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("知识库不存在"))
		return
	}

	// Check if there are embedded chunks
	var chunkCount int64
	mysql.DB.Model(&model.KnowledgeChunk{}).Where("knowledge_base_id = ? AND embedding_status = ?", kbID, "embedded").Count(&chunkCount)
	if chunkCount == 0 {
		c.JSON(http.StatusOK, model.ErrorMsg("请先同步数据来源并等待向量化完成"))
		return
	}

	// Check if RAPTOR service is initialized
	if service.GetRaptorService() == nil {
		c.JSON(http.StatusOK, model.ErrorMsg("RAPTOR服务未初始化，请检查AI模型配置"))
		return
	}

	// Build tree in background
	go func() {
		if err := service.BuildKBTree(kbID); err != nil {
			log.Printf("Failed to build RAPTOR tree for KB %d: %v", kbID, err)
		}
	}()

	c.JSON(http.StatusOK, model.Success(map[string]string{
		"message": "RAPTOR树构建已启动，请稍后查看结果",
	}))
}

// GetRaptorTreeStats returns statistics about the RAPTOR tree
func GetRaptorTreeStats(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	kbID := uint(id)

	raptorSvc := service.GetRaptorService()
	if raptorSvc == nil {
		c.JSON(http.StatusOK, model.ErrorMsg("RAPTOR服务未初始化"))
		return
	}

	stats, err := raptorSvc.GetRaptorTreeStats(kbID)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.Success(stats))
}

// GetEmbeddingStatus returns the embedding status for a knowledge base
func GetEmbeddingStatus(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	kbID := uint(id)

	var totalChunks, embeddedChunks, pendingChunks, failedChunks int64

	mysql.DB.Model(&model.KnowledgeChunk{}).Where("knowledge_base_id = ?", kbID).Count(&totalChunks)
	mysql.DB.Model(&model.KnowledgeChunk{}).Where("knowledge_base_id = ? AND embedding_status = ?", kbID, "embedded").Count(&embeddedChunks)
	mysql.DB.Model(&model.KnowledgeChunk{}).Where("knowledge_base_id = ? AND embedding_status = ?", kbID, "pending").Count(&pendingChunks)
	mysql.DB.Model(&model.KnowledgeChunk{}).Where("knowledge_base_id = ? AND embedding_status = ?", kbID, "failed").Count(&failedChunks)

	// Calculate progress percentage
	progress := 0.0
	if totalChunks > 0 {
		progress = float64(embeddedChunks) / float64(totalChunks) * 100
	}

	// Get recent embedding jobs
	var jobs []model.EmbeddingJob
	mysql.DB.Where("knowledge_base_id = ?", kbID).Order("created_at DESC").Limit(10).Find(&jobs)

	c.JSON(http.StatusOK, model.Success(map[string]interface{}{
		"totalChunks":     totalChunks,
		"embeddedChunks":  embeddedChunks,
		"pendingChunks":   pendingChunks,
		"failedChunks":    failedChunks,
		"progress":        progress,
		"recentJobs":      jobs,
	}))
}

// RebuildEmbeddings rebuilds embeddings for all chunks in a knowledge base
func RebuildEmbeddings(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	kbID := uint(id)

	// Check if there are documents in the KB
	var docCount int64
	mysql.DB.Model(&model.KnowledgeDocument{}).Where("knowledge_base_id = ?", kbID).Count(&docCount)
	if docCount == 0 {
		c.JSON(http.StatusOK, model.ErrorMsg("请先添加数据来源并同步"))
		return
	}

	// Check if embedding service is available
	if service.GetEmbeddingService() == nil {
		c.JSON(http.StatusOK, model.ErrorMsg("嵌入服务未初始化，请检查AI模型配置"))
		return
	}

	// Reset all chunks to pending
	mysql.DB.Model(&model.KnowledgeChunk{}).
		Where("knowledge_base_id = ?", kbID).
		Update("embedding_status", "pending")

	// Get all documents in the KB
	var kdList []model.KnowledgeDocument
	mysql.DB.Where("knowledge_base_id = ?", kbID).Find(&kdList)

	// Re-index all documents
	go func() {
		for _, kd := range kdList {
			service.IndexDocumentContent(kbID, kd.DocumentID)
		}
	}()

	c.JSON(http.StatusOK, model.Success(map[string]string{
		"message": "向量重建已启动",
	}))
}
