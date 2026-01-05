package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/gllog/models"
	"github.com/weibaohui/k8m/pkg/response"
	"k8s.io/klog/v2"
)

// RegisterManagementRoutes 注册全局日志插件管理路由 - Gin到Chi迁移
func RegisterManagementRoutes(arg chi.Router) {
	prefix := "/plugins/" + modules.PluginNameGlobalLog
	arg.Get(prefix+"/list", response.Adapter(ListGlobalLog))
	klog.V(6).Infof("注册全局日志插件管理路由(mgm)")
}

func ListGlobalLog(c *response.Context) {
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
