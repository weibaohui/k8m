package chat

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/service"
)

func History(c *gin.Context) {
	client, err := service.AIService().DefaultClient()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	ctx := amis.GetContextWithUser(c)
	history := client.GetHistory(ctx)
	amis.WriteJsonData(c, history)

}
func Reset(c *gin.Context) {
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
