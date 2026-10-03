package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
)

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

// RandNDigitInt generates a random number with n digits
func RandNDigitInt(n int) int {
	if n <= 0 {
		return 0
	}
	_min := intPow(10, n-1)
	_max := intPow(10, n) - 1

	// 使用 int64 进行计算以避免溢出
	r := int64(_max) - int64(_min) + 1
	if r <= 0 {
		// 如果范围无效，返回最小值
		return _min
	}

	// 使用 crypto/rand 生成安全随机数
	randomNum, err := rand.Int(rand.Reader, big.NewInt(r))
	if err != nil {
		// 如果生成随机数失败，返回最小值
		return _min
	}
	return int(randomNum.Int64()) + _min
}

// RandInt generates a random number between min and max
func RandInt(min, max int) int {
	if min > max {
		min, max = max, min
	}

	// 使用 int64 进行计算以避免溢出
	r := int64(max) - int64(min) + 1
	if r <= 0 {
		// 如果范围无效，返回最小值
		return min
	}

	// 使用 crypto/rand 生成安全随机数
	randomNum, err := rand.Int(rand.Reader, big.NewInt(r))
	if err != nil {
		// 如果生成随机数失败，返回最小值
		return min
	}
	return int(randomNum.Int64()) + min
}

// RandNLengthString generates a random string of specified length using the default charset
func RandNLengthString(n int) string {
	if n <= 0 {
		return ""
	}

	result := make([]byte, n)
	for i := range result {
		// 使用 crypto/rand 生成安全随机索引
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			// 如果生成随机数失败，使用第一个字符
			result[i] = charset[0]
			continue
		}
		result[i] = charset[randomIndex.Int64()]
	}
	return string(result)
}

// GenerateSecureToken 生成密码学安全的随机凭证（256位熵，十六进制编码，带 k8m_ 前缀便于识别）。
// 与 RandNLengthString 不同：本函数在任何错误下都返回 error 而不是弱回退值，用于会话凭证签发。
func GenerateSecureToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("生成随机凭证失败: %w", err)
	}
	return "k8m_" + hex.EncodeToString(buf), nil
}

// intPow is a helper function to calculate power of 10
func intPow(base, exp int) int {
	result := 1
	for exp > 0 {
		result *= base
		exp--
	}
	return result
}
