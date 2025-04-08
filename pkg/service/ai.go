package service

import (
	"fmt"

	"github.com/weibaohui/k8m/pkg/ai"
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
	enable := c.IsEnabled()
	if !enable {
		return nil, fmt.Errorf("ChatGPT功能未开启")
	}

	client, err := c.openAIClient()

	return client, err

}

func (c *aiService) openAIClient() (ai.IAI, error) {
	cfg := flag.Init()

	aiProvider := ai.Provider{
		Name:        "openai",
		Model:       cfg.ApiModel,
		Password:    cfg.ApiKey,
		BaseURL:     cfg.ApiURL,
		Temperature: 0.7,
		TopP:        1,
		TopK:        0,
		MaxTokens:   1000,
	}
	if cfg.EnableAI && cfg.UseBuiltInModel {
		aiProvider.BaseURL = c.innerApiUrl
		aiProvider.Password = c.innerApiKey
		aiProvider.Model = c.innerModel
	}

	aiClient := ai.NewClient(aiProvider.Name)
	if err := aiClient.Configure(&aiProvider); err != nil {
		return nil, err
	}
	return aiClient, nil
}

func (c *aiService) IsEnabled() bool {
	cfg := flag.Init()
	enable := cfg.EnableAI
	klog.V(4).Infof("ChatGPT 状态:%v\n", enable)
	return enable
}
