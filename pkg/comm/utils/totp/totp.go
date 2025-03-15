package totp

import (
	"encoding/base32"
	"fmt"
	"strings"

	"github.com/pquerna/otp/totp"
	"github.com/weibaohui/k8m/pkg/comm/utils"
)

// GenerateSecret 生成TOTP密钥
func GenerateSecret(username string) (string, string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "K8M",
		AccountName: username,
	})
	if err != nil {
		return "", "", err
	}

	// 返回密钥和二维码URL
	return key.Secret(), key.URL(), nil
}

// ValidateCode 验证TOTP代码
func ValidateCode(secret string, code string) bool {
	// 确保密钥是base32编码的
	secret = strings.ToUpper(secret)
	secret = strings.TrimSpace(secret)

	// 如果密钥不是有效的base32编码，返回false
	if _, err := base32.StdEncoding.DecodeString(secret); err != nil {
		return false
	}

	// 验证代码
	return totp.Validate(code, secret)
}

// GenerateBackupCodes 生成备用恢复码
func GenerateBackupCodes(count int) ([]string, error) {
	if count <= 0 {
		return nil, fmt.Errorf("count must be positive")
	}

	codes := make([]string, count)
	for i := 0; i < count; i++ {
		// 生成8位随机数字
		code := utils.RandNLengthString(8)
		codes[i] = code[:8]
	}

	return codes, nil
}
