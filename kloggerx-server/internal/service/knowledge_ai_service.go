package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/pkg/logger"
	"kloggerx-server/internal/repository/mysql"
	"kloggerx-server/internal/repository/qdrant"
)

// ─── Embedding Config Cache ──────────────────────────────────────────────────

var embeddingConfigCache struct {
	sync.RWMutex
	baseURL  string
	apiKey   string
	modelID  string
	cachedAt time.Time
	ttl      time.Duration
}

func init() {
	embeddingConfigCache.ttl = 5 * time.Minute
}

// ─── Knowledge Source Management ─────────────────────────────────────────────

func AddKnowledgeSource(kbID uint, sourceType string, sourceID uint) (*model.KnowledgeSource, error) {
	// Verify source exists and get its name
	var sourceName string
	var doc model.Document
	if err := mysql.DB.First(&doc, sourceID).Error; err != nil {
		return nil, fmt.Errorf("来源不存在")
	}
	sourceName = doc.Title

	src := model.KnowledgeSource{
		KnowledgeBaseID: kbID,
		SourceType:      sourceType,
		SourceID:        sourceID,
		SourceName:      sourceName,
		AutoSync:        true,
	}
	if err := mysql.DB.Where("knowledge_base_id = ? AND source_type = ? AND source_id = ?",
		kbID, sourceType, sourceID).FirstOrCreate(&src).Error; err != nil {
		return nil, err
	}
	// Immediately sync documents from the source
	go syncSourceToKB(kbID, sourceType, sourceID)
	return &src, nil
}

func RemoveKnowledgeSource(kbID, sourceID uint) error {
	return mysql.DB.Where("knowledge_base_id = ? AND id = ?", kbID, sourceID).Delete(&model.KnowledgeSource{}).Error
}

func GetKnowledgeSources(kbID uint) ([]model.KnowledgeSource, error) {
	var list []model.KnowledgeSource
	err := mysql.DB.Where("knowledge_base_id = ?", kbID).Order("created_at DESC").Find(&list).Error
	for i := range list {
		var cnt int64
		mysql.DB.Model(&model.KnowledgeChunk{}).
			Where("knowledge_base_id = ? AND document_id IN (?)",
				kbID,
				mysql.DB.Model(&model.Document{}).Select("id").Where("parent_id = ? OR id = ?", list[i].SourceID, list[i].SourceID),
			).Count(&cnt)
		list[i].DocCount = int(cnt)
	}
	return list, err
}

func SyncKnowledgeSource(sourceEntryID uint) error {
	var src model.KnowledgeSource
	if err := mysql.DB.First(&src, sourceEntryID).Error; err != nil {
		return err
	}
	syncSourceToKB(src.KnowledgeBaseID, src.SourceType, src.SourceID)
	now := time.Now()
	mysql.DB.Model(&model.KnowledgeSource{}).Where("id = ?", sourceEntryID).Update("last_sync_at", &now)
	return nil
}

// syncSourceToKB scans a folder (or single document) and indexes all documents into the KB.
func syncSourceToKB(kbID uint, sourceType string, sourceID uint) {
	var docIDs []uint

	if sourceType == "folder" {
		// Collect all document IDs recursively under the folder
		collectFolderDocIDs(sourceID, &docIDs)
	} else {
		docIDs = []uint{sourceID}
	}

	for _, docID := range docIDs {
		// Add to KnowledgeDocument if not already present
		kd := model.KnowledgeDocument{KnowledgeBaseID: kbID, DocumentID: docID}
		mysql.DB.Where("knowledge_base_id = ? AND document_id = ?", kbID, docID).FirstOrCreate(&kd)
		// Index document content for RAG
		IndexDocumentContent(kbID, docID)
	}
}

// collectFolderDocIDs recursively collects IDs of non-folder documents under a folder.
func collectFolderDocIDs(folderID uint, result *[]uint) {
	var children []model.Document
	mysql.DB.Where("parent_id = ? AND is_deleted = ?", folderID, false).Find(&children)
	for _, child := range children {
		if child.Type == "folder" {
			collectFolderDocIDs(child.ID, result)
		} else {
			*result = append(*result, child.ID)
		}
	}
}

// ─── RAG Indexing ────────────────────────────────────────────────────────────

const chunkSize = 500 // characters per chunk

// IndexDocumentContent chunks a document's content and stores chunks in knowledge_chunks.
// If vector search is enabled, it also generates embeddings and stores them in Qdrant.
func IndexDocumentContent(kbID, docID uint) {
	var doc model.Document
	if err := mysql.DB.First(&doc, docID).Error; err != nil {
		logger.Debugf("IndexDocumentContent: document %d not found: %v", docID, err)
		return
	}

	// Extract plain text from JSON content
	plainText := extractPlainTextFromContent(doc.Content)
	if plainText == "" {
		logger.Debugf("IndexDocumentContent: document %d (%s) has no extractable text, content length=%d", docID, doc.Title, len(doc.Content))
		return
	}

	logger.Debugf("IndexDocumentContent: processing document %d (%s), text length=%d", docID, doc.Title, len(plainText))

	// Delete old chunks for this document in this KB
	mysql.DB.Where("knowledge_base_id = ? AND document_id = ?", kbID, docID).Delete(&model.KnowledgeChunk{})

	// Delete old vectors in Qdrant
	if qdrant.DefaultVectorClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		qdrant.DefaultVectorClient.DeleteByDocument(ctx, kbID, docID)
	}

	// Split into chunks
	chunks := splitIntoChunks(plainText, chunkSize)

	// Store chunks in MySQL
	chunkRecords := make([]model.KnowledgeChunk, len(chunks))
	for i, chunk := range chunks {
		chunkRecords[i] = model.KnowledgeChunk{
			KnowledgeBaseID: kbID,
			DocumentID:      docID,
			DocumentTitle:   doc.Title,
			Content:         chunk,
			ChunkIndex:      i,
			EmbeddingStatus: "pending",
		}
		mysql.DB.Create(&chunkRecords[i])
	}

	// Generate embeddings and store in Qdrant (async)
	if globalEmbeddingService != nil && qdrant.DefaultVectorClient != nil {
		go generateEmbeddingsForChunks(kbID, docID, doc.Title, chunkRecords)
	}
}

// generateEmbeddingsForChunks generates embeddings for chunks and stores them in Qdrant
func generateEmbeddingsForChunks(kbID, docID uint, docTitle string, chunks []model.KnowledgeChunk) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Create embedding job record
	job := model.EmbeddingJob{
		KnowledgeBaseID: kbID,
		DocumentID:      docID,
		Status:          "processing",
		ChunkCount:      len(chunks),
	}
	mysql.DB.Create(&job)

	// Get embedding settings dynamically (not from global cached service)
	baseURL, apiKey, modelID, err := GetEmbeddingModelSettings()
	if err != nil {
		job.Status = "failed"
		job.ErrorMessage = fmt.Sprintf("Failed to get embedding settings: %v", err)
		mysql.DB.Save(&job)
		return
	}

	// Create a new embedding service with the current settings
	embedService := NewEmbeddingService(EmbeddingConfig{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Model:   modelID,
	})

	// Extract texts for embedding
	texts := make([]string, len(chunks))
	for i, c := range chunks {
		texts[i] = c.Content
	}

	// Generate embeddings
	embeddings, err := embedService.BatchGenerateEmbeddings(ctx, texts, 100)
	if err != nil {
		job.Status = "failed"
		job.ErrorMessage = err.Error()
		mysql.DB.Save(&job)
		return
	}

	// Update vector size based on the embedding model
	vectorDim := GetEmbeddingDimensionForModel(modelID)
	qdrant.SetVectorSize(vectorDim)

	// Prepare vector points
	points := make([]*qdrant.VectorPoint, len(chunks))
	for i, embedding := range embeddings {
		if i < len(chunks) {
			points[i] = &qdrant.VectorPoint{
				Vector:        embedding,
				DocumentID:    docID,
				ChunkID:       chunks[i].ID,
				ChunkIndex:    chunks[i].ChunkIndex,
				Content:       chunks[i].Content,
				DocumentTitle: docTitle,
				KBID:          kbID,
				RaptorLevel:   0,
			}
		}
	}

	// Upsert to Qdrant
	pointIDs, err := qdrant.DefaultVectorClient.BatchUpsertVectors(ctx, points)
	if err != nil {
		job.Status = "failed"
		job.ErrorMessage = fmt.Sprintf("Failed to store vectors: %v", err)
		mysql.DB.Save(&job)
		return
	}

	// Update chunk records with Qdrant point IDs
	for i, pointID := range pointIDs {
		if i < len(chunks) {
			mysql.DB.Model(&model.KnowledgeChunk{}).
				Where("id = ?", chunks[i].ID).
				Updates(map[string]interface{}{
					"qdrant_point_id":  pointID,
					"embedding_status": "embedded",
				})
		}
	}

	// Mark job as completed
	job.Status = "completed"
	mysql.DB.Save(&job)
}

func extractPlainTextFromContent(content string) string {
	if content == "" || content == "{}" {
		return ""
	}
	// Parse Tiptap/ProseMirror JSON and extract text nodes
	var doc map[string]interface{}
	if err := json.Unmarshal([]byte(content), &doc); err != nil {
		// Not JSON, treat as plain text
		return strings.TrimSpace(content)
	}

	// Check if it's a valid Tiptap document with content
	if _, hasContent := doc["content"]; hasContent {
		var sb strings.Builder
		extractTextNodes(doc, &sb)
		result := strings.TrimSpace(sb.String())
		if result != "" {
			return result
		}
	}

	// Try to extract text from any string values in the JSON
	var sb strings.Builder
	extractAllText(doc, &sb)
	result := strings.TrimSpace(sb.String())
	if result != "" {
		return result
	}

	// Fallback: return raw content stripped of JSON syntax
	return strings.TrimSpace(content)
}

// extractAllText recursively extracts all string values from a JSON structure
func extractAllText(data interface{}, sb *strings.Builder) {
	switch v := data.(type) {
	case map[string]interface{}:
		// Check for specific text fields
		if text, ok := v["text"].(string); ok && text != "" {
			sb.WriteString(text)
			sb.WriteString(" ")
		}
		// Recurse into all values
		for _, val := range v {
			extractAllText(val, sb)
		}
	case []interface{}:
		for _, item := range v {
			extractAllText(item, sb)
		}
	case string:
		if v != "" && len(v) > 1 {
			sb.WriteString(v)
			sb.WriteString(" ")
		}
	}
}

func extractTextNodes(node map[string]interface{}, sb *strings.Builder) {
	if t, ok := node["text"].(string); ok {
		sb.WriteString(t)
		sb.WriteString(" ")
	}
	if content, ok := node["content"].([]interface{}); ok {
		for _, child := range content {
			if childMap, ok := child.(map[string]interface{}); ok {
				extractTextNodes(childMap, sb)
			}
		}
		sb.WriteString("\n")
	}
}

func splitIntoChunks(text string, size int) []string {
	runes := []rune(text)
	var chunks []string
	for i := 0; i < len(runes); i += size {
		end := i + size
		if end > len(runes) {
			end = len(runes)
		}
		chunk := strings.TrimSpace(string(runes[i:end]))
		if chunk != "" {
			chunks = append(chunks, chunk)
		}
	}
	return chunks
}

// ─── RAG Search ──────────────────────────────────────────────────────────────

type ChunkRef struct {
	DocumentID    uint   `json:"documentId"`
	DocumentTitle string `json:"documentTitle"`
	Content       string `json:"content"`
	ChunkIndex    int    `json:"chunkIndex"`
}

// SearchChunks searches for relevant chunks using keywords extracted from a question.
func SearchChunks(kbID uint, question string, limit int) []ChunkRef {
	keywords := extractKeywords(question)
	if len(keywords) == 0 {
		return nil
	}

	// Build OR LIKE conditions for each keyword
	var conditions []string
	var args []interface{}
	for _, kw := range keywords {
		conditions = append(conditions, "content LIKE ?")
		args = append(args, "%"+kw+"%")
	}
	whereClause := "knowledge_base_id = ? AND (" + strings.Join(conditions, " OR ") + ")"
	args = append([]interface{}{kbID}, args...)

	var chunks []model.KnowledgeChunk
	mysql.DB.Where(whereClause, args...).
		Order("updated_at DESC").
		Limit(limit).
		Find(&chunks)

	var refs []ChunkRef
	for _, c := range chunks {
		refs = append(refs, ChunkRef{
			DocumentID:    c.DocumentID,
			DocumentTitle: c.DocumentTitle,
			Content:       c.Content,
			ChunkIndex:    c.ChunkIndex,
		})
	}
	return refs
}

func extractKeywords(text string) []string {
	// Remove common Chinese stop words, split by whitespace and punctuation
	replacer := strings.NewReplacer(
		"，", " ", "。", " ", "？", " ", "！", " ",
		"、", " ", "：", " ", "；", " ", "的", " ",
		"了", " ", "是", " ", "在", " ", "有", " ",
		"和", " ", "与", " ", "我", " ", "你", " ",
		"他", " ", "她", " ", "它", " ", "们", " ",
	)
	cleaned := replacer.Replace(text)
	words := strings.Fields(cleaned)

	seen := map[string]bool{}
	var result []string
	for _, w := range words {
		w = strings.TrimSpace(w)
		if len([]rune(w)) >= 2 && !seen[w] {
			seen[w] = true
			result = append(result, w)
		}
	}
	if len(result) > 8 {
		result = result[:8]
	}
	return result
}

// ─── AI Chat (RAG) ───────────────────────────────────────────────────────────

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type KBChatResult struct {
	Answer  string     `json:"answer"`
	Sources []ChunkRef `json:"sources"`
}

// ChatWithKnowledge performs a RAG-based Q&A against a knowledge base.
// It uses hybrid search (vector + fulltext) when available, falling back to keyword search.
func ChatWithKnowledge(kbID uint, question string, history []ChatMessage) (*KBChatResult, error) {
	return ChatWithKnowledgeOpts(kbID, question, history, "")
}

// ChatWithKnowledgeOpts performs a RAG-based Q&A against a knowledge base with optional model override.
func ChatWithKnowledgeOpts(kbID uint, question string, history []ChatMessage, modelOverride string) (*KBChatResult, error) {
	var chunks []ChunkRef

	// Check if knowledge base has any content first
	var totalChunks int64
	mysql.DB.Model(&model.KnowledgeChunk{}).Where("knowledge_base_id = ?", kbID).Count(&totalChunks)
	if totalChunks == 0 {
		return nil, fmt.Errorf("当前知识库内没有内容，请先在知识库设置中添加数据来源并同步")
	}

	// Try vector search first if Qdrant is available
	if qdrant.DefaultVectorClient != nil {
		// Get embedding settings dynamically
		embedBaseURL, embedAPIKey, embedModelID, err := GetEmbeddingModelSettings()
		if err == nil && embedAPIKey != "" {
			// Create embedding service for this query
			embedSvc := NewEmbeddingService(EmbeddingConfig{
				BaseURL: embedBaseURL,
				APIKey:  embedAPIKey,
				Model:   embedModelID,
			})

			// Create retrieval service dynamically
			retrievalSvc := NewRetrievalService(embedSvc, qdrant.DefaultVectorClient)

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			results, searchErr := retrievalSvc.HybridSearch(ctx, kbID, question, DefaultHybridSearchOptions())
			cancel()

			if searchErr != nil {
				logger.Debugf("HybridSearch error: %v", searchErr)
			} else if len(results) > 0 {
				logger.Debugf("HybridSearch found %d results", len(results))
				for _, r := range results {
					chunks = append(chunks, ChunkRef{
						DocumentID:    r.DocumentID,
						DocumentTitle: r.DocumentTitle,
						Content:       r.Content,
						ChunkIndex:    r.ChunkIndex,
					})
				}
			} else {
				logger.Debugf("HybridSearch returned 0 results")
			}
		} else {
			logger.Debugf("GetEmbeddingModelSettings error: %v", err)
		}
	} else {
		logger.Debugf("Qdrant client is nil, skipping vector search")
	}

	// Fallback to keyword search if vector search failed or returned no results
	if len(chunks) == 0 {
		logger.Debugf("Falling back to keyword search")
		chunks = SearchChunks(kbID, question, 6)
		logger.Debugf("Keyword search found %d results", len(chunks))
	}

	// 2. Build context from chunks
	var contextParts []string
	for i, c := range chunks {
		contextParts = append(contextParts, fmt.Sprintf("[%d] 来自《%s》:\n%s", i+1, c.DocumentTitle, c.Content))
	}
	contextText := strings.Join(contextParts, "\n\n")

	systemPrompt := "你是一个知识库助手，根据提供的文档内容准确回答用户问题。回答时请引用来源编号，格式为「来源[n]」。如果文档中没有相关信息，请如实说明。"
	if contextText != "" {
		systemPrompt += "\n\n以下是相关文档内容：\n" + contextText
	}

	// 3. Call AI API with optional model override
	answer, err := callAIChatWithModel(systemPrompt, question, history, modelOverride)
	if err != nil {
		return nil, err
	}

	return &KBChatResult{
		Answer:  answer,
		Sources: chunks,
	}, nil
}

// GetEmbeddingModelSettings reads embedding model settings from the database with caching.
// Returns baseURL, apiKey, modelID for the embedding model.
func GetEmbeddingModelSettings() (baseURL, apiKey, modelID string, err error) {
	embeddingConfigCache.RLock()
	if time.Since(embeddingConfigCache.cachedAt) < embeddingConfigCache.ttl {
		baseURL = embeddingConfigCache.baseURL
		apiKey = embeddingConfigCache.apiKey
		modelID = embeddingConfigCache.modelID
		embeddingConfigCache.RUnlock()
		return
	}
	embeddingConfigCache.RUnlock()

	// Cache expired, reload from DB
	embeddingConfigCache.Lock()
	defer embeddingConfigCache.Unlock()

	// Double-check after acquiring write lock
	if time.Since(embeddingConfigCache.cachedAt) < embeddingConfigCache.ttl {
		return embeddingConfigCache.baseURL, embeddingConfigCache.apiKey, embeddingConfigCache.modelID, nil
	}

	baseURL, apiKey, modelID, err = getEmbeddingModelSettingsFromDB()
	if err == nil {
		embeddingConfigCache.baseURL = baseURL
		embeddingConfigCache.apiKey = apiKey
		embeddingConfigCache.modelID = modelID
		embeddingConfigCache.cachedAt = time.Now()
	}
	return
}

// getEmbeddingModelSettingsFromDB is the actual DB lookup for embedding model settings.
func getEmbeddingModelSettingsFromDB() (baseURL, apiKey, modelID string, err error) {
	var setting model.SystemSetting
	if dbErr := mysql.DB.Where("`key` = ?", "ai_model_settings").First(&setting).Error; dbErr != nil {
		return "", "", "", fmt.Errorf("AI模型未配置，请在管理后台配置AI模型")
	}

	var data struct {
		Providers []struct {
			ID       int    `json:"id"`
			Name     string `json:"name"`
			BaseURL  string `json:"baseUrl"`
			APIKey   string `json:"apiKey"`
			IsActive bool   `json:"isActive"`
			Models   []struct {
				ID        string `json:"id"`
				Name      string `json:"name"`
				Type      string `json:"type"`
				IsDefault bool   `json:"isDefault"`
			} `json:"models"`
		} `json:"providers"`
		KBSettings struct {
			ChatModel      string `json:"chatModel"`
			EmbeddingModel string `json:"embeddingModel"`
		} `json:"kbSettings"`
	}
	if parseErr := json.Unmarshal([]byte(setting.Value), &data); parseErr != nil {
		return "", "", "", fmt.Errorf("AI配置解析失败")
	}

	logger.Debugf("GetEmbeddingModelSettings: embeddingModel='%s', providers=%d",
		data.KBSettings.EmbeddingModel, len(data.Providers))

	// If embeddingModel is specified, find it in all active providers
	if data.KBSettings.EmbeddingModel != "" {
		targetModel := data.KBSettings.EmbeddingModel
		for _, p := range data.Providers {
			if !p.IsActive || p.BaseURL == "" || p.APIKey == "" {
				logger.Debugf("Skipping inactive provider: %s (active=%v, baseUrl='%s')", p.Name, p.IsActive, p.BaseURL)
				continue
			}
			logger.Debugf("Checking provider '%s' with %d models", p.Name, len(p.Models))
			for _, m := range p.Models {
				if m.ID == targetModel {
					logger.Debugf("Found embedding model '%s' in provider '%s' (baseUrl='%s')", m.ID, p.Name, p.BaseURL)
					return p.BaseURL, p.APIKey, m.ID, nil
				}
			}
		}
		// Model ID not found in any active provider
		return "", "", "", fmt.Errorf("指定的嵌入模型 '%s' 未在任何活跃的提供商中找到，请检查配置", targetModel)
	}

	// Find first active provider with an embedding model
	for _, p := range data.Providers {
		if !p.IsActive || p.BaseURL == "" || p.APIKey == "" {
			continue
		}
		// First try to find default embedding model
		for _, m := range p.Models {
			if m.Type == "embedding" && m.IsDefault {
				return p.BaseURL, p.APIKey, m.ID, nil
			}
		}
		// If no default, use first embedding model
		for _, m := range p.Models {
			if m.Type == "embedding" {
				return p.BaseURL, p.APIKey, m.ID, nil
			}
		}
	}

	return "", "", "", fmt.Errorf("未找到可用的嵌入模型，请在管理后台配置")
}

// GetAISettings reads AI model settings from the database.
func GetAISettings() (baseURL, apiKey, modelID string, err error) {
	var setting model.SystemSetting
	if dbErr := mysql.DB.Where("`key` = ?", "ai_model_settings").First(&setting).Error; dbErr != nil {
		return "", "", "", fmt.Errorf("AI模型未配置，请在管理后台配置AI模型")
	}

	var data struct {
		Providers []struct {
			BaseURL  string `json:"baseUrl"`
			APIKey   string `json:"apiKey"`
			IsActive bool   `json:"isActive"`
			Models   []struct {
				ID        string `json:"id"`
				Type      string `json:"type"`
				IsDefault bool   `json:"isDefault"`
			} `json:"models"`
		} `json:"providers"`
		KBSettings struct {
			ChatModel string `json:"chatModel"`
		} `json:"kbSettings"`
	}
	if parseErr := json.Unmarshal([]byte(setting.Value), &data); parseErr != nil {
		return "", "", "", fmt.Errorf("AI配置解析失败")
	}

	// If kbSettings.chatModel is set, find that specific model
	if data.KBSettings.ChatModel != "" {
		for _, p := range data.Providers {
			if !p.IsActive || p.BaseURL == "" || p.APIKey == "" {
				continue
			}
			for _, m := range p.Models {
				if m.ID == data.KBSettings.ChatModel {
					return p.BaseURL, p.APIKey, m.ID, nil
				}
			}
		}
	}

	// Find first active provider with a chat model
	for _, p := range data.Providers {
		if !p.IsActive || p.BaseURL == "" || p.APIKey == "" {
			continue
		}
		// First try to find default chat model
		for _, m := range p.Models {
			if m.Type == "chat" && m.IsDefault {
				return p.BaseURL, p.APIKey, m.ID, nil
			}
		}
		// If no default, use first chat model
		for _, m := range p.Models {
			if m.Type == "chat" {
				return p.BaseURL, p.APIKey, m.ID, nil
			}
		}
	}
	return "", "", "", fmt.Errorf("未找到可用的AI聊天模型，请在管理后台配置")
}

// GetAISettingsWithModel reads AI model settings with optional model override
func GetAISettingsWithModel(modelOverride string) (baseURL, apiKey, modelID string, err error) {
	var setting model.SystemSetting
	if dbErr := mysql.DB.Where("`key` = ?", "ai_model_settings").First(&setting).Error; dbErr != nil {
		return "", "", "", fmt.Errorf("AI模型未配置，请在管理后台配置AI模型")
	}

	var data struct {
		Providers []struct {
			BaseURL  string `json:"baseUrl"`
			APIKey   string `json:"apiKey"`
			IsActive bool   `json:"isActive"`
			Models   []struct {
				ID        string `json:"id"`
				Type      string `json:"type"`
				IsDefault bool   `json:"isDefault"`
			} `json:"models"`
		} `json:"providers"`
		KBSettings struct {
			ChatModel string `json:"chatModel"`
		} `json:"kbSettings"`
	}
	if parseErr := json.Unmarshal([]byte(setting.Value), &data); parseErr != nil {
		return "", "", "", fmt.Errorf("AI配置解析失败")
	}

	// If modelOverride is provided, find that specific model
	targetModel := modelOverride
	if targetModel == "" {
		targetModel = data.KBSettings.ChatModel
	}

	// If we have a target model, find it
	if targetModel != "" {
		for _, p := range data.Providers {
			if !p.IsActive || p.BaseURL == "" || p.APIKey == "" {
				continue
			}
			for _, m := range p.Models {
				if m.ID == targetModel {
					return p.BaseURL, p.APIKey, m.ID, nil
				}
			}
		}
	}

	// Find first active provider with a chat model
	for _, p := range data.Providers {
		if !p.IsActive || p.BaseURL == "" || p.APIKey == "" {
			continue
		}
		// First try to find default chat model
		for _, m := range p.Models {
			if m.Type == "chat" && m.IsDefault {
				return p.BaseURL, p.APIKey, m.ID, nil
			}
		}
		// If no default, use first chat model
		for _, m := range p.Models {
			if m.Type == "chat" {
				return p.BaseURL, p.APIKey, m.ID, nil
			}
		}
	}
	return "", "", "", fmt.Errorf("未找到可用的AI聊天模型，请在管理后台配置")
}

func callAIChat(systemPrompt, question string, history []ChatMessage) (string, error) {
	return callAIChatWithModel(systemPrompt, question, history, "")
}

// ─── SSE Streaming Chat ──────────────────────────────────────────────────────

// SSEEvent represents a Server-Sent Event payload
type SSEEvent struct {
	Type    string      `json:"type"`              // "chunk" | "sources" | "done" | "error"
	Content string      `json:"content,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// ChatOptions holds parameters for streaming chat
type ChatOptions struct {
	UserID          uint
	ConversationID  uint   // 0 means create new conversation
	KnowledgeBaseID *uint  // nil means global
	Question        string
	History         []ChatMessage
	ModelOverride   string
}

// StreamChatWithKnowledge performs RAG-based streaming chat.
// 1. Execute retrieval (hybrid/keyword search)
// 2. Build system prompt with document context
// 3. Call AI API with stream:true
// 4. Push chunks via writer callback
// 5. Persist messages to database after completion
func StreamChatWithKnowledge(ctx context.Context, opts ChatOptions, writer func(SSEEvent)) error {
	start := time.Now()

	// Determine which KB to search
	var kbIDs []uint
	if opts.KnowledgeBaseID != nil {
		kbIDs = []uint{*opts.KnowledgeBaseID}
	} else {
		// Global: search all KBs the user is a member of
		var members []model.KnowledgeMember
		mysql.DB.Where("user_id = ?", opts.UserID).Find(&members)
		for _, m := range members {
			kbIDs = append(kbIDs, m.KnowledgeBaseID)
		}
	}

	if len(kbIDs) == 0 {
		writer(SSEEvent{Type: "error", Content: "您还没有加入任何知识库，请先创建或加入知识库"})
		return fmt.Errorf("no knowledge bases")
	}

	// Check if any chunks exist
	var totalChunks int64
	mysql.DB.Model(&model.KnowledgeChunk{}).Where("knowledge_base_id IN ?", kbIDs).Count(&totalChunks)
	if totalChunks == 0 {
		writer(SSEEvent{Type: "error", Content: "当前知识库内没有内容，请先添加文档并同步"})
		return fmt.Errorf("no content in knowledge base")
	}

	// 1. Execute retrieval
	var chunks []ChunkRef
	for _, kbID := range kbIDs {
		// Try vector search first
		if qdrant.DefaultVectorClient != nil {
			embedBaseURL, embedAPIKey, embedModelID, err := GetEmbeddingModelSettings()
			if err == nil && embedAPIKey != "" {
				embedSvc := NewEmbeddingService(EmbeddingConfig{
					BaseURL: embedBaseURL,
					APIKey:  embedAPIKey,
					Model:   embedModelID,
				})
				retrievalSvc := NewRetrievalService(embedSvc, qdrant.DefaultVectorClient)
				searchCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
				results, searchErr := retrievalSvc.HybridSearch(searchCtx, kbID, opts.Question, DefaultHybridSearchOptions())
				cancel()
				if searchErr == nil && len(results) > 0 {
					for _, r := range results {
						chunks = append(chunks, ChunkRef{
							DocumentID:    r.DocumentID,
							DocumentTitle: r.DocumentTitle,
							Content:       r.Content,
							ChunkIndex:    r.ChunkIndex,
						})
					}
				}
			}
		}
		// Fallback to keyword search
		if len(chunks) == 0 {
			chunks = append(chunks, SearchChunks(kbID, opts.Question, 6)...)
		}
	}

	// 2. Build context and system prompt
	var contextParts []string
	for i, c := range chunks {
		contextParts = append(contextParts, fmt.Sprintf("[%d] 来自《%s》:\n%s", i+1, c.DocumentTitle, c.Content))
	}
	contextText := strings.Join(contextParts, "\n\n")

	systemPrompt := "你是一个知识库助手，根据提供的文档内容准确回答用户问题。回答时请引用来源编号，格式为「来源[n]」。如果文档中没有相关信息，请如实说明。"
	if contextText != "" {
		systemPrompt += "\n\n以下是相关文档内容：\n" + contextText
	}

	// Send sources event
	writer(SSEEvent{Type: "sources", Data: chunks})

	// 3. Call AI API with streaming
	fullContent, tokensUsed, err := callAIChatStream(ctx, systemPrompt, opts.Question, opts.History, opts.ModelOverride, func(chunk string) {
		writer(SSEEvent{Type: "chunk", Content: chunk})
	})
	if err != nil {
		writer(SSEEvent{Type: "error", Content: fmt.Sprintf("AI调用失败: %v", err)})
		return err
	}

	durationMs := uint(time.Since(start).Milliseconds())

	// 5. Persist messages to database
	var conversationID uint
	if opts.ConversationID > 0 {
		conversationID = opts.ConversationID
	} else {
		// Create new conversation
		title := opts.Question
		if len([]rune(title)) > 50 {
			title = string([]rune(title)[:50]) + "..."
		}
		conv := model.KbConversation{
			UserID:          opts.UserID,
			KnowledgeBaseID: opts.KnowledgeBaseID,
			Title:           title,
			Model:           opts.ModelOverride,
			MessageCount:    0,
		}
		if dbErr := mysql.DB.Create(&conv).Error; dbErr != nil {
			logger.Errorf("Failed to create conversation: %v", dbErr)
		} else {
			conversationID = conv.ID
		}
	}

	var messageID uint
	if conversationID > 0 {
		// Save user message
		userMsg := model.KbMessage{
			ConversationID: conversationID,
			Role:           "user",
			Content:        opts.Question,
		}
		mysql.DB.Create(&userMsg)

		// Save assistant message
		var sourcesJSON *string
		if len(chunks) > 0 {
			if sj, e := json.Marshal(chunks); e == nil {
				s := string(sj)
				sourcesJSON = &s
			}
		}
		assistantMsg := model.KbMessage{
			ConversationID: conversationID,
			Role:           "assistant",
			Content:        fullContent,
			Sources:        sourcesJSON,
			TokensUsed:     uint(tokensUsed),
			DurationMs:     durationMs,
		}
		mysql.DB.Create(&assistantMsg)
		messageID = assistantMsg.ID

		// Update conversation message count
		mysql.DB.Model(&model.KbConversation{}).Where("id = ?", conversationID).
			Updates(map[string]interface{}{
				"message_count": mysql.DB.Raw("message_count + 2"),
				"updated_at":    time.Now(),
			})
	}

	// Send done event
	writer(SSEEvent{Type: "done", Data: map[string]interface{}{
		"messageId":      messageID,
		"conversationId": conversationID,
		"tokensUsed":     tokensUsed,
		"durationMs":     durationMs,
	}})

	// Generate follow-up suggestions after done event (if enabled)
	if IsFollowUpSuggestionsEnabled() {
		suggestions, sugErr := GenerateFollowUpSuggestions(opts.Question, fullContent, opts.History, opts.ModelOverride)
		if sugErr == nil && len(suggestions) > 0 {
			writer(SSEEvent{Type: "suggestions", Data: suggestions})
		}
	}

	return nil
}

// callAIChatStream calls the AI API with stream:true and invokes onChunk for each delta.
// Returns the full accumulated content, estimated token count, and any error.
func callAIChatStream(ctx context.Context, systemPrompt, question string, history []ChatMessage, modelOverride string, onChunk func(string)) (string, int, error) {
	baseURL, apiKey, modelID, err := GetAISettingsWithModel(modelOverride)
	if err != nil {
		return "", 0, err
	}

	messages := []map[string]string{
		{"role": "system", "content": systemPrompt},
	}
	for _, h := range history {
		messages = append(messages, map[string]string{"role": h.Role, "content": h.Content})
	}
	messages = append(messages, map[string]string{"role": "user", "content": question})

	reqBody := map[string]interface{}{
		"model":       modelID,
		"messages":    messages,
		"max_tokens":  2000,
		"temperature": 0.7,
		"stream":      true,
	}
	reqData, _ := json.Marshal(reqBody)

	baseURL = strings.TrimSuffix(baseURL, "/")
	chatURL := baseURL + "/chat/completions"

	httpReq, err := http.NewRequestWithContext(ctx, "POST", chatURL, bytes.NewReader(reqData))
	if err != nil {
		return "", 0, fmt.Errorf("创建请求失败: %v", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", 0, fmt.Errorf("AI服务连接失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", 0, fmt.Errorf("AI API返回错误 %d: %s", resp.StatusCode, string(body))
	}

	// Parse SSE stream
	var fullContent strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 64*1024), 64*1024)

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		// Parse data lines
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")

		// Check for stream end
		if data == "[DONE]" {
			break
		}

		// Parse JSON chunk
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
				FinishReason *string `json:"finish_reason"`
			} `json:"choices"`
			Usage *struct {
				TotalTokens int `json:"total_tokens"`
			} `json:"usage"`
		}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			content := chunk.Choices[0].Delta.Content
			fullContent.WriteString(content)
			onChunk(content)
		}
	}

	if err := scanner.Err(); err != nil {
		return fullContent.String(), 0, fmt.Errorf("读取流失败: %v", err)
	}

	// Estimate tokens (rough: 1 token ≈ 1.5 chars for Chinese)
	resultText := fullContent.String()
	estimatedTokens := len([]rune(resultText))*2/3 + len([]rune(question))*2/3

	return resultText, estimatedTokens, nil
}

// callAIChatLite performs a lightweight AI call with custom maxTokens and temperature.
// Used for auxiliary tasks like generating follow-up suggestions.
func callAIChatLite(prompt string, modelOverride string, maxTokens int, temperature float64) (string, error) {
	baseURL, apiKey, modelID, err := GetAISettingsWithModel(modelOverride)
	if err != nil {
		return "", err
	}

	messages := []map[string]string{
		{"role": "user", "content": prompt},
	}

	reqBody := map[string]interface{}{
		"model":       modelID,
		"messages":    messages,
		"max_tokens":  maxTokens,
		"temperature": temperature,
	}
	reqData, _ := json.Marshal(reqBody)

	baseURL = strings.TrimSuffix(baseURL, "/")
	chatURL := baseURL + "/chat/completions"
	httpReq, err := http.NewRequest("POST", chatURL, bytes.NewReader(reqData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("AI服务连接失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("AI API返回错误 %d: %s", resp.StatusCode, string(body))
	}

	var chatResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("解析AI响应失败: %v", err)
	}
	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("AI未返回任何内容")
	}
	return chatResp.Choices[0].Message.Content, nil
}

// GenerateFollowUpSuggestions generates 2-3 follow-up question suggestions
// based on the conversation context.
func GenerateFollowUpSuggestions(question, answer string, history []ChatMessage, modelOverride string) ([]string, error) {
	// Truncate answer if too long to save tokens
	answerRunes := []rune(answer)
	if len(answerRunes) > 500 {
		answer = string(answerRunes[:500]) + "..."
	}

	prompt := `基于以下对话内容，生成2-3个用户可能想继续追问的简短问题。
要求：
1. 问题要自然、有延展性
2. 与当前话题相关但角度不同
3. 每个问题不超过20个字
4. 直接返回JSON数组格式，如：["问题1", "问题2", "问题3"]

用户问题：` + question + `
AI回答：` + answer

	result, err := callAIChatLite(prompt, modelOverride, 500, 0.8)
	if err != nil {
		return nil, err
	}

	// Try to parse as JSON array
	var suggestions []string
	if parseErr := json.Unmarshal([]byte(strings.TrimSpace(result)), &suggestions); parseErr == nil {
		return suggestions, nil
	}

	// Fallback: try to extract JSON array from the response using regex
	re := regexp.MustCompile(`\[\s*"[^"]*"(?:\s*,\s*"[^"]*")*\s*\]`)
	matched := re.FindString(result)
	if matched != "" {
		if parseErr := json.Unmarshal([]byte(matched), &suggestions); parseErr == nil {
			return suggestions, nil
		}
	}

	// Could not parse suggestions
	return nil, fmt.Errorf("failed to parse suggestions from AI response")
}

// IsFollowUpSuggestionsEnabled checks if follow-up suggestions are enabled in system settings.
func IsFollowUpSuggestionsEnabled() bool {
	var setting model.SystemSetting
	if err := mysql.DB.Where("`key` = ?", "ai_model_settings").First(&setting).Error; err != nil {
		return true // default enabled
	}

	var data struct {
		KBSettings struct {
			EnableFollowUpSuggestions *bool `json:"enableFollowUpSuggestions"`
		} `json:"kbSettings"`
	}
	if err := json.Unmarshal([]byte(setting.Value), &data); err != nil {
		return true // default enabled
	}

	if data.KBSettings.EnableFollowUpSuggestions == nil {
		return true // default enabled
	}
	return *data.KBSettings.EnableFollowUpSuggestions
}

// UpdateFollowUpSuggestionsEnabled updates the follow-up suggestions setting.
func UpdateFollowUpSuggestionsEnabled(enabled bool) error {
	var setting model.SystemSetting
	if err := mysql.DB.Where("`key` = ?", "ai_model_settings").First(&setting).Error; err != nil {
		// Create with default structure
		data := map[string]interface{}{
			"providers": []interface{}{},
			"kbSettings": map[string]interface{}{
				"enableFollowUpSuggestions": enabled,
			},
		}
		jsonData, _ := json.Marshal(data)
		setting = model.SystemSetting{Key: "ai_model_settings", Value: string(jsonData)}
		return mysql.DB.Create(&setting).Error
	}

	// Parse existing and update
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(setting.Value), &data); err != nil {
		return err
	}

	kbSettings, ok := data["kbSettings"].(map[string]interface{})
	if !ok {
		kbSettings = map[string]interface{}{}
	}
	kbSettings["enableFollowUpSuggestions"] = enabled
	data["kbSettings"] = kbSettings

	jsonData, _ := json.Marshal(data)
	return mysql.DB.Model(&model.SystemSetting{}).Where("`key` = ?", "ai_model_settings").Update("value", string(jsonData)).Error
}

// callAIChatWithModel calls AI chat API with optional model override
func callAIChatWithModel(systemPrompt, question string, history []ChatMessage, modelOverride string) (string, error) {
	baseURL, apiKey, modelID, err := GetAISettingsWithModel(modelOverride)
	if err != nil {
		return "", err
	}

	messages := []map[string]string{
		{"role": "system", "content": systemPrompt},
	}
	for _, h := range history {
		messages = append(messages, map[string]string{"role": h.Role, "content": h.Content})
	}
	messages = append(messages, map[string]string{"role": "user", "content": question})

	reqBody := map[string]interface{}{
		"model":       modelID,
		"messages":    messages,
		"max_tokens":  2000,
		"temperature": 0.7,
	}
	reqData, _ := json.Marshal(reqBody)

	baseURL = strings.TrimSuffix(baseURL, "/")
	chatURL := baseURL + "/chat/completions"
	httpReq, err := http.NewRequest("POST", chatURL, bytes.NewReader(reqData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("AI服务连接失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("AI API返回错误 %d: %s", resp.StatusCode, string(body))
	}

	var chatResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("解析AI响应失败: %v", err)
	}
	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("AI未返回任何内容")
	}
	return chatResp.Choices[0].Message.Content, nil
}
