package dynamic

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func UpdateLabels(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	kind := c.Param("kind")
	group := c.Param("group")
	version := c.Param("version")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var req struct {
		Labels map[string]string `json:"labels"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var obj *unstructured.Unstructured
	err = kom.Cluster(selectedCluster).WithContext(ctx).
		Name(name).Namespace(ns).
		CRD(group, version, kind).
		Get(&obj).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	obj.SetLabels(req.Labels)

	err = kom.Cluster(selectedCluster).WithContext(ctx).
		Name(name).Namespace(ns).
		CRD(group, version, kind).
		Update(obj).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOK(c)
}
