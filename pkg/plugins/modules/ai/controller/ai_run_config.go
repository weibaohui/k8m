package controller

import (
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/plugins/modules/ai/models"
	"github.com/weibaohui/k8m/pkg/plugins/modules/ai/service"
	"github.com/weibaohui/k8m/pkg/response"
)

// AIRunConfigController AI运行配置控制器
// 提供AI运行配置的获取和更新功能

type AIRunConfigController struct{}

// @Summary 获取AI运行配置
// @Security BearerAuth
// @Success 200 {object} models.AIRunConfig
// @Router /admin/plugins/ai/run_config [get]
func (c *AIRunConfigController) GetRunConfig(ctx *response.Context) {
	config, err := service.AIRunConfigService().GetDefault()
	if err != nil {
		amis.WriteJsonError(ctx, err)
		return
	}
	amis.WriteJsonData(ctx, config)
}

// @Summary 更新AI运行配置
// @Security BearerAuth
// @Param config body models.AIRunConfig true "AI运行配置"
// @Success 200 {object} string
// @Router /admin/plugins/ai/run_config [post]
func (c *AIRunConfigController) UpdateRunConfig(ctx *response.Context) {
	var config models.AIRunConfig
	if err := ctx.ShouldBindJSON(&config); err != nil {
		amis.WriteJsonError(ctx, err)
		return
	}

	if err := service.AIRunConfigService().SaveDefault(&config); err != nil {
		amis.WriteJsonError(ctx, err)
		return
	}

	// 更新Flag配置，让新配置生效
	if err := service.AIService().UpdateFlagFromAIRunConfig(); err != nil {
		amis.WriteJsonError(ctx, err)
		return
	}

	amis.WriteJsonOK(ctx)
}