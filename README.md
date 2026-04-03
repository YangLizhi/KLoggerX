# KloggerX - 私有化云文档协作平台

<div align="center">

![Version](https://img.shields.io/badge/version-1.0.0-blue.svg)
![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)
![Vue](https://img.shields.io/badge/Vue-3.5+-4FC08D?logo=vue.js)
![License](https://img.shields.io/badge/license-MIT-green.svg)

**KloggerX 在线知识协作和管理平台**

[功能特性](#功能特性) • [快速开始](#快速开始) • [技术架构](#技术架构) • [配置说明](#配置说明) • [贡献指南](#贡献指南)

</div>

---

## 项目简介

KloggerX 是一款可私有化部署的在线知识协作和管理平台，支持多人实时协作、多格式文档编辑、知识库管理、远程存储接入等企业级功能，满足团队私有化部署和数据安全需求。

### 核心亮点

- **私有化部署**：完全自主可控，数据存储在自有服务器
- **实时协作**：基于 Yjs + WebSocket 的多人协同编辑
- **多格式支持**：富文本文档、表格、幻灯片、思维导图、代码文件
- **远程存储**：支持 SFTP、FTP、SMB、WebDAV 等协议接入
- **AI 能力**：集成知识库 AI 问答功能

---

## 功能特性

### 文档管理

| 功能 | 描述 |
|------|------|
| 文档库 | 树形目录管理，支持文件夹、标签、收藏、置顶 |
| 文档类型 | 富文本文档、表格、幻灯片、思维导图、多维表格、问卷、代码 |
| 文件预览 | PDF、Office 文档在线预览，图片预览 |
| 版本管理 | 文档版本历史，支持回滚 |
| 权限控制 | 文档级权限管理，支持分享、协作 |

### 在线编辑

#### 富文本文档
- 基于 Tiptap/ProseMirror 的块级编辑器
- 支持标题、列表、表格、代码块、图片、链接等
- 支持任务列表、高亮、下划线等富文本格式
- 实时协作编辑

#### 表格编辑
- 完整的电子表格功能
- 支持公式、格式设置、合并单元格
- 多工作表管理
- 数据排序、筛选

#### 幻灯片编辑
- 幻灯片创建、复制、删除、排序
- 主题切换、背景设置
- 切换动画、入场动画
- 幻灯片放映模式

#### 思维导图
- 节点创建、编辑、删除
- 拖拽布局
- 多种主题样式

### 远程存储

支持多种远程存储协议接入：

| 协议 | 功能 |
|------|------|
| SFTP | SSH 文件传输，支持密钥认证 |
| FTP | 文件传输协议，支持匿名/用户认证 |
| SMB | Windows 共享文件夹，支持域认证 |
| WebDAV | Web 分布式创作和版本控制 |

### AI 能力

- 基于 Qdrant 向量数据库的知识库
- 文档智能问答
- 知识检索与推荐

---

## 技术架构

### 前端技术栈

| 技术 | 版本 | 说明 |
|------|------|------|
| Vue | 3.5+ | 核心框架 |
| Vite | 8.0+ | 构建工具 |
| TypeScript | 5.9+ | 类型支持 |
| Element Plus | 2.13+ | UI 组件库 |
| Pinia | 3.0+ | 状态管理 |
| Vue Router | 4.6+ | 路由管理 |
| Tiptap | 3.20+ | 富文本编辑器 |
| Yjs | 13.6+ | 实时协作 |
| Axios | 1.13+ | HTTP 客户端 |

### 后端技术栈

| 技术 | 版本 | 说明 |
|------|------|------|
| Go | 1.22+ | 核心语言 |
| Gin | 1.9+ | Web 框架 |
| GORM | 1.25+ | ORM 框架 |
| MySQL | 8.0+ | 主数据库 |
| Redis | 7.0+ | 缓存/会话 |
| MinIO | - | 对象存储 |
| WebSocket | - | 实时通信 |
| JWT | - | 身份认证 |
| Qdrant | - | 向量数据库 |

### 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                         前端 (Vue3)                          │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐           │
│  │ 文档库  │ │  编辑器  │ │  云盘   │ │  设置   │           │
│  └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘           │
└───────┼──────────┼──────────┼──────────┼───────────────────┘
        │          │          │          │
        └──────────┴──────────┴──────────┘
                        │
                   HTTP/WebSocket
                        │
┌───────────────────────┴─────────────────────────────────────┐
│                     后端 (Go + Gin)                          │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐           │
│  │ 用户    │ │ 文档    │ │ 存储    │ │ AI     │           │
│  │ 服务    │ │ 服务    │ │ 服务    │ │ 服务    │           │
│  └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘           │
└───────┼──────────┼──────────┼──────────┼───────────────────┘
        │          │          │          │
   ┌────┴────┐ ┌───┴───┐ ┌────┴────┐ ┌────┴────┐
   │  MySQL  │ │ Redis │ │  MinIO  │ │ Qdrant  │
   └─────────┘ └───────┘ └─────────┘ └─────────┘
```

---

## 项目结构

```
iKloggerX/
├── kloggerx-server/          # 后端服务
│   ├── api/                  # API 接口
│   │   └── v1/               # v1 版本接口
│   ├── cmd/                  # 程序入口
│   ├── config/               # 配置文件
│   ├── deploy/               # 部署脚本
│   ├── internal/             # 内部模块
│   │   ├── model/            # 数据模型
│   │   ├── service/          # 业务逻辑
│   │   ├── middleware/       # 中间件
│   │   └── utils/            # 工具函数
│   ├── Dockerfile            # Docker 构建文件
│   └── go.mod                # Go 依赖管理
│
├── kloggerx-web/             # 前端应用
│   ├── src/
│   │   ├── api/              # API 接口封装
│   │   ├── assets/           # 静态资源
│   │   ├── components/       # 公共组件
│   │   │   ├── editor/       # 编辑器组件
│   │   │   ├── permission/   # 权限组件
│   │   │   └── share/        # 分享组件
│   │   ├── composables/      # 组合式函数
│   │   ├── hooks/            # 自定义 Hooks
│   │   ├── router/           # 路由配置
│   │   ├── store/            # 状态管理
│   │   ├── types/            # TypeScript 类型
│   │   ├── utils/            # 工具函数
│   │   └── views/            # 页面组件
│   │       ├── admin/        # 管理后台
│   │       ├── document/     # 文档相关
│   │       ├── home/         # 首页
│   │       └── user/         # 用户相关
│   ├── package.json          # NPM 依赖
│   └── vite.config.ts        # Vite 配置
│
├── docker-compose.yml        # Docker Compose 配置
└── README.md                 # 项目说明
```

---

## 快速开始

### 环境要求

- Go 1.22+
- Node.js 18+
- MySQL 8.0+
- Redis 7.0+
- MinIO (可选，用于对象存储)

### 方式一：Docker Compose 部署（推荐）

```bash
# 克隆项目
git clone https://github.com/your-username/iKloggerX.git
cd iKloggerX

# 一键启动
docker-compose up -d

# 访问应用
# 前端：http://localhost
# 后端：http://localhost:8178
# MinIO 控制台：http://localhost:9001
```

### 方式二：本地开发

#### 1. 克隆项目

```bash
git clone https://github.com/your-username/iKloggerX.git
cd iKloggerX
```

#### 2. 配置数据库

```sql
-- 创建数据库
CREATE DATABASE kloggerx CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

#### 3. 配置后端

```bash
cd kloggerx-server

# 修改配置文件
cp config/config.yaml.example config/config.yaml
# 编辑 config.yaml，配置数据库、Redis 等连接信息

# 安装依赖
go mod download

# 运行服务
go run cmd/main.go
```

#### 4. 配置前端

```bash
cd kloggerx-web

# 安装依赖
npm install

# 开发模式运行
npm run dev

# 生产构建
npm run build
```

#### 5. 访问应用

- 前端开发服务：http://localhost:5173
- 后端 API 服务：http://localhost:8178

---

## 配置说明

### 后端配置 (config.yaml)

```yaml
# 服务配置
server:
  port: 8178
  mode: debug    # debug/release

# 数据库配置
database:
  host: 127.0.0.1
  port: 3306
  user: root
  password: your_password
  dbname: kloggerx
  charset: utf8mb4

# Redis 配置
redis:
  addr: 127.0.0.1:6379
  password: ""
  db: 0

# JWT 配置
jwt:
  secret: your-secret-key-change-in-production
  expire: 168h

# MinIO 配置
minio:
  endpoint: 127.0.0.1:9000
  access_key: minioadmin
  secret_key: minioadmin
  bucket: kloggerx
  use_ssl: false

# OnlyOffice 配置（可选）
onlyoffice:
  server_url: "http://your-office-server:8082"
  jwt_secret: "your-jwt-secret"

# Qdrant 配置（AI 功能）
qdrant:
  host: localhost
  port: 6333
  collection: kloggerx_chunks
```

### 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `SERVER_PORT` | 服务端口 | 8178 |
| `DB_HOST` | 数据库地址 | 127.0.0.1 |
| `DB_PORT` | 数据库端口 | 3306 |
| `DB_USER` | 数据库用户 | root |
| `DB_PASSWORD` | 数据库密码 | - |
| `DB_NAME` | 数据库名 | kloggerx |
| `REDIS_ADDR` | Redis 地址 | 127.0.0.1:6379 |
| `JWT_SECRET` | JWT 密钥 | - |
| `MINIO_ENDPOINT` | MinIO 地址 | 127.0.0.1:9000 |

---

## API 文档

### 认证相关

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/auth/login` | 用户登录 |
| POST | `/api/v1/auth/register` | 用户注册 |
| POST | `/api/v1/auth/logout` | 用户登出 |
| GET | `/api/v1/auth/profile` | 获取用户信息 |

### 文档相关

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/documents` | 获取文档列表 |
| GET | `/api/v1/documents/:id` | 获取文档详情 |
| POST | `/api/v1/documents` | 创建文档 |
| PUT | `/api/v1/documents/:id` | 更新文档 |
| DELETE | `/api/v1/documents/:id` | 删除文档 |
| GET | `/api/v1/documents/:id/versions` | 获取版本历史 |

### 远程存储

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/remote-storages` | 获取存储列表 |
| POST | `/api/v1/remote-storages` | 添加存储配置 |
| POST | `/api/v1/remote-storages/:id/test` | 测试连接 |
| POST | `/api/v1/remote-storages/:id/connect` | 连接存储 |
| POST | `/api/v1/remote-storages/:id/disconnect` | 断开连接 |
| GET | `/api/v1/remote-storages/:id/files` | 获取文件列表 |

---

## 开发指南

### 代码规范

- 前端：遵循 Vue 3 Composition API 风格，使用 TypeScript
- 后端：遵循 Go 官方代码规范，使用 golangci-lint
- 提交：遵循 Conventional Commits 规范

### 分支管理

- `main`：主分支，稳定版本
- `develop`：开发分支
- `feature/*`：功能分支
- `fix/*`：修复分支

### 开发流程

1. Fork 本仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'feat: add amazing feature'`)
4. 推送分支 (`git push origin feature/amazing-feature`)
5. 提交 Pull Request

---

## 路线图

### v1.0 (当前)

- [x] 文档基础管理
- [x] 富文本编辑器
- [x] 表格编辑器
- [x] 幻灯片编辑器
- [x] 思维导图编辑器
- [x] 远程存储接入 (SFTP/FTP/SMB)
- [x] 文件预览 (PDF/Office)
- [x] 用户认证与权限

### v1.1 (计划中)

- [ ] 实时协作增强
- [ ] 评论与批注
- [ ] 文档模板库
- [ ] 全文搜索优化

### v1.2 (规划中)

- [ ] 移动端适配
- [ ] 离线编辑
- [ ] 国际化支持
- [ ] 插件系统

---

## 贡献指南

我们欢迎所有形式的贡献！

### 如何贡献

1. 提交 Issue 报告 Bug 或提出功能建议
2. Fork 项目并提交 Pull Request
3. 完善文档或翻译

### 行为准则

请阅读并遵守我们的 [行为准则](CODE_OF_CONDUCT.md)。

---

## 常见问题

**Q: 如何修改默认端口？**

A: 修改 `kloggerx-server/config/config.yaml` 中的 `server.port` 配置。

**Q: 如何配置 HTTPS？**

A: 建议使用 Nginx 反向代理并配置 SSL 证书。

**Q: 数据库迁移如何执行？**

A: 项目使用 GORM 自动迁移，启动服务时会自动创建表结构。

**Q: 如何对接 OnlyOffice？**

A: 在配置文件中配置 OnlyOffice 服务器地址和 JWT 密钥即可。

---

## 许可证

本项目基于 [MIT](LICENSE) 许可证开源。

---

## 致谢

感谢以下开源项目：

- [Vue.js](https://vuejs.org/)
- [Element Plus](https://element-plus.org/)
- [Tiptap](https://tiptap.dev/)
- [Gin](https://gin-gonic.com/)
- [GORM](https://gorm.io/)

---

<div align="center">

**如果这个项目对你有帮助，请给一个 ⭐️ Star！**

**捐赠请联系enluoo@163.com, 您的慷慨将会助力该项目的持续发展**

Made with ❤️ by KloggerX Team

</div>
