package utils

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
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

func UrlSafeBase64Encode(s string) string {
	encoded := base64.URLEncoding.EncodeToString([]byte(s))
	// 去掉填充的等号，使其更URL安全
	encoded = strings.TrimRight(encoded, "=")
	return encoded
}
func UrlSafeBase64Decode(s string) (string, error) {
	// 补等号
	if m := len(s) % 4; m != 0 {
		s += strings.Repeat("=", 4-m)
	}
	decodedBytes, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(decodedBytes), nil
}

// MD5Hex 中文函数注释：对输入字符串计算 MD5，并返回 32 位小写十六进制字符串。
func MD5Hex(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}
