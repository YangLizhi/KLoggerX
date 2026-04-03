package service

import (
	"io"
	"time"
)

// RemoteFileInfo represents file information from remote storage
type RemoteFileInfo struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	IsDir   bool      `json:"isDir"`
	ModTime time.Time `json:"modTime"`
	Path    string    `json:"path"`
}

// RemoteClient defines the interface for remote storage clients
type RemoteClient interface {
	// Connect establishes connection to the remote server
	Connect() error

	// Disconnect closes the connection
	Disconnect() error

	// List returns list of files in the specified directory
	List(path string) ([]RemoteFileInfo, error)

	// Download retrieves a file from remote storage
	Download(path string) (io.ReadCloser, error)

	// Upload uploads a file to remote storage
	Upload(path string, reader io.Reader, size int64) error

	// Delete removes a file or directory from remote storage
	Delete(path string) error

	// Mkdir creates a directory on remote storage
	Mkdir(path string) error

	// TestConnection tests if the connection is valid
	TestConnection() error

	// IsConnected returns current connection status
	IsConnected() bool
}

// RemoteStorageConfig holds configuration for remote storage connection
type RemoteStorageConfig struct {
	Type       string // ftp, sftp, smb, nfs, webdav, baidu, aliyun, tencent
	Server     string
	Port       int
	Username   string
	Password   string
	SharePath  string // For SMB/NFS
	Domain     string // For SMB AD authentication
	MountPoint string // Virtual mount point name

	// Cloud drive specific fields
	AccessToken  string `json:"accessToken"`  // OAuth access token
	RefreshToken string `json:"refreshToken"` // OAuth refresh token
	APIKey       string `json:"apiKey"`       // API key for some services
	ExpiresAt    int64  `json:"expiresAt"`    // Token expiration timestamp

	// WebDAV specific
	UseHTTPS bool `json:"useHTTPS"` // Use HTTPS for WebDAV

	// Additional settings
	RootPath string `json:"rootPath"` // Root path to start browsing
}
