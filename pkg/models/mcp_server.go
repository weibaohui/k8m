package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

// MCPServerConfig MCP服务器配置
type MCPServerConfig struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	URL       string    `gorm:"not null" json:"url,omitempty"`
	Name      string    `gorm:"uniqueIndex;not null;type:varchar(255)" json:"name,omitempty"`
	Enabled   bool      `gorm:"default:false" json:"enabled,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

func (c *MCPServerConfig) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*MCPServerConfig, int64, error) {
	return dao.GenericQuery(params, c, queryFuncs...)
}

func (c *MCPServerConfig) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, c, queryFuncs...)
}

func (c *MCPServerConfig) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, c, utils.ToInt64Slice(ids), queryFuncs...)
}

func (c *MCPServerConfig) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*MCPServerConfig, error) {
	return dao.GenericGetOne(params, c, queryFuncs...)
}
