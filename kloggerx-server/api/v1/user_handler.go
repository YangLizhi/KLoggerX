package v1

import (
	"net/http"
	"strconv"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/pkg/utils"
	"kloggerx-server/internal/service"

	"github.com/gin-gonic/gin"
)

// UserLogin godoc
// @Summary 用户登录
// @Description 用户账号密码登录
// @Tags 认证
// @Accept json
// @Produce json
// @Param body body service.LoginReq true "登录请求"
// @Success 200 {object} model.Response
// @Router /user/login [post]
func UserLogin(c *gin.Context) {
	var req service.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误: "+err.Error()))
		return
	}
	token, user, err := service.Login(req)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(gin.H{"token": token, "user": user}))
}

// UserRegister godoc
// @Summary 用户注册
// @Description 注册新用户
// @Tags 认证
// @Accept json
// @Produce json
// @Param body body service.RegisterReq true "注册请求"
// @Success 200 {object} model.Response
// @Router /user/register [post]
func UserRegister(c *gin.Context) {
	var req service.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误: "+err.Error()))
		return
	}
	if err := service.Register(req); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

// GetUserInfo godoc
// @Summary 获取当前用户信息
// @Description 获取当前登录用户的详细信息
// @Tags 用户
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /user/info [get]
func GetUserInfo(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	user, err := service.GetUserByID(uid)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("用户不存在"))
		return
	}
	c.JSON(http.StatusOK, model.Success(user))
}

// UpdateUserInfo godoc
// @Summary 更新用户信息
// @Description 更新当前用户的昵称、头像等信息
// @Tags 用户
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /user/update [post]
func UpdateUserInfo(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}
	allowed := map[string]bool{"nickname": true, "avatar": true}
	updates := make(map[string]interface{})
	for k, v := range body {
		if allowed[k] {
			updates[k] = v
		}
	}
	if err := service.UpdateUser(uid, updates); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

// ChangeUserPassword godoc
// @Summary 修改密码
// @Description 修改当前用户密码
// @Tags 用户
// @Accept json
// @Produce json
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /user/change-password [post]
func ChangeUserPassword(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	var body struct {
		OldPassword string `json:"oldPassword" binding:"required,max=128"`
		NewPassword string `json:"newPassword" binding:"required,min=6,max=128"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("参数错误"))
		return
	}
	if err := service.ChangePassword(uid, body.OldPassword, body.NewPassword); err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(nil))
}

// GetUserList godoc
// @Summary 获取用户列表
// @Description 分页获取用户列表
// @Tags 用户
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param keyword query string false "搜索关键词"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /user/list [get]
func GetUserList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	keyword := c.Query("keyword")
	users, total, err := service.GetUserList(page, pageSize, keyword)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.Success(model.PaginatedData{List: users, Total: total, Page: page, PageSize: pageSize}))
}

// SearchUsers godoc
// @Summary 搜索用户
// @Description 根据用户名搜索用户（用于@提及）
// @Tags 用户
// @Produce json
// @Param q query string true "搜索关键词"
// @Success 200 {object} model.Response
// @Security BearerAuth
// @Router /users/search [get]
func SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusOK, model.Success([]gin.H{}))
		return
	}

	// Limit to 10 results for mention dropdown
	users, err := service.SearchUsersByUsername(query, 10)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg(err.Error()))
		return
	}

	// Format response
	result := make([]gin.H, 0, len(users))
	for _, user := range users {
		result = append(result, gin.H{
			"id":       user.ID,
			"username": user.Username,
			"avatar":   user.Avatar,
		})
	}
	c.JSON(http.StatusOK, model.Success(result))
}
