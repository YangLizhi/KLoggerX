package service

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// NFSClient implements RemoteClient for NFS protocol
// Note: NFS requires system-level mount support
type NFSClient struct {
	config      RemoteStorageConfig
	connected   bool
	mountPoint  string
	basePath    string
}

// NewNFSClient creates a new NFS client
func NewNFSClient(config RemoteStorageConfig) *NFSClient {
	return &NFSClient{config: config}
}

// Connect mounts the NFS share
func (c *NFSClient) Connect() error {
	// Create mount point if not exists
	c.mountPoint = fmt.Sprintf("/tmp/nfs-mount-%d", time.Now().UnixNano())
	if err := os.MkdirAll(c.mountPoint, 0755); err != nil {
		return fmt.Errorf("failed to create mount point: %v", err)
	}

	// Build NFS server address
	server := c.config.Server
	if c.config.Port > 0 && c.config.Port != 2049 {
		server = fmt.Sprintf("%s:%d", c.config.Server, c.config.Port)
	}

	// Build the export path
	exportPath := c.config.SharePath
	if exportPath == "" {
		exportPath = "/"
	}
	
	nfsSource := fmt.Sprintf("%s:%s", server, exportPath)

	// Check if already mounted
	mountCmd := exec.Command("mountpoint", "-q", c.mountPoint)
	if mountCmd.Run() != nil {
		// Not mounted, so mount it
		mountArgs := []string{"-t", "nfs4", nfsSource, c.mountPoint}
		if c.config.Username != "" || c.config.Password != "" {
			// Add options for authentication (for NFSv4 with Kerberos)
			options := []string{}
			if c.config.Username != "" {
				options = append(options, fmt.Sprintf("username=%s", c.config.Username))
			}
			if c.config.Password != "" {
				options = append(options, fmt.Sprintf("password=%s", c.config.Password))
			}
			if len(options) > 0 {
				mountArgs = append([]string{"-o", strings.Join(options, ",")}, mountArgs...)
			}
		}

		mountCmd = exec.Command("mount", mountArgs...)
		mountCmd.Stdout = os.Stdout
		mountCmd.Stderr = os.Stderr
		
		if err := mountCmd.Run(); err != nil {
			os.RemoveAll(c.mountPoint)
			return fmt.Errorf("failed to mount NFS share: %v", err)
		}
	}

	c.basePath = c.mountPoint
	if c.config.RootPath != "" {
		c.basePath = filepath.Join(c.mountPoint, c.config.RootPath)
	}

	c.connected = true
	return nil
}

// Disconnect unmounts the NFS share
func (c *NFSClient) Disconnect() error {
	if c.mountPoint != "" {
		// Unmount
		umountCmd := exec.Command("umount", c.mountPoint)
		umountCmd.Run()
		
		// Remove mount point
		os.RemoveAll(c.mountPoint)
		c.mountPoint = ""
	}
	
	c.connected = false
	return nil
}

// List returns list of files in the specified directory
func (c *NFSClient) List(path string) ([]RemoteFileInfo, error) {
	if !c.connected {
		if err := c.Connect(); err != nil {
			return nil, err
		}
	}

	fullPath := filepath.Join(c.basePath, path)
	
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %v", err)
	}

	var files []RemoteFileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		filePath := filepath.Join(path, entry.Name())
		if strings.HasPrefix(filePath, "/") {
			filePath = filePath[1:]
		}

		files = append(files, RemoteFileInfo{
			Name:    entry.Name(),
			Size:    info.Size(),
			IsDir:   entry.IsDir(),
			ModTime: info.ModTime(),
			Path:    filePath,
		})
	}

	return files, nil
}

// Download retrieves a file from NFS
func (c *NFSClient) Download(filePath string) (io.ReadCloser, error) {
	if !c.connected {
		if err := c.Connect(); err != nil {
			return nil, err
		}
	}

	fullPath := filepath.Join(c.basePath, filePath)
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}

	return file, nil
}

// Upload uploads a file to NFS
func (c *NFSClient) Upload(filePath string, reader io.Reader, size int64) error {
	if !c.connected {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	fullPath := filepath.Join(c.basePath, filePath)
	
	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create parent directories: %v", err)
	}

	// Create the file
	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Copy content
	_, err = io.Copy(file, reader)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

// Delete removes a file or directory from NFS
func (c *NFSClient) Delete(filePath string) error {
	if !c.connected {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	fullPath := filepath.Join(c.basePath, filePath)
	
	return os.RemoveAll(fullPath)
}

// Mkdir creates a directory on NFS
func (c *NFSClient) Mkdir(dirPath string) error {
	if !c.connected {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	fullPath := filepath.Join(c.basePath, dirPath)
	
	return os.MkdirAll(fullPath, 0755)
}

// TestConnection tests if the connection is valid
func (c *NFSClient) TestConnection() error {
	if err := c.Connect(); err != nil {
		return err
	}
	defer c.Disconnect()

	// Try to list root directory
	_, err := c.List("/")
	return err
}

// IsConnected returns current connection status
func (c *NFSClient) IsConnected() bool {
	return c.connected
}
