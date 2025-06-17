package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

type WebhookReceiver struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	Platform      string    `json:"platform,omitempty"` // feishu,dingtalk
	TargetURL     string    `json:"target_url,omitempty"`
	Method        string    `json:"method,omitempty"`
	Template      string    `gorm:"type:text" json:"template,omitempty"`
	SignSecret    string    `json:"sign_secret,omitempty"`
	SignAlgo      string    `json:"sign_algo,omitempty"`       // e.g. "hmac-sha256", "feishu"
	SignHeaderKey string    `json:"sign_header_key,omitempty"` // e.g. "X-Signature" or unused
	CreatedAt     time.Time `json:"created_at,omitempty"`      // 创建时间
	UpdatedAt     time.Time `json:"updated_at,omitempty"`
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
