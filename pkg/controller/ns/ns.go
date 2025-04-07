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
)

func OptionList(c *gin.Context) {
	selectedCluster := amis.GetSelectedCluster(c)

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
	_, _, clusterUserRoles := amis.GetLoginUserWithClusterRoles(c)
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

func sortAndRespond(c *gin.Context, list []map[string]string) {

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
