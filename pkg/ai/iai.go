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
	"net/http"

	"github.com/sashabaranov/go-openai"
)

type IAI interface {
	Configure(config IAIConfig) error
	GetCompletion(ctx context.Context, contents ...any) (string, error)
	GetCompletionWithTools(ctx context.Context, contents ...any) ([]openai.ToolCall, string, error)
	GetStreamCompletion(ctx context.Context, contents ...any) (*openai.ChatCompletionStream, error)
	GetStreamCompletionWithTools(ctx context.Context, contents ...any) (*openai.ChatCompletionStream, error)
	GetName() string
	Close()
	SetTools(tools []openai.Tool)
	SaveAIHistory(ctx context.Context, content string)
	GetHistory(ctx context.Context) []openai.ChatCompletionMessage
	ClearHistory(ctx context.Context) error
}

type nopCloser struct{}

func (nopCloser) Close() {}

type IAIConfig interface {
	GetPassword() string
	GetModel() string
	GetBaseURL() string
	GetProxyEndpoint() string
	GetEndpointName() string
	GetEngine() string
	GetTemperature() float32
	GetProviderRegion() string
	GetTopP() float32
	GetTopK() int32
	GetMaxTokens() int
	GetMaxHistory() int32
	GetProviderId() string
	GetCompartmentId() string
	GetOrganizationId() string
	GetCustomHeaders() []http.Header
}

func NewClient(provider string) IAI {
	// default client
	return &OpenAIClient{}
}

type Configuration struct {
	Providers       []Provider
	DefaultProvider string
}

type Provider struct {
	Name           string
	Model          string
	Password       string
	BaseURL        string
	ProxyEndpoint  string
	ProxyPort      string
	EndpointName   string
	Engine         string
	Temperature    float32
	ProviderRegion string
	ProviderId     string
	CompartmentId  string
	TopP           float32
	TopK           int32
	MaxHistory     int32
	MaxTokens      int
	OrganizationId string
	CustomHeaders  []http.Header
}

func (p *Provider) GetBaseURL() string {
	return p.BaseURL
}

func (p *Provider) GetProxyEndpoint() string {
	return p.ProxyEndpoint
}

func (p *Provider) GetEndpointName() string {
	return p.EndpointName
}

func (p *Provider) GetTopP() float32 {
	return p.TopP
}

func (p *Provider) GetTopK() int32 {
	return p.TopK
}

func (p *Provider) GetMaxTokens() int {
	return p.MaxTokens
}
func (p *Provider) GetMaxHistory() int32 {
	return p.MaxHistory
}

func (p *Provider) GetPassword() string {
	return p.Password
}

func (p *Provider) GetModel() string {
	return p.Model
}

func (p *Provider) GetEngine() string {
	return p.Engine
}
func (p *Provider) GetTemperature() float32 {
	return p.Temperature
}

func (p *Provider) GetProviderRegion() string {
	return p.ProviderRegion
}

func (p *Provider) GetProviderId() string {
	return p.ProviderId
}

func (p *Provider) GetCompartmentId() string {
	return p.CompartmentId
}

func (p *Provider) GetOrganizationId() string {
	return p.OrganizationId
}

func (p *Provider) GetCustomHeaders() []http.Header {
	return p.CustomHeaders
}

var passwordlessProviders = []string{"localai", "ollama", "amazonsagemaker", "amazonbedrock", "googlevertexai", "oci"}

func NeedPassword(backend string) bool {
	for _, b := range passwordlessProviders {
		if b == backend {
			return false
		}
	}
	return true
}
