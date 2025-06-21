package cb

import (
	"fmt"
	"strings"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
	"k8s.io/klog/v2"
)

func RegisterDefaultCallbacks(cluster *service.ClusterConfig) func() {
	selectedCluster := service.ClusterService().ClusterID(cluster)

	getCallback := kom.Cluster(selectedCluster).Callback().Get()
	_ = getCallback.Before("*").Register("k8m:get", handleGet)

	describeCallback := kom.Cluster(selectedCluster).Callback().Describe()
	_ = describeCallback.Before("*").Register("k8m:describe", handleDescribe)

	listCallback := kom.Cluster(selectedCluster).Callback().List()
	_ = listCallback.Before("*").Register("k8m:list", handleList)

	logsCallback := kom.Cluster(selectedCluster).Callback().Logs()
	_ = logsCallback.Before("*").Register("k8m:logs", handleLogs)

	deleteCallback := kom.Cluster(selectedCluster).Callback().Delete()
	_ = deleteCallback.Before("*").Register("k8m:delete", handleDelete)

	updateCallback := kom.Cluster(selectedCluster).Callback().Update()
	_ = updateCallback.Before("*").Register("k8m:update", handleUpdate)

	patchCallback := kom.Cluster(selectedCluster).Callback().Patch()
	_ = patchCallback.Before("*").Register("k8m:patch", handlePatch)

	createCallback := kom.Cluster(selectedCluster).Callback().Create()
	_ = createCallback.Before("*").Register("k8m:create", handleCreate)

	execCallback := kom.Cluster(selectedCluster).Callback().Exec()
	_ = execCallback.Before("*").Register("k8m:pod-exec", handleExec)

	streamExecCallback := kom.Cluster(selectedCluster).Callback().StreamExec()
	_ = streamExecCallback.Before("*").Register("k8m:pod-stream-exec", handleExec)
	klog.V(6).Infof("registered callbacks for cluster %s", selectedCluster)
	return nil
}

// handleCommonLogic 根据用户在指定集群上的角色和命名空间权限，校验其是否有执行指定 Kubernetes 操作（如读取、变更、Exec 等）的权限。
// 返回用户名、角色列表和权限校验错误（如无权限则返回错误）。
// handleCommonLogic 根据用户在平台和集群中的角色，校验其对指定 Kubernetes 集群和命名空间的操作权限。
// 平台管理员拥有所有权限，集群管理员拥有全部操作权限，特定操作（如 Exec、只读）需具备对应角色及命名空间权限。
// 若为内部监听（如 node watch），则跳过权限校验。
// 返回用户名、角色列表及权限校验错误（如无权限或未授权）。
//
// 参数：
//
//	k8s: 封装了操作上下文的 Kubectl 实例。
//	action: 待校验的操作类型（如 exec、delete、update、patch、create、读取类操作等）。
//
// 返回：
//
//	用户名、角色列表，以及权限不足或异常时的错误信息。
func handleCommonLogic(k8s *kom.Kubectl, action string) (string, []string, error) {
	stmt := k8s.Statement
	cluster := k8s.ID
	ctx := stmt.Context
	nsList := stmt.NamespaceList
	nsList = append(nsList, stmt.Namespace)

	// 内部监听增加一个认证机制，不用做权限校验
	// 比如node watch
	if constants.RolePlatformAdmin == ctx.Value(constants.RolePlatformAdmin) {
		return constants.RolePlatformAdmin, []string{constants.RolePlatformAdmin}, nil
	}

	username := fmt.Sprintf("%s", ctx.Value(constants.JwtUserName))

	if username == "" {
		return "", nil, fmt.Errorf("用户为空%v，默认阻止", nil)
	}

	roleString := service.UserService().GetPlatformRolesByName(username)

	var err error

	roles, _ := service.UserService().GetClusterRole(cluster, username, roleString)
	// 先看是不是平台管理员
	if slice.Contain(roles, constants.RolePlatformAdmin) {
		// 平台管理员，可以执行任何操作
		return username, roles, nil
	}
	clusterUserRoles, err := service.UserService().GetClusters(username)

	if err != nil || len(clusterUserRoles) == 0 {
		// 没有集群权限，报错
		return "", nil, fmt.Errorf("用户[%s]获取集群授权错误，默认阻止", username)
	}

	if clusterUserRoles != nil && len(clusterUserRoles) == 0 {
		return "", nil, fmt.Errorf("用户[%s]没有集群授权", username)
	}
	if _, ok := slice.FindBy(clusterUserRoles, func(index int, item *models.ClusterUserRole) bool {
		return item.Cluster == cluster
	}); !ok {
		return "", nil, fmt.Errorf("用户[%s]没有集群[%s]访问权限", username, cluster)
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
				return "", nil, fmt.Errorf("用户[%s]没有集群[%s] 只读权限", username, cluster)
			}
			execClusters := slice.Filter(clusterUserRoles, func(index int, item *models.ClusterUserRole) bool {
				return item.Cluster == cluster && item.Role == constants.RoleClusterPodExec
			})
			if len(execClusters) == 0 {
				return "", nil, fmt.Errorf("用户[%s]没有集群[%s] Exec权限", username, cluster)
			}
			if len(nsList) > 0 {
				// 具备只读+Exec权限了，那么继续看是否有该ns的权限.
				// ns为空，或者ns列表中含有当前ns，那么就允许执行。

				// 首先看是否在ns黑名单中，在的话阻止
				execClustersWithBNs := slice.Filter(execClusters, func(index int, item *models.ClusterUserRole) bool {
					return item.BlacklistNamespaces != "" && utils.AnyIn(nsList, strings.Split(item.BlacklistNamespaces, ","))
				})
				if len(execClustersWithBNs) > 0 {
					return "", nil, fmt.Errorf("用户[%s]没有集群[%s] [%s] Exec权限-进入命名空间黑名单", username, cluster, strings.Join(nsList, ","))
				}
				execClustersWithNs := slice.Filter(execClusters, func(index int, item *models.ClusterUserRole) bool {
					return item.Namespaces == "" || utils.AllIn(nsList, strings.Split(item.Namespaces, ","))
				})
				if len(execClustersWithNs) == 0 {
					return "", nil, fmt.Errorf("用户[%s]没有集群[%s] [%s] Exec权限-不在命名空间白名单", username, cluster, strings.Join(nsList, ","))
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
					return "", nil, fmt.Errorf("用户[%s]没有集群[%s] [%s] Exec权限-进入命名空间黑名单", username, cluster, strings.Join(nsList, ","))
				}

				// ns为空，或者ns列表中含有当前ns，那么就允许执行。
				execClustersWithNs := slice.Filter(manageClusters, func(index int, item *models.ClusterUserRole) bool {
					return item.Namespaces == "" || utils.AllIn(nsList, strings.Split(item.Namespaces, ","))
				})
				if len(execClustersWithNs) == 0 {
					return "", nil, fmt.Errorf("用户[%s]没有集群[%s] [%s] Exec权限-不在命名空间白名单", username, cluster, strings.Join(nsList, ","))
				}
			}
		}

	case "delete", "update", "patch", "create":
		changeClusters := slice.Filter(clusterUserRoles, func(index int, item *models.ClusterUserRole) bool {
			return item.Cluster == cluster && item.Role == constants.RoleClusterAdmin
		})
		if len(changeClusters) == 0 {
			return "", nil, fmt.Errorf("用户[%s]没有集群[%s] 操作权限", username, cluster)
		}
		if len(nsList) > 0 {
			// 具备操作权限了，那么继续看是否有该ns的权限.
			// ns为空，或者ns列表中含有当前ns，那么就允许执行。

			// 首先看是否在ns黑名单中，在的话阻止
			execClustersWithBNs := slice.Filter(changeClusters, func(index int, item *models.ClusterUserRole) bool {
				return item.BlacklistNamespaces != "" && utils.AnyIn(nsList, strings.Split(item.BlacklistNamespaces, ","))
			})
			if len(execClustersWithBNs) > 0 {
				return "", nil, fmt.Errorf("用户[%s]没有集群[%s] [%s] 操作权限-进入命名空间黑名单", username, cluster, strings.Join(nsList, ","))
			}

			changeClustersWithNs := slice.Filter(changeClusters, func(index int, item *models.ClusterUserRole) bool {
				return item.Namespaces == "" || utils.AllIn(nsList, strings.Split(item.Namespaces, ","))
			})
			if len(changeClustersWithNs) == 0 {
				return "", nil, fmt.Errorf("用户[%s]没有集群[%s] [%s] 操作权限-不在命名空间白名单", username, cluster, strings.Join(nsList, ","))
			}
		}
	default:
		// 读取类的权限，走到这的可能是集群管理员，或者集群只读，exec在前面拦截了。

		// 必须得有集群只读或者集群管理员权限
		readClusters := slice.Filter(clusterUserRoles, func(index int, item *models.ClusterUserRole) bool {
			return item.Cluster == cluster && (item.Role == constants.RoleClusterReadonly || item.Role == constants.RoleClusterAdmin)
		})
		if len(readClusters) == 0 {
			return "", nil, fmt.Errorf("用户[%s]没有集群[%s] 读取/管理员 权限", username, cluster)
		}
		if len(nsList) > 0 {
			// 具备操作权限了，那么继续看是否有该ns的权限.
			// ns为空，或者ns列表中含有当前ns，那么就允许执行。

			// 首先看是否在ns黑名单中，在的话阻止
			execClustersWithBNs := slice.Filter(readClusters, func(index int, item *models.ClusterUserRole) bool {
				return item.BlacklistNamespaces != "" && utils.AnyIn(nsList, strings.Split(item.BlacklistNamespaces, ","))
			})
			if len(execClustersWithBNs) > 0 {
				return "", nil, fmt.Errorf("用户[%s]没有集群[%s] [%s] 读取权限-进入命名空间黑名单", username, cluster, strings.Join(nsList, ","))
			}

			readClustersWithNs := slice.Filter(readClusters, func(index int, item *models.ClusterUserRole) bool {
				return item.Namespaces == "" || utils.AllIn(nsList, strings.Split(item.Namespaces, ","))
			})
			if len(readClustersWithNs) == 0 {
				return "", nil, fmt.Errorf("用户[%s]没有集群[%s] [%s] 读取权限-不在命名空间白名单", username, cluster, strings.Join(nsList, ","))
			}
		}
	}

	klog.V(6).Infof("cb: cluster= %s,user= %s, role= %s，roleString=%v, operation=%s, gck=[%s], resource=[%s/%s] ",
		cluster, username, roleString, roles, action, stmt.GVK.String(), stmt.Namespace, stmt.Name)
	return username, roles, err
}
func saveLog2DB(k8s *kom.Kubectl, action string, err error) {
	stmt := k8s.Statement
	cluster := k8s.ID
	ctx := stmt.Context
	username := fmt.Sprintf("%s", ctx.Value(constants.JwtUserName))
	roleString := fmt.Sprintf("%s", ctx.Value(constants.JwtUserRole))

	log := models.OperationLog{
		Action:       action,
		Cluster:      cluster,
		Kind:         stmt.GVK.Kind,
		Name:         stmt.Name,
		Namespace:    stmt.Namespace,
		UserName:     username,
		Group:        stmt.GVK.Group,
		Role:         roleString,
		ActionResult: "success",
	}

	if err != nil {
		log.ActionResult = err.Error()
	}

	service.OperationLogService().Add(&log)

}
func handleDelete(k8s *kom.Kubectl) error {
	_, _, err := handleCommonLogic(k8s, "delete")
	saveLog2DB(k8s, "delete", err)
	return err
}

func handleUpdate(k8s *kom.Kubectl) error {
	_, _, err := handleCommonLogic(k8s, "update")
	saveLog2DB(k8s, "update", err)
	return err
}

func handlePatch(k8s *kom.Kubectl) error {
	_, _, err := handleCommonLogic(k8s, "patch")
	saveLog2DB(k8s, "patch", err)
	return err
}

func handleCreate(k8s *kom.Kubectl) error {
	_, _, err := handleCommonLogic(k8s, "create")
	saveLog2DB(k8s, "create", err)
	return err
}
func handleExec(k8s *kom.Kubectl) error {
	_, _, err := handleCommonLogic(k8s, "exec")
	saveLog2DB(k8s, "exec", err)
	return err
}

func handleList(k8s *kom.Kubectl) error {
	_, _, err := handleCommonLogic(k8s, "list")
	return err
}
func handleDescribe(k8s *kom.Kubectl) error {
	_, _, err := handleCommonLogic(k8s, "describe")
	return err
}
func handleLogs(k8s *kom.Kubectl) error {
	_, _, err := handleCommonLogic(k8s, "logs")
	return err
}
func handleGet(k8s *kom.Kubectl) error {
	_, _, err := handleCommonLogic(k8s, "get")
	return err
}
