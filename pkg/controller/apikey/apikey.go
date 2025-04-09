package apikey

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/flag"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/service"
	"gorm.io/gorm"
)

// Create 创建API密钥
func Create(c *gin.Context) {
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
	cfg := flag.Init()
	groupNames, _ := service.UserService().GetGroupNames(username)

	roles, _ := service.UserService().GetRolesByGroupNames(groupNames)
	if username == cfg.AdminUserName {
		roles = []string{constants.RolePlatformAdmin}
	}
	// 查询用户对应的集群
	clusters, _ := service.UserService().GetClusters(username)
	duration := time.Hour * 24 * 365
	token, _ := service.UserService().GenerateJWTToken(username, roles, clusters, duration)
	return token
}

// List 获取API密钥列表
func List(c *gin.Context) {
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

func Delete(c *gin.Context) {
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
