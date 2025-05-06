package utils

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/weibaohui/k8m/pkg/constants"
)

// GetJWTClaims 从 Gin 上下文的请求头或查询参数中提取并解析 JWT，返回其 claims。
// 若未提供 Token、Token 无效或 claims 解析失败，则返回相应错误。
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

// GetUsernameFromToken 从JWT令牌字符串中解析并返回用户名。
// 如果令牌无效或未包含用户名字段，则返回错误。
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

// GetJwtMapClaimsFromToken 解析并验证给定的 JWT token 字符串，返回其中的 claims。
// 如果 token 缺失、无效或 claims 解析失败，则返回相应错误。
//
// @param authToken 需要解析的 JWT token 字符串，可带有 "Bearer " 前缀。
// @param jwtTokenSecret 用于验证 token 的密钥。
// @return jwt.MapClaims 解析出的 JWT claims。
// @return error 解析或验证失败时返回的错误。
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
	// 主动检查有效期
	if err := claims.Valid(); err != nil {
		return nil, err
	}
	return claims, nil
}
