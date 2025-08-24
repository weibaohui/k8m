package utils

import (
	"crypto/rand"
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

	// 使用 crypto/rand 生成安全随机数
	rangeSize := _max - _min + 1
	randomNum, err := rand.Int(rand.Reader, big.NewInt(int64(rangeSize)))
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

	// 使用 crypto/rand 生成安全随机数
	rangeSize := max - min + 1
	randomNum, err := rand.Int(rand.Reader, big.NewInt(int64(rangeSize)))
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

// intPow is a helper function to calculate power of 10
func intPow(base, exp int) int {
	result := 1
	for exp > 0 {
		result *= base
		exp--
	}
	return result
}
