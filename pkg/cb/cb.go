package cb

import (
	"fmt"
	"time"

	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
	"k8s.io/klog/v2"
)

func RegisterDefaultCallbacks(cluster *service.ClusterConfig) func() {
	selectedCluster := service.ClusterService().ClusterID(cluster)

	deleteCallback := kom.Cluster(selectedCluster).Callback().Delete()
	_ = deleteCallback.Before("kom:delete").Register("k8m:delete", handleDelete)
	updateCallback := kom.Cluster(selectedCluster).Callback().Update()
	_ = updateCallback.Before("kom:update").Register("k8m:update", handleUpdate)
	patchCallback := kom.Cluster(selectedCluster).Callback().Patch()
	_ = patchCallback.Before("kom:patch").Register("k8m:patch", handlePatch)
	createCallback := kom.Cluster(selectedCluster).Callback().Create()
	_ = createCallback.Before("kom:create").Register("k8m:create", handleCreate)
	return nil
}

func handleCommonLogic(k8s *kom.Kubectl, action string) (string, string, error) {
	stmt := k8s.Statement
	cluster := k8s.ID
	ctx := stmt.Context
	username := fmt.Sprintf("%s", ctx.Value(constants.JwtUserName))
	role := fmt.Sprintf("%s", ctx.Value(constants.JwtUserRole))
	klog.V(6).Infof("cb: cluster= %s,user= %s, role= %s, operation=%s, gck=[%s], resource=[%s/%s] ",
		cluster, username, role, action, stmt.GVK.String(), stmt.Namespace, stmt.Name)

	log := models.OperationLog{
		Action:       action,
		Cluster:      cluster,
		Kind:         stmt.GVK.Kind,
		Name:         stmt.Name,
		Namespace:    stmt.Namespace,
		UserName:     username,
		Group:        stmt.GVK.Group,
		Role:         role,
		ActionResult: "success",
	}

	var err error
	clusterRole, err := service.UserService().GetClusterRole(cluster, username)
	//clusterRole 为最高优先级
	if clusterRole == models.RoleClusterAdmin {
		//管理员不做处理
	}
	if clusterRole == models.RoleClusterReadonly {
		err = fmt.Errorf("非管理员不能%s资源", action)
	}
	if clusterRole == "" {
		//没有对该用户单独指定该集群的操作权限，那么使用ctx传递过来的用户角色
		if role == models.RoleClusterReadonly {
			err = fmt.Errorf("非管理员不能%s资源", action)
		}
	}

	if err != nil {
		log.ActionResult = err.Error()
	}
	go func() {
		time.Sleep(1 * time.Second)
		service.OperationLogService().Add(&log)
	}()
	return username, role, err
}

func handleDelete(k8s *kom.Kubectl) error {
	_, _, err := handleCommonLogic(k8s, "delete")
	return err
}

func handleUpdate(k8s *kom.Kubectl) error {
	_, _, err := handleCommonLogic(k8s, "update")
	return err
}

func handlePatch(k8s *kom.Kubectl) error {
	_, _, err := handleCommonLogic(k8s, "patch")
	return err
}

func handleCreate(k8s *kom.Kubectl) error {
	_, _, err := handleCommonLogic(k8s, "create")
	return err
}
