package cb

import (
	"fmt"

	"github.com/weibaohui/k8m/pkg/comm"
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
	ns := stmt.Namespace
	if ns != "" {
		nsList = append(nsList, ns)
	}
	name := stmt.Name
	return comm.CheckPermissionLogic(ctx, cluster, nsList, ns, name, action)
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
