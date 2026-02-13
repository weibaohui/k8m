package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

type HelmChart struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	RepositoryID   uint      `gorm:"index:idx_helm_chart_repository_id;not null" json:"repository_id,omitempty"` // 关联仓库ID
	RepositoryName string    `gorm:"size:255" json:"repository_name,omitempty"`                       // 关联仓库名称
	Name           string    `gorm:"size:255;index:idx_helm_chart_name;not null" json:"name,omitempty"`          // Chart名称
	LatestVersion  string    `gorm:"size:64" json:"latest_version,omitempty"`                         // 最新版本（冗余字段，优化查询）
	Description    string    `gorm:"type:text" json:"description,omitempty"`                          // Chart描述
	Home           string    `gorm:"size:255" json:"home,omitempty"`                                   // 项目主页URL
	Icon           string    `gorm:"size:255" json:"icon,omitempty"`                                   // Chart图标链接
	Keywords       string    `gorm:"type:text" json:"keywords,omitempty"`                              // 关键词（PostgreSQL数组类型）
	KubeVersion    string    `gorm:"size:64" json:"kubeVersion,omitempty"`                             // 最低k8s版本要求
	AppVersion     string    `gorm:"size:64" json:"appVersion,omitempty"`                              // app应用版本
	Deprecated     bool      `json:"deprecated,omitempty"`                                             // Whether or not this chart is deprecated
	CreatedAt      time.Time `json:"created_at,omitempty" gorm:"<-:create"`
	UpdatedAt      time.Time `json:"updated_at"`
	Sources        string    `gorm:"type:text" json:"sources,omitempty"` // 源码主页
}

func (c *HelmChart) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*HelmChart, int64, error) {
	return dao.GenericQuery(params, c, queryFuncs...)
}

// BatchSave 批量保存 HelmChart 实例
func (c *HelmChart) BatchSave(params *dao.Params, events []*HelmChart, batchSize int, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericBatchSave(params, events, batchSize, queryFuncs...)
}
func (c *HelmChart) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, c, queryFuncs...)
}

func (c *HelmChart) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, c, utils.ToInt64Slice(ids), queryFuncs...)
}

func (c *HelmChart) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*HelmChart, error) {
	return dao.GenericGetOne(params, c, queryFuncs...)
}

// Release Helm Release 信息（来自 helm list 命令）
type Release struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	Revision    string `json:"revision"`
	Updated     string `json:"updated"`
	Status      string `json:"status"`
	Chart       string `json:"chart"`
	AppVersion  string `json:"app_version"`
	Description string `json:"description"`
}

// ReleaseHistory Helm Release 历史信息（来自 helm history 命令）
type ReleaseHistory struct {
	Revision   int    `json:"revision"`
	Updated    string `json:"updated"`
	Status     string `json:"status"`
	Chart      string `json:"chart"`
	AppVersion string `json:"app_version"`
}
