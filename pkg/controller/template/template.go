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
