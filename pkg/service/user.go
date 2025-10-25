package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/go-ldap/ldap/v3"
	"k8s.io/klog/v2"

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

// GetGroupMenuData 获取用户组的菜单数据
// groupNames: 用户组名称，逗号分隔
// 用户会属于多个组，每个组有不同的菜单数据，如何合并？
// 当前策略，保留第一个出现的项，以后再改成菜单切换形式，看到多套菜单，根据不同角色下的菜单数据进行切换
// return: 菜单数据，json字符串
func (u *userService) GetGroupMenuData(groupNameList []string) (any, error) {
	if len(groupNameList) == 0 {
		// 返回默认空菜单结构
		return []any{}, nil
	}

	// 尝试从所有用户组中获取菜单数据
	// 找到第一个不为空的cacheKey，返回其数据
	for _, groupName := range groupNameList {
		groupName = strings.TrimSpace(groupName)
		if groupName == "" {
			continue
		}

		cacheKey := u.formatCacheKey("user:groupmenu:%s", groupName)

		result, err := utils.GetOrSetCache(CacheService().CacheInstance(), cacheKey, 5*time.Minute, func() (any, error) {
			params := &dao.Params{}
			userGroup := &models.UserGroup{}
			queryFunc := func(db *gorm.DB) *gorm.DB {
				return db.Select("menu_data").Where("group_name = ?", groupName)
			}

			item, err := userGroup.GetOne(params, queryFunc)
			if err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, err
				}
				return []any{}, nil
			}

			if item.MenuData != "" {
				// 将JSON字符串解析为Go对象
				var menuData any
				if err := json.Unmarshal([]byte(item.MenuData), &menuData); err == nil {
					if menuData == nil {
						return []any{}, nil
					}
					return menuData, nil
				}
				// 如果解析失败，返回空数组
				return []any{}, nil
			}

			return []any{}, nil
		})

		// 如果找到了有效的菜单数据，直接返回
		if err == nil && result != nil {
			return result, nil
		}
	}

	// 如果没有找到任何菜单数据，返回默认空菜单结构
	return []any{}, nil
}

// addCacheKey 添加缓存key到列表中（并发安全）
func (u *userService) addCacheKey(key string) {
	u.cacheKeys.Store(key, struct{}{})
}

// getCacheKeys 获取所有缓存key（并发安全）
func (u *userService) getCacheKeys() []string {
	var keys []string
	u.cacheKeys.Range(func(key, value any) bool {
		if k, ok := key.(string); ok {
			keys = append(keys, k)
		}
		return true
	})
	return keys
}

// formatCacheKey 格式化缓存key并添加到列表中（并发安全）
func (u *userService) formatCacheKey(format string, a ...any) string {
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

// GetRolesByUserName 通过用户名获取用户的角色
func (u *userService) GetRolesByUserName(username string) ([]string, error) {
	// 先判断是否不是最大的临时管理员账户
	cfg := flag.Init()
	if cfg.EnableTempAdmin && username == cfg.AdminUserName {
		return []string{constants.RolePlatformAdmin}, nil
	}
	groupNames, err := u.GetGroupNames(username)
	if err != nil {
		return nil, err
	}
	roles, err := u.GetRolesByGroupNames(groupNames)
	if err != nil {
		return nil, err
	}
	// 平台管理员账户最大，返回最大角色
	for _, role := range roles {
		if role == constants.RolePlatformAdmin {
			return []string{role}, nil
		}
	}

	return roles, nil

}

// GetRolesByGroupNames 获取用户的角色
func (u *userService) GetRolesByGroupNames(groupNames []string) ([]string, error) {
	if len(groupNames) == 0 {
		return nil, nil
	}

	cacheKey := u.formatCacheKey("user:roles:%s", strings.Join(groupNames, ","))

	result, err := utils.GetOrSetCache(CacheService().CacheInstance(), cacheKey, 5*time.Minute, func() ([]string, error) {
		var ugList []models.UserGroup
		err := dao.DB().Model(&models.UserGroup{}).Where("group_name in ?", groupNames).Distinct("role").Find(&ugList).Error
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
		if groupNameList, err := u.GetGroupNames(username); err == nil {
			if len(groupNameList) > 0 {
				// 查找用户组对应的授权
				if items2, _, err := clusterRole.List(params, func(db *gorm.DB) *gorm.DB {
					return db.Where("authorization_type=? and  username in ? ", constants.ClusterAuthorizationTypeUserGroup, groupNameList)
				}); err == nil {
					items = append(items, items2...)
				}
			}
		}
		return items, nil
	})

	return result, err
}

// GenerateJWTTokenOnlyUserName  生成 Token，仅包含Username
func (u *userService) GenerateJWTTokenOnlyUserName(username string, duration time.Duration) (string, error) {
	if username == "" {
		return "", errors.New("username cannot be empty")
	}
	name := constants.JwtUserName

	var token = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		name:  username,
		"exp": time.Now().Add(duration).Unix(),
	})
	cfg := flag.Init()
	var jwtSecret = []byte(cfg.JwtTokenSecret)
	return token.SignedString(jwtSecret)
}

// GetGroupNames 获取用户所在的用户组
// return: 用户组名称列表
func (u *userService) GetGroupNames(username string) ([]string, error) {
	cacheKey := u.formatCacheKey("user:groupnames:%s", username)

	result, err := utils.GetOrSetCache(CacheService().CacheInstance(), cacheKey, 5*time.Minute, func() ([]string, error) {
		params := &dao.Params{}
		user := &models.User{}
		queryFunc := func(db *gorm.DB) *gorm.DB {
			return db.Select("group_names").Where(" username = ?", username)
		}
		item, err := user.GetOne(params, queryFunc)
		if err != nil {
			return nil, err
		}

		// 如果GroupNames为空返回空切片
		if item.GroupNames == "" {
			return []string{}, nil
		}

		// 将逗号分隔的字符串转为切片并去除空白
		groups := strings.Split(item.GroupNames, ",")
		var cleanGroups []string
		for _, g := range groups {
			if trimmed := strings.TrimSpace(g); trimmed != "" {
				cleanGroups = append(cleanGroups, trimmed)
			}
		}
		return cleanGroups, nil
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
func (u *userService) CheckAndCreateUser(username, source, groups string) error {
	params := dao.BuildDefaultParams()
	user := &models.User{}
	queryFunc := func(db *gorm.DB) *gorm.DB {
		return db.Where("username = ? and source=?", username, source)
	}
	du, err := user.GetOne(params, queryFunc)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 用户不存在，创建新用户
			du = &models.User{
				Username:   username,
				Source:     source,
				GroupNames: groups,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}
			return du.Save(params)
		}
		return err
	}

	// 数据库中已存在用户，检查是否需要更新用户组
	if groups != "" && du.GroupNames != groups {
		// 只更新 group_names 字段，避免更新其他字段导致 password 被清空
		err = du.UpdateColumn("group_names", groups)
		if err != nil {
			klog.V(6).Infof("更新%s用户组出错%v", username, err)
			return err
		}
	}

	return nil
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
func (u *userService) ldapConnection(config *models.LDAPConfig) (*ldap.Conn, error) {
	conn, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port))
	if err != nil {
		klog.Errorf("无法连接到ldap服务器: %v", err)
		return nil, err
	}

	// 设置超时时间
	conn.SetTimeout(5 * time.Second)
	return conn, nil
}

// ldap搜索
func (u *userService) searchRequest(conn *ldap.Conn, username string, config *models.LDAPConfig) (*ldap.Entry, error) {
	var (
		cur              *ldap.SearchResult
		ldapFieldsFilter = []string{
			"dn",
		}
	)

	// 解密管理员密码
	bindPassword, err := utils.AesDecrypt(config.BindPassword)
	if err != nil {
		klog.Errorf("LDAP密码解密失败: %v", err)
		return nil, errors.New("LDAP配置错误")
	}

	err = conn.Bind(config.BindDN, string(bindPassword))
	if err != nil {
		klog.Errorf("LDAP绑定失败: %v", err)
		return nil, errors.New("LDAP认证失败")
	}

	sql := ldap.NewSearchRequest(
		config.BaseDN,
		ldap.ScopeWholeSubtree,
		ldap.DerefAlways,
		0,
		0,
		false,
		fmt.Sprintf("(%v=%v)", config.UserFilter, username),
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

// LoginWithLdap 登录ldap
func (u *userService) LoginWithLdap(username string, password string, cfg *flag.Config) (*ldap.Entry, error) {
	// 从数据库获取启用的LDAP配置
	ldapConfig := &models.LDAPConfig{}
	params := &dao.Params{}

	queryFunc := func(db *gorm.DB) *gorm.DB {
		return db.Where("enabled = ?", true).Order("id desc").Limit(1)
	}

	config, err := ldapConfig.GetOne(params, queryFunc)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			klog.Error("未找到启用的LDAP配置")
			return nil, errors.New("LDAP未配置或未启用")
		}
		klog.Errorf("获取LDAP配置失败: %v", err)
		return nil, err
	}

	// 创建新连接
	conn, err := u.ldapConnection(config)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// 使用连接进行搜索
	userInfo, err := u.searchRequest(conn, username, config)
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

func (u *userService) IsUserPlatformAdmin(user string) bool {
	roles, err := u.GetRolesByUserName(user)
	if err != nil {
		return false
	}
	return slices.Contains(roles, constants.RolePlatformAdmin)
}
