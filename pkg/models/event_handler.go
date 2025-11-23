package models

import (
	"fmt"
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

// K8sEvent 事件处理器使用的K8s事件模型
type K8sEvent struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	EvtKey    string    `gorm:"uniqueIndex;" json:"evt_key"`
	Type      string    `gorm:"type:varchar(16);" json:"type"`
	Reason    string    `gorm:"type:varchar(128);" json:"reason"`
	Level     string    `gorm:"type:varchar(16);" json:"level"`
	Namespace string    `gorm:"type:varchar(64);;index" json:"namespace"`
	Name      string    `gorm:"type:varchar(128);" json:"name"`
	Message   string    `gorm:"type:text;" json:"message"`
	Timestamp time.Time `gorm:"index" json:"timestamp"`
	Processed bool      `gorm:"default:false;index" json:"processed"`
	Attempts  int       `gorm:"default:0" json:"attempts"`
	CreatedAt time.Time `json:"created_at,omitempty" gorm:"<-:create"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// TableName 设置表名
func (K8sEvent) TableName() string {
	return "k8s_events"
}

// IsWarning 判断是否为警告事件
func (e *K8sEvent) IsWarning() bool {
	return e.Type == "Warning" || e.Level == "warning"
}

// GenerateEvtKey 生成事件键
func GenerateEvtKey(namespace, kind, name, reason, message string) string {
	return fmt.Sprintf("%s/%s/%s/%s/%s", namespace, kind, name, reason, message)
}

// List 列出事件记录
// 参数使用统一的 Params 和可选查询方法
func (e *K8sEvent) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*K8sEvent, int64, error) {
	return dao.GenericQuery(params, e, queryFuncs...)
}

// Save 保存事件记录
// 支持根据查询函数限制可更新的字段
func (e *K8sEvent) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, e, queryFuncs...)
}

// Delete 根据ID删除事件记录
func (e *K8sEvent) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, e, utils.ToInt64Slice(ids), queryFuncs...)
}

// GetOne 获取单条事件记录
func (e *K8sEvent) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*K8sEvent, error) {
	return dao.GenericGetOne(params, e, queryFuncs...)
}

// GetByEvtKey 根据事件键获取事件
func (e *K8sEvent) GetByEvtKey(evtKey string) (*K8sEvent, error) {
	var item K8sEvent
	err := dao.DB().Where("evt_key = ?", evtKey).First(&item).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

// MarkProcessedByID 根据ID更新处理状态
func (e *K8sEvent) MarkProcessedByID(id int64, processed bool) error {
	return dao.DB().Model(&K8sEvent{}).Where("id = ?", id).Update("processed", processed).Error
}

// IncrementAttemptsByID 根据ID增加重试次数
func (e *K8sEvent) IncrementAttemptsByID(id int64) error {
	return dao.DB().Model(&K8sEvent{}).Where("id = ?", id).UpdateColumn("attempts", gorm.Expr("attempts + ?", 1)).Error
}

// ListUnprocessed 列出未处理的事件，按时间升序，限制条数
func (e *K8sEvent) ListUnprocessed(limit int) ([]*K8sEvent, error) {
	var list []*K8sEvent
	err := dao.DB().Where("processed = ?", false).Order("timestamp ASC").Limit(limit).Find(&list).Error
	return list, err
}

// UpsertByEvtKey 通过事件键进行插入或更新（幂等）
// 已存在则更新时间与消息，不存在则创建
func (e *K8sEvent) UpsertByEvtKey() error {
	var existing K8sEvent
	err := dao.DB().Where("evt_key = ?", e.EvtKey).First(&existing).Error
	if err == nil {
		return dao.DB().Model(&K8sEvent{}).Where("evt_key = ?", e.EvtKey).Updates(map[string]any{
			"timestamp": e.Timestamp,
			"message":   e.Message,
		}).Error
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	return dao.DB().Create(e).Error
}
