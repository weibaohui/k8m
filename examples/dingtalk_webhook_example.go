package main

import (
	"fmt"
	"log"

	"github.com/weibaohui/k8m/pkg/webhook"
)

func main() {
	// 创建钉钉接收器
	receiver := webhook.NewDingtalkReceiver(
		"https://oapi.dingtalk.com/robot/send?access_token=cd98b71753e1f5d227b43a8e8ff6c8cfbd9f5f06a333aa4391764fa2a2d00acc",
		"SEC9c105a1ecc49e341fac48db0bb3f462c8fbce55cfca899a0944c3d6a164dea89",
	)

	// 验证接收器配置
	if err := receiver.Validate(); err != nil {
		log.Fatalf("验证接收器配置失败: %v", err)
	}

	// 获取钉钉发送器
	sender, err := webhook.GetSender("dingtalk")
	if err != nil {
		log.Fatalf("获取钉钉发送器失败: %v", err)
	}

	// 发送消息
	result, err := sender.Send("这是一条测试消息", receiver)
	if err != nil {
		log.Fatalf("发送消息失败: %v", err)
	}

	fmt.Printf("发送结果: Status=%s, StatusCode=%d\n", result.Status, result.StatusCode)
	if result.RespBody != "" {
		fmt.Printf("响应内容: %s\n", result.RespBody)
	}
}
