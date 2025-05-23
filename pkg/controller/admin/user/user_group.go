package user

import (
	"fmt"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/service"
	"gorm.io/gorm"
)

func ListUserGroup(c *gin.Context) {
	params := dao.BuildParams(c)
	m := &models.UserGroup{}

	items, total, err := m.List(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}
func SaveUserGroup(c *gin.Context) {

	params := dao.BuildParams(c)
	m := &models.UserGroup{}
	err := c.ShouldBindJSON(&m)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	_, _, err = handleCommonLogic(c, "保存", m.GroupName)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	err = m.Save(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	// 清除用户组的缓存
	service.UserService().ClearCacheByKey(m.GroupName)
	amis.WriteJsonData(c, gin.H{
		"id": m.ID,
	})
}
func DeleteUserGroup(c *gin.Context) {
	ids := c.Param("ids")

	_, _, err := handleCommonLogic(c, "删除", ids)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	params := dao.BuildParams(c)
	m := &models.UserGroup{}

	err = m.Delete(params, ids)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	// 清除用户组的缓存
	service.UserService().ClearCacheByKey(m.GroupName)
	amis.WriteJsonOK(c)
}

func handleCommonLogic(c *gin.Context, action string, groupName string) (string, string, error) {
	cluster, _ := amis.GetSelectedCluster(c)
	ctx := amis.GetContextWithUser(c)
	username := fmt.Sprintf("%s", ctx.Value(constants.JwtUserName))
	role := fmt.Sprintf("%s", ctx.Value(constants.JwtUserRole))

	log := models.OperationLog{
		Action:       action,
		Cluster:      cluster,
		Kind:         "UserGroup",
		Name:         groupName,
		Namespace:    groupName,
		UserName:     username,
		Group:        groupName,
		Role:         role,
		ActionResult: "success",
	}

	var err error
	if !amis.IsCurrentUserPlatformAdmin(c) {
		err = fmt.Errorf("非平台管理员不能%s资源", action)
	}
	if err != nil {
		log.ActionResult = err.Error()
	}
	service.OperationLogService().Add(&log)
	return username, role, err
}

func GroupOptionList(c *gin.Context) {
	params := dao.BuildParams(c)
	m := &models.UserGroup{}
	items, _, err := m.List(params, func(db *gorm.DB) *gorm.DB {
		return db.Distinct("id,group_name")
	})
	if err != nil {
		amis.WriteJsonData(c, gin.H{
			"options": make([]map[string]string, 0),
		})
		return
	}
	var names []map[string]string
	for _, n := range items {
		names = append(names, map[string]string{
			"label": n.GroupName,
			"value": n.GroupName,
		})
	}
	slice.SortBy(names, func(a, b map[string]string) bool {
		return a["label"] < b["label"]
	})
	amis.WriteJsonData(c, gin.H{
		"options": names,
	})
}
