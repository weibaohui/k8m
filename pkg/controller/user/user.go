package user

import (
	"encoding/base64"
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

func List(c *gin.Context) {
	params := dao.BuildParams(c)
	m := &models.User{}

	queryFuncs := genQueryFuncs(c, params)

	items, total, err := m.List(params, queryFuncs...)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}
func Save(c *gin.Context) {
	params := dao.BuildParams(c)
	m := &models.User{}
	err := c.ShouldBindJSON(&m)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 用户名不能为admin
	if m.Username == "admin" {
		amis.WriteJsonError(c, fmt.Errorf("用户名不能为admin"))
		return
	}

	_, role := amis.GetLoginUser(c)

	if m.ID == 0 {
		// 新增
		switch role {
		case models.RoleClusterAdmin, models.RoleClusterReadonly:
			amis.WriteJsonError(c, fmt.Errorf("非管理员不能新增用户"))
			return
		}
	} else {
		switch role {
		case models.RoleClusterAdmin, models.RoleClusterReadonly:
			var originalUser models.User
			err = dao.DB().Model(&models.User{}).
				Where("id=?", m.ID).
				Find(&originalUser).Error
			if err != nil {
				amis.WriteJsonError(c, fmt.Errorf("无此用户[%d]", m.ID))
				return
			}

			// 如需限制不能修改的字段，请在下面赋值。
			// 用户名、角色不能修改
			m.Username = originalUser.Username
		}

	}

	queryFuncs := genQueryFuncs(c, params)

	// 保存的时候需要单独处理
	queryFuncs = append(queryFuncs, func(db *gorm.DB) *gorm.DB {
		if m.ID == 0 {
			// 新增
			return db
		} else {
			// 修改
			return db.Select([]string{"username", "group_names"})
		}
	})
	err = m.Save(params, queryFuncs...)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, gin.H{
		"id": m.ID,
	})
}
func Delete(c *gin.Context) {
	ids := c.Param("ids")
	params := dao.BuildParams(c)
	m := &models.User{}

	_, role := amis.GetLoginUser(c)

	switch role {
	case models.RoleClusterReadonly, models.RoleClusterAdmin:
		// 非平台管理员，不能删除
		amis.WriteJsonError(c, fmt.Errorf("非管理员不能删除用户"))
		return
	}

	queryFuncs := genQueryFuncs(c, params)

	err := m.Delete(params, ids, queryFuncs...)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}
func UpdatePsw(c *gin.Context) {

	id := c.Param("id")
	params := dao.BuildParams(c)
	m := &models.User{}
	err := c.ShouldBindJSON(m)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	m.ID = uint(utils.ToInt64(id))

	// 密码 + 盐重新计算
	pswBytes, err := utils.AesDecrypt(m.Password)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	m.Salt = utils.RandNLengthString(8)
	psw, err := utils.AesEncrypt([]byte(fmt.Sprintf("%s%s", string(pswBytes), m.Salt)))
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	m.Password = base64.StdEncoding.EncodeToString(psw)

	queryFuncs := genQueryFuncs(c, params)
	queryFuncs = append(queryFuncs, func(db *gorm.DB) *gorm.DB {
		return db.Select([]string{"password", "salt"}).Updates(m)
	})
	err = m.Save(params, queryFuncs...)
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
	// TODO 校验是否当前用户，只能给自己开启
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

	// 保存到数据库
	queryFuncs = append(queryFuncs, func(db *gorm.DB) *gorm.DB {
		return db.Select([]string{"two_fa_secret", "two_fa_type", "two_fa_backup_codes"})
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
		Code string `json:"code"`
	}
	var req Enable2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 获取用户信息
	// TODO 校验是否当前用户，只能给自己开启

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

	// 保存到数据库
	queryFuncs = append(queryFuncs, func(db *gorm.DB) *gorm.DB {
		return db.Select([]string{"two_fa_enabled"})
	})
	err = user.Save(params, queryFuncs...)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOK(c)
}

func genQueryFuncs(c *gin.Context, params *dao.Params) []func(*gorm.DB) *gorm.DB {
	//  管理页面，判断是否管理员，看到所有的用户，
	user, role := amis.GetLoginUser(c)
	var queryFuncs []func(*gorm.DB) *gorm.DB
	switch role {
	case models.RolePlatformAdmin:
		params.UserName = ""
		queryFuncs = []func(*gorm.DB) *gorm.DB{
			func(db *gorm.DB) *gorm.DB {
				return db
			},
		}
	case models.RoleClusterAdmin, models.RoleClusterReadonly:
		queryFuncs = []func(*gorm.DB) *gorm.DB{
			func(db *gorm.DB) *gorm.DB {
				return db.Where("username=?", user)
			},
		}

	}
	return queryFuncs
}
