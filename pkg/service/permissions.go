package service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/models"
	"gorm.io/gorm"
)

type permissionService struct {
}

// 权限缓存项
type CacheItem struct {
	UserPermissions      map[string]map[string][]string // map[clusterID]map[namespace][]operations
	UserGroupPermissions map[string]map[string][]string // map[clusterID]map[namespace][]operations
	LastUpdated          time.Time
}

// 用户权限缓存
var permissionCache = make(map[string]*CacheItem)

var localPermissionService = &permissionService{}

func PermissionService() *permissionService {
	return localPermissionService
}

// ClearUserCache 清除用户权限缓存
func (s *permissionService) ClearUserCache(userID string) {
	delete(permissionCache, userID)
}

// CheckUserIsAdmin 检查用户是否为管理员
func (s *permissionService) CheckUserIsAdmin(username string) (bool, error) {
	//TODO: 暂时不修改原user表
	// user := &models.User{}
	// user, err := user.GetByUsername(s.DBParams, username)
	// if err != nil {
	// 	return false, err
	// }
	// return user.IsAdmin, nil
	return true, nil
}

// GetUserID 通过用户名获取用户ID
func (s *permissionService) GetUserID(username string) (string, error) {
	user := &models.User{}
	params := &dao.Params{}
	user, err := user.GetByUsername(params, username)
	if err != nil {
		return "", err
	}
	return strconv.FormatUint(uint64(user.ID), 10), nil
}

// CheckPermission 检查用户是否有权限在特定集群和命名空间执行操作
func (s *permissionService) CheckPermission(userID, clusterID, namespace, operation string) (bool, error) {
	// 检查缓存是否存在且有效
	cacheItem, exists := permissionCache[userID]
	if !exists || time.Since(cacheItem.LastUpdated) > 5*time.Minute {
		// 缓存不存在或已过期，重新加载权限
		userPerms, groupPerms, err := s.loadUserPermissions(userID)
		if err != nil {
			return false, err
		}

		permissionCache[userID] = &CacheItem{
			UserPermissions:      userPerms,
			UserGroupPermissions: groupPerms,
			LastUpdated:          time.Now(),
		}
		cacheItem = permissionCache[userID]
	}

	// 检查用户直接权限
	if allowed := s.checkCachedPermission(cacheItem.UserPermissions, clusterID, namespace, operation); allowed {
		return true, nil
	}

	// 检查用户组权限
	return s.checkCachedPermission(cacheItem.UserGroupPermissions, clusterID, namespace, operation), nil
}

// loadUserPermissions 加载用户的所有权限
func (s *permissionService) loadUserPermissions(userID string) (map[string]map[string][]string, map[string]map[string][]string, error) {
	userPerms := make(map[string]map[string][]string)
	groupPerms := make(map[string]map[string][]string)

	params := &dao.Params{}

	// 1. 加载用户直接权限
	permModel := &models.ClusterPermission{}
	userPermissions, err := permModel.ListUserPermissions(params, userID)
	if err != nil {
		return nil, nil, err
	}

	// 处理用户直接权限
	for _, perm := range userPermissions {
		roleModel := &models.Role{}
		role, err := roleModel.GetOne(params, func(db *gorm.DB) *gorm.DB {
			return db.Where("role_id = ?", perm.RoleID)
		})
		if err != nil {
			continue
		}

		operations, err := role.GetOperations()
		if err != nil {
			continue
		}

		// 初始化集群映射
		if _, exists := userPerms[perm.ClusterID]; !exists {
			userPerms[perm.ClusterID] = make(map[string][]string)
		}

		// 添加命名空间操作
		userPerms[perm.ClusterID][perm.Namespace] = operations
	}

	// 2. 加载用户组权限
	ugModel := &models.UserGroupAssociation{}
	userGroups, err := ugModel.GetGr(params, userID)
	if err != nil {
		return userPerms, groupPerms, nil
	}

	// 如果用户不属于任何组，直接返回
	if len(userGroups) == 0 {
		return userPerms, groupPerms, nil
	}

	// 提取组ID
	groupIDs := make([]string, 0, len(userGroups))
	for _, ug := range userGroups {
		groupIDs = append(groupIDs, ug.GroupID)
	}

	// 加载组权限
	groupPermissions, err := permModel.ListGroupPermissions(params, groupIDs)
	if err != nil {
		return userPerms, groupPerms, nil
	}

	// 处理组权限
	for _, perm := range groupPermissions {
		roleModel := &models.Role{}
		role, err := roleModel.GetOne(params, func(db *gorm.DB) *gorm.DB {
			return db.Where("role_id = ?", perm.RoleID)
		})
		if err != nil {
			continue
		}

		operations, err := role.GetOperations()
		if err != nil {
			continue
		}

		// 初始化集群映射
		if _, exists := groupPerms[perm.ClusterID]; !exists {
			groupPerms[perm.ClusterID] = make(map[string][]string)
		}

		// 添加命名空间操作
		groupPerms[perm.ClusterID][perm.Namespace] = operations
	}

	return userPerms, groupPerms, nil
}

// checkCachedPermission 检查缓存的权限
func (s *permissionService) checkCachedPermission(perms map[string]map[string][]string, clusterID, namespace, operation string) bool {
	// 检查是否有此集群的权限
	clusterPerms, hasCluster := perms[clusterID]
	if !hasCluster {
		return false
	}

	// 优先检查精确命名空间匹配
	if namespaceOps, hasNS := clusterPerms[namespace]; hasNS {
		for _, op := range namespaceOps {
			if op == "*" || op == operation {
				return true
			}
		}
	}

	// 检查通配符命名空间
	if wildcardOps, hasWildcard := clusterPerms["*"]; hasWildcard {
		for _, op := range wildcardOps {
			if op == "*" || op == operation {
				return true
			}
		}
	}

	return false
}

// CreatePermissionBinding 创建新的权限绑定
func (s *permissionService) CreatePermissionBinding(targetType, targetID, clusterID, namespace, roleID, createdBy string) (string, error) {
	// 验证目标类型
	if targetType != "user" && targetType != "group" {
		return "", errors.New("目标类型必须是 'user' 或 'group'")
	}

	params := &dao.Params{}

	// 验证角色存在
	roleModel := &models.Role{}
	_, err := roleModel.GetOne(params, func(db *gorm.DB) *gorm.DB {
		return db.Where("role_id = ?", roleID)
	})
	if err != nil {
		return "", errors.New("指定的角色不存在")
	}

	existingPerm := &models.ClusterPermission{}
	res, err := dao.GenericGetOne(params, existingPerm, func(db *gorm.DB) *gorm.DB {
		return db.Where("target_type = ? AND target_id = ? AND cluster_id = ? AND namespace = ? AND role_id = ?",
			targetType, targetID, clusterID, namespace, roleID)
	})
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", errors.New("查询错误")
	}
	if res.ID > 0 {
		return "", errors.New("请勿重复创建")
	}

	// 创建绑定ID
	bindingID := strings.Replace(uuid.New().String(), "-", "", -1)[:32]

	// 创建权限绑定
	perm := &models.ClusterPermission{
		BindingID:  bindingID,
		TargetType: targetType,
		TargetID:   targetID,
		ClusterID:  clusterID,
		Namespace:  namespace,
		RoleID:     roleID,
		CreatedAt:  time.Now(),
		CreatedBy:  createdBy,
	}

	// 保存到数据库
	err = dao.GenericSave(params, perm)
	if err != nil {
		return "", err
	}

	// 如果是用户类型，清除该用户的缓存
	if targetType == "user" {
		s.ClearUserCache(targetID)
	} else {
		// 如果是组类型，清除该组所有成员的缓存
		// 这里简化处理，清除所有缓存
		permissionCache = make(map[string]*CacheItem)
	}

	return bindingID, nil
}

// DeletePermissionBinding 删除权限绑定
func (s *permissionService) DeletePermissionBinding(bindingID string) error {
	params := &dao.Params{}
	perm := &models.ClusterPermission{}

	permFound, err := dao.GenericGetOne(params, perm, func(db *gorm.DB) *gorm.DB {
		return db.Where("binding_id = ?", bindingID)
	})
	if err != nil {
		return err
	}

	// 删除权限绑定
	err = dao.GenericDelete(params, perm, []int64{int64(permFound.ID)})
	if err != nil {
		return err
	}

	// 清除相关用户的缓存
	if permFound.TargetType == "user" {
		s.ClearUserCache(permFound.TargetID)
	} else {
		// 如果是组类型，简化处理，清除所有缓存
		permissionCache = make(map[string]*CacheItem)
	}

	return nil
}

// ListPermissionBindings 列出权限绑定
func (s *permissionService) ListPermissionBindings(ctx context.Context, page, pageSize int, keyword string) (interface{}, int64, error) {
	if page == 0 {
		page = 1
	}

	if pageSize == 0 {
		pageSize = 10
	}

	params := &dao.Params{
		Page:    page,
		PerPage: pageSize,
	}

	perm := &models.ClusterPermission{}
	var queryFuncs []func(*gorm.DB) *gorm.DB

	if keyword != "" {
		queryFuncs = append(queryFuncs, func(db *gorm.DB) *gorm.DB {
			return db.Where("binding_id LIKE ? OR target_id LIKE ? OR cluster_id LIKE ?",
				"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
		})
	}

	perms, total, err := perm.List(params, queryFuncs...)
	if err != nil {
		return nil, 0, err
	}

	return perms, total, nil
}
