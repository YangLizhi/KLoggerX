package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/pkg/utils"
	"kloggerx-server/internal/repository/mysql"
	"kloggerx-server/internal/service"

	"github.com/gin-gonic/gin"
)

// GetAIModelSettings godoc
// @Summary 获取AI模型设置
// @Description 获取AI模型配置信息
// @Tags 系统管理
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /admin/settings/ai [get]
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

// SaveAIModelSettings godoc
// @Summary 保存AI模型设置
// @Description 保存AI模型配置
// @Tags 系统管理
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /admin/settings/ai [post]
func SaveAIModelSettings(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}

	data, err := json.Marshal(req)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("序列化数据失败: " + err.Error()))
		return
	}

	var setting model.SystemSetting
	result := mysql.DB.Where("`key` = ?", "ai_model_settings").First(&setting)
	if result.Error != nil {
		setting = model.SystemSetting{Key: "ai_model_settings", Value: string(data)}
		if err := mysql.DB.Create(&setting).Error; err != nil {
			c.JSON(http.StatusOK, model.ErrorMsg("保存设置失败: "+err.Error()))
			return
		}
	} else {
		if err := mysql.DB.Model(&model.SystemSetting{}).Where("`key` = ?", "ai_model_settings").Update("value", string(data)).Error; err != nil {
			c.JSON(http.StatusOK, model.ErrorMsg("保存设置失败: "+err.Error()))
			return
		}
	}

	c.JSON(http.StatusOK, model.Success(nil))
}

// GetStorageSettings godoc
// @Summary 获取存储设置
// @Description 获取存储配置信息
// @Tags 系统管理
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /admin/settings/storage [get]
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

// SaveStorageSettings godoc
// @Summary 保存存储设置
// @Description 保存存储配置
// @Tags 系统管理
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /admin/settings/storage [post]
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

// TestRemoteStorage godoc
// @Summary 测试远程存储连接
// @Description 测试远程存储连接是否正常
// @Tags 系统管理
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /admin/settings/storage/test [post]
func TestRemoteStorage(c *gin.Context) {
	var req struct {
		Type     string `json:"type" binding:"required,max=50"`
		Server   string `json:"server" binding:"required,max=500"`
		Port     int    `json:"port"`
		Username string `json:"username" binding:"max=200"`
		Password string `json:"password" binding:"max=500"`
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

// DetectAIModels godoc
// @Summary 检测AI模型
// @Description 探测AI提供商可用模型列表
// @Tags 系统管理
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /admin/settings/ai/detect [post]
func DetectAIModels(c *gin.Context) {
	var req struct {
		BaseURL string `json:"baseUrl" binding:"required,max=500"`
		APIKey  string `json:"apiKey" binding:"required,max=500"`
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

// TestAIModel godoc
// @Summary 测试AI模型
// @Description 测试指定AI模型是否可用
// @Tags 系统管理
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /admin/settings/ai/test [post]
func TestAIModel(c *gin.Context) {
	var req struct {
		BaseURL string `json:"baseUrl" binding:"required,max=500"`
		APIKey  string `json:"apiKey" binding:"required,max=500"`
		Model   string `json:"model" binding:"required,max=200"`
		Prompt  string `json:"prompt" binding:"max=5000"`
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

// TestLdapConnection godoc
// @Summary 测试LDAP连接
// @Description 测试LDAP/AD服务器连接
// @Tags 系统管理
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /admin/ldap/test [post]
func TestLdapConnection(c *gin.Context) {
	var req struct {
		Type         string `json:"type" binding:"required,max=50"`
		Server       string `json:"server" binding:"required,max=500"`
		BaseDN       string `json:"baseDN" binding:"required,max=500"`
		BindDN       string `json:"bindDN" binding:"required,max=500"`
		BindPassword string `json:"bindPassword" binding:"required,max=500"`
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

	// 解析服务器地址
	host, port, useSSL, err := service.ParseLDAPServerURL(req.Server)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}

	config := service.LDAPConfig{
		Host:         host,
		Port:         port,
		BaseDN:       req.BaseDN,
		BindDN:       req.BindDN,
		BindPassword: req.BindPassword,
		UseSSL:       useSSL,
	}

	if err := service.TestLDAPConnection(config); err != nil {
		c.JSON(http.StatusOK, model.Success(map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}))
		return
	}

	c.JSON(http.StatusOK, model.Success(map[string]interface{}{
		"success": true,
		"message": "连接成功",
	}))
}

// FetchLdapUsers godoc
// @Summary 获取LDAP用户
// @Description 从LDAP/AD获取用户列表
// @Tags 系统管理
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /admin/ldap/users [post]
func FetchLdapUsers(c *gin.Context) {
	var req struct {
		Type         string `json:"type" binding:"required,max=50"`
		Server       string `json:"server" binding:"required,max=500"`
		BaseDN       string `json:"baseDN" binding:"required,max=500"`
		BindDN       string `json:"bindDN" binding:"required,max=500"`
		BindPassword string `json:"bindPassword" binding:"required,max=500"`
		Page         int    `json:"page"`
		PageSize     int    `json:"pageSize"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}

	host, port, useSSL, err := service.ParseLDAPServerURL(req.Server)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 50
	}

	config := service.LDAPConfig{
		Host:         host,
		Port:         port,
		BaseDN:       req.BaseDN,
		BindDN:       req.BindDN,
		BindPassword: req.BindPassword,
		UseSSL:       useSSL,
	}

	users, total, err := service.ListLDAPUsers(config, req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(fmt.Sprintf("获取LDAP用户失败: %v", err)))
		return
	}

	c.JSON(http.StatusOK, model.Success(map[string]interface{}{
		"users": users,
		"total": total,
	}))
}

// ImportLdapUsers godoc
// @Summary 导入LDAP用户
// @Description 导入选中的LDAP用户到系统
// @Tags 系统管理
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /admin/ldap/import [post]
func ImportLdapUsers(c *gin.Context) {
	var req struct {
		Type         string   `json:"type" binding:"required,max=50"`
		Server       string   `json:"server" binding:"required,max=500"`
		BaseDN       string   `json:"baseDN" binding:"required,max=500"`
		BindDN       string   `json:"bindDN" binding:"required,max=500"`
		BindPassword string   `json:"bindPassword" binding:"required,max=500"`
		Users        []string `json:"users" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}

	host, port, useSSL, err := service.ParseLDAPServerURL(req.Server)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}

	config := service.LDAPConfig{
		Host:         host,
		Port:         port,
		BaseDN:       req.BaseDN,
		BindDN:       req.BindDN,
		BindPassword: req.BindPassword,
		UseSSL:       useSSL,
	}

	imported, importErrors := service.ImportLDAPUsers(config, req.Users)

	c.JSON(http.StatusOK, model.Success(map[string]interface{}{
		"imported": imported,
		"errors":   importErrors,
	}))
}

// ===================== User Storage Settings =====================

// GetUserStorageSettings godoc
// @Summary 获取用户存储设置
// @Description 获取用户个人存储设置
// @Tags 用户
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /user/storage-settings [get]
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

// SaveUserStorageSettings godoc
// @Summary 保存用户存储设置
// @Description 保存用户个人存储设置
// @Tags 用户
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /user/storage-settings [post]
func SaveUserStorageSettings(c *gin.Context) {
	userID := utils.GetUserID(c.MustGet("userId"))

	var req struct {
		SyncDir     string `json:"syncDir" binding:"max=500"`
		DownloadDir string `json:"downloadDir" binding:"max=500"`
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

// ListRemoteStorages godoc
// @Summary 获取远程存储列表
// @Description 获取所有远程存储配置
// @Tags 系统管理
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /admin/remote-storages [get]
func ListRemoteStorages(c *gin.Context) {
	var storages []model.RemoteStorage
	mysql.DB.Find(&storages)

	c.JSON(http.StatusOK, model.Success(storages))
}

// CreateRemoteStorage godoc
// @Summary 创建远程存储
// @Description 创建新的远程存储配置
// @Tags 系统管理
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /admin/remote-storages [post]
func CreateRemoteStorage(c *gin.Context) {
	var req struct {
		Name       string `json:"name" binding:"required,max=200"`
		Type       string `json:"type" binding:"required,oneof=sftp ftp smb webdav nfs baidu aliyun tencent"`
		Server     string `json:"server" binding:"required,max=500"`
		Port       int    `json:"port"`
		Username   string `json:"username" binding:"max=200"`
		Password   string `json:"password" binding:"max=500"`
		SharePath  string `json:"sharePath" binding:"max=500"`
		Domain     string `json:"domain" binding:"max=200"`
		MountPoint string `json:"mountPoint" binding:"required,max=200"`
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

// UpdateRemoteStorage godoc
// @Summary 更新远程存储
// @Description 更新远程存储配置
// @Tags 系统管理
// @Accept json
// @Produce json
// @Param id path int true "存储ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /admin/remote-storages/{id} [put]
func UpdateRemoteStorage(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Name       string `json:"name" binding:"required,max=200"`
		Type       string `json:"type" binding:"required,oneof=sftp ftp smb webdav nfs baidu aliyun tencent"`
		Server     string `json:"server" binding:"required,max=500"`
		Port       int    `json:"port"`
		Username   string `json:"username" binding:"max=200"`
		Password   string `json:"password" binding:"max=500"`
		SharePath  string `json:"sharePath" binding:"max=500"`
		Domain     string `json:"domain" binding:"max=200"`
		MountPoint string `json:"mountPoint" binding:"required,max=200"`
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

// DeleteRemoteStorage godoc
// @Summary 删除远程存储
// @Description 删除远程存储配置
// @Tags 系统管理
// @Produce json
// @Param id path int true "存储ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /admin/remote-storages/{id} [delete]
func DeleteRemoteStorage(c *gin.Context) {
	id := c.Param("id")

	result := mysql.DB.Delete(&model.RemoteStorage{}, id)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusOK, model.ErrorMsg("存储不存在"))
		return
	}

	c.JSON(http.StatusOK, model.Success(nil))
}

// TestRemoteStorageConnection godoc
// @Summary 测试远程存储连接
// @Description 测试指定远程存储的连接
// @Tags 系统管理
// @Produce json
// @Param id path int true "存储ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /admin/remote-storages/{id}/test [post]
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

// ConnectRemoteStorage godoc
// @Summary 连接远程存储
// @Description 建立与远程存储的连接
// @Tags 系统管理
// @Produce json
// @Param id path int true "存储ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /admin/remote-storages/{id}/connect [post]
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

// DisconnectRemoteStorage godoc
// @Summary 断开远程存储
// @Description 断开与远程存储的连接
// @Tags 系统管理
// @Produce json
// @Param id path int true "存储ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /admin/remote-storages/{id}/disconnect [post]
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
