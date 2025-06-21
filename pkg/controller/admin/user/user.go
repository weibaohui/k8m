package user

import (
	"encoding/base64"
	"fmt"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/service"
	"gorm.io/gorm"
)

type AdminUserController struct {
}

// AdminUser 用于用户相关接口
// 路由注册函数
func RegisterAdminUserRoutes(admin *gin.RouterGroup) {

	ctrl := AdminUserController{}
	// user 平台管理员可操作，管理用户
	admin.GET("/user/list", ctrl.List)
	admin.POST("/user/save/id/:id/status/:disabled", ctrl.UserStatusQuickSave)
	admin.POST("/user/save", ctrl.Save)
	admin.POST("/user/delete/:ids", ctrl.Delete)
	admin.POST("/user/update_psw/:id", ctrl.UpdatePsw)
	admin.GET("/user/option_list", ctrl.UserOptionList)
	// 2FA 平台管理员可操作，管理用户
	admin.POST("/user/2fa/disable/:id", ctrl.Disable2FA)

}

// @Summary 获取用户列表
// @Description 获取所有用户信息
// @Security BearerAuth
// @Success 200 {object} []models.User
// @Router /admin/user/list [get]
func (a *AdminUserController) List(c *gin.Context) {
	params := dao.BuildParams(c)
	m := &models.User{}

	queryFuncs := genQueryFuncs(c, params)
	queryFuncs = append(queryFuncs, func(db *gorm.DB) *gorm.DB {
		return db.Select([]string{"id", "group_names", "two_fa_enabled", "username", "two_fa_type", "two_fa_app_name", "source", "created_at", "updated_at", "disabled"})
	})
	items, total, err := m.List(params, queryFuncs...)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}

// @Summary 保存用户
// @Description 新增或更新用户信息
// @Security BearerAuth
// @Accept json
// @Param data body models.User true "用户信息"
// @Success 200 {object} map[string]interface{}
// @Router /admin/user/save [post]
func (a *AdminUserController) Save(c *gin.Context) {
	params := dao.BuildParams(c)
	m := models.User{}
	err := c.ShouldBindJSON(&m)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
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
	// 清除用户的缓存
	service.UserService().ClearCacheByKey(m.Username)
	amis.WriteJsonData(c, gin.H{
		"id": m.ID,
	})
}

// @Summary 删除用户
// @Description 根据ID批量删除用户
// @Security BearerAuth
// @Param ids path string true "用户ID，多个用逗号分隔"
// @Success 200 {object} string
// @Router /admin/user/delete/{ids} [post]
func (a *AdminUserController) Delete(c *gin.Context) {
	ids := c.Param("ids")
	params := dao.BuildParams(c)
	m := &models.User{}

	queryFuncs := genQueryFuncs(c, params)

	err := m.Delete(params, ids, queryFuncs...)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	// 清除用户的缓存
	service.UserService().ClearCacheByKey(m.Username)
	amis.WriteJsonOK(c)
}

// @Summary 更新用户密码
// @Description 根据ID更新用户密码
// @Security BearerAuth
// @Accept json
// @Param id path string true "用户ID"
// @Param data body models.User true "新密码信息"
// @Success 200 {object} string
// @Router /admin/user/update_psw/{id} [post]
func (a *AdminUserController) UpdatePsw(c *gin.Context) {

	id := c.Param("id")
	params := dao.BuildParams(c)
	m := models.User{}
	err := c.ShouldBindJSON(&m)
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

func genQueryFuncs(c *gin.Context, params *dao.Params) []func(*gorm.DB) *gorm.DB {
	params.UserName = ""
	queryFuncs := []func(*gorm.DB) *gorm.DB{
		func(db *gorm.DB) *gorm.DB {
			return db
		},
	}
	return queryFuncs
}

func (a *AdminUserController) UserOptionList(c *gin.Context) {
	params := dao.BuildParams(c)
	m := &models.User{}
	items, _, err := m.List(params, func(db *gorm.DB) *gorm.DB {
		return db.Distinct("username")
	})
	if err != nil {
		amis.WriteJsonData(c, gin.H{
			"options": make([]map[string]string, 0),
		})
		return
	}
	var names []map[string]string
	for _, n := range items {
		names = append(names, map[string]string{
			"label": n.Username,
			"value": n.Username,
		})
	}
	slice.SortBy(names, func(a, b map[string]string) bool {
		return a["label"] < b["label"]
	})
	amis.WriteJsonData(c, gin.H{
		"options": names,
	})
}

// Disable2FA 禁用2FA
// @Summary 禁用用户2FA
// @Description 禁用指定用户的二步验证
// @Security BearerAuth
// @Param id path string true "用户ID"
// @Success 200 {object} string
// @Router /admin/user/2fa/disable/{id} [post]
func (a *AdminUserController) Disable2FA(c *gin.Context) {
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

func (a *AdminUserController) UserStatusQuickSave(c *gin.Context) {
	id := c.Param("id")
	disabled := c.Param("disabled")

	var entity models.User
	entity.ID = utils.ToUInt(id)

	if disabled == "true" {
		entity.Disabled = true
	} else {
		entity.Disabled = false
	}
	err := dao.DB().Model(&entity).Select("disabled").Updates(entity).Error

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonErrorOrOK(c, err)
}
