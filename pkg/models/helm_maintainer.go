package models

import (
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

type HelmMaintainer struct {
	ID      uint   `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	ChartID uint   `gorm:"index" json:"chart_id,omitempty"`
	Name    string `gorm:"not null" json:"name,omitempty"`
	Email   string `json:"email,omitempty"`
	URL     string `json:"url,omitempty"`
}

func (c *HelmMaintainer) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*HelmMaintainer, int64, error) {
	return dao.GenericQuery(params, c, queryFuncs...)
}

func (c *HelmMaintainer) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, c, queryFuncs...)
}

func (c *HelmMaintainer) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, c, utils.ToInt64Slice(ids), queryFuncs...)
}

func (c *HelmMaintainer) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*HelmMaintainer, error) {
	return dao.GenericGetOne(params, c, queryFuncs...)
}
