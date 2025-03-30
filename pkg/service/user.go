package service

import (
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
func (u *userService) GetClusterRole(cluster string, username string, jwtUserRole string) (string, error) {
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

	// 如果数据库中没有,则使用jwtUserRole
	if jwtUserRole != "" {
		return jwtUserRole, nil
	}

	return "", nil
}
