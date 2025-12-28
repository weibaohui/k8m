package route

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/plugins/modules/webhook/admin"
	"k8s.io/klog/v2"
)

func RegisterPluginAdminRoutes(arg *gin.RouterGroup) {
	g := arg.Group("/plugins/webhook")
	ctrl := &admin.Controller{}

	g.GET("/list", ctrl.WebhookList)
	g.POST("/delete/:ids", ctrl.WebhookDelete)
	g.POST("/save", ctrl.WebhookSave)
	g.POST("/id/:id/test", ctrl.WebhookTest)
	g.GET("/option_list", ctrl.WebhookOptionList)

	g.GET("/records", ctrl.WebhookRecordList)
	g.GET("/records/:id", ctrl.WebhookRecordDetail)
	g.GET("/records/statistics", ctrl.WebhookRecordStatistics)

	klog.V(6).Infof("注册webhook插件管理路由(admin)")
}

