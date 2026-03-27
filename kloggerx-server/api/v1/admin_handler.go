package v1

import (
	"net/http"
	"strconv"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/pkg/utils"
	"kloggerx-server/internal/repository/mysql"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// GetUserList returns list of users with pagination
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

// CreateUser creates a new user
func AdminCreateUser(c *gin.Context) {
	var req struct {
		Username     string `json:"username" binding:"required"`
		Email        string `json:"email" binding:"required,email"`
		Password     string `json:"password" binding:"required,min=6"`
		Nickname     string `json:"nickname"`
		Role         string `json:"role"`
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

// UpdateUser updates a user
func AdminUpdateUser(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var req struct {
		Nickname     string `json:"nickname"`
		Email        string `json:"email"`
		Role         string `json:"role"`
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

// DeleteUser deletes a user
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

// ResetUserPassword resets a user's password
func AdminResetUserPassword(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var req struct {
		Password string `json:"password" binding:"required,min=6"`
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

// GetDepartmentTree returns department tree
func AdminGetDepartmentTree(c *gin.Context) {
	var departments []model.Department
	if err := mysql.DB.Find(&departments).Error; err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("获取部门列表失败"))
		return
	}

	// Build tree
	deptMap := make(map[uint]*model.Department)
	for i := range departments {
		deptMap[departments[i].ID] = &departments[i]
	}

	var roots []model.Department
	for i := range departments {
		if departments[i].ParentID == nil {
			roots = append(roots, departments[i])
		}
	}

	c.JSON(http.StatusOK, model.Success(roots))
}

// CreateDepartment creates a new department
func AdminCreateDepartment(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
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

// UpdateDepartment updates a department
func AdminUpdateDepartment(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var req struct {
		Name     string `json:"name" binding:"required"`
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

// DeleteDepartment deletes a department
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

// GetDepartmentMembers returns members of a department
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
