package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

// OperationLog 用户导入OperationLog
type OperationLog struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`   // 模板 ID，主键，自增
	UserName     string    `gorm:"size:255;index:idx_operation_log_username" json:"username,omitempty"`
	Role         string    `gorm:"size:50;index:idx_operation_log_role" json:"role,omitempty"`
	Cluster      string    `gorm:"size:100;index:idx_operation_log_cluster" json:"cluster,omitempty"`
	Namespace    string    `gorm:"size:100;index:idx_operation_log_namespace" json:"namespace,omitempty"`
	Name         string    `gorm:"size:255;index:idx_operation_log_name" json:"name,omitempty"`
	Group        string    `gorm:"size:100;index:idx_operation_log_group" json:"group,omitempty"`           // 资源group
	Kind         string    `gorm:"size:100;index:idx_operation_log_kind" json:"kind,omitempty"`             // 资源kind
	Action       string    `gorm:"size:50;index:idx_operation_log_action" json:"action,omitempty"`          // 操作类型
	Params       string    `gorm:"type:text" json:"params,omitempty"`                         // 操作参数
	ActionResult string    `gorm:"type:text" json:"action_result,omitempty"`                 // 操作结果
	CreatedAt    time.Time `json:"created_at,omitempty" gorm:"<-:create"`                     // Automatically managed by GORM for creation time
	UpdatedAt    time.Time `json:"updated_at,omitempty"`                                      // Automatically managed by GORM for update time

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
