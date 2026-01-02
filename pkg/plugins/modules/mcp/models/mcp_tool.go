package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

type MCPTool struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	ServerName  string    `json:"server_name,omitempty"`
	Name        string    `gorm:"uniqueIndex;not null;type:varchar(255)" json:"name,omitempty"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	InputSchema string    `gorm:"type:text" json:"input_schema,omitempty"`
	Enabled     bool      `gorm:"default:true" json:"enabled,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty" gorm:"<-:create"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
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

func (c *MCPTool) ListByServer(db *gorm.DB, serverName string) ([]*MCPTool, int64, error) {
	var items []*MCPTool
	var total int64
	query := db.Where("server_name = ?", serverName)
	if err := query.Model(c).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
