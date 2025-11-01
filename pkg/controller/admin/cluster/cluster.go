package cluster

import (
	"errors"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/service"
	"gorm.io/gorm"
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
	admin.POST("/cluster/aws/save", ctrl.SaveAWSEKSCluster)
	admin.POST("/cluster/token/save", ctrl.SaveTokenCluster)
	admin.GET("/cluster/config/:id", ctrl.GetClusterConfig)
	admin.POST("/cluster/config/save", ctrl.SaveClusterConfig)
}
func RegisterUserClusterRoutes(mgm *gin.RouterGroup) {
	ctrl := &Controller{}
	// 前端用户点击重连接按钮
	mgm.POST("/cluster/:cluster/reconnect", ctrl.Reconnect)
}

// @Summary 获取文件类型的集群选项
// @Description 获取所有已发现集群的kubeconfig文件名列表，用于下拉选项
// @Security BearerAuth
// @Success 200 {object} string
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
	var options []map[string]any
	for _, fn := range fileNames {
		options = append(options, map[string]any{
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
// @Success 200 {object} string
// @Router /admin/cluster/scan [post]
func (a *Controller) Scan(c *gin.Context) {
	service.ClusterService().Scan()
	amis.WriteJsonData(c, "ok")
}

// @Summary 重新连接集群
// @Description 重新连接一个已断开的集群
// @Security BearerAuth
// @Param cluster path string true "Base64编码的集群ID"
// @Success 200 {object} string "已执行，请稍后刷新"
// @Router /mgm/cluster/{cluster}/reconnect [post]
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
// @Success 200 {object} string "已执行，请稍后刷新"
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

// GetClusterConfig 获取集群配置参数
// @Summary 获取集群配置参数
// @Description 根据集群ID获取kom相关配置参数
// @Tags cluster
// @Accept json
// @Produce json
// @Param id path string true "集群ID"
// @Security BearerAuth
// @Success 200 {object} models.KubeConfig
// @Router /admin/cluster/config/{id} [get]
func (a *Controller) GetClusterConfig(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		amis.WriteJsonError(c, errors.New("集群ID不能为空"))
		return
	}

	params := dao.BuildParams(c)
	kubeConfig := &models.KubeConfig{}

	// 根据ID查询集群配置
	config, err := kubeConfig.GetOne(params, func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	})
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	//对零值进行默认值填充
	if config.Timeout == 0 {
		config.Timeout = 30
	}
	if config.QPS == 0 {
		config.QPS = 100
	}
	if config.Burst == 0 {
		config.Burst = 200
	}
	// 只返回配置相关的字段
	configData := map[string]any{
		"id":       config.ID,
		"proxyURL": config.ProxyURL,
		"timeout":  config.Timeout,
		"qps":      config.QPS,
		"burst":    config.Burst,
	}

	amis.WriteJsonData(c, configData)
}

// SaveClusterConfig 保存集群配置参数
// @Summary 保存集群配置参数
// @Description 保存集群的kom相关配置参数
// @Tags cluster
// @Accept json
// @Produce json
// @Param config body object true "集群配置参数"
// @Security BearerAuth
// @Success 200 {object} string
// @Router /admin/cluster/config/save [post]
func (a *Controller) SaveClusterConfig(c *gin.Context) {
	var configData struct {
		ID       uint    `json:"id" binding:"required"`
		ProxyURL string  `json:"proxyURL"`
		Timeout  int     `json:"timeout"`
		QPS      float32 `json:"qps"`
		Burst    int     `json:"burst"`
	}

	if err := c.ShouldBindJSON(&configData); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	params := dao.BuildParams(c)
	kubeConfig := &models.KubeConfig{}

	// 根据ID查询现有配置
	config, err := kubeConfig.GetOne(params, func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", configData.ID)
	})
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 更新配置字段
	config.ProxyURL = configData.ProxyURL
	config.Timeout = configData.Timeout
	config.QPS = configData.QPS
	config.Burst = configData.Burst

	// 保存更新
	if err := config.Save(params); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOKMsg(c, "配置保存成功")
}
