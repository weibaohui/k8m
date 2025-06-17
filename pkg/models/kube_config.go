package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

// KubeConfig 用户导入kubeconfig
type KubeConfig struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"` // 模板 ID，主键，自增
	Content     string    `gorm:"type:text" json:"content,omitempty"`           // 模板内容，支持大文本存储
	Server      string    `gorm:"index" json:"server,omitempty"`
	User        string    `gorm:"index" json:"user,omitempty"`
	Cluster     string    `gorm:"index" json:"cluster,omitempty"` // 模板类型，最大长度 100
	Namespace   string    `gorm:"index" json:"namespace,omitempty"`
	DisplayName string    `gorm:"index" json:"display_name,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"` // Automatically managed by GORM for creation time
	UpdatedAt   time.Time `json:"updated_at,omitempty"` // Automatically managed by GORM for update time
}

func (c *KubeConfig) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*KubeConfig, int64, error) {
	return dao.GenericQuery(params, c, queryFuncs...)
}

func (c *KubeConfig) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, c, queryFuncs...)
}

func (c *KubeConfig) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, c, utils.ToInt64Slice(ids), queryFuncs...)
}

func (c *KubeConfig) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*KubeConfig, error) {
	return dao.GenericGetOne(params, c, queryFuncs...)
}
