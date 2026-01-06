package service

import (
	"github.com/weibaohui/k8m/pkg/plugins/modules/ai/models"
)

// AIRunConfigService AI运行配置服务
func AIRunConfigService() *aiRunConfigService {
	return &aiRunConfigService{}
}

type aiRunConfigService struct{}

// GetDefault 获取默认的AI运行配置
func (s *aiRunConfigService) GetDefault() (*models.AIRunConfig, error) {
	config := &models.AIRunConfig{}
	return config.GetDefault()
}

// SaveDefault 保存默认的AI运行配置
func (s *aiRunConfigService) SaveDefault(config *models.AIRunConfig) error {
	// 先获取当前配置
	current, err := s.GetDefault()
	if err != nil {
		return err
	}

	// 更新配置
	if current != nil {
		config.ID = current.ID
	}

	return config.Save(nil)
}