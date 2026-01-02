package route

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/gllog/models"
	"k8s.io/klog/v2"
)

func RegisterManagementRoutes(arg *gin.RouterGroup) {
	g := arg.Group("/plugins/" + modules.PluginNameGlobalLog)
	g.GET("/list", ListGlobalLog)
	klog.V(6).Infof("注册全局日志插件管理路由(mgm)")
}

func ListGlobalLog(c *gin.Context) {
	cluster := c.Query("cluster")
	namespace := c.Query("namespace")
	nodeName := c.Query("node_name")
	podName := c.Query("pod_name")
	container := c.Query("container")
	keyword := c.Query("keyword")
	logLevel := c.Query("log_level")
	source := c.Query("source")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	ctx := amis.GetContextWithUser(c)

	logs, err := models.ListGlobalLog(ctx, cluster, namespace, nodeName, podName, container, keyword, logLevel, source, startTime, endTime)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonListWithTotal(c, int64(len(logs)), logs)
}
