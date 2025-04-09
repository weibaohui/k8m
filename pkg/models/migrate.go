package models

import (
	"fmt"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/flag"
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
	_ = AddInnerMCPServer()
	_ = FixRoleName()
	_ = InitConfigTable()
}
func AutoMigrate() error {

	// 添加需要迁移的所有模型
	err := dao.DB().AutoMigrate(
		&CustomTemplate{},
		&KubeConfig{},
		&User{},
		&ClusterUserRole{},
		&OperationLog{},
		&ShellLog{},
		&HelmRepository{},
		&HelmChart{},
		&UserGroup{},
		&MCPServerConfig{},
		&Config{},
	)
	if err != nil {
		klog.Errorf("数据库迁移报错: %v", err.Error())
	}
	// 删除 user 表 name 字段，已弃用
	err = dao.DB().Migrator().DropColumn(&User{}, "Role")
	if err != nil {
		klog.Errorf("数据库迁移 User 表 DropColumn Role 报错: %v", err.Error())
	}
	return nil
}
func FixRoleName() error {
	// 将用户组表中角色进行统一，除了平台管理员以外，都更新为普通用户guest
	result := dao.DB().Model(&UserGroup{}).Where("role != ?", "platform_admin").Update("role", "guest")
	if result.Error != nil {
		klog.Errorf("更新用户组表中角色失败: %v", result.Error)
		return result.Error
	}

	return nil
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
func AddInnerMCPServer() error {
	// 检查是否存在名为k8m的记录
	var count int64
	if err := dao.DB().Model(&MCPServerConfig{}).Where("name = ?", "k8m").Count(&count).Error; err != nil {
		klog.Errorf("查询MCP服务器配置失败: %v", err)
		return err
	}
	cfg := flag.Init()
	// 如果不存在，添加默认的内部MCP服务器配置
	if count == 0 {
		config := &MCPServerConfig{
			Name:      "k8m",
			URL:       fmt.Sprintf("http://localhost:%d/sse", cfg.MCPServerPort),
			Enabled:   true,
			CreatedBy: "system",
		}
		if err := dao.DB().Create(config).Error; err != nil {
			klog.Errorf("添加内部MCP服务器配置失败: %v", err)
			return err
		}
		klog.V(4).Info("成功添加内部MCP服务器配置")
	}

	return nil
}
func InitConfigTable() error {
	var count int64
	if err := dao.DB().Model(&Config{}).Count(&count).Error; err != nil {
		klog.Errorf("查询配置表: %v", err)
		return err
	}
	if count == 0 {
		config := &Config{
			PrintConfig: false,
			EnableAI:    true,
			AnySelect:   true,
			LoginType:   "password",
		}
		if err := dao.DB().Create(config).Error; err != nil {
			klog.Errorf("初始化配置表失败: %v", err)
			return err
		}
		klog.V(4).Info("成功初始化配置表")
	}

	return nil
}
