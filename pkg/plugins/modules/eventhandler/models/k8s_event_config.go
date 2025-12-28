package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

// K8sEventConfig 中文函数注释：Event 监听 转发 发送webhook配置表。
type K8sEventConfig struct {
	ID               uint   `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	Name             string `json:"name"`                                // 事件转发配置名称
	Description      string `json:"description"`                         // 事件转发配置描述
	Clusters         string `json:"clusters"`                            // 目标集群列表
	Webhooks         string `json:"webhooks"`                            // webhook列表
	WebhookNames     string `json:"webhook_names"`                       // webhook 名称列表
	Enabled          bool   `json:"enabled"`                             // 是否启用该任务
	AIEnabled        bool   `json:"ai_enabled"`                          // 是否启用AI总结功能
	AIPromptTemplate string `gorm:"type:text" json:"ai_prompt_template"` // AI总结提示词模板

	RuleNamespaces string `json:"rule_namespaces" gorm:"type:text"` // []string 精确匹配命名空间
	RuleNames      string `json:"rule_names" gorm:"type:text"`      // []string 包含匹配名称
	RuleReasons    string `json:"rule_reasons" gorm:"type:text"`    // []string 包含匹配Reason、Message两个字段
	RuleReverse    bool   `json:"rule_reverse" gorm:"default:false"`

	CreatedAt time.Time `json:"created_at,omitempty" gorm:"<-:create"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// List 中文函数注释：返回符合条件的 K8sEventConfig 列表及总数。
func (c *K8sEventConfig) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*K8sEventConfig, int64, error) {
	return dao.GenericQuery(params, c, queryFuncs...)
}

// Save 中文函数注释：保存或更新 K8sEventConfig 实例。
func (c *K8sEventConfig) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, c, queryFuncs...)
}

// Delete 中文函数注释：根据指定 ID 删除 K8sEventConfig 实例。
func (c *K8sEventConfig) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, c, utils.ToInt64Slice(ids), queryFuncs...)
}

// GetOne 中文函数注释：获取单个 K8sEventConfig 实例。
func (c *K8sEventConfig) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*K8sEventConfig, error) {
	return dao.GenericGetOne(params, c, queryFuncs...)
}

