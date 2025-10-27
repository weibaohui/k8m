package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

// WebhookLogRecord webhook发送日志记录
type WebhookLogRecord struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	WebhookID    uint      `json:"webhook_id,omitempty" gorm:"index"`     // webhook接收器ID
	WebhookName  string    `json:"webhook_name,omitempty"`                // webhook名称
	ReceiverID   string    `json:"receiver_id,omitempty"`                 // 接收器ID
	Method       string    `json:"method,omitempty"`                      // HTTP方法
	URL          string    `json:"url,omitempty"`                         // 请求URL
	StatusCode   int       `json:"status_code,omitempty"`                 // 响应状态码
	Success      bool      `json:"success,omitempty" gorm:"index"`        // 是否成功
	Duration     int64     `json:"duration,omitempty"`                    // 请求耗时(纳秒)
	ErrorMessage string    `json:"error_message,omitempty"`               // 错误信息
	Summary      string    `json:"summary,omitempty"`                     // 日志摘要
	Detail       string    `gorm:"type:text" json:"detail,omitempty"`     // 完整日志详情(JSON格式)
	RequestTime  time.Time `json:"request_time,omitempty" gorm:"index"`   // 请求时间
	CreatedAt    time.Time `json:"created_at,omitempty" gorm:"<-:create"` // 创建时间
	UpdatedAt    time.Time `json:"updated_at,omitempty"`                  // 更新时间
}

// List 查询webhook日志列表
func (w *WebhookLogRecord) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*WebhookLogRecord, int64, error) {
	return dao.GenericQuery(params, w, queryFuncs...)
}

// Save 保存webhook日志
func (w *WebhookLogRecord) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, w, queryFuncs...)
}

// Delete 删除webhook日志
func (w *WebhookLogRecord) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, w, utils.ToInt64Slice(ids), queryFuncs...)
}

// GetOne 获取单个webhook日志
func (w *WebhookLogRecord) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*WebhookLogRecord, error) {
	return dao.GenericGetOne(params, w, queryFuncs...)
}

// ListByWebhookID 根据webhook ID查询日志
func (w *WebhookLogRecord) ListByWebhookID(webhookID uint, params *dao.Params) ([]*WebhookLogRecord, int64, error) {
	return w.List(params, func(db *gorm.DB) *gorm.DB {
		return db.Where("webhook_id = ?", webhookID).Order("created_at DESC")
	})
}

// ListBySuccess 根据成功状态查询日志
func (w *WebhookLogRecord) ListBySuccess(success bool, params *dao.Params) ([]*WebhookLogRecord, int64, error) {
	return w.List(params, func(db *gorm.DB) *gorm.DB {
		return db.Where("success = ?", success).Order("created_at DESC")
	})
}

// ListByTimeRange 根据时间范围查询日志
func (w *WebhookLogRecord) ListByTimeRange(startTime, endTime time.Time, params *dao.Params) ([]*WebhookLogRecord, int64, error) {
	return w.List(params, func(db *gorm.DB) *gorm.DB {
		return db.Where("request_time BETWEEN ? AND ?", startTime, endTime).Order("created_at DESC")
	})
}

// GetStatistics 获取webhook发送统计信息
func (w *WebhookLogRecord) GetStatistics(webhookID uint, startTime, endTime time.Time) (map[string]interface{}, error) {
	var result struct {
		Total   int64 `json:"total"`
		Success int64 `json:"success"`
		Failed  int64 `json:"failed"`
	}

	db := dao.DB()
	query := db.Model(&WebhookLogRecord{})

	if webhookID > 0 {
		query = query.Where("webhook_id = ?", webhookID)
	}

	if !startTime.IsZero() && !endTime.IsZero() {
		query = query.Where("request_time BETWEEN ? AND ?", startTime, endTime)
	}

	// 总数
	if err := query.Count(&result.Total).Error; err != nil {
		return nil, err
	}

	// 成功数
	if err := query.Where("success = ?", true).Count(&result.Success).Error; err != nil {
		return nil, err
	}

	// 失败数
	result.Failed = result.Total - result.Success

	// 计算成功率
	successRate := float64(0)
	if result.Total > 0 {
		successRate = float64(result.Success) / float64(result.Total) * 100
	}

	return map[string]interface{}{
		"total":        result.Total,
		"success":      result.Success,
		"failed":       result.Failed,
		"success_rate": successRate,
	}, nil
}
