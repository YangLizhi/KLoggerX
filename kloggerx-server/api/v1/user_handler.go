package v1

import (
	"net/http"
	"strconv"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/pkg/utils"
	"kloggerx-server/internal/service"

	"github.com/gin-gonic/gin"
)

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

func GetUserInfo(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	user, err := service.GetUserByID(uid)
	if err != nil {
		c.JSON(http.StatusOK, model.ErrorMsg("用户不存在"))
		return
	}
	c.JSON(http.StatusOK, model.Success(user))
}

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

func ChangeUserPassword(c *gin.Context) {
	uid := utils.GetUserID(c.MustGet("userId"))
	var body struct {
		OldPassword string `json:"oldPassword" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required,min=6"`
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

// SearchUsers handles user search for @mention feature
// GET /api/v1/users/search?q=xxx
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
