package callback

import (
	"fmt"

	"github.com/weibaohui/k8m/internal/kubectl"
	"k8s.io/klog/v2"
)

func RegisterCallback() {
	queryCallback := kubectl.Init().Callback().Query()
	_ = queryCallback.Register("k8m:query", Query)
}

func Query(k8s *kubectl.Kubectl) error {
	json := k8s.Stmt.String()
	klog.V(2).Infof("k8s stmt json:\n%s\n", json)
	klog.V(2).Infof("QueryQueryQuery")
	klog.V(2).Infof("QueryQueryQuery")
	klog.V(2).Infof("QueryQueryQuery")
	klog.V(2).Infof("QueryQueryQuery")
	klog.V(2).Infof("QueryQueryQuery")
	return fmt.Errorf("无权限")
}
