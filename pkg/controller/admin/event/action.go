package event

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/eventhandler/worker"
	"github.com/weibaohui/k8m/pkg/models"
	"gorm.io/gorm"
)

// AdminEventController 事件转发配置管理控制器
// 负责事件配置的列表、保存、删除、快速启用/禁用等操作
type AdminEventController struct{}

// RegisterAdminEventRoutes 注册事件配置管理相关路由
// 路由前缀：/admin/event
func RegisterAdminEventRoutes(admin *gin.RouterGroup) {
	ctrl := &AdminEventController{}
	admin.GET("/event/list", ctrl.List)
	admin.POST("/event/save", ctrl.Save)
	admin.POST("/event/delete/:ids", ctrl.Delete)
	admin.POST("/event/save/id/:id/status/:enabled", ctrl.QuickSave)
}

// List 获取事件配置列表
// @Summary 获取事件配置列表
// @Security BearerAuth
// @Success 200 {object} string
// @Router /admin/event/list [get]
func (s *AdminEventController) List(c *gin.Context) {
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

// Save 保存或更新事件配置
// @Summary 保存事件配置
// @Security BearerAuth
// @Success 200 {object} string
// @Router /admin/event/save [post]
func (s *AdminEventController) Save(c *gin.Context) {
	params := dao.BuildParams(c)
	m := models.K8sEventConfig{}
	err := c.ShouldBindJSON(&m)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 验证AI总结配置
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

	// 保存webhook名称快照
	receiver := models.WebhookReceiver{}
	if names, nErr := receiver.GetNamesByIds(m.Webhooks); nErr == nil {
		m.WebhookNames = strings.Join(names, ",")
	} else {
		amis.WriteJsonError(c, nErr)
		return
	}

	// 规范化规则字段：将以逗号分隔的字符串转换为JSON数组字符串
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

	err = m.Save(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 保存成功后，通知事件处理Worker刷新配置，立即生效
	if w := worker.Instance(); w != nil {
		w.UpdateConfig()
	}

	amis.WriteJsonOK(c)
}

// Delete 删除事件配置
// @Summary 删除事件配置
// @Security BearerAuth
// @Param ids path string true "事件配置ID，多个用逗号分隔"
// @Success 200 {object} string
// @Router /admin/event/delete/{ids} [post]
func (s *AdminEventController) Delete(c *gin.Context) {
	ids := c.Param("ids")
	params := dao.BuildParams(c)

	m := &models.K8sEventConfig{}
	err := m.Delete(params, ids)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	// 删除成功后刷新Worker配置
	if w := worker.Instance(); w != nil {
		w.UpdateConfig()
	}
	amis.WriteJsonOK(c)
}

// QuickSave 快速更新事件配置启用状态
// @Summary 快速更新事件配置状态
// @Security BearerAuth
// @Param id path int true "事件配置ID"
// @Param enabled path string true "状态，例如：true、false"
// @Success 200 {object} string
// @Router /admin/event/save/id/{id}/status/{enabled} [post]
func (s *AdminEventController) QuickSave(c *gin.Context) {
	id := c.Param("id")
	enabled := c.Param("enabled")

	var entity models.K8sEventConfig
	entity.ID = utils.ToUInt(id)
	entity.Enabled = enabled == "true"

	err := dao.DB().Model(&entity).Select("enabled").Updates(entity).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	// 快速启用/禁用后刷新Worker配置
	if w := worker.Instance(); w != nil {
		w.UpdateConfig()
	}
	amis.WriteJsonOK(c)
}
