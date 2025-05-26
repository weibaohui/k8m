package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

// McpKey MCP访问密钥
type McpKey struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	Username    string    `gorm:"index;not null" json:"username,omitempty"` // 所属用户
	McpKey      string    `gorm:"type:text" json:"mcp_key,omitempty"`       // MCP密钥值
	Description string    `json:"description,omitempty"`                    // 描述信息
	Enabled     bool      `gorm:"default:true" json:"enabled,omitempty"`    // 是否启用
	Jwt         string    `gorm:"type:text" json:"jwt"`                     //  JWT
	LastUsedAt  time.Time `json:"last_used_at,omitempty"`                   // 最后使用时间
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"` // Automatically managed by GORM for update time
	CreatedBy   string    `json:"created_by,omitempty"` // 创建者
}

func (c *McpKey) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*McpKey, int64, error) {
	return dao.GenericQuery(params, c, queryFuncs...)
}

func (c *McpKey) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, c, queryFuncs...)
}

func (c *McpKey) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, c, utils.ToInt64Slice(ids), queryFuncs...)
}

func (c *McpKey) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*McpKey, error) {
	return dao.GenericGetOne(params, c, queryFuncs...)
}
