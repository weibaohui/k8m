package server

import (
	"context"
	"fmt"

	mcp2 "github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	"k8s.io/klog/v2"
)

// ExecToolOverride 覆盖 kom 库 mcp/tools/pod 中同名工具。
//
// 背景：上游 kom v0.2.71 中 mcp/tools/pod/exec.go 的 ExecHandler 把命令输出写到
// `var execResult string` 并 `Execute(&execResult)`，但 kom/callbacks/exec.go 的 ExecuteCommand
// 反射检查要求 dest 是 *[]byte，否则直接返回
// "请确保dest 是一个指向字节切片的指针。定义var s []byte 使用&s"，
// 导致 MCP 客户端调用 run_command_in_k8s_pod 时必然失败（issue #468）。
//
// 我们在 k8m 启动 MCP server 后再次用同名工具 AddTool 覆盖（mcp-go AddTool 走 map 直接覆盖）。
//
// 该实现的输入参数和工具描述保持与上游一致，方便将来上游修复后平滑移除此文件。
func ExecToolOverride() mcp2.Tool {
	return mcp2.NewTool(
		"run_command_in_k8s_pod",
		mcp2.WithDescription("在Pod内执行命令，需指定容器名称 (类似命令: kubectl exec -n <namespace> <pod-name> -c <container-name> -- <command> [args...]) / Execute command in pod with container name"),
		mcp2.WithTitleAnnotation("Execute Command in Pod"),
		mcp2.WithDestructiveHintAnnotation(true),
		mcp2.WithString("cluster", mcp2.Description("集群名称 （使用空字符串表示默认集群）/ Cluster name")),
		mcp2.WithString("namespace", mcp2.Required(), mcp2.Description("命名空间 / Namespace")),
		mcp2.WithString("name", mcp2.Required(), mcp2.Description("Pod名称 / Pod name")),
		mcp2.WithString("container", mcp2.Description("容器名称（必填） / Container name (required)")),
		mcp2.WithString("command", mcp2.Required(), mcp2.Description("要执行的命令 / Command to execute")),
		mcp2.WithArray("args",
			mcp2.Description("命令参数列表 / Command arguments"),
			mcp2.Items(map[string]interface{}{"type": "string"}),
		),
	)
}

// ExecHandlerOverride 修复版 ExecHandler：使用 *[]byte 接收命令输出。
func ExecHandlerOverride(ctx context.Context, request mcp2.CallToolRequest) (*mcp2.CallToolResult, error) {
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	containerName := request.GetString("container", "")
	argsVal := request.GetStringSlice("args", []string{})
	command := request.GetString("command", "")

	klog.V(6).Infof("Executing command in pod %s/%s container %s: %v %v", meta.Namespace, meta.Name, containerName, command, argsVal)

	// 关键修复：dest 必须是 *[]byte，上游 kom 库 ExecHandler 错误地使用了 string。
	var execResult []byte
	err = kom.Cluster(meta.Cluster).WithContext(ctx).
		Namespace(meta.Namespace).
		Name(meta.Name).
		Ctl().Pod().
		ContainerName(containerName).
		Command(command, argsVal...).
		Execute(&execResult).Error

	if err != nil {
		return nil, fmt.Errorf("command execution failed: %v", err)
	}

	// TextResult[T] 对 []byte 类型会做 string(v) 转换，输出给 MCP 客户端是文本。
	return tools.TextResult(execResult, meta)
}