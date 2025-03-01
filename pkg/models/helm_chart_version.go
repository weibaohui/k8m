package models

import (
	"time"

	"github.com/lib/pq"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

type HelmChartVersion struct {
	ID            uint           `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	ChartID       uint           `gorm:"index;not null" json:"chart_id,omitempty"`  // 关联Chart ID
	Version       string         `gorm:"index;not null" json:"version,omitempty"`   // SemVer 版本号（如 1.2.3）
	AppVersion    string         `json:"app_version,omitempty"`                     // 应用版本（如 nginx:1.23.0）
	Digest        string         `gorm:"unique;not null" json:"digest,omitempty"`   // Chart包摘要（SHA256）
	URLs          pq.StringArray `gorm:"type:text[]" json:"ur_ls,omitempty"`        // 下载URL列表（兼容多镜像）
	DownloadCount int            `gorm:"default:0" json:"download_count,omitempty"` // 下载次数统计
	CreatedAt     time.Time      `json:"created_at"`
}

func (c *HelmChartVersion) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*HelmChartVersion, int64, error) {
	return dao.GenericQuery(params, c, queryFuncs...)
}

func (c *HelmChartVersion) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, c, queryFuncs...)
}

func (c *HelmChartVersion) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, c, utils.ToInt64Slice(ids), queryFuncs...)
}

func (c *HelmChartVersion) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*HelmChartVersion, error) {
	return dao.GenericGetOne(params, c, queryFuncs...)
}
