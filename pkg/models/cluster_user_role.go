package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/constants"
	"gorm.io/gorm"
)

// ClusterUserRole 集群用户权限表
// AuthorizationType有两种类型（user、user_group），如果是用户，那么代表这个人有哪些权限
// 如果是Group，那么代表这个组有哪些权限，这个组可能会有多个用户，那么这多个用户都有相关的权限
type ClusterUserRole struct {
	ID                  uint                               `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	Cluster             string                             `gorm:"index" json:"cluster,omitempty"`  // 集群名称
	Username            string                             `gorm:"index" json:"username,omitempty"` // 用户名
	Role                string                             `gorm:"index" json:"role,omitempty"`     // 角色类型：只读、读写、Exec
	Namespaces          string                             `json:"namespaces,omitempty"`            // Namespaces列表，逗号分割 ，该用户可以访问的Ns
	BlacklistNamespaces string                             `json:"blacklist_namespaces,omitempty"`  // 黑名单Namespaces列表，逗号分割，禁止访问的Ns
	AuthorizationType   constants.ClusterAuthorizationType `json:"authorization_type,omitempty"`    // 用户类型。User\Group两种，默认为User，空为User。Group指用户组
	CreatedAt           time.Time                          `json:"created_at,omitempty"`
	UpdatedAt           time.Time                          `json:"updated_at,omitempty"`
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
