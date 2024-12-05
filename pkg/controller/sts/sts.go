package sts

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/apps/v1"
)

func History(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	ctx := c.Request.Context()
	list, err := kom.DefaultCluster().WithContext(ctx).Resource(&v1.StatefulSet{}).Namespace(ns).Name(name).
		Ctl().Rollout().History()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, list)
}

func Undo(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	revision := c.Param("revision")
	ctx := c.Request.Context()
	r := utils.ToInt(revision)
	result, err := kom.DefaultCluster().WithContext(ctx).Resource(&v1.StatefulSet{}).Namespace(ns).Name(name).
		Ctl().Rollout().Undo(r)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOKMsg(c, result)
}
