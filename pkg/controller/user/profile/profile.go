package profile

import (
	"encoding/base64"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/service"
	"gorm.io/gorm"
)

type Controller struct{}

func RegisterProfileRoutes(mgm *gin.RouterGroup) {
	ctrl := &Controller{}
	mgm.GET("/user/profile", ctrl.Profile)
	mgm.GET("/user/profile/cluster/permissions/list", ctrl.ListUserPermissions)
	mgm.POST("/user/profile/update_psw", ctrl.UpdatePsw)
	// user profile 2FA 用户自助操作
	mgm.POST("/user/profile/2fa/generate", ctrl.Generate2FASecret)
	mgm.POST("/user/profile/2fa/disable", ctrl.Disable2FA)
	mgm.POST("/user/profile/2fa/enable", ctrl.Enable2FA)
}

// @Summary 获取用户信息
// @Description 获取当前登录用户的详细信息
// @Security BearerAuth
// @Success 200 {object} string
// @Router /mgm/user/profile [get]
func (uc *Controller) Profile(c *gin.Context) {
	params := dao.BuildParams(c)
	m := &models.User{}

	m.Username = params.UserName
	params.UserName = "" // 避免增加CreatedBy字段,因为用户是管理员创建的，所以不需要CreatedBy字段

	items, total, err := m.List(params, func(db *gorm.DB) *gorm.DB {
		return db.
			Select([]string{"id", "group_names", "two_fa_enabled", "username", "two_fa_type", "two_fa_app_name", "source", "created_at", "updated_at"}).
			Where(m)
	})
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}

// ListUserPermissions 列出当前登录用户所拥有的集群权限
// @Summary 获取用户集群权限
// @Description 列出当前登录用户所拥有的集群权限
// @Security BearerAuth
// @Success 200 {object} string
// @Router /mgm/user/profile/cluster/permissions/list [get]
func (uc *Controller) ListUserPermissions(c *gin.Context) {
	params := dao.BuildParams(c)
	clusters, err := service.UserService().GetClusters(params.UserName)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonList(c, clusters)
}

// PasswordUpdateRequest 密码修改请求结构体
type PasswordUpdateRequest struct {
	Password        string `json:"password" binding:"required"`         // 新密码（加密后）
	ConfirmPassword string `json:"confirmPassword" binding:"required"` // 确认密码（加密后）
}

// @Summary 修改密码
// @Description 修改当前登录用户的密码，需要两次输入密码确认
// @Security BearerAuth
// @Param request body PasswordUpdateRequest true "密码修改请求"
// @Success 200 {object} string "操作成功"
// @Router /mgm/user/profile/update_psw [post]
func (uc *Controller) UpdatePsw(c *gin.Context) {
	params := dao.BuildParams(c)
	req := PasswordUpdateRequest{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 解密两个密码进行比较
	pswBytes, err := utils.AesDecrypt(req.Password)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	confirmPswBytes, err := utils.AesDecrypt(req.ConfirmPassword)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 验证两次密码是否一致
	if string(pswBytes) != string(confirmPswBytes) {
		amis.WriteJsonError(c, fmt.Errorf("两次输入的密码不一致"))
		return
	}

	// 密码 + 盐重新计算
	m := models.User{}
	m.Salt = utils.RandNLengthString(8)
	psw, err := utils.AesEncrypt([]byte(fmt.Sprintf("%s%s", string(pswBytes), m.Salt)))
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	m.Password = base64.StdEncoding.EncodeToString(psw)
	m.Username = params.UserName // 用户名是从token中获取的，不能使用用户前端传递过来的用户名
	params.UserName = ""         // 避免增加CreatedBy字段,因为查询用户集群权限，是管理员授权的，所以不需要CreatedBy字段

	err = dao.DB().Select([]string{"password", "salt"}).Where("username=?", m.Username).Updates(m).Error

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}
