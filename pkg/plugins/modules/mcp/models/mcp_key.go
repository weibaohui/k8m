package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

type McpKey struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	Username    string    `gorm:"index;not null" json:"username,omitempty"`
	McpKey      string    `gorm:"type:text" json:"mcp_key,omitempty"`
	Description string    `json:"description,omitempty"`
	Enabled     bool      `gorm:"default:true" json:"enabled,omitempty"`
	Jwt         string    `gorm:"type:text" json:"jwt"`
	LastUsedAt  time.Time `json:"last_used_at,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty" gorm:"<-:create"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
	CreatedBy   string    `json:"created_by,omitempty"`
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

func (c *McpKey) ListByUser(db *gorm.DB, username string) ([]*McpKey, int64, error) {
	var items []*McpKey
	var total int64
	query := db.Where("username = ?", username)
	if err := query.Model(c).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
