package log

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"gorm.io/gorm"
)

type Controller struct{}

func RegisterLogRoutes(mgm *gin.RouterGroup) {
	ctrl := &Controller{}
	mgm.GET("/log/shell/list", ctrl.ListShell)
	mgm.GET("/log/operation/list", ctrl.ListOperation)
}



// @Summary Shell日志列表
// @Description 获取所有Shell操作日志
// @Security BearerAuth
// @Success 200 {object} string
// @Router /mgm/log/shell/list [get]
func (lc *Controller) ListShell(c *gin.Context) {
	params := dao.BuildParams(c)
	m := &models.ShellLog{}

	// 处理时间范围查询
	var queryFuncs []func(*gorm.DB) *gorm.DB
	if queryFunc, ok := dao.BuildCreatedAtQuery(params); ok {
		queryFuncs = append(queryFuncs, queryFunc)
	}

	items, total, err := m.List(params, queryFuncs...)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}

// @Summary 操作日志列表
// @Description 获取所有操作日志
// @Security BearerAuth
// @Success 200 {object} string
// @Router /mgm/log/operation/list [get]
func (lc *Controller) ListOperation(c *gin.Context) {
	params := dao.BuildParams(c)
	m := &models.OperationLog{}

	// 处理时间范围查询
	var queryFuncs []func(*gorm.DB) *gorm.DB
	if queryFunc, ok := dao.BuildCreatedAtQuery(params); ok {
		queryFuncs = append(queryFuncs, queryFunc)
	}

	items, total, err := m.List(params, queryFuncs...)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}
