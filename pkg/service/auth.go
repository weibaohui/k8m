package service

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"k8s.io/klog/v2"
)

type authService struct{}

var localAuthService = &authService{}

func AuthService() *authService {
	return localAuthService
}

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

func (a *authService) EnsureUserHasRoles(c *gin.Context, roles ...string) (bool, error) {
	user := amis.GetLoginUser(c)
	if UserService().IsUserPlatformAdmin(user) {
		return true, nil
	}
	rolesOfUser, err := UserService().GetRolesByUserName(user)
	if err != nil {
		klog.V(6).Infof("权限校验失败：获取用户角色错误，用户=%s，错误=%v", user, err)
		return false, err
	}
	klog.V(6).Infof("权限校验：用户=%s，用户角色=%v，需要角色=%v", user, rolesOfUser, roles)
	if hasIntersect(rolesOfUser, roles) {
		return true, nil
	}
	klog.V(6).Infof("权限校验失败：需要角色=%v，用户=%s，用户角色=%v", roles, user, rolesOfUser)
	return false, errors.New("无权限访问该路由")
}

func (a *authService) EnsureUserIsPlatformAdmin(c *gin.Context) (bool, error) {
	user := amis.GetLoginUser(c)
	if UserService().IsUserPlatformAdmin(user) {
		return true, nil
	}
	klog.V(6).Infof("权限校验失败：仅平台管理员可访问，用户=%s", user)
	return false, errors.New("仅平台管理员可访问")
}

func (a *authService) EnsureUserIsLogined(c *gin.Context) (bool, error) {
	user := amis.GetLoginUser(c)
	if user == "" {
		return false, errors.New("未登录用户")
	}
	return true, nil
}
