package config

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/service"
)

// AIModelConfigController 用于管理AI模型配置

type AIModelConfigController struct {
}

// RegisterAIModelConfigRoutes 注册路由
func RegisterAIModelConfigRoutes(admin *gin.RouterGroup) {
	ctrl := &AIModelConfigController{}
	admin.GET("/ai/model/list", ctrl.List)
	admin.POST("/ai/model/save", ctrl.Save)
	admin.POST("/ai/model/delete/:ids", ctrl.Delete)
	admin.POST("/ai/model/id/:id/think/:status", ctrl.QuickSave)
	admin.POST("/ai/model/test/id/:id", ctrl.TestConnection)

}

// @Summary 快速保存AI模型思考状态
// @Security BearerAuth
// @Param id path int true "模型ID"
// @Param status path string true "状态，例如：true、false"
// @Success 200 {object} string
// @Router /admin/ai/model/id/{id}/think/{status} [post]
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

// @Summary 测试AI模型连接
// @Security BearerAuth
// @Param id path int true "模型ID"
// @Success 200 {object} string
// @Router /admin/ai/model/test/id/{id} [post]
func (m *AIModelConfigController) TestConnection(c *gin.Context) {
	id := c.Param("id")

	var entity models.AIModelConfig
	entity.ID = utils.ToUInt(id)

	err := dao.DB().Model(&entity).First(&entity).Error

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	client, err := service.AIService().TestClient(entity.ApiURL, entity.ApiKey, entity.ApiModel)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	ctx := amis.GetContextWithUser(c)
	completion, err := client.GetCompletion(ctx, "你是谁？")
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	if completion != "" {
		amis.WriteJsonOKMsg(c, "测试返回成功:"+completion)
		return
	}
	amis.WriteJsonError(c, fmt.Errorf("测试失败"))
}

// @Summary 创建或更新AI模型配置
// @Security BearerAuth
// @Success 200 {object} string
// @Router /admin/ai/model/save [post]
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

// @Summary 获取AI模型配置列表
// @Security BearerAuth
// @Success 200 {object} string
// @Router /admin/ai/model/list [get]
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

// @Summary 删除AI模型配置
// @Security BearerAuth
// @Param ids path string true "模型ID，多个用逗号分隔"
// @Success 200 {object} string
// @Router /admin/ai/model/delete/{ids} [post]
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
