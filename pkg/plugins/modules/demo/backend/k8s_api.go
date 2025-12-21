package backend

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/plugins"
	"k8s.io/klog/v2"
)

// RegisterClusterRoutes 注册Demo插件的集群相关路由
func RegisterClusterRoutes(api *gin.RouterGroup) {
	g := api.Group("/plugins/demo")
	// 列表
	g.GET("/items", List)

	klog.V(6).Infof("注册插件集群路由: %s", "/items/:id")
}

// List 返回演示列表数据
// 方法内进行角色校验，仅允许“user”角色访问（平台管理员通行）
func List(c *gin.Context) {
	// 方法内角色校验
	ok, err := plugins.EnsureUserIsLogined(c)
	if !ok {
		amis.WriteJsonError(c, err)
		return
	}
	klog.V(6).Infof("获取演示列表")

	params := dao.BuildParams(c)
	m := &Item{}
	items, total, err := dao.GenericQuery(params, m)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}
