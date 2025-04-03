package utils

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// GetJWTClaims 从请求中获取JWT claims
func GetJWTClaims(c *gin.Context, jwtTokenSecret string) (jwt.MapClaims, error) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		// 尝试从query中获取
		tokenString = c.Query("token")
		if tokenString == "" {
			return nil, fmt.Errorf("未提供 Token")
		}
	}

	if strings.HasPrefix(tokenString, "Bearer ") {
		// token Bear xxxx
		tokenString = tokenString[7:]
	}

	var jwtSecret = []byte(jwtTokenSecret)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("Token 无效")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid JWT claims")
	}

	return claims, nil
}
