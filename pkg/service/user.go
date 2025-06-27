package service

import (
	"errors"
	"fmt"
	"github.com/go-ldap/ldap/v3"
	"k8s.io/klog/v2"
	"strings"
	"sync"
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
	cacheKeys sync.Map // 用于存储所有使用过的缓存key
}

// addCacheKey 添加缓存key到列表中（并发安全）
func (u *userService) addCacheKey(key string) {
	u.cacheKeys.Store(key, struct{}{})
}

// getCacheKeys 获取所有缓存key（并发安全）
func (u *userService) getCacheKeys() []string {
	var keys []string
	u.cacheKeys.Range(func(key, value interface{}) bool {
		if k, ok := key.(string); ok {
			keys = append(keys, k)
		}
		return true
	})
	return keys
}

// formatCacheKey 格式化缓存key并添加到列表中（并发安全）
func (u *userService) formatCacheKey(format string, a ...interface{}) string {
	key := fmt.Sprintf(format, a...)
	u.addCacheKey(key)
	return key
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
	if groupNames == "" {
		return nil, nil
	}

	cacheKey := u.formatCacheKey("user:roles:%s", groupNames)

	result, err := utils.GetOrSetCache(CacheService().CacheInstance(), cacheKey, 5*time.Minute, func() ([]string, error) {
		var ugList []models.UserGroup
		err := dao.DB().Model(&models.UserGroup{}).Where("group_name in ?", strings.Split(groupNames, ",")).Distinct("role").Find(&ugList).Error
		if err != nil {
			return nil, err
		}
		var roles []string
		for _, ug := range ugList {
			roles = append(roles, ug.Role)
		}
		return roles, nil
	})

	return result, err
}

// GetClusterRole 获取用户在指定集群中的角色权限
// cluster: 集群名称
// username: 用户名
// jwtUserRole: JWT用户角色,从context传递
// 返回值：角色列表 [平台角色，集群角色合并了，后续考虑拆开]
// 包含用户本身的集群角色、以及所在用户组的集群角色
func (u *userService) GetClusterRole(cluster string, username string, jwtUserRoles string) ([]string, error) {
	// jwtUserRoles的检查逻辑保持不变
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

	cacheKey := u.formatCacheKey("user:clusterrole:%s:%s", username, cluster)

	result, err := utils.GetOrSetCache(CacheService().CacheInstance(), cacheKey, 5*time.Minute, func() ([]string, error) {
		// 查找用户本身、用户所在组，两个层面的集群权限。先形成查询名称的列表
		var unionNames []string
		if groupNames, err := u.GetGroupNames(username); err == nil {
			unionNames = append(unionNames, strings.Split(groupNames, ",")...)
		}
		unionNames = append(unionNames, username)

		// 查找用户本身、用户所在组，两个层面的集群权限。
		params := &dao.Params{}
		params.PerPage = 10000000
		clusterRole := &models.ClusterUserRole{}
		queryFunc := func(db *gorm.DB) *gorm.DB {
			return db.Distinct("role").Where("cluster = ? AND username in ?", cluster, unionNames)
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
	})

	return result, err
}

// GetClusterNames 获取用户有权限的集群名称数组
// username: 用户名
func (u *userService) GetClusterNames(username string) ([]string, error) {

	items, err := u.GetClusters(username)
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
	cacheKey := u.formatCacheKey("user:clusters:%s", username)

	result, err := utils.GetOrSetCache(CacheService().CacheInstance(), cacheKey, 5*time.Minute, func() ([]*models.ClusterUserRole, error) {
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
	})

	return result, err
}

// GenerateJWTTokenByUserName  生成 Token
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

	for _, cluster := range clusters {
		clusterNames = append(clusterNames, cluster.Cluster)
	}

	var token = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		name:         username,
		role:         strings.Join(roles, ","),        // 角色列表
		cst:          strings.Join(clusterNames, ","), // 集群名称列表
		cstUserRoles: utils.ToJSON(clusters),          // 集群用户角色列表 可以反序列化为[]*models.ClusterUserRole
		"exp":        time.Now().Add(duration).Unix(),
	})
	cfg := flag.Init()
	var jwtSecret = []byte(cfg.JwtTokenSecret)
	return token.SignedString(jwtSecret)
}

// GenerateJWTTokenOnlyUserName  生成 Token，仅包含Username
func (u *userService) GenerateJWTTokenOnlyUserName(username string, duration time.Duration) (string, error) {
	name := constants.JwtUserName

	var token = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		name:  username,
		"exp": time.Now().Add(duration).Unix(),
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

	for _, cluster := range clusters {
		clusterNames = append(clusterNames, cluster.Cluster)
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

// GetGroupNames 获取用户所在的用户组
func (u *userService) GetGroupNames(username string) (string, error) {
	cacheKey := u.formatCacheKey("user:groupnames:%s", username)

	result, err := utils.GetOrSetCache(CacheService().CacheInstance(), cacheKey, 5*time.Minute, func() (string, error) {
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
	})

	return result, err
}

func (u *userService) GetUserByMCPKey(mcpKey string) (string, error) {
	params := &dao.Params{}
	m := &models.McpKey{}
	queryFunc := func(db *gorm.DB) *gorm.DB {
		return db.Select("username").Where(" mcp_key = ?", mcpKey)
	}
	item, err := m.GetOne(params, queryFunc)
	if err != nil {
		return "", err
	}

	if item.Username == "" {
		return "", errors.New("username is empty")
	}

	// 检测用户是否被禁用
	user := &models.User{}
	disabled, err := user.IsDisabled(item.Username)
	if err != nil {
		return "", err
	}
	if disabled {
		return "", fmt.Errorf("用户[%s]被禁用", item.Username)
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

// ClearCacheByKey 清除指定关键字的所有相关缓存
func (u *userService) ClearCacheByKey(cacheKey string) {
	// 遍历所有已使用的缓存key
	for _, key := range u.getCacheKeys() {
		if strings.Contains(key, cacheKey) {
			utils.ClearCacheByKey(CacheService().CacheInstance(), key)
		}
	}
}

// ldap连接
func (u *userService) ldapConnection(cfg *flag.Config) (*ldap.Conn, error) {
	conn, err := ldap.Dial("tcp", cfg.LdapHost+":"+cfg.LdapPort)
	if err != nil {
		klog.Errorf("无法连接到ldap服务器: %v", err)
		return nil, err
	}

	// 设置超时时间
	conn.SetTimeout(5 * time.Second)
	return conn, nil
}

// ldap搜索
func (u *userService) searchRequest(conn *ldap.Conn, username string, cfg *flag.Config) (*ldap.Entry, error) {
	var (
		cur              *ldap.SearchResult
		ldapFieldsFilter = []string{
			"dn",
		}
	)
	err := conn.Bind(cfg.LdapBindUserDN, cfg.LdapPassword)
	if err != nil {
		klog.Errorf("LDAP绑定失败: %v", err)
		return nil, errors.New("LDAP认证失败")
	}

	sql := ldap.NewSearchRequest(
		cfg.LdapBaseDN,
		ldap.ScopeWholeSubtree,
		ldap.DerefAlways,
		0,
		0,
		false,
		fmt.Sprintf("(%v=%v)", cfg.LdapUserField, username),
		ldapFieldsFilter,
		nil)

	cur, err = conn.Search(sql)
	if err != nil {
		klog.Errorf("LDAP搜索用户失败: %v", err)
		return nil, errors.New("用户搜索失败")
	}

	if len(cur.Entries) == 0 {
		klog.Errorf("LDAP中未找到用户: %s", username)
		return nil, errors.New("用户不存在")
	}

	return cur.Entries[0], nil
}

// 登录ldap
func (u *userService) LoginWithLdap(username string, password string, cfg *flag.Config) (*ldap.Entry, error) {
	// 创建新连接
	conn, err := u.ldapConnection(cfg)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// 使用连接进行搜索
	userInfo, err := u.searchRequest(conn, username, cfg)
	if err != nil {
		return nil, err
	}

	// 验证用户密码
	err = conn.Bind(userInfo.DN, password)
	if err != nil {
		klog.Errorf("LDAP用户密码验证失败: %s", username)
		return nil, errors.New("用户或密码不正确")
	}

	return userInfo, nil
}
