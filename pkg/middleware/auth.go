package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"k8s.io/klog/v2"
)

func AuthMiddleware() gin.HandlerFunc {

	// 定义 JWT 密钥
	// todo 作为参数项
	var jwtSecret = []byte("your-secret-key")

	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			// 尝试从query中获取
			tokenString = c.Query("token")
			klog.V(6).Infof("从query中获取token为%v", tokenString)
			if tokenString == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "未提供 Token"})
				c.Abort()
				return
			}
		}

		if strings.HasPrefix(tokenString, "Bearer ") {
			// token Bear xxxx
			tokenString = tokenString[7:]
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Token 无效"})
			c.Abort()
			return
		}

		c.Next()
	}
}
