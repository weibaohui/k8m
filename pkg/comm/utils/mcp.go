package utils

import (
	"fmt"
	"strings"
)

// BuildMCPToolName 构建完整的工具名称
func BuildMCPToolName(toolName, serverName string) string {
	return fmt.Sprintf("%s@%s", toolName, serverName)
}

// ParseMCPToolName 从完整的工具名称中解析出服务器名称
func ParseMCPToolName(fullToolName string) (toolName, serverName string, err error) {
	lastIndex := strings.LastIndex(fullToolName, "@")
	if lastIndex == -1 {
		return "", "", fmt.Errorf("invalid tool name format: %s", fullToolName)
	}
	return fullToolName[:lastIndex], fullToolName[lastIndex+1:], nil
}
