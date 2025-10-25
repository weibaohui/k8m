package comm

import (
	"context"
	"fmt"
	"strings"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/service"
	"k8s.io/klog/v2"
)

// CheckPermissionLogic
// return err
func CheckPermissionLogic(ctx context.Context, cluster string, nsList []string, ns, name, action string) error {

	// 内部监听增加一个认证机制，不用做权限校验
	// 比如node watch
	if constants.RolePlatformAdmin == ctx.Value(constants.RolePlatformAdmin) {
		return nil
	}

	username := fmt.Sprintf("%s", ctx.Value(constants.JwtUserName))

	if username == "" {
		return fmt.Errorf("用户为空%v，默认阻止", nil)
	}
	var err error
	// 先看是不是平台管理员
	if service.UserService().IsUserPlatformAdmin(username) {
		// 平台管理员，可以执行任何操作
		return nil
	}

	clusterUserRoles, err := service.UserService().GetClusters(username)

	if err != nil || len(clusterUserRoles) == 0 {
		// 没有集群权限，报错
		return fmt.Errorf("用户[%s]获取集群授权错误，默认阻止", username)
	}

	if clusterUserRoles != nil && len(clusterUserRoles) == 0 {
		return fmt.Errorf("用户[%s]没有集群授权", username)
	}
	if _, ok := slice.FindBy(clusterUserRoles, func(index int, item *models.ClusterUserRole) bool {
		return item.Cluster == cluster
	}); !ok {
		return fmt.Errorf("用户[%s]没有集群[%s]访问权限", username, cluster)
	}

	// 下面都是有集群的访问权限的情况，需要进一步区分是什么类型的操作。
	// 以及是否有namespace的权限

	// 操作对象为带namespace的情况，那么需要进一步看用户是否有该ns的权限
	// 如果遍历权限表格，该集群对应的ns为空，说明不限制，如果ns不为空（是一个数组），说明限制了ns，就需要相等才能执行。
	// 先判断是否有集群、对应的操作权限，再看是否有命名空间的
	switch action {
	case "exec":
		manageClusters := slice.Filter(clusterUserRoles, func(index int, item *models.ClusterUserRole) bool {
			return item.Cluster == cluster && item.Role == constants.RoleClusterAdmin
		})
		// 对于给定的cluster这个集群，
		// 没有集群管理员权限，那么就需要进行Exec权限判断了
		// 有集群管理权限，就能有在pod中执行命令的权限

		if len(manageClusters) == 0 {
			// 如果没有集群管理员权限，那么就必须要有集群只读+exec权限
			rdOnlyClusters := slice.Filter(clusterUserRoles, func(index int, item *models.ClusterUserRole) bool {
				return item.Cluster == cluster && item.Role == constants.RoleClusterReadonly
			})
			if len(rdOnlyClusters) == 0 {
				return fmt.Errorf("用户[%s]没有集群[%s] 只读权限", username, cluster)
			}
			execClusters := slice.Filter(clusterUserRoles, func(index int, item *models.ClusterUserRole) bool {
				return item.Cluster == cluster && item.Role == constants.RoleClusterPodExec
			})
			if len(execClusters) == 0 {
				return fmt.Errorf("用户[%s]没有集群[%s] Exec权限", username, cluster)
			}
			if len(nsList) > 0 {
				// 具备只读+Exec权限了，那么继续看是否有该ns的权限.
				// ns为空，或者ns列表中含有当前ns，那么就允许执行。

				// 首先看是否在ns黑名单中，在的话阻止
				execClustersWithBNs := slice.Filter(execClusters, func(index int, item *models.ClusterUserRole) bool {
					return item.BlacklistNamespaces != "" && utils.AnyIn(nsList, strings.Split(item.BlacklistNamespaces, ","))
				})
				if len(execClustersWithBNs) > 0 {
					return fmt.Errorf("用户[%s]没有集群[%s] [%s] Exec权限-进入命名空间黑名单", username, cluster, strings.Join(nsList, ","))
				}
				execClustersWithNs := slice.Filter(execClusters, func(index int, item *models.ClusterUserRole) bool {
					return item.Namespaces == "" || utils.AllIn(nsList, strings.Split(item.Namespaces, ","))
				})
				if len(execClustersWithNs) == 0 {
					return fmt.Errorf("用户[%s]没有集群[%s] [%s] Exec权限-不在命名空间白名单", username, cluster, strings.Join(nsList, ","))
				}
			}
		} else {
			//  有集群权限，那么继续看是否有该ns的权限.
			if len(nsList) > 0 {

				// 首先看是否在ns黑名单中，在的话阻止
				execClustersWithBNs := slice.Filter(manageClusters, func(index int, item *models.ClusterUserRole) bool {
					return item.BlacklistNamespaces != "" && utils.AnyIn(nsList, strings.Split(item.BlacklistNamespaces, ","))
				})
				if len(execClustersWithBNs) > 0 {
					return fmt.Errorf("用户[%s]没有集群[%s] [%s] Exec权限-进入命名空间黑名单", username, cluster, strings.Join(nsList, ","))
				}

				// ns为空，或者ns列表中含有当前ns，那么就允许执行。
				execClustersWithNs := slice.Filter(manageClusters, func(index int, item *models.ClusterUserRole) bool {
					return item.Namespaces == "" || utils.AllIn(nsList, strings.Split(item.Namespaces, ","))
				})
				if len(execClustersWithNs) == 0 {
					return fmt.Errorf("用户[%s]没有集群[%s] [%s] Exec权限-不在命名空间白名单", username, cluster, strings.Join(nsList, ","))
				}
			}
		}

	case "delete", "update", "patch", "create":
		changeClusters := slice.Filter(clusterUserRoles, func(index int, item *models.ClusterUserRole) bool {
			return item.Cluster == cluster && item.Role == constants.RoleClusterAdmin
		})
		if len(changeClusters) == 0 {
			return fmt.Errorf("用户[%s]没有集群[%s] 操作权限", username, cluster)
		}
		if len(nsList) > 0 {
			// 具备操作权限了，那么继续看是否有该ns的权限.
			// ns为空，或者ns列表中含有当前ns，那么就允许执行。

			// 首先看是否在ns黑名单中，在的话阻止
			execClustersWithBNs := slice.Filter(changeClusters, func(index int, item *models.ClusterUserRole) bool {
				return item.BlacklistNamespaces != "" && utils.AnyIn(nsList, strings.Split(item.BlacklistNamespaces, ","))
			})
			if len(execClustersWithBNs) > 0 {
				return fmt.Errorf("用户[%s]没有集群[%s] [%s] 操作权限-进入命名空间黑名单", username, cluster, strings.Join(nsList, ","))
			}

			changeClustersWithNs := slice.Filter(changeClusters, func(index int, item *models.ClusterUserRole) bool {
				return item.Namespaces == "" || utils.AllIn(nsList, strings.Split(item.Namespaces, ","))
			})
			if len(changeClustersWithNs) == 0 {
				return fmt.Errorf("用户[%s]没有集群[%s] [%s] 操作权限-不在命名空间白名单", username, cluster, strings.Join(nsList, ","))
			}
		}
	default:
		// 读取类的权限，走到这的可能是集群管理员，或者集群只读，exec在前面拦截了。

		// 必须得有集群只读或者集群管理员权限
		readClusters := slice.Filter(clusterUserRoles, func(index int, item *models.ClusterUserRole) bool {
			return item.Cluster == cluster && (item.Role == constants.RoleClusterReadonly || item.Role == constants.RoleClusterAdmin)
		})
		if len(readClusters) == 0 {
			return fmt.Errorf("用户[%s]没有集群[%s] 读取/管理员 权限", username, cluster)
		}
		if len(nsList) > 0 {
			// 具备操作权限了，那么继续看是否有该ns的权限.
			// ns为空，或者ns列表中含有当前ns，那么就允许执行。

			// 首先看是否在ns黑名单中，在的话阻止
			execClustersWithBNs := slice.Filter(readClusters, func(index int, item *models.ClusterUserRole) bool {
				return item.BlacklistNamespaces != "" && utils.AnyIn(nsList, strings.Split(item.BlacklistNamespaces, ","))
			})
			if len(execClustersWithBNs) > 0 {
				return fmt.Errorf("用户[%s]没有集群[%s] [%s] 读取权限-进入命名空间黑名单", username, cluster, strings.Join(nsList, ","))
			}

			readClustersWithNs := slice.Filter(readClusters, func(index int, item *models.ClusterUserRole) bool {
				return item.Namespaces == "" || utils.AllIn(nsList, strings.Split(item.Namespaces, ","))
			})
			if len(readClustersWithNs) == 0 {
				return fmt.Errorf("用户[%s]没有集群[%s] [%s] 读取权限-不在命名空间白名单", username, cluster, strings.Join(nsList, ","))
			}
		}
	}
	klog.V(6).Infof("cb: cluster= %s,user= %s,  operation=%s,  resource=[%s/%s] ",
		cluster, username, action, ns, name)
	return err
}
