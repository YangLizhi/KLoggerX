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
