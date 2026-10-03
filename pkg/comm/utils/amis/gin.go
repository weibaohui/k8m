package amis

import (
	"context"
	"fmt"

	"github.com/weibaohui/k8m/pkg/constants"
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
