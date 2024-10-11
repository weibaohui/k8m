package utils

import (
	"encoding/base64"
	"strings"
)

// DecodeBase64 解密 Base64 编码的字符串
func DecodeBase64(encoded string) (string, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	return string(decodedBytes), nil
}
func MustDecodeBase64(encoded string) string {
	decodedBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return ""
	}
	// 去除多余的换行符
	decodedString := strings.TrimSpace(string(decodedBytes))
	return decodedString
}
func EncodeBase64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}
