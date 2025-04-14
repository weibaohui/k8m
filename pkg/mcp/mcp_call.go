package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sashabaranov/go-openai"
	"github.com/weibaohui/k8m/pkg/ai"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"k8s.io/klog/v2"
)

// ToolCallResult 存储工具调用的结果
type ToolCallResult struct {
	ToolName   string                 `json:"tool_name"`
	Parameters map[string]interface{} `json:"parameters"`
	Result     string                 `json:"result"`
	Error      string                 `json:"error,omitempty"`
}

func (m *MCPHost) ProcessWithOpenAI(ctx context.Context, ai ai.IAI, prompt string) (string, []ToolCallResult, error) {

	// 创建带有工具的聊天完成请求
	tools := m.GetAllTools(ctx)
	ai.SetTools(tools)
	toolCalls, content, err := ai.GetCompletionWithTools(ctx, prompt)
	if err != nil {
		return "", nil, err
	}

	results := m.ExecTools(ctx, toolCalls)

	return content, results, nil

}

func (m *MCPHost) ExecTools(ctx context.Context, toolCalls []openai.ToolCall) []ToolCallResult {
	// 存储所有工具调用的结果
	var results []ToolCallResult

	// 处理工具调用
	if toolCalls != nil {
		for _, toolCall := range toolCalls {
			startTime := time.Now()

			fullToolName := toolCall.Function.Name
			klog.V(6).Infof("Tool Name: %s\n", fullToolName)
			arguments := toolCall.Function.Arguments
			arguments = clean(arguments)
			klog.V(6).Infof("Tool Arguments: %s\n", arguments)

			result := ToolCallResult{
				ToolName: fullToolName,
			}

			// 解析参数
			var args map[string]interface{}
			if arguments != "" && arguments != "{}" && arguments != "null" {

				if err := json.Unmarshal([]byte(arguments), &args); err != nil {
					result.Error = fmt.Sprintf("failed to parse tool arguments: %v", err)
					klog.V(6).Infof("参数解析Error: %s\n", result.Error)
					results = append(results, result)
					continue
				}
			}

			result.Parameters = args

			var cli *client.SSEMCPClient
			var toolName, serverName string
			var err error
			if strings.Contains(fullToolName, "@") {
				// 如果识别的ToolName包含@，则解析ToolName
				toolName, serverName, _ = utils.ParseMCPToolName(fullToolName)
			} else {
				toolName = fullToolName
				serverName = m.GetServerNameByToolName(toolName)
			}
			klog.V(6).Infof("解析ToolName: %s, ServerName: %s\n", toolName, serverName)
			if serverName == "" {
				// 解析失败，尝试直接用toolName
				result.Error = fmt.Sprintf("根据Tool名称 %s 解析MCP Server 名称失败: %v", fullToolName, err)
				results = append(results, result)
				continue
			}
			klog.V(6).Infof("解析ToolName: %s, ServerName: %s\n", toolName, serverName)
			// 执行工具调用
			callRequest := mcp.CallToolRequest{}
			callRequest.Params.Name = toolName
			callRequest.Params.Arguments = args
			klog.V(6).Infof("执行工具调用: %s\n", utils.ToJSON(callRequest))
			cli, err = m.GetClient(ctx, serverName)
			if err != nil {
				klog.V(6).Infof("获取MCP Client 失败: %v\n", err)
				result.Error = fmt.Sprintf("获取MCP Client 失败: %v", err)
				results = append(results, result)
				continue
			}
			// 执行工具
			callResult, err := cli.CallTool(ctx, callRequest)
			// 记录执行日志
			executeTime := time.Since(startTime).Milliseconds()
			if err != nil {
				klog.V(6).Infof("工具执行失败: %v\n", err)
				result.Error = fmt.Sprintf("工具执行失败: %v", err)
				results = append(results, result)
				m.LogToolExecution(ctx, toolName, serverName, args, result, executeTime)
				continue
			}

			// 处理工具执行结果
			if len(callResult.Content) > 0 {
				if textContent, ok := callResult.Content[0].(mcp.TextContent); ok {
					result.Result = textContent.Text
				}
			}
			results = append(results, result)

			m.LogToolExecution(ctx, toolName, serverName, args, result, executeTime)
		}
	}
	return results
}

func clean(arguments string) string {
	arguments = strings.TrimSpace(arguments)
	arguments = strings.ReplaceAll(arguments, "\n", "")
	arguments = strings.ReplaceAll(arguments, "\t", "")
	if arguments == "{}}" {
		arguments = ""
	}
	if arguments == "{}" {
		arguments = ""
	}
	return arguments
}
