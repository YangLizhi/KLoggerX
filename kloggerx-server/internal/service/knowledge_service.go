package service

import (
	"kloggerx-server/internal/model"
	"kloggerx-server/internal/repository/mysql"
)

func CreateKnowledgeBase(userID uint, name, description string) (*model.KnowledgeBase, error) {
	kb := model.KnowledgeBase{
		Name:        name,
		Description: description,
		OwnerID:     userID,
	}
	if err := mysql.DB.Create(&kb).Error; err != nil {
		return nil, err
	}
	// add owner as member
	mysql.DB.Create(&model.KnowledgeMember{
		KnowledgeBaseID: kb.ID,
		UserID:          userID,
		Role:            "owner",
	})
	return &kb, nil
}

func GetKnowledgeBaseList(page, pageSize int, keyword string) ([]model.KnowledgeBase, int64, error) {
	var list []model.KnowledgeBase
	var total int64
	db := mysql.DB.Model(&model.KnowledgeBase{})
	if keyword != "" {
		db = db.Where("name LIKE ?", "%"+keyword+"%")
	}
	db.Count(&total)
	err := db.Offset((page - 1) * pageSize).Limit(pageSize).Order("updated_at DESC").Find(&list).Error
	for i := range list {
		var u model.User
		if mysql.DB.Select("username").First(&u, list[i].OwnerID).Error == nil {
			list[i].OwnerName = u.Username
		}
		mysql.DB.Model(&model.KnowledgeMember{}).Where("knowledge_base_id = ?", list[i].ID).Count(&total)
		list[i].MemberCount = int(total)
		var docCount int64
		mysql.DB.Model(&model.KnowledgeDocument{}).Where("knowledge_base_id = ?", list[i].ID).Count(&docCount)
		list[i].DocCount = int(docCount)
	}
	return list, total, err
}

func GetKnowledgeBaseDetail(id uint) (*model.KnowledgeBase, error) {
	var kb model.KnowledgeBase
	if err := mysql.DB.First(&kb, id).Error; err != nil {
		return nil, err
	}
	var u model.User
	if mysql.DB.Select("username").First(&u, kb.OwnerID).Error == nil {
		kb.OwnerName = u.Username
	}
	var mc int64
	mysql.DB.Model(&model.KnowledgeMember{}).Where("knowledge_base_id = ?", kb.ID).Count(&mc)
	kb.MemberCount = int(mc)
	var dc int64
	mysql.DB.Model(&model.KnowledgeDocument{}).Where("knowledge_base_id = ?", kb.ID).Count(&dc)
	kb.DocCount = int(dc)
	return &kb, nil
}

func DeleteKnowledgeBase(id uint) error {
	mysql.DB.Where("knowledge_base_id = ?", id).Delete(&model.KnowledgeMember{})
	mysql.DB.Where("knowledge_base_id = ?", id).Delete(&model.KnowledgeDocument{})
	return mysql.DB.Delete(&model.KnowledgeBase{}, id).Error
}

func UpdateKnowledgeBaseInfo(id uint, updates map[string]interface{}) error {
	return mysql.DB.Model(&model.KnowledgeBase{}).Where("id = ?", id).Updates(updates).Error
}

func GetKnowledgeBaseTree(kbID uint) ([]model.Document, error) {
	var kds []model.KnowledgeDocument
	mysql.DB.Where("knowledge_base_id = ?", kbID).Find(&kds)
	var docs []model.Document
	for _, kd := range kds {
		var d model.Document
		if mysql.DB.First(&d, kd.DocumentID).Error == nil {
			var u model.User
			if mysql.DB.Select("username").First(&u, d.OwnerID).Error == nil {
				d.OwnerName = u.Username
			}
			docs = append(docs, d)
		}
	}
	return docs, nil
}

func AddKnowledgeMember(kbID, userID uint, role string) error {
	m := model.KnowledgeMember{KnowledgeBaseID: kbID, UserID: userID, Role: role}
	return mysql.DB.Where("knowledge_base_id = ? AND user_id = ?", kbID, userID).FirstOrCreate(&m).Error
}

func RemoveKnowledgeMember(kbID, userID uint) error {
	return mysql.DB.Where("knowledge_base_id = ? AND user_id = ?", kbID, userID).Delete(&model.KnowledgeMember{}).Error
}

func GetKnowledgeMembers(kbID uint) ([]model.KnowledgeMember, error) {
	var members []model.KnowledgeMember
	err := mysql.DB.Where("knowledge_base_id = ?", kbID).Find(&members).Error
	for i := range members {
		var u model.User
		if mysql.DB.Select("username, avatar").First(&u, members[i].UserID).Error == nil {
			members[i].UserName = u.Username
			members[i].UserAvatar = u.Avatar
		}
	}
	return members, err
}

func PublishDocument(kbID, docID uint) error {
	kd := model.KnowledgeDocument{KnowledgeBaseID: kbID, DocumentID: docID}
	return mysql.DB.Where("knowledge_base_id = ? AND document_id = ?", kbID, docID).FirstOrCreate(&kd).Error
}

func SearchKnowledge(keyword string, kbID *uint, page, pageSize int) ([]model.Document, int64, error) {
	var docIDs []uint
	db := mysql.DB.Model(&model.KnowledgeDocument{})
	if kbID != nil {
		db = db.Where("knowledge_base_id = ?", *kbID)
	}
	db.Pluck("document_id", &docIDs)

	if len(docIDs) == 0 {
		return nil, 0, nil
	}

	var docs []model.Document
	var total int64
	qdb := mysql.DB.Model(&model.Document{}).Where("id IN ? AND is_deleted = ? AND title LIKE ?", docIDs, false, "%"+keyword+"%")
	qdb.Count(&total)
	err := qdb.Offset((page - 1) * pageSize).Limit(pageSize).Find(&docs).Error
	return docs, total, err
}
