package profile

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/comm/utils/totp"
	"github.com/weibaohui/k8m/pkg/models"
)

// Disable2FA 禁用2FA
func Disable2FA(c *gin.Context) {
	params := dao.BuildParams(c)

	// 获取用户信息

	user := &models.User{}
	user.Username = params.UserName
	params.UserName = "" //避免增加CreatedBy字段,因为查询用户集群权限，是管理员授权的，所以不需要CreatedBy字段

	// 清除2FA相关信息
	user.TwoFAEnabled = false
	user.TwoFASecret = ""
	user.TwoFAType = ""
	user.TwoFABackupCodes = ""

	err := dao.DB().
		Select([]string{"two_fa_enabled", "two_fa_secret", "two_fa_type", "two_fa_backup_codes"}).
		Where("username=?", user.Username).
		Updates(user).Error

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOK(c)
}

// Generate2FASecret 生成2FA密钥
func Generate2FASecret(c *gin.Context) {
	params := dao.BuildParams(c)

	// 获取用户信息
	user := &models.User{}
	user.Username = params.UserName
	params.UserName = "" //避免增加CreatedBy字段,因为查询用户集群权限，是管理员授权的，所以不需要CreatedBy字段

	err := dao.DB().
		Where("username=?", user.Username).
		First(user).Error

	if err != nil {
		amis.WriteJsonError(c, fmt.Errorf("用户不存在"))
		return
	}
	// 检查是否已启用2FA
	if user.TwoFAEnabled {
		amis.WriteJsonError(c, fmt.Errorf("2步验证已绑定，如需重新绑定，请先关闭。"))
		return
	}

	// 生成TOTP密钥和二维码URL
	secret, qrURL, err := totp.GenerateSecret(user.Username)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 生成备用恢复码
	backupCodes, err := totp.GenerateBackupCodes(8)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 更新用户2FA信息
	user.TwoFASecret = secret
	user.TwoFAType = "totp"
	user.TwoFABackupCodes = strings.Join(backupCodes, ",")
	user.TwoFAEnabled = false

	err = dao.DB().
		Select([]string{"two_fa_enabled", "two_fa_secret", "two_fa_type", "two_fa_backup_codes"}).
		Where("username=?", user.Username).
		Updates(user).Error

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 返回数据给前端
	amis.WriteJsonData(c, gin.H{
		"secret":       secret,
		"qr_url":       qrURL,
		"backup_codes": backupCodes,
	})

}

// Enable2FA 验证并启用2FA
func Enable2FA(c *gin.Context) {
	params := dao.BuildParams(c)

	// 获取用户输入的验证码
	type Enable2FARequest struct {
		Code    string `json:"code"`
		AppName string `json:"app_name"`
	}
	var req Enable2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 获取用户信息

	user := &models.User{}
	user.Username = params.UserName
	params.UserName = "" //避免增加CreatedBy字段,因为查询用户集群权限，是管理员授权的，所以不需要CreatedBy字段

	err := dao.DB().
		Where("username=?", user.Username).
		First(user).Error

	if err != nil {
		amis.WriteJsonError(c, fmt.Errorf("用户不存在"))
		return
	}

	// 验证TOTP代码
	if !totp.ValidateCode(user.TwoFASecret, req.Code) {
		amis.WriteJsonError(c, fmt.Errorf("验证码无效"))
		return
	}

	// 启用2FA
	user.TwoFAEnabled = true
	user.TwoFAAppName = req.AppName

	err = dao.DB().
		Select([]string{"two_fa_enabled", "two_fa_app_name"}).
		Where("username=?", user.Username).
		Updates(user).Error

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOK(c)
}
