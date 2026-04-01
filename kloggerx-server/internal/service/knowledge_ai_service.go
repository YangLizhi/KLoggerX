package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/repository/mysql"
)

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
func IndexDocumentContent(kbID, docID uint) {
	var doc model.Document
	if err := mysql.DB.First(&doc, docID).Error; err != nil {
		return
	}

	// Extract plain text from JSON content
	plainText := extractPlainTextFromContent(doc.Content)
	if plainText == "" {
		return
	}

	// Delete old chunks for this document in this KB
	mysql.DB.Where("knowledge_base_id = ? AND document_id = ?", kbID, docID).Delete(&model.KnowledgeChunk{})

	// Split into chunks
	chunks := splitIntoChunks(plainText, chunkSize)
	for i, chunk := range chunks {
		mysql.DB.Create(&model.KnowledgeChunk{
			KnowledgeBaseID: kbID,
			DocumentID:      docID,
			DocumentTitle:   doc.Title,
			Content:         chunk,
			ChunkIndex:      i,
		})
	}
}

func extractPlainTextFromContent(content string) string {
	if content == "" || content == "{}" {
		return ""
	}
	// Parse Tiptap/ProseMirror JSON and extract text nodes
	var doc map[string]interface{}
	if err := json.Unmarshal([]byte(content), &doc); err != nil {
		return content // treat as plain text
	}
	var sb strings.Builder
	extractTextNodes(doc, &sb)
	return strings.TrimSpace(sb.String())
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
func ChatWithKnowledge(kbID uint, question string, history []ChatMessage) (*KBChatResult, error) {
	// 1. Retrieve relevant chunks
	chunks := SearchChunks(kbID, question, 6)

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

	// 3. Call AI API
	answer, err := callAIChat(systemPrompt, question, history)
	if err != nil {
		return nil, err
	}

	return &KBChatResult{
		Answer:  answer,
		Sources: chunks,
	}, nil
}

// getAISettings reads AI model settings from the database.
func getAISettings() (baseURL, apiKey, modelID string, err error) {
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

func callAIChat(systemPrompt, question string, history []ChatMessage) (string, error) {
	baseURL, apiKey, modelID, err := getAISettings()
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
