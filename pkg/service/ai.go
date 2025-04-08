package service

import (
	"fmt"

	"github.com/weibaohui/k8m/pkg/ai"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/flag"
	"k8s.io/klog/v2"
)

type aiService struct {
	innerModel  string
	innerApiKey string
	innerApiUrl string
}

func (c *aiService) SetVars(apikey, apiUrl, model string) {
	c.innerModel = model
	c.innerApiUrl = apiUrl
	c.innerApiKey = apikey
}

func (c *aiService) DefaultClient() (ai.IAI, error) {
	apiKey, apiURL, model, enable := c.getChatGPTAuth()
	c.innerModel = model
	c.innerApiUrl = apiURL
	c.innerApiKey = apiKey
	if !enable {
		return nil, fmt.Errorf("ChatGPT功能未开启")
	}

	client, err := c.openAIClient()

	return client, err

}

func (c *aiService) openAIClient() (ai.IAI, error) {
	aiProvider := ai.Provider{
		Name:        "openai",
		Model:       c.innerModel,
		Password:    c.innerApiKey,
		BaseURL:     c.innerApiUrl,
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

	cfg := flag.Init()
	if !cfg.EnableAI {
		return
	}
	if cfg.ApiKey != "" {
		apiKey = cfg.ApiKey
	}
	if cfg.ApiURL != "" {
		apiURL = cfg.ApiURL
	}
	if cfg.ApiModel != "" {
		model = cfg.ApiModel
	}
	enable = cfg.EnableAI

	klog.V(4).Infof("ChatGPT 开关= %v\nurl=%s\nkey=%s\nmodel=%s\n", enable, apiURL, utils.MaskString(apiKey, 5), model)

	return
}
func (c *aiService) IsEnabled() bool {
	_, _, _, enable := c.getChatGPTAuth()
	klog.V(4).Infof("ChatGPT 状态:%v\n", enable)
	return enable
}
