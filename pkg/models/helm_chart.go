package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

type HelmChart struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	RepositoryID   uint      `gorm:"index;not null" json:"repository_id,omitempty"` // 关联仓库ID
	RepositoryName string    `json:"repository_name,omitempty"`                     // 关联仓库ID
	Name           string    `gorm:"index;not null" json:"name,omitempty"`          // Chart名称
	LatestVersion  string    `json:"latest_version,omitempty"`                      // 最新版本（冗余字段，优化查询）
	Description    string    `json:"description,omitempty"`                         // Chart描述
	Home           string    `json:"home,omitempty"`                                // 项目主页URL
	Icon           string    `json:"icon,omitempty"`                                // Chart图标链接
	Keywords       string    `json:"keywords,omitempty"`                            // 关键词（PostgreSQL数组类型）
	KubeVersion    string    `json:"kubeVersion,omitempty"`                         // 最低k8s版本要求
	AppVersion     string    `json:"appVersion,omitempty"`                          // app应用版本
	Deprecated     bool      `json:"deprecated,omitempty"`                          // Whether or not this chart is deprecated
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Sources        string    `json:"sources,omitempty"` // 源码主页
}

func (c *HelmChart) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*HelmChart, int64, error) {
	return dao.GenericQuery(params, c, queryFuncs...)
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
