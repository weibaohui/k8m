package config

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/models"
	"gorm.io/gorm"

	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
)

// AIModelConfigController 用于管理AI模型配置

type AIModelConfigController struct {
	DB *gorm.DB
}

// RegisterAIModelConfigRoutes 注册路由
func RegisterAIModelConfigRoutes(r *gin.RouterGroup) {
	ctrl := &AIModelConfigController{DB: dao.DB()}
	r.GET("/ai/model/list", ctrl.List)
	r.POST("/ai/model/save", ctrl.Save)
	r.POST("/ai/model/delete/:ids", ctrl.Delete)
}

// Save 创建或更新AI模型配置
func (m *AIModelConfigController) Save(c *gin.Context) {
	params := dao.BuildParams(c)

	var config models.AIModelConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 保存到数据库
	if err := config.Save(params); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOK(c)
}

// List 获取API密钥列表
func (m *AIModelConfigController) List(c *gin.Context) {
	params := dao.BuildParams(c)

	config := &models.AIModelConfig{}
	items, total, err := config.List(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}

func (m *AIModelConfigController) Delete(c *gin.Context) {
	ids := c.Param("ids")
	params := dao.BuildParams(c)
	config := &models.AIModelConfig{}

	err := config.Delete(params, ids)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}
