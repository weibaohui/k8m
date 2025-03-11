package utils

import (
	"strings"
)

// IsDecimal 检查字符串是否为小数
// 如果字符串包含小数点且小数点后有数字，则返回true
func IsDecimal(s string) bool {
	// 检查是否包含小数点
	if strings.Contains(s, ".") {
		// 分割字符串，获取小数点后的部分
		parts := strings.Split(s, ".")
		// 如果小数点后有数字，则为小数
		if len(parts) > 1 && len(parts[1]) > 0 {
			return true
		}
	}
	return false
}