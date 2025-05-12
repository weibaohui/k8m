package ai

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/sashabaranov/go-openai"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/constants"
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
	maxHistory  int32
	memory      *memoryService

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
	c.memory = NewMemoryService()
	return nil
}

func (c *OpenAIClient) GetCompletion(ctx context.Context, contents ...any) (string, error) {
	c.fillChatHistory(ctx, contents)

	// Create a completion request
	resp, err := c.client.CreateChatCompletion(ctx,
		openai.ChatCompletionRequest{
			Model:    c.model,
			Messages: c.GetHistory(ctx),
		})
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}
func (c *OpenAIClient) GetCompletionWithTools(ctx context.Context, contents ...any) ([]openai.ToolCall, string, error) {

	// Create a completion request
	c.fillChatHistory(ctx, contents)
	resp, err := c.client.CreateChatCompletion(ctx,
		openai.ChatCompletionRequest{
			Model:       c.model,
			Messages:    c.GetHistory(ctx),
			Temperature: c.temperature,
			TopP:        c.topP,
			Tools:       c.tools,
		})
	if err != nil {
		return nil, "", err
	}
	return resp.Choices[0].Message.ToolCalls, resp.Choices[0].Message.Content, nil
}

func (c *OpenAIClient) GetStreamCompletion(ctx context.Context, contents ...any) (*openai.ChatCompletionStream, error) {
	c.fillChatHistory(ctx, contents)
	stream, err := c.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:       c.model,
		Messages:    c.GetHistory(ctx),
		Temperature: c.temperature,
		TopP:        c.topP,
		Stream:      true,
	})
	return stream, err
}
func (c *OpenAIClient) GetStreamCompletionWithTools(ctx context.Context, contents ...any) (*openai.ChatCompletionStream, error) {
	c.fillChatHistory(ctx, contents)
	stream, err := c.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: c.GetHistory(ctx),
		Tools:    c.tools,
		Stream:   true,
	})
	klog.V(6).Infof("GetStreamCompletionWithTools history length: %d", len(c.GetHistory(ctx)))
	klog.V(8).Infof("GetStreamCompletionWithTools c.history: %v", utils.ToJSON(c.GetHistory(ctx)))
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

func (c *OpenAIClient) SaveAIHistory(ctx context.Context, contents string) {
	val := ctx.Value(constants.JwtUserName)
	if username, ok := val.(string); ok {
		c.memory.AppendUserHistory(username, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: contents,
		})
	} else {
		klog.Warningf("SaveAIHistory content but user not found: %s", contents)
	}
}

func (c *OpenAIClient) GetHistory(ctx context.Context) []openai.ChatCompletionMessage {
	val := ctx.Value(constants.JwtUserName)
	if username, ok := val.(string); ok {
		return c.memory.GetUserHistory(username)
	}
	return make([]openai.ChatCompletionMessage, 0)

}

func (c *OpenAIClient) fillChatHistory(ctx context.Context, contents ...any) {

	history := c.GetHistory(ctx)
	for _, content := range contents {
		switch item := content.(type) {
		case string:
			klog.V(2).Infof("Adding user message to history: %v", item)
			history = append(history, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: item,
			})
		case models.MCPToolCallResult:
			klog.V(2).Infof("Adding user message to history: %v", item)
			history = append(history, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: utils.ToJSON(item),
			})
		case []string:
			klog.V(2).Infof("Adding string array to history: %v", item)
			for _, m := range item {
				history = append(history, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Content: m,
				})
			}
		case []models.MCPToolCallResult:
			klog.V(2).Infof("Adding MCPToolCallResult array to history: %v", item)
			for _, m := range item {
				history = append(history, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Content: utils.ToJSON(m),
				})
			}
		case []interface{}:
			for _, m := range item {
				history = append(history, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Content: utils.ToJSON(m),
				})
			}
		default:
			klog.Warningf("Unhandled content type in Send: %T", item)
		}
	}

	// 保留最后 maxHistory 条（含系统提示）
	if c.maxHistory > 0 && int32(len(history)) > c.maxHistory {
		keep := history[len(history)-int(c.maxHistory):]
		history = keep
	}

	system := slice.Filter(history, func(index int, item openai.ChatCompletionMessage) bool {
		if item.Role == openai.ChatMessageRoleSystem {
			return true
		}
		return false
	})

	if len(system) == 0 {
		// 创建系统消息
		sysMsg := openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: sysPrompt,
		}
		// 将系统消息插入到历史记录最前面
		history = append([]openai.ChatCompletionMessage{sysMsg}, history...)
	}
	username := ctx.Value(constants.JwtUserName).(string)
	c.memory.SetUserHistory(username, history)

}
