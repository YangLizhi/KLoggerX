package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/repository/mysql"
	"kloggerx-server/internal/repository/qdrant"
)

// RaptorService provides RAPTOR (Recursive Abstractive Processing for Tree-Organized Retrieval) capabilities
type RaptorService struct {
	embeddingSvc *EmbeddingService
	vectorClient *qdrant.VectorClient
}

// NewRaptorService creates a new RaptorService
func NewRaptorService(embeddingSvc *EmbeddingService, vectorClient *qdrant.VectorClient) *RaptorService {
	return &RaptorService{
		embeddingSvc: embeddingSvc,
		vectorClient: vectorClient,
	}
}

// RaptorConfig holds configuration for RAPTOR tree building
type RaptorConfig struct {
	MaxLevel          int     // Maximum tree depth (default: 3)
	ClusterThreshold  float64 // Similarity threshold for clustering (default: 0.5)
	MaxClusterSize    int     // Maximum chunks per cluster (default: 10)
	SummaryMaxTokens  int     // Max tokens for summaries (default: 500)
	MinChunksForTree  int     // Minimum chunks to build tree (default: 5)
}

// DefaultRaptorConfig returns default RAPTOR configuration
func DefaultRaptorConfig() RaptorConfig {
	return RaptorConfig{
		MaxLevel:         3,
		ClusterThreshold: 0.5,
		MaxClusterSize:   10,
		SummaryMaxTokens: 500,
		MinChunksForTree: 5,
	}
}

// RaptorNode represents a node in the RAPTOR tree
type RaptorNode struct {
	ID          uint
	KBID        uint
	Level       int
	NodeType    string // "leaf", "cluster", "root"
	ChunkIDs    []uint // Original chunk IDs
	DocumentIDs []uint // Related document IDs
	Content     string // Summary or original content
	Embedding   []float32
	PointID     string // Qdrant point ID
	ParentID    *uint
	ChildIDs    []uint
}

// Cluster represents a cluster of similar items
type Cluster struct {
	ID       int
	Members  []int // Indices into the original slice
	Centroid []float32
}

// ClusterMember represents a member of a cluster
type ClusterMember struct {
	ChunkID       uint
	DocumentID    uint
	DocumentTitle string
	Content       string
	Embedding     []float32
}

// BuildRaptorTree builds a RAPTOR tree for a knowledge base
func (s *RaptorService) BuildRaptorTree(ctx context.Context, kbID uint, config RaptorConfig) error {
	// 1. Get all embedded chunks for the knowledge base
	chunks, err := s.getEmbeddedChunks(kbID)
	if err != nil {
		return fmt.Errorf("failed to get embedded chunks: %w", err)
	}

	if len(chunks) < config.MinChunksForTree {
		return fmt.Errorf("not enough chunks (%d) to build RAPTOR tree, need at least %d", len(chunks), config.MinChunksForTree)
	}

	// 2. Delete existing RAPTOR nodes for this KB
	s.deleteExistingNodes(kbID)

	// 3. Build tree level by level
	currentLevelNodes := s.createLeafNodes(chunks)
	level := 0

	for len(currentLevelNodes) > 1 && level < config.MaxLevel {
		level++
		parentNodes, err := s.buildNextLevel(ctx, kbID, currentLevelNodes, level, config)
		if err != nil {
			return fmt.Errorf("failed to build level %d: %w", level, err)
		}
		currentLevelNodes = parentNodes
	}

	return nil
}

// getEmbeddedChunks retrieves all embedded chunks for a knowledge base
func (s *RaptorService) getEmbeddedChunks(kbID uint) ([]ClusterMember, error) {
	var chunks []model.KnowledgeChunk
	if err := mysql.DB.Where("knowledge_base_id = ? AND embedding_status = ?", kbID, "embedded").
		Order("id ASC").
		Find(&chunks).Error; err != nil {
		return nil, err
	}

	members := make([]ClusterMember, 0, len(chunks))
	for _, chunk := range chunks {
		if chunk.QdrantPointID != "" {
			members = append(members, ClusterMember{
				ChunkID:       chunk.ID,
				DocumentID:    chunk.DocumentID,
				DocumentTitle: chunk.DocumentTitle,
				Content:       chunk.Content,
			})
		}
	}

	return members, nil
}

// deleteExistingNodes removes existing RAPTOR nodes for a KB
func (s *RaptorService) deleteExistingNodes(kbID uint) {
	mysql.DB.Where("knowledge_base_id = ?", kbID).Delete(&model.RaptorNode{})
}

// createLeafNodes creates leaf nodes from chunks (level 0)
func (s *RaptorService) createLeafNodes(chunks []ClusterMember) []*RaptorNode {
	nodes := make([]*RaptorNode, len(chunks))
	for i, chunk := range chunks {
		nodes[i] = &RaptorNode{
			Level:       0,
			NodeType:    "leaf",
			ChunkIDs:    []uint{chunk.ChunkID},
			DocumentIDs: []uint{chunk.DocumentID},
			Content:     chunk.Content,
		}
	}
	return nodes
}

// buildNextLevel builds the next level of the RAPTOR tree
func (s *RaptorService) buildNextLevel(ctx context.Context, kbID uint, nodes []*RaptorNode, level int, config RaptorConfig) ([]*RaptorNode, error) {
	// 1. Get embeddings for all nodes
	embeddings, err := s.getNodeEmbeddings(ctx, nodes)
	if err != nil {
		return nil, fmt.Errorf("failed to get embeddings: %w", err)
	}

	// 2. Cluster nodes based on embedding similarity
	clusters := s.clusterNodes(nodes, embeddings, config)

	// 3. Create parent nodes from clusters
	parentNodes := make([]*RaptorNode, 0, len(clusters))

	for _, cluster := range clusters {
		// Collect content from cluster members
		var contents []string
		var chunkIDs []uint
		var docIDs []uint
		docIDSet := make(map[uint]bool)

		for _, memberIdx := range cluster.Members {
			node := nodes[memberIdx]
			contents = append(contents, node.Content)
			chunkIDs = append(chunkIDs, node.ChunkIDs...)
			for _, docID := range node.DocumentIDs {
				if !docIDSet[docID] {
					docIDs = append(docIDs, docID)
					docIDSet[docID] = true
				}
			}
		}

		// Generate summary using AI
		summary, err := s.generateSummary(ctx, contents, config.SummaryMaxTokens)
		if err != nil {
			return nil, fmt.Errorf("failed to generate summary: %w", err)
		}

		// Generate embedding for summary
		summaryEmbedding, err := s.embeddingSvc.GenerateEmbedding(ctx, summary)
		if err != nil {
			return nil, fmt.Errorf("failed to generate summary embedding: %w", err)
		}

		// Create parent node
		parentNode := &RaptorNode{
			KBID:        kbID,
			Level:       level,
			NodeType:    "cluster",
			ChunkIDs:    chunkIDs,
			DocumentIDs: docIDs,
			Content:     summary,
			Embedding:   summaryEmbedding,
		}

		// Save to MySQL
		raptorNode := model.RaptorNode{
			KnowledgeBaseID: kbID,
			NodeType:        "cluster",
			DocumentIDs:     toJSONString(docIDs),
			Summary:         summary,
			Level:           level,
		}
		if err := mysql.DB.Create(&raptorNode).Error; err != nil {
			return nil, fmt.Errorf("failed to save raptor node: %w", err)
		}
		parentNode.ID = raptorNode.ID

		// Store in Qdrant
		if s.vectorClient != nil {
			point := &qdrant.VectorPoint{
				Vector:        summaryEmbedding,
				DocumentID:    0, // Summary doesn't belong to a single document
				ChunkID:       raptorNode.ID,
				ChunkIndex:    0,
				Content:       summary,
				DocumentTitle: fmt.Sprintf("RAPTOR Summary L%d", level),
				KBID:          kbID,
				RaptorLevel:   level,
			}
			pointID, err := s.vectorClient.UpsertVector(ctx, point)
			if err != nil {
				return nil, fmt.Errorf("failed to store summary vector: %w", err)
			}

			// Update Qdrant point ID
			mysql.DB.Model(&model.RaptorNode{}).Where("id = ?", raptorNode.ID).
				Update("qdrant_point_id", pointID)
			parentNode.PointID = pointID
		}

		parentNodes = append(parentNodes, parentNode)
	}

	// If only one cluster at max level, mark as root
	if level >= config.MaxLevel && len(parentNodes) == 1 {
		parentNodes[0].NodeType = "root"
		mysql.DB.Model(&model.RaptorNode{}).Where("id = ?", parentNodes[0].ID).
			Update("node_type", "root")
	}

	return parentNodes, nil
}

// getNodeEmbeddings gets embeddings for nodes
func (s *RaptorService) getNodeEmbeddings(ctx context.Context, nodes []*RaptorNode) ([][]float32, error) {
	texts := make([]string, len(nodes))
	for i, node := range nodes {
		texts[i] = node.Content
	}

	// Generate embeddings in batch
	embeddings, err := s.embeddingSvc.BatchGenerateEmbeddings(ctx, texts, 100)
	if err != nil {
		return nil, err
	}

	return embeddings, nil
}

// clusterNodes clusters nodes using k-means style clustering
func (s *RaptorService) clusterNodes(nodes []*RaptorNode, embeddings [][]float32, config RaptorConfig) []Cluster {
	n := len(nodes)
	if n == 0 {
		return nil
	}

	// Determine number of clusters
	maxClusters := (n + config.MaxClusterSize - 1) / config.MaxClusterSize
	if maxClusters < 1 {
		maxClusters = 1
	}

	// Use hierarchical agglomerative clustering
	clusters := s.hierarchicalClustering(n, embeddings, maxClusters, config.ClusterThreshold)

	return clusters
}

// hierarchicalClustering performs hierarchical agglomerative clustering
func (s *RaptorService) hierarchicalClustering(n int, embeddings [][]float32, maxClusters int, threshold float64) []Cluster {
	if n <= maxClusters {
		// Each item is its own cluster
		clusters := make([]Cluster, n)
		for i := 0; i < n; i++ {
			clusters[i] = Cluster{
				ID:       i,
				Members:  []int{i},
				Centroid: embeddings[i],
			}
		}
		return clusters
	}

	// Initialize: each item is a cluster
	clusters := make([]Cluster, n)
	for i := 0; i < n; i++ {
		clusters[i] = Cluster{
			ID:       i,
			Members:  []int{i},
			Centroid: embeddings[i],
		}
	}

	// Compute distance matrix
	distances := make([][]float64, n)
	for i := range distances {
		distances[i] = make([]float64, n)
		for j := range distances[i] {
			if i == j {
				distances[i][j] = 0
			} else {
				distances[i][j] = 1 - cosineSimilarity(embeddings[i], embeddings[j])
			}
		}
	}

	// Agglomerative clustering
	for len(clusters) > maxClusters {
		// Find closest pair
		minDist := math.Inf(1)
		minI, minJ := -1, -1

		for i := 0; i < len(clusters); i++ {
			for j := i + 1; j < len(clusters); j++ {
				// Average linkage
				avgDist := averageDistance(clusters[i].Members, clusters[j].Members, distances)
				if avgDist < minDist {
					minDist = avgDist
					minI, minJ = i, j
				}
			}
		}

		if minDist > 1-threshold && len(clusters) <= maxClusters*2 {
			break
		}

		// Merge clusters
		if minI >= 0 && minJ >= 0 {
			clusters[minI].Members = append(clusters[minI].Members, clusters[minJ].Members...)
			clusters[minI].Centroid = computeCentroid(clusters[minI].Members, embeddings)

			// Remove cluster j
			clusters = append(clusters[:minJ], clusters[minJ+1:]...)
		} else {
			break
		}
	}

	return clusters
}

// averageDistance computes average distance between two clusters
func averageDistance(cluster1, cluster2 []int, distances [][]float64) float64 {
	sum := 0.0
	count := 0
	for _, i := range cluster1 {
		for _, j := range cluster2 {
			sum += distances[i][j]
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

// computeCentroid computes the centroid of a set of embeddings
func computeCentroid(indices []int, embeddings [][]float32) []float32 {
	if len(indices) == 0 {
		return nil
	}

	dim := len(embeddings[0])
	centroid := make([]float32, dim)

	for _, idx := range indices {
		for d := 0; d < dim; d++ {
			centroid[d] += embeddings[idx][d]
		}
	}

	for d := 0; d < dim; d++ {
		centroid[d] /= float32(len(indices))
	}

	return centroid
}

// cosineSimilarity computes cosine similarity between two vectors
func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// generateSummary generates a summary using AI
func (s *RaptorService) generateSummary(ctx context.Context, contents []string, maxTokens int) (string, error) {
	// Combine contents with truncation
	combined := strings.Join(contents, "\n\n---\n\n")
	if len(combined) > 8000 {
		combined = combined[:8000]
	}

	prompt := fmt.Sprintf(`请对以下文档片段进行总结，提取关键信息，生成一个简洁的摘要（不超过%d字）：

%s

请直接输出摘要内容：`, maxTokens, combined)

	// Use the AI chat function
	answer, err := callAIChat(prompt, "", nil)
	if err != nil {
		return "", err
	}

	return answer, nil
}

// HierarchicalSearch performs hierarchical search across RAPTOR levels
func (s *RaptorService) HierarchicalSearch(ctx context.Context, kbID uint, query string, topK int) ([]SearchResult, error) {
	// Generate query embedding
	queryEmbedding, err := s.embeddingSvc.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Search at different levels
	var allResults []SearchResult

	// Level 0: Original chunks
	chunkResults, err := s.vectorClient.Search(ctx, kbID, queryEmbedding, topK/2, intPtr(0))
	if err == nil {
		for _, r := range chunkResults {
			allResults = append(allResults, SearchResult{
				ChunkID:       r.ChunkID,
				DocumentID:    r.DocumentID,
				DocumentTitle: r.DocumentTitle,
				Content:       r.Content,
				Score:         r.Score,
				Source:        "raptor_level_0",
			})
		}
	}

	// Level 1+: RAPTOR summaries
	for level := 1; level <= 3; level++ {
		summaryResults, err := s.vectorClient.Search(ctx, kbID, queryEmbedding, topK/4, intPtr(level))
		if err == nil {
			for _, r := range summaryResults {
				allResults = append(allResults, SearchResult{
					ChunkID:       r.ChunkID,
					DocumentID:    r.DocumentID,
					DocumentTitle: r.DocumentTitle,
					Content:       r.Content,
					Score:         r.Score * 1.1, // Boost summaries slightly
					Source:        fmt.Sprintf("raptor_level_%d", level),
				})
			}
		}
	}

	// Sort by score and return top K
	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].Score > allResults[j].Score
	})

	if len(allResults) > topK {
		allResults = allResults[:topK]
	}

	return allResults, nil
}

// GetRaptorTreeStats returns statistics about the RAPTOR tree
func (s *RaptorService) GetRaptorTreeStats(kbID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Count nodes by level
	var nodes []model.RaptorNode
	mysql.DB.Where("knowledge_base_id = ?", kbID).Find(&nodes)

	levelCounts := make(map[int]int)
	for _, node := range nodes {
		levelCounts[node.Level]++
	}

	stats["total_nodes"] = len(nodes)
	stats["level_counts"] = levelCounts
	stats["has_tree"] = len(nodes) > 0

	return stats, nil
}

// Helper functions

func toJSONString(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func intPtr(i int) *int {
	return &i
}

// Global RAPTOR service instance
var globalRaptorService *RaptorService

// InitRaptorService initializes the global RAPTOR service
func InitRaptorService(embeddingSvc *EmbeddingService, vectorClient *qdrant.VectorClient) {
	globalRaptorService = NewRaptorService(embeddingSvc, vectorClient)
}

// GetRaptorService returns the global RAPTOR service
func GetRaptorService() *RaptorService {
	return globalRaptorService
}

// BuildKBTree builds RAPTOR tree for a knowledge base
func BuildKBTree(kbID uint) error {
	if globalRaptorService == nil {
		return fmt.Errorf("RAPTOR service not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	return globalRaptorService.BuildRaptorTree(ctx, kbID, DefaultRaptorConfig())
}

// KMeansClustering provides k-means clustering as an alternative
func KMeansClustering(embeddings [][]float32, k int, maxIterations int) [][]int {
	n := len(embeddings)
	if n == 0 || k <= 0 {
		return nil
	}

	if k > n {
		k = n
	}

	// Initialize centroids randomly
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	centroidIndices := r.Perm(n)[:k]
	centroids := make([][]float32, k)
	for i, idx := range centroidIndices {
		centroids[i] = make([]float32, len(embeddings[0]))
		copy(centroids[i], embeddings[idx])
	}

	// Assignments
	assignments := make([]int, n)

	for iter := 0; iter < maxIterations; iter++ {
		// Assign points to nearest centroid
		changed := false
		for i, emb := range embeddings {
			minDist := math.Inf(1)
			minCluster := 0
			for j, centroid := range centroids {
				dist := 1 - cosineSimilarity(emb, centroid)
				if dist < minDist {
					minDist = dist
					minCluster = j
				}
			}
			if assignments[i] != minCluster {
				assignments[i] = minCluster
				changed = true
			}
		}

		if !changed {
			break
		}

		// Update centroids
		for j := range centroids {
			var count int
			newCentroid := make([]float32, len(centroids[j]))
			for i, cluster := range assignments {
				if cluster == j {
					count++
					for d := range embeddings[i] {
						newCentroid[d] += embeddings[i][d]
					}
				}
			}
			if count > 0 {
				for d := range newCentroid {
					newCentroid[d] /= float32(count)
				}
				centroids[j] = newCentroid
			}
		}
	}

	// Group by cluster
	clusters := make([][]int, k)
	for i, cluster := range assignments {
		clusters[cluster] = append(clusters[cluster], i)
	}

	return clusters
}
