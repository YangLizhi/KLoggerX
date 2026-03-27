package v1

import (
	"net/http"
	"strconv"

	"kloggerx-server/internal/service"

	"github.com/gin-gonic/gin"
)

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
