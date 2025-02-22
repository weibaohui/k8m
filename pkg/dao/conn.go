package dao

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/weibaohui/k8m/pkg/flag"
	"gorm.io/driver/sqlite" // Sqlite driver based on CGO
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"k8s.io/klog/v2"
)

// 定义全局变量
var (
	once       sync.Once
	dbInstance *gorm.DB
	dbErr      error
)

// connDB 返回数据库连接的单例
func connDB() (*gorm.DB, error) {
	once.Do(func() {
		customLogger := logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold:             time.Second, // 慢 SQL 阈值
				LogLevel:                  logger.Info, // 日志级别
				IgnoreRecordNotFoundError: true,        // 忽略记录未找到错误
				Colorful:                  true,        // 禁用彩色打印
			},
		)

		cfg := flag.Init()

		db, err := gorm.Open(sqlite.Open(cfg.SqlitePath), &gorm.Config{
			Logger: customLogger,
		})
		if err != nil {
			dbErr = err
			return
		}
		klog.V(4).Infof("已连接数据库.")

		dbInstance = db
	})

	return dbInstance, dbErr
}
func DB() *gorm.DB {
	_, err := connDB()
	if err != nil {
		klog.Errorf("数据库连接失败:%v", err)
	}
	return dbInstance
}
