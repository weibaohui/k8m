package admin

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	coremodels "github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/plugins/modules/eventhandler/models"
	"github.com/weibaohui/k8m/pkg/plugins/modules/eventhandler/worker"
	"gorm.io/gorm"
)

// Controller 中文函数注释：事件转发配置管理控制器。
type Controller struct{}

// List 中文函数注释：获取事件配置列表。
func (s *Controller) List(c *gin.Context) {
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
func (s *Controller) Save(c *gin.Context) {
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

	receiver := coremodels.WebhookReceiver{}
	if names, nErr := receiver.GetNamesByIds(m.Webhooks); nErr == nil {
		m.WebhookNames = strings.Join(names, ",")
	} else {
		amis.WriteJsonError(c, nErr)
		return
	}

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
func (s *Controller) Delete(c *gin.Context) {
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
func (s *Controller) QuickSave(c *gin.Context) {
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

