package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

type HelmRelease struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	ReleaseName  string    `gorm:"index" json:"release_name,omitempty"`       // Release 名称
	RepoName     string    `json:"repo_name,omitempty"`                       // 仓库名称
	Namespace    string    `gorm:"not null;index" json:"namespace,omitempty"` // 命名空间
	ChartName    string    `gorm:"not null" json:"chart_name,omitempty"`      // Chart 名称
	ChartVersion string    `json:"chart_version,omitempty"`                   // Chart 版本
	Values       string    `json:"values,omitempty"`                          // values.yaml 内容
	Status       string    `json:"status,omitempty"`                          // 安装状态
	Result       string    `json:"result,omitempty"`                          // 描述
	CreatedAt    time.Time `json:"created_at,omitempty"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
}

func (r *HelmRelease) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*HelmRelease, int64, error) {
	return dao.GenericQuery(params, r, queryFuncs...)
}

func (r *HelmRelease) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, r, queryFuncs...)
}

func (r *HelmRelease) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, r, utils.ToInt64Slice(ids), queryFuncs...)
}

func (r *HelmRelease) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*HelmRelease, error) {
	return dao.GenericGetOne(params, r, queryFuncs...)
}

// GetHelmReleaseByNsAndReleaseName 通过namespace和releaseName获取repo名称
func GetHelmReleaseByNsAndReleaseName(namespace, releaseName string) (*HelmRelease, error) {
	r := &HelmRelease{}
	db := dao.DB().Where("namespace = ? AND release_name = ?", namespace, releaseName)
	err := db.First(r).Error
	if err != nil {
		return nil, err
	}
	return r, nil
}

// DeleteHelmReleaseByNsAndReleaseName 删除指定namespace和releaseName的HelmRelease记录
func DeleteHelmReleaseByNsAndReleaseName(namespace, releaseName string) error {
	return dao.DB().Where("namespace = ? AND release_name = ?", namespace, releaseName).Delete(&HelmRelease{}).Error
}
