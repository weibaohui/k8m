package apikey

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/service"
	"gorm.io/gorm"
)

type Controller struct{}

func RegisterAPIKeysRoutes(mgm *gin.RouterGroup) {
	ctrl := &Controller{}
	mgm.GET("/user/profile/api_keys/list", ctrl.List)
	mgm.POST("/user/profile/api_keys/create", ctrl.Create)
	mgm.POST("/user/profile/api_keys/delete/:id", ctrl.Delete)
}

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
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 从JWT中获取用户信息
	username := c.GetString(constants.JwtUserName)

	// 生成API密钥
	apiKey := &models.ApiKey{
		Username:    username,
		Key:         generateAPIKey(username),
		Description: req.Description,
	}

	// 保存到数据库
	if err := apiKey.Save(params); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOK(c)
}
func generateAPIKey(username string) string {
	// 查询用户对应的集群
	// todo 有效期应该是一个可配置项
	duration := time.Hour * 24 * 365
	token, _ := service.UserService().GenerateJWTTokenOnlyUserName(username, duration)
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
