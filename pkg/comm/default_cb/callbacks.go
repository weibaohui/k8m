package default_cb

import (
	"github.com/weibaohui/k8m/pkg/comm/kubectl"
)

func RegisterDefaultCallbacks() {

	queryCallback := kubectl.Init().Callback().Get()
	_ = queryCallback.Register("k8m:get", Get)

	listCallback := kubectl.Init().Callback().List()
	_ = listCallback.Register("k8m:list", List)

	createCallback := kubectl.Init().Callback().Create()
	_ = createCallback.Register("k8m:create", Create)

}
