package service

import (
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/models"
)

type ssoConfigService struct{}

var ssoConfigSvc = &ssoConfigService{}

// SSOConfigService 获取SSO配置服务实例
func SSOConfigService() *ssoConfigService {
	return ssoConfigSvc
}

// Get 获取SSO配置
func (s *ssoConfigService) Get() (*models.SSOConfig, error) {
	var config models.SSOConfig
	result := dao.DB().First(&config)
	return &config, result.Error
}

// Update 更新SSO配置
func (s *ssoConfigService) Update(config *models.SSOConfig) error {
	return dao.DB().Save(config).Error
}

// Create 创建SSO配置
func (s *ssoConfigService) Create(config *models.SSOConfig) error {
	return dao.DB().Create(config).Error
}

// Delete 删除SSO配置
func (s *ssoConfigService) Delete(id uint) error {
	return dao.DB().Delete(&models.SSOConfig{}, id).Error
}
