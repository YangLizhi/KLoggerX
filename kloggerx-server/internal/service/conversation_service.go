package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/repository/mysql"

	"gorm.io/gorm"
)

// GetConversationList 获取用户的对话列表（分页，按更新时间倒序）
func GetConversationList(userID uint, knowledgeBaseID *uint, page, pageSize int) ([]model.KbConversation, int64, error) {
	var list []model.KbConversation
	var total int64
	db := mysql.DB.Model(&model.KbConversation{}).Where("user_id = ?", userID)
	if knowledgeBaseID != nil {
		db = db.Where("knowledge_base_id = ?", *knowledgeBaseID)
	}
	db.Count(&total)
	err := db.Offset((page - 1) * pageSize).Limit(pageSize).Order("is_pinned DESC, updated_at DESC").Find(&list).Error
	return list, total, err
}

// CreateConversation 创建新对话
func CreateConversation(userID uint, knowledgeBaseID *uint, modelName string) (*model.KbConversation, error) {
	conv := model.KbConversation{
		UserID:          userID,
		KnowledgeBaseID: knowledgeBaseID,
		Title:           "新对话",
		Model:           modelName,
	}
	if err := mysql.DB.Create(&conv).Error; err != nil {
		return nil, err
	}
	return &conv, nil
}

// GetConversationDetail 获取对话详情（含消息列表，支持分页）
func GetConversationDetail(conversationID, userID uint, msgPage, msgPageSize int) (*model.KbConversation, []model.KbMessage, int64, error) {
	var conv model.KbConversation
	if err := mysql.DB.Where("id = ? AND user_id = ?", conversationID, userID).First(&conv).Error; err != nil {
		return nil, nil, 0, fmt.Errorf("对话不存在")
	}

	var messages []model.KbMessage
	var msgTotal int64
	msgDB := mysql.DB.Model(&model.KbMessage{}).Where("conversation_id = ?", conversationID)
	msgDB.Count(&msgTotal)
	err := msgDB.Offset((msgPage - 1) * msgPageSize).Limit(msgPageSize).Order("created_at ASC").Find(&messages).Error
	if err != nil {
		return nil, nil, 0, err
	}
	return &conv, messages, msgTotal, nil
}

// UpdateConversation 更新对话（重命名、置顶等）
func UpdateConversation(conversationID, userID uint, updates map[string]interface{}) error {
	result := mysql.DB.Model(&model.KbConversation{}).
		Where("id = ? AND user_id = ?", conversationID, userID).
		Updates(updates)
	if result.RowsAffected == 0 {
		return fmt.Errorf("对话不存在或无权限")
	}
	return result.Error
}

// DeleteConversation 删除对话（软删除，同时标记消息）
func DeleteConversation(conversationID, userID uint) error {
	// 校验所有权
	var conv model.KbConversation
	if err := mysql.DB.Where("id = ? AND user_id = ?", conversationID, userID).First(&conv).Error; err != nil {
		return fmt.Errorf("对话不存在或无权限")
	}

	return mysql.DB.Transaction(func(tx *gorm.DB) error {
		// 软删除对话
		if err := tx.Delete(&model.KbConversation{}, conversationID).Error; err != nil {
			return err
		}
		// 级联删除相关消息（硬删除，因为消息无 DeletedAt 字段）
		if err := tx.Where("conversation_id = ?", conversationID).Delete(&model.KbMessage{}).Error; err != nil {
			return err
		}
		// 删除相关反馈
		if err := tx.Where("conversation_id = ?", conversationID).Delete(&model.KbFeedback{}).Error; err != nil {
			return err
		}
		return nil
	})
}

// BatchDeleteConversations 批量删除对话
func BatchDeleteConversations(conversationIDs []uint, userID uint) error {
	if len(conversationIDs) == 0 {
		return nil
	}
	return mysql.DB.Transaction(func(tx *gorm.DB) error {
		// 只删除属于该用户的对话
		result := tx.Where("id IN ? AND user_id = ?", conversationIDs, userID).Delete(&model.KbConversation{})
		if result.Error != nil {
			return result.Error
		}
		// 级联删除消息和反馈
		if err := tx.Where("conversation_id IN ?", conversationIDs).Delete(&model.KbMessage{}).Error; err != nil {
			return err
		}
		if err := tx.Where("conversation_id IN ?", conversationIDs).Delete(&model.KbFeedback{}).Error; err != nil {
			return err
		}
		return nil
	})
}

// ─── Export ──────────────────────────────────────────────────────────────────

// ExportConversation exports a conversation as Markdown (or PDF placeholder).
// Returns file content bytes, filename, content-type, and error.
func ExportConversation(conversationID, userID uint, format string) ([]byte, string, string, error) {
	// Verify ownership
	var conv model.KbConversation
	if err := mysql.DB.Where("id = ? AND user_id = ?", conversationID, userID).First(&conv).Error; err != nil {
		return nil, "", "", fmt.Errorf("对话不存在或无权限")
	}

	// Get all messages
	var messages []model.KbMessage
	mysql.DB.Where("conversation_id = ?", conversationID).Order("created_at ASC").Find(&messages)

	// Build Markdown content
	var sb strings.Builder
	sb.WriteString("# " + conv.Title + "\n\n")
	sb.WriteString(fmt.Sprintf("导出时间：%s\n\n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString("---\n\n")

	for _, msg := range messages {
		switch msg.Role {
		case "user":
			sb.WriteString("## 用户\n\n")
		case "assistant":
			sb.WriteString("## AI助手\n\n")
		case "system":
			sb.WriteString("## 系统\n\n")
		}
		sb.WriteString(msg.Content + "\n\n")
		sb.WriteString("---\n\n")
	}

	mdContent := []byte(sb.String())
	filename := conv.Title + ".md"

	switch format {
	case "markdown", "md":
		return mdContent, filename, "text/markdown", nil
	case "pdf":
		// PDF export not yet supported, return error
		return nil, "", "", fmt.Errorf("PDF导出即将支持，请暂时使用Markdown格式")
	default:
		return mdContent, filename, "text/markdown", nil
	}
}

// ─── Share ───────────────────────────────────────────────────────────────────

// CreateShareLink creates a share link for a conversation.
func CreateShareLink(conversationID, userID uint, expiresInHours *int) (*model.KbShareLink, error) {
	// Verify ownership
	var conv model.KbConversation
	if err := mysql.DB.Where("id = ? AND user_id = ?", conversationID, userID).First(&conv).Error; err != nil {
		return nil, fmt.Errorf("对话不存在或无权限")
	}

	// Revoke existing active links for this conversation
	mysql.DB.Model(&model.KbShareLink{}).Where("conversation_id = ? AND user_id = ? AND is_active = ?", conversationID, userID, true).
		Update("is_active", false)

	// Generate random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("生成分享令牌失败")
	}
	token := hex.EncodeToString(tokenBytes)

	var expiresAt *time.Time
	if expiresInHours != nil && *expiresInHours > 0 {
		t := time.Now().Add(time.Duration(*expiresInHours) * time.Hour)
		expiresAt = &t
	}

	link := model.KbShareLink{
		ConversationID: conversationID,
		UserID:         userID,
		ShareToken:     token,
		ExpiresAt:      expiresAt,
		IsActive:       true,
	}
	if err := mysql.DB.Create(&link).Error; err != nil {
		return nil, fmt.Errorf("创建分享链接失败")
	}
	return &link, nil
}

// RevokeShareLink revokes all active share links for a conversation.
func RevokeShareLink(conversationID, userID uint) error {
	result := mysql.DB.Model(&model.KbShareLink{}).
		Where("conversation_id = ? AND user_id = ? AND is_active = ?", conversationID, userID, true).
		Update("is_active", false)
	if result.RowsAffected == 0 {
		return fmt.Errorf("没有可取消的分享链接")
	}
	return result.Error
}

// GetSharedConversation retrieves a shared conversation by token (public access).
func GetSharedConversation(token string) (*model.KbConversation, []model.KbMessage, error) {
	var link model.KbShareLink
	if err := mysql.DB.Where("share_token = ? AND is_active = ?", token, true).First(&link).Error; err != nil {
		return nil, nil, fmt.Errorf("分享链接无效或已过期")
	}

	// Check expiration
	if link.ExpiresAt != nil && link.ExpiresAt.Before(time.Now()) {
		return nil, nil, fmt.Errorf("分享链接已过期")
	}

	// Increment view count
	mysql.DB.Model(&link).Update("view_count", gorm.Expr("view_count + 1"))

	// Get conversation
	var conv model.KbConversation
	if err := mysql.DB.First(&conv, link.ConversationID).Error; err != nil {
		return nil, nil, fmt.Errorf("对话不存在")
	}

	// Get messages
	var messages []model.KbMessage
	mysql.DB.Where("conversation_id = ?", link.ConversationID).Order("created_at ASC").Find(&messages)

	return &conv, messages, nil
}

// GenerateConversationTitle 自动生成对话标题（异步，调用LLM总结首次对话内容）
func GenerateConversationTitle(conversationID uint, firstMessage string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("GenerateConversationTitle panic: %v", r)
			}
		}()

		systemPrompt := "你是一个标题生成助手。请根据用户的第一条消息，生成一个简短的对话标题（不超过20个字）。只返回标题文本，不要加引号或其他格式。"
		title, err := callAIChat(systemPrompt, firstMessage, nil)
		if err != nil {
			log.Printf("GenerateConversationTitle error for conv %d: %v", conversationID, err)
			return
		}

		// 截断过长标题
		if len([]rune(title)) > 50 {
			title = string([]rune(title)[:50])
		}

		mysql.DB.Model(&model.KbConversation{}).
			Where("id = ?", conversationID).
			Updates(map[string]interface{}{
				"title":      title,
				"updated_at": time.Now(),
			})
	}()
}
