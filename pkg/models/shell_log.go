package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

// ShellLog 用户导入ShellLog
type ShellLog struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"` // 模板 ID，主键，自增
	UserName      string    `json:"username,omitempty"`
	Cluster       string    `json:"cluster,omitempty"`
	Namespace     string    `json:"namespace,omitempty"`
	PodName       string    `json:"pod_name,omitempty"`
	ContainerName string    `json:"container_name,omitempty"`
	Command       string    `json:"command,omitempty"` // shell 执行命令
	Role          string    `json:"role,omitempty"`
	CreatedAt     time.Time `json:"created_at,omitempty"` // Automatically managed by GORM for creation time
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
