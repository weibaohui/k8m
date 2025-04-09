package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

// ApiKey 用户API密钥
type ApiKey struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	Username    string    `gorm:"index;not null" json:"username,omitempty"` // 所属用户
	Key         string    `gorm:"type:text" json:"key,omitempty"`           // API密钥值
	Description string    `json:"description,omitempty"`                    // 描述信息
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`   // Automatically managed by GORM for update time
	CreatedBy   string    `json:"created_by,omitempty"`   // 创建者
	LastUsedAt  time.Time `json:"last_used_at,omitempty"` // 最后使用时间
}

func (c *ApiKey) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*ApiKey, int64, error) {
	return dao.GenericQuery(params, c, queryFuncs...)
}

func (c *ApiKey) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, c, queryFuncs...)
}

func (c *ApiKey) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, c, utils.ToInt64Slice(ids), queryFuncs...)
}

func (c *ApiKey) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*ApiKey, error) {
	return dao.GenericGetOne(params, c, queryFuncs...)
}
