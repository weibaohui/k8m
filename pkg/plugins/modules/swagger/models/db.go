package models

import (
	"github.com/weibaohui/k8m/internal/dao"
	"k8s.io/klog/v2"
)

type SwaggerConfig struct {
	ID           uint `gorm:"primaryKey;autoIncrement" json:"id"`
	Enabled      bool `gorm:"default:true" json:"enabled"`
}

func (s *SwaggerConfig) TableName() string {
	return "plugin_swagger_config"
}

func InitDB() error {
	return dao.DB().AutoMigrate(&SwaggerConfig{})
}

func GetConfig() (*SwaggerConfig, error) {
	var cfg SwaggerConfig
	if err := dao.DB().FirstOrCreate(&cfg).Error; err != nil {
		return nil, err
	}
	return &cfg, nil
}

func UpdateConfig(enabled bool) error {
	cfg, err := GetConfig()
	if err != nil {
		return err
	}
	cfg.Enabled = enabled
	return dao.DB().Save(cfg).Error
}

func IsEnabled() bool {
	cfg, err := GetConfig()
	if err != nil {
		klog.V(6).Infof("获取Swagger插件配置失败，使用默认启用状态: %v", err)
		return true
	}
	return cfg.Enabled
}
