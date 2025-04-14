package mcp

import (
	"context"
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/models"
)

// LogToolExecution 记录工具执行日志
func (m *MCPHost) LogToolExecution(ctx context.Context, toolName, serverName string, parameters interface{}, result ToolCallResult, executeTime int64) {

	log := &models.MCPToolLog{
		ToolName:    toolName,
		ServerName:  serverName,
		Parameters:  utils.ToJSON(parameters),
		Result:      result.Result,
		ExecuteTime: executeTime,
		CreatedAt:   time.Now(),
		Error:       result.Error,
	}

	username, _ := m.getUserRoleFromMCPCtx(ctx)

	log.CreatedBy = username
	dao.DB().Create(log)
}

func (m *MCPHost) addLog(log *models.MCPToolLog) {

	m.bufferMux.Lock()
	m.buffer = append(m.buffer, log)
	m.bufferMux.Unlock()

	if len(m.buffer) >= 100 {
		m.flushBuffer()
	}
}

func (m *MCPHost) startFlushLoop() {
	for {
		select {
		case <-m.ticker.C:
			m.bufferMux.Lock()
			m.flushBuffer()
			m.bufferMux.Unlock()
		case <-m.stopChan:
			m.ticker.Stop()
			m.bufferMux.Lock()
			m.flushBuffer()
			m.bufferMux.Unlock()
			return
		}
	}
}

func (m *MCPHost) flushBuffer() {
	if len(m.buffer) == 0 {
		return
	}
	dao.DB().CreateInBatches(m.buffer, 100)
	m.buffer = m.buffer[:0]
}
