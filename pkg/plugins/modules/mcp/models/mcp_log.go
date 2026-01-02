package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

type MCPToolLog struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	ToolName    string    `gorm:"index" json:"tool_name,omitempty"`
	ServerName  string    `gorm:"index" json:"server_name,omitempty"`
	Parameters  string    `gorm:"type:text" json:"parameters,omitempty"`
	Prompt      string    `gorm:"type:text" json:"prompt,omitempty"`
	Result      string    `gorm:"type:text" json:"result,omitempty"`
	Error       string    `gorm:"type:text" json:"error,omitempty"`
	ExecuteTime int64     `json:"execute_time,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty" gorm:"<-:create"`
	CreatedBy   string    `gorm:"index" json:"created_by,omitempty"`
}

func (c *MCPToolLog) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*MCPToolLog, int64, error) {
	return dao.GenericQuery(params, c, queryFuncs...)
}

func (c *MCPToolLog) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, c, queryFuncs...)
}

func (c *MCPToolLog) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, c, utils.ToInt64Slice(ids), queryFuncs...)
}

func (c *MCPToolLog) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*MCPToolLog, error) {
	return dao.GenericGetOne(params, c, queryFuncs...)
}
