package inspection

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
)

type AdminWebhookController struct {
}

func RegisterAdminWebhookRoutes(admin *gin.RouterGroup) {
	ctrl := &AdminWebhookController{}
	admin.GET("/inspection/webhook/list", ctrl.WebhookList)
	admin.POST("/inspection/webhook/delete/:ids", ctrl.WebhookDelete)
	admin.POST("/inspection/webhook/save", ctrl.WebhookSave)
}

func (s *AdminWebhookController) WebhookList(c *gin.Context) {
	params := dao.BuildParams(c)
	m := &models.WebhookReceiver{}

	items, total, err := m.List(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}

func (s *AdminWebhookController) WebhookSave(c *gin.Context) {
	params := dao.BuildParams(c)
	m := models.WebhookReceiver{}
	err := c.ShouldBindJSON(&m)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	err = m.Save(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

func (s *AdminWebhookController) WebhookDelete(c *gin.Context) {
	ids := c.Param("ids")
	params := dao.BuildParams(c)
	m := &models.WebhookReceiver{}
	err := m.Delete(params, ids)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}
