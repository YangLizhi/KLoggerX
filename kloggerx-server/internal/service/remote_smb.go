package service

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/hirochachacha/go-smb2"
)

// SMBClient implements RemoteClient for SMB/CIFS protocol
type SMBClient struct {
	config    RemoteStorageConfig
	session   *smb2.Session
	share     *smb2.Share
	mu        sync.Mutex
	connected bool
}

// NewSMBClient creates a new SMB client
func NewSMBClient(config RemoteStorageConfig) *SMBClient {
	return &SMBClient{
		config: config,
	}
}

// Connect establishes SMB connection
func (c *SMBClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return nil
	}

	// Default port
	port := c.config.Port
	if port == 0 {
		port = 445
	}

	// Connect to SMB server via TCP first
	addr := fmt.Sprintf("%s:%d", c.config.Server, port)
	tcpConn, err := net.DialTimeout("tcp", addr, 30*time.Second)
	if err != nil {
		return fmt.Errorf("SMB TCP connection failed: %w", err)
	}

	// SMB dialer - support guest access when no credentials provided
	var d *smb2.Dialer
	if c.config.Username == "" {
		// Guest access mode
		d = &smb2.Dialer{
			Initiator: &smb2.NTLMInitiator{
				User:     "Guest",
				Password: "",
				Domain:   c.config.Domain,
			},
		}
	} else {
		d = &smb2.Dialer{
			Initiator: &smb2.NTLMInitiator{
				User:     c.config.Username,
				Password: c.config.Password,
				Domain:   c.config.Domain,
			},
		}
	}

	// Negotiate SMB session
	session, err := d.Dial(tcpConn)
	if err != nil {
		tcpConn.Close()
		return fmt.Errorf("SMB session failed: %w", err)
	}

	// Get share path (default to IPC$ if not specified)
	sharePath := c.config.SharePath
	if sharePath == "" {
		sharePath = "IPC$"
	}
	// Clean share path - remove leading slashes
	sharePath = strings.TrimPrefix(sharePath, "\\")
	sharePath = strings.TrimPrefix(sharePath, "/")

	// Mount share
	share, err := session.Mount(sharePath)
	if err != nil {
		session.Logoff()
		return fmt.Errorf("SMB mount failed: %w", err)
	}

	c.session = session
	c.share = share
	c.connected = true
	return nil
}

// Disconnect closes the SMB connection
func (c *SMBClient) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	var errs []error
	if c.share != nil {
		if err := c.share.Umount(); err != nil {
			errs = append(errs, err)
		}
	}
	if c.session != nil {
		if err := c.session.Logoff(); err != nil {
			errs = append(errs, err)
		}
	}

	c.connected = false
	c.share = nil
	c.session = nil

	if len(errs) > 0 {
		return fmt.Errorf("disconnect errors: %v", errs)
	}
	return nil
}

// List returns files in the specified directory
func (c *SMBClient) List(path string) ([]RemoteFileInfo, error) {
	if err := c.ensureConnected(); err != nil {
		return nil, err
	}

	// Validate path security
	if _, err := ValidatePath(path); err != nil {
		return nil, err
	}

	// Normalize path for SMB
	path = normalizePathSMB(path)

	// Use backslash for SMB
	path = strings.ReplaceAll(path, "/", "\\")

	// Read directory
	f, err := c.share.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		// Try as directory
		entries, err := c.share.ReadDir(path)
		if err != nil {
			return nil, fmt.Errorf("failed to list directory: %w", err)
		}
		return c.entriesToRemoteFiles(entries, path), nil
	}
	defer f.Close()

	// Check if it's a file
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		// It's a file, return it as single item
		return []RemoteFileInfo{
			{
				Name:    info.Name(),
				Size:    info.Size(),
				IsDir:   false,
				ModTime: info.ModTime(),
				Path:    path,
			},
		}, nil
	}

	// It's a directory, read entries
	f.Close()
	entries, err := c.share.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory: %w", err)
	}
	return c.entriesToRemoteFiles(entries, path), nil
}

// entriesToRemoteFiles converts os.FileInfo slice to RemoteFileInfo slice
func (c *SMBClient) entriesToRemoteFiles(entries []os.FileInfo, basePath string) []RemoteFileInfo {
	result := make([]RemoteFileInfo, 0, len(entries))
	for _, entry := range entries {
		filePath := basePath + "\\" + entry.Name()
		// Convert backslashes to forward slashes
		filePath = strings.ReplaceAll(filePath, "\\", "/")
		// Ensure path starts with single /
		if !strings.HasPrefix(filePath, "/") {
			filePath = "/" + filePath
		}
		// Remove duplicate slashes
		for strings.Contains(filePath, "//") {
			filePath = strings.ReplaceAll(filePath, "//", "/")
		}
		result = append(result, RemoteFileInfo{
			Name:    entry.Name(),
			Size:    entry.Size(),
			IsDir:   entry.IsDir(),
			ModTime: entry.ModTime(),
			Path:    filePath,
		})
	}
	return result
}

// Download retrieves a file from remote storage
func (c *SMBClient) Download(path string) (io.ReadCloser, error) {
	if err := c.ensureConnected(); err != nil {
		return nil, err
	}

	// Validate path security
	if _, err := ValidatePath(path); err != nil {
		return nil, err
	}

	path = normalizePathSMB(path)
	path = strings.ReplaceAll(path, "/", "\\")

	file, err := c.share.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return file, nil
}

// Upload uploads a file to remote storage
func (c *SMBClient) Upload(path string, reader io.Reader, size int64) error {
	if err := c.ensureConnected(); err != nil {
		return err
	}

	// Validate path security
	if _, err := ValidatePath(path); err != nil {
		return err
	}

	path = normalizePathSMB(path)
	smbPath := strings.ReplaceAll(path, "/", "\\")

	// Create parent directories if needed
	dir := path[:strings.LastIndex(path, "/")]
	if dir != "" && dir != "/" {
		smbDir := strings.ReplaceAll(dir, "/", "\\")
		c.share.MkdirAll(smbDir, 0755)
	}

	// Create/truncate file
	file, err := c.share.OpenFile(smbPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	return err
}

// Delete removes a file or directory
func (c *SMBClient) Delete(path string) error {
	if err := c.ensureConnected(); err != nil {
		return err
	}

	// Validate path security
	if _, err := ValidatePath(path); err != nil {
		return err
	}

	path = normalizePathSMB(path)
	smbPath := strings.ReplaceAll(path, "/", "\\")

	// Try to remove as file first
	err := c.share.Remove(smbPath)
	if err == nil {
		return nil
	}

	// Try as directory
	return c.share.RemoveAll(smbPath)
}

// Mkdir creates a directory
func (c *SMBClient) Mkdir(path string) error {
	if err := c.ensureConnected(); err != nil {
		return err
	}
	// Validate path security
	if _, err := ValidatePath(path); err != nil {
		return err
	}
	path = normalizePathSMB(path)
	smbPath := strings.ReplaceAll(path, "/", "\\")
	return c.share.MkdirAll(smbPath, 0755)
}

// TestConnection tests the connection
func (c *SMBClient) TestConnection() error {
	if err := c.Connect(); err != nil {
		return err
	}
	// Try to list root directory
	_, err := c.List("/")
	return err
}

// IsConnected returns connection status
func (c *SMBClient) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected
}

// ensureConnected checks and establishes connection if needed
func (c *SMBClient) ensureConnected() error {
	c.mu.Lock()
	connected := c.connected
	c.mu.Unlock()

	if connected {
		return nil
	}
	return c.Connect()
}

// normalizePathSMB ensures consistent path format for SMB
func normalizePathSMB(p string) string {
	if p == "" {
		return "\\"
	}
	// Convert forward slashes to backslashes
	p = strings.ReplaceAll(p, "/", "\\")
	// Ensure it starts with backslash
	if !strings.HasPrefix(p, "\\") {
		p = "\\" + p
	}
	// Remove trailing backslash
	if len(p) > 1 && strings.HasSuffix(p, "\\") {
		p = p[:len(p)-1]
	}
	return p
}

// Ensure SMBClient implements RemoteClient
var _ RemoteClient = (*SMBClient)(nil)
