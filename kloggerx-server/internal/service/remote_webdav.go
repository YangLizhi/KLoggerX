package service

import (
	"io"
	"strconv"
	"strings"

	"github.com/studio-b12/gowebdav"
)

// WebDAVClient implements RemoteClient for WebDAV protocol
type WebDAVClient struct {
	config    RemoteStorageConfig
	client    *gowebdav.Client
	connected bool
}

// NewWebDAVClient creates a new WebDAV client
func NewWebDAVClient(config RemoteStorageConfig) *WebDAVClient {
	return &WebDAVClient{config: config}
}

// Connect establishes connection to WebDAV server
func (c *WebDAVClient) Connect() error {
	// Build the base URL
	scheme := "http"
	if c.config.UseHTTPS || c.config.Port == 443 {
		scheme = "https"
	}

	// Construct the URL
	var url string
	if c.config.Port > 0 && c.config.Port != 80 && c.config.Port != 443 {
		url = scheme + "://" + c.config.Server + ":" + strconv.Itoa(c.config.Port)
	} else {
		url = scheme + "://" + c.config.Server
	}

	// Add root path if specified
	if c.config.RootPath != "" {
		url = strings.TrimSuffix(url, "/") + "/" + strings.TrimPrefix(c.config.RootPath, "/")
	}

	// Support anonymous access - use empty credentials if username is not provided
	username := c.config.Username
	password := c.config.Password
	// gowebdav will handle empty credentials for anonymous access

	c.client = gowebdav.NewClient(url, username, password)

	// Test connection
	err := c.client.Connect()
	if err != nil {
		return err
	}

	c.connected = true
	return nil
}

// Disconnect closes the connection
func (c *WebDAVClient) Disconnect() error {
	c.connected = false
	return nil
}

// List returns list of files in the specified directory
func (c *WebDAVClient) List(path string) ([]RemoteFileInfo, error) {
	if !c.connected {
		if err := c.Connect(); err != nil {
			return nil, err
		}
	}

	path = strings.TrimPrefix(path, "/")
	if path == "" {
		path = "/"
	}

	files, err := c.client.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var result []RemoteFileInfo
	for _, f := range files {
		// Skip current and parent directory entries
		name := f.Name()
		if name == "." || name == ".." {
			continue
		}

		filePath := path
		if !strings.HasSuffix(filePath, "/") && filePath != "/" {
			filePath += "/"
		}
		if filePath == "/" {
			filePath += name
		} else {
			filePath += name
		}

		modTime := f.ModTime()

		result = append(result, RemoteFileInfo{
			Name:    name,
			Size:    f.Size(),
			IsDir:   f.IsDir(),
			ModTime: modTime,
			Path:    filePath,
		})
	}

	return result, nil
}

// Download retrieves a file from WebDAV
func (c *WebDAVClient) Download(path string) (io.ReadCloser, error) {
	if !c.connected {
		if err := c.Connect(); err != nil {
			return nil, err
		}
	}

	reader, err := c.client.ReadStream(path)
	if err != nil {
		return nil, err
	}

	return reader, nil
}

// Upload uploads a file to WebDAV
func (c *WebDAVClient) Upload(path string, reader io.Reader, size int64) error {
	if !c.connected {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	return c.client.WriteStream(path, reader, 0644)
}

// Delete removes a file or directory from WebDAV
func (c *WebDAVClient) Delete(path string) error {
	if !c.connected {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	return c.client.Remove(path)
}

// Mkdir creates a directory on WebDAV
func (c *WebDAVClient) Mkdir(path string) error {
	if !c.connected {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	return c.client.Mkdir(path, 0755)
}

// TestConnection tests if the connection is valid
func (c *WebDAVClient) TestConnection() error {
	if err := c.Connect(); err != nil {
		return err
	}
	defer c.Disconnect()

	// Try to list root directory
	_, err := c.List("/")
	return err
}

// IsConnected returns current connection status
func (c *WebDAVClient) IsConnected() bool {
	return c.connected
}
