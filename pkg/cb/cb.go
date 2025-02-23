package cb

import (
	"fmt"

	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
	"k8s.io/klog/v2"
)

func RegisterCallback() {
	clusters := service.ClusterService().ConnectedClusters()
	for _, cluster := range clusters {
		selectedCluster := service.ClusterService().ClusterID(cluster)

		deleteCallback := kom.Cluster(selectedCluster).Callback().Delete()
		_ = deleteCallback.Register("k8m:delete", handleDelete)
		updateCallback := kom.Cluster(selectedCluster).Callback().Update()
		_ = updateCallback.Register("k8m:update", handleUpdate)
		patchCallback := kom.Cluster(selectedCluster).Callback().Patch()
		_ = patchCallback.Register("k8m:patch", handlePatch)
		createCallback := kom.Cluster(selectedCluster).Callback().Create()
		_ = createCallback.Register("k8m:create", handleCreate)
	}
}

func handleDelete(k8s *kom.Kubectl) error {
	stmt := k8s.Statement
	cluster := k8s.ID

	username := stmt.Context.Value(constants.JwtUserName)
	role := stmt.Context.Value(constants.JwtUserRole)
	klog.V(6).Infof("cb: cluster= %s,user= %s, role= %s, operation=delete, gck=[%s], resource=[%s/%s] ", cluster, username, role, stmt.GVK.String(), stmt.Namespace, stmt.Name)
	switch role {
	case models.RoleClusterReadonly:
		return fmt.Errorf("非管理员不能删除资源")
	}
	return nil
}

func handleUpdate(k8s *kom.Kubectl) error {
	stmt := k8s.Statement
	cluster := k8s.ID

	username := stmt.Context.Value(constants.JwtUserName)
	role := stmt.Context.Value(constants.JwtUserRole)
	klog.V(6).Infof("cb: cluster= %s,user= %s, role= %s, operation=update, gck=[%s], resource=[%s/%s] ", cluster, username, role, stmt.GVK.String(), stmt.Namespace, stmt.Name)
	switch role {
	case models.RoleClusterReadonly:
		return fmt.Errorf("非管理员不能更新资源")
	}
	return nil
}

func handlePatch(k8s *kom.Kubectl) error {
	stmt := k8s.Statement
	cluster := k8s.ID

	username := stmt.Context.Value(constants.JwtUserName)
	role := stmt.Context.Value(constants.JwtUserRole)
	klog.V(6).Infof("cb: cluster= %s,user= %s, role= %s, operation=patch, gck=[%s], resource=[%s/%s] ", cluster, username, role, stmt.GVK.String(), stmt.Namespace, stmt.Name)
	switch role {
	case models.RoleClusterReadonly:
		return fmt.Errorf("非管理员不能修改资源")
	}
	return nil
}

func handleCreate(k8s *kom.Kubectl) error {
	stmt := k8s.Statement
	cluster := k8s.ID

	username := stmt.Context.Value(constants.JwtUserName)
	role := stmt.Context.Value(constants.JwtUserRole)
	klog.V(6).Infof("cb: cluster= %s,user= %s, role= %s, operation=create, gck=[%s], resource=[%s/%s] ", cluster, username, role, stmt.GVK.String(), stmt.Namespace, stmt.Name)
	switch role {
	case models.RoleClusterReadonly:
		return fmt.Errorf("非管理员不能创建资源")
	}
	return nil
}
