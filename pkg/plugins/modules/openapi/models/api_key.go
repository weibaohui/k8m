package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

type ApiKey struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"` // 主键ID
	Username    string    `gorm:"index;not null" json:"username,omitempty"`     // 用户名
	Key         string    `gorm:"type:text" json:"key,omitempty"`               // API密钥值
	Description string    `json:"description,omitempty"`                        // 密钥描述信息
	ExpiresAt   time.Time `json:"expires_at,omitempty"`                         // 过期时间
	CreatedAt   time.Time `json:"created_at,omitempty" gorm:"<-:create"`        // 创建时间
	UpdatedAt   time.Time `json:"updated_at,omitempty"`                         // 更新时间
	CreatedBy   string    `json:"created_by,omitempty"`                         // 创建人
	LastUsedAt  time.Time `json:"last_used_at,omitempty"`                       // 最后使用时间
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
