package template

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
)

type Controller struct {
}

func RegisterTemplateRoutes(mgm *gin.RouterGroup) {
	ctrl := &Controller{}
	mgm.GET("/custom/template/kind/list", ctrl.ListKind)
	mgm.GET("/custom/template/list", ctrl.List)
	mgm.POST("/custom/template/save", ctrl.Save)
	mgm.POST("/custom/template/delete/:ids", ctrl.Delete)
}

// @Summary 模板列表
// @Description 获取所有自定义模板信息
// @Security BearerAuth
// @Success 200 {object} string
// @Router /mgm/custom/template/list [get]
func (t *Controller) List(c *gin.Context) {
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
func (t *Controller) Save(c *gin.Context) {
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
	amis.WriteJsonData(c, gin.H{
		"id": m.ID,
	})
}

// @Summary 删除模板
// @Description 删除一个或多个自定义模板
// @Security BearerAuth
// @Param ids path string true "要删除的模板ID，多个用逗号分隔"
// @Success 200 {object} string "操作成功"
// @Router /mgm/custom/template/delete/{ids} [post]
func (t *Controller) Delete(c *gin.Context) {
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
