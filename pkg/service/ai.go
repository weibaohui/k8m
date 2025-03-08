package service

import (
	"fmt"
	"os"

	"github.com/weibaohui/k8m/pkg/ai"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"k8s.io/klog/v2"
)

type aiService struct {
	model  string
	apiKey string
	apiUrl string
}

func (c *aiService) SetVars(apikey, apiUrl, model string) {
	c.model = model
	c.apiUrl = apiUrl
	c.apiKey = apikey
}

func (c *aiService) DefaultClient() (ai.IAI, error) {
	apiKey, apiURL, model, enable := c.getChatGPTAuth()
	c.model = model
	c.apiUrl = apiURL
	c.apiKey = apiKey
	if !enable {
		return nil, fmt.Errorf("ChatGPT功能未开启")
	}

	client, err := c.openAIClient()

	return client, err

}

func (c *aiService) openAIClient() (ai.IAI, error) {
	aiProvider := ai.Provider{
		Name:        "openai",
		Model:       c.model,
		Password:    c.apiKey,
		BaseURL:     c.apiUrl,
		Temperature: 0.7,
		TopP:        1,
		TopK:        0,
		MaxTokens:   1000,
	}

	aiClient := ai.NewClient(aiProvider.Name)
	if err := aiClient.Configure(&aiProvider); err != nil {
		return nil, err
	}
	return aiClient, nil
}

func (c *aiService) getChatGPTAuth() (apiKey string, apiURL string, model string, enable bool) {
	// 从环境变量读取OpenAI API Key和API URL
	// 环境变量优先
	apiKey = os.Getenv("OPENAI_API_KEY")
	apiURL = os.Getenv("OPENAI_API_URL")
	model = os.Getenv("OPENAI_MODEL")
	if apiKey == "" && apiURL == "" {
		apiKey = c.apiKey
		apiURL = c.apiUrl
		klog.V(4).Infof("ChatGPT 环境变量没有设置 , 尝试使用默认配置 key:%s,url:%s\n", utils.MaskString(apiKey, 5), apiURL)
	} else {
		klog.V(4).Infof("ChatGPT 环境变量已设置, key:%s,url:%s\n", utils.MaskString(apiKey, 5), apiURL)
	}
	if apiKey != "" && apiURL != "" {
		enable = true
		klog.V(4).Infof("ChatGPT 启用 key:%s,url:%s\n", utils.MaskString(apiKey, 5), apiURL)
	}
	if model == "" {
		// 如果环境变量没设置，保底使用内置的
		model = c.model
		klog.V(4).Infof("ChatGPT 默认模型:%s\n", model)
	}

	if model != "" {
		// model 确实有值，且到这里应该为ENV环境变量的值
		// 那么model优先使用环境变量的值
		klog.V(4).Infof("ChatGPT 使用环境变量中设置的模型:%s\n", model)
		c.model = model
	}
	c.apiKey = apiKey
	c.apiUrl = apiURL
	return
}
func (c *aiService) IsEnabled() bool {
	_, _, _, enable := c.getChatGPTAuth()
	klog.V(4).Infof("ChatGPT 状态:%v\n", enable)
	return enable
}
