package default_cb

import (
	"github.com/weibaohui/k8m/pkg/comm/kubectl"
)

func RegisterDefaultCallbacks() {

	queryCallback := kubectl.Init().Callback().Get()
	queryCallback.Register("k8m:get", Get)

	listCallback := kubectl.Init().Callback().List()
	listCallback.Register("k8m:list", List)

}
