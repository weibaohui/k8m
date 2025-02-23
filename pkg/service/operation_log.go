package service

import (
	"github.com/weibaohui/k8m/pkg/models"
)

type operationLogService struct {
}

func (s *operationLogService) Add(m *models.OperationLog) {
	_ = m.Save(nil)
}
