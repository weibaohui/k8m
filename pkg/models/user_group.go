package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

type UserGroup struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	GroupName   string    `gorm:"index" json:"group_name,omitempty"`
	Description string    `json:"description,omitempty"`
	Role        string    `gorm:"index" json:"role,omitempty"` // 管理员/只读
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}

func (c *UserGroup) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*UserGroup, int64, error) {
	return dao.GenericQuery(params, c, queryFuncs...)
}

func (c *UserGroup) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, c, queryFuncs...)
}

func (c *UserGroup) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, c, utils.ToInt64Slice(ids), queryFuncs...)
}

func (c *UserGroup) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*UserGroup, error) {
	return dao.GenericGetOne(params, c, queryFuncs...)
}
