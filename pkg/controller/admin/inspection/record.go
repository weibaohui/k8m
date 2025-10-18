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

// @Summary 获取巡检记录列表
// @Description 根据巡检计划ID获取对应的巡检记录列表
// @Security BearerAuth
// @Param id path string false "巡检计划ID"
// @Success 200 {object} string
// @Router /admin/inspection/schedule/id/{id}/record/list [get]
// @Router /admin/inspection/record/list [get]
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

// @Summary 推送巡检记录
// @Description 将指定巡检记录的AI总结推送到所有配置的Webhook接收器
// @Security BearerAuth
// @Param id path string true "巡检记录ID"
// @Success 200 {object} string
// @Router /admin/inspection/schedule/record/id/{id}/push [post]
func (r *AdminRecordController) Push(c *gin.Context) {
	recordIDStr := c.Param("id")
	recordID := utils.ToUInt(recordIDStr)
	record := &models.InspectionRecord{}
	summary, resultRaw, err := record.GetRecordBothContentById(recordID)
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

	webhook.PushMsgToAllTargets(summary, resultRaw, receivers)

	amis.WriteJsonOK(c)
}
