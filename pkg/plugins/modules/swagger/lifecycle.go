package swagger

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules/swagger/models"
	"k8s.io/klog/v2"
)

type SwaggerLifecycle struct{}

func (s *SwaggerLifecycle) Install(ctx plugins.InstallContext) error {
	klog.V(6).Infof("安装Swagger插件")
	if err := models.InitDB(); err != nil {
		return err
	}
	return nil
}

func (s *SwaggerLifecycle) Upgrade(ctx plugins.UpgradeContext) error {
	klog.V(6).Infof("升级Swagger插件：从版本 %s 到版本 %s", ctx.FromVersion(), ctx.ToVersion())
	return nil
}

func (s *SwaggerLifecycle) Enable(ctx plugins.EnableContext) error {
	klog.V(6).Infof("启用Swagger插件")
	return nil
}

func (s *SwaggerLifecycle) Disable(ctx plugins.BaseContext) error {
	klog.V(6).Infof("禁用Swagger插件")
	return nil
}

func (s *SwaggerLifecycle) Uninstall(ctx plugins.UninstallContext) error {
	klog.V(6).Infof("卸载Swagger插件")
	return nil
}

func (s *SwaggerLifecycle) Start(ctx plugins.BaseContext) error {
	klog.V(6).Infof("启动Swagger插件")
	return nil
}

func (s *SwaggerLifecycle) StartCron(ctx plugins.BaseContext, spec string) error {
	klog.V(6).Infof("执行Swagger插件定时任务，表达式: %s", spec)
	return nil
}
