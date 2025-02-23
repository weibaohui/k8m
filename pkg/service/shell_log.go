package service

import (
	"github.com/weibaohui/k8m/pkg/models"
)

type shellLogService struct {
}

func (s *shellLogService) Add(m *models.ShellLog) {
	_ = m.Save(nil)
}
