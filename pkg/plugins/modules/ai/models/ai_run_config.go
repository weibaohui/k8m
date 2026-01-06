package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"

	"gorm.io/gorm"
)

type AIRunConfig struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	UseBuiltInModel bool      `gorm:"default:true" json:"use_built_in_model"` // 是否使用内置模型，默认开启
	ModelID         uint      `json:"model_id"`                               // 选择的模型ID
	MaxHistory      int32     `gorm:"default:10" json:"max_history"`          // 模型对话上下文历史记录数
	MaxIterations   int32     `gorm:"default:10" json:"max_iterations"`       // 模型自动对话的最大轮数
	AnySelect       bool      `gorm:"default:true" json:"any_select"`         // 是否开启任意选择
	CreatedAt       time.Time `json:"created_at,omitempty" gorm:"<-:create"`
	UpdatedAt       time.Time `json:"updated_at,omitempty"`
}

func (c *AIRunConfig) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*AIRunConfig, int64, error) {
	return dao.GenericQuery(params, c, queryFuncs...)
}

func (c *AIRunConfig) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, c, queryFuncs...)
}

func (c *AIRunConfig) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, c, utils.ToInt64Slice(ids), queryFuncs...)
}

func (c *AIRunConfig) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*AIRunConfig, error) {
	return dao.GenericGetOne(params, c, queryFuncs...)
}

// GetDefault 获取默认的AI运行配置
func (c *AIRunConfig) GetDefault() (*AIRunConfig, error) {
	var config AIRunConfig
	err := dao.DB().First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果没有配置，创建默认配置
			config = AIRunConfig{
				UseBuiltInModel: true,
				MaxHistory:      10,
				MaxIterations:   10,
				AnySelect:       true,
			}
			err = dao.DB().Create(&config).Error
			if err != nil {
				return nil, err
			}
			return &config, nil
		}
		return nil, err
	}
	return &config, nil
}
