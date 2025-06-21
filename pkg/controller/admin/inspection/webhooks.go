package inspection

import (
	"fmt"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/webhook"
)

type AdminWebhookController struct {
}

func RegisterAdminWebhookRoutes(admin *gin.RouterGroup) {
	ctrl := &AdminWebhookController{}
	admin.GET("/inspection/webhook/list", ctrl.WebhookList)
	admin.POST("/inspection/webhook/delete/:ids", ctrl.WebhookDelete)
	admin.POST("/inspection/webhook/save", ctrl.WebhookSave)
	admin.POST("/inspection/webhook/id/:id/test", ctrl.WebhookTest)
	admin.GET("/inspection/webhook/option_list", ctrl.WebhookOptionList)

}

func (s *AdminWebhookController) WebhookOptionList(c *gin.Context) {
	m := models.WebhookReceiver{}
	params := dao.BuildParams(c)
	params.PerPage = 100000
	list, _, err := m.List(params)

	if err != nil {
		amis.WriteJsonData(c, gin.H{
			"options": make([]map[string]string, 0),
		})
		return
	}
	var hooks []map[string]string
	for _, n := range list {
		hooks = append(hooks, map[string]string{
			"label": n.Name,
			"value": fmt.Sprintf("%d", n.ID),
		})
	}
	slice.SortBy(hooks, func(a, b map[string]string) bool {
		return a["label"] < b["label"]
	})
	amis.WriteJsonData(c, gin.H{
		"options": hooks,
	})

}
func (s *AdminWebhookController) WebhookTest(c *gin.Context) {
	id := c.Param("id")
	params := dao.BuildParams(c)
	m := &models.WebhookReceiver{
		ID: utils.ToUInt(id),
	}
	m, err := m.GetOne(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	ret := webhook.PushMsgToSingleTarget("test", m)
	if ret != nil {
		amis.WriteJsonOKMsg(c, ret.RespBody)
		return
	}

	amis.WriteJsonError(c, fmt.Errorf("unsupported platform: %s", m.Platform))
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
