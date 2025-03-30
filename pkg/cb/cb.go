package cb

import (
	"fmt"
	"time"

	"github.com/duke-git/lancet/v2/slice"
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
	execCallback := kom.Cluster(selectedCluster).Callback().Exec()
	_ = execCallback.Before("kom:exec").Register("k8m:exec", handleExec)

	return nil
}

func handleCommonLogic(k8s *kom.Kubectl, action string) (string, []string, error) {
	stmt := k8s.Statement
	cluster := k8s.ID
	ctx := stmt.Context
	username := fmt.Sprintf("%s", ctx.Value(constants.JwtUserName))
	roles := fmt.Sprintf("%s", ctx.Value(constants.JwtUserRole))

	log := models.OperationLog{
		Action:       action,
		Cluster:      cluster,
		Kind:         stmt.GVK.Kind,
		Name:         stmt.Name,
		Namespace:    stmt.Namespace,
		UserName:     username,
		Group:        stmt.GVK.Group,
		Role:         roles,
		ActionResult: "success",
	}

	var err error
	clusterRoles, _ := service.UserService().GetClusterRole(cluster, username, roles)

	if len(clusterRoles) == 0 {
		err = fmt.Errorf("非管理员不能%s资源", action)
	}

	switch action {
	case "exec":
		// 只允许具备exec权限的用户，包括管理员和平台管理员、exec权限的用户
		if !(slice.Contain(clusterRoles, models.RolePlatformAdmin) || slice.Contain(clusterRoles, models.RoleClusterAdmin) || slice.Contain(clusterRoles, models.RoleClusterPodExec)) {
			err = fmt.Errorf("非管理员,且无exec权限，不能%s资源", action)
		}
	case "delete", "update", "patch", "create":
		if !(slice.Contain(clusterRoles, models.RolePlatformAdmin) || slice.Contain(clusterRoles, models.RoleClusterAdmin)) {
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
	klog.V(6).Infof("cb: cluster= %s,user= %s, role= %s，roles=%v, operation=%s, gck=[%s], resource=[%s/%s] ",
		cluster, username, roles, clusterRoles, action, stmt.GVK.String(), stmt.Namespace, stmt.Name)
	klog.V(6).Infof("final error=%v", err)
	return username, clusterRoles, err
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
func handleExec(k8s *kom.Kubectl) error {
	_, _, err := handleCommonLogic(k8s, "exec")
	return err
}
