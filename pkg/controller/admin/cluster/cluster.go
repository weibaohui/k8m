package cluster

import (
	"fmt"
	"time"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/service"
)

func List(c *gin.Context) {
	user, role := amis.GetLoginUser(c)

	clusters := service.ClusterService().AllClusters()
	if role != models.RolePlatformAdmin {
		userCluster, err := service.UserService().GetClusters(user)
		if err != nil {
			amis.WriteJsonData(c, gin.H{
				"options": make([]map[string]string, 0),
			})
			return
		}
		clusters = slice.Filter(clusters, func(index int, cluster *service.ClusterConfig) bool {
			return slice.Contain(userCluster, cluster.GetClusterID())
		})
	}
	amis.WriteJsonData(c, clusters)
}

func FileOptionList(c *gin.Context) {
	clusters := service.ClusterService().AllClusters()

	if len(clusters) == 0 {
		amis.WriteJsonData(c, gin.H{
			"options": make([]map[string]string, 0),
		})
		return
	}

	var fileNames []string
	for _, cluster := range clusters {
		fileNames = append(fileNames, cluster.FileName)
	}
	fileNames = slice.Unique(fileNames)
	var options []map[string]interface{}
	for _, fn := range fileNames {
		options = append(options, map[string]interface{}{
			"label": fn,
			"value": fn,
		})
	}

	amis.WriteJsonData(c, gin.H{
		"options": options,
	})
}

func Scan(c *gin.Context) {
	service.ClusterService().Scan()
	amis.WriteJsonData(c, "ok")
}

func Reconnect(c *gin.Context) {
	fileName := c.Param("fileName")
	contextName := c.Param("contextName")
	service.ClusterService().Connect(fileName, contextName)
	amis.WriteJsonOKMsg(c, "已执行，请查看最新状态")
}
func Disconnect(c *gin.Context) {
	fileName := c.Param("fileName")
	contextName := c.Param("contextName")
	service.ClusterService().Disconnect(fileName, contextName)
	amis.WriteJsonOKMsg(c, "已执行，请查看最新状态")
}

func SetDefault(c *gin.Context) {
	fileName := c.Param("fileName")
	contextName := c.Param("contextName")
	cookieValue := fileName + "/" + contextName
	if cookieValue == "/" {
		return
	}

	c.SetCookie(
		"selectedCluster",
		cookieValue,
		int(24*time.Hour.Seconds()), // 有效期（秒），这里是 1 天,
		"/",
		"",
		false,
		false,
	)

	go func() {
		// 如果没有连接，那么进行一次连接。
		if !service.ClusterService().IsConnected(cookieValue) {
			service.ClusterService().Connect(fileName, contextName)
		}
	}()
	amis.WriteJsonOK(c)

}

func SetDefaultInCluster(c *gin.Context) {
	c.SetCookie(
		"selectedCluster",
		"InCluster",
		int(24*time.Hour.Seconds()), // 有效期（秒），这里是 1 天,
		"/",
		"",
		false,
		false,
	)
	amis.WriteJsonOK(c)

}
