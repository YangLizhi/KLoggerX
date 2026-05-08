package service

import (
	"errors"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/repository/mysql"
)

// CreateComment 创建评论
func CreateComment(userID, docID uint, content string, parentID *uint, quotedText string) (*model.Comment, error) {
	comment := model.Comment{
		DocumentID: docID,
		UserID:     userID,
		Content:    content,
		ParentID:   parentID,
		QuotedText: quotedText,
	}
	if err := mysql.DB.Create(&comment).Error; err != nil {
		return nil, err
	}

	// Preload user info
	mysql.DB.Preload("User").First(&comment, comment.ID)

	// Parse @mentions and create notifications asynchronously
	go func() {
		if err := processMentions(docID, userID, content); err != nil {
			println("CreateComment processMentions error:", err.Error())
		}
	}()

	return &comment, nil
}

// ListComments 获取文档所有评论（树形结构，按时间正序）
func ListComments(docID uint) ([]model.Comment, error) {
	var comments []model.Comment
	err := mysql.DB.Where("document_id = ? AND parent_id IS NULL", docID).
		Preload("User").
		Preload("Replies").
		Preload("Replies.User").
		Order("created_at ASC").
		Find(&comments).Error
	return comments, err
}

// DeleteCommentByUser 删除评论（仅作者或文档owner可删）
func DeleteCommentByUser(userID, commentID uint) error {
	var comment model.Comment
	if err := mysql.DB.First(&comment, commentID).Error; err != nil {
		return errors.New("评论不存在")
	}

	// Check permission: comment author or document owner
	if comment.UserID != userID {
		var doc model.Document
		if err := mysql.DB.Select("owner_id").First(&doc, comment.DocumentID).Error; err != nil {
			return errors.New("文档不存在")
		}
		if doc.OwnerID != userID {
			return errors.New("无权限删除此评论")
		}
	}

	// Delete replies first
	mysql.DB.Where("parent_id = ?", commentID).Delete(&model.Comment{})
	return mysql.DB.Delete(&model.Comment{}, commentID).Error
}

// ResolveCommentByUser 标记评论为已解决
func ResolveCommentByUser(userID, commentID uint) error {
	var comment model.Comment
	if err := mysql.DB.First(&comment, commentID).Error; err != nil {
		return errors.New("评论不存在")
	}

	// Check permission: comment author, document owner, or admin can resolve
	if comment.UserID != userID {
		var doc model.Document
		if err := mysql.DB.Select("owner_id").First(&doc, comment.DocumentID).Error; err != nil {
			return errors.New("文档不存在")
		}
		if doc.OwnerID != userID {
			return errors.New("无权限标记此评论")
		}
	}

	return mysql.DB.Model(&model.Comment{}).Where("id = ?", commentID).Update("resolved", true).Error
}

// GetCommentCount 获取文档评论数
func GetCommentCount(docID uint) (int64, error) {
	var count int64
	err := mysql.DB.Model(&model.Comment{}).Where("document_id = ?", docID).Count(&count).Error
	return count, err
}
