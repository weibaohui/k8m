package models

import (
	"strings"
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

type WebhookReceiver struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	Name          string    `json:"name,omitempty"`     // webhook名称
	Platform      string    `json:"platform,omitempty"` // feishu,dingtalk
	TargetURL     string    `json:"target_url,omitempty"`
	Method        string    `json:"method,omitempty"`
	Template      string    `gorm:"type:text" json:"template,omitempty"`
	SignSecret    string    `json:"sign_secret,omitempty"`
	SignAlgo      string    `json:"sign_algo,omitempty"`       // e.g. "hmac-sha256", "feishu"
	SignHeaderKey string    `json:"sign_header_key,omitempty"` // e.g. "X-Signature" or unused
	CreatedAt     time.Time `json:"created_at,omitempty" gorm:"<-:create"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"` // Automatically managed by GORM for update time
}

func (c *WebhookReceiver) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*WebhookReceiver, int64, error) {
	return dao.GenericQuery(params, c, queryFuncs...)
}

func (c *WebhookReceiver) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, c, queryFuncs...)
}

func (c *WebhookReceiver) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, c, utils.ToInt64Slice(ids), queryFuncs...)
}

func (c *WebhookReceiver) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*WebhookReceiver, error) {
	return dao.GenericGetOne(params, c, queryFuncs...)
}
func (c *WebhookReceiver) ListByRecordID(recordID uint) ([]*WebhookReceiver, error) {
	// 1. 查询 InspectionRecord
	record := &InspectionRecord{}
	record, err := record.GetOne(nil, func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", recordID)
	})
	if err != nil {
		return nil, err

	}
	// 2. 查询 Schedule，查找webhook
	schedule := &InspectionSchedule{}
	schedule, err = schedule.GetOne(nil, func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", record.ScheduleID)
	})
	if err != nil {
		return nil, err

	}

	// 检查 webhooks 字段是否为空
	if strings.TrimSpace(schedule.Webhooks) == "" {
		return []*WebhookReceiver{}, nil
	}
	
	receiver := &WebhookReceiver{}
	receivers, _, err := receiver.List(dao.BuildDefaultParams(), func(db *gorm.DB) *gorm.DB {
		return db.Where("id in ?", strings.Split(schedule.Webhooks, ","))
	})
	if err != nil {
		return nil, err
	}
	return receivers, nil
}
func (c *WebhookReceiver) GetNamesByIds(ids string) ([]string, error) {

	receivers, _, err := c.List(dao.BuildDefaultParams(), func(db *gorm.DB) *gorm.DB {
		return db.Select("name").Where("id in ?", strings.Split(ids, ","))
	})
	if err != nil {
		return nil, err
	}
	var names []string
	for _, receiver := range receivers {
		names = append(names, receiver.Name)
	}
	return names, nil
}
