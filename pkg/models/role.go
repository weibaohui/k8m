package models

import (
	"encoding/json"
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

// Role 角色模型
type Role struct {
	ID       uint   `gorm:"primaryKey;autoIncrement;comment:主键" json:"id,omitempty"`
	RoleID   string `gorm:"uniqueIndex;not null;size:32;comment:角色ID" json:"role_id,omitempty"`
	RoleName string `gorm:"not null;size:64;comment:角色名称" json:"role_name,omitempty"`
	// 格式为"["name:method"]"
	Operations      string    `gorm:"type:json;comment:操作列表" json:"-"`
	CreatedAt       time.Time `gorm:"comment:创建时间" json:"created_at,omitempty"`
	UpdatedAt       time.Time `gorm:"comment:更新时间" json:"updated_at,omitempty"`
	CreatedBy       string    `gorm:"index;comment:创建者" json:"created_by,omitempty"`
	FrontOperations []string  `json:"frontOperations,omitempty" gorm:"-"`
}

func (r *Role) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*Role, int64, error) {
	return dao.GenericQuery(params, r, queryFuncs...)
}

func (r *Role) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, r, queryFuncs...)
}

func (r *Role) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, r, utils.ToInt64Slice(ids), queryFuncs...)
}

func (r *Role) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*Role, error) {
	return dao.GenericGetOne(params, r, queryFuncs...)
}

func (r *Role) GetOperations() ([]string, error) {
	var operations []string
	err := json.Unmarshal([]byte(r.Operations), &operations)
	return operations, err
}
