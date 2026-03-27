package service

import (
	"kloggerx-server/internal/model"
	"kloggerx-server/internal/repository/mysql"
)

func AddComment(docID, userID uint, content, selection string, parentID *uint) (*model.Comment, error) {
	c := model.Comment{
		DocumentID: docID,
		UserID:     userID,
		Content:    content,
		Selection:  selection,
		ParentID:   parentID,
	}
	if err := mysql.DB.Create(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func GetComments(docID uint) ([]model.Comment, error) {
	var comments []model.Comment
	err := mysql.DB.Where("document_id = ?", docID).Order("created_at ASC").Find(&comments).Error
	for i := range comments {
		var u model.User
		if mysql.DB.Select("username").First(&u, comments[i].UserID).Error == nil {
			comments[i].UserName = u.Username
		}
	}
	return comments, err
}

func ResolveComment(id uint) error {
	return mysql.DB.Model(&model.Comment{}).Where("id = ?", id).Update("resolved", true).Error
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
