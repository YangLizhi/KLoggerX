package qdrant

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"kloggerx-server/config"

	"github.com/google/uuid"
)

// HTTP Client for Qdrant REST API
var httpClient *http.Client
var qdrantHost string
var qdrantPort int
var qdrantAPIKey string

// Init initializes the Qdrant HTTP client
func Init(cfg config.QdrantConfig) error {
	httpClient = &http.Client{
		Timeout: 30 * time.Second,
	}
	qdrantHost = cfg.Host
	qdrantPort = cfg.Port
	qdrantAPIKey = cfg.APIKey

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := GetCollections(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to Qdrant: %w", err)
	}

	log.Printf("[Qdrant] Connected to Qdrant at %s:%d", cfg.Host, cfg.Port)
	return nil
}

// Close is a no-op for HTTP client
func Close() {}

// doRequest performs an HTTP request to Qdrant
func doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	url := fmt.Sprintf("http://%s:%d%s", qdrantHost, qdrantPort, path)

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if qdrantAPIKey != "" {
		req.Header.Set("api-key", qdrantAPIKey)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Qdrant API error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// CollectionResponse represents a collection in the response
type CollectionResponse struct {
	Name string `json:"name"`
}

// GetCollections returns all collections
func GetCollections(ctx context.Context) ([]CollectionResponse, error) {
	respBody, err := doRequest(ctx, "GET", "/collections", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Result struct {
			Collections []CollectionResponse `json:"collections"`
		} `json:"result"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result.Result.Collections, nil
}

// CreateCollection creates a new collection
func CreateCollection(ctx context.Context, name string, vectorSize int) error {
	body := map[string]interface{}{
		"vectors": map[string]interface{}{
			"size":     vectorSize,
			"distance": "Cosine",
		},
	}

	_, err := doRequest(ctx, "PUT", fmt.Sprintf("/collections/%s", name), body)
	return err
}

// DeleteCollection deletes a collection
func DeleteCollection(ctx context.Context, name string) error {
	_, err := doRequest(ctx, "DELETE", fmt.Sprintf("/collections/%s", name), nil)
	return err
}

// Point represents a vector point
type Point struct {
	ID      string                 `json:"id"`
	Vector  []float32              `json:"vector"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

// UpsertPoints upserts points into a collection
func UpsertPoints(ctx context.Context, collection string, points []Point) error {
	body := map[string]interface{}{
		"points": points,
	}

	_, err := doRequest(ctx, "PUT", fmt.Sprintf("/collections/%s/points", collection), body)
	return err
}

// SearchPoint represents a search result point
type SearchPoint struct {
	ID      string                 `json:"id"`
	Version int64                  `json:"version"`
	Score   float64                `json:"score"`
	Payload map[string]interface{} `json:"payload"`
}

// SearchResponse represents a search response
type SearchResponse struct {
	Result []SearchPoint `json:"result"`
}

// SearchPoints searches for similar vectors
func SearchPoints(ctx context.Context, collection string, vector []float32, topK int, filter map[string]interface{}) (*SearchResponse, error) {
	body := map[string]interface{}{
		"vector":       vector,
		"limit":        topK,
		"with_payload": true,
	}

	if filter != nil {
		body["filter"] = filter
	}

	respBody, err := doRequest(ctx, "POST", fmt.Sprintf("/collections/%s/points/search", collection), body)
	if err != nil {
		return nil, err
	}

	var result SearchResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	return &result, nil
}

// DeletePoints deletes points from a collection
func DeletePoints(ctx context.Context, collection string, filter map[string]interface{}) error {
	body := map[string]interface{}{
		"filter": filter,
	}

	_, err := doRequest(ctx, "POST", fmt.Sprintf("/collections/%s/points/delete", collection), body)
	return err
}

// DeletePointsByIDs deletes points by their IDs
func DeletePointsByIDs(ctx context.Context, collection string, ids []string) error {
	body := map[string]interface{}{
		"points": ids,
	}

	_, err := doRequest(ctx, "POST", fmt.Sprintf("/collections/%s/points/delete", collection), body)
	return err
}

// SetPayload sets payload fields on points matching a filter
func SetPayload(ctx context.Context, collection string, payload map[string]interface{}, filter map[string]interface{}) error {
	body := map[string]interface{}{
		"payload": payload,
		"filter":  filter,
	}

	_, err := doRequest(ctx, "POST", fmt.Sprintf("/collections/%s/points/payload", collection), body)
	return err
}

// SetPayloadByPointIDs sets payload fields on specific points by IDs
func SetPayloadByPointIDs(ctx context.Context, collection string, payload map[string]interface{}, pointIDs []string) error {
	body := map[string]interface{}{
		"payload": payload,
		"points":  pointIDs,
	}

	_, err := doRequest(ctx, "POST", fmt.Sprintf("/collections/%s/points/payload", collection), body)
	return err
}

// GetCollectionInfo returns collection info
func GetCollectionInfo(ctx context.Context, collection string) (map[string]interface{}, error) {
	respBody, err := doRequest(ctx, "GET", fmt.Sprintf("/collections/%s", collection), nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Result map[string]interface{} `json:"result"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse collection info: %w", err)
	}

	return result.Result, nil
}

// GenerateUUID generates a new UUID string
func GenerateUUID() string {
	return uuid.New().String()
}
