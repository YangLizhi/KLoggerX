package model

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"size:50;uniqueIndex;not null" json:"username"`
	Email        string    `gorm:"size:100;uniqueIndex;not null" json:"email"`
	Password     string    `gorm:"size:255;not null" json:"-"`
	Nickname     string    `gorm:"size:50" json:"nickname"`
	Avatar       string    `gorm:"size:500" json:"avatar"`
	Role         string    `gorm:"size:20;default:member" json:"role"`
	DepartmentID uint      `gorm:"default:0" json:"departmentId"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type Department struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	Name        string       `gorm:"size:100;not null" json:"name"`
	ParentID    *uint        `json:"parentId"`
	Children    []Department `gorm:"foreignKey:ParentID" json:"children,omitempty"`
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
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
	Children         []Document `gorm:"foreignKey:ParentID" json:"children,omitempty"`
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
	ID         uint      `gorm:"primaryKey" json:"id"`
	DocumentID uint      `gorm:"index;not null" json:"documentId"`
	UserID     uint      `gorm:"not null" json:"userId"`
	UserName   string    `gorm:"-" json:"userName"`
	Content    string    `gorm:"type:text;not null" json:"content"`
	Selection  string    `gorm:"type:text" json:"selection"`
	ParentID   *uint     `json:"parentId"`
	Resolved   bool      `gorm:"default:false" json:"resolved"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
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
	Action        string    `gorm:"size:30;not null;index" json:"action"`
	ResourceType  string    `gorm:"size:30;not null;index" json:"resourceType"`
	ResourceID    uint      `gorm:"index;not null" json:"resourceId"`
	ResourceTitle string    `gorm:"size:500" json:"resourceTitle"`
	Detail        string    `gorm:"type:text" json:"detail"`
	IP            string    `gorm:"size:50" json:"ip"`
	CreatedAt     time.Time `json:"createdAt"`
}

type SystemSetting struct {
	Key   string `gorm:"primaryKey;size:100" json:"key"`
	Value string `gorm:"type:text" json:"value"`
}
