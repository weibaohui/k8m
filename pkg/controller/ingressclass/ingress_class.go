package ingressclass

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/networking/v1"
)

func SetDefault(c *gin.Context) {
	name := c.Param("name")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	err := kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.IngressClass{}).Name(name).
		Ctl().IngressClass().SetDefault()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}
