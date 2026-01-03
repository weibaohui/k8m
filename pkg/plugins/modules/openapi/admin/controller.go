package admin

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/plugins/modules/openapi/models"
	"github.com/weibaohui/k8m/pkg/service"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

type Controller struct{}

// Create 创建API密钥
// @Summary 创建API密钥
// @Description 为当前用户创建一个新的API密钥
// @Security BearerAuth
// @Param description body string false "密钥描述"
// @Success 200 {object} string "操作成功"
// @Router /mgm/user/profile/api_keys/create [post]
func (ac *Controller) Create(c *gin.Context) {
	params := dao.BuildParams(c)

	var req struct {
		Description string `json:"description"`
		ExpiresAt   string `json:"expires_at"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 从JWT中获取用户信息
	username := c.GetString(constants.JwtUserName)

	// 解析过期时间
	var expiresAt time.Time
	var err error
	if req.ExpiresAt != "" {
		// 尝试解析完整时间格式 "2006-01-02 15:04:05"
		expiresAt, err = time.Parse("2006-01-02 15:04:05", req.ExpiresAt)
		if err != nil {
			// 如果失败，尝试解析日期格式 "2006-01-02"，解析为当天 23:59:59
			expiresAt, err = time.Parse("2006-01-02", req.ExpiresAt)
			if err != nil {
				amis.WriteJsonError(c, fmt.Errorf("过期时间格式错误: %v", err))
				return
			}
			// 设置为当天 23:59:59
			expiresAt = time.Date(expiresAt.Year(), expiresAt.Month(), expiresAt.Day(), 23, 59, 59, 0, time.Local)
		}
	} else {
		// 默认1年后过期
		expiresAt = time.Now().Add(time.Hour * 24 * 365)
	}

	// 计算duration用于生成JWT
	duration := time.Until(expiresAt)
	if duration <= 0 {
		amis.WriteJsonError(c, fmt.Errorf("过期时间必须大于当前时间"))
		return
	}

	// 生成API密钥
	apiKey := &models.ApiKey{
		Username:    username,
		Key:         generateAPIKey(username, duration),
		Description: req.Description,
		ExpiresAt:   expiresAt,
		LastUsedAt:  time.Now(),
	}
	if apiKey.Key == "" {
		amis.WriteJsonError(c, fmt.Errorf("生成API密钥失败"))
		return
	}
	// 保存到数据库
	if err := apiKey.Save(params); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOK(c)
}
func generateAPIKey(username string, duration time.Duration) string {
	// 查询用户对应的集群
	token, err := service.UserService().GenerateJWTTokenOnlyUserName(username, duration)
	if err != nil {
		klog.Errorf("generateAPIKey error: %v", err)
		return ""
	}
	return token
}

// List 获取API密钥列表
// @Summary 获取API密钥列表
// @Description 获取当前用户的所有API密钥
// @Security BearerAuth
// @Success 200 {object} string
// @Router /mgm/user/profile/api_keys/list [get]
func (ac *Controller) List(c *gin.Context) {
	username := c.GetString(constants.JwtUserName)
	params := dao.BuildParams(c)

	apiKey := &models.ApiKey{}
	list, _, err := apiKey.List(params, func(db *gorm.DB) *gorm.DB {
		return db.Where("username = ?", username)
	})

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonData(c, list)
}

// @Summary 删除API密钥
// @Description 删除指定ID的API密钥
// @Security BearerAuth
// @Param id path string true "API密钥ID"
// @Success 200 {object} string "操作成功"
// @Router /mgm/user/profile/api_keys/delete/{id} [post]
func (ac *Controller) Delete(c *gin.Context) {
	id := c.Param("id")
	params := dao.BuildParams(c)

	apiKey := &models.ApiKey{}

	err := apiKey.Delete(params, id)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}
