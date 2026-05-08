package service

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/pkg/logger"
	"kloggerx-server/internal/repository/mysql"
)

// ─── Knowledge Graph Construction Service ────────────────────────────────────

const maxContentLengthPerDoc = 4000 // 单个文档传给 LLM 的最大字符数

// BuildKnowledgeGraph triggers async knowledge graph construction for the given knowledge base.
func BuildKnowledgeGraph(kbID uint) error {
	// 检查是否已有 building 状态的记录（防重复构建）
	var existing model.KnowledgeGraph
	result := mysql.DB.Where("knowledge_base_id = ?", kbID).First(&existing)
	if result.Error == nil && existing.Status == "building" {
		return fmt.Errorf("知识图谱正在构建中，请稍后再试")
	}

	// 创建或更新 KnowledgeGraph 记录
	if result.Error != nil {
		// 不存在，创建新记录
		existing = model.KnowledgeGraph{
			KnowledgeBaseID: kbID,
			Status:          "building",
		}
		if err := mysql.DB.Create(&existing).Error; err != nil {
			return fmt.Errorf("创建图谱记录失败: %v", err)
		}
	} else {
		// 已存在，更新状态
		mysql.DB.Model(&existing).Updates(map[string]interface{}{
			"status":        "building",
			"error_message": "",
		})
	}

	graphID := existing.ID

	// 异步执行构建
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Errorf("GraphBuild: panic recovered in goroutine for KB %d: %v", kbID, r)
				mysql.DB.Model(&model.KnowledgeGraph{}).Where("id = ?", graphID).Updates(map[string]interface{}{
					"status":        "failed",
					"error_message": fmt.Sprintf("内部错误: %v", r),
				})
			}
		}()

		if err := doBuildKnowledgeGraph(kbID, graphID); err != nil {
			logger.Errorf("GraphBuild: KB %d build failed: %v", kbID, err)
			mysql.DB.Model(&model.KnowledgeGraph{}).Where("id = ?", graphID).Updates(map[string]interface{}{
				"status":        "failed",
				"error_message": err.Error(),
			})
		}
	}()

	return nil
}

// doBuildKnowledgeGraph performs the actual graph building synchronously.
func doBuildKnowledgeGraph(kbID uint, graphID uint) error {
	logger.Infof("GraphBuild: Starting build for KB %d", kbID)

	// 1. 获取所有已嵌入的 chunks
	var chunks []model.KnowledgeChunk
	if err := mysql.DB.Where("knowledge_base_id = ? AND embedding_status = ?", kbID, "embedded").
		Order("document_id, chunk_index").Find(&chunks).Error; err != nil {
		return fmt.Errorf("查询知识块失败: %v", err)
	}

	if len(chunks) == 0 {
		return fmt.Errorf("没有可用的知识块，请先同步数据来源并等待向量化完成")
	}

	logger.Infof("GraphBuild: Found %d embedded chunks for KB %d", len(chunks), kbID)

	// 2. 按 document_id 分组 chunks
	docChunks := make(map[uint][]model.KnowledgeChunk)
	for _, chunk := range chunks {
		docChunks[chunk.DocumentID] = append(docChunks[chunk.DocumentID], chunk)
	}

	// 3. 对每个文档提取实体和关系
	var allEntities []map[string]interface{}
	var allRelationships []map[string]interface{}

	for docID, docChunkList := range docChunks {
		// 获取文档标题
		docTitle := docChunkList[0].DocumentTitle
		if docTitle == "" {
			var doc model.Document
			if err := mysql.DB.Select("title").First(&doc, docID).Error; err == nil {
				docTitle = doc.Title
			} else {
				docTitle = fmt.Sprintf("文档%d", docID)
			}
		}

		// 合并 chunks 的内容
		var contentParts []string
		totalLen := 0
		for _, chunk := range docChunkList {
			if totalLen+len([]rune(chunk.Content)) > maxContentLengthPerDoc {
				break
			}
			contentParts = append(contentParts, chunk.Content)
			totalLen += len([]rune(chunk.Content))
		}
		content := strings.Join(contentParts, "\n")

		if content == "" {
			continue
		}

		logger.Debugf("GraphBuild: Extracting entities from doc %d (%s), content length=%d", docID, docTitle, len([]rune(content)))

		entities, relationships, err := ExtractEntitiesFromDocument(docTitle, content)
		if err != nil {
			logger.Warnf("GraphBuild: Failed to extract entities from doc %d: %v", docID, err)
			continue // 单个文档失败不影响整体
		}

		// 给实体 ID 加上文档前缀，避免跨文档 ID 冲突
		prefix := fmt.Sprintf("doc%d_", docID)
		for i := range entities {
			if id, ok := entities[i]["id"].(string); ok {
				entities[i]["id"] = prefix + id
			}
			entities[i]["source_doc_id"] = fmt.Sprintf("%d", docID)
		}
		for i := range relationships {
			if src, ok := relationships[i]["source"].(string); ok {
				relationships[i]["source"] = prefix + src
			}
			if tgt, ok := relationships[i]["target"].(string); ok {
				relationships[i]["target"] = prefix + tgt
			}
		}

		allEntities = append(allEntities, entities...)
		allRelationships = append(allRelationships, relationships...)

		logger.Debugf("GraphBuild: Doc %d: extracted %d entities, %d relationships", docID, len(entities), len(relationships))
	}

	if len(allEntities) == 0 {
		return fmt.Errorf("未能从文档中提取到任何实体")
	}

	logger.Infof("GraphBuild: Total extracted: %d entities, %d relationships", len(allEntities), len(allRelationships))

	// 4. 写入临时 entities.json
	entitiesData := map[string]interface{}{
		"entities":      allEntities,
		"relationships": allRelationships,
	}

	tmpDir := os.TempDir()
	inputPath := filepath.Join(tmpDir, fmt.Sprintf("kg_entities_%d_%d.json", kbID, time.Now().UnixNano()))
	outputPath := filepath.Join(tmpDir, fmt.Sprintf("kg_graph_%d_%d.json", kbID, time.Now().UnixNano()))

	// 清理临时文件
	defer os.Remove(inputPath)
	defer os.Remove(outputPath)

	inputJSON, err := json.Marshal(entitiesData)
	if err != nil {
		return fmt.Errorf("序列化实体数据失败: %v", err)
	}

	if err := os.WriteFile(inputPath, inputJSON, 0644); err != nil {
		return fmt.Errorf("写入临时文件失败: %v", err)
	}

	// 5. 调用 Python 脚本
	scriptPath, err := findBuildGraphScript()
	if err != nil {
		return fmt.Errorf("查找图构建脚本失败: %v", err)
	}

	logger.Debugf("GraphBuild: Running: python3 %s --input %s --output %s", scriptPath, inputPath, outputPath)

	cmd := exec.Command("python3", scriptPath, "--input", inputPath, "--output", outputPath)
	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Python脚本执行失败: %v, 输出: %s", err, string(cmdOutput))
	}

	logger.Debugf("GraphBuild: Python script output: %s", string(cmdOutput))

	// 6. 读取 graph.json
	graphJSON, err := os.ReadFile(outputPath)
	if err != nil {
		return fmt.Errorf("读取图谱结果失败: %v", err)
	}

	// 7. 解析 metadata
	var graphData map[string]interface{}
	if err := json.Unmarshal(graphJSON, &graphData); err != nil {
		return fmt.Errorf("解析图谱数据失败: %v", err)
	}

	nodeCount := 0
	edgeCount := 0
	communityCount := 0

	if metadata, ok := graphData["metadata"].(map[string]interface{}); ok {
		if v, ok := metadata["node_count"].(float64); ok {
			nodeCount = int(v)
		}
		if v, ok := metadata["edge_count"].(float64); ok {
			edgeCount = int(v)
		}
		if v, ok := metadata["community_count"].(float64); ok {
			communityCount = int(v)
		}
	}

	// 8. 更新数据库记录
	if err := mysql.DB.Model(&model.KnowledgeGraph{}).Where("id = ?", graphID).Updates(map[string]interface{}{
		"graph_data":      string(graphJSON),
		"node_count":      nodeCount,
		"edge_count":      edgeCount,
		"community_count": communityCount,
		"status":          "ready",
		"error_message":   "",
	}).Error; err != nil {
		return fmt.Errorf("更新图谱数据失败: %v", err)
	}

	fmt.Printf("[GraphBuild] KB %d build completed: %d nodes, %d edges, %d communities\n",
		kbID, nodeCount, edgeCount, communityCount)

	return nil
}

// findBuildGraphScript locates the build_graph.py script.
func findBuildGraphScript() (string, error) {
	// 优先从工作目录查找
	wd, err := os.Getwd()
	if err == nil {
		scriptPath := filepath.Join(wd, "scripts", "build_graph.py")
		if _, err := os.Stat(scriptPath); err == nil {
			return scriptPath, nil
		}
	}

	// 尝试从可执行文件所在目录查找
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		scriptPath := filepath.Join(execDir, "scripts", "build_graph.py")
		if _, err := os.Stat(scriptPath); err == nil {
			return scriptPath, nil
		}
		// 向上一层
		scriptPath = filepath.Join(filepath.Dir(execDir), "scripts", "build_graph.py")
		if _, err := os.Stat(scriptPath); err == nil {
			return scriptPath, nil
		}
	}

	return "", fmt.Errorf("未找到 build_graph.py 脚本，请确认 scripts/build_graph.py 存在于工作目录中")
}

// ─── Entity Extraction ───────────────────────────────────────────────────────

// ExtractEntitiesFromDocument calls LLM to extract entities and relationships from document content.
func ExtractEntitiesFromDocument(docTitle string, content string) ([]map[string]interface{}, []map[string]interface{}, error) {
	systemPrompt := "你是一个知识图谱构建专家。请从以下文档内容中提取关键实体和它们之间的关系。"

	userPrompt := fmt.Sprintf(`文档标题：%s
文档内容：
%s

请严格按照以下 JSON 格式返回，不要包含其他文字：
{
  "entities": [
    {"id": "唯一标识（使用简短英文，如 transformer_model）", "label": "实体名称", "type": "concept|person|technology|event|organization", "description": "一句话描述"}
  ],
  "relationships": [
    {"source": "实体id", "target": "实体id", "type": "关系类型", "description": "关系描述", "confidence": 0.9}
  ]
}

实体类型说明：
- concept: 概念、理论、方法论
- person: 人物
- technology: 技术、工具、框架、产品
- event: 事件、会议、里程碑
- organization: 组织、公司、团队

关系类型包括：uses, depends_on, related_to, part_of, created_by, implements, extends, describes, similar_to, influences

注意：
1. 每个实体的 id 必须唯一且简短
2. 关系的 source 和 target 必须是已定义的实体 id
3. confidence 取值 0.6-1.0
4. 尽量提取 5-20 个实体和相应的关系`, docTitle, content)

	// 复用现有 AI 调用方式
	response, err := callAIChat(systemPrompt, userPrompt, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("调用AI失败: %v", err)
	}

	// 提取 JSON 内容（处理 markdown 代码块等情况）
	jsonStr := extractJSONFromResponse(response)
	if jsonStr == "" {
		return nil, nil, fmt.Errorf("AI返回内容中未找到有效JSON")
	}

	// 解析 JSON
	var result struct {
		Entities      []map[string]interface{} `json:"entities"`
		Relationships []map[string]interface{} `json:"relationships"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, nil, fmt.Errorf("解析AI返回的JSON失败: %v, 内容: %s", err, truncateStr(jsonStr, 200))
	}

	return result.Entities, result.Relationships, nil
}

// ─── Graph Data Query ────────────────────────────────────────────────────────

// GetKnowledgeGraphData returns the full graph data for visualization.
func GetKnowledgeGraphData(kbID uint) (interface{}, error) {
	var graph model.KnowledgeGraph
	if err := mysql.DB.Where("knowledge_base_id = ?", kbID).First(&graph).Error; err != nil {
		return nil, nil // 不存在时返回空
	}

	if graph.GraphData == "" {
		return nil, nil
	}

	// 将 GraphData JSON 字符串解析为 map
	var graphData map[string]interface{}
	if err := json.Unmarshal([]byte(graph.GraphData), &graphData); err != nil {
		return nil, fmt.Errorf("解析图谱数据失败: %v", err)
	}

	return graphData, nil
}

// GetKnowledgeGraphStatus returns the build status without the large GraphData field.
func GetKnowledgeGraphStatus(kbID uint) (interface{}, error) {
	var graph model.KnowledgeGraph
	if err := mysql.DB.Where("knowledge_base_id = ?", kbID).First(&graph).Error; err != nil {
		// 不存在时返回默认状态
		return map[string]interface{}{
			"status":         "idle",
			"nodeCount":      0,
			"edgeCount":      0,
			"communityCount": 0,
			"errorMessage":   "",
			"updatedAt":      nil,
		}, nil
	}

	return map[string]interface{}{
		"status":         graph.Status,
		"nodeCount":      graph.NodeCount,
		"edgeCount":      graph.EdgeCount,
		"communityCount": graph.CommunityCount,
		"errorMessage":   graph.ErrorMessage,
		"updatedAt":      graph.UpdatedAt,
	}, nil
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// extractJSONFromResponse extracts JSON content from LLM response that may be
// wrapped in markdown code blocks like ```json ... ```.
func extractJSONFromResponse(response string) string {
	// 尝试提取 markdown 代码块中的内容
	if idx := strings.Index(response, "```json"); idx != -1 {
		start := idx + len("```json")
		if end := strings.Index(response[start:], "```"); end != -1 {
			return strings.TrimSpace(response[start : start+end])
		}
	}

	if idx := strings.Index(response, "```"); idx != -1 {
		start := idx + len("```")
		// 跳过可能的语言标识行
		if nlIdx := strings.Index(response[start:], "\n"); nlIdx != -1 {
			afterFirstLine := start + nlIdx + 1
			if end := strings.Index(response[afterFirstLine:], "```"); end != -1 {
				candidate := strings.TrimSpace(response[afterFirstLine : afterFirstLine+end])
				if len(candidate) > 0 && candidate[0] == '{' {
					return candidate
				}
			}
		}
	}

	// 查找第一个 { 和最后一个 } 之间的内容
	firstBrace := strings.Index(response, "{")
	lastBrace := strings.LastIndex(response, "}")
	if firstBrace != -1 && lastBrace > firstBrace {
		return response[firstBrace : lastBrace+1]
	}

	return ""
}

// truncateStr truncates a string to the given max length.
func truncateStr(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
