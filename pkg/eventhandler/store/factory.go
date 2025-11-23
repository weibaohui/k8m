// Package store 存储工厂函数
package store

import (
	"fmt"

	"k8s.io/klog/v2"
)

// NewStore 创建GORM存储实例
func NewStore() (EventStore, error) {
	gormStore := NewGORMStore()
	if gormStore == nil {
		return nil, fmt.Errorf("创建GORM存储失败")
	}

	if err := gormStore.Init(); err != nil {
		return nil, err
	}

	klog.V(6).Infof("GORM存储初始化成功")
	return gormStore, nil
}
