package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

// User 用户导入User
type User struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	Username   string    `gorm:"uniqueIndex;not null" json:"username,omitempty"`
	Salt       string    `gorm:"not null" json:"salt,omitempty"`
	Password   string    `gorm:"not null" json:"password,omitempty"`
	GroupNames string    `json:"group_names"`
	CreatedAt  time.Time `json:"created_at,omitempty"`
	UpdatedAt  time.Time `json:"updated_at,omitempty"`                             // Automatically managed by GORM for update time
	CreatedBy  string    `gorm:"index:idx_created_by" json:"created_by,omitempty"` // 创建者

}

const (
	RoleClusterAdmin    = "cluster_admin"
	RoleClusterReadonly = "cluster_readonly"
	RolePlatformAdmin   = "platform_admin"
)

func (c *User) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*User, int64, error) {

	return dao.GenericQuery(params, c, queryFuncs...)
}

func (c *User) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, c, queryFuncs...)
}

func (c *User) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, c, utils.ToInt64Slice(ids), queryFuncs...)
}

func (c *User) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*User, error) {
	return dao.GenericGetOne(params, c, queryFuncs...)
}
