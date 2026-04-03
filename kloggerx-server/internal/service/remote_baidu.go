package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

// BaiduClient implements RemoteClient for Baidu Netdisk
type BaiduClient struct {
	config    RemoteStorageConfig
	connected bool
}

// BaiduFileInfo represents file info from Baidu API
type BaiduFileInfo struct {
	FsID     int64  `json:"fs_id"`
	Path     string `json:"path"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	IsDir    int    `json:"isdir"`
	Mtime    int64  `json:"server_mtime"`
	Ctime    int64  `json:"server_ctime"`
}

// BaiduListResponse represents list response from Baidu API
type BaiduListResponse struct {
	Errno   int            `json:"errno"`
	Errmsg  string         `json:"errmsg"`
	List    []BaiduFileInfo `json:"list"`
}

// BaiduDownloadResponse represents download info response
type BaiduDownloadResponse struct {
	Errno  int    `json:"errno"`
	Errmsg string `json:"errmsg"`
	Dlink  string `json:"dlink"`
}

// NewBaiduClient creates a new Baidu Netdisk client
func NewBaiduClient(config RemoteStorageConfig) *BaiduClient {
	return &BaiduClient{config: config}
}

// Connect validates the access token
func (c *BaiduClient) Connect() error {
	if c.config.AccessToken == "" {
		return fmt.Errorf("access token is required for Baidu Netdisk")
	}

	// Verify token by getting user info
	url := "https://pan.baidu.com/rest/2.0/xpan/nas?method=uinfo&access_token=" + c.config.AccessToken
	
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to connect to Baidu: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Errno   int    `json:"errno"`
		Errmsg  string `json:"errmsg"`
		BaiduName string `json:"baidu_name"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	if result.Errno != 0 {
		return fmt.Errorf("Baidu API error: %s (errno: %d)", result.Errmsg, result.Errno)
	}

	c.connected = true
	return nil
}

// Disconnect closes the connection
func (c *BaiduClient) Disconnect() error {
	c.connected = false
	return nil
}

// List returns list of files in the specified directory
func (c *BaiduClient) List(dirPath string) ([]RemoteFileInfo, error) {
	if !c.connected {
		if err := c.Connect(); err != nil {
			return nil, err
		}
	}

	// Normalize path
	if dirPath == "" || dirPath == "/" {
		dirPath = "/"
	} else if !strings.HasPrefix(dirPath, "/") {
		dirPath = "/" + dirPath
	}

	// Build request URL
	apiURL := fmt.Sprintf("https://pan.baidu.com/rest/2.0/xpan/file?method=list&dir=%s&order=time&access_token=%s",
		url.QueryEscape(dirPath), c.config.AccessToken)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %v", err)
	}
	defer resp.Body.Close()

	var result BaiduListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	if result.Errno != 0 {
		return nil, fmt.Errorf("Baidu API error: %s (errno: %d)", result.Errmsg, result.Errno)
	}

	var files []RemoteFileInfo
	for _, f := range result.List {
		files = append(files, RemoteFileInfo{
			Name:    f.Filename,
			Size:    f.Size,
			IsDir:   f.IsDir == 1,
			ModTime: time.Unix(f.Mtime, 0),
			Path:    f.Path,
		})
	}

	return files, nil
}

// Download retrieves a file from Baidu Netdisk
func (c *BaiduClient) Download(filePath string) (io.ReadCloser, error) {
	if !c.connected {
		if err := c.Connect(); err != nil {
			return nil, err
		}
	}

	// Get download link
	apiURL := fmt.Sprintf("https://pan.baidu.com/rest/2.0/xpan/multimedia?method=filemetas&fsids=[%s]&dlink=1&access_token=%s",
		url.QueryEscape(filePath), c.config.AccessToken)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get download link: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Errno int `json:"errno"`
		List  []struct {
			Dlink string `json:"dlink"`
		} `json:"list"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	if result.Errno != 0 || len(result.List) == 0 {
		return nil, fmt.Errorf("failed to get download link")
	}

	// Download file using the dlink
	downloadURL := result.List[0].Dlink + "&access_token=" + c.config.AccessToken
	
	dlResp, err := http.Get(downloadURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %v", err)
	}

	return dlResp.Body, nil
}

// Upload uploads a file to Baidu Netdisk
func (c *BaiduClient) Upload(uploadPath string, reader io.Reader, size int64) error {
	if !c.connected {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	// Normalize path
	if !strings.HasPrefix(uploadPath, "/") {
		uploadPath = "/" + uploadPath
	}

	// Get upload URL
	apiURL := fmt.Sprintf("https://pan.baidu.com/rest/2.0/xpan/file?method=precreate&access_token=%s", c.config.AccessToken)
	
	// First, precreate the file
	precreateData := url.Values{}
	precreateData.Set("path", uploadPath)
	precreateData.Set("size", fmt.Sprintf("%d", size))
	precreateData.Set("isdir", "0")
	precreateData.Set("autoinit", "1")

	resp, err := http.Post(apiURL, "application/x-www-form-urlencoded", strings.NewReader(precreateData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to precreate file: %v", err)
	}
	resp.Body.Close()

	// Upload the file using superfile2 API
	uploadURL := fmt.Sprintf("https://d.pcs.baidu.com/rest/2.0/pcs/superfile2?method=upload&access_token=%s&path=%s",
		c.config.AccessToken, url.QueryEscape(uploadPath))

	// Create multipart form
	body := &bytes.Buffer{}
	body.WriteString(fmt.Sprintf("--boundary\r\nContent-Disposition: form-data; name=\"file\"; filename=\"%s\"\r\n\r\n", path.Base(uploadPath)))
	io.Copy(body, reader)
	body.WriteString("\r\n--boundary--\r\n")

	req, err := http.NewRequest("POST", uploadURL, body)
	if err != nil {
		return fmt.Errorf("failed to create upload request: %v", err)
	}
	req.Header.Set("Content-Type", "multipart/form-data; boundary=boundary")

	client := &http.Client{Timeout: 30 * time.Minute}
	uploadResp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload file: %v", err)
	}
	defer uploadResp.Body.Close()

	if uploadResp.StatusCode != 200 {
		return fmt.Errorf("upload failed with status: %d", uploadResp.StatusCode)
	}

	return nil
}

// Delete removes a file or directory from Baidu Netdisk
func (c *BaiduClient) Delete(filePath string) error {
	if !c.connected {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	apiURL := fmt.Sprintf("https://pan.baidu.com/rest/2.0/xpan/file?method=filemanager&opera=delete&access_token=%s",
		c.config.AccessToken)

	data := url.Values{}
	data.Set("filelist", fmt.Sprintf("[\"%s\"]", filePath))

	resp, err := http.Post(apiURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Errno int `json:"errno"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	if result.Errno != 0 {
		return fmt.Errorf("delete failed with errno: %d", result.Errno)
	}

	return nil
}

// Mkdir creates a directory on Baidu Netdisk
func (c *BaiduClient) Mkdir(dirPath string) error {
	if !c.connected {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	if !strings.HasPrefix(dirPath, "/") {
		dirPath = "/" + dirPath
	}

	apiURL := fmt.Sprintf("https://pan.baidu.com/rest/2.0/xpan/file?method=mkdir&access_token=%s",
		c.config.AccessToken)

	data := url.Values{}
	data.Set("path", dirPath)

	resp, err := http.Post(apiURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Errno int `json:"errno"`
		Errmsg string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	if result.Errno != 0 {
		return fmt.Errorf("mkdir failed: %s (errno: %d)", result.Errmsg, result.Errno)
	}

	return nil
}

// TestConnection tests if the connection is valid
func (c *BaiduClient) TestConnection() error {
	return c.Connect()
}

// IsConnected returns current connection status
func (c *BaiduClient) IsConnected() bool {
	return c.connected
}
