package ns

import (
	"context"
	"strings"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

func OptionList(c *gin.Context) {
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 处理集群中有限制namespace的情况
	if list, ok := handleRestrictedNamespace(selectedCluster); ok {
		sortAndRespond(c, list)
		return
	}

	// 处理平台管理员的情况
	if list, ok := handlePlatformAdmin(c, selectedCluster); ok {
		sortAndRespond(c, list)
		return
	}

	// 处理普通用户的情况
	if list, ok := handleNormalUser(c, selectedCluster); ok {
		sortAndRespond(c, list)
		return
	}

}

func handleRestrictedNamespace(selectedCluster string) ([]map[string]string, bool) {
	cluster := service.ClusterService().GetClusterByID(selectedCluster)
	if cluster != nil && cluster.Namespace != "" {
		list := []map[string]string{{
			"label": cluster.Namespace,
			"value": cluster.Namespace,
		}}
		return list, true
	}
	return nil, false
}

func handlePlatformAdmin(c *gin.Context, selectedCluster string) ([]map[string]string, bool) {
	if amis.IsCurrentUserPlatformAdmin(c) {
		ctx := amis.GetContextWithUser(c)
		nsList, err := getClusterNsList(ctx, selectedCluster)
		if err != nil {
			return make([]map[string]string, 0), true
		}
		return nsList, true
	}
	return nil, false
}

func handleNormalUser(c *gin.Context, selectedCluster string) ([]map[string]string, bool) {
	user, _ := amis.GetLoginUser(c)
	clusterUserRoles, err := service.UserService().GetClusters(user)
	if err != nil {
		return make([]map[string]string, 0), true
	}
	if clusterUserRoles == nil {
		return nil, false
	}

	// 筛选带有ns的授权列表
	clusterUserRoles = slice.Filter(clusterUserRoles, func(index int, item *models.ClusterUserRole) bool {
		return item.Namespaces != "" && item.Cluster == selectedCluster
	})

	if len(clusterUserRoles) > 0 {
		// 处理有限制namespace的用户
		return handleUserWithNamespaceRestriction(clusterUserRoles)
	}

	// 处理没有限制namespace的用户
	ctx := amis.GetContextWithUser(c)
	return handleUserWithoutNamespaceRestriction(ctx, selectedCluster)
}

func handleUserWithNamespaceRestriction(roles []*models.ClusterUserRole) ([]map[string]string, bool) {
	var list []map[string]string
	for _, item := range roles {
		if item.Namespaces != "" {
			ns := strings.Split(item.Namespaces, ",")
			for _, n := range ns {
				list = append(list, map[string]string{
					"label": n,
					"value": n,
				})
			}
		}
	}
	return list, true
}

func handleUserWithoutNamespaceRestriction(ctx context.Context, selectedCluster string) ([]map[string]string, bool) {
	nsList, err := getClusterNsList(ctx, selectedCluster)
	if err != nil {
		return make([]map[string]string, 0), true
	}
	return nsList, true
}

func handleBlacklist(c *gin.Context, selectedCluster string, list []map[string]string) ([]map[string]string, bool) {
	user, _ := amis.GetLoginUser(c)
	clusterUserRoles, err := service.UserService().GetClusters(user)
	if err != nil {
		return list, true
	}
	// 筛选带有黑名单 ns的授权列表
	clusterUserRoles = slice.Filter(clusterUserRoles, func(index int, item *models.ClusterUserRole) bool {
		return item.BlacklistNamespaces != "" && item.Cluster == selectedCluster
	})

	// 如果没有黑名单配置，直接返回原列表
	if len(clusterUserRoles) == 0 {
		return list, true
	}

	// 获取所有黑名单namespace
	blacklistNs := make(map[string]bool)
	for _, role := range clusterUserRoles {
		namespaces := strings.Split(role.BlacklistNamespaces, ",")
		for _, ns := range namespaces {
			if ns = strings.TrimSpace(ns); ns != "" {
				blacklistNs[ns] = true
			}
		}
	}

	// 从列表中剔除黑名单namespace
	result := slice.Filter(list, func(index int, item map[string]string) bool {
		ns := item["value"]
		return !blacklistNs[ns]
	})

	return result, true
}

// sortAndRespond 对命名空间列表进行去重、排序并返回响应
// 每个list item中都有一个map，key为label和value，value为命名空间名称
func sortAndRespond(c *gin.Context, list []map[string]string) {
	// 使用map进行去重
	uniqMap := make(map[string]map[string]string)
	for _, item := range list {
		uniqMap[item["value"]] = item
	}

	// 转换回切片
	list = make([]map[string]string, 0, len(uniqMap))
	for _, item := range uniqMap {
		list = append(list, item)
	}
	klog.V(6).Infof("ns list: %v", list)

	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	// 剔除黑名单Namespace
	list, _ = handleBlacklist(c, selectedCluster, list)

	slice.SortBy(list, func(a, b map[string]string) bool {
		return a["label"] < b["label"]
	})
	amis.WriteJsonData(c, gin.H{
		"options": list,
	})
}

func getClusterNsList(ctx context.Context, selectedCluster string) ([]map[string]string, error) {
	// 那么读取集群中的ns
	var ns []v1.Namespace
	err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Namespace{}).List(&ns).Error
	if err != nil {
		return nil, err
	}
	var list []map[string]string
	for _, n := range ns {
		list = append(list, map[string]string{
			"label": n.Name,
			"value": n.Name,
		})
	}
	slice.SortBy(list, func(a, b map[string]string) bool {
		return a["label"] < b["label"]
	})
	return list, nil
}
