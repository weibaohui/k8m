package service

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/flag"
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

// GetRolesByGroupNames 获取用户的角色
func (u *userService) GetRolesByGroupNames(groupNames string) ([]string, error) {
	var ugList []models.UserGroup
	err := dao.DB().Model(&models.UserGroup{}).Where("group_name in ?", strings.Split(groupNames, ",")).Distinct("role").Find(&ugList).Error
	if err != nil {
		return nil, err
	}
	// 查询所有的用户组，判断用户组的角色
	// 形成一个用户组对应的角色列表
	var roles []string
	for _, ug := range ugList {
		roles = append(roles, ug.Role)
	}
	return roles, nil
}

// GetClusterRole 获取用户在指定集群中的角色权限
// cluster: 集群名称
// username: 用户名
// jwtUserRole: JWT用户角色,从context传递
// 返回值：角色列表 [平台角色，集群角色合并了，后续考虑拆开]
func (u *userService) GetClusterRole(cluster string, username string, jwtUserRoles string) ([]string, error) {
	// jwtUserRoles可能为一个字符串逗号分隔的角色列表
	if jwtUserRoles != "" {
		roles := strings.SplitSeq(jwtUserRoles, ",")
		for role := range roles {
			// 只有平台管理员才返回，这是最大权限了
			// 不是平台管理员就是普通用户，这是权限系统的设定，只有这两种角色
			// 普通用户需要接受集群权限授权，那么就往下执行，查看是否具有集群授权
			if role == constants.RolePlatformAdmin {
				return []string{role}, nil
			}
		}
	}
	// 先从jwt字符串中读取，没有再读数据库
	params := &dao.Params{}
	params.PerPage = 10000000
	clusterRole := &models.ClusterUserRole{}
	queryFunc := func(db *gorm.DB) *gorm.DB {
		return db.Distinct("role").Where("cluster = ? AND username = ?", cluster, username)
	}
	items, _, err := clusterRole.List(params, queryFunc)
	if err != nil {
		return []string{}, err
	}
	var roles []string
	for _, item := range items {
		roles = append(roles, item.Role)
	}

	return roles, nil
}

// GetClusterNames 获取用户有权限的集群名称数组
// username: 用户名
func (u *userService) GetClusterNames(username string) ([]string, error) {
	params := &dao.Params{}
	params.PerPage = 10000000
	clusterRole := &models.ClusterUserRole{}
	queryFunc := func(db *gorm.DB) *gorm.DB {
		return db.Distinct("cluster").Where(" username = ?", username)
	}
	items, _, err := clusterRole.List(params, queryFunc)
	if err != nil {
		return []string{}, err
	}
	var clusters []string
	for _, item := range items {
		clusters = append(clusters, item.Cluster)
	}

	return clusters, nil
}

// GetClusters 获取用户有权限的集群列表
// username: 用户名
// 最终结果包含两种情况：
// 1. 用户授权类型为用户
// 2. 用户授权类型为用户组,当前用户所在的用户组，如果有授权，那么也提取出来
func (u *userService) GetClusters(username string) ([]*models.ClusterUserRole, error) {
	params := &dao.Params{}
	params.PerPage = 10000000
	clusterRole := &models.ClusterUserRole{}
	queryFunc := func(db *gorm.DB) *gorm.DB {
		return db.Where(" username = ?", username)
	}
	items, _, err := clusterRole.List(params, queryFunc)
	if err != nil {
		return nil, err
	}
	// 以上为授权类型为用户的情况
	// 以下为授权类型为用户组的情况
	// 先获取用户所在用户组名称，可能多个
	if groupNames, err := u.GetGroupNames(username); err == nil {
		goupNameList := strings.Split(groupNames, ",")
		if len(goupNameList) > 0 {
			// 查找用户组对应的授权
			if items2, _, err := clusterRole.List(params, func(db *gorm.DB) *gorm.DB {
				return db.Where("authorization_type=? and  username in ? ", constants.ClusterAuthorizationTypeUserGroup, goupNameList)
			}); err == nil {
				items = append(items, items2...)
			}
		}
	}

	return items, nil
}

// GenerateJWTToken 生成 Token
func (u *userService) GenerateJWTTokenByUserName(username string, duration time.Duration) (string, error) {
	role := constants.JwtUserRole
	name := constants.JwtUserName
	cst := constants.JwtClusters
	cstUserRoles := constants.JwtClusterUserRoles

	groupNames, _ := u.GetGroupNames(username)
	roles, _ := u.GetRolesByGroupNames(groupNames)
	// 查询用户对应的集群
	clusters, _ := u.GetClusters(username)

	var clusterNames []string
	if clusters != nil {
		for _, cluster := range clusters {
			clusterNames = append(clusterNames, cluster.Cluster)
		}
	}

	var token = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		name:         username,
		role:         strings.Join(roles, ","),        // 角色列表
		cst:          strings.Join(clusterNames, ","), // 集群名称列表
		cstUserRoles: utils.ToJSON(clusters),          // 集群用户角色列表 可以反序列化为[]*models.ClusterUserRole
		"exp":        time.Now().Add(duration).Unix(), // 国企时间
	})
	cfg := flag.Init()
	var jwtSecret = []byte(cfg.JwtTokenSecret)
	return token.SignedString(jwtSecret)
}

// GenerateJWTToken 生成 Token
func (u *userService) GenerateJWTToken(username string, roles []string, clusters []*models.ClusterUserRole, duration time.Duration) (string, error) {
	role := constants.JwtUserRole
	name := constants.JwtUserName
	cst := constants.JwtClusters
	cstUserRoles := constants.JwtClusterUserRoles

	var clusterNames []string
	if clusters != nil {
		for _, cluster := range clusters {
			clusterNames = append(clusterNames, cluster.Cluster)
		}
	}

	var token = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		name:         username,
		role:         strings.Join(roles, ","),        // 角色列表
		cst:          strings.Join(clusterNames, ","), // 集群名称列表
		cstUserRoles: utils.ToJSON(clusters),          // 集群用户角色列表 可以反序列化为[]*models.ClusterUserRole
		"exp":        time.Now().Add(duration).Unix(), // 国企时间
	})
	cfg := flag.Init()
	var jwtSecret = []byte(cfg.JwtTokenSecret)
	return token.SignedString(jwtSecret)
}

func (u *userService) GetGroupNames(username string) (string, error) {
	params := &dao.Params{}
	user := &models.User{}
	queryFunc := func(db *gorm.DB) *gorm.DB {
		return db.Select("group_names").Where(" username = ?", username)
	}
	item, err := user.GetOne(params, queryFunc)
	if err != nil {
		return "", err
	}

	return item.GroupNames, nil
}

func (u *userService) GetUserByMCPKey(mcpKey string) (string, error) {
	params := &dao.Params{}
	user := &models.McpKey{}
	queryFunc := func(db *gorm.DB) *gorm.DB {
		return db.Select("username").Where(" key = ?", mcpKey)
	}
	item, err := user.GetOne(params, queryFunc)
	if err != nil {
		return "", err
	}

	return item.Username, nil
}

// CheckAndCreateUser 检查用户是否存在，如果不存在则创建一个新用户
func (u *userService) CheckAndCreateUser(username, source string) error {
	params := &dao.Params{}
	user := &models.User{}
	queryFunc := func(db *gorm.DB) *gorm.DB {
		return db.Where("username = ?", username)
	}
	_, err := user.GetOne(params, queryFunc)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 用户不存在，创建新用户
			newUser := &models.User{
				Username:  username,
				Source:    source,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			return newUser.Save(params)
		}
		return err
	}
	return nil
}

// GetPlatformRolesByName 通过用户名获取用户的平台角色
func (u *userService) GetPlatformRolesByName(username string) string {
	cfg := flag.Init()
	if cfg.EnableTempAdmin && username == cfg.AdminUserName {
		return constants.RolePlatformAdmin
	}
	if names, err := u.GetGroupNames(username); err == nil {
		if rolesByGroupNames, err := u.GetRolesByGroupNames(names); err == nil {
			return strings.Join(rolesByGroupNames, ",")
		}
	}
	return ""
}
