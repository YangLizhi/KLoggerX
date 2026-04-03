package v1

import (
	"net/http"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/pkg/utils"
	"kloggerx-server/internal/service"

	"github.com/gin-gonic/gin"
)

// GetStorageStats returns system storage statistics
func GetStorageStats(c *gin.Context) {
	stats, err := service.GetSystemStorageStats()
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("获取存储统计失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.Success(stats))
}

// GetStorageUsage returns storage usage breakdown
func GetStorageUsage(c *gin.Context) {
	userID := utils.GetUserID(c.MustGet("userId"))

	usage, totalSize, fileCount, err := service.GetUserStorageUsage(userID)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("获取存储使用情况失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.Success(map[string]interface{}{
		"byType":    usage,
		"totalSize": totalSize,
		"fileCount": fileCount,
	}))
}

// GetAdminStorageUsage returns storage usage for all users (admin only)
func GetAdminStorageUsage(c *gin.Context) {
	usage, totalSize, fileCount, err := service.GetAllStorageUsage()
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("获取存储使用情况失败: "+err.Error()))
		return
	}

	// Also get disk stats
	stats, err := service.GetSystemStorageStats()
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("获取磁盘统计失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.Success(map[string]interface{}{
		"diskStats": stats,
		"byType":    usage,
		"totalSize": totalSize,
		"fileCount": fileCount,
	}))
}
