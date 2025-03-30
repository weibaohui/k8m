package user

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/service"
)

func ListClusterPermissions(c *gin.Context) {
	params := dao.BuildParams(c)
	m := &models.ClusterUserRole{}

	items, total, err := m.List(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}

func SaveClusterPermission(c *gin.Context) {
	params := dao.BuildParams(c)
	m := &models.ClusterUserRole{}
	err := c.ShouldBindJSON(&m)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	_, _, err = handlePermissionCommonLogic(c, "保存", m.Cluster)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	err = m.Save(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, gin.H{"id": m.ID})
}

func DeleteClusterPermission(c *gin.Context) {
	ids := c.Param("ids")

	_, _, err := handlePermissionCommonLogic(c, "删除", ids)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	params := dao.BuildParams(c)
	m := &models.ClusterUserRole{}

	err = m.Delete(params, ids)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

func handlePermissionCommonLogic(c *gin.Context, action string, clusterName string) (string, string, error) {
	ctx := amis.GetContextWithUser(c)
	username := fmt.Sprintf("%s", ctx.Value(constants.JwtUserName))
	role := fmt.Sprintf("%s", ctx.Value(constants.JwtUserRole))

	log := models.OperationLog{
		Action:       action,
		Cluster:      clusterName,
		Kind:         "ClusterPermission",
		Name:         clusterName,
		UserName:     username,
		Group:        clusterName,
		Role:         role,
		ActionResult: "success",
	}

	var err error
	if role != models.RolePlatformAdmin {
		err = fmt.Errorf("非平台管理员不能%s权限配置", action)
	}

	if err != nil {
		log.ActionResult = err.Error()
	}

	go func() {
		time.Sleep(1 * time.Second)
		service.OperationLogService().Add(&log)
	}()

	return username, role, err
}
