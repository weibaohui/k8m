package cluster

import (
	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/service"
)

type Controller struct {
}

func RegisterAdminClusterRoutes(admin *gin.RouterGroup) {
	ctrl := &Controller{}
	admin.POST("/cluster/scan", ctrl.Scan)
	admin.GET("/cluster/file/option_list", ctrl.FileOptionList)
	admin.POST("/cluster/kubeconfig/save", ctrl.SaveKubeConfig)
	admin.POST("/cluster/kubeconfig/remove", ctrl.RemoveKubeConfig)
	admin.POST("/cluster/:cluster/disconnect", ctrl.Disconnect)
}
func RegisterUserClusterRoutes(mgm *gin.RouterGroup) {
	ctrl := &Controller{}
	// 前端用户点击重连接按钮
	mgm.POST("/cluster/:cluster/reconnect", ctrl.Reconnect)
}

// @Summary 获取文件类型的集群选项
// @Description 获取所有已发现集群的kubeconfig文件名列表，用于下拉选项
// @Security BearerAuth
// @Success 200 {object} amis.Response
// @Router /admin/cluster/file/option_list [get]
func (a *Controller) FileOptionList(c *gin.Context) {
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

// @Summary 扫描集群
// @Description 扫描本地Kubeconfig文件目录以发现新的集群
// @Security BearerAuth
// @Success 200 {object} amis.Response "ok"
// @Router /admin/cluster/scan [post]
func (a *Controller) Scan(c *gin.Context) {
	service.ClusterService().Scan()
	amis.WriteJsonData(c, "ok")
}

// @Summary 重新连接集群
// @Description 重新连接一个已断开的集群
// @Security BearerAuth
// @Param cluster path string true "Base64编码的集群ID"
// @Success 200 {object} amis.Response "已执行，请稍后刷新"
// @Router /admin/cluster/{cluster}/reconnect [post]
func (a *Controller) Reconnect(c *gin.Context) {
	clusterBase64 := c.Param("cluster")
	clusterID, err := utils.DecodeBase64(clusterBase64)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	go service.ClusterService().Connect(clusterID)
	amis.WriteJsonOKMsg(c, "已执行，请稍后刷新")
}

// @Summary 断开集群连接
// @Description 断开一个正在运行的集群的连接
// @Security BearerAuth
// @Param cluster path string true "Base64编码的集群ID"
// @Success 200 {object} amis.Response "已执行，请稍后刷新"
// @Router /admin/cluster/{cluster}/disconnect [post]
func (a *Controller) Disconnect(c *gin.Context) {
	clusterBase64 := c.Param("cluster")
	clusterID, err := utils.DecodeBase64(clusterBase64)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	service.ClusterService().Disconnect(clusterID)
	amis.WriteJsonOKMsg(c, "已执行，请稍后刷新")
}
