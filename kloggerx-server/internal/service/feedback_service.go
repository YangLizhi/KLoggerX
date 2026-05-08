package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/repository/mysql"
	"kloggerx-server/internal/repository/qdrant"

	"gorm.io/gorm/clause"
)

// SubmitFeedback 提交反馈（赞/踩）
// - 如果已有反馈则更新，没有则创建（基于 message_id + user_id 唯一索引）
// - rating: 1=赞, -1=踩
// - feedbackType: inaccurate/incomplete/irrelevant/outdated (踩时必填)
// - comment: 用户补充说明(可选)
// - correctAnswer: 用户提供的正确答案(可选)
func SubmitFeedback(userID, messageID uint, rating int8, feedbackType, comment, correctAnswer string) error {
	// 获取 message 以拿到 conversation_id
	var msg model.KbMessage
	if err := mysql.DB.First(&msg, messageID).Error; err != nil {
		return err
	}

	feedback := model.KbFeedback{
		MessageID:      messageID,
		ConversationID: msg.ConversationID,
		UserID:         userID,
		Rating:         rating,
		FeedbackType:   feedbackType,
		CreatedAt:      time.Now(),
	}
	if comment != "" {
		feedback.Comment = &comment
	}
	if correctAnswer != "" {
		feedback.CorrectAnswer = &correctAnswer
		feedback.ReviewStatus = "pending"
	}

	// Upsert: ON DUPLICATE KEY UPDATE
	err := mysql.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "message_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"rating", "feedback_type", "comment", "correct_answer", "review_status"}),
	}).Create(&feedback).Error
	if err != nil {
		return err
	}

	// 异步应用反馈到检索权重
	go ApplyFeedbackToRetrieval(messageID, rating, feedbackType)

	return nil
}

// UpdateFeedback 更新反馈
func UpdateFeedback(userID, messageID uint, updates map[string]interface{}) error {
	return mysql.DB.Model(&model.KbFeedback{}).
		Where("message_id = ? AND user_id = ?", messageID, userID).
		Updates(updates).Error
}

// GetMessageFeedback 获取消息的反馈（用于前端展示状态）
func GetMessageFeedback(userID, messageID uint) (*model.KbFeedback, error) {
	var feedback model.KbFeedback
	err := mysql.DB.Where("message_id = ? AND user_id = ?", messageID, userID).First(&feedback).Error
	if err != nil {
		return nil, err
	}
	return &feedback, nil
}

// ApplyFeedbackToRetrieval 将反馈应用到检索权重（异步）
// - 点赞：提升关联chunks的quality_score (+0.1, 上限2.0)
// - 点踩(irrelevant)：降低关联chunks的quality_score (-0.1, 下限0.5)
func ApplyFeedbackToRetrieval(messageID uint, rating int8, feedbackType string) {
	// 只对赞和踩(irrelevant)做权重调整
	if rating != 1 && (rating != -1 || feedbackType != "irrelevant") {
		return
	}

	// 从 kb_messages 获取 sources
	var msg model.KbMessage
	if err := mysql.DB.First(&msg, messageID).Error; err != nil {
		log.Printf("[Feedback] Failed to get message %d: %v", messageID, err)
		return
	}

	if msg.Sources == nil || *msg.Sources == "" {
		return
	}

	// 解析 sources JSON: [{docId, title, chunkContent, ...}]
	var sources []struct {
		DocID   uint   `json:"docId"`
		Title   string `json:"title"`
		ChunkID uint   `json:"chunkId"`
	}
	if err := json.Unmarshal([]byte(*msg.Sources), &sources); err != nil {
		log.Printf("[Feedback] Failed to parse sources for message %d: %v", messageID, err)
		return
	}

	if len(sources) == 0 {
		return
	}

	// 获取 VectorClient
	vectorClient := qdrant.DefaultVectorClient
	if vectorClient == nil {
		log.Printf("[Feedback] Vector client not initialized, skipping quality_score update")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 收集关联的 chunk 的 qdrant point IDs
	var pointIDs []string
	for _, src := range sources {
		if src.ChunkID == 0 && src.DocID == 0 {
			continue
		}

		// 通过 chunk_id 查找对应的 qdrant_point_id
		if src.ChunkID > 0 {
			var chunk model.KnowledgeChunk
			if err := mysql.DB.First(&chunk, src.ChunkID).Error; err == nil && chunk.QdrantPointID != "" {
				pointIDs = append(pointIDs, chunk.QdrantPointID)
			}
		} else if src.DocID > 0 {
			// 如果没有 chunkID，按 docID 查找所有 chunks
			var chunks []model.KnowledgeChunk
			mysql.DB.Where("document_id = ? AND qdrant_point_id != ''", src.DocID).Find(&chunks)
			for _, c := range chunks {
				pointIDs = append(pointIDs, c.QdrantPointID)
			}
		}
	}

	if len(pointIDs) == 0 {
		return
	}

	// 计算新的 quality_score
	// 由于 Qdrant 不支持原子增减操作，我们需要使用固定值更新
	// 降级方案：直接设置一个调整后的值
	// 为了更精确，我们从数据库的反馈统计来决定 quality_score
	qualityScore := calculateQualityScore(pointIDs, rating)

	// 更新 Qdrant payload
	payload := map[string]interface{}{
		"quality_score": qualityScore,
	}
	if err := vectorClient.UpdatePayloadByPointIDs(ctx, payload, pointIDs); err != nil {
		log.Printf("[Feedback] Failed to update quality_score in Qdrant: %v", err)
	} else {
		log.Printf("[Feedback] Updated quality_score to %.2f for %d points (message %d, rating %d)",
			qualityScore, len(pointIDs), messageID, rating)
	}
}

// calculateQualityScore 基于反馈计算 quality_score
func calculateQualityScore(pointIDs []string, rating int8) float64 {
	// 简化逻辑：赞 -> 1.1, 踩 -> 0.9
	// 实际场景中可以查询历史反馈次数做累计计算
	if rating == 1 {
		return 1.1
	}
	return 0.9
}

// ─── Review Workflow ─────────────────────────────────────────────────────────

// FeedbackReviewItem 审核列表项
type FeedbackReviewItem struct {
	model.KbFeedback
	OriginalQuestion string `json:"originalQuestion"` // 原始用户问题
	OriginalAnswer   string `json:"originalAnswer"`   // 原始AI回答
	KnowledgeBaseID  *uint  `json:"knowledgeBaseId"`  // 所属知识库
	UserName         string `json:"userName"`          // 提交者用户名
}

// GetPendingReviews 获取待审核列表（管理员用）
func GetPendingReviews(page, pageSize int) ([]FeedbackReviewItem, int64, error) {
	var total int64
	mysql.DB.Model(&model.KbFeedback{}).
		Where("correct_answer IS NOT NULL AND correct_answer != '' AND review_status = 'pending'").
		Count(&total)

	var feedbacks []model.KbFeedback
	err := mysql.DB.Where("correct_answer IS NOT NULL AND correct_answer != '' AND review_status = 'pending'").
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&feedbacks).Error
	if err != nil {
		return nil, 0, err
	}

	var items []FeedbackReviewItem
	for _, fb := range feedbacks {
		item := FeedbackReviewItem{KbFeedback: fb}

		// 获取原始消息内容（AI回答）
		var msg model.KbMessage
		if err := mysql.DB.First(&msg, fb.MessageID).Error; err == nil {
			item.OriginalAnswer = msg.Content
		}

		// 获取该消息之前的用户问题
		var userMsg model.KbMessage
		if err := mysql.DB.Where("conversation_id = ? AND role = 'user' AND id < ?", fb.ConversationID, fb.MessageID).
			Order("id DESC").First(&userMsg).Error; err == nil {
			item.OriginalQuestion = userMsg.Content
		}

		// 获取所属知识库
		var conv model.KbConversation
		if err := mysql.DB.First(&conv, fb.ConversationID).Error; err == nil {
			item.KnowledgeBaseID = conv.KnowledgeBaseID
		}

		// 获取提交者用户名
		var user model.User
		if err := mysql.DB.First(&user, fb.UserID).Error; err == nil {
			item.UserName = user.Username
		}

		items = append(items, item)
	}

	return items, total, nil
}

// ApproveReview 通过审核
func ApproveReview(feedbackID, reviewerID uint, comment string) error {
	var feedback model.KbFeedback
	if err := mysql.DB.First(&feedback, feedbackID).Error; err != nil {
		return fmt.Errorf("反馈记录不存在")
	}
	if feedback.ReviewStatus != "pending" {
		return fmt.Errorf("该记录已被审核")
	}
	if feedback.CorrectAnswer == nil || *feedback.CorrectAnswer == "" {
		return fmt.Errorf("没有可审核的正确答案")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"review_status":  "approved",
		"reviewer_id":    reviewerID,
		"review_comment": comment,
		"reviewed_at":    &now,
	}
	if err := mysql.DB.Model(&model.KbFeedback{}).Where("id = ?", feedbackID).Updates(updates).Error; err != nil {
		return err
	}

	// 异步将 correct_answer 存入知识库
	go addCorrectAnswerToKnowledge(feedback)

	return nil
}

// RejectReview 拒绝审核
func RejectReview(feedbackID, reviewerID uint, comment string) error {
	var feedback model.KbFeedback
	if err := mysql.DB.First(&feedback, feedbackID).Error; err != nil {
		return fmt.Errorf("反馈记录不存在")
	}
	if feedback.ReviewStatus != "pending" {
		return fmt.Errorf("该记录已被审核")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"review_status":  "rejected",
		"reviewer_id":    reviewerID,
		"review_comment": comment,
		"reviewed_at":    &now,
	}
	return mysql.DB.Model(&model.KbFeedback{}).Where("id = ?", feedbackID).Updates(updates).Error
}

// GetReviewStats 获取审核统计
func GetReviewStats() (map[string]int64, error) {
	stats := map[string]int64{}

	base := mysql.DB.Model(&model.KbFeedback{}).Where("correct_answer IS NOT NULL AND correct_answer != ''")

	var pending, approved, rejected int64
	base.Where("review_status = 'pending'").Count(&pending)
	base.Where("review_status = 'approved'").Count(&approved)
	base.Where("review_status = 'rejected'").Count(&rejected)

	stats["pending"] = pending
	stats["approved"] = approved
	stats["rejected"] = rejected

	return stats, nil
}

// addCorrectAnswerToKnowledge 将审核通过的正确答案作为新chunk存入知识库
func addCorrectAnswerToKnowledge(feedback model.KbFeedback) {
	// 获取对话关联的知识库
	var conv model.KbConversation
	if err := mysql.DB.First(&conv, feedback.ConversationID).Error; err != nil {
		log.Printf("[Review] Failed to get conversation %d: %v", feedback.ConversationID, err)
		return
	}

	// 如果对话没有关联知识库（全库对话），跳过知识补充
	if conv.KnowledgeBaseID == nil {
		log.Printf("[Review] Conversation %d has no specific KB, skipping knowledge supplement", feedback.ConversationID)
		return
	}

	kbID := *conv.KnowledgeBaseID
	correctAnswer := *feedback.CorrectAnswer

	// 创建 KnowledgeChunk 记录
	chunk := model.KnowledgeChunk{
		KnowledgeBaseID: kbID,
		DocumentID:      0, // 用户贡献不关联特定文档
		DocumentTitle:   "用户贡献",
		Content:         correctAnswer,
		ChunkIndex:      0,
		SourceType:      "user_contribution",
		EmbeddingStatus: "pending",
	}
	if err := mysql.DB.Create(&chunk).Error; err != nil {
		log.Printf("[Review] Failed to create chunk for feedback %d: %v", feedback.ID, err)
		return
	}

	// 异步生成 embedding 并存入 Qdrant
	if qdrant.DefaultVectorClient != nil {
		go generateEmbeddingForUserContribution(kbID, chunk)
	}
}

// generateEmbeddingForUserContribution 为用户贡献的chunk生成embedding
func generateEmbeddingForUserContribution(kbID uint, chunk model.KnowledgeChunk) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// 获取 embedding 设置
	baseURL, apiKey, modelID, err := GetEmbeddingModelSettings()
	if err != nil {
		log.Printf("[Review] Failed to get embedding settings: %v", err)
		mysql.DB.Model(&model.KnowledgeChunk{}).Where("id = ?", chunk.ID).
			Update("embedding_status", "failed")
		return
	}

	embedService := NewEmbeddingService(EmbeddingConfig{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Model:   modelID,
	})

	// 生成 embedding
	embedding, err := embedService.GenerateEmbedding(ctx, chunk.Content)
	if err != nil {
		log.Printf("[Review] Failed to generate embedding for chunk %d: %v", chunk.ID, err)
		mysql.DB.Model(&model.KnowledgeChunk{}).Where("id = ?", chunk.ID).
			Update("embedding_status", "failed")
		return
	}

	// 更新向量维度
	vectorDim := GetEmbeddingDimensionForModel(modelID)
	qdrant.SetVectorSize(vectorDim)

	// 存入 Qdrant
	point := &qdrant.VectorPoint{
		Vector:        embedding,
		DocumentID:    0,
		ChunkID:       chunk.ID,
		ChunkIndex:    chunk.ChunkIndex,
		Content:       chunk.Content,
		DocumentTitle: chunk.DocumentTitle,
		KBID:          kbID,
		RaptorLevel:   0,
	}

	pointIDs, err := qdrant.DefaultVectorClient.BatchUpsertVectors(ctx, []*qdrant.VectorPoint{point})
	if err != nil {
		log.Printf("[Review] Failed to upsert vector for chunk %d: %v", chunk.ID, err)
		mysql.DB.Model(&model.KnowledgeChunk{}).Where("id = ?", chunk.ID).
			Update("embedding_status", "failed")
		return
	}

	if len(pointIDs) > 0 {
		mysql.DB.Model(&model.KnowledgeChunk{}).Where("id = ?", chunk.ID).
			Updates(map[string]interface{}{
				"qdrant_point_id":  pointIDs[0],
				"embedding_status": "embedded",
			})
		log.Printf("[Review] Successfully embedded user contribution chunk %d into KB %d", chunk.ID, kbID)
	}
}
