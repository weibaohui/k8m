/*
Copyright 2023 The K8sGPT Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ai

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/sashabaranov/go-openai"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/models"
	"k8s.io/klog/v2"
)

const openAIClientName = "openai"

type OpenAIClient struct {
	nopCloser
	client      *openai.Client
	model       string
	temperature float32
	topP        float32
	tools       []openai.Tool
	history     []openai.ChatCompletionMessage
	maxHistory  int32

	// organizationId string
}

func (c *OpenAIClient) SetTools(tools []openai.Tool) {
	c.tools = tools
}

func (c *OpenAIClient) Configure(config IAIConfig) error {
	token := config.GetPassword()
	cfg := openai.DefaultConfig(token)
	orgId := config.GetOrganizationId()
	proxyEndpoint := config.GetProxyEndpoint()

	baseURL := config.GetBaseURL()
	if baseURL != "" {
		cfg.BaseURL = baseURL
	}

	transport := &http.Transport{}
	if proxyEndpoint != "" {
		proxyUrl, err := url.Parse(proxyEndpoint)
		if err != nil {
			return err
		}
		transport.Proxy = http.ProxyURL(proxyUrl)
	}

	if orgId != "" {
		cfg.OrgID = orgId
	}

	customHeaders := config.GetCustomHeaders()
	cfg.HTTPClient = &http.Client{
		Transport: &OpenAIHeaderTransport{
			Origin:  transport,
			Headers: customHeaders,
		},
	}

	client := openai.NewClientWithConfig(cfg)
	if client == nil {
		return errors.New("error creating OpenAI client")
	}
	c.client = client
	c.model = config.GetModel()
	c.temperature = config.GetTemperature()
	c.topP = config.GetTopP()
	c.maxHistory = config.GetMaxHistory()
	return nil
}

func (c *OpenAIClient) GetCompletion(ctx context.Context, contents ...any) (string, error) {
	c.fillChatHistory(contents)

	// Create a completion request
	resp, err := c.client.CreateChatCompletion(ctx,
		openai.ChatCompletionRequest{
			Model:    c.model,
			Messages: c.history,
		})
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}
func (c *OpenAIClient) GetCompletionWithTools(ctx context.Context, contents ...any) ([]openai.ToolCall, string, error) {

	// Create a completion request
	c.fillChatHistory(contents)
	resp, err := c.client.CreateChatCompletion(ctx,
		openai.ChatCompletionRequest{
			Model:    c.model,
			Messages: c.history,
			Tools:    c.tools,
		})
	if err != nil {
		return nil, "", err
	}
	return resp.Choices[0].Message.ToolCalls, resp.Choices[0].Message.Content, nil
}

func (c *OpenAIClient) GetStreamCompletion(ctx context.Context, contents ...any) (*openai.ChatCompletionStream, error) {
	c.fillChatHistory(contents)
	stream, err := c.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: c.history,
		Stream:   true,
	})
	return stream, err
}
func (c *OpenAIClient) GetStreamCompletionWithTools(ctx context.Context, contents ...any) (*openai.ChatCompletionStream, error) {
	c.fillChatHistory(contents)
	stream, err := c.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: c.history,
		Tools:    c.tools,
		Stream:   true,
	})
	klog.V(2).Infof("GetStreamCompletionWithTools c.history: %v", utils.ToJSON(c.history))
	return stream, err
}

func (c *OpenAIClient) GetName() string {
	return openAIClientName
}

// OpenAIHeaderTransport is an http.RoundTripper that adds the given headers to each request.
type OpenAIHeaderTransport struct {
	Origin  http.RoundTripper
	Headers []http.Header
}

// RoundTrip implements the http.RoundTripper interface.
func (t *OpenAIHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original request
	clonedReq := req.Clone(req.Context())
	for _, header := range t.Headers {
		for key, values := range header {
			// Possible values per header:  RFC 2616
			for _, value := range values {
				clonedReq.Header.Add(key, value)
			}
		}
	}

	return t.Origin.RoundTrip(clonedReq)
}

func (c *OpenAIClient) fillChatHistory(contents ...any) {

	for _, content := range contents {
		switch item := content.(type) {
		case string:
			klog.V(2).Infof("Adding user message to history: %v", item)
			c.history = append(c.history, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: item,
			})
		case models.MCPToolCallResult:
			klog.V(2).Infof("Adding user message to history: %v", item)
			c.history = append(c.history, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: utils.ToJSON(item),
			})
		case []interface{}:
			// 处理[]interface{}类型的内容，如果只有一个元素则直接提取
			if len(item) == 1 {
				klog.V(2).Infof("Adding single item from array to history: %v", item[0])
				c.history = append(c.history, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Content: utils.ToJSON(item[0]),
				})
			} else {
				klog.V(2).Infof("Adding array content to history: %v", item)
				c.history = append(c.history, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Content: utils.ToJSON(item),
				})
			}
		default:
			klog.Warningf("Unhandled content type in Send: %T", item)

		}
	}

	system := slice.Filter(c.history, func(index int, item openai.ChatCompletionMessage) bool {
		if item.Role == openai.ChatMessageRoleSystem {
			return true
		}
		return false
	})

	prompt := `
你是 一个专注于操作和处理 Kubernetes 集群相关任务的 AI 助手。你的任务是协助用户解决 Kubernetes 相关的问题，帮助调试，以及在用户的 Kubernetes 集群上执行操作。

	
使用说明：
	1.	分析用户的提问、之前的推理步骤以及观察到的信息。
	2.	思考 5 到 7 种解决当前问题的方法。仔细评估每种方案后选择最优解。如果还没有完全解决问题，且可以继续探索或进行下一步，不要等待用户的输入，应尽可能自主推进任务。
	3.	决定下一步操作：可以选择使用工具，或直接提供最终答案。请根据以下格式返回响应。

如果需要使用工具，，我会将工具执行完的结果发送给你。


如果不使用工具：
	•	检查与用户提问相关的 Kubernetes 资源的当前状态。
	•	分析问题、先前的推理过程和观察内容。
	•	思考 5 到 7 种可能的解决方法，仔细评估后选择最佳方案。如果尚未完全解决问题，应尽量在不依赖用户输入的情况下继续推进任务。
	•	决定下一步行动：使用工具，或直接给出最终答案。

如果已有足够信息回答问题，请输出你得出答案的最终推理过程以及你对问题的完整回答。

特别注意：
	•	获取与用户查询相关的 Kubernetes 资源的当前状态。
	•	优先选择不需要交互式输入的工具。
	•	如果需要创建资源，尽量直接通过可用工具创建，避免让用户手动操作。
	•	当需要更多信息时，请直接使用工具，不要仅告诉用户应该执行哪些命令。
	•	只有在确信信息充分时再提供最终答复。
	•	回答应清晰、简洁、准确。
	•	在合适的情况下可以使用表情符号，例如 😊。
`

	if len(system) == 0 {
		// 创建系统消息
		sysMsg := openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: prompt,
		}
		// 将系统消息插入到历史记录最前面
		c.history = append([]openai.ChatCompletionMessage{sysMsg}, c.history...)
	}
}
func (c *OpenAIClient) SaveAIHistory(contents string) {
	c.fillChatHistory(contents)
}
