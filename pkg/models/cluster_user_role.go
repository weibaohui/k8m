package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

// ClusterUserRole 集群用户权限表
type ClusterUserRole struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	Cluster    string    `gorm:"index" json:"cluster,omitempty"`    // 集群名称
	Username   string    `gorm:"index" json:"username,omitempty"`   // 用户名
	Role       string    `gorm:"index" json:"role,omitempty"`       // 角色类型：只读、读写、Exec
	Namespaces string    `json:"namespaces,omitempty"`              // Namespaces列表，逗号分割 ，该用户可以访问的Ns
	CreatedBy  string    `gorm:"index" json:"created_by,omitempty"` // 创建者
	CreatedAt  time.Time `json:"created_at,omitempty"`
	UpdatedAt  time.Time `json:"updated_at,omitempty"`
}

func (c *ClusterUserRole) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*ClusterUserRole, int64, error) {
	return dao.GenericQuery(params, c, queryFuncs...)
}

func (c *ClusterUserRole) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, c, queryFuncs...)
}

func (c *ClusterUserRole) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, c, utils.ToInt64Slice(ids), queryFuncs...)
}

func (c *ClusterUserRole) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*ClusterUserRole, error) {
	return dao.GenericGetOne(params, c, queryFuncs...)
}
