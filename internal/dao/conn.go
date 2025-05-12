package dao

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/weibaohui/k8m/internal/sqlite"
	"github.com/weibaohui/k8m/pkg/flag"

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

// connDB 初始化并返回全局唯一的 GORM 数据库连接实例。
// 若数据库文件不存在，则自动创建所需目录和文件，并设置最大连接数为 1。
// 返回数据库连接实例和初始化过程中遇到的错误。
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

		// 检测数据库文件是否存在
		if _, err := os.Stat(cfg.SqlitePath); os.IsNotExist(err) {
			// 创建文件夹、数据库文件
			dir := filepath.Dir(cfg.SqlitePath)
			err := os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				klog.Errorf("创建数据库文件[%s]失败: %v", dir, err.Error())
				return
			}
			// 创建数据库文件
			file, err := os.Create(cfg.SqlitePath)
			defer file.Close()

			if err != nil {
				klog.Errorf("创建数据库文件[%s]失败: %v", cfg.SqlitePath, err.Error())
				return
			}
		}
		db, err := gorm.Open(sqlite.Open(cfg.SqlitePath), &gorm.Config{
			Logger: customLogger,
		})
		if err != nil {
			dbErr = err
			return
		}
		klog.V(4).Infof("已连接数据库[%s].", cfg.SqlitePath)
		s, err := db.DB()
		if err != nil {
			dbErr = err
			return
		}
		s.SetMaxOpenConns(1)
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
