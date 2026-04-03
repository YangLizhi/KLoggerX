package service

import (
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/jlaffaye/ftp"
)

// FTPClient implements RemoteClient for FTP protocol
type FTPClient struct {
	config    RemoteStorageConfig
	conn      *ftp.ServerConn
	mu        sync.Mutex
	connected bool
}

// NewFTPClient creates a new FTP client
func NewFTPClient(config RemoteStorageConfig) *FTPClient {
	return &FTPClient{
		config: config,
	}
}

// Connect establishes FTP connection
func (c *FTPClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return nil
	}

	// Debug: log config
	log.Printf("[FTP] Connecting to %s:%d, username=%q, sharePath=%q",
		c.config.Server, c.config.Port, c.config.Username, c.config.SharePath)

	// Default port
	port := c.config.Port
	if port == 0 {
		port = 21
	}

	// FTP connection options
	options := []ftp.DialOption{
		ftp.DialWithTimeout(30 * time.Second),
	}

	// Connect to FTP server
	addr := fmt.Sprintf("%s:%d", c.config.Server, port)
	conn, err := ftp.Dial(addr, options...)
	if err != nil {
		return fmt.Errorf("FTP connection failed: %w", err)
	}

	// Login - support anonymous access
	username := c.config.Username
	password := c.config.Password
	// Use anonymous login if username is empty or explicitly "anonymous"
	if username == "" || strings.ToLower(username) == "anonymous" {
		username = "anonymous"
		password = "anonymous@example.com"
		log.Printf("[FTP] Using anonymous login")
	}
	err = conn.Login(username, password)
	if err != nil {
		conn.Quit()
		return fmt.Errorf("FTP login failed: %w", err)
	}

	// Change to initial directory if SharePath is specified
	if c.config.SharePath != "" {
		log.Printf("[FTP] Changing to directory: %s", c.config.SharePath)
		err = conn.ChangeDir(c.config.SharePath)
		if err != nil {
			conn.Quit()
			return fmt.Errorf("failed to change to initial directory %s: %w", c.config.SharePath, err)
		}
	}

	c.conn = conn
	c.connected = true
	log.Printf("[FTP] Connected successfully")
	return nil
}

// Disconnect closes the FTP connection
func (c *FTPClient) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected || c.conn == nil {
		return nil
	}

	err := c.conn.Quit()
	c.connected = false
	c.conn = nil
	return err
}

// List returns files in the specified directory
func (c *FTPClient) List(path string) ([]RemoteFileInfo, error) {
	if err := c.ensureConnected(); err != nil {
		return nil, err
	}

	// Normalize path
	path = normalizePathFTP(path)

	// If SharePath is set and path is root, use current directory (already changed to SharePath)
	// If path is "/" and no SharePath, list from root
	if path == "/" {
		if c.config.SharePath != "" {
			// Already in SharePath directory, list current dir
			path = "."
		}
	} else if path == "" {
		path = "."
	}

	entries, err := c.conn.List(path)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory: %w", err)
	}

	result := make([]RemoteFileInfo, 0, len(entries))
	for _, entry := range entries {
		// Skip . and ..
		if entry.Name == "." || entry.Name == ".." {
			continue
		}

		filePath := path + "/" + entry.Name
		if path == "/" {
			filePath = "/" + entry.Name
		} else if path == "." {
			filePath = entry.Name
		}

		result = append(result, RemoteFileInfo{
			Name:    entry.Name,
			Size:    int64(entry.Size),
			IsDir:   entry.Type == ftp.EntryTypeFolder,
			ModTime: entry.Time,
			Path:    filePath,
		})
	}
	return result, nil
}

// Download retrieves a file from remote storage
func (c *FTPClient) Download(path string) (io.ReadCloser, error) {
	if err := c.ensureConnected(); err != nil {
		return nil, err
	}

	path = normalizePathFTP(path)

	// Use Retr to get a reader
	r, err := c.conn.Retr(path)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	return r, nil
}

// Upload uploads a file to remote storage
func (c *FTPClient) Upload(path string, reader io.Reader, size int64) error {
	if err := c.ensureConnected(); err != nil {
		return err
	}

	path = normalizePathFTP(path)

	// Change to parent directory if needed
	dir := dirFromPathFTP(path)
	if dir != "" && dir != "/" {
		c.conn.ChangeDir(dir)
	}

	// Upload file
	return c.conn.Stor(path, reader)
}

// Delete removes a file or directory
func (c *FTPClient) Delete(path string) error {
	if err := c.ensureConnected(); err != nil {
		return err
	}

	path = normalizePathFTP(path)

	// Try to delete as file first
	err := c.conn.Delete(path)
	if err == nil {
		return nil
	}

	// If it's a directory, try RemoveDir
	// Note: RemoveDir only works on empty directories
	err = c.conn.RemoveDir(path)
	if err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}
	return nil
}

// Mkdir creates a directory
func (c *FTPClient) Mkdir(path string) error {
	if err := c.ensureConnected(); err != nil {
		return err
	}
	path = normalizePathFTP(path)
	return c.conn.MakeDir(path)
}

// TestConnection tests the connection
func (c *FTPClient) TestConnection() error {
	if err := c.Connect(); err != nil {
		return err
	}
	// Try to list root directory
	_, err := c.List("/")
	return err
}

// IsConnected returns connection status
func (c *FTPClient) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected
}

// ensureConnected checks and establishes connection if needed
func (c *FTPClient) ensureConnected() error {
	c.mu.Lock()
	connected := c.connected
	c.mu.Unlock()

	if connected {
		return nil
	}
	return c.Connect()
}

// normalizePathFTP ensures consistent path format for FTP
func normalizePathFTP(path string) string {
	if path == "" {
		return "."
	}
	// Convert backslashes to forward slashes
	path = strings.ReplaceAll(path, "\\", "/")
	// Remove trailing slash
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	return path
}

// dirFromPathFTP extracts directory from file path
func dirFromPathFTP(path string) string {
	idx := strings.LastIndex(path, "/")
	if idx <= 0 {
		return ""
	}
	return path[:idx]
}

// Ensure FTPClient implements RemoteClient
var _ RemoteClient = (*FTPClient)(nil)
