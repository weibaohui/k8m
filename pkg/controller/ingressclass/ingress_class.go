package ingressclass

import (
	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/networking/v1"
)

type Controller struct{}

func RegisterRoutes(api *gin.RouterGroup) {
	ctrl := &Controller{}
	api.POST("/ingress_class/set_default/name/:name", ctrl.SetDefault)
	api.GET("/ingress_class/option_list", ctrl.OptionList)
}

// @Summary 设置默认的 IngressClass
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param name path string true "IngressClass 名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/ingress_class/set_default/name/{name} [post]
func (cc *Controller) SetDefault(c *gin.Context) {
	name := c.Param("name")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	err = kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.IngressClass{}).Name(name).
		Ctl().IngressClass().SetDefault()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

// @Summary 获取 IngressClass 选项列表
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/ingress_class/option_list [get]
func (cc *Controller) OptionList(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var list []v1.IngressClass
	err = kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.IngressClass{}).List(&list).Error
	if err != nil {
		amis.WriteJsonData(c, gin.H{
			"options": make([]map[string]string, 0),
		})
		return
	}
	var names []map[string]string
	for _, n := range list {
		names = append(names, map[string]string{
			"label": n.Name,
			"value": n.Name,
		})
	}
	slice.SortBy(names, func(a, b map[string]string) bool {
		return a["label"] < b["label"]
	})
	amis.WriteJsonData(c, gin.H{
		"options": names,
	})
}
