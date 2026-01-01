package plugins

import (
	"fmt"

	"github.com/robfig/cron/v3"
	"github.com/weibaohui/k8m/pkg/plugins/eventbus"
	"k8s.io/klog/v2"
)

// setCronRunning 设置某个插件某条 cron 的运行状态
func (m *Manager) setCronRunning(name, spec string, running bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.cronRunning[name]; !ok {
		m.cronRunning[name] = make(map[string]bool)
	}
	m.cronRunning[name][spec] = running
}

// getCronEntry 获取某个插件某条 cron 的调度条目
func (m *Manager) getCronEntry(name, spec string) (cron.Entry, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	idMap, ok := m.cronIDs[name]
	if !ok {
		return cron.Entry{}, false
	}
	id, ok := idMap[spec]
	if !ok {
		return cron.Entry{}, false
	}
	return m.cron.Entry(id), true
}

// EnsureCron 确保某条 cron 已注册（不存在则注册）
func (m *Manager) EnsureCron(name, spec string) error {
	m.mu.Lock()
	mod, ok := m.modules[name]
	m.mu.Unlock()
	if !ok || mod.Lifecycle == nil {
		return fmt.Errorf("插件未注册或未实现生命周期: %s", name)
	}
	if _, err := cron.ParseStandard(spec); err != nil {
		return fmt.Errorf("cron 表达式非法: %s，错误: %v", spec, err)
	}
	if _, ok := m.getCronEntry(name, spec); ok {
		return nil
	}
	ctx := baseContextImpl{meta: mod.Meta, bus: eventbus.New()}
	n := name
	s := spec
	id, err := m.cron.AddFunc(s, func() {
		m.setCronRunning(n, s, true)
		if err := mod.Lifecycle.StartCron(ctx, s); err != nil {
			klog.V(6).Infof("执行插件定时任务失败: %s，表达式: %s，错误: %v", n, s, err)
		} else {
			klog.V(6).Infof("执行插件定时任务成功: %s，表达式: %s", n, s)
		}
		m.setCronRunning(n, s, false)
	})
	if err != nil {
		return err
	}
	m.mu.Lock()
	if _, ok := m.cronIDs[name]; !ok {
		m.cronIDs[name] = make(map[string]cron.EntryID)
	}
	m.cronIDs[name][spec] = id
	m.mu.Unlock()
	return nil
}

// RemoveCron 删除某条 cron
func (m *Manager) RemoveCron(name, spec string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	idMap, ok := m.cronIDs[name]
	if !ok {
		return
	}
	id, ok := idMap[spec]
	if !ok {
		return
	}
	m.cron.Remove(id)
	delete(idMap, spec)
	if runMap, ok := m.cronRunning[name]; ok {
		delete(runMap, spec)
	}
	klog.V(6).Infof("强制停止插件定时任务: %s，表达式: %s", name, spec)
}

// RunCronOnce 立即执行一次某条 cron 的任务
func (m *Manager) RunCronOnce(name, spec string) error {
	m.mu.RLock()
	mod, ok := m.modules[name]
	m.mu.RUnlock()
	if !ok || mod.Lifecycle == nil {
		return fmt.Errorf("插件未注册或未实现生命周期: %s", name)
	}
	ctx := baseContextImpl{meta: mod.Meta, bus: eventbus.New()}
	go func() {
		m.setCronRunning(name, spec, true)
		if err := mod.Lifecycle.StartCron(ctx, spec); err != nil {
			klog.V(6).Infof("手动执行插件定时任务失败: %s，表达式: %s，错误: %v", name, spec, err)
		} else {
			klog.V(6).Infof("手动执行插件定时任务成功: %s，表达式: %s", name, spec)
		}
		m.setCronRunning(name, spec, false)
	}()
	return nil
}
