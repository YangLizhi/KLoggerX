package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"
)

// AliyunClient implements RemoteClient for Aliyun Drive (阿里云盘)
type AliyunClient struct {
	config      RemoteStorageConfig
	connected   bool
	driveID     string
	uploadToken string
}

// AliyunFileInfo represents file info from Aliyun API
type AliyunFileInfo struct {
	DriveID       string `json:"drive_id"`
	FileID        string `json:"file_id"`
	ParentFileID  string `json:"parent_file_id"`
	Name          string `json:"name"`
	Size          int64  `json:"size"`
	Type          string `json:"type"` // file or folder
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// AliyunListResponse represents list response
type AliyunListResponse struct {
	Items         []AliyunFileInfo `json:"items"`
	NextMarker    string           `json:"next_marker"`
	StatusCode    int              `json:"status_code"`
	Code          string           `json:"code"`
	Message       string           `json:"message"`
}

// AliyunUserResponse represents user info response
type AliyunUserResponse struct {
	StatusCode int `json:"status_code"`
	User       struct {
		DriveID string `json:"drive_id"`
		UserID  string `json:"user_id"`
	} `json:"user"`
}

// AliyunUploadResponse represents upload response
type AliyunUploadResponse struct {
	StatusCode    int    `json:"status_code"`
	UploadID      string `json:"upload_id"`
	UploadURL     string `json:"upload_url"`
	FileID        string `json:"file_id"`
	Code          string `json:"code"`
	Message       string `json:"message"`
}

// NewAliyunClient creates a new Aliyun Drive client
func NewAliyunClient(config RemoteStorageConfig) *AliyunClient {
	return &AliyunClient{config: config}
}

// Connect validates the access token and gets drive info
func (c *AliyunClient) Connect() error {
	if c.config.AccessToken == "" && c.config.RefreshToken == "" {
		return fmt.Errorf("access token or refresh token is required for Aliyun Drive")
	}

	// If we only have refresh token, refresh access token
	if c.config.AccessToken == "" && c.config.RefreshToken != "" {
		if err := c.refreshAccessToken(); err != nil {
			return err
		}
	}

	// Get user info and drive ID
	req, _ := http.NewRequest("GET", "https://api.alipan.com/adrive/v1.0/user/getDriveInfo", nil)
	req.Header.Set("Authorization", "Bearer "+c.config.AccessToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Aliyun Drive: %v", err)
	}
	defer resp.Body.Close()

	var result AliyunUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	if result.StatusCode != 0 || result.User.DriveID == "" {
		return fmt.Errorf("failed to get drive info")
	}

	c.driveID = result.User.DriveID
	c.connected = true
	return nil
}

// refreshAccessToken refreshes the access token using refresh token
func (c *AliyunClient) refreshAccessToken() error {
	data := map[string]string{
		"refresh_token": c.config.RefreshToken,
		"grant_type":    "refresh_token",
	}
	body, _ := json.Marshal(data)

	req, _ := http.NewRequest("POST", "https://api.alipan.com/oauth/access_token", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to refresh token: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		StatusCode   int    `json:"status_code"`
		Code         string `json:"code"`
		Message      string `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	if result.StatusCode != 0 {
		return fmt.Errorf("token refresh failed: %s", result.Message)
	}

	c.config.AccessToken = result.AccessToken
	c.config.RefreshToken = result.RefreshToken
	return nil
}

// Disconnect closes the connection
func (c *AliyunClient) Disconnect() error {
	c.connected = false
	return nil
}

// List returns list of files in the specified directory
func (c *AliyunClient) List(dirPath string) ([]RemoteFileInfo, error) {
	if !c.connected {
		if err := c.Connect(); err != nil {
			return nil, err
		}
	}

	// Get parent file ID from path
	parentFileID := "root" // root directory
	if dirPath != "" && dirPath != "/" {
		// Need to find the file ID for the path
		fileID, err := c.getFileIDByPath(dirPath)
		if err != nil {
			return nil, err
		}
		parentFileID = fileID
	}

	// List files
	data := map[string]interface{}{
		"drive_id":       c.driveID,
		"parent_file_id": parentFileID,
		"limit":          100,
		"order_by":       "updated_at",
		"order_direction": "DESC",
	}
	body, _ := json.Marshal(data)

	req, _ := http.NewRequest("POST", "https://api.alipan.com/adrive/v1.0/openFile/list", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+c.config.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %v", err)
	}
	defer resp.Body.Close()

	var result AliyunListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	if result.StatusCode != 0 {
		return nil, fmt.Errorf("API error: %s", result.Message)
	}

	var files []RemoteFileInfo
	for _, f := range result.Items {
		modTime, _ := time.Parse(time.RFC3339, f.UpdatedAt)
		files = append(files, RemoteFileInfo{
			Name:    f.Name,
			Size:    f.Size,
			IsDir:   f.Type == "folder",
			ModTime: modTime,
			Path:    f.FileID, // Use file ID as path for operations
		})
	}

	return files, nil
}

// getFileIDByPath gets file ID by path
func (c *AliyunClient) getFileIDByPath(filePath string) (string, error) {
	parts := strings.Split(strings.Trim(filePath, "/"), "/")
	currentID := "root"

	for _, part := range parts {
		if part == "" {
			continue
		}

		// List files in current directory
		data := map[string]interface{}{
			"drive_id":       c.driveID,
			"parent_file_id": currentID,
			"limit":          100,
		}
		body, _ := json.Marshal(data)

		req, _ := http.NewRequest("POST", "https://api.alipan.com/adrive/v1.0/openFile/list", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+c.config.AccessToken)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}

		var result AliyunListResponse
		json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()

		found := false
		for _, f := range result.Items {
			if f.Name == part {
				currentID = f.FileID
				found = true
				break
			}
		}

		if !found {
			return "", fmt.Errorf("path not found: %s", part)
		}
	}

	return currentID, nil
}

// Download retrieves a file from Aliyun Drive
func (c *AliyunClient) Download(fileID string) (io.ReadCloser, error) {
	if !c.connected {
		if err := c.Connect(); err != nil {
			return nil, err
		}
	}

	// Get download URL
	data := map[string]interface{}{
		"drive_id": c.driveID,
		"file_id":  fileID,
	}
	body, _ := json.Marshal(data)

	req, _ := http.NewRequest("POST", "https://api.alipan.com/adrive/v1.0/openFile/getDownloadUrl", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+c.config.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get download URL: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		URL        string `json:"url"`
		StatusCode int    `json:"status_code"`
		Message    string `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	if result.StatusCode != 0 {
		return nil, fmt.Errorf("failed to get download URL: %s", result.Message)
	}

	// Download the file
	dlResp, err := http.Get(result.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %v", err)
	}

	return dlResp.Body, nil
}

// Upload uploads a file to Aliyun Drive
func (c *AliyunClient) Upload(uploadPath string, reader io.Reader, size int64) error {
	if !c.connected {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	// Get parent directory ID
	parentFileID := "root"
	dir := path.Dir(uploadPath)
	if dir != "" && dir != "/" {
		fileID, err := c.getFileIDByPath(dir)
		if err != nil {
			// Try to create the directory
			if err := c.Mkdir(dir); err != nil {
				return err
			}
			fileID, _ = c.getFileIDByPath(dir)
		}
		if fileID != "" {
			parentFileID = fileID
		}
	}

	// Create file with upload
	data := map[string]interface{}{
		"drive_id":       c.driveID,
		"parent_file_id": parentFileID,
		"name":           path.Base(uploadPath),
		"type":           "file",
		"size":           size,
		"check_name_mode": "auto_rename",
	}
	body, _ := json.Marshal(data)

	req, _ := http.NewRequest("POST", "https://api.alipan.com/adrive/v1.0/openFile/create", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+c.config.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}

	var result AliyunUploadResponse
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()

	if result.StatusCode != 0 {
		return fmt.Errorf("failed to create file: %s", result.Message)
	}

	// Upload to the upload URL
	if result.UploadURL != "" {
		uploadReq, _ := http.NewRequest("PUT", result.UploadURL, reader)
		uploadReq.Header.Set("Content-Length", fmt.Sprintf("%d", size))
		
		uploadResp, err := client.Do(uploadReq)
		if err != nil {
			return fmt.Errorf("failed to upload file: %v", err)
		}
		uploadResp.Body.Close()
	}

	return nil
}

// Delete removes a file or directory from Aliyun Drive
func (c *AliyunClient) Delete(fileID string) error {
	if !c.connected {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	data := map[string]interface{}{
		"drive_id": c.driveID,
		"file_id":  fileID,
	}
	body, _ := json.Marshal(data)

	req, _ := http.NewRequest("POST", "https://api.alipan.com/adrive/v1.0/openFile/delete", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+c.config.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}
	defer resp.Body.Close()

	return nil
}

// Mkdir creates a directory on Aliyun Drive
func (c *AliyunClient) Mkdir(dirPath string) error {
	if !c.connected {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	// Create parent directories if needed
	parts := strings.Split(strings.Trim(dirPath, "/"), "/")
	currentParentID := "root"
	currentPath := ""

	for _, part := range parts {
		if part == "" {
			continue
		}
		currentPath += "/" + part

		// Check if directory exists
		fileID, _ := c.getFileIDByPath(currentPath)
		if fileID != "" {
			currentParentID = fileID
			continue
		}

		// Create directory
		data := map[string]interface{}{
			"drive_id":       c.driveID,
			"parent_file_id": currentParentID,
			"name":           part,
			"type":           "folder",
			"check_name_mode": "auto_rename",
		}
		body, _ := json.Marshal(data)

		req, _ := http.NewRequest("POST", "https://api.alipan.com/adrive/v1.0/openFile/create", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+c.config.AccessToken)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}

		var result struct {
			FileID      string `json:"file_id"`
			StatusCode  int    `json:"status_code"`
			Message     string `json:"message"`
		}
		json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()

		if result.StatusCode != 0 {
			return fmt.Errorf("failed to create directory: %s", result.Message)
		}

		currentParentID = result.FileID
	}

	return nil
}

// TestConnection tests if the connection is valid
func (c *AliyunClient) TestConnection() error {
	return c.Connect()
}

// IsConnected returns current connection status
func (c *AliyunClient) IsConnected() bool {
	return c.connected
}
