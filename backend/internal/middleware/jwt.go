// JWT中间件 - 用户认证
// 文件名: jwt.go
// 路径: /workspace/backend/internal/middleware/jwt.go

package middleware

import (
	"net/http"
	"strings"
	"time"

	"blog-cms/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT声明结构
type Claims struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	Role     int    `json:"role"`
	jwt.RegisteredClaims
}

var jwtConfig *config.JWTConfig

// InitJWT 初始化JWT配置
func InitJWT(cfg *config.JWTConfig) {
	jwtConfig = cfg
}

// GenerateToken 生成JWT令牌
func GenerateToken(userID uint64, username string, role int) (string, error) {
	now := time.Now()
	expireTime := now.Add(time.Duration(jwtConfig.ExpireHour) * time.Hour)

	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "blog-cms",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtConfig.Secret))
}

// ParseToken 解析JWT令牌
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtConfig.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

// JWTAuth JWT认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未提供认证令牌",
			})
			c.Abort()
			return
		}

		// 检查Bearer格式
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "认证令牌格式错误",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "认证令牌无效或已过期",
			})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// RequireRole 角色权限检查中间件
func RequireRole(requiredRoles ...int) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "请先登录",
			})
			c.Abort()
			return
		}

		userRole := role.(int)
		hasPermission := false
		for _, requiredRole := range requiredRoles {
			if userRole >= requiredRole {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限不足",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserIDFromContext 从上下文中获取用户ID
func GetUserIDFromContext(c *gin.Context) (uint64, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	return userID.(uint64), true
}

// GetUserFromContext 从上下文中获取用户信息
func GetUserFromContext(c *gin.Context) (uint64, string, int, bool) {
	userID, uidExists := c.Get("user_id")
	username, nameExists := c.Get("username")
	role, roleExists := c.Get("role")

	if !uidExists || !nameExists || !roleExists {
		return 0, "", 0, false
	}

	return userID.(uint64), username.(string), role.(int), true
}
