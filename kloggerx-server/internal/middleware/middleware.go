package middleware

import (
	"net/http"
	"strings"
	"time"

	"kloggerx-server/internal/model"
	"kloggerx-server/internal/pkg/jwt"
	"kloggerx-server/internal/service"

	"github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, model.Error(401, "未登录"))
			c.Abort()
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, model.Error(401, "Token格式错误"))
			c.Abort()
			return
		}
		claims, err := jwt.ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, model.Error(401, "Token无效或已过期"))
			c.Abort()
			return
		}
		c.Set("userId", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, Accept")
		c.Header("Access-Control-Max-Age", "86400")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func Logger() gin.HandlerFunc {
	return gin.Logger()
}

// AuditLog 审计日志中间件
// 仅对写操作(POST/PUT/DELETE)生效
func AuditLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" {
			c.Next()
			return
		}

		startTime := time.Now()
		c.Next()

		// 记录操作日志
		userID, _ := c.Get("userId")
		userName, _ := c.Get("username")
		uid, _ := userID.(uint)
		uname, _ := userName.(string)

		log := &model.OperationLog{
			UserID:    uid,
			UserName:  uname,
			Action:    c.Request.Method + " " + c.FullPath(),
			Resource:  c.Request.URL.Path,
			IP:        c.ClientIP(),
			UserAgent: c.GetHeader("User-Agent"),
			Status:    c.Writer.Status(),
			Duration:  time.Since(startTime).Milliseconds(),
		}
		// 异步写入数据库
		go service.CreateAuditLog(log)
	}
}

// AdminOnly checks if user is admin
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			c.JSON(http.StatusOK, model.Error(403, "无权限访问"))
			c.Abort()
			return
		}
		c.Next()
	}
}

// AuthQueryToken supports authentication via query parameter "token" (for iframe/embedded scenarios)
// Falls back to standard Authorization header if query param is not present.
func AuthQueryToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.Query("token")
		if tokenStr == "" {
			// Fallback to Authorization header
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				c.JSON(http.StatusUnauthorized, model.Error(401, "未登录"))
				c.Abort()
				return
			}
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				c.JSON(http.StatusUnauthorized, model.Error(401, "Token格式错误"))
				c.Abort()
				return
			}
			tokenStr = parts[1]
		}

		claims, err := jwt.ParseToken(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, model.Error(401, "Token无效或已过期"))
			c.Abort()
			return
		}
		c.Set("userId", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}
