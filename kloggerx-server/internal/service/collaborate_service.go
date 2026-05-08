package service

import (
	"regexp"
	"strings"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/repository/mysql"
)

// mentionRegex matches @username pattern, username can contain letters, numbers, underscores, hyphens and Chinese characters
var mentionRegex = regexp.MustCompile(`@([a-zA-Z0-9_\-一-龥]+)`)

func AddComment(docID, userID uint, content, quotedText string, parentID *uint) (*model.Comment, error) {
	c := model.Comment{
		DocumentID: docID,
		UserID:     userID,
		Content:    content,
		QuotedText: quotedText,
		ParentID:   parentID,
	}
	if err := mysql.DB.Create(&c).Error; err != nil {
		return nil, err
	}

	// Preload user
	mysql.DB.Preload("User").First(&c, c.ID)

	// Parse @mentions and create notifications asynchronously
	go func() {
		if err := processMentions(docID, userID, content); err != nil {
			// Log error but don't fail the comment creation
			println("Process mentions error:", err.Error())
		}
	}()

	return &c, nil
}

// processMentions parses @username from content and creates notifications
func processMentions(docID, fromUserID uint, content string) error {
	// Find all mentions
	matches := mentionRegex.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}

	// Get unique usernames
	usernameMap := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			username := strings.TrimSpace(match[1])
			if username != "" {
				usernameMap[username] = true
			}
		}
	}

	if len(usernameMap) == 0 {
		return nil
	}

	// Get commenter info
	var commenter model.User
	if err := mysql.DB.Select("username").First(&commenter, fromUserID).Error; err != nil {
		return err
	}

	// Process each mentioned user
	for username := range usernameMap {
		// Find user by username
		var user model.User
		if err := mysql.DB.Where("username = ?", username).First(&user).Error; err != nil {
			// User not found, skip
			continue
		}

		// Skip self-mention
		if user.ID == fromUserID {
			continue
		}

		// Build notification content (first 50 chars of comment)
		contentPreview := content
		if len(contentPreview) > 50 {
			contentPreview = contentPreview[:50] + "..."
		}
		notificationContent := commenter.Username + " 在评论中提及了你: " + contentPreview

		// Create notification
		n := model.Notification{
			Type:       "comment_mention",
			Title:      "评论提及",
			Content:    notificationContent,
			FromUserID: fromUserID,
			ToUserID:   user.ID,
			DocumentID: docID,
		}
		if err := mysql.DB.Create(&n).Error; err != nil {
			// Log error but continue with other mentions
			println("Create mention notification error:", err.Error())
		}
	}

	return nil
}

func GetComments(docID uint) ([]model.Comment, error) {
	var comments []model.Comment
	err := mysql.DB.Where("document_id = ? AND parent_id IS NULL", docID).
		Preload("User").
		Preload("Replies").
		Preload("Replies.User").
		Order("created_at ASC").Find(&comments).Error
	return comments, err
}

func ResolveComment(id uint) error {
	return mysql.DB.Model(&model.Comment{}).Where("id = ?", id).Update("resolved", true).Error
}

func GetCommentByID(id uint) (*model.Comment, error) {
	var comment model.Comment
	if err := mysql.DB.Preload("User").First(&comment, id).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}

func DeleteComment(id uint) error {
	return mysql.DB.Delete(&model.Comment{}, id).Error
}

func GetNotifications(userID uint, page, pageSize int) ([]model.Notification, int64, error) {
	var list []model.Notification
	var total int64
	db := mysql.DB.Model(&model.Notification{}).Where("to_user_id = ?", userID)
	db.Count(&total)
	err := db.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&list).Error
	for i := range list {
		var u model.User
		if mysql.DB.Select("username").First(&u, list[i].FromUserID).Error == nil {
			list[i].FromUserName = u.Username
		}
	}
	return list, total, err
}

func MarkNotificationRead(id uint) error {
	return mysql.DB.Model(&model.Notification{}).Where("id = ?", id).Update("is_read", true).Error
}

func MarkAllNotificationsRead(userID uint) error {
	return mysql.DB.Model(&model.Notification{}).Where("to_user_id = ? AND is_read = ?", userID, false).Update("is_read", true).Error
}

func CreateNotification(typ, title, content string, fromUserID, toUserID, docID uint) error {
	n := model.Notification{
		Type:       typ,
		Title:      title,
		Content:    content,
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		DocumentID: docID,
	}
	return mysql.DB.Create(&n).Error
}

// RecordCollaborateEvent records a collaboration event to the notification table
// This function is designed to be called asynchronously and should not block
// the main WebSocket message flow
func RecordCollaborateEvent(eventType string, documentID uint, userID uint, content string) error {
	// Generate title based on event type
	var title string
	switch eventType {
	case "join":
		title = "用户加入协作"
	case "leave":
		title = "用户离开协作"
	case "concurrent_edit":
		title = "并发编辑提醒"
	default:
		title = "协作事件"
	}

	// Get document owner ID to send notification
	var doc model.Document
	if err := mysql.DB.Select("owner_id").First(&doc, documentID).Error; err != nil {
		return err
	}

	// Don't send notification to self
	if doc.OwnerID == userID {
		return nil
	}

	// Create notification for document owner
	notification := model.Notification{
		Type:       "collaborate_event",
		Title:      title,
		Content:    content,
		FromUserID: userID,
		ToUserID:   doc.OwnerID,
		DocumentID: documentID,
	}

	return mysql.DB.Create(&notification).Error
}

// RecordCollaborateEventAsync is a non-blocking version of RecordCollaborateEvent
// It runs in a goroutine and logs errors instead of returning them
func RecordCollaborateEventAsync(eventType string, documentID uint, userID uint, content string) {
	go func() {
		if err := RecordCollaborateEvent(eventType, documentID, userID, content); err != nil {
			// Log error but don't panic - this is a non-critical operation
			// In production, use proper logging
			println("RecordCollaborateEvent error:", err.Error())
		}
	}()
}

// InviteCollaborators invites users to collaborate on a document
// Creates notifications for each invited user
func InviteCollaborators(documentID uint, inviterID uint, userIDs []uint, message string) error {
	// Get document info
	var doc model.Document
	if err := mysql.DB.Select("id", "title", "owner_id").First(&doc, documentID).Error; err != nil {
		return err
	}

	// Get inviter info
	var inviter model.User
	if err := mysql.DB.Select("username").First(&inviter, inviterID).Error; err != nil {
		return err
	}

	inviterName := inviter.Username
	docTitle := doc.Title
	if docTitle == "" {
		docTitle = "无标题文档"
	}

	// Build notification content
	content := inviterName + " 邀请你协作编辑 " + docTitle
	if message != "" {
		content = content + "\n" + message
	}

	// Create notifications for each invited user
	for _, userID := range userIDs {
		// Skip self-invitation
		if userID == inviterID {
			continue
		}

		n := model.Notification{
			Type:       "collaborate_invite",
			Title:      "协作邀请",
			Content:    content,
			FromUserID: inviterID,
			ToUserID:   userID,
			DocumentID: documentID,
		}
		if err := mysql.DB.Create(&n).Error; err != nil {
			// Log error but continue with other invitations
			println("Create invite notification error:", err.Error())
		}
	}

	return nil
}
