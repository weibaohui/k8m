package cb

import (
	"fmt"
	"strings"

	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
	"k8s.io/klog/v2"
)

func RegisterCallback() {
	clusters := service.ClusterService().ConnectedClusters()
	for _, cluster := range clusters {
		selectedCluster := service.ClusterService().ClusterID(cluster)
		klog.V(6).Infof("注册回调%s", selectedCluster)
		queryCallback := kom.Cluster(selectedCluster).Callback().Get()
		_ = queryCallback.Register("k8m:get", Get)
		execCallback := kom.Cluster(selectedCluster).Callback().Exec()
		_ = execCallback.Register("k8m:exec", Audit)
		streamExecCallback := kom.Cluster(selectedCluster).Callback().StreamExec()
		_ = streamExecCallback.Register("k8m:streamExec", Audit)
	}

}

func Get(k8s *kom.Kubectl) error {
	stmt := k8s.Statement
	cluster := k8s.ID
	klog.V(2).Infof("k8s [%s] Get %s %s/%s \n", cluster, stmt.GVR.Resource, stmt.Namespace, stmt.Name)
	return nil
}
func Audit(k8s *kom.Kubectl) error {
	stmt := k8s.Statement
	cluster := k8s.ID
	cmd := fmt.Sprintf("%s %s", stmt.Command, strings.Join(stmt.Args, " "))
	klog.V(2).Infof("k8s [%s] Exec cmd in %s %s/%s : %s \n", cluster, stmt.GVR.Resource, stmt.Namespace, stmt.Name, cmd)
	return nil
}
