package storageclass

import (
	"github.com/duke-git/lancet/v2/slice"
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/response"
	"github.com/weibaohui/kom/kom"

	v1 "k8s.io/api/storage/v1"
)

type Controller struct{}

// 从 gin 切换到 chi，使用 chi.Router 替代 gin.RouterGroup
func RegisterRoutes(r chi.Router) {
	ctrl := &Controller{}
	r.Post("/storage_class/set_default/name/{name}", response.Adapter(ctrl.SetDefault))
	r.Get("/storage_class/option_list", response.Adapter(ctrl.OptionList))
}

// SetDefault 设置默认存储类
// @Summary 设置默认存储类
// @Security BearerAuth
// @Param cluster path string true "集群名称"
// @Param name path string true "存储类名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/storage_class/set_default/name/{name} [post]
func (cc *Controller) SetDefault(c *response.Context) {
	name := c.Param("name")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	err = kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.StorageClass{}).Name(name).
		Ctl().StorageClass().SetDefault()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

// OptionList 获取存储类选项列表
// @Summary 获取存储类选项列表
// @Security BearerAuth
// @Param cluster path string true "集群名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/storage_class/option_list [get]
func (cc *Controller) OptionList(c *response.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var list []v1.StorageClass
	err = kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.StorageClass{}).List(&list).Error
	if err != nil {
		amis.WriteJsonData(c, response.H{
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
	amis.WriteJsonData(c, response.H{
		"options": names,
	})
}
