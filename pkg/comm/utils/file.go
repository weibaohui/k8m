package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// ExpandHomePath 处理路径中的 ~ 符号，将其展开为用户主目录
func ExpandHomePath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, path[2:]), nil
	}
	return path, nil
}
