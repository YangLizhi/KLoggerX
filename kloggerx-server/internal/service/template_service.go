package service

import (
	"encoding/json"
	"regexp"
	"strings"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/repository/mysql"
)

func GetTemplates(category string) ([]model.Template, error) {
	var list []model.Template
	db := mysql.DB.Model(&model.Template{})
	if category != "" {
		db = db.Where("category = ?", category)
	}
	err := db.Order("id ASC").Find(&list).Error
	return list, err
}

// GetTemplatesByCategory returns templates by category with pagination.
func GetTemplatesByCategory(category string, page, pageSize int) ([]model.Template, int64, error) {
	var list []model.Template
	var total int64
	db := mysql.DB.Model(&model.Template{})
	if category != "" {
		db = db.Where("category = ?", category)
	}
	db.Count(&total)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	err := db.Offset((page - 1) * pageSize).Limit(pageSize).Order("id ASC").Find(&list).Error
	return list, total, err
}

// SearchTemplates searches templates by keyword (name fuzzy match).
func SearchTemplates(keyword string) ([]model.Template, error) {
	var list []model.Template
	err := mysql.DB.Where("name LIKE ?", "%"+keyword+"%").Order("id ASC").Find(&list).Error
	return list, err
}

// FavoriteTemplate adds a template to user's favorites.
func FavoriteTemplate(userID, templateID uint) error {
	fav := model.TemplateFavorite{
		UserID:     userID,
		TemplateID: templateID,
	}
	return mysql.DB.Where("user_id = ? AND template_id = ?", userID, templateID).FirstOrCreate(&fav).Error
}

// UnfavoriteTemplate removes a template from user's favorites.
func UnfavoriteTemplate(userID, templateID uint) error {
	return mysql.DB.Where("user_id = ? AND template_id = ?", userID, templateID).Delete(&model.TemplateFavorite{}).Error
}

// GetUserFavoriteTemplateIDs returns the template IDs a user has favorited.
func GetUserFavoriteTemplateIDs(userID uint) ([]uint, error) {
	var ids []uint
	err := mysql.DB.Model(&model.TemplateFavorite{}).Where("user_id = ?", userID).Pluck("template_id", &ids).Error
	return ids, err
}

func GetTemplateCategories() ([]string, error) {
	var cats []string
	err := mysql.DB.Model(&model.Template{}).Distinct("category").Pluck("category", &cats).Error
	return cats, err
}

func GetTemplateDetail(id uint) (*model.Template, error) {
	var t model.Template
	err := mysql.DB.First(&t, id).Error
	return &t, err
}

// CreateTemplate creates a new template
func CreateTemplate(t *model.Template) error {
	// Generate preview from content
	GenerateTemplatePreview(t)
	return mysql.DB.Create(t).Error
}

// UpdateTemplate updates a template by ID
func UpdateTemplate(id uint, updates map[string]interface{}) error {
	// If content is being updated, regenerate preview
	if content, ok := updates["content"].(string); ok {
		updates["preview"] = generatePreview(content)
	}
	return mysql.DB.Model(&model.Template{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteTemplate deletes a template by ID
func DeleteTemplate(id uint) error {
	return mysql.DB.Delete(&model.Template{}, id).Error
}

// GetAllTemplates returns all templates including non-builtin ones (for admin)
func GetAllTemplates() ([]model.Template, error) {
	var list []model.Template
	err := mysql.DB.Order("id ASC").Find(&list).Error
	return list, err
}

// extractTextFromContent extracts plain text from various content formats (Tiptap JSON, HTML, plain text)
func extractTextFromContent(content string) string {
	if content == "" || content == "{}" {
		return ""
	}

	// Try to parse as Tiptap JSON format
	var doc map[string]interface{}
	if err := json.Unmarshal([]byte(content), &doc); err == nil {
		// Successfully parsed as JSON, extract text recursively
		text := extractTextFromNode(doc)
		if text != "" {
			return strings.TrimSpace(text)
		}
	}

	// If not valid JSON or no text extracted, treat as HTML or plain text
	// Remove HTML tags
	text := removeHTMLTags(content)
	return strings.TrimSpace(text)
}

// extractTextFromNode recursively extracts text from Tiptap JSON nodes
func extractTextFromNode(node interface{}) string {
	var result strings.Builder

	switch n := node.(type) {
	case map[string]interface{}:
		// Check if this node has text
		if text, ok := n["text"].(string); ok {
			result.WriteString(text)
		}

		// Process content array if exists
		if content, ok := n["content"].([]interface{}); ok {
			for _, child := range content {
				childText := extractTextFromNode(child)
				if childText != "" {
					if result.Len() > 0 {
						result.WriteString(" ")
					}
					result.WriteString(childText)
				}
			}
		}

		// Process attrs.text for some node types
		if attrs, ok := n["attrs"].(map[string]interface{}); ok {
			if text, ok := attrs["text"].(string); ok {
				if result.Len() > 0 {
					result.WriteString(" ")
				}
				result.WriteString(text)
			}
		}
	case []interface{}:
		for _, item := range n {
			itemText := extractTextFromNode(item)
			if itemText != "" {
				if result.Len() > 0 {
					result.WriteString(" ")
				}
				result.WriteString(itemText)
			}
		}
	}

	return result.String()
}

// removeHTMLTags removes HTML tags from content
func removeHTMLTags(content string) string {
	// Remove script and style tags with their content
	scriptRegex := regexp.MustCompile(`(?i)<(script|style)[^>]*>[\s\S]*?</\1>`)
	content = scriptRegex.ReplaceAllString(content, " ")

	// Remove HTML tags
	tagRegex := regexp.MustCompile(`<[^>]+>`)
	content = tagRegex.ReplaceAllString(content, " ")

	// Decode common HTML entities
	content = strings.ReplaceAll(content, "&nbsp;", " ")
	content = strings.ReplaceAll(content, "&lt;", "<")
	content = strings.ReplaceAll(content, "&gt;", ">")
	content = strings.ReplaceAll(content, "&amp;", "&")
	content = strings.ReplaceAll(content, "&quot;", "\"")
	content = strings.ReplaceAll(content, "&#39;", "'")

	// Normalize whitespace
	spaceRegex := regexp.MustCompile(`\s+`)
	content = spaceRegex.ReplaceAllString(content, " ")

	return strings.TrimSpace(content)
}

// generatePreview generates a preview text from content (first 200 chars)
func generatePreview(content string) string {
	text := extractTextFromContent(content)
	if text == "" {
		return ""
	}

	// Take first 200 characters
	maxLen := 200
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}

// GenerateTemplatePreview generates preview for a template and updates it
func GenerateTemplatePreview(t *model.Template) {
	if t.Content != "" && t.Content != "{}" {
		t.Preview = generatePreview(t.Content)
	}
}
