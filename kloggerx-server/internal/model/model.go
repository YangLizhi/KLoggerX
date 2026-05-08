package model

import (
	"time"

	"kloggerx-server/internal/pkg/crypto"

	"gorm.io/gorm"
)

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"size:50;uniqueIndex;not null" json:"username"`
	Email        string    `gorm:"size:100;uniqueIndex;not null" json:"email"`
	Password     string    `gorm:"size:255;not null" json:"-"`
	Nickname     string    `gorm:"size:50" json:"nickname"`
	Avatar       string    `gorm:"size:500" json:"avatar"`
	Role         string    `gorm:"size:20;default:member" json:"role"`
	AuthSource   string    `gorm:"size:20;default:local" json:"authSource"` // local, ldap
	DepartmentID uint      `gorm:"default:0" json:"departmentId"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type Department struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	Name        string       `gorm:"size:100;not null" json:"name"`
	ParentID    *uint        `json:"parentId"`
	Children    []Department `gorm:"foreignKey:ParentID" json:"children"`
	MemberCount int          `gorm:"-" json:"memberCount"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
}

type Document struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	Title            string     `gorm:"size:500;not null;default:'无标题文档'" json:"title"`
	Type             string     `gorm:"size:20;not null;index" json:"type"`
	ParentID         *uint      `gorm:"index" json:"parentId"`
	OwnerID          uint       `gorm:"index;not null" json:"ownerId"`
	OwnerName        string     `gorm:"-" json:"ownerName"`
	Content          string     `gorm:"type:longtext" json:"content"`
	IsPinned         bool       `gorm:"default:false" json:"isPinned"`
	IsFavorite       bool       `gorm:"-" json:"isFavorite"`
	IsShortcut       bool       `gorm:"default:false" json:"isShortcut"`
	ShortcutTargetID *uint      `json:"shortcutTargetId"`
	IsDeleted        bool       `gorm:"default:false;index" json:"isDeleted"`
	DeletedAt        *time.Time `json:"deletedAt"`
	Version          int        `gorm:"default:1" json:"version"`
	// File metadata (populated for uploaded files)
	FileSize     int64  `gorm:"default:0" json:"fileSize"`
	FileExt      string `gorm:"size:20" json:"fileExt"`
	OriginalName string `gorm:"size:500" json:"originalName"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	Children     []Document `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

type DocumentVersion struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	DocumentID uint      `gorm:"index;not null" json:"documentId"`
	Version    int       `gorm:"not null" json:"version"`
	Content    string    `gorm:"type:longtext" json:"content"`
	EditorID   uint      `gorm:"not null" json:"editorId"`
	EditorName string    `gorm:"-" json:"editorName"`
	CreatedAt  time.Time `json:"createdAt"`
}

type Favorite struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"uniqueIndex:idx_user_doc;not null" json:"userId"`
	DocumentID uint      `gorm:"uniqueIndex:idx_user_doc;not null" json:"documentId"`
	CreatedAt  time.Time `json:"createdAt"`
}

type Permission struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	DocumentID uint      `gorm:"uniqueIndex:idx_doc_user;not null" json:"documentId"`
	UserID     uint      `gorm:"uniqueIndex:idx_doc_user;not null" json:"userId"`
	UserName   string    `gorm:"-" json:"userName"`
	UserAvatar string    `gorm:"-" json:"userAvatar"`
	Level      string    `gorm:"size:20;not null" json:"level"`
	Inherited  bool      `gorm:"default:false" json:"inherited"`
	CreatedAt  time.Time `json:"createdAt"`
}

type ShareSetting struct {
	ID                uint   `gorm:"primaryKey" json:"id"`
	DocumentID        uint   `gorm:"uniqueIndex;not null" json:"documentId"`
	Scope             string `gorm:"size:20;default:collaborator" json:"scope"`
	DefaultPermission string `gorm:"size:20;default:view" json:"defaultPermission"`
	LinkEnabled       bool   `gorm:"default:false" json:"linkEnabled"`
	ShareLink         string `gorm:"size:500" json:"shareLink"`
	IncludeChildren   bool   `gorm:"default:false" json:"includeChildren"`
}

type KnowledgeBase struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:200;not null" json:"name"`
	Description string    `gorm:"size:1000" json:"description"`
	OwnerID     uint      `gorm:"not null" json:"ownerId"`
	OwnerName   string    `gorm:"-" json:"ownerName"`
	MemberCount int       `gorm:"-" json:"memberCount"`
	DocCount    int       `gorm:"-" json:"docCount"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type KnowledgeMember struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	KnowledgeBaseID uint      `gorm:"uniqueIndex:idx_kb_user;not null" json:"knowledgeBaseId"`
	UserID          uint      `gorm:"uniqueIndex:idx_kb_user;not null" json:"userId"`
	UserName        string    `gorm:"-" json:"userName"`
	UserAvatar      string    `gorm:"-" json:"userAvatar"`
	Role            string    `gorm:"size:20;default:member" json:"role"`
	CreatedAt       time.Time `json:"createdAt"`
}

type KnowledgeDocument struct {
	ID              uint `gorm:"primaryKey" json:"id"`
	KnowledgeBaseID uint `gorm:"uniqueIndex:idx_kb_doc;not null" json:"knowledgeBaseId"`
	DocumentID      uint `gorm:"uniqueIndex:idx_kb_doc;not null" json:"documentId"`
}

type Comment struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	DocumentID uint      `json:"document_id" gorm:"index;not null"`
	UserID     uint      `json:"user_id" gorm:"index;not null"`
	User       User      `json:"user" gorm:"foreignKey:UserID"`
	Content    string    `json:"content" gorm:"type:text;not null" binding:"required,max=5000"`
	ParentID   *uint     `json:"parent_id" gorm:"index"`
	QuotedText string    `json:"quoted_text" gorm:"type:text"`
	Resolved   bool      `json:"resolved" gorm:"default:false"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Replies    []Comment `json:"replies" gorm:"foreignKey:ParentID"`
}

type Notification struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Type         string    `gorm:"size:20;not null" json:"type"`
	Title        string    `gorm:"size:200;not null" json:"title"`
	Content      string    `gorm:"size:500" json:"content"`
	FromUserID   uint      `gorm:"not null" json:"fromUserId"`
	FromUserName string    `gorm:"-" json:"fromUserName"`
	ToUserID     uint      `gorm:"index;not null" json:"toUserId"`
	DocumentID   uint      `json:"documentId"`
	IsRead       bool      `gorm:"default:false" json:"isRead"`
	CreatedAt    time.Time `json:"createdAt"`
}

type RecentDocument struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"uniqueIndex:idx_user_doc_recent;not null" json:"userId"`
	DocumentID uint      `gorm:"uniqueIndex:idx_user_doc_recent;not null" json:"documentId"`
	AccessedAt time.Time `json:"accessedAt"`
}

type FileRecord struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Name       string    `gorm:"size:500;not null" json:"name"`
	Path       string    `gorm:"size:1000;not null" json:"path"`
	Size       int64     `json:"size"`
	MimeType   string    `gorm:"size:100" json:"mimeType"`
	UploaderID uint      `gorm:"not null" json:"uploaderId"`
	CreatedAt  time.Time `json:"createdAt"`
}

type OperationLog struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"index;not null" json:"userId"`
	UserName      string    `gorm:"size:50" json:"userName"`
	Action        string    `gorm:"size:50;not null;index" json:"action"`
	Resource      string    `gorm:"size:500" json:"resource"`
	ResourceType  string    `gorm:"size:30;index" json:"resourceType"`
	ResourceID    uint      `gorm:"index" json:"resourceId"`
	ResourceTitle string    `gorm:"size:500" json:"resourceTitle"`
	Detail        string    `gorm:"type:text" json:"detail"`
	IP            string    `gorm:"size:45" json:"ip"`
	UserAgent     string    `gorm:"size:500" json:"userAgent"`
	Status        int       `json:"status"`
	Duration      int64     `json:"duration"`
	CreatedAt     time.Time `gorm:"index" json:"createdAt"`
}

type SystemSetting struct {
	Key   string `gorm:"primaryKey;size:100" json:"key"`
	Value string `gorm:"type:longtext" json:"value"`
}

// Template is a pre-built document template that can be used to quickly create documents.
type Template struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:200;not null" json:"name"`
	Description string    `gorm:"size:500" json:"description"`
	Type        string    `gorm:"size:20;not null" json:"type"` // doc, sheet, slide, mindnote, bitable
	Category    string    `gorm:"size:50;index" json:"category"`
	Content     string    `gorm:"type:longtext" json:"content"`
	Preview     string    `gorm:"type:text" json:"preview"` // Text preview extracted from content
	IsBuiltin   bool      `gorm:"default:false" json:"isBuiltin"`
	CreatedAt   time.Time `json:"createdAt"`
}

// TemplateFavorite stores user's favorite templates.
type TemplateFavorite struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"uniqueIndex:idx_user_template;not null" json:"userId"`
	TemplateID uint      `gorm:"uniqueIndex:idx_user_template;not null" json:"templateId"`
	CreatedAt  time.Time `json:"createdAt"`
}

// KnowledgeSource links a KnowledgeBase to a cloud-drive folder or a document.
// When linked, documents from the source are indexed into the knowledge base for AI Q&A.
type KnowledgeSource struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	KnowledgeBaseID uint       `gorm:"index;not null" json:"knowledgeBaseId"`
	SourceType      string     `gorm:"size:20;not null" json:"sourceType"` // "folder" | "document"
	SourceID        uint       `gorm:"not null" json:"sourceId"`
	SourceName      string     `gorm:"size:500" json:"sourceName"`
	AutoSync        bool       `gorm:"default:true" json:"autoSync"`
	DocCount        int        `gorm:"-" json:"docCount"`
	LastSyncAt      *time.Time `json:"lastSyncAt"`
	CreatedAt       time.Time  `json:"createdAt"`
}

// KnowledgeChunk stores chunked content from documents for RAG retrieval.
type KnowledgeChunk struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	KnowledgeBaseID uint      `gorm:"index;not null" json:"knowledgeBaseId"`
	DocumentID      uint      `gorm:"index;not null" json:"documentId"`
	DocumentTitle   string    `gorm:"size:500" json:"documentTitle"`
	Content         string    `gorm:"type:text;not null" json:"content"`
	ChunkIndex      int       `gorm:"not null" json:"chunkIndex"`
	SourceType      string    `gorm:"size:30;default:'document'" json:"sourceType"` // document, user_contribution
	QdrantPointID   string    `gorm:"size:64" json:"qdrantPointId"`           // Qdrant vector point ID
	EmbeddingStatus string    `gorm:"size:20;default:pending" json:"embeddingStatus"` // pending, embedded, failed
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// EmbeddingJob tracks the vectorization status of documents.
type EmbeddingJob struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	KnowledgeBaseID uint       `gorm:"index;not null" json:"knowledgeBaseId"`
	DocumentID      uint       `gorm:"index;not null" json:"documentId"`
	Status          string     `gorm:"size:20;default:pending" json:"status"` // pending, processing, completed, failed
	ChunkCount      int        `gorm:"default:0" json:"chunkCount"`
	ErrorMessage    string     `gorm:"type:text" json:"errorMessage"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

// RaptorNode represents a node in the RAPTOR tree for hierarchical document summarization.
type RaptorNode struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	KnowledgeBaseID uint       `gorm:"index;not null" json:"knowledgeBaseId"`
	NodeType        string     `gorm:"size:20;not null" json:"nodeType"` // leaf, cluster, root
	ParentID        *uint      `gorm:"index" json:"parentId"`
	DocumentIDs     string     `gorm:"type:json" json:"documentIds"`     // JSON array of document IDs
	Summary         string     `gorm:"type:longtext;not null" json:"summary"`
	QdrantPointID   string     `gorm:"size:64" json:"qdrantPointId"`
	Level           int        `gorm:"default:0" json:"level"`           // 0=leaf, 1=cluster, 2+=higher
	CreatedAt       time.Time  `json:"createdAt"`
}

// UserStorageSetting stores user-specific storage path configurations.
type UserStorageSetting struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"uniqueIndex;not null" json:"userId"`
	SyncDir     string    `gorm:"size:500" json:"syncDir"`      // Personal sync directory path
	DownloadDir string    `gorm:"size:500" json:"downloadDir"`  // Personal download directory path
	AutoSync    bool      `gorm:"default:true" json:"autoSync"` // Auto-sync toggle
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// RemoteStorage stores remote storage connection configurations (SMB, FTP, etc.)
type RemoteStorage struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"size:100;not null" json:"name"`
	Type         string    `gorm:"size:20;not null" json:"type"`         // ftp, sftp, smb, nfs, webdav, baidu, aliyun, tencent
	Server       string    `gorm:"size:255" json:"server"`               // Server address (optional for cloud drives)
	Port         int       `json:"port"`
	Username     string    `gorm:"size:100" json:"username"`
	Password     string    `gorm:"size:255" json:"-"`                    // Encrypted, not returned to frontend
	SharePath    string    `gorm:"size:500" json:"sharePath"`            // Share path for SMB/NFS
	Domain       string    `gorm:"size:100" json:"domain"`               // AD Domain for SMB
	MountPoint   string    `gorm:"size:255;not null" json:"mountPoint"`  // Mount point name (virtual directory name)
	Status       string    `gorm:"size:20;default:disconnected" json:"status"` // connected, disconnected, error
	IsEnabled    bool      `gorm:"default:true" json:"isEnabled"`
	
	// Cloud drive specific fields
	AccessToken  string    `gorm:"size:1000" json:"-"`       // OAuth access token (encrypted)
	RefreshToken string    `gorm:"size:1000" json:"-"`       // OAuth refresh token (encrypted)
	APIKey       string    `gorm:"size:255" json:"-"`        // API key for some services (encrypted)
	ExpiresAt    int64     `json:"expiresAt"`                // Token expiration timestamp
	RootPath     string    `gorm:"size:500" json:"rootPath"` // Root path to start browsing
	
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// BeforeSave encrypts sensitive fields before saving to database.
func (r *RemoteStorage) BeforeSave(tx *gorm.DB) error {
	if r.Password != "" {
		encrypted, err := crypto.Encrypt(r.Password)
		if err != nil {
			return err
		}
		r.Password = encrypted
	}
	if r.AccessToken != "" {
		encrypted, err := crypto.Encrypt(r.AccessToken)
		if err != nil {
			return err
		}
		r.AccessToken = encrypted
	}
	if r.RefreshToken != "" {
		encrypted, err := crypto.Encrypt(r.RefreshToken)
		if err != nil {
			return err
		}
		r.RefreshToken = encrypted
	}
	if r.APIKey != "" {
		encrypted, err := crypto.Encrypt(r.APIKey)
		if err != nil {
			return err
		}
		r.APIKey = encrypted
	}
	return nil
}

// AfterFind decrypts sensitive fields after reading from database.
// Silently ignores decryption errors to remain compatible with legacy plaintext data.
func (r *RemoteStorage) AfterFind(tx *gorm.DB) error {
	if r.Password != "" {
		if decrypted, err := crypto.Decrypt(r.Password); err == nil {
			r.Password = decrypted
		}
	}
	if r.AccessToken != "" {
		if decrypted, err := crypto.Decrypt(r.AccessToken); err == nil {
			r.AccessToken = decrypted
		}
	}
	if r.RefreshToken != "" {
		if decrypted, err := crypto.Decrypt(r.RefreshToken); err == nil {
			r.RefreshToken = decrypted
		}
	}
	if r.APIKey != "" {
		if decrypted, err := crypto.Decrypt(r.APIKey); err == nil {
			r.APIKey = decrypted
		}
	}
	return nil
}

// KnowledgeGraph 知识图谱
type KnowledgeGraph struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	KnowledgeBaseID uint      `gorm:"uniqueIndex;not null" json:"knowledgeBaseId"`
	GraphData       string    `gorm:"type:longtext" json:"-"`
	NodeCount       int       `json:"nodeCount"`
	EdgeCount       int       `json:"edgeCount"`
	CommunityCount  int       `json:"communityCount"`
	Status          string    `gorm:"size:20;default:'idle'" json:"status"`
	ErrorMessage    string    `gorm:"type:text" json:"errorMessage,omitempty"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// StorageUsage tracks storage usage statistics per user by file type.
type StorageUsage struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"uniqueIndex:idx_user_type;not null" json:"userId"`
	FileType   string    `gorm:"size:20;uniqueIndex:idx_user_type;not null" json:"fileType"` // doc, sheet, slide, image, other
	TotalSize  int64     `json:"totalSize"`  // Total size in bytes
	FileCount  int       `json:"fileCount"`  // Number of files
	UpdatedAt  time.Time `json:"updatedAt"`
}

// KbConversation stores AI assistant conversation sessions.
type KbConversation struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	UserID          uint           `gorm:"not null;index:idx_user_kb;index:idx_user_updated" json:"userId"`
	KnowledgeBaseID *uint          `gorm:"index:idx_user_kb" json:"knowledgeBaseId"` // NULL表示全库对话
	Title           string         `gorm:"size:255;not null;default:'新对话'" json:"title"`
	Model           string         `gorm:"size:100;default:''" json:"model"`
	MessageCount    uint           `gorm:"default:0" json:"messageCount"`
	IsPinned        bool           `gorm:"default:false" json:"isPinned"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// KbMessage stores individual messages within a conversation.
type KbMessage struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	ConversationID uint      `gorm:"not null;index:idx_conversation" json:"conversationId"`
	Role           string    `gorm:"type:enum('user','assistant','system');not null" json:"role"`
	Content        string    `gorm:"type:longtext;not null" json:"content"`
	Sources        *string   `gorm:"type:json" json:"sources"` // JSON: [{docId, title, chunkContent}]
	TokensUsed     uint      `gorm:"default:0" json:"tokensUsed"`
	DurationMs     uint      `gorm:"default:0" json:"durationMs"`
	CreatedAt      time.Time `json:"createdAt"`
}

// KbFeedback stores user feedback on AI assistant responses.
type KbFeedback struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	MessageID      uint       `gorm:"not null;uniqueIndex:idx_message_user" json:"messageId"`
	ConversationID uint       `gorm:"not null;index:idx_conversation" json:"conversationId"`
	UserID         uint       `gorm:"not null;uniqueIndex:idx_message_user" json:"userId"`
	Rating         int8       `gorm:"not null" json:"rating"` // 1=赞, -1=踩
	FeedbackType   string     `gorm:"size:50;default:''" json:"feedbackType"` // inaccurate/incomplete/irrelevant/outdated
	Comment        *string    `gorm:"type:text" json:"comment"`
	CorrectAnswer  *string    `gorm:"type:text" json:"correctAnswer"`
	ReviewStatus   string     `gorm:"size:20;default:'pending'" json:"reviewStatus"` // pending/approved/rejected
	ReviewerID     *uint      `json:"reviewerId"`
	ReviewComment  string     `gorm:"size:500;default:''" json:"reviewComment"`
	ReviewedAt     *time.Time `json:"reviewedAt"`
	CreatedAt      time.Time  `json:"createdAt"`
}

// KbShareLink stores share links for AI conversations.
type KbShareLink struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	ConversationID uint       `gorm:"not null;index" json:"conversationId"`
	UserID         uint       `gorm:"not null" json:"userId"`
	ShareToken     string     `gorm:"size:64;uniqueIndex" json:"shareToken"`
	ExpiresAt      *time.Time `json:"expiresAt"`
	IsActive       bool       `gorm:"default:true" json:"isActive"`
	ViewCount      uint       `gorm:"default:0" json:"viewCount"`
	CreatedAt      time.Time  `json:"createdAt"`
}
