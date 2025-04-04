package service

import (
	"context"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/models"
	"gorm.io/gorm"
)

type configService struct {
	db *gorm.DB
}

func NewConfigService() *configService {
	return &configService{db: dao.DB()}
}

func (s *configService) GetConfig(ctx context.Context) (*models.Config, error) {
	var config models.Config
	if err := s.db.WithContext(ctx).First(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

func (s *configService) UpdateConfig(ctx context.Context, config *models.Config) error {
	return s.db.WithContext(ctx).Save(config).Error
}

func (s *configService) InitDefaultConfig(ctx context.Context) error {
	var count int64
	if err := s.db.WithContext(ctx).Model(&models.Config{}).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return s.db.WithContext(ctx).Create(&models.Config{}).Error
	}
	return nil
}
