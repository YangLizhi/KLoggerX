package v1

import (
	"log"
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
		Model    string                 `json:"model"` // Optional: override model
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
		Question string                `json:"question" binding:"required"`
		History  []service.ChatMessage `json:"history"`
		KBIDs    []uint                `json:"kbIds"`
		Model    string                `json:"model"` // Optional: override model
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

// ─── RAPTOR Tree Management ────────────────────────────────────────────────

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
