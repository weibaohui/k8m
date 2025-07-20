package dynamic

import (
	"context"
	"time"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type CRDController struct{}

func RegisterCRDRoutes(api *gin.RouterGroup) {
	ctrl := &CRDController{}
	api.GET("/crd/group/option_list", ctrl.GroupOptionList)
	api.GET("/crd/kind/option_list", ctrl.KindOptionList)
	api.GET("/crd/status", ctrl.CRDStatus)
}

func (cc *CRDController) GroupOptionList(c *gin.Context) {
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

	amis.WriteJsonData(c, gin.H{
		"options": options,
	})
}

func (cc *CRDController) KindOptionList(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	g := c.Query("spec[group]")
	if g == "" {
		// 还没选group
		amis.WriteJsonData(c, gin.H{
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

	amis.WriteJsonData(c, gin.H{
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
