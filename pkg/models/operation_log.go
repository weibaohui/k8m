package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

// OperationLog 用户导入OperationLog
type OperationLog struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"` // 模板 ID，主键，自增
	Cluster       string    `gorm:"index" json:"cluster,omitempty"`
	Namespace     string    `json:"namespace,omitempty"`
	Name          string    `json:"name,omitempty"`
	UserName      string    `json:"username,omitempty"`
	Group         string    `json:"group,omitempty"`
	Kind          string    `json:"kind,omitempty"`
	Action        string    `json:"action,omitempty"` // 操作类型
	ContainerName string    `json:"container_name,omitempty"`
	CreatedAt     time.Time `json:"created_at,omitempty"` // Automatically managed by GORM for creation time
	UpdatedAt     time.Time `json:"updated_at,omitempty"` // Automatically managed by GORM for update time

}

func (c *OperationLog) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*OperationLog, int64, error) {
	return dao.GenericQuery(params, c, queryFuncs...)
}

func (c *OperationLog) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, c, queryFuncs...)
}

func (c *OperationLog) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, c, utils.ToInt64Slice(ids), queryFuncs...)
}

func (c *OperationLog) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*OperationLog, error) {
	return dao.GenericGetOne(params, c, queryFuncs...)
}
