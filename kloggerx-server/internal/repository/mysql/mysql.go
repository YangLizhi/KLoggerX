package mysql

import (
	"fmt"
	"log"
	"os"
	"time"

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
		Logger: gormlogger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			gormlogger.Config{
				SlowThreshold: 200 * time.Millisecond,
				LogLevel:      gormlogger.Warn,
				Colorful:      true,
			},
		),
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
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)
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
		&model.KnowledgeGraph{},
		&model.Template{},
		&model.TemplateFavorite{},
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
		// Knowledge AI assistant models
		&model.KbConversation{},
		&model.KbMessage{},
		&model.KbFeedback{},
		&model.KbShareLink{},
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
	upgradeBuiltinTemplates()
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

// builtinTemplateContent defines rich Tiptap JSON content for each doc-type builtin template.
var builtinTemplateContent = map[string]string{
	"会议记录": `{"type":"doc","content":[` +
		`{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"会议记录"}]},` +
		`{"type":"paragraph","content":[{"type":"text","text":"会议日期：____年__月__日  时间：__:__ ~ __:__"}]},` +
		`{"type":"paragraph","content":[{"type":"text","text":"会议地点："}]},` +
		`{"type":"paragraph","content":[{"type":"text","text":"主持人："}]},` +
		`{"type":"paragraph","content":[{"type":"text","text":"参会人员："}]},` +
		`{"type":"paragraph","content":[{"type":"text","text":"记录人："}]},` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"会议议题"}]},` +
		`{"type":"bulletList","content":[` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"议题一：请在此填写第一个讨论议题"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"议题二：请在此填写第二个讨论议题"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"议题三：请在此填写第三个讨论议题"}]}]}` +
		`]},` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"讨论要点"}]},` +
		`{"type":"bulletList","content":[` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"关于议题一的讨论结论和关键意见"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"关于议题二的讨论结论和关键意见"}]}]}` +
		`]},` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"决议事项"}]},` +
		`{"type":"bulletList","content":[` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"决议一：经讨论决定..."}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"决议二：经讨论决定..."}]}]}` +
		`]},` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"行动项"}]},` +
		`{"type":"taskList","content":[` +
			`{"type":"taskItem","attrs":{"checked":false},"content":[{"type":"paragraph","content":[{"type":"text","text":"【负责人A】完成XX任务 — 截止日期：__月__日"}]}]},` +
			`{"type":"taskItem","attrs":{"checked":false},"content":[{"type":"paragraph","content":[{"type":"text","text":"【负责人B】跟进YY事项 — 截止日期：__月__日"}]}]},` +
			`{"type":"taskItem","attrs":{"checked":false},"content":[{"type":"paragraph","content":[{"type":"text","text":"【负责人C】确认ZZ方案 — 截止日期：__月__日"}]}]}` +
		`]},` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"下次会议安排"}]},` +
		`{"type":"paragraph","content":[{"type":"text","text":"预计时间：____年__月__日  议题预告："}]}` +
		`]}`,

	"周报": `{"type":"doc","content":[` +
		`{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"周报"}]},` +
		`{"type":"paragraph","content":[{"type":"text","text":"姓名：________  部门：________  日期：____年__月__日 ~ __月__日"}]},` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"本周完成"}]},` +
		`{"type":"bulletList","content":[` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"完成了XX模块的开发与自测，已提交代码审查"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"参与了YY需求评审会议，明确了技术方案"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"修复了ZZ线上问题，已发布修复版本"}]}]}` +
		`]},` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"下周计划"}]},` +
		`{"type":"bulletList","content":[` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"继续推进XX功能的联调与集成测试"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"启动YY模块的技术方案设计"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"完成ZZ文档的编写与评审"}]}]}` +
		`]},` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"需要协助"}]},` +
		`{"type":"bulletList","content":[` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"需要产品同学确认XX交互细节"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"需要运维协助配置YY测试环境"}]}]}` +
		`]},` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"风险与问题"}]},` +
		`{"type":"bulletList","content":[` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"XX依赖的第三方接口响应较慢，可能影响上线时间"}]}]}` +
		`]}` +
		`]}`,

	"OKR 目标": `{"type":"doc","content":[` +
		`{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"OKR 目标"}]},` +
		`{"type":"paragraph","content":[{"type":"text","text":"周期：____年 Q__（__月 ~ __月）"}]},` +
		`{"type":"paragraph","content":[{"type":"text","text":"制定人：________  部门：________"}]},` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"Objective 1：提升产品核心体验"}]},` +
		`{"type":"paragraph","content":[{"type":"text","text":"描述：聚焦用户反馈最集中的痛点，系统性优化产品核心使用流程"}]},` +
		`{"type":"bulletList","content":[` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"KR1：用户满意度评分从 X 分提升至 Y 分（进度：__/%）"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"KR2：核心功能操作步骤减少 30%（进度：__/%）"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"KR3：页面加载时间 P95 降低至 2 秒以内（进度：__/%）"}]}]}` +
		`]},` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"Objective 2：建设高效协作团队"}]},` +
		`{"type":"paragraph","content":[{"type":"text","text":"描述：通过流程优化与工具建设，提升团队整体交付效率"}]},` +
		`{"type":"bulletList","content":[` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"KR1：Sprint 交付准时率达到 90% 以上（进度：__/%）"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"KR2：代码审查覆盖率达到 100%（进度：__/%）"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"KR3：线上故障平均恢复时间缩短至 30 分钟以内（进度：__/%）"}]}]}` +
		`]},` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"Objective 3：拓展业务增长空间"}]},` +
		`{"type":"paragraph","content":[{"type":"text","text":"描述：探索新场景与新渠道，为业务增长注入新动力"}]},` +
		`{"type":"bulletList","content":[` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"KR1：新增合作渠道 X 个（进度：__/%）"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"KR2：月活跃用户增长 20%（进度：__/%）"}]}]}` +
		`]}` +
		`]}`,

	"项目计划": `{"type":"doc","content":[` +
		`{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"项目计划"}]},` +
		`{"type":"paragraph","content":[{"type":"text","text":"项目名称：________  负责人：________  启动日期：____年__月__日"}]},` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"项目背景"}]},` +
		`{"type":"paragraph","content":[{"type":"text","text":"简要描述项目发起的背景与动因，说明为什么需要启动该项目，以及期望解决的核心问题。"}]},` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"项目目标"}]},` +
		`{"type":"bulletList","content":[` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"目标一：明确的、可衡量的业务/技术目标"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"目标二：量化的预期成果指标"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"目标三：关键交付物及质量标准"}]}]}` +
		`]},` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"项目范围"}]},` +
		`{"type":"paragraph","content":[{"type":"text","text":"明确项目包含和不包含的内容，界定边界，避免范围蔓延。"}]},` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"里程碑计划"}]},` +
		`{"type":"bulletList","content":[` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"M1 - 需求确认与方案评审（__月__日）"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"M2 - 开发完成与内部测试（__月__日）"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"M3 - UAT验收与上线准备（__月__日）"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"M4 - 正式发布与项目复盘（__月__日）"}]}]}` +
		`]},` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"团队分工"}]},` +
		`{"type":"bulletList","content":[` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"产品经理：负责需求分析与验收"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"技术负责人：负责架构设计与技术评审"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"开发工程师：负责编码实现与单元测试"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"测试工程师：负责测试方案设计与执行"}]}]}` +
		`]},` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"风险评估"}]},` +
		`{"type":"bulletList","content":[` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"风险一：需求变更频繁 → 应对措施：建立变更审批流程"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"风险二：关键人员请假 → 应对措施：交叉培训与备份机制"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"风险三：第三方依赖延迟 → 应对措施：提前沟通并预留缓冲时间"}]}]}` +
		`]}` +
		`]}`,

	"需求文档": `{"type":"doc","content":[` +
		`{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"需求文档"}]},` +
		`{"type":"paragraph","content":[{"type":"text","text":"产品名称：________  版本：V__.__  编写日期：____年__月__日"}]},` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"产品概述"}]},` +
		`{"type":"paragraph","content":[{"type":"text","text":"简要描述产品的定位、目标用户群体和核心价值主张。说明该需求所属的产品模块以及与其他模块的关系。"}]},` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"用户故事"}]},` +
		`{"type":"bulletList","content":[` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"作为【普通用户】，我希望能够【快速搜索文档】，以便【在大量文档中迅速找到所需内容】"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"作为【团队管理者】，我希望能够【查看团队成员的文档活动】，以便【了解团队工作进展】"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"作为【系统管理员】，我希望能够【批量管理用户权限】，以便【高效维护系统安全】"}]}]}` +
		`]},` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"功能需求"}]},` +
		`{"type":"bulletList","content":[` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"FR-001：支持按关键词、标签、创建时间等多维度搜索文档"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"FR-002：搜索结果支持高亮显示匹配关键词"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"FR-003：支持搜索结果按相关度、时间排序"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"FR-004：支持保存常用搜索条件"}]}]}` +
		`]},` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"非功能需求"}]},` +
		`{"type":"bulletList","content":[` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"性能：搜索响应时间 P95 ≤ 500ms"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"安全：搜索结果严格遵循权限控制，用户只能看到有权访问的文档"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"兼容：支持主流浏览器 Chrome、Firefox、Safari、Edge 最新两个版本"}]}]}` +
		`]},` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"验收标准"}]},` +
		`{"type":"bulletList","content":[` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"所有功能需求通过测试用例验证"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"性能指标达标，经压力测试确认"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"用户体验走查通过，无阻断性问题"}]}]}` +
		`]}` +
		`]}`,

	"技术方案": `{"type":"doc","content":[` +
		`{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"技术方案"}]},` +
		`{"type":"paragraph","content":[{"type":"text","text":"方案名称：________  编写人：________  日期：____年__月__日"}]},` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"背景与目标"}]},` +
		`{"type":"paragraph","content":[{"type":"text","text":"描述技术改进的背景，当前系统存在的问题或面临的挑战，以及本次技术方案期望达成的目标。"}]},` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"方案设计"}]},` +
		`{"type":"paragraph","content":[{"type":"text","text":"阐述整体架构设计思路，包括系统模块划分、核心流程、数据流向等。建议配合架构图说明。"}]},` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"技术选型"}]},` +
		`{"type":"bulletList","content":[` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"编程语言：Go / Java / Python（选型理由）"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"Web 框架：Gin / Spring Boot / FastAPI（选型理由）"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"数据库：MySQL / PostgreSQL / MongoDB（选型理由）"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"缓存：Redis / Memcached（选型理由）"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"消息队列：Kafka / RabbitMQ（选型理由）"}]}]}` +
		`]},` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"接口设计"}]},` +
		`{"type":"paragraph","content":[{"type":"text","text":"列出核心 API 接口定义，包括请求方法、路径、参数和返回值。"}]},` +
		`{"type":"bulletList","content":[` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"GET /api/v1/resource - 获取资源列表"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"POST /api/v1/resource - 创建新资源"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"PUT /api/v1/resource/:id - 更新指定资源"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"DELETE /api/v1/resource/:id - 删除指定资源"}]}]}` +
		`]},` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"数据模型"}]},` +
		`{"type":"paragraph","content":[{"type":"text","text":"描述核心数据表结构、字段定义、索引策略和表间关系。"}]},` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"部署方案"}]},` +
		`{"type":"paragraph","content":[{"type":"text","text":"说明部署架构、环境要求、CI/CD 流水线配置、监控告警策略等。"}]},` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"风险与备选方案"}]},` +
		`{"type":"bulletList","content":[` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"风险一：XX组件可能存在性能瓶颈 → 备选：采用YY方案替代"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"风险二：第三方服务 SLA 不可控 → 备选：增加降级与熔断机制"}]}]}` +
		`]}` +
		`]}`,

	"读书笔记": `{"type":"doc","content":[` +
		`{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"读书笔记"}]},` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"书籍信息"}]},` +
		`{"type":"bulletList","content":[` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"书名：《________》"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"作者：________"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"出版日期：____年__月"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"阅读日期：____年__月__日 ~ __月__日"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"推荐指数：★★★★☆"}]}]}` +
		`]},` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"内容概要"}]},` +
		`{"type":"paragraph","content":[{"type":"text","text":"用几段话概括全书的主要内容和结构，让读者快速了解本书讲了什么。"}]},` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"核心观点"}]},` +
		`{"type":"bulletList","content":[` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"观点一：作者提出的第一个核心论点及其论证逻辑"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"观点二：作者提出的第二个核心论点及其论证逻辑"}]}]},` +
			`{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"观点三：作者提出的第三个核心论点及其论证逻辑"}]}]}` +
		`]},` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"精彩摘录"}]},` +
		`{"type":"blockquote","content":[{"type":"paragraph","content":[{"type":"text","text":"在此记录书中令你印象深刻的原文段落，注明页码方便日后查阅。"}]}]},` +
		`{"type":"blockquote","content":[{"type":"paragraph","content":[{"type":"text","text":"再记录一段精彩内容..."}]}]},` +
		`{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"个人感悟"}]},` +
		`{"type":"paragraph","content":[{"type":"text","text":"写下你的阅读感受和思考。这本书哪些内容与你的经历产生了共鸣？它改变了你对某些事物的看法吗？你打算如何将书中的理念应用到工作或生活中？"}]}` +
		`]}`,
}

// builtinTemplatePreviews defines preview text for each doc-type builtin template.
var builtinTemplatePreviews = map[string]string{
	"会议记录": "会议记录 会议日期：____年__月__日 会议地点： 主持人： 参会人员： 记录人： 会议议题 议题一：请在此填写第一个讨论议题 议题二 议题三 讨论要点 决议事项 行动项 下次会议安排",
	"周报":   "周报 姓名：________ 部门：________ 本周完成 完成了XX模块的开发与自测 参与了YY需求评审会议 修复了ZZ线上问题 下周计划 继续推进XX功能的联调 启动YY模块的技术方案设计 需要协助 风险与问题",
	"OKR 目标": "OKR 目标 周期：____年 Q__ Objective 1：提升产品核心体验 KR1：用户满意度评分提升 KR2：核心功能操作步骤减少30% KR3：页面加载时间P95降低至2秒 Objective 2：建设高效协作团队 Objective 3：拓展业务增长空间",
	"项目计划": "项目计划 项目背景 简要描述项目发起的背景与动因 项目目标 里程碑计划 M1需求确认 M2开发完成 M3 UAT验收 M4正式发布 团队分工 风险评估 需求变更 关键人员 第三方依赖",
	"需求文档": "需求文档 产品概述 简要描述产品的定位、目标用户群体和核心价值主张 用户故事 作为普通用户我希望能够快速搜索文档 功能需求 FR-001多维度搜索 FR-002高亮显示 非功能需求 性能 安全 兼容 验收标准",
	"技术方案": "技术方案 背景与目标 描述技术改进的背景 方案设计 阐述整体架构设计思路 技术选型 编程语言 Web框架 数据库 缓存 消息队列 接口设计 数据模型 部署方案 风险与备选方案",
	"读书笔记": "读书笔记 书籍信息 书名 作者 出版日期 阅读日期 推荐指数 内容概要 用几段话概括全书的主要内容和结构 核心观点 精彩摘录 个人感悟 写下你的阅读感受和思考",
}

func seedTemplates() {
	var count int64
	DB.Model(&model.Template{}).Count(&count)
	if count > 0 {
		return
	}
	templates := []model.Template{
		{Name: "会议记录", Category: "工作", Type: "doc", IsBuiltin: true, Description: "记录会议议题、讨论结果与行动项",
			Content: builtinTemplateContent["会议记录"], Preview: builtinTemplatePreviews["会议记录"]},
		{Name: "周报", Category: "工作", Type: "doc", IsBuiltin: true, Description: "周工作汇报模板",
			Content: builtinTemplateContent["周报"], Preview: builtinTemplatePreviews["周报"]},
		{Name: "项目计划", Category: "项目", Type: "doc", IsBuiltin: true, Description: "项目启动与里程碑规划",
			Content: builtinTemplateContent["项目计划"], Preview: builtinTemplatePreviews["项目计划"]},
		{Name: "需求文档", Category: "产品", Type: "doc", IsBuiltin: true, Description: "产品需求规格说明书",
			Content: builtinTemplateContent["需求文档"], Preview: builtinTemplatePreviews["需求文档"]},
		{Name: "技术方案", Category: "技术", Type: "doc", IsBuiltin: true, Description: "技术设计与架构文档",
			Content: builtinTemplateContent["技术方案"], Preview: builtinTemplatePreviews["技术方案"]},
		{Name: "OKR 目标", Category: "工作", Type: "doc", IsBuiltin: true, Description: "季度OKR目标与关键结果",
			Content: builtinTemplateContent["OKR 目标"], Preview: builtinTemplatePreviews["OKR 目标"]},
		{Name: "空白表格", Category: "表格", Type: "sheet", IsBuiltin: true, Description: "从空白表格开始", Content: "{}"},
		{Name: "数据报表", Category: "表格", Type: "sheet", IsBuiltin: true, Description: "数据统计与分析", Content: "{}"},
		{Name: "项目演示", Category: "幻灯片", Type: "slide", IsBuiltin: true, Description: "项目汇报演示文稿", Content: "{}"},
		{Name: "个人介绍", Category: "幻灯片", Type: "slide", IsBuiltin: true, Description: "个人/团队介绍幻灯片", Content: "{}"},
		{Name: "思维导图", Category: "思维笔记", Type: "mindnote", IsBuiltin: true, Description: "知识梳理思维导图", Content: "{}"},
		{Name: "读书笔记", Category: "学习", Type: "doc", IsBuiltin: true, Description: "书籍阅读与知识整理",
			Content: builtinTemplateContent["读书笔记"], Preview: builtinTemplatePreviews["读书笔记"]},
	}
	for i := range templates {
		DB.Create(&templates[i])
	}
	log.Printf("[Seed] %d default templates created", len(templates))
}

// upgradeBuiltinTemplates updates existing builtin templates that have empty content.
// This ensures already-deployed databases get the rich template content.
func upgradeBuiltinTemplates() {
	var templates []model.Template
	if err := DB.Where("is_builtin = ? AND (content = '' OR content IS NULL)", true).Find(&templates).Error; err != nil {
		log.Printf("[Upgrade] Failed to query builtin templates: %v", err)
		return
	}
	if len(templates) == 0 {
		return
	}
	updated := 0
	for _, t := range templates {
		content, hasContent := builtinTemplateContent[t.Name]
		if !hasContent {
			continue
		}
		preview := builtinTemplatePreviews[t.Name]
		if err := DB.Model(&t).Updates(map[string]interface{}{
			"content": content,
			"preview": preview,
		}).Error; err != nil {
			log.Printf("[Upgrade] Failed to update template '%s': %v", t.Name, err)
		} else {
			updated++
		}
	}
	if updated > 0 {
		log.Printf("[Upgrade] Updated %d builtin templates with rich content", updated)
	}
}
