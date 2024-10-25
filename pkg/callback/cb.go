package callback

import (
	"github.com/weibaohui/kom/kom"
	"k8s.io/klog/v2"
)

func RegisterCallback() {
	queryCallback := kom.DefaultCluster().Callback().Get()
	_ = queryCallback.Register("k8m:get11", Get)
}

func Get(k8s *kom.Kubectl) error {
	// todo 在这里可以统一进行权限认证等操作，返回error即可阻断执行
	u := k8s.Statement.Context.Value("user")
	klog.V(2).Infof("%s k8s Get \n", u)
	return nil
}
