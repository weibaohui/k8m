package inspection

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/webhook"
	"gorm.io/gorm"
)

type AdminRecordController struct {
}

func RegisterAdminRecordRoutes(admin *gin.RouterGroup) {
	ctrl := &AdminRecordController{}
	admin.GET("/inspection/schedule/id/:id/record/list", ctrl.RecordList)
	admin.GET("/inspection/record/list", ctrl.RecordList)
	admin.POST("/inspection/schedule/record/id/:id/push", ctrl.Push)

}
func (r *AdminRecordController) RecordList(c *gin.Context) {
	params := dao.BuildParams(c)

	m := &models.InspectionRecord{}
	id := c.Param("id")
	if id != "" {
		m.ScheduleID = utils.UintPtr(utils.ToUInt(id))
	}

	items, total, err := m.List(params, func(db *gorm.DB) *gorm.DB {
		return db.Where(m)
	})
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}

func (r *AdminRecordController) Push(c *gin.Context) {
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
