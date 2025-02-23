package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

// User 用户导入User
type User struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"` // 模板 ID，主键，自增
	Username  string    `gorm:"uniqueIndex;not null" json:"username,omitempty"`
	Password  string    `gorm:"not null;index:idx_password" json:"password,omitempty"`
	Role      string    `gorm:"not null;index:idx_role" json:"role,omitempty"` // 管理员/只读
	Salt      string    `gorm:"not null" json:"salt,omitempty"`
	CreatedBy string    `gorm:"index:idx_created_by" json:"created_by,omitempty"` // 创建者
	CreatedAt time.Time `json:"created_at,omitempty"`                             // Automatically managed by GORM for creation time
	UpdatedAt time.Time `json:"updated_at,omitempty"`                             // Automatically managed by GORM for update time
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
