package v1

import (
	"net/http"
	"strconv"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/service"

	"github.com/gin-gonic/gin"
)

// GetTemplates godoc
// @Summary 获取模板列表
// @Description 获取可用模板列表，支持分类筛选和关键词搜索
// @Tags 模板
// @Produce json
// @Param category query string false "分类筛选"
// @Param keyword query string false "搜索关键词"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /template/list [get]
func GetTemplates(c *gin.Context) {
	category := c.Query("category")
	keyword := c.Query("keyword")

	var list []model.Template
	var err error

	if keyword != "" {
		list, err = service.SearchTemplates(keyword)
	} else {
		list, err = service.GetTemplates(category)
	}
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}

	// Attach favorite info if user is authenticated
	var favoriteIDs []uint
	uid := extractUserID(c)
	if uid > 0 {
		favoriteIDs, _ = service.GetUserFavoriteTemplateIDs(uid)
	}

	cats, _ := service.GetTemplateCategories()
	c.JSON(http.StatusOK, model.Success(gin.H{"list": list, "categories": cats, "favoriteIds": favoriteIDs}))
}

// GetTemplateDetail godoc
// @Summary 获取模板详情
// @Description 获取指定模板的详细信息
// @Tags 模板
// @Produce json
// @Param id path int true "模板ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /template/{id} [get]
func GetTemplateDetail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	t, err := service.GetTemplateDetail(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("模板不存在"))
		return
	}
	c.JSON(http.StatusOK, model.Success(t))
}

// UseTemplate godoc
// @Summary 使用模板
// @Description 基于模板创建新文档
// @Tags 模板
// @Accept json
// @Produce json
// @Param id path int true "模板ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /template/{id}/use [post]
func UseTemplate(c *gin.Context) {
	uid := extractUserID(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		ParentID *uint  `json:"parentId"`
		Title    string `json:"title" binding:"max=500"`
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

// FavoriteTemplateHandler godoc
// @Summary 收藏模板
// @Description 将模板添加到用户收藏
// @Tags 模板
// @Produce json
// @Param id path int true "模板ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /template/{id}/favorite [post]
func FavoriteTemplateHandler(c *gin.Context) {
	uid := extractUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("无效的模板ID"))
		return
	}
	if err := service.FavoriteTemplate(uid, uint(id)); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("收藏失败: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

// UnfavoriteTemplateHandler godoc
// @Summary 取消收藏模板
// @Description 将模板从用户收藏中移除
// @Tags 模板
// @Produce json
// @Param id path int true "模板ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /template/{id}/favorite [delete]
func UnfavoriteTemplateHandler(c *gin.Context) {
	uid := extractUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("无效的模板ID"))
		return
	}
	if err := service.UnfavoriteTemplate(uid, uint(id)); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("取消收藏失败: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

// Admin handlers for template management

// CreateTemplateHandler godoc
// @Summary 创建模板
// @Description 管理员创建新模板
// @Tags 系统管理
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /admin/template/create [post]
func CreateTemplateHandler(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required,max=200"`
		Description string `json:"description" binding:"max=1000"`
		Category    string `json:"category" binding:"max=100"`
		Type        string `json:"type" binding:"required,oneof=doc sheet slide mind code survey bitable"`
		Content     string `json:"content" binding:"max=5000000"`
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

// UpdateTemplateHandler godoc
// @Summary 更新模板
// @Description 管理员更新模板
// @Tags 系统管理
// @Accept json
// @Produce json
// @Param id path int true "模板ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /admin/template/{id} [put]
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

// DeleteTemplateHandler godoc
// @Summary 删除模板
// @Description 管理员删除模板
// @Tags 系统管理
// @Produce json
// @Param id path int true "模板ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /admin/template/{id} [delete]
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

// GetAllTemplatesHandler godoc
// @Summary 获取所有模板
// @Description 管理员获取所有模板包括非内置模板
// @Tags 系统管理
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /admin/templates [get]
func GetAllTemplatesHandler(c *gin.Context) {
	list, err := service.GetAllTemplates()
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	cats, _ := service.GetTemplateCategories()
	c.JSON(http.StatusOK, model.Success(gin.H{"list": list, "categories": cats}))
}
