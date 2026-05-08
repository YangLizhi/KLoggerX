package v1

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"kloggerx-server/config"
	"kloggerx-server/internal/model"
	"kloggerx-server/internal/pkg/utils"
	"kloggerx-server/internal/repository/mysql"
	redisRepo "kloggerx-server/internal/repository/redis"
	"kloggerx-server/internal/service"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// AdminGetUserList godoc
// @Summary 管理员获取用户列表
// @Description 管理员分页获取用户列表
// @Tags 系统管理
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param keyword query string false "搜索关键词"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /admin/users [get]
func AdminGetUserList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	keyword := c.Query("keyword")
	role := c.Query("role")

	var users []model.User
	var total int64

	query := mysql.DB.Model(&model.User{})
	if keyword != "" {
		query = query.Where("username LIKE ? OR email LIKE ? OR nickname LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if role != "" {
		query = query.Where("role = ?", role)
	}

	query.Count(&total)
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error; err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("获取用户列表失败"))
		return
	}

	// Hide passwords
	for i := range users {
		users[i].Password = ""
	}

	c.JSON(http.StatusOK, model.Success(model.PaginatedData{List: users, Total: total, Page: page, PageSize: pageSize}))
}

// AdminCreateUser godoc
// @Summary 管理员创建用户
// @Description 管理员创建新用户
// @Tags 系统管理
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /admin/users [post]
func AdminCreateUser(c *gin.Context) {
	var req struct {
		Username     string `json:"username" binding:"required,max=50"`
		Email        string `json:"email" binding:"required,email,max=200"`
		Password     string `json:"password" binding:"required,min=6,max=128"`
		Nickname     string `json:"nickname" binding:"max=100"`
		Role         string `json:"role" binding:"max=50"`
		DepartmentID uint   `json:"departmentId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}

	// Check if username or email exists
	var count int64
	mysql.DB.Model(&model.User{}).Where("username = ? OR email = ?", req.Username, req.Email).Count(&count)
	if count > 0 {
		c.JSON(http.StatusOK, model.ErrorMsg("用户名或邮箱已存在"))
		return
	}

	// Hash password
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("密码加密失败"))
		return
	}

	if req.Role == "" {
		req.Role = "member"
	}

	user := model.User{
		Username:     req.Username,
		Email:        req.Email,
		Password:     string(hashedPwd),
		Nickname:     req.Nickname,
		Role:         req.Role,
		DepartmentID: req.DepartmentID,
	}
	if err := mysql.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("创建用户失败"))
		return
	}

	user.Password = ""
	c.JSON(http.StatusOK, model.Success(user))
}

// AdminUpdateUser godoc
// @Summary 管理员更新用户
// @Description 管理员更新用户信息
// @Tags 系统管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /admin/users/{id} [put]
func AdminUpdateUser(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var req struct {
		Nickname     string `json:"nickname" binding:"max=100"`
		Email        string `json:"email" binding:"max=200"`
		Role         string `json:"role" binding:"max=50"`
		DepartmentID uint   `json:"departmentId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}

	updates := map[string]interface{}{
		"nickname":      req.Nickname,
		"email":         req.Email,
		"role":          req.Role,
		"department_id": req.DepartmentID,
	}

	if err := mysql.DB.Model(&model.User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("更新用户失败"))
		return
	}

	c.JSON(http.StatusOK, model.Success(nil))
}

// AdminDeleteUser godoc
// @Summary 管理员删除用户
// @Description 管理员删除用户
// @Tags 系统管理
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /admin/users/{id} [delete]
func AdminDeleteUser(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	uid := utils.GetUserID(c.MustGet("userId"))

	if uint(id) == uid {
		c.JSON(http.StatusOK, model.ErrorMsg("不能删除自己"))
		return
	}

	if err := mysql.DB.Delete(&model.User{}, id).Error; err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("删除用户失败"))
		return
	}

	c.JSON(http.StatusOK, model.Success(nil))
}

// AdminResetUserPassword godoc
// @Summary 重置用户密码
// @Description 管理员重置用户密码
// @Tags 系统管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /admin/users/{id}/reset-password [post]
func AdminResetUserPassword(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var req struct {
		Password string `json:"password" binding:"required,min=6,max=128"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("密码加密失败"))
		return
	}

	if err := mysql.DB.Model(&model.User{}).Where("id = ?", id).Update("password", string(hashedPwd)).Error; err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("重置密码失败"))
		return
	}

	c.JSON(http.StatusOK, model.Success(nil))
}

// AdminGetDepartmentTree godoc
// @Summary 获取部门树
// @Description 获取部门树结构
// @Tags 系统管理
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /admin/departments [get]
func AdminGetDepartmentTree(c *gin.Context) {
	var departments []model.Department
	if err := mysql.DB.Find(&departments).Error; err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("获取部门列表失败"))
		return
	}

	// Build tree structure using recursive approach
	// Helper function to build tree recursively
	var buildTree func(parentID *uint) []model.Department
	buildTree = func(parentID *uint) []model.Department {
		var result []model.Department
		for i := range departments {
			// Check if this department has the specified parent
			var isChild bool
			if parentID == nil && departments[i].ParentID == nil {
				isChild = true
			} else if parentID != nil && departments[i].ParentID != nil && *departments[i].ParentID == *parentID {
				isChild = true
			}

			if isChild {
				dept := departments[i]
				// Recursively get children
				dept.Children = buildTree(&dept.ID)
				result = append(result, dept)
			}
		}
		return result
	}

	// Build tree starting from root (nil parent)
	roots := buildTree(nil)

	c.JSON(http.StatusOK, model.Success(roots))
}

// AdminCreateDepartment godoc
// @Summary 创建部门
// @Description 管理员创建新部门
// @Tags 系统管理
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /admin/departments [post]
func AdminCreateDepartment(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required,max=100"`
		ParentID *uint  `json:"parentId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}

	dept := model.Department{
		Name:     req.Name,
		ParentID: req.ParentID,
	}
	if err := mysql.DB.Create(&dept).Error; err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("创建部门失败"))
		return
	}

	c.JSON(http.StatusOK, model.Success(dept))
}

// AdminUpdateDepartment godoc
// @Summary 更新部门
// @Description 管理员更新部门信息
// @Tags 系统管理
// @Accept json
// @Produce json
// @Param id path int true "部门ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /admin/departments/{id} [put]
func AdminUpdateDepartment(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var req struct {
		Name     string `json:"name" binding:"required,max=100"`
		ParentID *uint  `json:"parentId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}

	updates := map[string]interface{}{
		"name":      req.Name,
		"parent_id": req.ParentID,
	}

	if err := mysql.DB.Model(&model.Department{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("更新部门失败"))
		return
	}

	c.JSON(http.StatusOK, model.Success(nil))
}

// AdminDeleteDepartment godoc
// @Summary 删除部门
// @Description 管理员删除部门
// @Tags 系统管理
// @Produce json
// @Param id path int true "部门ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /admin/departments/{id} [delete]
func AdminDeleteDepartment(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	// Check if department has members
	var userCount int64
	mysql.DB.Model(&model.User{}).Where("department_id = ?", id).Count(&userCount)
	if userCount > 0 {
		c.JSON(http.StatusOK, model.ErrorMsg("部门下有成员，不能删除"))
		return
	}

	// Check if department has children
	var childCount int64
	mysql.DB.Model(&model.Department{}).Where("parent_id = ?", id).Count(&childCount)
	if childCount > 0 {
		c.JSON(http.StatusOK, model.ErrorMsg("部门下有子部门，不能删除"))
		return
	}

	if err := mysql.DB.Delete(&model.Department{}, id).Error; err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("删除部门失败"))
		return
	}

	c.JSON(http.StatusOK, model.Success(nil))
}

// AdminGetDepartmentMembers godoc
// @Summary 获取部门成员
// @Description 获取部门下的成员列表
// @Tags 系统管理
// @Produce json
// @Param id path int true "部门ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /admin/departments/{id}/members [get]
func AdminGetDepartmentMembers(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var users []model.User
	if err := mysql.DB.Where("department_id = ?", id).Find(&users).Error; err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("获取部门成员失败"))
		return
	}

	for i := range users {
		users[i].Password = ""
	}

	c.JSON(http.StatusOK, model.Success(users))
}

// ─── Feedback Review ──────────────────────────────────────────────────────────

// GetPendingReviewsHandler godoc
// @Summary 获取待审核列表
// @Description 获取待审核的反馈列表
// @Tags 系统管理
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /admin/feedback/reviews [get]
func GetPendingReviewsHandler(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	items, total, err := service.GetPendingReviews(page, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("获取审核列表失败"))
		return
	}

	c.JSON(http.StatusOK, model.Success(model.PaginatedData{List: items, Total: total, Page: page, PageSize: pageSize}))
}

// ApproveReviewHandler godoc
// @Summary 通过审核
// @Description 通过反馈审核
// @Tags 系统管理
// @Accept json
// @Produce json
// @Param id path int true "审核ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /admin/feedback/reviews/{id}/approve [post]
func ApproveReviewHandler(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	reviewerID := utils.GetUserID(c.MustGet("userId"))

	var req struct {
		Comment string `json:"comment" binding:"max=2000"`
	}
	c.ShouldBindJSON(&req)

	if err := service.ApproveReview(uint(id), reviewerID, req.Comment); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.Success(nil))
}

// RejectReviewHandler godoc
// @Summary 拒绝审核
// @Description 拒绝反馈审核
// @Tags 系统管理
// @Accept json
// @Produce json
// @Param id path int true "审核ID"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /admin/feedback/reviews/{id}/reject [post]
func RejectReviewHandler(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	reviewerID := utils.GetUserID(c.MustGet("userId"))

	var req struct {
		Comment string `json:"comment"`
	}
	c.ShouldBindJSON(&req)

	if err := service.RejectReview(uint(id), reviewerID, req.Comment); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.Success(nil))
}

// GetReviewStatsHandler godoc
// @Summary 审核统计
// @Description 获取审核统计信息
// @Tags 系统管理
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /admin/feedback/reviews/stats [get]
func GetReviewStatsHandler(c *gin.Context) {
	stats, err := service.GetReviewStats()
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("获取统计失败"))
		return
	}

	c.JSON(http.StatusOK, model.Success(stats))
}

// ─── Dashboard ──────────────────────────────────────────────────────────────

var serverStartTime = time.Now()

// GetDashboardStats 获取仪表盘统计
func GetDashboardStats(c *gin.Context) {
	var totalUsers int64
	var totalDocuments int64
	var totalKnowledgeBases int64

	mysql.DB.Model(&model.User{}).Count(&totalUsers)
	mysql.DB.Model(&model.Document{}).Where("is_deleted = ?", false).Count(&totalDocuments)
	mysql.DB.Model(&model.KnowledgeBase{}).Count(&totalKnowledgeBases)

	// Total storage: sum of file sizes from file_records
	var totalStorageUsed int64
	mysql.DB.Model(&model.FileRecord{}).Select("COALESCE(SUM(size), 0)").Scan(&totalStorageUsed)

	// Today active users (users who logged in today)
	today := time.Now().Format("2006-01-02")
	var todayActiveUsers int64
	mysql.DB.Model(&model.User{}).Where("DATE(updated_at) = ?", today).Count(&todayActiveUsers)

	// New documents today
	var newDocumentsToday int64
	mysql.DB.Model(&model.Document{}).Where("DATE(created_at) = ? AND is_deleted = ?", today, false).Count(&newDocumentsToday)

	// Weekly trend: last 7 days
	type DayTrend struct {
		Date        string `json:"date"`
		Documents   int64  `json:"documents"`
		ActiveUsers int64  `json:"activeUsers"`
	}
	var weeklyTrend []DayTrend
	for i := 6; i >= 0; i-- {
		day := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		var docs int64
		var users int64
		mysql.DB.Model(&model.Document{}).Where("DATE(created_at) = ? AND is_deleted = ?", day, false).Count(&docs)
		mysql.DB.Model(&model.User{}).Where("DATE(updated_at) = ?", day).Count(&users)
		weeklyTrend = append(weeklyTrend, DayTrend{Date: day, Documents: docs, ActiveUsers: users})
	}

	c.JSON(http.StatusOK, model.Success(gin.H{
		"total_users":           totalUsers,
		"total_documents":       totalDocuments,
		"total_knowledge_bases": totalKnowledgeBases,
		"total_storage_used":    totalStorageUsed,
		"today_active_users":    todayActiveUsers,
		"new_documents_today":   newDocumentsToday,
		"weekly_trend":          weeklyTrend,
	}))
}

// GetSystemInfo 获取系统信息
func GetSystemInfo(c *gin.Context) {
	// Go version & OS info
	goVersion := runtime.Version()
	osInfo := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
	cpuCores := runtime.NumCPU()

	// Memory usage
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// MySQL status
	mysqlStatus := "connected"
	sqlDB, err := mysql.DB.DB()
	if err != nil || sqlDB.Ping() != nil {
		mysqlStatus = "disconnected"
	}

	// Redis status
	redisStatus := "connected"
	if redisRepo.RDB == nil || redisRepo.RDB.Ping(context.Background()).Err() != nil {
		redisStatus = "disconnected"
	}

	// MinIO status
	minioStatus := "configured"
	if config.Cfg.MinIO.Endpoint == "" {
		minioStatus = "not_configured"
	}

	// Uptime
	uptime := time.Since(serverStartTime).String()

	// Disk usage (uploads directory)
	var diskUsage int64
	uploadsDir := filepath.Join(".", "uploads")
	filepath.Walk(uploadsDir, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			diskUsage += info.Size()
		}
		return nil
	})

	c.JSON(http.StatusOK, model.Success(gin.H{
		"go_version":   goVersion,
		"os":           osInfo,
		"cpu_cores":    cpuCores,
		"memory_usage": memStats.Alloc,
		"memory_sys":   memStats.Sys,
		"mysql_status": mysqlStatus,
		"redis_status": redisStatus,
		"minio_status": minioStatus,
		"uptime":       uptime,
		"disk_usage":   diskUsage,
	}))
}

// AdminListDirectories godoc
// @Summary 列出目录
// @Description 列出服务器目录结构
// @Tags 系统管理
// @Produce json
// @Param path query string false "目录路径"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /admin/directories [get]
func AdminListDirectories(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		path = "/"
	}

	// Security check: only allow absolute paths
	if len(path) == 0 || path[0] != '/' {
		c.JSON(http.StatusOK, model.ErrorMsg("路径必须是绝对路径"))
		return
	}

	dirs, err := utils.ListDirectories(path)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("获取目录列表失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.Success(dirs))
}
