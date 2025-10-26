package inspection

import (
	"context"
	"fmt"
	"strings"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/lua"
	"github.com/weibaohui/k8m/pkg/models"
	"gorm.io/gorm"
)

type AdminScheduleController struct {
}

func RegisterAdminScheduleRoutes(admin *gin.RouterGroup) {
	ctrl := &AdminScheduleController{}
	admin.GET("/inspection/schedule/list", ctrl.List)
	admin.GET("/inspection/schedule/record/id/:id/event/list", ctrl.EventList)
	admin.POST("/inspection/schedule/record/id/:id/summary", ctrl.SummaryByRecordID)
	admin.GET("/inspection/schedule/record/id/:id/output/list", ctrl.OutputList)
	admin.POST("/inspection/schedule/save", ctrl.Save)
	admin.POST("/inspection/schedule/delete/:ids", ctrl.Delete)
	admin.POST("/inspection/schedule/save/id/:id/status/:enabled", ctrl.QuickSave)
	admin.POST("/inspection/schedule/start/id/:id", ctrl.Start)
	admin.POST("/inspection/schedule/id/:id/update_script_code", ctrl.UpdateScriptCode)
	admin.POST("/inspection/schedule/id/:id/summary", ctrl.SummaryBySchedule)
	admin.POST("/inspection/schedule/id/:id/summary/cluster/:cluster/start_time/:start_time/end_time/:end_time", ctrl.SummaryBySchedule)
	admin.GET("/inspection/event/status/option_list", ctrl.EventStatusOptionList)
}

// @Summary 获取巡检计划列表
// @Security BearerAuth
// @Success 200 {object} string
// @Router /admin/inspection/schedule/list [get]
func (s *AdminScheduleController) List(c *gin.Context) {
	params := dao.BuildParams(c)
	m := &models.InspectionSchedule{}

	items, total, err := m.List(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}

// @Summary 获取巡检脚本输出列表
// @Security BearerAuth
// @Param id path string true "巡检记录ID"
// @Success 200 {object} string
// @Router /admin/inspection/schedule/record/id/{id}/output/list [get]
func (s *AdminScheduleController) OutputList(c *gin.Context) {
	params := dao.BuildParams(c)
	params.PerPage = 10000
	id := c.Param("id")
	m := &models.InspectionScriptResult{
		RecordID: utils.ToUInt(id),
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

// @Summary 获取巡检事件列表
// @Security BearerAuth
// @Param id path string true "巡检记录ID"
// @Success 200 {object} string
// @Router /admin/inspection/schedule/record/id/{id}/event/list [get]
func (s *AdminScheduleController) EventList(c *gin.Context) {
	params := dao.BuildParams(c)
	params.PerPage = 10000
	id := c.Param("id")
	m := &models.InspectionCheckEvent{
		RecordID: utils.ToUInt(id),
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

// @Summary 获取巡检事件状态选项列表
// @Security BearerAuth
// @Success 200 {object} string
// @Router /admin/inspection/event/status/option_list [get]
func (s *AdminScheduleController) EventStatusOptionList(c *gin.Context) {
	m := &models.InspectionCheckEvent{}
	events, _, err := m.List(nil, func(db *gorm.DB) *gorm.DB {
		return db.Distinct("event_status")
	})
	if err != nil {
		amis.WriteJsonData(c, gin.H{
			"options": make([]map[string]string, 0),
		})
		return
	}
	var names []map[string]string
	for _, n := range events {
		names = append(names, map[string]string{
			"label": n.EventStatus,
			"value": n.EventStatus,
		})
	}
	slice.SortBy(names, func(a, b map[string]string) bool {
		return a["label"] < b["label"]
	})
	amis.WriteJsonData(c, gin.H{
		"options": names,
	})
}

// @Summary 保存巡检计划
// @Security BearerAuth
// @Success 200 {object} string
// @Router /admin/inspection/schedule/save [post]
func (s *AdminScheduleController) Save(c *gin.Context) {
	params := dao.BuildParams(c)
	m := models.InspectionSchedule{}
	err := c.ShouldBindJSON(&m)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 检测cron表达式是否正确
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	// 尝试解析
	_, err = parser.Parse(m.Cron)
	if err != nil {
		amis.WriteJsonError(c, fmt.Errorf("cron表达式错误: %w", err))
		return
	}

	// 验证AI总结配置
	if m.AIEnabled {
		// 检查AI提示词模板长度
		if len(m.AIPromptTemplate) > 2000 {
			amis.WriteJsonError(c, fmt.Errorf("AI提示词模板长度不能超过2000个字符"))
			return
		}

		// 可以添加更多AI配置验证逻辑
		// 例如：检查模板格式、关键词等
		if strings.TrimSpace(m.AIPromptTemplate) != "" {
			// 简单验证：确保模板不包含危险字符
			if strings.Contains(m.AIPromptTemplate, "<script>") ||
				strings.Contains(m.AIPromptTemplate, "javascript:") {
				amis.WriteJsonError(c, fmt.Errorf("AI提示词模板包含不安全的内容"))
				return
			}
		}
	}

	// 保存webhookNames
	receiver := models.WebhookReceiver{}
	if names, nErr := receiver.GetNamesByIds(m.Webhooks); nErr == nil {
		m.WebhookNames = strings.Join(names, ",")
	} else {
		amis.WriteJsonError(c, nErr)
		return
	}

	err = m.Save(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	sb := lua.NewScheduleBackground()
	if m.Enabled {
		go func() {
			sb.Update(m.ID)
		}()
	}

	amis.WriteJsonOK(c)
}

// @Summary 删除巡检计划
// @Security BearerAuth
// @Param ids path string true "巡检计划ID，多个用逗号分隔"
// @Success 200 {object} string
// @Router /admin/inspection/schedule/delete/{ids} [post]
func (s *AdminScheduleController) Delete(c *gin.Context) {
	ids := c.Param("ids")
	params := dao.BuildParams(c)
	// 清除定时 任务
	go func() {
		for id := range strings.SplitSeq(ids, ",") {
			sb := lua.NewScheduleBackground()
			sb.Remove(utils.ToUInt(id))
		}
	}()

	// 查询到需清除的执行记录
	var records []*models.InspectionRecord
	if err := dao.DB().Model(&records).Where("schedule_id in (?)", ids).Find(&records).Error; err == nil {
		recordIds := make([]uint, len(records))
		for i, record := range records {
			recordIds[i] = record.ID
		}
		// 先清除检测历史事件
		events := &models.InspectionCheckEvent{}
		dao.DB().Model(&events).Where("record_id in (?)", recordIds).Delete(&events)
		scriptResult := models.InspectionScriptResult{}
		dao.DB().Model(&scriptResult).Where("record_id in (?)", recordIds).Delete(&scriptResult)

		// 再清除执行记录
		intIds := utils.ToInt64Slice(ids)
		dao.DB().Model(&records).Where("schedule_id in (?)", intIds).Delete(&records)
	}

	// 删除计划
	m := &models.InspectionSchedule{}
	err := m.Delete(params, ids)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOK(c)
}

// @Summary 快速更新巡检计划状态
// @Security BearerAuth
// @Param id path int true "巡检计划ID"
// @Param enabled path string true "状态，例如：true、false"
// @Success 200 {object} string
// @Router /admin/inspection/schedule/save/id/{id}/status/{enabled} [post]
func (s *AdminScheduleController) QuickSave(c *gin.Context) {
	id := c.Param("id")
	enabled := c.Param("enabled")

	var entity models.InspectionSchedule
	entity.ID = utils.ToUInt(id)

	if enabled == "true" {
		entity.Enabled = true
	} else {
		entity.Enabled = false
	}
	err := dao.DB().Model(&entity).Select("enabled").Updates(entity).Error

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	go func() {
		// 存储后，按照开关状态确定执行cron
		sb := lua.NewScheduleBackground()
		if entity.Enabled {
			sb.Update(entity.ID)
		} else {
			sb.Remove(entity.ID)
		}
	}()

	amis.WriteJsonErrorOrOK(c, err)
}

// @Summary 启动巡检计划，马上执行一次
// @Security BearerAuth
// @Param id path int true "巡检计划ID"
// @Success 200 {object} string
// @Router /admin/inspection/schedule/start/id/{id} [post]
func (s *AdminScheduleController) Start(c *gin.Context) {
	id := c.Param("id")
	m := &models.InspectionSchedule{
		ID: utils.ToUInt(id),
	}

	one, err := m.GetOne(nil)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	if strings.TrimSpace(one.ScriptCodes) == "" {
		amis.WriteJsonError(c, fmt.Errorf("无检测规则，请先在 操作-管理规则 菜单中配置"))
		return
	}
	go func() {
		// 立马执行一次
		sb := lua.NewScheduleBackground()
		clusters := strings.Split(one.Clusters, ",")
		for _, cluster := range clusters {
			_, _ = sb.RunByCluster(context.Background(), &one.ID, cluster, lua.TriggerTypeManual)
		}
	}()

	amis.WriteJsonOKMsg(c, "巡检开始，请稍后刷新查看结果")
}

// @Summary 更新巡检脚本代码
// @Security BearerAuth
// @Param id path int true "巡检计划ID"
// @Param script_codes body string true "脚本代码"
// @Success 200 {object} string
// @Router /admin/inspection/schedule/id/{id}/update_script_code [post]
func (s *AdminScheduleController) UpdateScriptCode(c *gin.Context) {
	id := c.Param("id")
	type requestBody struct {
		ScriptCodes string `json:"script_codes"`
	}
	var codes requestBody

	err := c.ShouldBindJSON(&codes)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	params := dao.BuildParams(c)
	m := &models.InspectionSchedule{}
	m.ID = utils.ToUInt(id)
	m.ScriptCodes = codes.ScriptCodes
	err = m.Save(params, func(db *gorm.DB) *gorm.DB {
		return db.Select("script_codes")
	})
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	go func() {
		sb := lua.NewScheduleBackground()
		sb.Remove(m.ID)
		sb.Add(m.ID)
	}()

	amis.WriteJsonOK(c)
}
