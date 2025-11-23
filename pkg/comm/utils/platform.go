package utils

import (
	"context"

	"github.com/weibaohui/k8m/pkg/constants"
)

// GetContextWithAdmin 返回一个包含平台管理员角色信息的新上下文对象。
func GetContextWithAdmin() context.Context {
	return context.WithValue(context.Background(), constants.RolePlatformAdmin, constants.RolePlatformAdmin)
}

// GetContextWithAdminFromCtx 返回一个包含平台管理员角色信息的新上下文对象。
func GetContextWithAdminFromCtx(ctx context.Context) context.Context {
	return context.WithValue(ctx, constants.RolePlatformAdmin, constants.RolePlatformAdmin)
}
