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

var local ai.IAI

func (c *aiService) DefaultClient() (ai.IAI, error) {
	enable := c.IsEnabled()
	if !enable {
		return nil, fmt.Errorf("ChatGPT功能未开启")
	}

	if local != nil {
		return local, nil
	}

	if client, err := c.openAIClient(); err == nil {
		local = client
	}
	return local, nil

}

func (c *aiService) openAIClient() (ai.IAI, error) {
	cfg := flag.Init()

	aiProvider := ai.AIProvider{
		Name:        "openai",
		Model:       cfg.ApiModel,
		Password:    cfg.ApiKey,
		BaseURL:     cfg.ApiURL,
		Temperature: 0.7,
		TopP:        1,
		TopK:        0,
		MaxTokens:   1000,
		MaxHistory:  10,
	}
	if cfg.EnableAI && cfg.UseBuiltInModel {
		aiProvider.BaseURL = c.innerApiUrl
		aiProvider.Password = c.innerApiKey
		aiProvider.Model = c.innerModel
	}

	if cfg.Debug {
		klog.V(4).Infof("ai BaseURL: %v\n", aiProvider.BaseURL)
		klog.V(4).Infof("ai Model : %v\n", aiProvider.Model)
		klog.V(4).Infof("ai Key: %v\n", utils.MaskString(aiProvider.Password, 5))
	}

	aiClient := ai.NewAIClient(aiProvider.Name)
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
