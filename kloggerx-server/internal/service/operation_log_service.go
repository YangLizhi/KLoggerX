package service

import (
	"kloggerx-server/internal/model"
	"kloggerx-server/internal/repository/mysql"
)

func CreateOperationLog(userID uint, userName, action, resourceType string, resourceID uint, resourceTitle, detail, ip string) {
	log := model.OperationLog{
		UserID:        userID,
		UserName:      userName,
		Action:        action,
		ResourceType:  resourceType,
		ResourceID:    resourceID,
		ResourceTitle: resourceTitle,
		Detail:        detail,
		IP:            ip,
	}
	mysql.DB.Create(&log)
}

func GetOperationLogs(resourceType string, resourceID uint, page, pageSize int) ([]model.OperationLog, int64, error) {
	var logs []model.OperationLog
	var total int64

	query := mysql.DB.Where("resource_type = ? AND resource_id = ?", resourceType, resourceID)
	query.Model(&model.OperationLog{}).Count(&total)

	err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&logs).Error

	return logs, total, err
}
