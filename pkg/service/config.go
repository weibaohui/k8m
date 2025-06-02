package service

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/flag"
	"github.com/weibaohui/k8m/pkg/models"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

type configService struct {
	db *gorm.DB
}

func NewConfigService() *configService {
	return &configService{db: dao.DB()}
}

func (s *configService) GetConfig() (*models.Config, error) {
	var config models.Config
	if err := s.db.First(&config).Error; err != nil {
		return nil, err
	}

	if config.MaxHistory == 0 {
		config.MaxHistory = 10
	}
	if config.MaxIterations == 0 {
		config.MaxIterations = 10
	}

	return &config, nil
}

func (s *configService) UpdateConfig(config *models.Config) error {

	err := s.db.Save(config).Error
	if err != nil {
		return err
	}
	// 保存后，让其生效
	err = s.UpdateFlagFromDBConfig()
	if err != nil {
		return err
	}
	return nil
}

// UpdateFlagFromDBConfig 从数据库中加载配置，更新Flag方法中的值
func (s *configService) UpdateFlagFromDBConfig() error {
	cfg := flag.Init()
	m, err := s.GetConfig()
	if err != nil {
		return err
	}

	cfg.AnySelect = m.AnySelect
	cfg.UseBuiltInModel = m.UseBuiltInModel
	if !m.UseBuiltInModel {
		if m.ModelID == 0 {
			klog.Errorf("UpdateFlagFromDBConfig 未指定有效的模型ID")
			return fmt.Errorf("未指定有效的模型ID")
		}
		// 不使用内置模型，从数据库中加载配置
		modelConfig := &models.AIModelConfig{
			ID: m.ModelID,
		}
		modelConfig, err = modelConfig.GetOne(nil)
		if err != nil {
			return err
		}
		cfg.ApiKey = modelConfig.ApiKey
		cfg.ApiModel = modelConfig.ApiModel
		cfg.ApiURL = modelConfig.ApiURL
		cfg.Think = modelConfig.Think
		if modelConfig.Temperature > 0 {
			cfg.Temperature = modelConfig.Temperature
		}
		if modelConfig.TopP > 0 {
			cfg.TopP = modelConfig.TopP
		}
	}

	// if m.KubeConfig != "" {
	// 	cfg.KubeConfig = m.KubeConfig
	// }
	if m.KubectlShellImage != "" {
		cfg.KubectlShellImage = m.KubectlShellImage
	}
	if m.NodeShellImage != "" {
		cfg.NodeShellImage = m.NodeShellImage
	}
	// 默认为30秒
	if m.ImagePullTimeout != 30 {
		cfg.ImagePullTimeout = m.ImagePullTimeout
	}
	if m.ProductName != "" {
		cfg.ProductName = m.ProductName
	}

	cfg.PrintConfig = m.PrintConfig
	cfg.EnableAI = m.EnableAI
	if m.ResourceCacheTimeout > 0 {
		cfg.ResourceCacheTimeout = m.ResourceCacheTimeout
	}
	if cfg.ResourceCacheTimeout == 0 {
		cfg.ResourceCacheTimeout = 60
	}

	if m.MaxHistory > 0 {
		cfg.MaxHistory = m.MaxHistory
	}
	if m.MaxIterations > 0 {
		cfg.MaxIterations = m.MaxIterations
	}

	// JwtTokenSecret 暂不启用，因为前端也要处理
	// cfg.JwtTokenSecret = m.JwtTokenSecret
	// LoginType 暂不启用，因为就一种password
	// cfg.LoginType = m.LoginType
	if cfg.PrintConfig {
		klog.Infof("已开启配置信息打印选项。下面是数据库配置的回显.\n%s:\n %+v\n%s\n", color.RedString("↓↓↓↓↓↓生产环境请务必关闭↓↓↓↓↓↓"), utils.ToJSON(m), color.RedString("↑↑↑↑↑↑生产环境请务必关闭↑↑↑↑↑↑"))
		cfg.ShowConfigCloseMethod()
	}
	_ = AIService().ResetDefaultClient()
	return nil
}
