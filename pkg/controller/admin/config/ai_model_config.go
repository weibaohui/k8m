package config

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
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
	r.POST("/ai/model/id/:id/think/:status", ctrl.QuickSave)

}

func (m *AIModelConfigController) QuickSave(c *gin.Context) {
	id := c.Param("id")
	status := c.Param("status")

	var entity models.AIModelConfig
	entity.ID = utils.ToUInt(id)

	if status == "true" {
		entity.Think = true
	} else {
		entity.Think = false
	}
	err := dao.DB().Model(&entity).Select("think").Updates(entity).Error

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonErrorOrOK(c, err)
}

// Save 创建或更新AI模型配置
func (m *AIModelConfigController) Save(c *gin.Context) {
	params := dao.BuildParams(c)

	var config models.AIModelConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 添加业务逻辑验证
	if config.ApiURL == "" {
		amis.WriteJsonError(c, fmt.Errorf("API URL不能为空"))
		return
	}
	if config.ApiKey == "" {
		amis.WriteJsonError(c, fmt.Errorf("API Key不能为空"))
		return
	}
	if config.Temperature < 0 || config.Temperature > 2 {
		amis.WriteJsonError(c, fmt.Errorf("Temperature参数应在0-2之间"))
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
