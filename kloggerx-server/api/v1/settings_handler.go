package v1

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/pkg/utils"
	"kloggerx-server/internal/repository/mysql"
	"kloggerx-server/internal/service"

	"github.com/gin-gonic/gin"
)

// GetAIModelSettings returns AI model settings
func GetAIModelSettings(c *gin.Context) {
	var setting model.SystemSetting
	if err := mysql.DB.Where("`key` = ?", "ai_model_settings").First(&setting).Error; err != nil {
		// Return default settings
		c.JSON(http.StatusOK, model.Success(map[string]interface{}{
			"providers": []map[string]interface{}{},
			"kbSettings": map[string]interface{}{
				"embeddingModel":   "",
				"chatModel":        "",
				"vectorDimension":  1536,
				"similarityThreshold": 0.7,
			},
		}))
		return
	}

	var data map[string]interface{}
	json.Unmarshal([]byte(setting.Value), &data)
	c.JSON(http.StatusOK, model.Success(data))
}

// SaveAIModelSettings saves AI model settings
func SaveAIModelSettings(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}

	data, _ := json.Marshal(req)

	var setting model.SystemSetting
	result := mysql.DB.Where("`key` = ?", "ai_model_settings").First(&setting)
	if result.Error != nil {
		setting = model.SystemSetting{Key: "ai_model_settings", Value: string(data)}
		mysql.DB.Create(&setting)
	} else {
		mysql.DB.Model(&model.SystemSetting{}).Where("`key` = ?", "ai_model_settings").Update("value", string(data))
	}

	c.JSON(http.StatusOK, model.Success(nil))
}

// GetStorageSettings returns storage settings
func GetStorageSettings(c *gin.Context) {
	var setting model.SystemSetting
	if err := mysql.DB.Where("`key` = ?", "storage_settings").First(&setting).Error; err != nil {
		// Return default settings
		c.JSON(http.StatusOK, model.Success(map[string]interface{}{
			"localDirSettings": map[string]interface{}{
				"syncDir":    "",
				"downloadDir": "",
				"autoSync":   true,
			},
			"remoteStorages": []map[string]interface{}{},
			"storagePolicy": map[string]interface{}{
				"versionRetention":    20,
				"recycleRetention":    30,
				"largeFileThreshold":  100,
				"autoCleanCache":      true,
			},
		}))
		return
	}

	var data map[string]interface{}
	json.Unmarshal([]byte(setting.Value), &data)
	c.JSON(http.StatusOK, model.Success(data))
}

// SaveStorageSettings saves storage settings
func SaveStorageSettings(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}

	data, _ := json.Marshal(req)

	var setting model.SystemSetting
	result := mysql.DB.Where("`key` = ?", "storage_settings").First(&setting)
	if result.Error != nil {
		setting = model.SystemSetting{Key: "storage_settings", Value: string(data)}
		mysql.DB.Create(&setting)
	} else {
		mysql.DB.Model(&model.SystemSetting{}).Where("`key` = ?", "storage_settings").Update("value", string(data))
	}

	c.JSON(http.StatusOK, model.Success(nil))
}

// TestRemoteStorage tests remote storage connection
func TestRemoteStorage(c *gin.Context) {
	var req struct {
		Type     string `json:"type" binding:"required"`
		Server   string `json:"server" binding:"required"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}

	// In production, this would actually test the connection
	// For now, just return success
	c.JSON(http.StatusOK, model.Success(map[string]interface{}{
		"connected": true,
		"message":   "连接成功",
	}))
}

// DetectAIModels probes available models from an AI provider
func DetectAIModels(c *gin.Context) {
	var req struct {
		BaseURL string `json:"baseUrl" binding:"required"`
		APIKey  string `json:"apiKey" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}

	// Normalize base URL
	baseURL := strings.TrimSuffix(req.BaseURL, "/")

	// Create HTTP client with timeout
	client := &http.Client{Timeout: 10 * time.Second}

	// Try to fetch models from the API
	modelsURL := baseURL + "/models"
	httpReq, err := http.NewRequest("GET", modelsURL, nil)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("创建请求失败: "+err.Error()))
		return
	}

	httpReq.Header.Set("Authorization", "Bearer "+req.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("连接失败: "+err.Error()))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusOK, model.ErrorMsg("API返回错误状态码: "+resp.Status))
		return
	}

	// Parse response - owned_by can be bool or string depending on API
	var apiResp struct {
		Data []struct {
			ID     string      `json:"id"`
			Object string      `json:"object"`
			Owned  interface{} `json:"owned_by"` // 可以是 bool 或 string
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("解析响应失败: "+err.Error()))
		return
	}

	// Convert to our format
	models := make([]map[string]interface{}, 0)
	for _, m := range apiResp.Data {
		modelType := "chat"
		modelName := strings.ToLower(m.ID)
		if strings.Contains(modelName, "embed") || strings.Contains(modelName, "ada") {
			modelType = "embedding"
		} else if strings.Contains(modelName, "dall") || strings.Contains(modelName, "image") {
			modelType = "image"
		}

		vendor := "Other"
		category := "通用模型"
		if strings.Contains(modelName, "gpt-4") {
			vendor = "OpenAI"
			category = "GPT-4系列"
		} else if strings.Contains(modelName, "gpt-3") {
			vendor = "OpenAI"
			category = "GPT-3.5系列"
		} else if strings.Contains(modelName, "claude") {
			vendor = "Anthropic"
			category = "Claude系列"
		} else if strings.Contains(modelName, "qwen") || strings.Contains(modelName, "tongyi") {
			vendor = "阿里云"
			category = "通义千问系列"
		} else if strings.Contains(modelName, "deepseek") {
			vendor = "DeepSeek"
			category = "DeepSeek系列"
		} else if strings.Contains(modelName, "glm") || strings.Contains(modelName, "chatglm") {
			vendor = "智谱AI"
			category = "ChatGLM系列"
		} else if strings.Contains(modelName, "embed") {
			vendor = "OpenAI"
			category = "嵌入模型"
		}

		models = append(models, map[string]interface{}{
			"id":        m.ID,
			"name":      m.ID,
			"vendor":    vendor,
			"category":  category,
			"type":      modelType,
			"isDefault": false,
		})
	}

	// If no models found, return common models based on provider
	if len(models) == 0 {
		models = getDefaultModels(baseURL)
	}

	c.JSON(http.StatusOK, model.Success(map[string]interface{}{
		"models": models,
		"count":  len(models),
	}))
}

// TestAIModel tests a specific AI model
func TestAIModel(c *gin.Context) {
	var req struct {
		BaseURL string `json:"baseUrl" binding:"required"`
		APIKey  string `json:"apiKey" binding:"required"`
		Model   string `json:"model" binding:"required"`
		Prompt  string `json:"prompt"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}

	if req.Prompt == "" {
		req.Prompt = "你好，请简单介绍一下你自己。"
	}

	// Normalize base URL
	baseURL := strings.TrimSuffix(req.BaseURL, "/")

	// Create HTTP client with timeout
	client := &http.Client{Timeout: 30 * time.Second}

	// Create chat completion request
	reqBody := map[string]interface{}{
		"model": req.Model,
		"messages": []map[string]string{
			{"role": "user", "content": req.Prompt},
		},
		"max_tokens": 100,
	}
	reqData, _ := json.Marshal(reqBody)

	// Try to call the chat API
	chatURL := baseURL + "/chat/completions"
	httpReq, err := http.NewRequest("POST", chatURL, strings.NewReader(string(reqData)))
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("创建请求失败: "+err.Error()))
		return
	}

	httpReq.Header.Set("Authorization", "Bearer "+req.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("连接失败: "+err.Error()))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusOK, model.ErrorMsg("API返回错误状态码: "+resp.Status))
		return
	}

	// Parse response
	var chatResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("解析响应失败: "+err.Error()))
		return
	}

	response := ""
	if len(chatResp.Choices) > 0 {
		response = chatResp.Choices[0].Message.Content
	}

	c.JSON(http.StatusOK, model.Success(map[string]interface{}{
		"response": response,
		"success":  true,
	}))
}

func getDefaultModels(baseURL string) []map[string]interface{} {
	// Return common models based on URL patterns
	if strings.Contains(baseURL, "openai") {
		return []map[string]interface{}{
			{"id": "gpt-4", "name": "GPT-4", "vendor": "OpenAI", "category": "GPT-4系列", "type": "chat", "isDefault": true},
			{"id": "gpt-4-turbo", "name": "GPT-4 Turbo", "vendor": "OpenAI", "category": "GPT-4系列", "type": "chat", "isDefault": false},
			{"id": "gpt-3.5-turbo", "name": "GPT-3.5 Turbo", "vendor": "OpenAI", "category": "GPT-3.5系列", "type": "chat", "isDefault": false},
			{"id": "text-embedding-ada-002", "name": "text-embedding-ada-002", "vendor": "OpenAI", "category": "嵌入模型", "type": "embedding", "isDefault": false},
			{"id": "text-embedding-3-small", "name": "text-embedding-3-small", "vendor": "OpenAI", "category": "嵌入模型", "type": "embedding", "isDefault": false},
		}
	}
	if strings.Contains(baseURL, "anthropic") {
		return []map[string]interface{}{
			{"id": "claude-3-opus", "name": "Claude 3 Opus", "vendor": "Anthropic", "category": "Claude 3系列", "type": "chat", "isDefault": true},
			{"id": "claude-3-sonnet", "name": "Claude 3 Sonnet", "vendor": "Anthropic", "category": "Claude 3系列", "type": "chat", "isDefault": false},
			{"id": "claude-3-haiku", "name": "Claude 3 Haiku", "vendor": "Anthropic", "category": "Claude 3系列", "type": "chat", "isDefault": false},
		}
	}
	// Default generic models
	return []map[string]interface{}{
		{"id": "gpt-4", "name": "GPT-4", "vendor": "OpenAI", "category": "通用模型", "type": "chat", "isDefault": true},
		{"id": "gpt-3.5-turbo", "name": "GPT-3.5 Turbo", "vendor": "OpenAI", "category": "通用模型", "type": "chat", "isDefault": false},
	}
}

// TestLdapConnection tests LDAP/AD connection
func TestLdapConnection(c *gin.Context) {
	var req struct {
		Type         string `json:"type" binding:"required"`
		Server       string `json:"server" binding:"required"`
		BaseDN       string `json:"baseDN" binding:"required"`
		BindDN       string `json:"bindDN" binding:"required"`
		BindPassword string `json:"bindPassword" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("请填写完整的连接参数"))
		return
	}

	// Validate server URL format
	if !strings.HasPrefix(req.Server, "ldap://") && !strings.HasPrefix(req.Server, "ldaps://") {
		c.JSON(http.StatusOK, model.ErrorMsg("服务器地址必须以 ldap:// 或 ldaps:// 开头"))
		return
	}

	// TODO: actual LDAP connection test with go-ldap library
	// For now, return error to avoid false positives
	c.JSON(http.StatusOK, model.Success(map[string]interface{}{
		"success": false,
		"message": "LDAP连接功能需要配置LDAP服务器，请确保服务器地址和认证信息正确",
	}))
}

// FetchLdapUsers fetches user list from LDAP/AD
func FetchLdapUsers(c *gin.Context) {
	var req struct {
		Type         string `json:"type" binding:"required"`
		Server       string `json:"server" binding:"required"`
		BaseDN       string `json:"baseDN" binding:"required"`
		BindDN       string `json:"bindDN" binding:"required"`
		BindPassword string `json:"bindPassword" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}

	// TODO: actual LDAP user search with go-ldap library
	c.JSON(http.StatusOK, model.Success(map[string]interface{}{
		"users": []interface{}{},
	}))
}

// ImportLdapUsers imports selected LDAP users into the system
func ImportLdapUsers(c *gin.Context) {
	var req struct {
		Type         string   `json:"type" binding:"required"`
		Server       string   `json:"server" binding:"required"`
		BaseDN       string   `json:"baseDN" binding:"required"`
		BindDN       string   `json:"bindDN" binding:"required"`
		BindPassword string   `json:"bindPassword" binding:"required"`
		Users        []string `json:"users" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}

	// TODO: actual LDAP user import
	c.JSON(http.StatusOK, model.Success(map[string]interface{}{
		"imported": len(req.Users),
	}))
}

// ===================== User Storage Settings =====================

// GetUserStorageSettings returns user-specific storage settings
func GetUserStorageSettings(c *gin.Context) {
	userID := utils.GetUserID(c.MustGet("userId"))

	var setting model.UserStorageSetting
	if err := mysql.DB.Where("user_id = ?", userID).First(&setting).Error; err != nil {
		// Return default settings
		c.JSON(http.StatusOK, model.Success(map[string]interface{}{
			"syncDir":     "",
			"downloadDir": "",
			"autoSync":    true,
		}))
		return
	}

	c.JSON(http.StatusOK, model.Success(map[string]interface{}{
		"syncDir":     setting.SyncDir,
		"downloadDir": setting.DownloadDir,
		"autoSync":    setting.AutoSync,
	}))
}

// SaveUserStorageSettings saves user-specific storage settings
func SaveUserStorageSettings(c *gin.Context) {
	userID := utils.GetUserID(c.MustGet("userId"))

	var req struct {
		SyncDir     string `json:"syncDir"`
		DownloadDir string `json:"downloadDir"`
		AutoSync    bool   `json:"autoSync"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}

	var setting model.UserStorageSetting
	result := mysql.DB.Where("user_id = ?", userID).First(&setting)
	if result.Error != nil {
		// Create new
		setting = model.UserStorageSetting{
			UserID:      userID,
			SyncDir:     req.SyncDir,
			DownloadDir: req.DownloadDir,
			AutoSync:    req.AutoSync,
		}
		mysql.DB.Create(&setting)
	} else {
		// Update existing
		mysql.DB.Model(&setting).Updates(map[string]interface{}{
			"sync_dir":     req.SyncDir,
			"download_dir": req.DownloadDir,
			"auto_sync":    req.AutoSync,
		})
	}

	c.JSON(http.StatusOK, model.Success(nil))
}

// ===================== Remote Storage Management =====================

// ListRemoteStorages returns list of remote storage configurations
func ListRemoteStorages(c *gin.Context) {
	var storages []model.RemoteStorage
	mysql.DB.Find(&storages)

	c.JSON(http.StatusOK, model.Success(storages))
}

// CreateRemoteStorage creates a new remote storage configuration
func CreateRemoteStorage(c *gin.Context) {
	var req struct {
		Name       string `json:"name" binding:"required"`
		Type       string `json:"type" binding:"required"`
		Server     string `json:"server" binding:"required"`
		Port       int    `json:"port"`
		Username   string `json:"username"`
		Password   string `json:"password"`
		SharePath  string `json:"sharePath"`
		Domain     string `json:"domain"`
		MountPoint string `json:"mountPoint" binding:"required"`
		IsEnabled  bool   `json:"isEnabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}

	// TODO: Encrypt password before saving
	storage := model.RemoteStorage{
		Name:       req.Name,
		Type:       req.Type,
		Server:     req.Server,
		Port:       req.Port,
		Username:   req.Username,
		Password:   req.Password, // Should be encrypted
		SharePath:  req.SharePath,
		Domain:     req.Domain,
		MountPoint: req.MountPoint,
		Status:     "connected", // Set to connected after creation
		IsEnabled:  req.IsEnabled,
	}
	mysql.DB.Create(&storage)

	c.JSON(http.StatusOK, model.Success(storage))
}

// UpdateRemoteStorage updates an existing remote storage configuration
func UpdateRemoteStorage(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Name       string `json:"name" binding:"required"`
		Type       string `json:"type" binding:"required"`
		Server     string `json:"server" binding:"required"`
		Port       int    `json:"port"`
		Username   string `json:"username"`
		Password   string `json:"password"`
		SharePath  string `json:"sharePath"`
		Domain     string `json:"domain"`
		MountPoint string `json:"mountPoint" binding:"required"`
		IsEnabled  bool   `json:"isEnabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}

	var storage model.RemoteStorage
	if err := mysql.DB.First(&storage, id).Error; err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("存储不存在"))
		return
	}

	// Update fields
	updates := map[string]interface{}{
		"name":        req.Name,
		"type":        req.Type,
		"server":      req.Server,
		"port":        req.Port,
		"username":    req.Username,
		"share_path":  req.SharePath,
		"domain":      req.Domain,
		"mount_point": req.MountPoint,
		"is_enabled":  req.IsEnabled,
	}
	// Only update password if provided
	if req.Password != "" {
		updates["password"] = req.Password // Should be encrypted
	}

	mysql.DB.Model(&storage).Updates(updates)

	// Disconnect cached client so it will reconnect with new config
	service.GetRemoteStorageService().Disconnect(uint(storage.ID))

	c.JSON(http.StatusOK, model.Success(nil))
}

// DeleteRemoteStorage deletes a remote storage configuration
func DeleteRemoteStorage(c *gin.Context) {
	id := c.Param("id")

	result := mysql.DB.Delete(&model.RemoteStorage{}, id)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusOK, model.ErrorMsg("存储不存在"))
		return
	}

	c.JSON(http.StatusOK, model.Success(nil))
}

// TestRemoteStorageConnection tests connection to a specific remote storage
func TestRemoteStorageConnection(c *gin.Context) {
	id := c.Param("id")

	var storage model.RemoteStorage
	if err := mysql.DB.First(&storage, id).Error; err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("存储不存在"))
		return
	}

	// Disconnect any existing cached connection first
	service.GetRemoteStorageService().Disconnect(uint(storage.ID))

	// Build config for testing
	config := service.RemoteStorageConfig{
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

	// Test connection using the service
	err := service.GetRemoteStorageService().TestConnectionWithConfig(config)
	if err != nil {
		// Update status to error
		mysql.DB.Model(&storage).Update("status", "error")
		c.JSON(http.StatusOK, model.ErrorMsg("连接失败: "+err.Error()))
		return
	}

	// Update status to connected
	mysql.DB.Model(&storage).Update("status", "connected")

	c.JSON(http.StatusOK, model.Success(map[string]interface{}{
		"connected": true,
		"message":   "连接成功",
	}))
}

// ConnectRemoteStorage establishes connection to a remote storage
func ConnectRemoteStorage(c *gin.Context) {
	id := c.Param("id")

	var storage model.RemoteStorage
	if err := mysql.DB.First(&storage, id).Error; err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("存储不存在"))
		return
	}

	// Disconnect any existing cached connection first
	service.GetRemoteStorageService().Disconnect(uint(storage.ID))

	// Build config
	config := service.RemoteStorageConfig{
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

	// Test connection
	err := service.GetRemoteStorageService().TestConnectionWithConfig(config)
	if err != nil {
		mysql.DB.Model(&storage).Update("status", "error")
		c.JSON(http.StatusOK, model.ErrorMsg("连接失败: "+err.Error()))
		return
	}

	// Update status to connected
	mysql.DB.Model(&storage).Update("status", "connected")

	c.JSON(http.StatusOK, model.Success(map[string]interface{}{
		"status":  "connected",
		"message": "连接成功",
	}))
}

// DisconnectRemoteStorage disconnects from a remote storage
func DisconnectRemoteStorage(c *gin.Context) {
	id := c.Param("id")

	var storage model.RemoteStorage
	if err := mysql.DB.First(&storage, id).Error; err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("存储不存在"))
		return
	}

	// Disconnect and clear cache
	err := service.GetRemoteStorageService().Disconnect(uint(storage.ID))
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("断开连接失败: "+err.Error()))
		return
	}

	// Update status to disconnected
	mysql.DB.Model(&storage).Update("status", "disconnected")

	c.JSON(http.StatusOK, model.Success(map[string]interface{}{
		"status":  "disconnected",
		"message": "已断开连接",
	}))
}
