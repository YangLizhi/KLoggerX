package service

import (
	"errors"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/pkg/jwt"
	"kloggerx-server/internal/pkg/utils"
	"kloggerx-server/internal/repository/mysql"

	"gorm.io/gorm"
)

type LoginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type RegisterReq struct {
	Username        string `json:"username" binding:"required,min=2,max=20"`
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirmPassword" binding:"required"`
}

func Login(req LoginReq) (string, *model.User, error) {
	var user model.User
	if err := mysql.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, errors.New("用户不存在")
		}
		return "", nil, err
	}
	if !utils.CheckPassword(req.Password, user.Password) {
		return "", nil, errors.New("密码错误")
	}
	token, err := jwt.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return "", nil, err
	}
	return token, &user, nil
}

func Register(req RegisterReq) error {
	if req.Password != req.ConfirmPassword {
		return errors.New("两次密码不一致")
	}
	var count int64
	mysql.DB.Model(&model.User{}).Where("email = ?", req.Email).Count(&count)
	if count > 0 {
		return errors.New("邮箱已注册")
	}
	mysql.DB.Model(&model.User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		return errors.New("用户名已存在")
	}
	hashed, err := utils.HashPassword(req.Password)
	if err != nil {
		return err
	}
	user := model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashed,
		Nickname: req.Username,
		Role:     "member",
	}
	return mysql.DB.Create(&user).Error
}

func GetUserByID(id uint) (*model.User, error) {
	var user model.User
	err := mysql.DB.First(&user, id).Error
	return &user, err
}

func UpdateUser(id uint, updates map[string]interface{}) error {
	return mysql.DB.Model(&model.User{}).Where("id = ?", id).Updates(updates).Error
}

func ChangePassword(id uint, oldPwd, newPwd string) error {
	user, err := GetUserByID(id)
	if err != nil {
		return err
	}
	if !utils.CheckPassword(oldPwd, user.Password) {
		return errors.New("原密码错误")
	}
	hashed, err := utils.HashPassword(newPwd)
	if err != nil {
		return err
	}
	return mysql.DB.Model(user).Update("password", hashed).Error
}

func GetUserList(page, pageSize int, keyword string) ([]model.User, int64, error) {
	var users []model.User
	var total int64
	db := mysql.DB.Model(&model.User{})
	if keyword != "" {
		db = db.Where("username LIKE ? OR nickname LIKE ? OR email LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	db.Count(&total)
	err := db.Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error
	return users, total, err
}
