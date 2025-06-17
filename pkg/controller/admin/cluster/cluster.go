package cluster

import (
	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/service"
)

type AdminClusterController struct {
}

func RegisterAdminClusterRoutes(admin *gin.RouterGroup) {
	ctrl := &AdminClusterController{}
	admin.POST("/cluster/scan", ctrl.Scan)
	admin.GET("/cluster/file/option_list", ctrl.FileOptionList)
	admin.POST("/cluster/kubeconfig/save", ctrl.SaveKubeConfig)
	admin.POST("/cluster/kubeconfig/remove", ctrl.RemoveKubeConfig)
	admin.POST("/cluster/:cluster/disconnect", ctrl.Disconnect)
}
func RegisterUserClusterRoutes(mgm *gin.RouterGroup) {
	ctrl := &AdminClusterController{}
	//前端用户点击重连接按钮
	mgm.POST("/cluster/:cluster/reconnect", ctrl.Reconnect)
}

// @Summary 获取文件类型的集群选项
// @Security BearerAuth
// @Success 200   {object} string
// @Router /admin/cluster/file/option_list [get]
func (a *AdminClusterController) FileOptionList(c *gin.Context) {
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

func (a *AdminClusterController) Scan(c *gin.Context) {
	service.ClusterService().Scan()
	amis.WriteJsonData(c, "ok")
}

func (a *AdminClusterController) Reconnect(c *gin.Context) {
	clusterBase64 := c.Param("cluster")
	clusterID, err := utils.DecodeBase64(clusterBase64)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	go service.ClusterService().Connect(clusterID)
	amis.WriteJsonOKMsg(c, "已执行，请稍后刷新")
}
func (a *AdminClusterController) Disconnect(c *gin.Context) {
	clusterBase64 := c.Param("cluster")
	clusterID, err := utils.DecodeBase64(clusterBase64)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	service.ClusterService().Disconnect(clusterID)
	amis.WriteJsonOKMsg(c, "已执行，请稍后刷新")
}
