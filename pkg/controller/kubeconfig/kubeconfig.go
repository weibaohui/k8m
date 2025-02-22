package kubeconfig

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"gorm.io/gorm"
)

func Save(c *gin.Context) {
	params := dao.BuildParams(c)
	m := &models.KubeConfig{}
	err := c.ShouldBindJSON(&m)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 因先删除再创建
	// 可能会更新地址的kubeconfig

	kc := &models.KubeConfig{}
	kc.Cluster = m.Cluster
	kc.Server = m.Server
	kc.User = m.User
	list, _, err := kc.List(params, func(d *gorm.DB) *gorm.DB {
		return d.Where("cluster = ?", m.Cluster).
			Where("server = ?", m.Server).
			Where("user = ?", m.User)
	})
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	for _, config := range list {
		_ = kc.Delete(params, fmt.Sprintf("%d", config.ID))
	}

	err = m.Save(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, gin.H{
		"id": m.ID,
	})
}
