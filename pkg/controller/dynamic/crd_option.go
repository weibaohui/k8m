package dynamic

import (
	"context"
	"time"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/response"
	"github.com/weibaohui/kom/kom"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type CRDController struct{}

// 从 gin 切换到 chi，使用 chi.Router 替代 gin.RouterGroup
func RegisterCRDRoutes(r chi.Router) {
	ctrl := &CRDController{}
	r.Get("/crd/group/option_list", response.Adapter(ctrl.GroupOptionList))
	r.Get("/crd/kind/option_list", response.Adapter(ctrl.KindOptionList))
	r.Get("/crd/status", response.Adapter(ctrl.CRDStatus))
}

// @Summary 获取CRD组选项列表
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/crd/group/option_list [get]
func (cc *CRDController) GroupOptionList(c *response.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	groups := getCrdGroupList(ctx, selectedCluster)
	var options []map[string]string
	for _, n := range groups {
		options = append(options, map[string]string{
			"label": n,
			"value": n,
		})
	}

	amis.WriteJsonData(c, response.H{
		"options": options,
	})
}

// @Summary 获取指定组的CRD类型选项列表
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param spec[group] query string true "CRD组名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/crd/kind/option_list [get]
func (cc *CRDController) KindOptionList(c *response.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	g := c.Query("spec[group]")
	if g == "" {
		// 还没选group
		amis.WriteJsonData(c, response.H{
			"options": make([]map[string]string, 0),
		})
		return
	}

	kinds := getCrdKindListByGroup(ctx, selectedCluster, g)

	var options []map[string]string
	for _, n := range kinds {
		options = append(options, map[string]string{
			"label": n,
			"value": n,
		})
	}

	amis.WriteJsonData(c, response.H{
		"options": options,
	})
}

func getCrdGroupList(ctx context.Context, selectedCluster string) []string {
	list, err := getCrdList(ctx, selectedCluster)
	if err != nil {
		return make([]string, 0)
	}

	var groups []string
	for _, item := range list {
		group, found, err := unstructured.NestedString(item.Object, "spec", "group")
		if err != nil || !found {
			continue
		}

		groups = append(groups, group)
	}
	groups = slice.Unique(groups)
	slice.Sort(groups, "asc")
	return groups
}

func getCrdList(ctx context.Context, selectedCluster string) ([]*unstructured.Unstructured, error) {
	var list []*unstructured.Unstructured
	err := kom.Cluster(selectedCluster).WithContext(ctx).GVK(
		"apiextensions.k8s.io",
		"v1",
		"CustomResourceDefinition").
		WithCache(time.Second * 30).
		List(&list).Error
	return list, err
}
func getCrdKindListByGroup(ctx context.Context, selectedCluster string, group string) []string {
	var list []*unstructured.Unstructured
	err := kom.Cluster(selectedCluster).WithContext(ctx).GVK(
		"apiextensions.k8s.io",
		"v1",
		"CustomResourceDefinition").
		Where("`spec.group`=?", group).
		WithCache(time.Second * 30).
		List(&list).Error
	if err != nil {
		return make([]string, 0)
	}

	var kinds []string
	for _, item := range list {
		kind, found, err := unstructured.NestedString(item.Object, "spec", "names", "kind")
		if err != nil || !found {
			continue
		}
		kinds = append(kinds, kind)
	}
	kinds = slice.Unique(kinds)
	slice.Sort(kinds, "asc")
	return kinds
}
