package ns

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/kubectl"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
)

func OptionList(c *gin.Context) {
	ctx := c.Request.Context()
	namespace, err := kubectl.Init().ListNamespace(ctx)
	if err != nil {
		amis.WriteJsonError(c, err)
	}
	var list []map[string]string
	for _, ns := range namespace {
		list = append(list, map[string]string{
			"label": ns.Name,
			"value": ns.Name,
		})
	}
	amis.WriteJsonData(c, gin.H{
		"options": list,
	})
}
