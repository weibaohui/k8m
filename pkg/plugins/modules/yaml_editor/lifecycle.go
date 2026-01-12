package yaml_editor

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules/yaml_editor/models"
	"k8s.io/klog/v2"
)

type YamlEditorLifecycle struct{}

func (l *YamlEditorLifecycle) Install(ctx plugins.InstallContext) error {
	if err := models.InitDB(); err != nil {
		klog.V(6).Infof("安装 YAML 编辑器插件失败: %v", err)
		return err
	}
	klog.V(6).Infof("安装 YAML 编辑器插件成功")
	return nil
}

func (l *YamlEditorLifecycle) Upgrade(ctx plugins.UpgradeContext) error {
	klog.V(6).Infof("升级 YAML 编辑器插件：从版本 %s 到版本 %s", ctx.FromVersion(), ctx.ToVersion())
	if err := models.UpgradeDB(ctx.FromVersion(), ctx.ToVersion()); err != nil {
		klog.V(6).Infof("升级 YAML 编辑器插件失败: %v", err)
		return err
	}
	return nil
}

func (l *YamlEditorLifecycle) Enable(ctx plugins.EnableContext) error {
	klog.V(6).Infof("启用 YAML 编辑器插件")
	return nil
}

func (l *YamlEditorLifecycle) Disable(ctx plugins.BaseContext) error {
	klog.V(6).Infof("禁用 YAML 编辑器插件")
	return nil
}

func (l *YamlEditorLifecycle) Uninstall(ctx plugins.UninstallContext) error {
	if !ctx.KeepData() {
		if err := models.DropDB(); err != nil {
			klog.V(6).Infof("卸载 YAML 编辑器插件失败: %v", err)
			return err
		}
	}
	klog.V(6).Infof("卸载 YAML 编辑器插件成功")
	return nil
}

func (l *YamlEditorLifecycle) Start(ctx plugins.BaseContext) error {
	klog.V(6).Infof("启动 YAML 编辑器插件后台任务")
	return nil
}

func (l *YamlEditorLifecycle) StartCron(ctx plugins.BaseContext, spec string) error {
	return nil
}

func (l *YamlEditorLifecycle) Stop(ctx plugins.BaseContext) error {
	klog.V(6).Infof("停止 YAML 编辑器插件后台任务")
	return nil
}
