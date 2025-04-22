package utils

import (
	"context"

	"github.com/weibaohui/k8m/pkg/constants"
)

func GetContextWithAdmin() context.Context {
	ctx := context.WithValue(context.Background(), constants.JwtUserRole, constants.RolePlatformAdmin)
	ctx = context.WithValue(ctx, constants.JwtUserName, "admin")
	return ctx
}
