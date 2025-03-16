package user

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/comm/utils/totp"
	"github.com/weibaohui/k8m/pkg/models"
	"gorm.io/gorm"
)

// Disable2FA 禁用2FA
func Disable2FA(c *gin.Context) {
	params := dao.BuildParams(c)
	userID := c.Param("id")

	// 获取用户信息

	user := &models.User{}
	user.ID = uint(utils.ToInt64(userID))
	queryFuncs := genQueryFuncs(c, params)
	user, err := user.GetOne(params, queryFuncs...)
	if err != nil {
		amis.WriteJsonError(c, fmt.Errorf("用户不存在"))
		return
	}

	// 检查是否已启用2FA
	if !user.TwoFAEnabled {
		amis.WriteJsonError(c, fmt.Errorf("2FA未启用"))
		return
	}

	// 清除2FA相关信息
	user.TwoFAEnabled = false
	user.TwoFASecret = ""
	user.TwoFAType = ""
	user.TwoFABackupCodes = ""

	// 保存到数据库
	queryFuncs = append(queryFuncs, func(db *gorm.DB) *gorm.DB {
		return db.Select([]string{"two_fa_enabled", "two_fa_secret", "two_fa_type", "two_fa_backup_codes"})
	})
	err = user.Save(params, queryFuncs...)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOK(c)
}

// Generate2FASecret 生成2FA密钥
func Generate2FASecret(c *gin.Context) {
	params := dao.BuildParams(c)
	userID := c.Param("id")

	// 获取用户信息
	user := &models.User{}
	user.ID = uint(utils.ToInt64(userID))
	queryFuncs := genQueryFuncs(c, params)
	user, err := user.GetOne(params, queryFuncs...)
	if err != nil {
		amis.WriteJsonError(c, fmt.Errorf("用户不存在"))
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

	// 保存到数据库
	queryFuncs = append(queryFuncs, func(db *gorm.DB) *gorm.DB {
		return db.Select([]string{"two_fa_secret", "two_fa_type", "two_fa_backup_codes", "two_fa_enabled", "two_fa_app_name"})
	})
	err = user.Save(params, queryFuncs...)
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
	userID := c.Param("id")

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
	user.ID = uint(utils.ToInt64(userID))
	queryFuncs := genQueryFuncs(c, params)
	user, err := user.GetOne(params, queryFuncs...)
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
	// 保存到数据库
	queryFuncs = append(queryFuncs, func(db *gorm.DB) *gorm.DB {
		return db.Select([]string{"two_fa_enabled", "two_fa_app_name"})
	})
	err = user.Save(params, queryFuncs...)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOK(c)
}
