package service

import (
	"context"
	"sort"
	"strings"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/repository/mysql"
	"kloggerx-server/internal/repository/qdrant"
)

// RetrievalService provides hybrid search capabilities
type RetrievalService struct {
	embeddingSvc *EmbeddingService
	vectorClient *qdrant.VectorClient
}

// NewRetrievalService creates a new RetrievalService
func NewRetrievalService(embeddingSvc *EmbeddingService, vectorClient *qdrant.VectorClient) *RetrievalService {
	return &RetrievalService{
		embeddingSvc: embeddingSvc,
		vectorClient: vectorClient,
	}
}

// SearchResult represents a search result with ranking information
type SearchResult struct {
	ChunkID       uint    `json:"chunkId"`
	DocumentID    uint    `json:"documentId"`
	DocumentTitle string  `json:"documentTitle"`
	Content       string  `json:"content"`
	ChunkIndex    int     `json:"chunkIndex"`
	Score         float64 `json:"score"`
	Source        string  `json:"source"` // "vector", "fulltext", or "hybrid"
	VectorScore   float64 `json:"vectorScore,omitempty"`
	FulltextScore float64 `json:"fulltextScore,omitempty"`
}

// HybridSearchOptions holds options for hybrid search
type HybridSearchOptions struct {
	TopK         int
	VectorWeight float64 // Weight for vector search results (default: 0.5)
	RRF_K        int     // RRF constant (default: 60)
	RaptorLevel  *int    // Filter by RAPTOR level (nil = all levels)
}

// DefaultHybridSearchOptions returns default search options
func DefaultHybridSearchOptions() HybridSearchOptions {
	return HybridSearchOptions{
		TopK:         6,
		VectorWeight: 0.5,
		RRF_K:        60,
	}
}

// HybridSearch performs hybrid search combining vector and full-text search
func (s *RetrievalService) HybridSearch(ctx context.Context, kbID uint, query string, opts HybridSearchOptions) ([]SearchResult, error) {
	if opts.TopK <= 0 {
		opts.TopK = 6
	}
	if opts.RRF_K <= 0 {
		opts.RRF_K = 60
	}
	if opts.VectorWeight <= 0 {
		opts.VectorWeight = 0.5
	}

	// Run vector and fulltext searches in parallel
	type searchResult struct {
		results []SearchResult
		err     error
		name    string
	}

	vectorCh := make(chan searchResult, 1)
	fulltextCh := make(chan searchResult, 1)

	// Vector search
	go func() {
		if s.embeddingSvc == nil || s.vectorClient == nil {
			vectorCh <- searchResult{nil, nil, "vector"}
			return
		}

		// Generate query embedding
		embedding, err := s.embeddingSvc.GenerateEmbedding(ctx, query)
		if err != nil {
			vectorCh <- searchResult{nil, err, "vector"}
			return
		}

		// Search in Qdrant
		vectorResults, err := s.vectorClient.Search(ctx, kbID, embedding, opts.TopK*2, opts.RaptorLevel)
		if err != nil {
			vectorCh <- searchResult{nil, err, "vector"}
			return
		}

		// Convert to SearchResult
		results := make([]SearchResult, len(vectorResults))
		for i, r := range vectorResults {
			results[i] = SearchResult{
				ChunkID:       r.ChunkID,
				DocumentID:    r.DocumentID,
				DocumentTitle: r.DocumentTitle,
				Content:       r.Content,
				ChunkIndex:    r.ChunkIndex,
				Score:         r.Score,
				Source:        "vector",
				VectorScore:   r.Score,
			}
		}
		vectorCh <- searchResult{results, nil, "vector"}
	}()

	// Full-text search
	go func() {
		results, err := s.fulltextSearch(kbID, query, opts.TopK*2)
		fulltextCh <- searchResult{results, err, "fulltext"}
	}()

	// Collect results
	vectorResult := <-vectorCh
	fulltextResult := <-fulltextCh

	// Handle errors gracefully (use available results)
	var vectorResults, fulltextResults []SearchResult
	if vectorResult.err == nil {
		vectorResults = vectorResult.results
	}
	if fulltextResult.err == nil {
		fulltextResults = fulltextResult.results
	}

	// If no results from either, return empty
	if len(vectorResults) == 0 && len(fulltextResults) == 0 {
		return nil, nil
	}

	// Apply RRF fusion
	merged := s.rrfFusion(vectorResults, fulltextResults, opts.RRF_K)

	// Return top K results
	if len(merged) > opts.TopK {
		merged = merged[:opts.TopK]
	}

	return merged, nil
}

// fulltextSearch performs MySQL FULLTEXT search
func (s *RetrievalService) fulltextSearch(kbID uint, query string, limit int) ([]SearchResult, error) {
	// Extract keywords from query
	keywords := extractKeywords(query)
	if len(keywords) == 0 {
		return nil, nil
	}

	// Build search condition
	var conditions []string
	var args []interface{}
	for _, kw := range keywords {
		conditions = append(conditions, "content LIKE ?")
		args = append(args, "%"+kw+"%")
	}

	// Get chunk IDs from knowledge base
	var chunks []model.KnowledgeChunk
	whereClause := "knowledge_base_id = ? AND (" + strings.Join(conditions, " OR ") + ")"
	args = append([]interface{}{kbID}, args...)

	if err := mysql.DB.Where(whereClause, args...).
		Order("updated_at DESC").
		Limit(limit).
		Find(&chunks).Error; err != nil {
		return nil, err
	}

	// Convert to SearchResult
	results := make([]SearchResult, len(chunks))
	for i, c := range chunks {
		// Calculate a simple relevance score based on keyword matches
		score := s.calculateFulltextScore(c.Content, keywords)
		results[i] = SearchResult{
			ChunkID:       c.ID,
			DocumentID:    c.DocumentID,
			DocumentTitle: c.DocumentTitle,
			Content:       c.Content,
			ChunkIndex:    c.ChunkIndex,
			Score:         score,
			Source:        "fulltext",
			FulltextScore: score,
		}
	}

	return results, nil
}

// calculateFulltextScore calculates a simple relevance score for fulltext results
func (s *RetrievalService) calculateFulltextScore(content string, keywords []string) float64 {
	contentLower := strings.ToLower(content)
	score := 0.0
	for _, kw := range keywords {
		count := strings.Count(contentLower, strings.ToLower(kw))
		score += float64(count)
	}
	// Normalize to 0-1 range roughly
	return score / 10.0
}

// rrfFusion applies Reciprocal Rank Fusion to merge search results
func (s *RetrievalService) rrfFusion(vectorResults, fulltextResults []SearchResult, k int) []SearchResult {
	// Map to track combined scores by chunk ID
	scores := make(map[uint]*SearchResult)

	// Process vector results
	for i, r := range vectorResults {
		rrfScore := 1.0 / float64(k+i+1)
		if existing, ok := scores[r.ChunkID]; ok {
			existing.Score += rrfScore
			existing.VectorScore = r.VectorScore
			existing.Source = "hybrid"
		} else {
			r.Score = rrfScore
			r.Source = "vector"
			scores[r.ChunkID] = &r
		}
	}

	// Process fulltext results
	for i, r := range fulltextResults {
		rrfScore := 1.0 / float64(k+i+1)
		if existing, ok := scores[r.ChunkID]; ok {
			existing.Score += rrfScore
			existing.FulltextScore = r.FulltextScore
			existing.Source = "hybrid"
		} else {
			r.Score = rrfScore
			r.Source = "fulltext"
			scores[r.ChunkID] = &r
		}
	}

	// Convert to slice and sort by score
	results := make([]SearchResult, 0, len(scores))
	for _, r := range scores {
		results = append(results, *r)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

// VectorSearch performs pure vector similarity search
func (s *RetrievalService) VectorSearch(ctx context.Context, kbID uint, query string, topK int) ([]SearchResult, error) {
	if s.embeddingSvc == nil || s.vectorClient == nil {
		return nil, nil
	}

	embedding, err := s.embeddingSvc.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, err
	}

	vectorResults, err := s.vectorClient.Search(ctx, kbID, embedding, topK, nil)
	if err != nil {
		return nil, err
	}

	results := make([]SearchResult, len(vectorResults))
	for i, r := range vectorResults {
		results[i] = SearchResult{
			ChunkID:       r.ChunkID,
			DocumentID:    r.DocumentID,
			DocumentTitle: r.DocumentTitle,
			Content:       r.Content,
			ChunkIndex:    r.ChunkIndex,
			Score:         r.Score,
			Source:        "vector",
			VectorScore:   r.Score,
		}
	}

	return results, nil
}

// Global retrieval service instance
var globalRetrievalService *RetrievalService

// InitRetrievalService initializes the global retrieval service
func InitRetrievalService(embeddingSvc *EmbeddingService, vectorClient *qdrant.VectorClient) {
	globalRetrievalService = NewRetrievalService(embeddingSvc, vectorClient)
}

// GetRetrievalService returns the global retrieval service
func GetRetrievalService() *RetrievalService {
	return globalRetrievalService
}
