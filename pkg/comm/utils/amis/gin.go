package amis

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/kom/kom"
)

func GetSelectedCluster(c *gin.Context) (string, error) {
	selectedCluster := c.GetString("cluster")
	if kom.Cluster(selectedCluster) == nil {
		return "", fmt.Errorf("cluster %s not found", selectedCluster)
	}
	return selectedCluster, nil
}

// GetLoginUser 获取当前登录用户名
func GetLoginUser(c *gin.Context) string {
	user := c.GetString(constants.JwtUserName)
	return user
}

func GetContextWithUser(c *gin.Context) context.Context {
	user := GetLoginUser(c)
	ctx := context.WithValue(c.Request.Context(), constants.JwtUserName, user)

	return ctx
}
