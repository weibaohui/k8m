package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

type WebhookReceiver struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	Name         string    `json:"name,omitempty"`     // webhook名称
	Platform     string    `json:"platform,omitempty"` // feishu,dingtalk
	TargetURL    string    `json:"target_url,omitempty"`
	BodyTemplate string    `gorm:"type:text" json:"body_template,omitempty"` // 发送到webhook的body模板
	SignSecret   string    `json:"sign_secret,omitempty"`
	CreatedAt    time.Time `json:"created_at,omitempty" gorm:"<-:create"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"` // Automatically managed by GORM for update time
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

// GetNamesByIds 根据webhook ID列表获取对应的名称列表
func (c *WebhookReceiver) GetNamesByIds(ids []string) ([]string, error) {
	receivers, _, err := c.List(dao.BuildDefaultParams(), func(db *gorm.DB) *gorm.DB {
		return db.Select("name").Where("id in ?", ids)
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

// GetReceiversByIds 根据逗号分隔的ID字符串查询webhook接收器列表
// 参数：ids - 逗号分隔的webhook ID字符串，例如 "1,2,3"
// 返回：webhook接收器列表和错误信息
// 如果ids为空，返回空列表
func (c *WebhookReceiver) GetReceiversByIds(ids []string) ([]*WebhookReceiver, error) {
	// 检查 ids 是否为空
	if len(ids) == 0 {
		return []*WebhookReceiver{}, nil
	}

	receiver := &WebhookReceiver{}
	receivers, _, err := receiver.List(dao.BuildDefaultParams(), func(db *gorm.DB) *gorm.DB {
		return db.Where("id in ?", ids)
	})
	if err != nil {
		return nil, err
	}
	return receivers, nil
}
