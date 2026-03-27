package mysql

import (
	"fmt"
	"log"

	"kloggerx-server/config"
	"kloggerx-server/internal/model"
	"kloggerx-server/internal/pkg/utils"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init(cfg config.DatabaseConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.Charset)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Info),
	})
	if err != nil {
		return err
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	return nil
}

func AutoMigrate() {
	DB.AutoMigrate(
		&model.User{},
		&model.Department{},
		&model.Document{},
		&model.DocumentVersion{},
		&model.Favorite{},
		&model.Permission{},
		&model.ShareSetting{},
		&model.KnowledgeBase{},
		&model.KnowledgeMember{},
		&model.KnowledgeDocument{},
		&model.Comment{},
		&model.Notification{},
		&model.RecentDocument{},
		&model.FileRecord{},
		&model.OperationLog{},
		&model.SystemSetting{},
	)

	seedData()
}

func seedData() {
	var count int64
	DB.Model(&model.User{}).Count(&count)
	if count > 0 {
		return
	}

	log.Println("[Seed] Creating default admin user and sample data...")

	hashedPwd, err := utils.HashPassword("admin")
	if err != nil {
		log.Printf("[Seed] Failed to hash password: %v", err)
		return
	}

	admin := model.User{
		Username: "admin",
		Email:    "admin@kloggerx.com",
		Password: hashedPwd,
		Nickname: "管理员",
		Role:     "admin",
	}
	if err := DB.Create(&admin).Error; err != nil {
		log.Printf("[Seed] Failed to create admin: %v", err)
		return
	}

	dept := model.Department{Name: "KLoggerX"}
	DB.Create(&dept)

	DB.Model(&admin).Update("department_id", dept.ID)

	welcomeDoc := model.Document{
		Title:   "欢迎使用 KLoggerX",
		Type:    "doc",
		OwnerID: admin.ID,
		Content: `{"type":"doc","content":[{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"欢迎使用 KLoggerX"}]},{"type":"paragraph","content":[{"type":"text","text":"KLoggerX 是一款私有化在线文档与知识管理平台，支持多人实时协作编辑、结构化知识沉淀、细粒度权限管控。"}]},{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"快速开始"}]},{"type":"bulletList","content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"点击左侧「新建文档」创建你的第一篇文档"}]}]},{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"使用工具栏进行富文本编辑"}]}]},{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"邀请团队成员一起协作"}]}]}]}]}`,
		Version: 1,
	}
	DB.Create(&welcomeDoc)

	log.Printf("[Seed] Default admin created: admin@kloggerx.com / admin")
}
