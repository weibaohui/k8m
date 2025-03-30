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
func (u *userService) GetClusterRole(cluster string, username string) (string, error) {
	params := &dao.Params{}
	clusterRole := &models.ClusterUserRole{}
	queryFunc := func(db *gorm.DB) *gorm.DB {
		return db.Where("cluster = ? AND username = ?", cluster, username)
	}
	role, err := clusterRole.GetOne(params, queryFunc)

	if err != nil {
		return "", err
	}
	return role.Role, nil
}
