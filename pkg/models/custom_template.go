package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

// CustomTemplate 表示用户自定义模板表的结构体
type CustomTemplate struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"` // 模板 ID，主键，自增
	Name      string    `gorm:"index" json:"name,omitempty"`                  // 模板名称，非空，最大长度 255
	Content   string    `gorm:"type:text" json:"content,omitempty"`           // 模板内容，支持大文本存储
	Kind      string    `gorm:"index" json:"kind,omitempty"`                  // 模板类型，最大长度 100
	Cluster   string    `gorm:"index" json:"cluster,omitempty"`               // 模板类型，最大长度 100
	IsGlobal  bool      `gorm:"index" json:"is_global,omitempty"`             // 模板类型，最大长度 100
	CreatedBy string    `gorm:"index" json:"created_by,omitempty"`            // 创建者
	CreatedAt time.Time `json:"created_at,omitempty"`                         // Automatically managed by GORM for creation time
	UpdatedAt time.Time `json:"updated_at,omitempty"`                         // Automatically managed by GORM for update time
}

func (c *CustomTemplate) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*CustomTemplate, int64, error) {
	return dao.GenericQuery(params, c, queryFuncs...)
}

func (c *CustomTemplate) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, c, queryFuncs...)
}

func (c *CustomTemplate) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, c, utils.ToInt64Slice(ids), queryFuncs...)
}

func (c *CustomTemplate) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*CustomTemplate, error) {
	return dao.GenericGetOne(params, c, queryFuncs...)
}
