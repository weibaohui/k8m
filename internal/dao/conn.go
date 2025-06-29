package dao

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/weibaohui/k8m/pkg/flag"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
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
		if cfg.DBDriver == "sqlite" {
			dbInstance, dbErr = openSqliteDB(cfg, customLogger)
			return
		} else if cfg.DBDriver == "mysql" {
			dbInstance, dbErr = openMysqlDB(cfg, customLogger)
			return
		} else if cfg.DBDriver == "postgresql" {
			dbInstance, dbErr = openPostgresDB(cfg, customLogger)
			return
		}
	})
	return dbInstance, dbErr
}

// openSqliteDB 负责初始化并返回sqlite数据库连接实例。
func openSqliteDB(cfg *flag.Config, customLogger logger.Interface) (*gorm.DB, error) {
	if _, err := os.Stat(cfg.SqlitePath); os.IsNotExist(err) {
		dir := filepath.Dir(cfg.SqlitePath)
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			klog.Errorf("创建数据库文件[%s]失败: %v", dir, err.Error())
			return nil, err
		}
		file, err := os.Create(cfg.SqlitePath)
		defer file.Close()
		if err != nil {
			klog.Errorf("创建数据库文件[%s]失败: %v", cfg.SqlitePath, err.Error())
			return nil, err
		}
	}

	db, err := gorm.Open(sqlite.Open(cfg.SqlitePath), &gorm.Config{
		Logger: customLogger,
	})
	if err != nil {
		return nil, err
	}
	klog.V(4).Infof("已连接数据库[%s].", cfg.SqlitePath)
	s, err := db.DB()
	if err != nil {
		return nil, err
	}
	s.SetMaxOpenConns(1)
	return db, nil
}

// openMysqlDB 负责初始化并返回mysql数据库连接实例。
func openMysqlDB(cfg *flag.Config, customLogger logger.Interface) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&collation=%s&%s",
		cfg.MysqlUser,
		cfg.MysqlPassword,
		cfg.MysqlHost,
		cfg.MysqlPort,
		cfg.MysqlDatabase,
		cfg.MysqlCharset,
		cfg.MysqlCollation,
		cfg.MysqlQuery,
	)
	showDsn := fmt.Sprintf("%s:******@tcp(%s:%d)/%s?charset=%s&collation=%s&%s",
		cfg.MysqlUser,
		cfg.MysqlHost,
		cfg.MysqlPort,
		cfg.MysqlDatabase,
		cfg.MysqlCharset,
		cfg.MysqlCollation,
		cfg.MysqlQuery,
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger:                                   customLogger,
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		klog.Errorf("初始化mysql数据库异常: %v", err)
		return nil, err
	}
	if cfg.MysqlLogMode {
		db = db.Debug()
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Minute * 5)
	klog.V(2).Infof("初始化mysql数据库完成! dsn: %s", showDsn)
	return db, nil
}

// openPostgresDB 负责初始化并返回postgres数据库连接实例。
func openPostgresDB(cfg *flag.Config, customLogger logger.Interface) (*gorm.DB, error) {
	// 构建 PostgreSQL DSN
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		cfg.PgHost,
		cfg.PgUser,
		cfg.PgPassword,
		cfg.PgDatabase,
		cfg.PgPort,
		cfg.PgSSLMode,
		cfg.PgTimeZone,
	)
	showDsn := fmt.Sprintf("host=%s user=%s password=****** dbname=%s port=%d sslmode=%s TimeZone=%s",
		cfg.PgHost,
		cfg.PgUser,
		cfg.PgDatabase,
		cfg.PgPort,
		cfg.PgSSLMode,
		cfg.PgTimeZone,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: customLogger,
	})
	if err != nil {
		klog.Errorf("初始化postgres数据库异常: %v", err)
		return nil, err
	}
	if cfg.PgLogMode {
		db = db.Debug()
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Minute * 5)
	klog.V(2).Infof("初始化postgres数据库完成! dsn: %s", showDsn)
	return db, nil
}

// DB 获取数据库连接实例
func DB() *gorm.DB {
	_, err := connDB()
	if err != nil {
		klog.Errorf("数据库连接失败:%v", err)
	}
	return dbInstance
}
