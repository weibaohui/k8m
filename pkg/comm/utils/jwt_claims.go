package utils

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/weibaohui/k8m/pkg/constants"
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

// GetUsernameFromToken 从Authorization中解析出username
func GetUsernameFromToken(authToken string, jwtTokenSecret string) (string, error) {
	claims, err := GetJwtMapClaimsFromToken(authToken, jwtTokenSecret)
	if err != nil {
		return "", err
	}

	username, ok := claims[constants.JwtUserName].(string)
	if !ok {
		return "", fmt.Errorf("Token中未包含username")
	}

	return username, nil
}

// GetJwtMapClaimsFromToken 从token中解析出claims
func GetJwtMapClaimsFromToken(authToken string, jwtTokenSecret string) (jwt.MapClaims, error) {
	if authToken == "" {
		return nil, fmt.Errorf("未提供 Token")
	}

	if strings.HasPrefix(authToken, "Bearer ") {
		// token Bear xxxx
		authToken = authToken[7:]
	}

	var jwtSecret = []byte(jwtTokenSecret)

	token, err := jwt.Parse(authToken, func(token *jwt.Token) (interface{}, error) {
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
