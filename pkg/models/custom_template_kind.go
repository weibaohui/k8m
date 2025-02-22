package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
)

// CustomTemplateKind 表示用户自定义模板分类表
type CustomTemplateKind struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"` // 模板 ID，主键，自增
	Name      string    `gorm:"index" json:"name,omitempty"`                  // 模板名称，非空，最大长度 255
	CreatedBy string    `gorm:"index" json:"created_by,omitempty"`            // 创建者
	CreatedAt time.Time `json:"created_at" json:"created_at"`                 // Automatically managed by GORM for creation time
	UpdatedAt time.Time `json:"updated_at" json:"updated_at"`                 // Automatically managed by GORM for update time
}

func (c *CustomTemplateKind) List(params *dao.Params) ([]*CustomTemplateKind, int64, error) {
	return dao.GenericQuery(params, c)
}
func (c *CustomTemplateKind) Save(params *dao.Params) error {
	return dao.GenericSave(params, c)
}

func (c *CustomTemplateKind) Delete(params *dao.Params, ids string) error {
	return dao.GenericDelete(params, c, utils.ToInt64Slice(ids))
}

func (c *CustomTemplateKind) GetOne(params *dao.Params) (*CustomTemplateKind, error) {
	return dao.GenericGetOne(params, c)
}
