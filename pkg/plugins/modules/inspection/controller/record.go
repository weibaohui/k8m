package controller

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/inspection/models"
	"github.com/weibaohui/k8m/pkg/plugins/modules/webhook"
	"github.com/weibaohui/k8m/pkg/response"

	"gorm.io/gorm"
)

type AdminRecordController struct {
}

// RegisterAdminRecordRoutes 注册巡检记录路由
// 从 gin 切换到 chi，使用 chi.Router 替代 gin.RouterGroup
// Chi 中使用 chi.NewRouter() 创建子路由
func RegisterAdminRecordRoutes(arg chi.Router) {
	admin := chi.NewRouter()
	ctrl := &AdminRecordController{}
	admin.Get("/schedule/id/{id}/record/list", response.Adapter(ctrl.RecordList))
	admin.Get("/record/list", response.Adapter(ctrl.RecordList))
	admin.Post("/schedule/record/id/{id}/push", response.Adapter(ctrl.Push))

	arg.Mount("/plugins/"+modules.PluginNameInspection, admin)
}

// @Summary 获取巡检记录列表
// @Description 根据巡检计划ID获取对应的巡检记录列表
// @Security BearerAuth
// @Param id path string false "巡检计划ID"
// @Success 200 {object} string
// @Router /admin/plugins/inspection/schedule/id/{id}/record/list [get]
// @Router /admin/plugins/inspection/record/list [get]
func (r *AdminRecordController) RecordList(c *response.Context) {
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
// @Router /admin/plugins/inspection/schedule/record/id/{id}/push [post]
func (r *AdminRecordController) Push(c *response.Context) {

	recordIDStr := c.Param("id")

	recordID := utils.ToUInt(recordIDStr)
	record := &models.InspectionRecord{}
	summary, resultRaw, _, _, err := record.GetRecordBothContentById(recordID)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	// 使用inspection插件的辅助函数获取webhook接收器
	receivers, err := models.GetWebhookReceiverIDsByRecordID(recordID)

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	webhook.PushMsgToAllTargetByIDs(summary, resultRaw, receivers)

	amis.WriteJsonOK(c)
}
