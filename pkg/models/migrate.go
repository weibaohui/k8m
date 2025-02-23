package models

import (
	"github.com/weibaohui/k8m/internal/dao"
	"k8s.io/klog/v2"
)

func init() {

	err := AutoMigrate()
	if err != nil {
		klog.Errorf("数据库迁移失败: %v", err.Error())
	}
	klog.V(4).Info("数据库自动迁移完成")
}
func AutoMigrate() error {

	// 添加需要迁移的所有模型
	return dao.DB().AutoMigrate(
		&CustomTemplate{},
		&KubeConfig{},
		&User{},
		&OperationLog{},
		&ShellLog{},
	)
}
