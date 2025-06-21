package inspection

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/webhook"
)

func Push(c *gin.Context) {
	recordIDStr := c.Param("id")
	recordID := utils.ToUInt(recordIDStr)
	record := &models.InspectionRecord{}
	summary, err := record.GetAISummaryById(recordID)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	receiver := &models.WebhookReceiver{}
	receivers, err := receiver.ListByRecordID(recordID)

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	webhook.PushMsgToAllTargets(summary, receivers)

	amis.WriteJsonOK(c)
}
