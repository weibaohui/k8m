package service

import (
	"fmt"
	"sync"

	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/flag"
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/ai/core"
	"github.com/weibaohui/k8m/pkg/plugins/modules/ai/models"
	"k8s.io/klog/v2"
)

type aiService struct {
	innerModel  string
	innerApiKey string
	innerApiUrl string
	// AI配置参数
	UseBuiltInModel bool    // 是否使用内置大模型参数
	AnySelect       bool    // 是否开启任意选择
	MaxHistory      int32   // 模型对话上下文历史记录数
	MaxIterations   int32   // 模型自动对话的最大轮数
	ApiKey          string  // 大模型的自定义API Key
	ApiModel        string  // 大模型的自定义模型名称
	ApiURL          string  // 大模型的自定义API URL
	Think           bool    // AI是否开启思考过程输出
	Temperature     float32 // 模型温度
	TopP            float32 // 模型topP参数
}

var (
	// aiInstance 单例实例
	aiInstance *aiService
	// aiOnce 用于确保单例只被初始化一次
	aiOnce sync.Once

	local core.IAI
)

// AIService 获取AI服务的单例实例
// 返回值:
//   - *aiService: AI服务实例
func AIService() *aiService {
	aiOnce.Do(func() {
		aiInstance = &aiService{}
	})
	return aiInstance
}

func (c *aiService) SetVars(apikey, apiUrl, model string) {
	c.innerModel = model
	c.innerApiUrl = apiUrl
	c.innerApiKey = apikey
}

func (c *aiService) DefaultClient() (core.IAI, error) {
	enable := c.IsEnabled()
	if !enable {
		return nil, fmt.Errorf("AI 功能未开启")
	}

	if local != nil {
		return local, nil
	}

	if client, err := c.openAIClient(); err == nil {
		local = client
	} else {
		return nil, err
	}
	return local, nil
}

// ResetDefaultClient 重置 local ，适用于切换
func (c *aiService) ResetDefaultClient() error {
	enable := c.IsEnabled()
	if !enable {
		return fmt.Errorf("AI功能未开启")
	}
	local = nil
	klog.V(6).Infof("AI DefaultClient Reset ")
	return nil
}

func (c *aiService) openAIClient() (core.IAI, error) {
	aiProvider := core.Provider{
		Name:        "openai",
		Model:       c.ApiModel,
		Password:    c.ApiKey,
		BaseURL:     c.ApiURL,
		Temperature: 0.7,
		TopP:        1,
		MaxHistory:  10,
		TopK:        0,
		MaxTokens:   1000,
		Think:       c.Think,
	}
	if c.UseBuiltInModel {
		aiProvider.BaseURL = c.innerApiUrl
		aiProvider.Password = c.innerApiKey
		aiProvider.Model = c.innerModel
	}

	if c.Temperature > 0 {
		aiProvider.Temperature = c.Temperature
	}
	if c.TopP > 0 {
		aiProvider.TopP = c.TopP
	}
	if c.MaxHistory > 0 {
		aiProvider.MaxHistory = c.MaxHistory
	}

	// 检查全局调试模式
	if flag.Init().Debug {
		klog.V(4).Infof("ai BaseURL: %v\n", aiProvider.BaseURL)
		klog.V(4).Infof("ai Model : %v\n", aiProvider.Model)
		klog.V(4).Infof("ai Key: %v\n", utils.MaskString(aiProvider.Password, 5))
	}

	aiClient := core.NewClient(aiProvider.Name)
	if err := aiClient.Configure(&aiProvider); err != nil {
		return nil, err
	}
	return aiClient, nil
}

func (c *aiService) IsEnabled() bool {
	enable := plugins.ManagerInstance().IsEnabled(modules.PluginNameAI)
	klog.V(4).Infof("ChatGPT 状态:%v\n", enable)
	return enable
}

func (c *aiService) TestClient(url string, key string, model string) (core.IAI, error) {
	klog.V(6).Infof("TestClient url:%v key:%v model:%v\n", url, utils.MaskString(key, 5), model)
	aiProvider := core.Provider{
		Name:     "test",
		Model:    model,
		Password: key,
		BaseURL:  url,
	}

	aiClient := core.NewClient(aiProvider.Name)
	if err := aiClient.Configure(&aiProvider); err != nil {
		return nil, err
	}
	return aiClient, nil
}

// UpdateFlagFromAIRunConfig 从AI运行配置表更新AI服务配置
func (c *aiService) UpdateFlagFromAIRunConfig() error {
	// 获取AI运行配置
	runConfig, err := AIRunConfigService().GetDefault()
	if err != nil {
		klog.Errorf("UpdateFlagFromAIRunConfig 获取AI运行配置失败: %v", err)
		return err
	}

	// 更新标志配置
	c.UseBuiltInModel = runConfig.UseBuiltInModel
	c.AnySelect = runConfig.AnySelect
	c.MaxHistory = runConfig.MaxHistory
	c.MaxIterations = runConfig.MaxIterations

	// 如果不使用内置模型，加载模型配置
	if !runConfig.UseBuiltInModel {
		if runConfig.ModelID == 0 {
			klog.Errorf("UpdateFlagFromAIRunConfig 未指定有效的模型ID")
			return fmt.Errorf("未指定有效的模型ID")
		}

		modelConfig := &models.AIModelConfig{ID: runConfig.ModelID}
		modelConfig, err := modelConfig.GetOne(nil)
		if err != nil {
			klog.Errorf("UpdateFlagFromAIRunConfig 获取模型配置失败: %v", err)
			return err
		}

		c.ApiKey = modelConfig.ApiKey
		c.ApiModel = modelConfig.ApiModel
		c.ApiURL = modelConfig.ApiURL
		c.Think = modelConfig.Think
		if modelConfig.Temperature > 0 {
			c.Temperature = modelConfig.Temperature
		}
		if modelConfig.TopP > 0 {
			c.TopP = modelConfig.TopP
		}
	}

	// 重置默认客户端，使新配置生效
	return c.ResetDefaultClient()
}
