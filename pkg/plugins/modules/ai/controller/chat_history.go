package controller

import (
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/plugins/modules/ai/service"
	"github.com/weibaohui/k8m/pkg/response"
)

// @Summary 获取聊天历史记录
// @Security BearerAuth
// @Success 200 {object} string
// @Router /ai/chat/history [get]
func (cc *Controller) History(c *response.Context) {
	client, err := service.AIService().DefaultClient()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	ctx := amis.GetContextWithUser(c)
	history := client.GetHistory(ctx)
	amis.WriteJsonData(c, history)

}

// @Summary 重置聊天历史记录
// @Security BearerAuth
// @Success 200 {object} string
// @Router /ai/chat/reset [post]
func (cc *Controller) Reset(c *response.Context) {
	client, err := service.AIService().DefaultClient()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	ctx := amis.GetContextWithUser(c)
	err = client.ClearHistory(ctx)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}
