package ai

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules/ai/models"
	"github.com/weibaohui/k8m/pkg/plugins/modules/ai/service"
	"k8s.io/klog/v2"
)

type AILifecycle struct{}

func (l *AILifecycle) Install(ctx plugins.InstallContext) error {
	if err := models.InitDB(); err != nil {
		klog.V(6).Infof("安装 AI 插件失败: %v", err)
		return err
	}
	if err := models.MigrateAIModel(); err != nil {
		klog.V(6).Infof("迁移 AI 模型配置失败: %v", err)
		return err
	}
	if err := models.InitBuiltinAIPrompts(); err != nil {
		klog.V(6).Infof("初始化内置 AI 提示词失败: %v", err)
		return err
	}
	klog.V(6).Infof("安装 AI 插件成功")
	return nil
}

func (l *AILifecycle) Upgrade(ctx plugins.UpgradeContext) error {
	klog.V(6).Infof("升级 AI 插件：从版本 %s 到版本 %s", ctx.FromVersion(), ctx.ToVersion())
	if err := models.UpgradeDB(ctx.FromVersion(), ctx.ToVersion()); err != nil {
		klog.V(6).Infof("升级 AI 插件失败: %v", err)
		return err
	}
	return nil
}

func (l *AILifecycle) Enable(ctx plugins.EnableContext) error {
	klog.V(6).Infof("启用 AI 插件")
	return nil
}

func (l *AILifecycle) Disable(ctx plugins.BaseContext) error {
	klog.V(6).Infof("禁用 AI 插件")
	return nil
}

func (l *AILifecycle) Uninstall(ctx plugins.UninstallContext) error {
	if !ctx.KeepData() {
		if err := models.DropDB(); err != nil {
			klog.V(6).Infof("卸载 AI 插件失败: %v", err)
			return err
		}
	}
	klog.V(6).Infof("卸载 AI 插件成功")
	return nil
}

func (l *AILifecycle) Start(ctx plugins.BaseContext) error {
	klog.V(6).Infof("启动 AI 插件后台任务")
	klog.V(6).Infof("更新 AI 插件 运行配置")
	service.AIService().UpdateFlagFromAIRunConfig()
	return nil
}

func (l *AILifecycle) StartCron(ctx plugins.BaseContext, spec string) error {
	klog.V(6).Infof("启动 AI 插件定时任务，表达式: %s", spec)
	return nil
}

func (l *AILifecycle) Stop(ctx plugins.BaseContext) error {
	klog.V(6).Infof("停止 AI 插件后台任务")
	return nil
}
