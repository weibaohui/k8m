package plugins

import (
	"sync"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"
)

var (
	routeRuleRegistry sync.Map // pluginName -> []RouteRule
)

// RegisterRouteRule 注册路由权限规则到全局注册表
func RegisterRouteRule(pluginName string, rule RouteRule) {
	v, _ := routeRuleRegistry.LoadOrStore(pluginName, []RouteRule{})
	existing := v.([]RouteRule)
	existing = append(existing, rule)
	routeRuleRegistry.Store(pluginName, existing)
	klog.V(6).Infof("为插件[%s]注册路由权限: %s %s，类型=%s，角色=%v", pluginName, rule.Method, rule.Path, rule.Kind, rule.Roles)
}

// GetRouteRules 获取插件已注册的动态路由权限规则
func GetRouteRules(pluginName string) []RouteRule {
	v, ok := routeRuleRegistry.Load(pluginName)
	if !ok {
		return nil
	}
	rules, _ := v.([]RouteRule)
	return rules
}

// SecuredGroup 封装后的安全路由分组
// 在注册 Gin 路由的同时，自动记录路由权限规则到注册表
type SecuredGroup struct {
	group      *gin.RouterGroup
	pluginName string
}

// NewSecuredGroup 创建安全路由分组，自动以 /plugins/<pluginName> 为前缀
func NewSecuredGroup(api *gin.RouterGroup, pluginName string) *SecuredGroup {
	g := api.Group("/plugins/" + pluginName)
	return &SecuredGroup{
		group:      g,
		pluginName: pluginName,
	}
}

// GET 注册 GET 路由并记录权限
func (s *SecuredGroup) GET(path string, kind RouteAccessKind, handler gin.HandlerFunc, roles ...string) {
	s.group.GET(path, handler)
	RegisterRouteRule(s.pluginName, RouteRule{
		Method: "GET",
		Path:   path,
		Kind:   kind,
		Roles:  roles,
	})
}

// POST 注册 POST 路由并记录权限
func (s *SecuredGroup) POST(path string, kind RouteAccessKind, handler gin.HandlerFunc, roles ...string) {
	s.group.POST(path, handler)
	RegisterRouteRule(s.pluginName, RouteRule{
		Method: "POST",
		Path:   path,
		Kind:   kind,
		Roles:  roles,
	})
}

