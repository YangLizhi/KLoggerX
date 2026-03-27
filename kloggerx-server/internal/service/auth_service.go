package service

import (
	"kloggerx-server/internal/model"
	"kloggerx-server/internal/repository/mysql"
)

func GetDocumentPermissions(docID uint) ([]model.Permission, error) {
	var perms []model.Permission
	err := mysql.DB.Where("document_id = ?", docID).Find(&perms).Error
	for i := range perms {
		var u model.User
		if mysql.DB.Select("username, avatar").First(&u, perms[i].UserID).Error == nil {
			perms[i].UserName = u.Username
			perms[i].UserAvatar = u.Avatar
		}
	}
	return perms, err
}

func SetPermission(docID, userID uint, level string) error {
	var perm model.Permission
	result := mysql.DB.Where("document_id = ? AND user_id = ?", docID, userID).First(&perm)
	if result.Error != nil {
		perm = model.Permission{DocumentID: docID, UserID: userID, Level: level}
		return mysql.DB.Create(&perm).Error
	}
	return mysql.DB.Model(&perm).Update("level", level).Error
}

func RemovePermission(docID, userID uint) error {
	return mysql.DB.Where("document_id = ? AND user_id = ?", docID, userID).Delete(&model.Permission{}).Error
}

func GetShareSetting(docID uint) (*model.ShareSetting, error) {
	var s model.ShareSetting
	err := mysql.DB.Where("document_id = ?", docID).First(&s).Error
	if err != nil {
		return &model.ShareSetting{DocumentID: docID, Scope: "collaborator", DefaultPermission: "view"}, nil
	}
	return &s, nil
}

func UpdateShareSetting(s *model.ShareSetting) error {
	var existing model.ShareSetting
	result := mysql.DB.Where("document_id = ?", s.DocumentID).First(&existing)
	if result.Error != nil {
		return mysql.DB.Create(s).Error
	}
	return mysql.DB.Model(&existing).Updates(s).Error
}

func CheckPermission(docID, userID uint) string {
	var doc model.Document
	if mysql.DB.First(&doc, docID).Error == nil && doc.OwnerID == userID {
		return "owner"
	}
	var perm model.Permission
	if mysql.DB.Where("document_id = ? AND user_id = ?", docID, userID).First(&perm).Error == nil {
		return perm.Level
	}
	var share model.ShareSetting
	if mysql.DB.Where("document_id = ?", docID).First(&share).Error == nil {
		if share.Scope == "public" || share.Scope == "organization" {
			return share.DefaultPermission
		}
	}
	return ""
}
