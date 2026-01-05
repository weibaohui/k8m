package service

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/plugins/modules/ai/models"
	"gorm.io/gorm"
)

// promptService 提示词服务结构体
// 用于管理和获取AI提示词相关的业务逻辑
type promptService struct {
}

var (
	// instance 单例实例
	instance *promptService
	// once 用于确保单例只被初始化一次
	once sync.Once
)

// GetPromptService 获取提示词服务的单例实例
// 返回值:
//   - *promptService: 提示词服务实例
func GetPromptService() *promptService {
	once.Do(func() {
		instance = &promptService{}
	})
	return instance
}

// GetPrompt 根据提示词类型获取提示词内容
// 参数:
//   - ctx: 上下文对象
//   - promptType: 提示词类型
//
// 返回值:
//   - string: 提示词内容
//   - error: 错误信息
func (p *promptService) GetPrompt(ctx context.Context, promptType constants.AIPromptType) (string, error) {
	// 验证输入参数
	if promptType == "" {
		return "", errors.New("提示词类型不能为空")
	}

	// 创建查询参数，直接根据 prompt_type 字段查询
	params := dao.Params{
		Queries: map[string]any{
			"prompt_type": string(promptType),
		},
		UserName: "", // 不匹配创建人，也就是不管谁创建的
	}

	// 查询数据库中匹配该类型的第一个提示词
	var prompt models.AIPrompt
	queryFunc := func(db *gorm.DB) *gorm.DB {
		return db.Where("prompt_type = ? AND is_enabled = ?", promptType, true)
	}

	result, err := prompt.GetOne(&params, queryFunc)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", fmt.Errorf("未找到类型为 '%s' 的提示词", promptType)
		}
		return "", fmt.Errorf("查询提示词失败: %v", err)
	}

	// 返回提示词内容
	return result.Content, nil
}

// ListPrompts 获取提示词列表
// 参数:
//   - ctx: 上下文对象，用于控制请求的生命周期
//   - promptType: 可选的提示词类型过滤条件，如果为空则返回所有类型
//
// 返回值:
//   - []*models.AIPrompt: 提示词列表
//   - error: 错误信息，如果查询失败则返回相应错误
func (p *promptService) ListPrompts(ctx context.Context, promptType constants.AIPromptType) ([]*models.AIPrompt, error) {
	// 创建查询参数
	params := &dao.Params{}

	// 创建AIPrompt模型实例
	prompt := &models.AIPrompt{}

	// 构建查询条件
	var queryFunc func(*gorm.DB) *gorm.DB
	if promptType != "" {
		queryFunc = func(db *gorm.DB) *gorm.DB {
			return db.Where("prompt_type = ? AND is_enabled = ?", promptType, true)
		}
	} else {
		queryFunc = func(db *gorm.DB) *gorm.DB {
			return db.Where("is_enabled = ?", true)
		}
	}

	// 执行查询获取提示词列表
	results, _, err := prompt.List(params, queryFunc)
	if err != nil {
		return nil, fmt.Errorf("查询提示词列表失败: %w", err)
	}

	return results, nil
}
