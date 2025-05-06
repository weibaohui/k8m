package utils

import (
	"context"

	"github.com/weibaohui/k8m/pkg/constants"
)

func GetContextWithAdmin() context.Context {
	return context.WithValue(context.Background(), constants.RolePlatformAdmin, constants.RolePlatformAdmin)
}
