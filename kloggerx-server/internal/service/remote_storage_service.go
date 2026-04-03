package service

import (
	"fmt"
	"io"
	"log"
	"sync"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/repository/mysql"
)

// RemoteStorageService manages remote storage connections and operations
type RemoteStorageService struct {
	clients sync.Map // storageID -> RemoteClient
	mu      sync.RWMutex
}

// Global remote storage service instance
var remoteStorageService = &RemoteStorageService{}

// GetRemoteStorageService returns the global remote storage service
func GetRemoteStorageService() *RemoteStorageService {
	return remoteStorageService
}

// GetClient returns a client for the specified storage ID
// If client is not cached, it creates and connects a new one
func (s *RemoteStorageService) GetClient(storageID uint) (RemoteClient, error) {
	// Check cache first
	if client, ok := s.clients.Load(storageID); ok {
		rc := client.(RemoteClient)
		if rc.IsConnected() {
			log.Printf("[RemoteStorage] Using cached client for storage %d", storageID)
			return rc, nil
		}
		// Client is disconnected, remove from cache
		s.clients.Delete(storageID)
		log.Printf("[RemoteStorage] Cached client for storage %d was disconnected, creating new one", storageID)
	}

	// Load storage config from database
	var storage model.RemoteStorage
	if err := mysql.DB.First(&storage, storageID).Error; err != nil {
		return nil, fmt.Errorf("storage not found: %w", err)
	}

	// Debug: log loaded config
	log.Printf("[RemoteStorage] Loaded config from DB: id=%d, type=%s, server=%s, port=%d, username=%s, sharePath=%q",
		storage.ID, storage.Type, storage.Server, storage.Port, storage.Username, storage.SharePath)

	// Check if enabled
	if !storage.IsEnabled {
		return nil, fmt.Errorf("storage is disabled")
	}

	// Create new client with all config fields
	config := RemoteStorageConfig{
		Type:        storage.Type,
		Server:      storage.Server,
		Port:        storage.Port,
		Username:    storage.Username,
		Password:    storage.Password,
		SharePath:   storage.SharePath,
		Domain:      storage.Domain,
		MountPoint:  storage.MountPoint,
		AccessToken: storage.AccessToken,
		RefreshToken: storage.RefreshToken,
		APIKey:      storage.APIKey,
		ExpiresAt:   storage.ExpiresAt,
		RootPath:    storage.RootPath,
	}

	client, err := s.createClient(config)
	if err != nil {
		return nil, err
	}

	// Connect
	if err := client.Connect(); err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}

	// Cache the client
	s.clients.Store(storageID, client)
	return client, nil
}

// createClient creates a new remote client based on the storage type
func (s *RemoteStorageService) createClient(config RemoteStorageConfig) (RemoteClient, error) {
	switch config.Type {
	case "sftp":
		return NewSFTPClient(config), nil
	case "ftp":
		return NewFTPClient(config), nil
	case "smb":
		return NewSMBClient(config), nil
	case "webdav":
		return NewWebDAVClient(config), nil
	case "nfs":
		return NewNFSClient(config), nil
	case "baidu":
		return NewBaiduClient(config), nil
	case "aliyun":
		return NewAliyunClient(config), nil
	case "tencent":
		return NewTencentClient(config), nil
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", config.Type)
	}
}

// ListFiles lists files in a remote storage directory
func (s *RemoteStorageService) ListFiles(storageID uint, path string) ([]RemoteFileInfo, error) {
	client, err := s.GetClient(storageID)
	if err != nil {
		return nil, err
	}
	return client.List(path)
}

// DownloadFile downloads a file from remote storage
func (s *RemoteStorageService) DownloadFile(storageID uint, path string) (io.ReadCloser, error) {
	client, err := s.GetClient(storageID)
	if err != nil {
		return nil, err
	}
	return client.Download(path)
}

// UploadFile uploads a file to remote storage
func (s *RemoteStorageService) UploadFile(storageID uint, path string, reader io.Reader, size int64) error {
	client, err := s.GetClient(storageID)
	if err != nil {
		return err
	}
	return client.Upload(path, reader, size)
}

// DeleteFile deletes a file or directory from remote storage
func (s *RemoteStorageService) DeleteFile(storageID uint, path string) error {
	client, err := s.GetClient(storageID)
	if err != nil {
		return err
	}
	return client.Delete(path)
}

// CreateDir creates a directory in remote storage
func (s *RemoteStorageService) CreateDir(storageID uint, path string) error {
	client, err := s.GetClient(storageID)
	if err != nil {
		return err
	}
	return client.Mkdir(path)
}

// TestConnection tests connection to a remote storage
func (s *RemoteStorageService) TestConnection(storageID uint) error {
	client, err := s.GetClient(storageID)
	if err != nil {
		return err
	}
	return client.TestConnection()
}

// Disconnect closes connection to a remote storage
func (s *RemoteStorageService) Disconnect(storageID uint) error {
	if client, ok := s.clients.Load(storageID); ok {
		s.clients.Delete(storageID)
		return client.(RemoteClient).Disconnect()
	}
	return nil
}

// DisconnectAll closes all remote storage connections
func (s *RemoteStorageService) DisconnectAll() {
	s.clients.Range(func(key, value interface{}) bool {
		if client, ok := value.(RemoteClient); ok {
			client.Disconnect()
		}
		s.clients.Delete(key)
		return true
	})
}

// TestConnectionWithConfig tests connection with provided config (without saving)
func (s *RemoteStorageService) TestConnectionWithConfig(config RemoteStorageConfig) error {
	client, err := s.createClient(config)
	if err != nil {
		return err
	}
	defer client.Disconnect()
	return client.TestConnection()
}
