package backend

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
)

// Item 演示数据模型（使用数据库存储）
type Item struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `json:"created_at,omitempty" gorm:"<-:create"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
	CreatedBy   string    `json:"created_by,omitempty" gorm:"size:255"`
}

// TableName 使用插件名前缀
func (Item) TableName() string {
	return "demo_items"
}

// InitDB 初始化数据库表（GORM自动迁移）
func InitDB() error {
	return dao.DB().AutoMigrate(&Item{})
}

