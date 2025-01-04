package cluster

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/service"
)

func List(c *gin.Context) {
	clusters := service.ClusterService().AllClusters()
	amis.WriteJsonData(c, clusters)
}

func Scan(c *gin.Context) {
	service.ClusterService().Scan()
	amis.WriteJsonData(c, "ok")
}

func Reconnect(c *gin.Context) {
	fileName := c.Param("fileName")
	contextName := c.Param("contextName")
	service.ClusterService().Reconnect(fileName, contextName)
	amis.WriteJsonOKMsg(c, "已执行，请查看最新状态")
}

func SetDefault(c *gin.Context) {
	fileName := c.Param("fileName")
	contextName := c.Param("contextName")
	cookieValue := fileName + "/" + contextName
	c.SetCookie(
		"selectedCluster",
		cookieValue,
		int(24*time.Hour.Seconds()), // 有效期（秒），这里是 1 天,
		"/",
		"",
		false,
		false,
	)
}
