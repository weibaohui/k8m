package models

import (
	"encoding/base64"
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

// KubeConfig 用户导入kubeconfig
type KubeConfig struct {
	ID          uint   `gorm:"primaryKey;autoIncrement" json:"id,omitempty"` // 模板 ID，主键，自增
	Content     string `gorm:"type:text" json:"content,omitempty"`           // 模板内容，支持大文本存储
	Server      string `json:"server,omitempty"`
	User        string `json:"user,omitempty"`
	Cluster     string `json:"cluster,omitempty"` // 类型，最大长度 100
	Namespace   string `json:"namespace,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	// aws 集群相关
	AccessKey       string `json:"-"`                    // AWS Access Key ID
	SecretAccessKey string `json:"-"`                    // AWS Secret Access Key
	ClusterName     string `json:"cluster_name"`         // AWS EKS 集群名称
	Region          string `json:"region"`               // AWS 区域
	IsAWSEKS        bool   `json:"is_aws_eks,omitempty"` // 标识是否为AWS EKS集群
	// token 纳管相关 server\token\cadata
	Token  string `gorm:"type:text" json:"token,omitempty"`   // token 内容，支持大文本存储
	CACert string `gorm:"type:text" json:"ca_data,omitempty"` // ca 证书内容，支持大文本存储

	CreatedAt time.Time `json:"created_at,omitempty" gorm:"<-:create"`
	UpdatedAt time.Time `json:"updated_at,omitempty"` // Automatically managed by GORM for update time
}

func (c *KubeConfig) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*KubeConfig, int64, error) {
	return dao.GenericQuery(params, c, queryFuncs...)
}

func (c *KubeConfig) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, c, queryFuncs...)
}

func (c *KubeConfig) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, c, utils.ToInt64Slice(ids), queryFuncs...)
}

func (c *KubeConfig) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*KubeConfig, error) {
	return dao.GenericGetOne(params, c, queryFuncs...)
}

// BeforeSave 在保存前加密敏感字段
func (c *KubeConfig) BeforeSave(tx *gorm.DB) error {
	if c.AccessKey != "" {
		encrypted, err := encryptField(c.AccessKey)
		if err != nil {
			return err
		}
		c.AccessKey = encrypted
	}
	if c.SecretAccessKey != "" {
		encrypted, err := encryptField(c.SecretAccessKey)
		if err != nil {
			return err
		}
		c.SecretAccessKey = encrypted
	}
	return nil
}

// AfterFind 在查询后解密敏感字段
func (c *KubeConfig) AfterFind(tx *gorm.DB) error {
	if c.AccessKey != "" {
		decrypted, err := decryptField(c.AccessKey)
		if err != nil {
			return err
		}
		c.AccessKey = decrypted
	}
	if c.SecretAccessKey != "" {
		decrypted, err := decryptField(c.SecretAccessKey)
		if err != nil {
			return err
		}
		c.SecretAccessKey = decrypted
	}
	return nil
}

// encryptField 加密字段
func encryptField(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	encrypted, err := utils.AesEncrypt([]byte(plaintext))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// decryptField 解密字段
func decryptField(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}
	decrypted, err := utils.AesDecrypt(ciphertext)
	if err != nil {
		return "", err
	}
	return string(decrypted), nil
}
