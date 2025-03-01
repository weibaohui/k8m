package models

import (
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

type HelmDependency struct {
	ID         uint   `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	VersionID  uint   `gorm:"index" json:"version_id,omitempty"` // 关联ChartVersion ID
	Name       string `gorm:"not null" json:"name,omitempty"`    // 依赖Chart名称
	Version    string `gorm:"not null" json:"version,omitempty"` // 依赖版本范围（如 ^1.0.0）
	Repository string `json:"repository,omitempty"`              // 依赖仓库地址（可选）
}

func (c *HelmDependency) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*HelmDependency, int64, error) {
	return dao.GenericQuery(params, c, queryFuncs...)
}

func (c *HelmDependency) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, c, queryFuncs...)
}

func (c *HelmDependency) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, c, utils.ToInt64Slice(ids), queryFuncs...)
}

func (c *HelmDependency) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*HelmDependency, error) {
	return dao.GenericGetOne(params, c, queryFuncs...)
}
