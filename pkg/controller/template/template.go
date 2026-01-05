package template

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/response"
)

type Controller struct {
}

// RegisterTemplateRoutes 注册模板相关路由

func RegisterTemplateRoutes(r chi.Router) {
	ctrl := &Controller{}
	r.Get("/custom/template/kind/list", response.Adapter(ctrl.ListKind))
	r.Get("/custom/template/list", response.Adapter(ctrl.List))
	r.Post("/custom/template/save", response.Adapter(ctrl.Save))
	r.Post("/custom/template/delete/{ids}", response.Adapter(ctrl.Delete))
}

// @Summary 模板列表
// @Description 获取所有自定义模板信息
// @Security BearerAuth
// @Success 200 {object} string
// @Router /mgm/custom/template/list [get]
func (t *Controller) List(c *response.Context) {
	params := dao.BuildParams(c)
	m := &models.CustomTemplate{}

	items, total, err := m.List(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}

// @Summary 保存模板
// @Description 新增或更新自定义模板
// @Security BearerAuth
// @Param template body models.CustomTemplate true "模板信息"
// @Success 200 {object} string "返回模板ID"
// @Router /mgm/custom/template/save [post]
func (t *Controller) Save(c *response.Context) {
	params := dao.BuildParams(c)
	m := models.CustomTemplate{}
	err := c.ShouldBindJSON(&m)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 处理分类
	if m.Kind == "" {
		m.Kind = "未分类"
	}

	err = m.Save(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, response.H{
		"id": m.ID,
	})
}

// @Summary 删除模板
// @Description 删除一个或多个自定义模板
// @Security BearerAuth
// @Param ids path string true "要删除的模板ID，多个用逗号分隔"
// @Success 200 {object} string "操作成功"
// @Router /mgm/custom/template/delete/{ids} [post]
func (t *Controller) Delete(c *response.Context) {
	ids := c.Param("ids")
	params := dao.BuildParams(c)
	m := &models.CustomTemplate{}
	err := m.Delete(params, ids)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}
