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

	updateCallback := kubectl.Init().Callback().Update()
	_ = updateCallback.Register("k8m:update", Update)

	patchCallback := kubectl.Init().Callback().Patch()
	_ = patchCallback.Register("k8m:patch", Patch)

	deleteCallback := kubectl.Init().Callback().Delete()
	_ = deleteCallback.Register("k8m:delete", Delete)
}
