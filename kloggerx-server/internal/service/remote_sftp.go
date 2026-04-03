package service

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// SFTPClient implements RemoteClient for SFTP protocol
type SFTPClient struct {
	config     RemoteStorageConfig
	sshClient  *ssh.Client
	sftpClient *sftp.Client
	mu         sync.Mutex
	connected  bool
}

// NewSFTPClient creates a new SFTP client
func NewSFTPClient(config RemoteStorageConfig) *SFTPClient {
	return &SFTPClient{
		config: config,
	}
}

// Connect establishes SFTP connection
func (c *SFTPClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return nil
	}

	// Debug: log config
	log.Printf("[SFTP] Connecting to %s:%d, user=%s, sharePath=%q",
		c.config.Server, c.config.Port, c.config.Username, c.config.SharePath)

	// Default port
	port := c.config.Port
	if port == 0 {
		port = 22
	}

	// SSH config
	sshConfig := &ssh.ClientConfig{
		User: c.config.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(c.config.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	// Connect to SSH server
	addr := fmt.Sprintf("%s:%d", c.config.Server, port)
	sshClient, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return fmt.Errorf("SSH connection failed: %w", err)
	}

	// Create SFTP client
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return fmt.Errorf("SFTP client creation failed: %w", err)
	}

	c.sshClient = sshClient
	c.sftpClient = sftpClient
	c.connected = true
	log.Printf("[SFTP] Connected successfully")
	return nil
}

// GetBasePath returns the base path (SharePath if set, otherwise root)
func (c *SFTPClient) GetBasePath() string {
	if c.config.SharePath != "" {
		return normalizePath(c.config.SharePath)
	}
	return "/"
}

// Disconnect closes the SFTP connection
func (c *SFTPClient) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	var errs []error
	if c.sftpClient != nil {
		if err := c.sftpClient.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if c.sshClient != nil {
		if err := c.sshClient.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	c.connected = false
	c.sftpClient = nil
	c.sshClient = nil

	if len(errs) > 0 {
		return fmt.Errorf("disconnect errors: %v", errs)
	}
	return nil
}

// List returns files in the specified directory
func (c *SFTPClient) List(path string) ([]RemoteFileInfo, error) {
	if err := c.ensureConnected(); err != nil {
		return nil, err
	}

	// Normalize path
	path = normalizePath(path)

	// If path is root and SharePath is set, use SharePath as base
	basePath := c.GetBasePath()
	log.Printf("[SFTP] List: inputPath=%q, basePath=%q", path, basePath)
	if path == "/" && basePath != "/" {
		path = basePath
	} else if path != "/" && basePath != "/" && !strings.HasPrefix(path, basePath) {
		// Prepend base path only if path doesn't already start with it
		path = basePath + path
	}
	log.Printf("[SFTP] List: finalPath=%q", path)

	files, err := c.sftpClient.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory: %w", err)
	}

	result := make([]RemoteFileInfo, 0, len(files))
	for _, f := range files {
		filePath := path + "/" + f.Name()
		if path == "/" {
			filePath = "/" + f.Name()
		}
		result = append(result, RemoteFileInfo{
			Name:    f.Name(),
			Size:    f.Size(),
			IsDir:   f.IsDir(),
			ModTime: f.ModTime(),
			Path:    filePath,
		})
	}
	return result, nil
}

// Download retrieves a file from remote storage
func (c *SFTPClient) Download(path string) (io.ReadCloser, error) {
	if err := c.ensureConnected(); err != nil {
		return nil, err
	}

	path = normalizePath(path)

	// Prepend base path if SharePath is set
	basePath := c.GetBasePath()
	if basePath != "/" && !strings.HasPrefix(path, basePath) {
		path = basePath + path
	}

	file, err := c.sftpClient.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return file, nil
}

// Upload uploads a file to remote storage
func (c *SFTPClient) Upload(path string, reader io.Reader, size int64) error {
	if err := c.ensureConnected(); err != nil {
		return err
	}

	path = normalizePath(path)

	// Create parent directories if needed
	dir := dirFromPath(path)
	if dir != "" && dir != "/" {
		c.sftpClient.MkdirAll(dir)
	}

	file, err := c.sftpClient.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	return err
}

// Delete removes a file or directory
func (c *SFTPClient) Delete(path string) error {
	if err := c.ensureConnected(); err != nil {
		return err
	}

	path = normalizePath(path)

	// Check if it's a directory
	info, err := c.sftpClient.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat: %w", err)
	}

	if info.IsDir() {
		return c.sftpClient.RemoveAll(path)
	}
	return c.sftpClient.Remove(path)
}

// Mkdir creates a directory
func (c *SFTPClient) Mkdir(path string) error {
	if err := c.ensureConnected(); err != nil {
		return err
	}
	path = normalizePath(path)
	return c.sftpClient.MkdirAll(path)
}

// TestConnection tests the connection
func (c *SFTPClient) TestConnection() error {
	if err := c.Connect(); err != nil {
		return err
	}
	// Try to list root directory
	_, err := c.List("/")
	return err
}

// IsConnected returns connection status
func (c *SFTPClient) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected
}

// ensureConnected checks and establishes connection if needed
func (c *SFTPClient) ensureConnected() error {
	c.mu.Lock()
	connected := c.connected
	c.mu.Unlock()

	if connected {
		return nil
	}
	return c.Connect()
}

// normalizePath ensures consistent path format
func normalizePath(path string) string {
	if path == "" {
		return "/"
	}
	// Convert backslashes to forward slashes
	path = strings.ReplaceAll(path, "\\", "/")
	// Ensure it starts with /
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	// Remove trailing slash (except for root)
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	return path
}

// dirFromPath extracts directory from file path
func dirFromPath(path string) string {
	idx := strings.LastIndex(path, "/")
	if idx <= 0 {
		return "/"
	}
	return path[:idx]
}

// Ensure SFTPClient implements RemoteClient
var _ RemoteClient = (*SFTPClient)(nil)

// Check if port is available (helper function)
func isPortOpen(host string, port int, timeout time.Duration) bool {
	target := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", target, timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
