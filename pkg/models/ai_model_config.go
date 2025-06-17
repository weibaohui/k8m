package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

// AIModelConfig 用于存储多种AI模型配置
// 支持后续选择不同模型

type AIModelConfig struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ApiKey      string    `json:"api_key"`
	ApiURL      string    `json:"api_url"`
	ApiModel    string    `json:"api_model"`
	Temperature float32   `json:"temperature"`
	TopP        float32   `json:"top_p"`
	Think       bool      `json:"think"` // 是否关闭思考模式
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}

func (c *AIModelConfig) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*AIModelConfig, int64, error) {
	return dao.GenericQuery(params, c, queryFuncs...)
}

func (c *AIModelConfig) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, c, queryFuncs...)
}

func (c *AIModelConfig) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, c, utils.ToInt64Slice(ids), queryFuncs...)
}

func (c *AIModelConfig) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*AIModelConfig, error) {
	return dao.GenericGetOne(params, c, queryFuncs...)
}
