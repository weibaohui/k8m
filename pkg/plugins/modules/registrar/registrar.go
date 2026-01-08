package registrar

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules/ai"
	"github.com/weibaohui/k8m/pkg/plugins/modules/demo"
	"github.com/weibaohui/k8m/pkg/plugins/modules/eventhandler"
	"github.com/weibaohui/k8m/pkg/plugins/modules/gatewayapi"
	"github.com/weibaohui/k8m/pkg/plugins/modules/gllog"
	"github.com/weibaohui/k8m/pkg/plugins/modules/heartbeat"
	"github.com/weibaohui/k8m/pkg/plugins/modules/helm"
	"github.com/weibaohui/k8m/pkg/plugins/modules/inspection"
	"github.com/weibaohui/k8m/pkg/plugins/modules/istio"
	k8m_mcp_server "github.com/weibaohui/k8m/pkg/plugins/modules/k8m_mcp_server"
	"github.com/weibaohui/k8m/pkg/plugins/modules/k8sgpt"
	k8swatch "github.com/weibaohui/k8m/pkg/plugins/modules/k8swatch"
	"github.com/weibaohui/k8m/pkg/plugins/modules/leader"
	mcp "github.com/weibaohui/k8m/pkg/plugins/modules/mcp_runtime"
	"github.com/weibaohui/k8m/pkg/plugins/modules/openapi"
	"github.com/weibaohui/k8m/pkg/plugins/modules/openkruise"
	"github.com/weibaohui/k8m/pkg/plugins/modules/swagger"
	"github.com/weibaohui/k8m/pkg/plugins/modules/webhook"
	"k8s.io/klog/v2"
)

// init 插件集中注册器
// 在系统启动时设置plugins的集中注册函数，统一注册各插件
func init() {
	plugins.SetRegistrar(func(m *plugins.Manager) {
		if err := m.Register(demo.Metadata); err != nil {
			klog.V(6).Infof("注册demo插件失败: %v", err)
		} else {
			klog.V(6).Infof("注册demo插件成功")
		}
		if err := m.Register(leader.Metadata); err != nil {
			klog.V(6).Infof("注册leader插件失败: %v", err)
		} else {
			klog.V(6).Infof("注册leader插件成功")
		}
		if err := m.Register(webhook.Metadata); err != nil {
			klog.V(6).Infof("注册webhook插件失败: %v", err)
		} else {
			klog.V(6).Infof("注册webhook插件成功")
		}
		if err := m.Register(eventhandler.Metadata); err != nil {
			klog.V(6).Infof("注册eventhandler插件失败: %v", err)
		} else {
			klog.V(6).Infof("注册eventhandler插件成功")
		}
		if err := m.Register(inspection.Metadata); err != nil {
			klog.V(6).Infof("注册inspection插件失败: %v", err)
		} else {
			klog.V(6).Infof("注册inspection插件成功")
		}
		if err := m.Register(helm.Metadata); err != nil {
			klog.V(6).Infof("注册helm插件失败: %v", err)
		} else {
			klog.V(6).Infof("注册helm插件成功")
		}
		if err := m.Register(gllog.Metadata); err != nil {
			klog.V(6).Infof("注册gllog插件失败: %v", err)
		} else {
			klog.V(6).Infof("注册gllog插件成功")
		}
		if err := m.Register(swagger.Metadata); err != nil {
			klog.V(6).Infof("注册swagger插件失败: %v", err)
		} else {
			klog.V(6).Infof("注册swagger插件成功")
		}
		if err := m.Register(mcp.Metadata); err != nil {
			klog.V(6).Infof("注册mcp插件失败: %v", err)
		} else {
			klog.V(6).Infof("注册mcp插件成功")
		}
		if err := m.Register(openapi.Metadata); err != nil {
			klog.V(6).Infof("注册openapi插件失败: %v", err)
		} else {
			klog.V(6).Infof("注册openapi插件成功")
		}
		if err := m.Register(k8m_mcp_server.Metadata); err != nil {
			klog.V(6).Infof("注册k8m_mcp_server插件失败: %v", err)
		} else {
			klog.V(6).Infof("注册k8m_mcp_server插件成功")
		}
		if err := m.Register(k8sgpt.Metadata); err != nil {
			klog.V(6).Infof("注册k8sgpt插件失败: %v", err)
		} else {
			klog.V(6).Infof("注册k8sgpt插件成功")
		}
		if err := m.Register(ai.Metadata); err != nil {
			klog.V(6).Infof("注册ai插件失败: %v", err)
		} else {
			klog.V(6).Infof("注册ai插件成功")
		}
		if err := m.Register(heartbeat.Metadata); err != nil {
			klog.V(6).Infof("注册heartbeat插件失败: %v", err)
		} else {
			klog.V(6).Infof("注册heartbeat插件成功")
		}
		if err := m.Register(k8swatch.Metadata); err != nil {
			klog.V(6).Infof("注册k8swatch插件失败: %v", err)
		} else {
			klog.V(6).Infof("注册k8swatch插件成功")
		}
		if err := m.Register(gatewayapi.Metadata); err != nil {
			klog.V(6).Infof("注册gatewayapi插件失败: %v", err)
		} else {
			klog.V(6).Infof("注册gatewayapi插件成功")
		}
		if err := m.Register(istio.Metadata); err != nil {
			klog.V(6).Infof("注册istio插件失败: %v", err)
		} else {
			klog.V(6).Infof("注册istio插件成功")
		}
		if err := m.Register(openkruise.Metadata); err != nil {
			klog.V(6).Infof("注册openkruise插件失败: %v", err)
		} else {
			klog.V(6).Infof("注册openkruise插件成功")
		}
	})
}
