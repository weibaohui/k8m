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

func List(c *gin.Context) {
	params := dao.BuildParams(c)
	m := &models.InspectionSchedule{}

	items, total, err := m.List(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}
func RecordList(c *gin.Context) {
	params := dao.BuildParams(c)
	params.PerPage = 10000

	id := c.Param("id")
	m := &models.InspectionRecord{
		ScheduleID: utils.UintPtr(utils.ToUInt(id)),
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

func OutputList(c *gin.Context) {
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

func EventList(c *gin.Context) {
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

func EventStatusOptionList(c *gin.Context) {
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

func Save(c *gin.Context) {
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
	err = m.Save(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 存储后，按照开关状态确定执行cron
	sb := lua.ScheduleBackground{}
	if m.Enabled {
		sb.Add(m.ID)
	} else {
		sb.Remove(m.ID)
	}

	amis.WriteJsonOK(c)
}
func Delete(c *gin.Context) {
	ids := c.Param("ids")
	params := dao.BuildParams(c)
	// 清除定时 任务
	intIds := utils.ToInt64Slice(ids)
	for _, id := range intIds {
		sb := lua.ScheduleBackground{}
		sb.Remove(uint(id))
	}

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

func QuickSave(c *gin.Context) {
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

	// 存储后，按照开关状态确定执行cron
	sb := lua.ScheduleBackground{}
	if entity.Enabled {
		sb.Add(entity.ID)
	} else {
		sb.Remove(entity.ID)
	}

	amis.WriteJsonErrorOrOK(c, err)
}

func Start(c *gin.Context) {
	id := c.Param("id")
	m := &models.InspectionSchedule{
		ID: utils.ToUInt(id),
	}

	one, err := m.GetOne(nil)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	func() {
		sb := lua.ScheduleBackground{}
		clusters := strings.Split(one.Clusters, ",")
		for _, cluster := range clusters {
			_, _ = sb.RunByCluster(context.Background(), &one.ID, cluster, lua.TriggerTypeManual)
		}
	}()

	amis.WriteJsonOKMsg(c, "巡检开始，请稍后刷新查看结果")
}
func UpdateScriptCode(c *gin.Context) {
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
	amis.WriteJsonOK(c)
}
