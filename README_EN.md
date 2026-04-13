# KloggerX - Private Cloud Document Collaboration Platform

<div align="center">

![Version](https://img.shields.io/badge/version-1.1.0-blue.svg)
![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)
![Vue](https://img.shields.io/badge/Vue-3.5+-4FC08D?logo=vue.js)
![License](https://img.shields.io/badge/license-MIT-green.svg)

**KloggerX Online Knowledge Collaboration and Management Platform**

[Features](#features) • [Quick Start](#quick-start) • [Architecture](#architecture) • [Configuration](#configuration) • [Contributing](#contributing)

</div>

---

## Introduction

KloggerX is a self-hosted online knowledge collaboration and management platform that supports multi-user real-time collaboration, multi-format document editing, knowledge base management, and remote storage integration. It meets enterprise needs for private deployment and data security.

### Key Highlights

- **Private Deployment**: Fully self-controlled, data stored on your own servers
- **Real-time Collaboration**: Multi-user collaborative editing based on Yjs + WebSocket
- **Multi-format Support**: Rich text documents, spreadsheets, presentations, mind maps, code files
- **Remote Storage**: Support for SFTP, FTP, SMB, WebDAV, NFS protocols
- **AI Capabilities**: Integrated knowledge base AI Q&A functionality
- **Comments & Annotations**: Document-level comments and inline annotation features
- **Document Template Library**: Built-in template library for quickly creating standard documents
- **Collaborator Awareness**: Real-time display of online collaborators and join/leave notifications

---

## Features

### Document Management

| Feature | Description |
|---------|-------------|
| Document Library | Tree-based directory management with folders, tags, favorites, and pinning |
| Document Types | Rich text, spreadsheets, presentations, mind maps, multi-dimensional tables, surveys, code |
| File Preview | Online preview for PDF, Office documents, and images |
| Version Management | Document version history with rollback support |
| Permission Control | Document-level permission management with sharing and collaboration |

### Online Editing

#### Rich Text Documents
- Block-based editor powered by Tiptap/ProseMirror
- Support for headings, lists, tables, code blocks, images, links, etc.
- Rich text formatting including task lists, highlights, underlines
- Real-time collaborative editing

#### Spreadsheet Editing
- Professional spreadsheet engine based on Univer Sheet
- Excel-level formula engine
- Complete formatting toolbar
- Multi-sheet management

#### Presentation Editing
- Slide creation, duplication, deletion, and reordering
- Theme switching and background settings
- Transition and entrance animations
- Presentation mode

#### Mind Maps
- Node creation, editing, and deletion
- Drag-and-drop layout
- Multiple theme styles

### Comments & Annotations

- Inline comment markers
- Comment panel (sidebar format)
- Comment bubbles (floating markers)

### Document Template Library

- Built-in template management
- Template preview
- Quick document creation from templates
- Save documents as templates

### Collaboration Management

- Multi-user collaborator online status display
- Real-time user join/leave notifications
- WebSocket real-time communication

### Remote Storage

Support for multiple remote storage protocols:

| Protocol | Function |
|----------|----------|
| SFTP | SSH file transfer with key authentication |
| FTP | File Transfer Protocol with anonymous/user authentication |
| SMB | Windows shared folders with domain authentication |
| WebDAV | Web Distributed Authoring and Versioning |
| NFS | Network File System |

### Cloud Storage

- Alibaba Cloud OSS
- Tencent Cloud COS
- Baidu Cloud BOS

### AI Capabilities

- Knowledge base powered by Qdrant vector database
- Intelligent document Q&A
- Knowledge retrieval and recommendations

---

## Architecture

### Frontend Tech Stack

| Technology | Version | Description |
|------------|---------|-------------|
| Vue | 3.5+ | Core framework |
| Vite | 8.0+ | Build tool |
| TypeScript | 5.9+ | Type support |
| Element Plus | 2.13+ | UI component library |
| Pinia | 3.0+ | State management |
| Vue Router | 4.6+ | Routing |
| Tiptap | 3.20+ | Rich text editor |
| Univer Sheet | 0.20+ | Spreadsheet engine |
| Yjs | 13.6+ | Real-time collaboration |
| Axios | 1.13+ | HTTP client |

### Backend Tech Stack

| Technology | Version | Description |
|------------|---------|-------------|
| Go | 1.22+ | Core language |
| Gin | 1.9+ | Web framework |
| GORM | 1.25+ | ORM framework |
| MySQL | 8.0+ | Primary database |
| Redis | 7.0+ | Cache/session |
| MinIO | - | Object storage |
| WebSocket | - | Real-time communication |
| JWT | - | Authentication |
| Qdrant | - | Vector database |

### System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Frontend (Vue3)                         │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐           │
│  │Document │ │ Editor  │ │  Cloud  │ │Settings │           │
│  │ Library │ │         │ │  Drive  │ │         │           │
│  └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘           │
└───────┼──────────┼──────────┼──────────┼───────────────────┘
        │          │          │          │
        └──────────┴──────────┴──────────┘
                        │
                   HTTP/WebSocket
                        │
┌───────────────────────┴─────────────────────────────────────┐
│                     Backend (Go + Gin)                       │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐           │
│  │  User   │ │Document │ │ Storage │ │   AI    │           │
│  │ Service │ │ Service │ │ Service │ │ Service │           │
│  └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘           │
└───────┼──────────┼──────────┼──────────┼───────────────────┘
        │          │          │          │
   ┌────┴────┐ ┌───┴───┐ ┌────┴────┐ ┌────┴────┐
   │  MySQL  │ │ Redis │ │  MinIO  │ │ Qdrant  │
   └─────────┘ └───────┘ └─────────┘ └─────────┘
```

---

## Project Structure

```
iKloggerX/
├── kloggerx-server/          # Backend service
│   ├── api/                  # API endpoints
│   │   └── v1/               # v1 API version
│   ├── cmd/                  # Application entry
│   ├── config/               # Configuration files
│   ├── deploy/               # Deployment scripts
│   ├── internal/             # Internal modules
│   │   ├── model/            # Data models
│   │   ├── service/          # Business logic
│   │   ├── middleware/       # Middleware
│   │   └── utils/            # Utility functions
│   ├── Dockerfile            # Docker build file
│   └── go.mod                # Go dependency management
│
├── kloggerx-web/             # Frontend application
│   ├── src/
│   │   ├── api/              # API client
│   │   ├── assets/           # Static assets
│   │   ├── components/       # Common components
│   │   │   ├── editor/       # Editor components
│   │   │   ├── permission/   # Permission components
│   │   │   └── share/        # Sharing components
│   │   ├── composables/      # Composable functions
│   │   ├── hooks/            # Custom hooks
│   │   ├── router/           # Router configuration
│   │   ├── store/            # State management
│   │   ├── types/            # TypeScript types
│   │   ├── utils/            # Utility functions
│   │   └── views/            # Page components
│   │       ├── admin/        # Admin dashboard
│   │       ├── document/     # Document related
│   │       ├── home/         # Home page
│   │       └── user/         # User related
│   ├── package.json          # NPM dependencies
│   └── vite.config.ts        # Vite configuration
│
├── docker-compose.yml        # Docker Compose configuration
└── README.md                 # Project documentation
```

---

## Quick Start

### Requirements

- Go 1.22+
- Node.js 18+
- MySQL 8.0+
- Redis 7.0+
- MinIO (optional, for object storage)

### Option 1: Docker Compose Deployment (Recommended)

```bash
# Clone the project
git clone https://github.com/your-username/iKloggerX.git
cd iKloggerX

# Start with one command
docker-compose up -d

# Access the application
# Frontend: http://localhost
# Backend: http://localhost:8178
# MinIO Console: http://localhost:9001
```

### Option 2: Local Development

#### 1. Clone the Project

```bash
git clone https://github.com/your-username/iKloggerX.git
cd iKloggerX
```

#### 2. Configure Database

```sql
-- Create database
CREATE DATABASE kloggerx CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

#### 3. Configure Backend

```bash
cd kloggerx-server

# Copy configuration file
cp config/config.yaml.example config/config.yaml
# Edit config.yaml with database, Redis connection info

# Install dependencies
go mod download

# Run the service
go run cmd/main.go
```

#### 4. Configure Frontend

```bash
cd kloggerx-web

# Install dependencies
npm install

# Run in development mode
npm run dev

# Production build
npm run build
```

#### 5. Access the Application

- Frontend dev server: http://localhost:5178
- Backend API server: http://localhost:8178

---

## Configuration

### Backend Configuration (config.yaml)

```yaml
# Server configuration
server:
  port: 8178
  mode: debug    # debug/release

# Database configuration
database:
  host: 127.0.0.1
  port: 3306
  user: root
  password: your_password
  dbname: kloggerx
  charset: utf8mb4

# Redis configuration
redis:
  addr: 127.0.0.1:6379
  password: ""
  db: 0

# JWT configuration
jwt:
  secret: your-secret-key-change-in-production
  expire: 168h

# MinIO configuration
minio:
  endpoint: 127.0.0.1:9000
  access_key: minioadmin
  secret_key: minioadmin
  bucket: kloggerx
  use_ssl: false

# OnlyOffice configuration (optional)
onlyoffice:
  server_url: "http://your-office-server:8082"
  jwt_secret: "your-jwt-secret"

# Qdrant configuration (AI features)
qdrant:
  host: localhost
  port: 6333
  collection: kloggerx_chunks
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | Server port | 8178 |
| `DB_HOST` | Database host | 127.0.0.1 |
| `DB_PORT` | Database port | 3306 |
| `DB_USER` | Database user | root |
| `DB_PASSWORD` | Database password | - |
| `DB_NAME` | Database name | kloggerx |
| `REDIS_ADDR` | Redis address | 127.0.0.1:6379 |
| `JWT_SECRET` | JWT secret | - |
| `MINIO_ENDPOINT` | MinIO endpoint | 127.0.0.1:9000 |

---

## API Documentation

### Authentication

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/auth/login` | User login |
| POST | `/api/v1/auth/register` | User registration |
| POST | `/api/v1/auth/logout` | User logout |
| GET | `/api/v1/auth/profile` | Get user info |

### Documents

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/documents` | Get document list |
| GET | `/api/v1/documents/:id` | Get document details |
| POST | `/api/v1/documents` | Create document |
| PUT | `/api/v1/documents/:id` | Update document |
| DELETE | `/api/v1/documents/:id` | Delete document |
| GET | `/api/v1/documents/:id/versions` | Get version history |

### Remote Storage

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/remote-storages` | Get storage list |
| POST | `/api/v1/remote-storages` | Add storage configuration |
| POST | `/api/v1/remote-storages/:id/test` | Test connection |
| POST | `/api/v1/remote-storages/:id/connect` | Connect storage |
| POST | `/api/v1/remote-storages/:id/disconnect` | Disconnect storage |
| GET | `/api/v1/remote-storages/:id/files` | Get file list |

---

## Development Guide

### Code Standards

- Frontend: Follow Vue 3 Composition API style with TypeScript
- Backend: Follow Go official code standards with golangci-lint
- Commits: Follow Conventional Commits specification

### Branch Management

- `main`: Main branch, stable releases
- `develop`: Development branch
- `feature/*`: Feature branches
- `fix/*`: Fix branches

### Development Workflow

1. Fork this repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'feat: add amazing feature'`)
4. Push the branch (`git push origin feature/amazing-feature`)
5. Submit a Pull Request

---

## Roadmap

### v1.1 (Current)

- [x] Comments & Annotations system
- [x] Document template library
- [x] Collaborator online awareness
- [x] Univer professional spreadsheet engine integration
- [x] Operation logs and audit trail
- [x] Storage policy management

### v1.0

- [x] Document basic management
- [x] Rich text editor
- [x] Spreadsheet editor
- [x] Presentation editor
- [x] Mind map editor
- [x] Remote storage integration (SFTP/FTP/SMB/WebDAV/NFS)
- [x] Cloud storage (Alibaba OSS/Tencent COS/Baidu BOS)
- [x] File preview (PDF/Office)
- [x] User authentication and permissions
- [x] AI knowledge base Q&A

### v1.2 (Planned)

- [ ] Mobile adaptation
- [ ] Offline editing
- [ ] Internationalization support
- [ ] Plugin system

---

## Contributing

We welcome all forms of contributions!

### How to Contribute

1. Submit an Issue to report bugs or suggest features
2. Fork the project and submit a Pull Request
3. Improve documentation or translations

### Code of Conduct

Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md).

---

## FAQ

**Q: How to change the default port?**

A: Modify the `server.port` configuration in `kloggerx-server/config/config.yaml`.

**Q: How to configure HTTPS?**

A: We recommend using Nginx reverse proxy with SSL certificates.

**Q: How to run database migrations?**

A: The project uses GORM auto-migration, tables are created automatically when starting the service.

**Q: How to integrate with OnlyOffice?**

A: Configure the OnlyOffice server URL and JWT secret in the configuration file.

---

## License

This project is open-sourced under the [MIT](LICENSE) license.

---

## Acknowledgments

Thanks to the following open-source projects:

- [Vue.js](https://vuejs.org/)
- [Element Plus](https://element-plus.org/)
- [Tiptap](https://tiptap.dev/)
- [Univer](https://univer.ai/) - Spreadsheet engine
- [Yjs](https://yjs.dev/) - Real-time collaboration
- [Gin](https://gin-gonic.com/)
- [GORM](https://gorm.io/)

---

<div align="center">

**If this project helps you, please give it a ⭐️ Star!**

**For donations, please contact enluoo@163.com. Your generosity will help the continued development of this project.**

Made with ❤️ by KloggerX Team

</div>
