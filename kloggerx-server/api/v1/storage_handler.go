package v1

import (
	"net/http"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/pkg/utils"
	"kloggerx-server/internal/service"

	"github.com/gin-gonic/gin"
)

// GetStorageStats godoc
// @Summary 获取存储统计
// @Description 获取系统存储统计信息
// @Tags 文件
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /storage/stats [get]
func GetStorageStats(c *gin.Context) {
	stats, err := service.GetSystemStorageStats()
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("获取存储统计失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.Success(stats))
}

// GetStorageUsage godoc
// @Summary 获取存储使用情况
// @Description 获取当前用户的存储使用情况
// @Tags 文件
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /storage/usage [get]
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

// GetAdminStorageUsage godoc
// @Summary 获取全局存储使用情况
// @Description 管理员获取所有用户的存储使用情况
// @Tags 系统管理
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /admin/storage/usage [get]
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
