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

// EnsureUserHasRoles 方法内角色校验，支持平台管理员通行
// 该函数用于在具体处理方法内进行角色权限校验
// 返回 (true, nil) 表示校验通过；返回 (false, error) 表示校验失败
// 不直接写入响应，由上层处理错误写入
func EnsureUserHasRoles(c *gin.Context, roles ...string) (bool, error) {
	user := amis.GetLoginUser(c)
	// 平台管理员拥有所有权限
	if service.UserService().IsUserPlatformAdmin(user) {
		return true, nil
	}
	// 查询用户角色
	rolesOfUser, err := service.UserService().GetRolesByUserName(user)
	if err != nil {
		klog.V(6).Infof("权限校验失败：获取用户角色错误，用户=%s，错误=%v", user, err)
		return false, err
	}
	klog.V(6).Infof("权限校验：用户=%s，用户角色=%v，需要角色=%v", user, rolesOfUser, roles)
	// 角色匹配
	if hasIntersect(rolesOfUser, roles) {
		return true, nil
	}
	klog.V(6).Infof("权限校验失败：需要角色=%v，用户=%s，用户角色=%v", roles, user, rolesOfUser)
	return false, errors.New("无权限访问该路由")
}

// EnsurePlatformAdmin 方法内平台管理员校验
// 返回 (true, nil) 表示是平台管理员；返回 (false, error) 表示校验失败
// 不直接写入响应，由上层处理错误写入
func EnsureUserIsPlatformAdmin(c *gin.Context) (bool, error) {
	user := amis.GetLoginUser(c)
	if service.UserService().IsUserPlatformAdmin(user) {
		return true, nil
	}
	klog.V(6).Infof("权限校验失败：仅平台管理员可访问，用户=%s", user)
	return false, errors.New("仅平台管理员可访问")
}
func EnsureUserIsLogined(c *gin.Context) (bool, error) {
	user := amis.GetLoginUser(c)
	if user == "" {
		return false, errors.New("未登录用户")
	}
	return true, nil
}
