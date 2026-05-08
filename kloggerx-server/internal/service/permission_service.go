package service

import (
	"errors"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/repository/mysql"
)

var (
	ErrPermissionDenied = errors.New("无权限执行此操作")
)

// CheckDocumentPermission 检查用户对文档的权限
// action: "read" | "edit" | "delete" | "admin"
func CheckDocumentPermission(userID uint, docID uint, action string) error {
	// 1. 查用户角色，admin bypass
	var user model.User
	if err := mysql.DB.Select("id, role").First(&user, userID).Error; err != nil {
		return ErrPermissionDenied
	}
	if user.Role == "admin" {
		return nil
	}

	// 2. 查文档获取 ownerID
	var doc model.Document
	if err := mysql.DB.Select("id, owner_id").First(&doc, docID).Error; err != nil {
		return errors.New("文档不存在")
	}

	// 3. 如果用户是 owner，允许所有操作
	if doc.OwnerID == userID {
		return nil
	}

	// 4. 查 permissions 表获取用户对该文档的权限级别
	var perm model.Permission
	err := mysql.DB.Where("document_id = ? AND user_id = ?", docID, userID).First(&perm).Error
	if err != nil {
		// 没有任何权限记录
		return ErrPermissionDenied
	}

	// 5. 对比 action 需要的权限级别
	// 权限级别: view < edit < admin
	switch action {
	case "read":
		// view, edit, admin 都可以
		return nil
	case "edit":
		if perm.Level == "edit" || perm.Level == "admin" {
			return nil
		}
		return ErrPermissionDenied
	case "delete", "admin":
		if perm.Level == "admin" {
			return nil
		}
		return ErrPermissionDenied
	default:
		return ErrPermissionDenied
	}
}

// CheckKnowledgePermission 检查用户对知识库的权限
// action: "read" | "edit" | "delete" | "admin"
func CheckKnowledgePermission(userID uint, kbID uint, action string) error {
	// 1. 查用户角色，admin bypass
	var user model.User
	if err := mysql.DB.Select("id, role").First(&user, userID).Error; err != nil {
		return ErrPermissionDenied
	}
	if user.Role == "admin" {
		return nil
	}

	// 2. 查知识库获取 ownerID
	var kb model.KnowledgeBase
	if err := mysql.DB.Select("id, owner_id").First(&kb, kbID).Error; err != nil {
		return errors.New("知识库不存在")
	}

	// 3. 如果用户是 owner，允许所有操作
	if kb.OwnerID == userID {
		return nil
	}

	// 4. 查 knowledge_members 表获取用户角色
	var member model.KnowledgeMember
	err := mysql.DB.Where("knowledge_base_id = ? AND user_id = ?", kbID, userID).First(&member).Error
	if err != nil {
		// 不是成员
		return ErrPermissionDenied
	}

	// 5. 对比所需权限
	// 成员角色: member < admin
	switch action {
	case "read":
		// 任何成员都可以读
		return nil
	case "edit":
		if member.Role == "admin" {
			return nil
		}
		return ErrPermissionDenied
	case "delete", "admin":
		// 只有 owner 才能删除，已在上面判断过
		return ErrPermissionDenied
	default:
		return ErrPermissionDenied
	}
}
