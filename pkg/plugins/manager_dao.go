package plugins

import (
	"fmt"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/models"
	"k8s.io/klog/v2"
)

// PersistStatus 将插件状态持久化到数据库
// 管理员API调用该方法写入配置，实际生效需要重启
func (m *Manager) PersistStatus(name string, status Status, params *dao.Params) error {
	if _, ok := m.modules[name]; !ok {
		return fmt.Errorf("插件未注册: %s", name)
	}
	cfg := &models.PluginConfig{
		Name:    name,
		Status:  statusToString(status),
		Version: m.modules[name].Meta.Version,
	}
	return cfg.SaveByName(params)
}

// ApplyConfigFromDB 启动时从数据库加载插件配置并应用
// 根据持久化状态执行安装或启用操作；未配置的插件默认启用并写入数据库
func (m *Manager) ApplyConfigFromDB() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 读取所有持久化配置
	params := dao.BuildDefaultParams()
	cfgModel := &models.PluginConfig{}
	cfgs, _, err := cfgModel.List(params)
	if err != nil {
		klog.V(6).Infof("读取插件配置失败: %v", err)
	}
	cfgMap := make(map[string]Status, len(cfgs))
	for _, c := range cfgs {
		cfgMap[c.Name] = statusFromString(c.Status)
	}

	// 应用配置
	for name, mod := range m.modules {
		st, ok := cfgMap[name]
		if !ok {
			// 默认标记为已发现，不写入数据库
			st = StatusDiscovered
		}
		m.status[name] = st

		switch st {
		case StatusInstalled:
			if mod.Lifecycle != nil {
				ctx := installContextImpl{baseContextImpl{meta: mod.Meta}}
				if err := mod.Lifecycle.Install(ctx); err != nil {
					klog.V(6).Infof("启动时安装插件失败: %s，错误: %v", name, err)
				} else {
					klog.V(6).Infof("启动时安装插件成功: %s", name)
				}
			}
		case StatusEnabled:
			if mod.Lifecycle != nil {
				ctx := enableContextImpl{baseContextImpl{meta: mod.Meta}}
				if err := mod.Lifecycle.Enable(ctx); err != nil {
					klog.V(6).Infof("启动时启用插件失败: %s，错误: %v", name, err)
				} else {
					klog.V(6).Infof("启动时启用插件成功: %s", name)
				}
			}
		case StatusDisabled:
			klog.V(6).Infof("启动时禁用插件: %s", name)
		case StatusDiscovered:
			klog.V(6).Infof("启动时标记插件为已发现: %s", name)
		}
	}
}
