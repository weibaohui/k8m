package apikey

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/models"
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
		Key:         utils.RandNLengthString(32), // 生成32位随机字符串作为密钥
		Description: req.Description,
	}

	// 保存到数据库
	if err := apiKey.Save(params); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOK(c)
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
	ids := c.Param("ids")
	username := c.GetString(constants.JwtUserName)
	params := dao.BuildParams(c)

	apiKey := &models.ApiKey{}

	err := apiKey.Delete(params, ids, func(db *gorm.DB) *gorm.DB {
		return db.Where("username = ?", username)
	})
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}
