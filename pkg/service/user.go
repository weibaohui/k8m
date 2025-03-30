package service

import (
	"strings"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/models"
	"gorm.io/gorm"
)

type userService struct {
}

func (u *userService) List() ([]*models.User, error) {
	user := &models.User{}
	params := dao.Params{
		PerPage: 10000000,
	}
	list, _, err := user.List(&params)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// GetClusterRole 获取用户在指定集群中的角色权限
// cluster: 集群名称
// username: 用户名
// jwtUserRole: JWT用户角色,从context传递
func (u *userService) GetClusterRole(cluster string, username string, jwtUserRoles string) (string, error) {
	//jwtUserRoles可能为一个字符串逗号分隔的角色列表
	if jwtUserRoles != "" {
		roles := strings.SplitSeq(jwtUserRoles, ",")
		for role := range roles {
			//只有平台管理员才返回，这是最大权限了
			//不是平台管理员就是普通用户，这是权限系统的设定，只有这两种角色
			//普通用户需要接受集群权限授权，那么就往下执行，查看是否具有集群授权
			if role == models.RolePlatformAdmin {
				return role, nil
			}
		}
	}

	params := &dao.Params{}
	params.PerPage = 10000000
	clusterRole := &models.ClusterUserRole{}
	queryFunc := func(db *gorm.DB) *gorm.DB {
		return db.Where("cluster = ? AND username = ?", cluster, username)
	}
	roles, _, err := clusterRole.List(params, queryFunc)
	if err != nil {
		return "", err
	}
	// 遍历所有角色，如果存在admin权限就返回admin
	for _, role := range roles {
		if role.Role == models.RoleClusterAdmin || role.Role == models.RolePlatformAdmin {
			return role.Role, nil
		}
	}
	// 如果没有找到admin权限，返回readonly权限（如果有的话）
	for _, role := range roles {
		if role.Role == models.RoleClusterReadonly {
			return role.Role, nil
		}
	}

	return "", nil
}
