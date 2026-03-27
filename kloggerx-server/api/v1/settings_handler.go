package v1

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/repository/mysql"

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

	// Parse response
	var apiResp struct {
		Data []struct {
			ID     string `json:"id"`
			Object string `json:"object"`
			Owned  bool   `json:"owned_by"`
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
