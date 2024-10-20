package default_cb

import (
	"github.com/weibaohui/k8m/pkg/comm/kubectl"
)

func RegisterDefaultCallbacks() {

	queryCallback := kubectl.Init().Callback().Query()
	queryCallback.Register("k8m:query1", Query)

}
