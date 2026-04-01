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

func GetKnowledgeBaseDetail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	kb, err := service.GetKnowledgeBaseDetail(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("知识库不存在"))
		return
	}
	c.JSON(http.StatusOK, model.Success(kb))
}

func CreateKnowledgeBase(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	var body struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
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

func UpdateKnowledgeBase(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
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

func DeleteKnowledgeBase(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
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
		SourceType string `json:"sourceType" binding:"required"` // "folder" | "document"
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

func KnowledgeChat(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		Question string                 `json:"question" binding:"required"`
		History  []service.ChatMessage  `json:"history"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}
	result, err := service.ChatWithKnowledge(uint(id), body.Question, body.History)
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
		Question    string                `json:"question" binding:"required"`
		History     []service.ChatMessage `json:"history"`
		KBIDs       []uint                `json:"kbIds"`
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

	// Search chunks across all KBs
	var allChunks []service.ChunkRef
	for _, kbID := range body.KBIDs {
		chunks := service.SearchChunks(kbID, body.Question, 3)
		allChunks = append(allChunks, chunks...)
	}

	// If we have chunks, use first KB for full RAG; else just call AI without context
	if len(body.KBIDs) > 0 {
		result, err := service.ChatWithKnowledge(body.KBIDs[0], body.Question, body.History)
		if err != nil {
			c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
			return
		}
		result.Sources = allChunks
		c.JSON(http.StatusOK, model.Success(result))
		return
	}

	c.JSON(http.StatusOK, model.ErrorMsg("暂无知识库内容，请先向知识库添加文档"))
}
