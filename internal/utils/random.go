package utils

import (
	"math/rand"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

// RandNDigitInt generates a random number with n digits
func RandNDigitInt(n int) int {
	if n <= 0 {
		return 0
	}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	_min := intPow(10, n-1)
	_max := intPow(10, n) - 1
	return rng.Intn(_max-_min+1) + _min
}

// RandNLengthString generates a random string of specified length using the default charset
func RandNLengthString(n int) string {
	if n <= 0 {
		return ""
	}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	result := make([]byte, n)
	for i := range result {
		result[i] = charset[rng.Intn(len(charset))]
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
