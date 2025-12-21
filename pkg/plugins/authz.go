package plugins

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/service"
	"k8s.io/klog/v2"
)

// RouteAccessMiddlewareFactory 创建一个全局路由访问控制中间件
// 该中间件会根据请求路径推断插件名称（/plugins/<name>/...），
// 并依据对应插件在 Metadata.RouteRules 中声明的规则进行权限校验。
func RouteAccessMiddlewareFactory(modules map[string]Module) gin.HandlerFunc {
	return func(c *gin.Context) {
		full := c.FullPath()
		method := c.Request.Method
		// 解析插件名：路径以 /plugins/<name>/ 开头
		pluginName := ""
		if strings.HasPrefix(full, "/plugins/") {
			rest := strings.TrimPrefix(full, "/plugins/")
			parts := strings.SplitN(rest, "/", 2)
			if len(parts) >= 1 {
				pluginName = parts[0]
			}
		}
		// 非插件路由，放行
		if pluginName == "" {
			c.Next()
			return
		}
		mod, ok := modules[pluginName]
		if !ok {
			// 未注册插件，直接放行（或拒绝，这里选择放行避免误伤非插件路径）
			c.Next()
			return
		}
		// 计算相对路径（去掉 /plugins/<name> 前缀）
		relative := strings.TrimPrefix(full, "/plugins/"+mod.Meta.Name)
		if relative == "" {
			relative = "/"
		}
		// 查找匹配规则：优先按相对路径，其次按完整路径
		var matched *RouteRule
		for i := range mod.RouteRules {
			r := &mod.RouteRules[i]
			if !methodEquals(method, r.Method) {
				continue
			}
			if pathEquals(relative, r.Path) || pathEquals(full, r.Path) {
				matched = r
				break
			}
		}
		// 未配置规则，默认放行（视为普通用户可访问）
		if matched == nil {
			c.Next()
			return
		}
		// 执行权限校验
		user := amis.GetLoginUser(c)
		switch matched.Kind {
		case AccessAnyUser:
			// 已登录即可
			c.Next()
			return
		case AccessPlatformAdmin:
			if service.UserService().IsUserPlatformAdmin(user) {
				c.Next()
				return
			}
			klog.V(6).Infof("权限校验失败：仅平台管理员可访问，用户=%s，路径=%s", user, full)
			amis.WriteJsonError(c, errors.New("仅平台管理员可访问"))
			c.Abort()
			return
		case AccessRoles:
			roles, err := service.UserService().GetRolesByUserName(user)
			if err != nil {
				klog.V(6).Infof("权限校验失败：获取用户角色错误，用户=%s，错误=%v", user, err)
				amis.WriteJsonError(c, err)
				c.Abort()
				return
			}
			// 平台管理员拥有所有权限
			if service.UserService().IsUserPlatformAdmin(user) {
				c.Next()
				return
			}
			if hasIntersect(roles, matched.Roles) {
				c.Next()
				return
			}
			klog.V(6).Infof("权限校验失败：需要角色=%v，用户=%s，用户角色=%v，路径=%s", matched.Roles, user, roles, full)
			amis.WriteJsonError(c, errors.New("无权限访问该路由"))
			c.Abort()
			return
		default:
			// 未知类型，视为拒绝
			klog.V(6).Infof("权限校验失败：未知访问控制类型，用户=%s，路径=%s", user, full)
			amis.WriteJsonError(c, errors.New("访问控制配置错误"))
			c.Abort()
			return
		}
	}
}

// methodEquals 比较HTTP方法（区分大小写）
func methodEquals(a, b string) bool {
	return strings.TrimSpace(a) == strings.TrimSpace(b)
}

// pathEquals 比较路径，忽略多余的斜杠
func pathEquals(a, b string) bool {
	na := normalizePath(a)
	nb := normalizePath(b)
	return na == nb
}

// normalizePath 规范化路径：确保以 / 开头，移除尾部多余斜杠
func normalizePath(p string) string {
	if p == "" {
		return "/"
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	// 移除尾部斜杠（但保留根路径 /）
	for len(p) > 1 && strings.HasSuffix(p, "/") {
		p = strings.TrimSuffix(p, "/")
	}
	return p
}

// hasIntersect 判断两个字符串切片是否有交集
func hasIntersect(a, b []string) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	set := make(map[string]struct{}, len(a))
	for _, x := range a {
		set[x] = struct{}{}
	}
	for _, y := range b {
		if _, ok := set[y]; ok {
			return true
		}
	}
	return false
}
