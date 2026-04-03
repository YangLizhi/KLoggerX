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

// TencentClient implements RemoteClient for Tencent WeDrive (腾讯微盘)
type TencentClient struct {
	config    RemoteStorageConfig
	connected bool
	userID    string
	orgID     string
}

// TencentFileInfo represents file info from Tencent API
type TencentFileInfo struct {
	FileID      string `json:"file_id"`
	FileName    string `json:"file_name"`
	FileSize    int64  `json:"file_size"`
	FileType    int    `json:"file_type"` // 1=folder, 2=file
	CreatedTime string `json:"created_time"`
	ModifiedTime string `json:"modified_time"`
	ParentID    string `json:"parent_id"`
}

// TencentListResponse represents list response
type TencentListResponse struct {
	ErrCode int                `json:"err_code"`
	ErrMsg  string             `json:"err_msg"`
	Data    struct {
		FileList []TencentFileInfo `json:"file_list"`
		NextPage string            `json:"next_page"`
	} `json:"data"`
}

// TencentUserResponse represents user info response
type TencentUserResponse struct {
	ErrCode int `json:"err_code"`
	ErrMsg  string `json:"err_msg"`
	Data    struct {
		UserID string `json:"user_id"`
		OrgID  string `json:"org_id"`
	} `json:"data"`
}

// NewTencentClient creates a new Tencent WeDrive client
func NewTencentClient(config RemoteStorageConfig) *TencentClient {
	return &TencentClient{config: config}
}

// Connect validates the access token and gets user info
func (c *TencentClient) Connect() error {
	if c.config.AccessToken == "" {
		return fmt.Errorf("access token is required for Tencent WeDrive")
	}

	// Get user info
	data := map[string]interface{}{}
	body, _ := json.Marshal(data)

	req, _ := http.NewRequest("POST", "https://pan.tencent.com/api/v1/user/info", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+c.config.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Tencent WeDrive: %v", err)
	}
	defer resp.Body.Close()

	var result TencentUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	if result.ErrCode != 0 {
		return fmt.Errorf("Tencent API error: %s (code: %d)", result.ErrMsg, result.ErrCode)
	}

	c.userID = result.Data.UserID
	c.orgID = result.Data.OrgID
	c.connected = true
	return nil
}

// Disconnect closes the connection
func (c *TencentClient) Disconnect() error {
	c.connected = false
	return nil
}

// List returns list of files in the specified directory
func (c *TencentClient) List(dirPath string) ([]RemoteFileInfo, error) {
	if !c.connected {
		if err := c.Connect(); err != nil {
			return nil, err
		}
	}

	// Get directory ID from path
	dirID := "" // Empty means root
	if dirPath != "" && dirPath != "/" {
		fileID, err := c.getFileIDByPath(dirPath)
		if err != nil {
			return nil, err
		}
		dirID = fileID
	}

	// List files
	data := map[string]interface{}{
		"folder_id": dirID,
		"page_size": 100,
	}
	body, _ := json.Marshal(data)

	req, _ := http.NewRequest("POST", "https://pan.tencent.com/api/v1/file/list", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+c.config.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %v", err)
	}
	defer resp.Body.Close()

	var result TencentListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("Tencent API error: %s", result.ErrMsg)
	}

	var files []RemoteFileInfo
	for _, f := range result.Data.FileList {
		modTime, _ := time.Parse(time.RFC3339, f.ModifiedTime)
		files = append(files, RemoteFileInfo{
			Name:    f.FileName,
			Size:    f.FileSize,
			IsDir:   f.FileType == 1,
			ModTime: modTime,
			Path:    f.FileID,
		})
	}

	return files, nil
}

// getFileIDByPath gets file ID by path
func (c *TencentClient) getFileIDByPath(filePath string) (string, error) {
	parts := strings.Split(strings.Trim(filePath, "/"), "/")
	currentID := "" // root

	for _, part := range parts {
		if part == "" {
			continue
		}

		// List files in current directory
		data := map[string]interface{}{
			"folder_id": currentID,
			"page_size": 100,
		}
		body, _ := json.Marshal(data)

		req, _ := http.NewRequest("POST", "https://pan.tencent.com/api/v1/file/list", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+c.config.AccessToken)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}

		var result TencentListResponse
		json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()

		found := false
		for _, f := range result.Data.FileList {
			if f.FileName == part {
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

// Download retrieves a file from Tencent WeDrive
func (c *TencentClient) Download(fileID string) (io.ReadCloser, error) {
	if !c.connected {
		if err := c.Connect(); err != nil {
			return nil, err
		}
	}

	// Get download URL
	data := map[string]interface{}{
		"file_id": fileID,
	}
	body, _ := json.Marshal(data)

	req, _ := http.NewRequest("POST", "https://pan.tencent.com/api/v1/file/download", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+c.config.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get download URL: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		ErrCode  int    `json:"err_code"`
		ErrMsg   string `json:"err_msg"`
		Data     struct {
			DownloadURL string `json:"download_url"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("failed to get download URL: %s", result.ErrMsg)
	}

	// Download the file
	dlResp, err := http.Get(result.Data.DownloadURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %v", err)
	}

	return dlResp.Body, nil
}

// Upload uploads a file to Tencent WeDrive
func (c *TencentClient) Upload(uploadPath string, reader io.Reader, size int64) error {
	if !c.connected {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	// Get parent directory ID
	parentID := "" // root
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
			parentID = fileID
		}
	}

	// Get upload URL
	data := map[string]interface{}{
		"parent_id":   parentID,
		"file_name":   path.Base(uploadPath),
		"file_size":   size,
	}
	body, _ := json.Marshal(data)

	req, _ := http.NewRequest("POST", "https://pan.tencent.com/api/v1/file/upload", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+c.config.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get upload URL: %v", err)
	}

	var result struct {
		ErrCode    int    `json:"err_code"`
		ErrMsg     string `json:"err_msg"`
		Data       struct {
			UploadID  string `json:"upload_id"`
			UploadURL string `json:"upload_url"`
		} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()

	if result.ErrCode != 0 {
		return fmt.Errorf("failed to get upload URL: %s", result.ErrMsg)
	}

	// Upload to the upload URL
	if result.Data.UploadURL != "" {
		uploadReq, _ := http.NewRequest("PUT", result.Data.UploadURL, reader)
		uploadReq.Header.Set("Content-Length", fmt.Sprintf("%d", size))
		
		uploadResp, err := client.Do(uploadReq)
		if err != nil {
			return fmt.Errorf("failed to upload file: %v", err)
		}
		uploadResp.Body.Close()
	}

	return nil
}

// Delete removes a file or directory from Tencent WeDrive
func (c *TencentClient) Delete(fileID string) error {
	if !c.connected {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	data := map[string]interface{}{
		"file_ids": []string{fileID},
	}
	body, _ := json.Marshal(data)

	req, _ := http.NewRequest("POST", "https://pan.tencent.com/api/v1/file/delete", bytes.NewReader(body))
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

// Mkdir creates a directory on Tencent WeDrive
func (c *TencentClient) Mkdir(dirPath string) error {
	if !c.connected {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	// Create parent directories if needed
	parts := strings.Split(strings.Trim(dirPath, "/"), "/")
	currentParentID := ""

	for _, part := range parts {
		if part == "" {
			continue
		}

		// Check if directory exists
		fileID, _ := c.getFileIDByPath(part)
		if fileID != "" {
			currentParentID = fileID
			continue
		}

		// Create directory
		data := map[string]interface{}{
			"parent_id":  currentParentID,
			"name":       part,
			"type":       "folder",
		}
		body, _ := json.Marshal(data)

		req, _ := http.NewRequest("POST", "https://pan.tencent.com/api/v1/file/create", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+c.config.AccessToken)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}

		var result struct {
			ErrCode int    `json:"err_code"`
			ErrMsg  string `json:"err_msg"`
			Data    struct {
				FileID string `json:"file_id"`
			} `json:"data"`
		}
		json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()

		if result.ErrCode != 0 {
			return fmt.Errorf("failed to create directory: %s", result.ErrMsg)
		}

		currentParentID = result.Data.FileID
	}

	return nil
}

// TestConnection tests if the connection is valid
func (c *TencentClient) TestConnection() error {
	return c.Connect()
}

// IsConnected returns current connection status
func (c *TencentClient) IsConnected() bool {
	return c.connected
}
