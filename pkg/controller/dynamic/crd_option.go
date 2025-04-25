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

func GroupOptionList(c *gin.Context) {
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

func KindOptionList(c *gin.Context) {
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

// Deprecated: 废弃，请不要使用
type CrdTree struct {
	Label    string     `json:"label,omitempty"`
	Value    string     `json:"value,omitempty"`
	Group    string     `json:"group,omitempty"`
	Kind     string     `json:"kind,omitempty"`
	Version  string     `json:"version,omitempty"`
	Children []*CrdTree `json:"children"`
}

// Deprecated: 废弃，请不要使用
func CrdGuidTreeThree(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	list, err := getCrdList(ctx, selectedCluster)

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	//
	var crdTreeList []*CrdTree

	for _, item := range list {
		group, found, err := unstructured.NestedString(item.Object, "spec", "group")
		if err != nil || !found {
			continue
		}
		groupNode := &CrdTree{
			Label:    group,
			Value:    group,
			Group:    group,
			Children: make([]*CrdTree, 0),
		}
		crdTreeList = append(crdTreeList, groupNode)

		kind, found, err := unstructured.NestedString(item.Object, "spec", "names", "kind")
		if err != nil || !found {
			continue
		}
		kindNode := &CrdTree{
			Label:    kind,
			Value:    kind,
			Group:    group,
			Kind:     kind,
			Children: make([]*CrdTree, 0),
		}
		groupNode.Children = append(
			groupNode.Children, kindNode)

		// 获取 spec.versions 数组
		versions, found, err := unstructured.NestedSlice(item.Object, "spec", "versions")
		if err != nil || !found {
			continue
		}

		// 提取每个版本的 name
		for _, version := range versions {
			versionMap, ok := version.(map[string]interface{})
			if !ok {
				continue
			}
			version, found, err := unstructured.NestedString(versionMap, "name")
			if err != nil || !found {
				continue
			}

			versionNode := &CrdTree{
				Label:   version,
				Value:   version,
				Group:   group,
				Kind:    kind,
				Version: version,
			}

			kindNode.Children = append(kindNode.Children, versionNode)

		}

	}

	amis.WriteJsonData(c, gin.H{
		"options": crdTreeList,
	})
}

// Deprecated: 废弃，请不要使用
func CrdGuidTree(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	groups := getCrdGroupList(ctx, selectedCluster)
	var crdTreeList []*CrdTree
	for _, group := range groups {
		node := &CrdTree{
			Label:    group,
			Value:    group,
			Group:    group,
			Children: make([]*CrdTree, 0),
		}
		crdTreeList = append(crdTreeList, node)
		kinds := getCrdKindListByGroup(ctx, selectedCluster, group)
		for _, kind := range kinds {
			kindNode := &CrdTree{
				Label:    kind,
				Value:    kind,
				Group:    group,
				Kind:     kind,
				Children: make([]*CrdTree, 0),
			}
			node.Children = append(node.Children, kindNode)
		}
	}
	amis.WriteJsonData(c, gin.H{
		"options": crdTreeList,
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

func getCrdList(ctx context.Context, selectedCluster string) ([]unstructured.Unstructured, error) {
	var list []unstructured.Unstructured
	err := kom.Cluster(selectedCluster).WithContext(ctx).GVK(
		"apiextensions.k8s.io",
		"v1",
		"CustomResourceDefinition").
		WithCache(time.Second * 30).
		List(&list).Error
	return list, err
}
func getCrdKindListByGroup(ctx context.Context, selectedCluster string, group string) []string {
	var list []unstructured.Unstructured
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
