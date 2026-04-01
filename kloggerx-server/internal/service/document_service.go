package service

import (
	"errors"
	"time"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/repository/mysql"
	rds "kloggerx-server/internal/repository/redis"
)

type CreateDocReq struct {
	Title        string `json:"title" binding:"required"`
	Type         string `json:"type" binding:"required"`
	ParentID     *uint  `json:"parentId"`
	FileSize     int64  `json:"fileSize"`
	FileExt      string `json:"fileExt"`
	OriginalName string `json:"originalName"`
}

func CreateDocument(userID uint, req CreateDocReq) (*model.Document, error) {
	doc := model.Document{
		Title:        req.Title,
		Type:         req.Type,
		ParentID:     req.ParentID,
		OwnerID:      userID,
		Version:      1,
		FileSize:     req.FileSize,
		FileExt:      req.FileExt,
		OriginalName: req.OriginalName,
	}
	if err := mysql.DB.Create(&doc).Error; err != nil {
		return nil, err
	}
	rds.InvalidateDocTreeCache(userID)
	return &doc, nil
}

func GetDocumentTree(userID uint, parentID *uint) ([]model.Document, error) {
	// Try cache for root-level tree only
	if parentID == nil {
		cacheKey := rds.DocTreeCacheKey(userID)
		var cached []model.Document
		if err := rds.GetCache(cacheKey, &cached); err == nil {
			return cached, nil
		}
	}

	var docs []model.Document
	db := mysql.DB.Where("owner_id = ? AND is_deleted = ?", userID, false)
	if parentID == nil {
		db = db.Where("parent_id IS NULL")
	} else {
		db = db.Where("parent_id = ?", *parentID)
	}
	err := db.Order("is_pinned DESC, updated_at DESC").Find(&docs).Error
	if err != nil {
		return nil, err
	}
	for i := range docs {
		var fav model.Favorite
		if mysql.DB.Where("user_id = ? AND document_id = ?", userID, docs[i].ID).First(&fav).Error == nil {
			docs[i].IsFavorite = true
		}
		var owner model.User
		if mysql.DB.Select("username").First(&owner, docs[i].OwnerID).Error == nil {
			docs[i].OwnerName = owner.Username
		}
	}

	// Cache root-level tree for 5 minutes
	if parentID == nil {
		rds.SetCache(rds.DocTreeCacheKey(userID), docs, 5*time.Minute)
	}

	return docs, nil
}

func GetDocumentDetail(docID, userID uint) (*model.Document, error) {
	var doc model.Document
	if err := mysql.DB.First(&doc, docID).Error; err != nil {
		return nil, err
	}
	var owner model.User
	if mysql.DB.Select("username").First(&owner, doc.OwnerID).Error == nil {
		doc.OwnerName = owner.Username
	}
	var fav model.Favorite
	if mysql.DB.Where("user_id = ? AND document_id = ?", userID, docID).First(&fav).Error == nil {
		doc.IsFavorite = true
	}
	// record recent
	now := time.Now()
	mysql.DB.Where("user_id = ? AND document_id = ?", userID, docID).Delete(&model.RecentDocument{})
	mysql.DB.Create(&model.RecentDocument{UserID: userID, DocumentID: docID, AccessedAt: now})
	return &doc, nil
}

func SaveDocumentContent(docID, userID uint, content string) error {
	var doc model.Document
	if err := mysql.DB.First(&doc, docID).Error; err != nil {
		return err
	}
	// save version
	mysql.DB.Create(&model.DocumentVersion{
		DocumentID: docID,
		Version:    doc.Version,
		Content:    doc.Content,
		EditorID:   userID,
	})
	return mysql.DB.Model(&doc).Updates(map[string]interface{}{
		"content": content,
		"version": doc.Version + 1,
	}).Error
}

func UpdateDocumentTitle(docID uint, title string) error {
	return mysql.DB.Model(&model.Document{}).Where("id = ?", docID).Update("title", title).Error
}

func DeleteDocument(docID uint) error {
	now := time.Now()
	err := mysql.DB.Model(&model.Document{}).Where("id = ?", docID).Updates(map[string]interface{}{
		"is_deleted": true,
		"deleted_at": now,
	}).Error
	if err == nil {
		rds.InvalidateAllDocTreeCache()
	}
	return err
}

func RestoreDocument(docID uint) error {
	err := mysql.DB.Model(&model.Document{}).Where("id = ?", docID).Updates(map[string]interface{}{
		"is_deleted": false,
		"deleted_at": nil,
	}).Error
	if err == nil {
		rds.InvalidateAllDocTreeCache()
	}
	return err
}

func PermanentDeleteDocument(docID uint) error {
	err := mysql.DB.Delete(&model.Document{}, docID).Error
	if err == nil {
		rds.InvalidateAllDocTreeCache()
	}
	return err
}

func MoveDocument(docID uint, targetParentID *uint) error {
	err := mysql.DB.Model(&model.Document{}).Where("id = ?", docID).Update("parent_id", targetParentID).Error
	if err == nil {
		rds.InvalidateAllDocTreeCache()
	}
	return err
}

func CopyDocument(docID, userID uint) (*model.Document, error) {
	var src model.Document
	if err := mysql.DB.First(&src, docID).Error; err != nil {
		return nil, err
	}
	doc := model.Document{
		Title:        src.Title + " (副本)",
		Type:         src.Type,
		ParentID:     src.ParentID,
		OwnerID:      userID,
		Content:      src.Content,
		Version:      1,
		FileSize:     src.FileSize,
		FileExt:      src.FileExt,
		OriginalName: src.OriginalName,
	}
	if err := mysql.DB.Create(&doc).Error; err != nil {
		return nil, err
	}
	rds.InvalidateDocTreeCache(userID)
	return &doc, nil
}

func PinDocument(docID uint, isPinned bool) error {
	err := mysql.DB.Model(&model.Document{}).Where("id = ?", docID).Update("is_pinned", isPinned).Error
	if err == nil {
		rds.InvalidateAllDocTreeCache()
	}
	return err
}

func FavoriteDocument(userID, docID uint, isFavorite bool) error {
	if isFavorite {
		fav := model.Favorite{UserID: userID, DocumentID: docID}
		return mysql.DB.Where("user_id = ? AND document_id = ?", userID, docID).FirstOrCreate(&fav).Error
	}
	return mysql.DB.Where("user_id = ? AND document_id = ?", userID, docID).Delete(&model.Favorite{}).Error
}

func TransferOwnership(docID, targetUserID uint) error {
	return mysql.DB.Model(&model.Document{}).Where("id = ?", docID).Update("owner_id", targetUserID).Error
}

func GetRecycleBin(userID uint, page, pageSize int) ([]model.Document, int64, error) {
	var docs []model.Document
	var total int64
	db := mysql.DB.Model(&model.Document{}).Where("owner_id = ? AND is_deleted = ?", userID, true)
	db.Count(&total)
	err := db.Offset((page - 1) * pageSize).Limit(pageSize).Order("deleted_at DESC").Find(&docs).Error
	return docs, total, err
}

func GetFavorites(userID uint, page, pageSize int) ([]model.Document, int64, error) {
	var favs []model.Favorite
	var total int64
	db := mysql.DB.Model(&model.Favorite{}).Where("user_id = ?", userID)
	db.Count(&total)
	db.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&favs)
	var docs []model.Document
	for _, f := range favs {
		var d model.Document
		if mysql.DB.First(&d, f.DocumentID).Error == nil {
			d.IsFavorite = true
			docs = append(docs, d)
		}
	}
	return docs, total, nil
}

func GetPinnedDocuments(userID uint) ([]model.Document, error) {
	var docs []model.Document
	err := mysql.DB.Where("owner_id = ? AND is_pinned = ? AND is_deleted = ?", userID, true, false).Find(&docs).Error
	return docs, err
}

func GetRecentDocuments(userID uint, page, pageSize int) ([]model.Document, int64, error) {
	var recents []model.RecentDocument
	var total int64
	db := mysql.DB.Model(&model.RecentDocument{}).Where("user_id = ?", userID)
	db.Count(&total)
	db.Offset((page - 1) * pageSize).Limit(pageSize).Order("accessed_at DESC").Find(&recents)
	var docs []model.Document
	for _, r := range recents {
		var d model.Document
		if mysql.DB.First(&d, r.DocumentID).Error == nil && !d.IsDeleted {
			docs = append(docs, d)
		}
	}
	return docs, total, nil
}

func SearchDocuments(userID uint, keyword string, page, pageSize int) ([]model.Document, int64, error) {
	var docs []model.Document
	var total int64
	db := mysql.DB.Model(&model.Document{}).Where("owner_id = ? AND is_deleted = ?", userID, false).
		Where("title LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	db.Count(&total)
	err := db.Offset((page - 1) * pageSize).Limit(pageSize).Order("updated_at DESC").Find(&docs).Error
	return docs, total, err
}

func GetDocumentVersions(docID uint) ([]model.DocumentVersion, error) {
	var versions []model.DocumentVersion
	err := mysql.DB.Where("document_id = ?", docID).Order("version DESC").Find(&versions).Error
	for i := range versions {
		var u model.User
		if mysql.DB.Select("username").First(&u, versions[i].EditorID).Error == nil {
			versions[i].EditorName = u.Username
		}
	}
	return versions, err
}

func RollbackVersion(docID uint, version int) error {
	var v model.DocumentVersion
	if err := mysql.DB.Where("document_id = ? AND version = ?", docID, version).First(&v).Error; err != nil {
		return errors.New("版本不存在")
	}
	return mysql.DB.Model(&model.Document{}).Where("id = ?", docID).Updates(map[string]interface{}{
		"content": v.Content,
		"version": version,
	}).Error
}
