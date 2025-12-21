package plugins

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/service"
	"k8s.io/klog/v2"
)

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

// EnsureRoles 方法内角色校验，支持平台管理员通行
// 该函数用于在具体的处理方法内进行角色权限校验，避免仅依赖路径规则。
// - 当用户为平台管理员时，直接放行
// - 当用户拥有指定角色集合的任意一个时，放行
// - 其他情况写入错误响应并中止请求
func EnsureRoles(c *gin.Context, roles ...string) bool {
	user := amis.GetLoginUser(c)
	// 平台管理员拥有所有权限
	if service.UserService().IsUserPlatformAdmin(user) {
		klog.V(6).Infof("权限校验：用户=%s，为平台管理员，路径=%s", user, c.FullPath())
		return true
	}
	// 查询用户角色
	rolesOfUser, err := service.UserService().GetRolesByUserName(user)
	if err != nil {
		klog.V(6).Infof("权限校验失败：获取用户角色错误，用户=%s，错误=%v", user, err)
		amis.WriteJsonError(c, err)
		c.Abort()
		return false
	}
	klog.V(6).Infof("权限校验：用户=%s，用户角色=%v，需要角色=%v，路径=%s", user, rolesOfUser, roles, c.FullPath())
	// 角色匹配
	if hasIntersect(rolesOfUser, roles) {
		return true
	}
	klog.V(6).Infof("权限校验失败：需要角色=%v，用户=%s，用户角色=%v，路径=%s", roles, user, rolesOfUser, c.FullPath())
	amis.WriteJsonError(c, errors.New("无权限访问该路由"))
	c.Abort()
	return false
}

// EnsurePlatformAdmin 方法内平台管理员校验
// 该函数用于在具体处理方法内校验是否为平台管理员，是则放行，否则返回错误并中止请求。
func EnsurePlatformAdmin(c *gin.Context) bool {
	user := amis.GetLoginUser(c)
	if service.UserService().IsUserPlatformAdmin(user) {
		return true
	}
	klog.V(6).Infof("权限校验失败：仅平台管理员可访问，用户=%s，路径=%s", user, c.FullPath())
	amis.WriteJsonError(c, errors.New("仅平台管理员可访问"))
	c.Abort()
	return false
}
