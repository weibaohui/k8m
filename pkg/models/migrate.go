package models

import (
	"github.com/weibaohui/k8m/internal/dao"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

func init() {

	err := AutoMigrate()
	if err != nil {
		klog.Errorf("数据库迁移失败: %v", err.Error())
	}
	klog.V(4).Info("数据库自动迁移完成")

	_ = FixClusterName()
}
func AutoMigrate() error {

	// 添加需要迁移的所有模型
	return dao.DB().AutoMigrate(
		&CustomTemplate{},
		&KubeConfig{},
		&User{},
		&OperationLog{},
		&ShellLog{},
		&HelmRepository{},
		&HelmChart{},
		&HelmDependency{},
		&HelmChartVersion{},
		&HelmMaintainer{},
	)
}
func FixClusterName() error {
	// 将display_name为空的记录更新为cluster字段
	result := dao.DB().Model(&KubeConfig{}).Where("display_name = ?", "").Update("display_name", gorm.Expr("cluster"))
	if result.Error != nil {
		klog.Errorf("更新cluster_name失败: %v", result.Error)
		return result.Error
	}
	return nil
}
