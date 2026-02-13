package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

// ShellLog 用户导入ShellLog
type ShellLog struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`      // 模板 ID，主键，自增
	UserName      string    `gorm:"size:255;index:idx_username" json:"username,omitempty"`
	Cluster       string    `gorm:"size:100;index:idx_cluster" json:"cluster,omitempty"`
	Namespace     string    `gorm:"size:100;index:idx_namespace" json:"namespace,omitempty"`
	PodName       string    `gorm:"size:255;index:idx_pod_name" json:"pod_name,omitempty"`
	ContainerName string    `gorm:"size:255" json:"container_name,omitempty"`
	Command       string    `gorm:"type:text" json:"command,omitempty"` // shell 执行命令
	Role          string    `gorm:"size:50" json:"role,omitempty"`
	CreatedAt     time.Time `json:"created_at,omitempty" gorm:"<-:create"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"` // Automatically managed by GORM for update time
}

func (c *ShellLog) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*ShellLog, int64, error) {
	return dao.GenericQuery(params, c, queryFuncs...)
}

func (c *ShellLog) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, c, queryFuncs...)
}

func (c *ShellLog) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, c, utils.ToInt64Slice(ids), queryFuncs...)
}

func (c *ShellLog) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*ShellLog, error) {
	return dao.GenericGetOne(params, c, queryFuncs...)
}
