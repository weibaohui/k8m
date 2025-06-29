package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

type HelmRepository struct {
	ID                    uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	Name                  string    `gorm:"not null" json:"name,omitempty"` // 仓库名称（唯一）
	URL                   string    `gorm:"not null" json:"url,omitempty"`  // 仓库地址（如 https://charts.example.com）
	Type                  string    `gorm:"comment:仓库类型（OCI/HTTP）" json:"type,omitempty"`
	Description           string    `json:"description,omitempty"` // 仓库描述
	AuthType              string    `gorm:"comment:认证类型（Basic/AuthToken/OAuth）" json:"auth_type,omitempty"`
	Username              string    `json:"username,omitempty"` // 认证用户名（加密存储）
	Password              string    `gorm:"-;comment:密码（临时字段，存储时需加密）" json:"password,omitempty"`
	EncryptedSecret       string    `gorm:"comment:加密后的凭据" json:"encrypted_secret,omitempty"`
	IsActive              bool      `gorm:"default:true" json:"is_active,omitempty"` // 是否启用
	Content               string    `json:"content,omitempty"`                       // 模板内容，支持大文本存储
	Generated             string    `json:"generated,omitempty"`                     // repo 索引文件创建时间
	CertFile              string    `json:"certFile,omitempty"`
	KeyFile               string    `json:"keyFile,omitempty"`
	CAFile                string    `json:"caFile,omitempty"`
	InsecureSkipTLSverify bool      `json:"insecure_skip_tls_verify,omitempty"`
	PassCredentialsAll    bool      `json:"pass_credentials_all,omitempty"`
	CreatedAt             time.Time `json:"created_at,omitempty"` // Automatically managed by GORM for creation time
	UpdatedAt             time.Time `json:"updated_at,omitempty"` // Automatically managed by GORM for update time
}

func (c *HelmRepository) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*HelmRepository, int64, error) {
	return dao.GenericQuery(params, c, queryFuncs...)
}

func (c *HelmRepository) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, c, queryFuncs...)
}

func (c *HelmRepository) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, c, utils.ToInt64Slice(ids), queryFuncs...)
}

func (c *HelmRepository) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*HelmRepository, error) {
	return dao.GenericGetOne(params, c, queryFuncs...)
}

func (c *HelmRepository) GetIDByNameAndURL(params *dao.Params) (uint, error) {
	t, err := c.GetOne(params, func(db *gorm.DB) *gorm.DB {
		return db.Select("id").Where("name = ? AND url = ?", c.Name, c.URL).First(c)
	})
	if err != nil {
		return 0, err
	}
	return t.ID, err
}

func (c *HelmRepository) UpdateContent(params *dao.Params) error {
	return c.Save(params, func(db *gorm.DB) *gorm.DB {
		return db.Select([]string{"content", "generated"}).Where("id = ?", c.ID).Save(c)
	})
}
