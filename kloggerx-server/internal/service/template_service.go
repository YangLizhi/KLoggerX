package service

import (
	"kloggerx-server/internal/model"
	"kloggerx-server/internal/repository/mysql"
)

func GetTemplates(category string) ([]model.Template, error) {
	var list []model.Template
	db := mysql.DB.Model(&model.Template{})
	if category != "" {
		db = db.Where("category = ?", category)
	}
	err := db.Order("id ASC").Find(&list).Error
	return list, err
}

func GetTemplateCategories() ([]string, error) {
	var cats []string
	err := mysql.DB.Model(&model.Template{}).Distinct("category").Pluck("category", &cats).Error
	return cats, err
}

func GetTemplateDetail(id uint) (*model.Template, error) {
	var t model.Template
	err := mysql.DB.First(&t, id).Error
	return &t, err
}
