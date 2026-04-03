package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// EmbeddingService provides text embedding generation using OpenAI API
type EmbeddingService struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	model      string
}

// EmbeddingConfig holds configuration for the embedding service
type EmbeddingConfig struct {
	BaseURL string
	APIKey  string
	Model   string // default: text-embedding-3-small
}

// NewEmbeddingService creates a new EmbeddingService
func NewEmbeddingService(cfg EmbeddingConfig) *EmbeddingService {
	if cfg.Model == "" {
		cfg.Model = "text-embedding-3-small"
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.openai.com/v1"
	}

	return &EmbeddingService{
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		baseURL: strings.TrimSuffix(cfg.BaseURL, "/"),
		apiKey:   cfg.APIKey,
		model:    cfg.Model,
	}
}

// embeddingRequest represents the request body for OpenAI embeddings API
type embeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// embeddingResponse represents the response from OpenAI embeddings API
type embeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Index     int       `json:"index"`
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

// GenerateEmbedding generates an embedding for a single text
func (s *EmbeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := s.GenerateEmbeddings(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}
	return embeddings[0], nil
}

// GenerateEmbeddings generates embeddings for multiple texts in a batch
func (s *EmbeddingService) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	// Build request
	reqBody := embeddingRequest{
		Model: s.model,
		Input: texts,
	}

	reqData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := s.baseURL + "/embeddings"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var embResp embeddingResponse
	if err := json.Unmarshal(respBody, &embResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w (body: %s)", err, string(respBody))
	}

	// Check for API error
	if embResp.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s (type: %s, code: %s)", 
			embResp.Error.Message, embResp.Error.Type, embResp.Error.Code)
	}

	// Extract embeddings in order
	embeddings := make([][]float32, len(texts))
	for _, data := range embResp.Data {
		if data.Index < len(embeddings) {
			embeddings[data.Index] = data.Embedding
		}
	}

	return embeddings, nil
}

// BatchGenerateEmbeddings generates embeddings for a large number of texts in batches
func (s *EmbeddingService) BatchGenerateEmbeddings(ctx context.Context, texts []string, batchSize int) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	if batchSize <= 0 {
		batchSize = 100 // Default batch size for OpenAI
	}

	var allEmbeddings [][]float32

	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}

		batch := texts[i:end]
		embeddings, err := s.GenerateEmbeddings(ctx, batch)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embeddings for batch %d-%d: %w", i, end, err)
		}

		allEmbeddings = append(allEmbeddings, embeddings...)
	}

	return allEmbeddings, nil
}

// GetEmbeddingDimension returns the dimension of the embeddings
func (s *EmbeddingService) GetEmbeddingDimension() int {
	return GetEmbeddingDimensionForModel(s.model)
}

// GetEmbeddingDimensionForModel returns the dimension for a given model
func GetEmbeddingDimensionForModel(model string) int {
	switch model {
	case "text-embedding-3-small":
		return 1536
	case "text-embedding-3-large":
		return 3072
	case "text-embedding-ada-002":
		return 1536
	// Common embedding models from other providers
	case "text-embedding-v1", "text-embedding-v2", "text-embedding-v3":
		return 1536 // Alibaba/DashScope
	case "bge-large-zh-v1.5", "bge-large-en-v1.5":
		return 1024
	case "bge-m3":
		return 1024
	case "embedding-001", "text-embedding-004":
		return 768 // Google
	case "voyage-large-2", "voyage-code-2":
		return 1536
	default:
		// Try to detect from model name patterns
		if strings.Contains(model, "large") {
			return 3072
		}
		return 1536 // Default dimension
	}
}

// GetModel returns the current model name
func (s *EmbeddingService) GetModel() string {
	return s.model
}

// Global embedding service instance
var globalEmbeddingService *EmbeddingService

// InitEmbeddingService initializes the global embedding service
func InitEmbeddingService(baseURL, apiKey, model string) {
	globalEmbeddingService = NewEmbeddingService(EmbeddingConfig{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Model:   model,
	})
}

// GetEmbeddingService returns the global embedding service
func GetEmbeddingService() *EmbeddingService {
	return globalEmbeddingService
}
