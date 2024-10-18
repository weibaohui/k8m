package callback

import (
	"context"

	"github.com/weibaohui/k8m/internal/kubectl"
	"k8s.io/klog/v2"
)

func RegisterCallback() {
	queryCallback := kubectl.Init().Callback().Query()
	_ = queryCallback.Register("k8m:query", Query)
}

func Query(ctx context.Context, k8s *kubectl.Kubectl) error {
	json := k8s.Stmt.String()
	// todo 在这里可以统一进行权限认证等操作，返回error即可阻断执行
	u := ctx.Value("user")
	klog.V(2).Infof("%s k8s Query stmt json:\n %s\n", u, json)
	return nil
}
