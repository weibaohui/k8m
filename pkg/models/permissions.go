package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

// Group 组模型
type Group struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;comment:主键" json:"id,omitempty"`
	GroupID     string    `gorm:"uniqueIndex;not null;size:32;comment:组ID" json:"group_id,omitempty"`
	GroupName   string    `gorm:"not null;size:64;comment:组名称" json:"group_name,omitempty"`
	Description string    `gorm:"type:text;comment:描述" json:"description,omitempty"`
	CreatedAt   time.Time `gorm:"comment:创建时间" json:"created_at,omitempty"`
	UpdatedAt   time.Time `gorm:"comment:更新时间" json:"updated_at,omitempty"`
	CreatedBy   string    `gorm:"index;comment:创建者" json:"created_by,omitempty"`
}

// UserGroupAssociation 用户组关联模型
type UserGroupAssociation struct {
	ID        uint      `gorm:"primaryKey;autoIncrement;comment:主键" json:"id,omitempty"`
	UserID    string    `gorm:"not null;size:32;index;comment:用户ID" json:"user_id,omitempty"`
	GroupID   string    `gorm:"not null;size:32;index;comment:组ID" json:"group_id,omitempty"`
	CreatedAt time.Time `gorm:"comment:创建时间" json:"created_at,omitempty"`
	UpdatedAt time.Time `gorm:"comment:更新时间" json:"updated_at,omitempty"`
}

// ClusterPermission 集群权限绑定模型
type ClusterPermission struct {
	ID         uint      `gorm:"primaryKey;autoIncrement;comment:主键" json:"id,omitempty"`
	BindingID  string    `gorm:"uniqueIndex;not null;size:32;comment:绑定ID" json:"binding_id,omitempty"`
	TargetType string    `gorm:"not null;size:10;comment:目标类型" json:"target_type,omitempty"` // user 或 group
	TargetID   string    `gorm:"not null;size:32;index;comment:目标ID" json:"target_id,omitempty"`
	ClusterID  string    `gorm:"not null;size:32;index;comment:集群ID" json:"cluster_id,omitempty"`
	Namespace  string    `gorm:"not null;size:64;comment:命名空间" json:"namespace,omitempty"`
	RoleID     string    `gorm:"not null;size:32;index;comment:角色ID" json:"role_id,omitempty"`
	CreatedAt  time.Time `gorm:"comment:创建时间" json:"created_at,omitempty"`
	UpdatedAt  time.Time `gorm:"comment:更新时间" json:"updated_at,omitempty"`
	CreatedBy  string    `gorm:"size:64;comment:创建者" json:"created_by,omitempty"`
}

// 用户模型的数据库操作方法
func (u *User) GetByUsername(params *dao.Params, username string) (*User, error) {
	return dao.GenericGetOne(params, u, func(db *gorm.DB) *gorm.DB {
		return db.Where("username = ?", username)
	})
}

func (g *Group) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*Group, int64, error) {
	return dao.GenericQuery(params, g, queryFuncs...)
}

func (g *Group) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, g, queryFuncs...)
}

func (g *Group) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, g, utils.ToInt64Slice(ids), queryFuncs...)
}

func (g *Group) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*Group, error) {
	return dao.GenericGetOne(params, g, queryFuncs...)
}

func (p *ClusterPermission) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*ClusterPermission, int64, error) {
	return dao.GenericQuery(params, p, queryFuncs...)
}

func (p *ClusterPermission) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, p, queryFuncs...)
}

func (p *ClusterPermission) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, p, utils.ToInt64Slice(ids), queryFuncs...)
}

func (p *ClusterPermission) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*ClusterPermission, error) {
	return dao.GenericGetOne(params, p, queryFuncs...)
}

// 集群权限模型的数据库操作方法
func (p *ClusterPermission) ListUserPermissions(params *dao.Params, userID string) ([]*ClusterPermission, error) {
	perms, _, err := dao.GenericQuery[*ClusterPermission](params, p, func(db *gorm.DB) *gorm.DB {
		return db.Where("target_type = ? AND target_id = ?", "user", userID)
	})
	return perms, err
}

func (p *ClusterPermission) ListGroupPermissions(params *dao.Params, groupIDs []string) ([]*ClusterPermission, error) {
	if len(groupIDs) == 0 {
		return []*ClusterPermission{}, nil
	}

	perms, _, err := dao.GenericQuery[*ClusterPermission](params, p, func(db *gorm.DB) *gorm.DB {
		return db.Where("target_type = ? AND target_id IN ?", "group", groupIDs)
	})
	return perms, err
}

func (ug *UserGroupAssociation) GetGr(params *dao.Params, userID string) ([]*UserGroupAssociation, error) {
	groups, _, err := dao.GenericQuery[*UserGroupAssociation](params, ug, func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ?", userID)
	})
	return groups, err
}
