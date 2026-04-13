package v1

import (
	"net/http"
	"strconv"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/service"

	"github.com/gin-gonic/gin"
)

func GetTemplates(c *gin.Context) {
	category := c.Query("category")
	list, err := service.GetTemplates(category)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	cats, _ := service.GetTemplateCategories()
	c.JSON(http.StatusOK, model.Success(gin.H{"list": list, "categories": cats}))
}

func GetTemplateDetail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	t, err := service.GetTemplateDetail(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("模板不存在"))
		return
	}
	c.JSON(http.StatusOK, model.Success(t))
}

// UseTemplate creates a new document pre-filled with a template's content.
func UseTemplate(c *gin.Context) {
	uid := extractUserID(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		ParentID *uint  `json:"parentId"`
		Title    string `json:"title"`
	}
	c.ShouldBindJSON(&body)

	t, err := service.GetTemplateDetail(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("模板不存在"))
		return
	}
	title := t.Name
	if body.Title != "" {
		title = body.Title
	}
	doc, err := service.CreateDocument(uid, service.CreateDocReq{
		Title:    title,
		Type:     t.Type,
		ParentID: body.ParentID,
	})
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	if t.Content != "" && t.Content != "{}" {
		service.SaveDocumentContent(doc.ID, uid, t.Content)
	}
	service.CreateOperationLog(uid, "", "create", "document", doc.ID, doc.Title, "from_template:"+t.Name, c.ClientIP())
	c.JSON(http.StatusOK, model.Success(doc))
}

func extractUserID(c *gin.Context) uint {
	v, _ := c.Get("userId")
	switch id := v.(type) {
	case uint:
		return id
	case float64:
		return uint(id)
	}
	return 0
}

// Admin handlers for template management

// CreateTemplateHandler creates a new template (admin only)
func CreateTemplateHandler(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Category    string `json:"category"`
		Type        string `json:"type" binding:"required"`
		Content     string `json:"content"`
		IsBuiltin   bool   `json:"isBuiltin"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("请填写必填字段: "+err.Error()))
		return
	}

	t := &model.Template{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Type:        req.Type,
		Content:     req.Content,
		IsBuiltin:   req.IsBuiltin,
	}
	if err := service.CreateTemplate(t); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("创建模板失败: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(t))
}

// UpdateTemplateHandler updates an existing template (admin only)
func UpdateTemplateHandler(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("无效的模板ID"))
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("请求参数错误"))
		return
	}

	if err := service.UpdateTemplate(uint(id), req); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("更新模板失败: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(gin.H{"id": id}))
}

// DeleteTemplateHandler deletes a template (admin only)
func DeleteTemplateHandler(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("无效的模板ID"))
		return
	}

	if err := service.DeleteTemplate(uint(id)); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("删除模板失败: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(gin.H{"id": id}))
}

// GetAllTemplatesHandler returns all templates including non-builtin (admin only)
func GetAllTemplatesHandler(c *gin.Context) {
	list, err := service.GetAllTemplates()
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	cats, _ := service.GetTemplateCategories()
	c.JSON(http.StatusOK, model.Success(gin.H{"list": list, "categories": cats}))
}
