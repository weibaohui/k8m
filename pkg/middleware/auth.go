package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/flag"
	"k8s.io/klog/v2"
)

func AuthMiddleware() gin.HandlerFunc {

	// 定义 JWT 密钥
	cfg := flag.Init()
	var jwtSecret = []byte(cfg.JwtTokenSecret)

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
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			amis.WriteJsonError(c, fmt.Errorf("invalid JWT claims"))
			c.Abort()
			return
		}
		c.Set(constants.JwtUserName, claims[constants.JwtUserName])
		c.Set(constants.JwtUserRole, claims[constants.JwtUserRole])
		c.Next()
	}
}
