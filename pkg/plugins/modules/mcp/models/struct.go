package models

// MCPToolCallResult 存储工具调用的结果
type MCPToolCallResult struct {
	ToolName   string `json:"tool_name"`
	Parameters any    `json:"parameters"`
	Result     string `json:"result"`
	Error      string `json:"error,omitempty"`
}
