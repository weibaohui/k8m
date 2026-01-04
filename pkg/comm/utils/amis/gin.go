package amis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/flag"
	"github.com/weibaohui/k8m/pkg/response"
	"github.com/weibaohui/kom/kom"
)

func GetSelectedCluster(c *response.Context) (string, error) {
	selectedCluster := c.GetString("cluster")
	if kom.Cluster(selectedCluster) == nil {
		return "", fmt.Errorf("cluster %s not found", selectedCluster)
	}
	return selectedCluster, nil
}

// GetLoginUser 获取当前登录用户名
func GetLoginUser(c *response.Context) string {
	user := c.GetString(constants.JwtUserName)
	return user
}

func GetContextWithUser(c *response.Context) context.Context {
	user := GetLoginUser(c)
	ctx := context.WithValue(c.Request.Context(), constants.JwtUserName, user)

	return ctx
}

func GetContextForAdmin() context.Context {
	// cfg := flag.Init()
	// todo 内部使用逻辑
	ctx := context.WithValue(context.Background(), constants.JwtUserName, "admin")
	return ctx
}

// GenerateJWTTokenOnlyUserName  生成 Token，仅包含Username
func GenerateJWTTokenOnlyUserNameInMCP(username string, duration time.Duration) (string, error) {
	if username == "" {
		return "", errors.New("username cannot be empty")
	}
	name := constants.JwtUserName

	var token = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		name:  username,
		"exp": time.Now().Add(duration).Unix(),
	})
	cfg := flag.Init()
	var jwtSecret = []byte(cfg.JwtTokenSecret)
	return token.SignedString(jwtSecret)
}
