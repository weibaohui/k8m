package profile

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"gorm.io/gorm"
)

// 列出当前登录用户所拥有的集群权限
func ListUserPermissions(c *gin.Context) {
	params := dao.BuildParams(c)

	m := &models.ClusterUserRole{}
	m.Username = params.UserName
	params.UserName = "" //避免增加CreatedBy字段,因为查询用户集群权限，是管理员授权的，所以不需要CreatedBy字段
	items, total, err := m.List(params, func(db *gorm.DB) *gorm.DB {
		return db.Where(m)
	})
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}