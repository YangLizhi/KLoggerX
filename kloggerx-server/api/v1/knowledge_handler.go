package v1

import (
	"net/http"
	"strconv"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/pkg/utils"
	"kloggerx-server/internal/service"

	"github.com/gin-gonic/gin"
)

func GetKnowledgeBaseList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	keyword := c.Query("keyword")
	list, total, err := service.GetKnowledgeBaseList(page, pageSize, keyword)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(model.PaginatedData{List: list, Total: total, Page: page, PageSize: pageSize}))
}

func GetKnowledgeBaseDetail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	kb, err := service.GetKnowledgeBaseDetail(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("知识库不存在"))
		return
	}
	c.JSON(http.StatusOK, model.Success(kb))
}

func CreateKnowledgeBase(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	var body struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}
	kb, err := service.CreateKnowledgeBase(uid, body.Name, body.Description)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(kb))
}

func UpdateKnowledgeBase(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	c.ShouldBindJSON(&body)
	updates := map[string]interface{}{}
	if body.Name != "" {
		updates["name"] = body.Name
	}
	if body.Description != "" {
		updates["description"] = body.Description
	}
	if len(updates) > 0 {
		service.UpdateKnowledgeBaseInfo(uint(id), updates)
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

func DeleteKnowledgeBase(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeleteKnowledgeBase(uint(id)); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

func GetKnowledgeBaseTree(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	docs, err := service.GetKnowledgeBaseTree(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(docs))
}

func AddKnowledgeMember(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		UserID uint   `json:"userId" binding:"required"`
		Role   string `json:"role"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}
	if body.Role == "" {
		body.Role = "member"
	}
	if err := service.AddKnowledgeMember(uint(id), body.UserID, body.Role); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

func RemoveKnowledgeMember(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		UserID uint `json:"userId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}
	if err := service.RemoveKnowledgeMember(uint(id), body.UserID); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

func GetKnowledgeMembers(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	members, err := service.GetKnowledgeMembers(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(members))
}

func SearchKnowledge(c *gin.Context) {
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	kbIDStr := c.Query("knowledgeBaseId")
	var kbID *uint
	if kbIDStr != "" {
		id, _ := strconv.ParseUint(kbIDStr, 10, 64)
		p := uint(id)
		kbID = &p
	}
	docs, total, err := service.SearchKnowledge(keyword, kbID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(model.PaginatedData{List: docs, Total: total, Page: page, PageSize: pageSize}))
}

func PublishDocument(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		DocumentID uint `json:"documentId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}
	if err := service.PublishDocument(uint(id), body.DocumentID); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}
