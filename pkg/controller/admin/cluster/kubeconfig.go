package cluster

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/service"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

func (a *AdminClusterController) SaveKubeConfig(c *gin.Context) {

	params := dao.BuildParams(c)
	m := models.KubeConfig{}
	err := c.ShouldBindJSON(&m)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	if m.DisplayName == "" {
		m.DisplayName = m.Cluster
	}
	// 因先删除再创建
	// 可能会更新地址的kubeconfig

	config, err := clientcmd.Load([]byte(m.Content))
	if err != nil {
		klog.V(6).Infof("解析 集群 [%s]失败: %v", m.Server, err)
		return
	}
	index := 0
	total := len(config.Contexts)
	for contextName, _ := range config.Contexts {
		index += 1
		context := config.Contexts[contextName]
		cluster := config.Clusters[context.Cluster]

		kc := &models.KubeConfig{
			Cluster:   context.Cluster,
			Server:    cluster.Server,
			User:      context.AuthInfo,
			Namespace: context.Namespace,
			Content:   m.Content,
		}

		kc.DisplayName = m.DisplayName
		// 大于1个，则名称加序列号
		if total != 1 {
			kc.DisplayName = fmt.Sprintf("%s-%d", m.DisplayName, index)
		}

		if list, _, err := kc.List(params); err == nil && list != nil {
			for _, item := range list {
				_ = kc.Delete(params, fmt.Sprintf("%d", item.ID))
			}
		}

		err = kc.Save(params)
		if err != nil {
			klog.V(6).Infof("保存 集群 [%s]失败: %v", m.Server, err)
			amis.WriteJsonError(c, err)
			return
		}

	}

	// 执行一下扫描
	service.ClusterService().ScanClustersInDB()
	// 初始化本项目中的回调
	amis.WriteJsonOK(c)
}
func (a *AdminClusterController) RemoveKubeConfig(c *gin.Context) {

	params := dao.BuildParams(c)
	m := models.KubeConfig{}
	err := c.ShouldBindJSON(&m)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	if list, _, err := m.List(params); err == nil && list != nil {
		for _, item := range list {
			_ = m.Delete(params, fmt.Sprintf("%d", item.ID))
		}
	}

	// 执行一下扫描
	service.ClusterService().ScanClustersInDB()

	amis.WriteJsonOK(c)
}
