package service

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/url"
	"time"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/pkg/utils"
	"kloggerx-server/internal/repository/mysql"

	ldapv3 "github.com/go-ldap/ldap/v3"
	"gorm.io/gorm"
)

// LDAPConfig LDAP连接配置
type LDAPConfig struct {
	Host         string   `json:"host"`
	Port         int      `json:"port"`
	BaseDN       string   `json:"base_dn"`
	BindDN       string   `json:"bind_dn"`
	BindPassword string   `json:"bind_password"`
	UseSSL       bool     `json:"use_ssl"`
	UserFilter   string   `json:"user_filter"`
	Attributes   []string `json:"attributes"`
}

// LDAPUser LDAP用户信息
type LDAPUser struct {
	DN         string `json:"dn"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	Department string `json:"department"`
}

// ParseLDAPServerURL 从 ldap://host:port 格式解析出LDAPConfig的Host/Port/UseSSL
func ParseLDAPServerURL(server string) (host string, port int, useSSL bool, err error) {
	u, err := url.Parse(server)
	if err != nil {
		return "", 0, false, fmt.Errorf("无效的服务器地址: %v", err)
	}

	useSSL = u.Scheme == "ldaps"
	host = u.Hostname()
	if host == "" {
		return "", 0, false, fmt.Errorf("服务器地址缺少主机名")
	}

	portStr := u.Port()
	if portStr == "" {
		if useSSL {
			port = 636
		} else {
			port = 389
		}
	} else {
		fmt.Sscanf(portStr, "%d", &port)
	}

	return host, port, useSSL, nil
}

// TestLDAPConnection 测试LDAP连接
func TestLDAPConnection(config LDAPConfig) error {
	conn, err := connectLDAP(config)
	if err != nil {
		return fmt.Errorf("连接失败: %v", err)
	}
	defer conn.Close()

	// 尝试bind验证凭证
	err = conn.Bind(config.BindDN, config.BindPassword)
	if err != nil {
		return fmt.Errorf("认证失败: %v", err)
	}
	return nil
}

// ListLDAPUsers 列出LDAP用户
func ListLDAPUsers(config LDAPConfig, page, pageSize int) ([]LDAPUser, int, error) {
	conn, err := connectLDAP(config)
	if err != nil {
		return nil, 0, fmt.Errorf("连接失败: %v", err)
	}
	defer conn.Close()

	err = conn.Bind(config.BindDN, config.BindPassword)
	if err != nil {
		return nil, 0, fmt.Errorf("认证失败: %v", err)
	}

	filter := config.UserFilter
	if filter == "" {
		filter = "(objectClass=person)"
	}

	searchReq := ldapv3.NewSearchRequest(
		config.BaseDN,
		ldapv3.ScopeWholeSubtree, ldapv3.NeverDerefAliases,
		0, 0, false,
		filter,
		[]string{"cn", "mail", "uid", "sAMAccountName", "department", "displayName"},
		nil,
	)

	result, err := conn.Search(searchReq)
	if err != nil {
		return nil, 0, fmt.Errorf("搜索失败: %v", err)
	}

	total := len(result.Entries)
	var users []LDAPUser

	// 分页
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	for _, entry := range result.Entries[start:end] {
		username := entry.GetAttributeValue("sAMAccountName")
		if username == "" {
			username = entry.GetAttributeValue("uid")
		}
		name := entry.GetAttributeValue("displayName")
		if name == "" {
			name = entry.GetAttributeValue("cn")
		}

		users = append(users, LDAPUser{
			DN:         entry.DN,
			Username:   username,
			Email:      entry.GetAttributeValue("mail"),
			Name:       name,
			Department: entry.GetAttributeValue("department"),
		})
	}

	return users, total, nil
}

// ImportLDAPUsers 导入LDAP用户到本地系统
func ImportLDAPUsers(config LDAPConfig, userDNs []string) (imported int, errs []string) {
	conn, err := connectLDAP(config)
	if err != nil {
		return 0, []string{fmt.Sprintf("连接失败: %v", err)}
	}
	defer conn.Close()

	err = conn.Bind(config.BindDN, config.BindPassword)
	if err != nil {
		return 0, []string{fmt.Sprintf("认证失败: %v", err)}
	}

	for _, dn := range userDNs {
		// 搜索该用户
		searchReq := ldapv3.NewSearchRequest(
			dn, ldapv3.ScopeBaseObject, ldapv3.NeverDerefAliases,
			0, 0, false, "(objectClass=*)",
			[]string{"cn", "mail", "uid", "sAMAccountName", "displayName"},
			nil,
		)
		result, searchErr := conn.Search(searchReq)
		if searchErr != nil || len(result.Entries) == 0 {
			errs = append(errs, fmt.Sprintf("找不到用户: %s", dn))
			continue
		}

		entry := result.Entries[0]
		username := entry.GetAttributeValue("sAMAccountName")
		if username == "" {
			username = entry.GetAttributeValue("uid")
		}
		if username == "" {
			errs = append(errs, fmt.Sprintf("用户无用户名: %s", dn))
			continue
		}
		email := entry.GetAttributeValue("mail")
		displayName := entry.GetAttributeValue("displayName")
		if displayName == "" {
			displayName = entry.GetAttributeValue("cn")
		}

		// 创建本地用户(如果不存在)
		createErr := CreateUserFromLDAP(username, email, displayName)
		if createErr != nil {
			errs = append(errs, fmt.Sprintf("导入用户 %s 失败: %v", username, createErr))
			continue
		}
		imported++
	}
	return
}

// CreateUserFromLDAP 从LDAP信息创建本地用户
func CreateUserFromLDAP(username, email, displayName string) error {
	var existingUser model.User

	// 检查用户名是否已存在
	err := mysql.DB.Where("username = ?", username).First(&existingUser).Error
	if err == nil {
		// 用户已存在，更新auth_source标记
		mysql.DB.Model(&existingUser).Updates(map[string]interface{}{
			"auth_source": "ldap",
		})
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// 如果email不为空，也检查email是否已存在
	if email != "" {
		err = mysql.DB.Where("email = ?", email).First(&existingUser).Error
		if err == nil {
			mysql.DB.Model(&existingUser).Updates(map[string]interface{}{
				"auth_source": "ldap",
			})
			return nil
		}
	}

	// 生成随机密码(LDAP用户通过LDAP认证，不使用本地密码)
	randomPwd := utils.GenerateRandomString(16)
	hashed, err := utils.HashPassword(randomPwd)
	if err != nil {
		return fmt.Errorf("生成密码失败: %v", err)
	}

	// 如果没有email，使用 username@ldap.local 作为占位
	if email == "" {
		email = username + "@ldap.local"
	}

	nickname := displayName
	if nickname == "" {
		nickname = username
	}

	user := model.User{
		Username:   username,
		Email:      email,
		Password:   hashed,
		Nickname:   nickname,
		Role:       "member",
		AuthSource: "ldap",
	}

	return mysql.DB.Create(&user).Error
}

// connectLDAP 建立LDAP连接
func connectLDAP(config LDAPConfig) (*ldapv3.Conn, error) {
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	var conn *ldapv3.Conn
	var err error

	if config.UseSSL {
		conn, err = ldapv3.DialTLS("tcp", addr, &tls.Config{InsecureSkipVerify: true})
	} else {
		conn, err = ldapv3.Dial("tcp", addr)
	}
	if err != nil {
		return nil, err
	}

	// 设置超时
	conn.SetTimeout(10 * time.Second)

	return conn, nil
}
