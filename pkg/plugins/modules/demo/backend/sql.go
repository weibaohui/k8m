package backend

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"k8s.io/klog/v2"
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

// DropDB 删除Demo插件相关的表及数据
// 通过GORM Migrator执行删除操作，兼容多种数据库
func DropDB() error {
	db := dao.DB()
	if db.Migrator().HasTable(&Item{}) {
		if err := db.Migrator().DropTable(&Item{}); err != nil {
			klog.V(6).Infof("删除 Demo 插件表失败: %v", err)
			return err
		}
		klog.V(6).Infof("已删除 Demo 插件表及数据")
	}
	klog.V(6).Infof("Demo 插件表不存在，无需删除")
	return nil
}
