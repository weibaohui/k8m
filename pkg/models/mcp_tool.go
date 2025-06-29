package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

// MCPTool MCP工具配置
type MCPTool struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	ServerName  string    `json:"server_name,omitempty"`                                        // mcp server id，唯一
	Name        string    `gorm:"uniqueIndex;not null;type:varchar(255)" json:"name,omitempty"` // 工具名称，唯一
	Description string    `gorm:"type:text" json:"description,omitempty"`                       // 工具描述
	InputSchema string    `gorm:"type:text" json:"input_schema,omitempty"`                      // 输入模式，JSON格式
	Enabled     bool      `gorm:"default:true" json:"enabled,omitempty"`                        // 是否启用
	CreatedAt   time.Time `json:"created_at,omitempty"`                                         // 创建时间
	UpdatedAt   time.Time `json:"updated_at,omitempty"`                                         // 更新时间
}

func (c *MCPTool) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*MCPTool, int64, error) {
	return dao.GenericQuery(params, c, queryFuncs...)
}

func (c *MCPTool) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, c, queryFuncs...)
}

func (c *MCPTool) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, c, utils.ToInt64Slice(ids), queryFuncs...)
}

func (c *MCPTool) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*MCPTool, error) {
	return dao.GenericGetOne(params, c, queryFuncs...)
}

func (c *MCPTool) BatchSave(params *dao.Params, tools []*MCPTool, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericBatchSave(params, tools, 100, queryFuncs...)
}
