package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"

	"gorm.io/gorm"
)

type AIRunConfig struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"` // 主键ID
	UseBuiltInModel bool      `gorm:"default:true" json:"use_built_in_model"`       // 是否使用内置模型
	ModelID         uint      `json:"model_id"`                                     // 模型ID
	MaxHistory      int32     `gorm:"default:10" json:"max_history"`                // 最大历史记录数
	MaxIterations   int32     `gorm:"default:10" json:"max_iterations"`             // 最大迭代次数
	AnySelect       bool      `gorm:"default:true" json:"any_select"`               // 是否开启任意选择
	FloatingWindow  bool      `gorm:"default:true" json:"floating_window"`          // 是否开启浮动窗口
	CreatedAt       time.Time `json:"created_at,omitempty" gorm:"<-:create"`        // 创建时间
	UpdatedAt       time.Time `json:"updated_at,omitempty"`                         // 更新时间
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

func (c *AIRunConfig) GetDefault() (*AIRunConfig, error) {
	var config AIRunConfig
	//使用数据库默认值
	err := dao.DB().FirstOrCreate(&config, AIRunConfig{}).Error
	return &config, err
}
