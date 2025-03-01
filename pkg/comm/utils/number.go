package utils

import (
	"strconv"
	"unicode"
)

// ExtractNumbers 函数用于从版本号中提取数字并转换为整数
func ExtractNumbers(version string) (int, error) {
	result := ""
	// 遍历版本号字符串的每个字符
	for _, char := range version {
		// 判断字符是否为数字
		if unicode.IsDigit(char) {
			result += string(char)
		}
	}
	// 如果提取的结果为空字符串，返回 0
	if result == "" {
		return 0, nil
	}
	// 将提取到的数字字符串转换为整数
	num, err := strconv.Atoi(result)
	if err != nil {
		return 0, err
	}
	return num, nil
}
