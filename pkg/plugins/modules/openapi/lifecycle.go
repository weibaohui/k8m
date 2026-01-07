package openapi

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules/openapi/models"
	"k8s.io/klog/v2"
)

type OpenAPILifecycle struct{}

func (o *OpenAPILifecycle) Install(ctx plugins.InstallContext) error {
	klog.V(6).Infof("开始安装OpenAPI插件")
	if err := models.InitDB(); err != nil {
		klog.V(6).Infof("安装OpenAPI插件失败，初始化数据库失败: %v", err)
		return err
	}
	klog.V(6).Infof("安装OpenAPI插件成功")
	return nil
}

func (o *OpenAPILifecycle) Upgrade(ctx plugins.UpgradeContext) error {
	klog.V(6).Infof("升级OpenAPI插件：从版本 %s 到版本 %s", ctx.FromVersion(), ctx.ToVersion())
	if err := models.UpgradeDB(ctx.FromVersion(), ctx.ToVersion()); err != nil {
		klog.V(6).Infof("升级OpenAPI插件失败: %v", err)
		return err
	}
	return nil
}

func (o *OpenAPILifecycle) Enable(ctx plugins.EnableContext) error {
	klog.V(6).Infof("启用OpenAPI插件")
	return nil
}

func (o *OpenAPILifecycle) Disable(ctx plugins.BaseContext) error {
	klog.V(6).Infof("禁用OpenAPI插件")
	return nil
}

func (o *OpenAPILifecycle) Uninstall(ctx plugins.UninstallContext) error {
	klog.V(6).Infof("开始卸载OpenAPI插件")

	if !ctx.KeepData() {
		if err := models.DropDB(); err != nil {
			klog.V(6).Infof("卸载OpenAPI插件失败: %v", err)
			return err
		}
	}
	klog.V(6).Infof("卸载OpenAPI插件成功")
	return nil
}

func (o *OpenAPILifecycle) Start(ctx plugins.BaseContext) error {
	klog.V(6).Infof("启动OpenAPI插件成功")
	return nil
}

func (o *OpenAPILifecycle) StartCron(ctx plugins.BaseContext, spec string) error {
	return nil
}

func (o *OpenAPILifecycle) Stop(ctx plugins.BaseContext) error {
	klog.V(6).Infof("停止OpenAPI插件后台任务")
	return nil
}
