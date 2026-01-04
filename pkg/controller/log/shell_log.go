package log

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/response"
	"gorm.io/gorm"
)

type Controller struct{}

// 从 gin 切换到 chi，使用 chi.Router 替代 gin.RouterGroup
func RegisterLogRoutes(r chi.Router) {
	ctrl := &Controller{}
	r.Get("/log/shell/list", response.Adapter(ctrl.ListShell))
	r.Get("/log/operation/list", response.Adapter(ctrl.ListOperation))
}

// @Summary Shell日志列表
// @Description 获取所有Shell操作日志
// @Security BearerAuth
// @Success 200 {object} string
// @Router /mgm/log/shell/list [get]
func (lc *Controller) ListShell(c *response.Context) {
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
func (lc *Controller) ListOperation(c *response.Context) {
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
