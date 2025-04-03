package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/models"
	"gorm.io/gorm"
)

type roleService struct {
}

var localRoleService = &roleService{}

func RoleService() *roleService {
	return localRoleService
}

// List 获取角色列表
func (r *roleService) List(ctx context.Context, page, pageSize int, keyword string) (interface{}, int64, error) {
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

	role := &models.Role{}
	var queryFuncs []func(*gorm.DB) *gorm.DB

	if keyword != "" {
		queryFuncs = append(queryFuncs, func(db *gorm.DB) *gorm.DB {
			return db.Where("role_name LIKE ?", "%"+keyword+"%")
		})
	}

	roles, total, err := role.List(params, queryFuncs...)
	if err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

// Create 创建角色
func (r *roleService) Create(ctx context.Context, roleName string, operations []string) (string, error) {
	if roleName == "" {
		return "", errors.New("role name cannot be empty")
	}

	operationsJSON, err := json.Marshal(operations)
	if err != nil {
		return "", err
	}
	username := fmt.Sprintf("%s", ctx.Value(constants.JwtUserName))
	role := &models.Role{
		RoleID:     uuid.New().String(),
		RoleName:   roleName,
		Operations: string(operationsJSON),
		CreatedBy:  username,
	}

	params := &dao.Params{}
	if err := role.Save(params); err != nil {
		return "", err
	}

	return role.RoleID, nil
}

// GetByID 根据ID获取角色
func (r *roleService) GetByID(ctx context.Context, roleID string) (interface{}, error) {
	if roleID == "" {
		return nil, errors.New("role ID cannot be empty")
	}

	params := &dao.Params{}
	role := &models.Role{}

	queryFunc := func(db *gorm.DB) *gorm.DB {
		return db.Where("role_id = ?", roleID)
	}

	result, err := role.GetOne(params, queryFunc)
	if err != nil {
		return nil, err
	}

	// 将operations映射为list类型
	var operationsList []string
	err = json.Unmarshal([]byte(result.Operations), &operationsList)
	if err != nil {
		return nil, err
	}
	result.FrontOperations = operationsList

	if result == nil {
		return nil, errors.New("role not found")
	}

	return result, nil
}

// Update 更新角色
func (r *roleService) Update(ctx context.Context, roleID string, roleName string, operations []string) error {
	if roleID == "" {
		return errors.New("role ID cannot be empty")
	}

	params := &dao.Params{}
	role := &models.Role{}

	queryFunc := func(db *gorm.DB) *gorm.DB {
		return db.Where("role_id = ?", roleID)
	}

	existingRole, err := role.GetOne(params, queryFunc)
	if err != nil {
		return err
	}

	if existingRole == nil {
		return errors.New("role not found")
	}

	if roleName != "" {
		existingRole.RoleName = roleName
	}

	if operations != nil {
		operationsJSON, err := json.Marshal(operations)
		if err != nil {
			return err
		}
		existingRole.Operations = string(operationsJSON)
	}

	return existingRole.Save(params)
}

// Delete 删除角色
func (r *roleService) Delete(ctx context.Context, roleID string) error {
	if roleID == "" {
		return errors.New("role ID cannot be empty")
	}

	params := &dao.Params{}
	role := &models.Role{}

	// 检查角色是否存在
	queryFunc := func(db *gorm.DB) *gorm.DB {
		return db.Where("role_id = ?", roleID)
	}

	existingRole, err := role.GetOne(params, queryFunc)
	if err != nil {
		return err
	}

	if existingRole == nil {
		return errors.New("role not found")
	}

	// 删除角色
	return role.Delete(params, roleID)
}
