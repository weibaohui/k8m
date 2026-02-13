package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

type HelmRelease struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	ReleaseName  string    `gorm:"size:255;index:idx_helm_release_release_name" json:"release_name,omitempty"` // Release 名称
	RepoName     string    `gorm:"size:255" json:"repo_name,omitempty"`                           // 仓库名称
	Namespace    string    `gorm:"size:100;not null;index:idx_helm_release_namespace" json:"namespace,omitempty"` // 命名空间
	ChartName    string    `gorm:"size:255;not null" json:"chart_name,omitempty"`                 // Chart 名称
	ChartVersion string    `gorm:"size:64" json:"chart_version,omitempty"`                        // Chart 版本
	Values       string    `gorm:"type:text" json:"values,omitempty"`                              // values.yaml 内容
	Status       string    `gorm:"size:50" json:"status,omitempty"`                                // 安装状态
	Cluster      string    `gorm:"size:100;index:idx_helm_release_cluster" json:"cluster,omitempty"`
	Result       string    `gorm:"type:text" json:"result,omitempty"` // 描述
	CreatedAt    time.Time `json:"created_at,omitempty" gorm:"<-:create"`
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
func GetHelmReleaseByNsAndReleaseName(namespace, releaseName, cluster string) (*HelmRelease, error) {
	r := &HelmRelease{}
	db := dao.DB().Where("namespace = ? AND release_name = ? AND cluster = ? ", namespace, releaseName, cluster)
	err := db.First(r).Error
	if err != nil {
		return nil, err
	}
	return r, nil
}

// DeleteHelmReleaseByNsAndReleaseName 删除指定namespace和releaseName的HelmRelease记录
func DeleteHelmReleaseByNsAndReleaseName(namespace, releaseName, cluster string) error {
	return dao.DB().Where("namespace = ? AND release_name = ? AND cluster =? ", namespace, releaseName, cluster).Delete(&HelmRelease{}).Error
}
