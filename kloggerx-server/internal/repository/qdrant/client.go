package qdrant

import (
	"context"
	"fmt"
)

const (
	DefaultVectorSize = 1536 // OpenAI text-embedding-3-small dimension
	DistanceCos       = "Cosine"
)

// VectorSize is the current vector dimension (can be set dynamically)
var VectorSize = DefaultVectorSize

// SetVectorSize sets the vector dimension
func SetVectorSize(size int) {
	VectorSize = size
}

// VectorPoint represents a vector point with metadata
type VectorPoint struct {
	ID            string
	Vector        []float32
	Payload       map[string]interface{}
	DocumentID    uint
	ChunkID       uint
	ChunkIndex    int
	Content       string
	DocumentTitle string
	KBID          uint
	RaptorLevel   int // 0 for normal chunks, >0 for RAPTOR summaries
}

// VectorClient wraps Qdrant operations for KLoggerX
type VectorClient struct {
	collectionName string
}

// NewVectorClient creates a new VectorClient
func NewVectorClient(collectionName string) *VectorClient {
	return &VectorClient{collectionName: collectionName}
}

// DefaultVectorClient is the default vector client using the global Client
var DefaultVectorClient *VectorClient

// InitDefaultVectorClient initializes the default vector client
func InitDefaultVectorClient(collectionName string) {
	DefaultVectorClient = NewVectorClient(collectionName)
}

// EnsureCollection creates the collection if it doesn't exist
func (vc *VectorClient) EnsureCollection(ctx context.Context) error {
	collections, err := GetCollections(ctx)
	if err != nil {
		return fmt.Errorf("failed to get collections: %w", err)
	}

	// Check if collection exists
	for _, col := range collections {
		if col.Name == vc.collectionName {
			return nil
		}
	}

	// Create collection
	return CreateCollection(ctx, vc.collectionName, VectorSize)
}

// UpsertVector inserts or updates a vector point
func (vc *VectorClient) UpsertVector(ctx context.Context, point *VectorPoint) (string, error) {
	if point.ID == "" {
		point.ID = GenerateUUID()
	}

	// Ensure collection exists
	if err := vc.EnsureCollection(ctx); err != nil {
		return "", err
	}

	// Build payload
	payload := map[string]interface{}{
		"kb_id":          int64(point.KBID),
		"doc_id":         int64(point.DocumentID),
		"chunk_id":       int64(point.ChunkID),
		"chunk_index":    int64(point.ChunkIndex),
		"content":        point.Content,
		"document_title": point.DocumentTitle,
		"raptor_level":   int64(point.RaptorLevel),
	}

	// Upsert point
	p := Point{
		ID:      point.ID,
		Vector:  point.Vector,
		Payload: payload,
	}

	if err := UpsertPoints(ctx, vc.collectionName, []Point{p}); err != nil {
		return "", fmt.Errorf("failed to upsert vector: %w", err)
	}

	return point.ID, nil
}

// BatchUpsertVectors inserts or updates multiple vector points
func (vc *VectorClient) BatchUpsertVectors(ctx context.Context, points []*VectorPoint) ([]string, error) {
	if len(points) == 0 {
		return nil, nil
	}

	// Ensure collection exists
	if err := vc.EnsureCollection(ctx); err != nil {
		return nil, err
	}

	// Convert to qdrant points
	qdrantPoints := make([]Point, len(points))
	pointIDs := make([]string, len(points))

	for i, point := range points {
		if point.ID == "" {
			point.ID = GenerateUUID()
		}
		pointIDs[i] = point.ID

		payload := map[string]interface{}{
			"kb_id":          int64(point.KBID),
			"doc_id":         int64(point.DocumentID),
			"chunk_id":       int64(point.ChunkID),
			"chunk_index":    int64(point.ChunkIndex),
			"content":        point.Content,
			"document_title": point.DocumentTitle,
			"raptor_level":   int64(point.RaptorLevel),
		}

		qdrantPoints[i] = Point{
			ID:      point.ID,
			Vector:  point.Vector,
			Payload: payload,
		}
	}

	// Upsert in batches of 100
	batchSize := 100
	for i := 0; i < len(qdrantPoints); i += batchSize {
		end := i + batchSize
		if end > len(qdrantPoints) {
			end = len(qdrantPoints)
		}

		if err := UpsertPoints(ctx, vc.collectionName, qdrantPoints[i:end]); err != nil {
			return nil, fmt.Errorf("failed to batch upsert vectors: %w", err)
		}
	}

	return pointIDs, nil
}

// SearchResult represents a search result
type SearchResult struct {
	PointID       string
	Score         float64
	DocumentID    uint
	ChunkID       uint
	ChunkIndex    int
	Content       string
	DocumentTitle string
	KBID          uint
	RaptorLevel   int
	QualityScore  float64 // quality_score from payload, default 1.0
}

// Search performs a vector similarity search
func (vc *VectorClient) Search(ctx context.Context, kbID uint, vector []float32, topK int, raptorLevel *int) ([]SearchResult, error) {
	// Build filter for kb_id
	filter := map[string]interface{}{
		"must": []map[string]interface{}{
			{
				"key": "kb_id",
				"match": map[string]interface{}{
					"value": int64(kbID),
				},
			},
		},
	}

	// Add raptor_level filter if specified
	if raptorLevel != nil {
		filter["must"] = append(filter["must"].([]map[string]interface{}), map[string]interface{}{
			"key": "raptor_level",
			"match": map[string]interface{}{
				"value": int64(*raptorLevel),
			},
		})
	}

	// Perform search
	resp, err := SearchPoints(ctx, vc.collectionName, vector, topK, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to search vectors: %w", err)
	}

	// Convert results
	searchResults := make([]SearchResult, len(resp.Result))
	for i, result := range resp.Result {
		searchResults[i] = SearchResult{
			PointID:       result.ID,
			Score:         result.Score,
			DocumentID:    getPayloadUint(result.Payload, "doc_id"),
			ChunkID:       getPayloadUint(result.Payload, "chunk_id"),
			ChunkIndex:    getPayloadInt(result.Payload, "chunk_index"),
			Content:       getPayloadString(result.Payload, "content"),
			DocumentTitle: getPayloadString(result.Payload, "document_title"),
			KBID:          getPayloadUint(result.Payload, "kb_id"),
			RaptorLevel:   getPayloadInt(result.Payload, "raptor_level"),
			QualityScore:  getPayloadFloat(result.Payload, "quality_score", 1.0),
		}
	}

	return searchResults, nil
}

// DeleteByDocument deletes all vectors for a specific document
func (vc *VectorClient) DeleteByDocument(ctx context.Context, kbID, docID uint) error {
	filter := map[string]interface{}{
		"must": []map[string]interface{}{
			{
				"key": "kb_id",
				"match": map[string]interface{}{
					"value": int64(kbID),
				},
			},
			{
				"key": "doc_id",
				"match": map[string]interface{}{
					"value": int64(docID),
				},
			},
		},
	}

	return DeletePoints(ctx, vc.collectionName, filter)
}

// DeleteByKnowledgeBase deletes all vectors for a specific knowledge base
func (vc *VectorClient) DeleteByKnowledgeBase(ctx context.Context, kbID uint) error {
	filter := map[string]interface{}{
		"must": []map[string]interface{}{
			{
				"key": "kb_id",
				"match": map[string]interface{}{
					"value": int64(kbID),
				},
			},
		},
	}

	return DeletePoints(ctx, vc.collectionName, filter)
}

// DeleteByPointID deletes a specific vector point by ID
func (vc *VectorClient) DeleteByPointID(ctx context.Context, pointID string) error {
	return DeletePointsByIDs(ctx, vc.collectionName, []string{pointID})
}

// UpdatePayloadByPointIDs updates payload fields on specific points
func (vc *VectorClient) UpdatePayloadByPointIDs(ctx context.Context, payload map[string]interface{}, pointIDs []string) error {
	return SetPayloadByPointIDs(ctx, vc.collectionName, payload, pointIDs)
}

// UpdatePayloadByFilter updates payload fields on points matching a filter
func (vc *VectorClient) UpdatePayloadByFilter(ctx context.Context, payload map[string]interface{}, filter map[string]interface{}) error {
	return SetPayload(ctx, vc.collectionName, payload, filter)
}

// GetCollectionStats returns statistics about the collection
func (vc *VectorClient) GetCollectionStats(ctx context.Context) (map[string]interface{}, error) {
	info, err := GetCollectionInfo(ctx, vc.collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection info: %w", err)
	}

	result := make(map[string]interface{})

	// Try to extract from nested result structure
	if nested, ok := info["result"].(map[string]interface{}); ok {
		if v, ok := nested["points_count"]; ok {
			result["points_count"] = v
		}
		if v, ok := nested["vectors_count"]; ok {
			result["vectors_count"] = v
		}
		if v, ok := nested["status"]; ok {
			result["status"] = v
		}
	} else {
		// Try direct access
		if v, ok := info["points_count"]; ok {
			result["points_count"] = v
		}
		if v, ok := info["vectors_count"]; ok {
			result["vectors_count"] = v
		}
		if v, ok := info["status"]; ok {
			result["status"] = v
		}
	}

	return result, nil
}

// Helper functions for payload extraction
func getPayloadUint(payload map[string]interface{}, key string) uint {
	if v, ok := payload[key]; ok {
		switch val := v.(type) {
		case float64:
			return uint(val)
		case int64:
			return uint(val)
		}
	}
	return 0
}

func getPayloadInt(payload map[string]interface{}, key string) int {
	if v, ok := payload[key]; ok {
		switch val := v.(type) {
		case float64:
			return int(val)
		case int64:
			return int(val)
		}
	}
	return 0
}

func getPayloadString(payload map[string]interface{}, key string) string {
	if v, ok := payload[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getPayloadFloat(payload map[string]interface{}, key string, defaultVal float64) float64 {
	if v, ok := payload[key]; ok {
		switch val := v.(type) {
		case float64:
			return val
		case int64:
			return float64(val)
		}
	}
	return defaultVal
}
