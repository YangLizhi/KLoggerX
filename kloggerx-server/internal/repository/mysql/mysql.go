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
		&model.KnowledgeSource{},
		&model.KnowledgeChunk{},
		&model.EmbeddingJob{},
		&model.RaptorNode{},
		&model.Template{},
		&model.Comment{},
		&model.Notification{},
		&model.RecentDocument{},
		&model.FileRecord{},
		&model.OperationLog{},
		&model.SystemSetting{},
		// New storage-related models
		&model.UserStorageSetting{},
		&model.RemoteStorage{},
		&model.StorageUsage{},
	)

	// Add FULLTEXT index on knowledge_chunks.content for efficient RAG retrieval
	DB.Exec("ALTER TABLE knowledge_chunks ADD FULLTEXT INDEX IF NOT EXISTS idx_chunk_content (content)")

	// Extend system_settings.value to LONGTEXT for large JSON data (AI model settings)
	if err := DB.Exec("ALTER TABLE system_settings MODIFY COLUMN `value` LONGTEXT").Error; err != nil {
		log.Printf("[AutoMigrate] Warning: Failed to alter system_settings.value to LONGTEXT: %v", err)
	} else {
		log.Println("[AutoMigrate] Successfully altered system_settings.value to LONGTEXT")
	}

	seedData()
	seedTemplates()
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

func seedTemplates() {
	var count int64
	DB.Model(&model.Template{}).Count(&count)
	if count > 0 {
		return
	}
	templates := []model.Template{
		{Name: "会议记录", Category: "工作", Type: "doc", IsBuiltin: true, Description: "记录会议议题、讨论结果与行动项",
			Content: `{"type":"doc","content":[{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"会议记录"}]},{"type":"paragraph","content":[{"type":"text","text":"会议时间："}]},{"type":"paragraph","content":[{"type":"text","text":"参会人员："}]},{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"议题"}]},{"type":"bulletList","content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"议题1"}]}]}]},{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"讨论内容"}]},{"type":"paragraph","content":[{"type":"text","text":""}]},{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"行动项"}]},{"type":"taskList","content":[{"type":"taskItem","attrs":{"checked":false},"content":[{"type":"paragraph","content":[{"type":"text","text":"负责人 - 截止日期"}]}]}]}]}`},
		{Name: "周报", Category: "工作", Type: "doc", IsBuiltin: true, Description: "周工作汇报模板",
			Content: `{"type":"doc","content":[{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"周报"}]},{"type":"paragraph","content":[{"type":"text","text":"姓名：  | 日期："}]},{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"本周完成"}]},{"type":"bulletList","content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":""}]}]}]},{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"下周计划"}]},{"type":"bulletList","content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":""}]}]}]},{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"问题与风险"}]},{"type":"paragraph","content":[{"type":"text","text":""}]}]}`},
		{Name: "项目计划", Category: "项目", Type: "doc", IsBuiltin: true, Description: "项目启动与里程碑规划",
			Content: `{"type":"doc","content":[{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"项目计划"}]},{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"项目背景"}]},{"type":"paragraph","content":[{"type":"text","text":""}]},{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"目标与范围"}]},{"type":"paragraph","content":[{"type":"text","text":""}]},{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"里程碑"}]},{"type":"bulletList","content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"阶段1："}]}]},{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"阶段2："}]}]}]},{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"风险管控"}]},{"type":"paragraph","content":[{"type":"text","text":""}]}]}`},
		{Name: "需求文档", Category: "产品", Type: "doc", IsBuiltin: true, Description: "产品需求规格说明书",
			Content: `{"type":"doc","content":[{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"需求文档"}]},{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"需求背景"}]},{"type":"paragraph","content":[{"type":"text","text":""}]},{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"用户故事"}]},{"type":"paragraph","content":[{"type":"text","text":"作为 [用户类型]，我希望 [功能]，以便 [价值]"}]},{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"功能需求"}]},{"type":"bulletList","content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":""}]}]}]},{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"非功能需求"}]},{"type":"paragraph","content":[{"type":"text","text":""}]},{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"验收标准"}]},{"type":"paragraph","content":[{"type":"text","text":""}]}]}`},
		{Name: "技术方案", Category: "技术", Type: "doc", IsBuiltin: true, Description: "技术设计与架构文档",
			Content: `{"type":"doc","content":[{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"技术方案"}]},{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"背景与目标"}]},{"type":"paragraph","content":[{"type":"text","text":""}]},{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"架构设计"}]},{"type":"paragraph","content":[{"type":"text","text":""}]},{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"接口设计"}]},{"type":"paragraph","content":[{"type":"text","text":""}]},{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"数据模型"}]},{"type":"paragraph","content":[{"type":"text","text":""}]},{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"部署方案"}]},{"type":"paragraph","content":[{"type":"text","text":""}]}]}`},
		{Name: "OKR 目标", Category: "工作", Type: "doc", IsBuiltin: true, Description: "季度OKR目标与关键结果",
			Content: `{"type":"doc","content":[{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"OKR 目标"}]},{"type":"paragraph","content":[{"type":"text","text":"周期：Q  年"}]},{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"Objective 1"}]},{"type":"paragraph","content":[{"type":"text","text":"目标描述"}]},{"type":"bulletList","content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"KR1: "}]}]},{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"KR2: "}]}]}]}]}`},
		{Name: "空白表格", Category: "表格", Type: "sheet", IsBuiltin: true, Description: "从空白表格开始", Content: "{}"},
		{Name: "数据报表", Category: "表格", Type: "sheet", IsBuiltin: true, Description: "数据统计与分析", Content: "{}"},
		{Name: "项目演示", Category: "幻灯片", Type: "slide", IsBuiltin: true, Description: "项目汇报演示文稿", Content: "{}"},
		{Name: "个人介绍", Category: "幻灯片", Type: "slide", IsBuiltin: true, Description: "个人/团队介绍幻灯片", Content: "{}"},
		{Name: "思维导图", Category: "思维笔记", Type: "mindnote", IsBuiltin: true, Description: "知识梳理思维导图", Content: "{}"},
		{Name: "读书笔记", Category: "学习", Type: "doc", IsBuiltin: true, Description: "书籍阅读与知识整理",
			Content: `{"type":"doc","content":[{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"读书笔记"}]},{"type":"paragraph","content":[{"type":"text","text":"书名：  | 作者：  | 评分："}]},{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"核心观点"}]},{"type":"paragraph","content":[{"type":"text","text":""}]},{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"精彩摘录"}]},{"type":"blockquote","content":[{"type":"paragraph","content":[{"type":"text","text":""}]}]},{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"我的思考"}]},{"type":"paragraph","content":[{"type":"text","text":""}]}]}`},
	}
	for i := range templates {
		DB.Create(&templates[i])
	}
	log.Printf("[Seed] %d default templates created", len(templates))
}
