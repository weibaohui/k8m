package log

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
)

type Controller struct{}

func RegisterLogRoutes(mgm *gin.RouterGroup) {
	ctrl := &Controller{}
	mgm.GET("/log/shell/list", ctrl.ListShell)
	mgm.GET("/log/operation/list", ctrl.ListOperation)
}
func (lc *Controller) ListShell(c *gin.Context) {
	params := dao.BuildParams(c)
	m := &models.ShellLog{}

	items, total, err := m.List(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}

func (lc *Controller) ListOperation(c *gin.Context) {
	params := dao.BuildParams(c)
	m := &models.OperationLog{}

	items, total, err := m.List(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}
