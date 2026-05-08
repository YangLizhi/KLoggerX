package v1

import (
	"net/http"
	"strconv"
	"time"

	"kloggerx-server/internal/service"

	"github.com/gin-gonic/gin"
)

// GetOperationLogs 获取文档操作日志（原有接口）
// GetOperationLogs godoc
// @Summary 获取操作日志
// @Description 获取文档的操作日志列表
// @Tags 操作日志
// @Produce json
// @Param id path int true "文档ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /document/{id}/logs [get]
func GetOperationLogs(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	logs, total, err := service.GetOperationLogs("document", uint(id), page, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": "获取操作日志失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"list":     logs,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

// ListAllOperationLogs 查询所有操作日志（管理员接口，支持过滤）
func ListAllOperationLogs(c *gin.Context) {
	filter := service.OperationLogFilter{}

	if uid := c.Query("user_id"); uid != "" {
		id, _ := strconv.ParseUint(uid, 10, 64)
		filter.UserID = uint(id)
	}

	filter.Action = c.Query("action")

	if st := c.Query("start_time"); st != "" {
		t, err := time.Parse("2006-01-02", st)
		if err == nil {
			filter.StartTime = &t
		}
	}

	if et := c.Query("end_time"); et != "" {
		t, err := time.Parse("2006-01-02", et)
		if err == nil {
			// 包含结束日期当天
			end := t.Add(24*time.Hour - time.Second)
			filter.EndTime = &end
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	filter.Page = page
	filter.PageSize = pageSize

	logs, total, err := service.ListOperationLogs(filter)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": "获取操作日志失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"list":     logs,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}
