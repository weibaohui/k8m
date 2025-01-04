package cb

import (
	"fmt"

	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
	"k8s.io/klog/v2"
)

func RegisterCallback() {
	clusters := service.ClusterService().ConnectedClusters()
	for _, cluster := range clusters {
		selectedCluster := fmt.Sprintf("%s/%s", cluster.FileName, cluster.ContextName)
		klog.V(6).Infof("注册回调%s", selectedCluster)
		queryCallback := kom.Cluster(selectedCluster).Callback().Get()
		_ = queryCallback.Register("k8m:get11", Get)
	}

}

func Get(k8s *kom.Kubectl) error {
	// // todo 在这里可以统一进行权限认证等操作，返回error即可阻断执行
	// u := k8s.Statement.Context.Value("user")
	// klog.V(2).Infof("%s k8s Get \n", u)
	return nil
}
