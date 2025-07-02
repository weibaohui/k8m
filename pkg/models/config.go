package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

type Config struct {
	ID                   uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	ProductName          string    `json:"product_name,omitempty"` // 产品名称
	LoginType            string    `json:"login_type,omitempty"`
	JwtTokenSecret       string    `json:"jwt_token_secret,omitempty"`
	NodeShellImage       string    `json:"node_shell_image,omitempty"`
	KubectlShellImage    string    `json:"kubectl_shell_image,omitempty"`
	ImagePullTimeout     int       `gorm:"default:30" json:"image_pull_timeout,omitempty"` // 镜像拉取超时时间（秒）
	AnySelect            bool      `gorm:"default:true" json:"any_select"`
	PrintConfig          bool      `json:"print_config"`
	EnableAI             bool      `gorm:"default:true" json:"enable_ai"` // 是否启用AI功能，默认开启
	UseBuiltInModel      bool      `gorm:"default:true" json:"use_built_in_model"`
	MaxIterations        int32     `json:"max_iterations"`                                     //  模型自动对话的最大轮数
	MaxHistory           int32     `json:"max_history"`                                        //  模型对话上下文历史记录数
	ResourceCacheTimeout int       `gorm:"default:60" json:"resource_cache_timeout,omitempty"` // 资源缓存时间（秒）
	ModelID              uint      `json:"model_id"`
	CreatedAt            time.Time `json:"created_at,omitempty"` // Automatically managed by GORM for creation time
	UpdatedAt            time.Time `json:"updated_at,omitempty"` // Automatically managed by GORM for update time
}

func (c *Config) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*Config, int64, error) {
	return dao.GenericQuery(params, c, queryFuncs...)
}

func (c *Config) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, c, queryFuncs...)
}

func (c *Config) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, c, utils.ToInt64Slice(ids), queryFuncs...)
}

func (c *Config) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*Config, error) {
	return dao.GenericGetOne(params, c, queryFuncs...)
}
