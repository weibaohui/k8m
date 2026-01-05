package admin

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/response"

	"github.com/weibaohui/k8m/pkg/plugins/modules/eventhandler/models"
	"github.com/weibaohui/k8m/pkg/plugins/modules/eventhandler/worker"
	"github.com/weibaohui/k8m/pkg/plugins/modules/webhook"
	"gorm.io/gorm"
)

// Controller 中文函数注释：事件转发配置管理控制器。
type Controller struct{}

// GetSetting 中文函数注释：获取事件转发总开关与运行参数配置。
func (s *Controller) GetSetting(c *response.Context) {
	setting, err := models.GetOrCreateEventForwardSetting()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, setting)
}

// UpdateSetting 中文函数注释：更新事件转发总开关与运行参数配置。
func (s *Controller) UpdateSetting(c *response.Context) {
	var in models.EventForwardSetting
	if err := c.ShouldBindJSON(&in); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	if in.EventWorkerProcessInterval <= 0 {
		in.EventWorkerProcessInterval = 10
	}
	if in.EventWorkerBatchSize <= 0 {
		in.EventWorkerBatchSize = 50
	}
	if in.EventWorkerMaxRetries <= 0 {
		in.EventWorkerMaxRetries = 3
	}
	if in.EventWatcherBufferSize <= 0 {
		in.EventWatcherBufferSize = 1000
	}

	if _, err := models.UpdateEventForwardSetting(&in); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	if w := worker.Instance(); w != nil {
		w.UpdateConfig()
	}
	amis.WriteJsonOK(c)
}

// List 中文函数注释：获取事件配置列表。
func (s *Controller) List(c *response.Context) {
	params := dao.BuildParams(c)
	m := &models.K8sEventConfig{}
	items, total, err := m.List(params, func(db *gorm.DB) *gorm.DB {
		return db.Order("id desc")
	})
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}

// Save 中文函数注释：保存或更新事件配置。
func (s *Controller) Save(c *response.Context) {
	params := dao.BuildParams(c)
	m := models.K8sEventConfig{}
	if err := c.ShouldBindJSON(&m); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	if m.AIEnabled {
		if len(m.AIPromptTemplate) > 2000 {
			amis.WriteJsonError(c, fmt.Errorf("AI提示词模板长度不能超过2000个字符"))
			return
		}
		if strings.TrimSpace(m.AIPromptTemplate) != "" {
			if strings.Contains(m.AIPromptTemplate, "<script>") || strings.Contains(m.AIPromptTemplate, "javascript:") {
				amis.WriteJsonError(c, fmt.Errorf("AI提示词模板包含不安全的内容"))
				return
			}
		}
	}

	// 保存webhookNames
	names, err := webhook.GetNamesByIds(strings.Split(m.Webhooks, ","))
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	m.WebhookNames = strings.Join(names, ",")

	normalize := func(v string) string {
		t := strings.TrimSpace(v)
		if t == "" {
			return ""
		}
		if strings.HasPrefix(t, "[") {
			return t
		}
		parts := strings.Split(t, ",")
		arr := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				arr = append(arr, p)
			}
		}
		b, _ := json.Marshal(arr)
		return string(b)
	}

	m.RuleNamespaces = normalize(m.RuleNamespaces)
	m.RuleNames = normalize(m.RuleNames)
	m.RuleReasons = normalize(m.RuleReasons)

	if err := m.Save(params); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	if w := worker.Instance(); w != nil {
		w.UpdateConfig()
	}
	amis.WriteJsonOK(c)
}

// Delete 中文函数注释：删除事件配置。
func (s *Controller) Delete(c *response.Context) {
	ids := c.Param("ids")
	params := dao.BuildParams(c)
	m := &models.K8sEventConfig{}
	if err := m.Delete(params, ids); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	if w := worker.Instance(); w != nil {
		w.UpdateConfig()
	}
	amis.WriteJsonOK(c)
}

// QuickSave 中文函数注释：快速更新事件配置启用状态。
func (s *Controller) QuickSave(c *response.Context) {
	id := c.Param("id")
	enabled := c.Param("enabled")

	var entity models.K8sEventConfig
	entity.ID = utils.ToUInt(id)
	entity.Enabled = enabled == "true"

	if err := dao.DB().Model(&entity).Select("enabled").Updates(entity).Error; err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	if w := worker.Instance(); w != nil {
		w.UpdateConfig()
	}
	amis.WriteJsonOK(c)
}
