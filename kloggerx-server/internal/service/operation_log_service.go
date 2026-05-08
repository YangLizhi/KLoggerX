package service

import (
	"time"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/repository/mysql"
)

// OperationLogFilter 操作日志过滤条件
type OperationLogFilter struct {
	UserID    uint
	Action    string
	StartTime *time.Time
	EndTime   *time.Time
	Page      int
	PageSize  int
}

// CreateOperationLog 创建操作日志（业务层手动调用）
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

// CreateAuditLog 审计中间件异步写入日志
func CreateAuditLog(log *model.OperationLog) error {
	return mysql.DB.Create(log).Error
}

// GetOperationLogs 按资源查询操作日志（原有接口，保持兼容）
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

// ListOperationLogs 查询操作日志（支持过滤）
func ListOperationLogs(filter OperationLogFilter) ([]model.OperationLog, int64, error) {
	var logs []model.OperationLog
	var total int64

	query := mysql.DB.Model(&model.OperationLog{})

	if filter.UserID > 0 {
		query = query.Where("user_id = ?", filter.UserID)
	}
	if filter.Action != "" {
		query = query.Where("action LIKE ?", "%"+filter.Action+"%")
	}
	if filter.StartTime != nil {
		query = query.Where("created_at >= ?", *filter.StartTime)
	}
	if filter.EndTime != nil {
		query = query.Where("created_at <= ?", *filter.EndTime)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}

	err := query.Order("created_at DESC").
		Offset((filter.Page - 1) * filter.PageSize).
		Limit(filter.PageSize).
		Find(&logs).Error

	return logs, total, err
}
