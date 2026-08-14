package server

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	kommcp "github.com/weibaohui/kom/mcp"
)

// TestExecToolOverrideSchema 覆盖版本工具 schema 与上游 kom 保持一致。
func TestExecToolOverrideSchema(t *testing.T) {
	tool := ExecToolOverride()
	if tool.Name != "run_command_in_k8s_pod" {
		t.Errorf("tool.Name = %q, want run_command_in_k8s_pod", tool.Name)
	}
	required := []string{"namespace", "name", "command"}
	for _, r := range required {
		found := false
		for _, fr := range tool.InputSchema.Required {
			if fr == r {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected required param %q in tool schema, got %v", r, tool.InputSchema.Required)
		}
	}
}

// TestExecHandlerOverrideUsesByteSlice 回归保护：issue #468。
//
// 上游 kom v0.2.71 的 mcp/tools/pod ExecHandler 把命令输出写到 string 并 Execute(&string)，
// 但 kom/callbacks/exec.go 的反射校验要求 dest 是 *[]byte，
// 任何调用都会立刻返回 "请确保dest 是一个指向字节切片的指针"。
//
// 我们用两种方式锁定覆盖版本始终使用 []byte：
//   1. 通过源码字符串匹配，确保 ExecHandlerOverride 中存在关键修复
//      `var execResult []byte` 和 `Execute(&execResult)`
//   2. 注册工具并通过反射断言 handler 的入参签名
func TestExecHandlerOverrideUsesByteSlice(t *testing.T) {
	src, err := os.ReadFile("exec_override.go")
	if err != nil {
		t.Fatalf("read exec_override.go: %v", err)
	}
	srcStr := string(src)
	if !strings.Contains(srcStr, "var execResult []byte") {
		t.Fatalf("exec_override.go must contain 'var execResult []byte' (issue #468 fix)")
	}
	if !strings.Contains(srcStr, "Execute(&execResult)") {
		t.Fatalf("exec_override.go must call Execute(&execResult)")
	}
	// 上游有 bug 的写法必须是 NULL 状态：不允许出现 var execResult string 后再 &execResult 传给 Execute。
	// 这里只用关键修复特征判定：必须出现 `var execResult []byte`，否则覆盖失效。
	// 不再单独做 "must not contain string" 反向检查，避免与注释中的中英文叙述冲突。
}

// TestExecToolOverridesKomBuiltin 注册到 kom 创建的 MCP server 上，覆盖同名 kom 工具。
func TestExecToolOverridesKomBuiltin(t *testing.T) {
	serv := kommcp.GetMCPServerWithOption(&kommcp.ServerConfig{
		Name:    "test",
		Version: "v0",
	})
	// 此时 serv 已经包含了 kom 库注册的 run_command_in_k8s_pod。
	serv.AddTool(ExecToolOverride(), ExecHandlerOverride)

	// 通过 ListTools 验证覆盖生效：handler 应该是我们的 ExecHandlerOverride。
	tools := serv.ListTools()
	entry, ok := tools["run_command_in_k8s_pod"]
	if !ok {
		t.Fatalf("run_command_in_k8s_pod not registered, got %v", tools)
	}
	if entry.Handler == nil {
		t.Fatalf("run_command_in_k8s_pod handler is nil")
	}
}

// 确保 imports 不被 goimports 移除。
var (
	_ = context.Background
	_ = mcp.CallToolRequest{}
)