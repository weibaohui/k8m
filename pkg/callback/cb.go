package callback

import (
	"context"

	"github.com/weibaohui/kom/kom"
	"k8s.io/klog/v2"
)

func RegisterCallback() {
	queryCallback := kom.Init().Callback().Get()
	_ = queryCallback.Register("k8m:get11", Get)
}

func Get(ctx context.Context, k8s *kom.Kom) error {
	json := k8s.Statement.String()
	// todo 在这里可以统一进行权限认证等操作，返回error即可阻断执行
	u := ctx.Value("user")
	klog.V(2).Infof("%s k8s Get stmt json:\n %s\n", u, json)
	return nil
}
