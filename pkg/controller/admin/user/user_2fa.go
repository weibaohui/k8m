package user

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"gorm.io/gorm"
)

// Disable2FA 禁用2FA
func (a *AdminClusterPermission) Disable2FA(c *gin.Context) {
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
