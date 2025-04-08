package service

import (
	"sync"
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/models"
)

type operationLogService struct {
	buffer    []*models.OperationLog
	bufferMux sync.Mutex
	ticker    *time.Ticker
	stopChan  chan bool
}

func NewOperationLogService() *operationLogService {
	s := &operationLogService{
		buffer:   make([]*models.OperationLog, 0, 100),
		ticker:   time.NewTicker(2 * time.Second),
		stopChan: make(chan bool),
	}
	go s.startFlushLoop()
	return s
}

func (s *operationLogService) Add(m *models.OperationLog, params ...any) {
	if len(params) > 0 {
		for _, param := range params {
			m.Params += utils.ToJSON(param)
		}
	}
	s.bufferMux.Lock()
	s.buffer = append(s.buffer, m)
	if len(s.buffer) >= 100 {
		s.flushBuffer()
	}
	s.bufferMux.Unlock()
}

func (s *operationLogService) startFlushLoop() {
	for {
		select {
		case <-s.ticker.C:
			s.bufferMux.Lock()
			s.flushBuffer()
			s.bufferMux.Unlock()
		case <-s.stopChan:
			s.ticker.Stop()
			s.bufferMux.Lock()
			s.flushBuffer()
			s.bufferMux.Unlock()
			return
		}
	}
}

func (s *operationLogService) flushBuffer() {
	if len(s.buffer) == 0 {
		return
	}
	dao.DB().CreateInBatches(s.buffer, 100)
	s.buffer = s.buffer[:0]
}

func (s *operationLogService) Close() {
	s.stopChan <- true
}
