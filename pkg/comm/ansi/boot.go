package ansi

import (
	"fmt"
	"log"

	"github.com/weibaohui/k8m/pkg/comm/utils"
)

func ShowBootInfo(version string, port int) {

	// 获取本机所有 IP 地址
	ips, err := utils.GetLocalIPs()
	if err != nil {
		log.Fatalf("获取本机 IP 失败: %v", err)
	}

	// 打印 Vite 风格的启动信息
	fmt.Printf("%s k8m %s %s  启动成功\n", colorGreen, version, colorReset)
	fmt.Printf("%s➜%s  %sLocal:%s    %shttp://localhost:%d/%s\n", colorGreen, colorReset, colorBold, colorReset, colorPurple, port, colorReset)
	for _, ip := range ips {
		fmt.Printf("%s➜%s  %sNetwork:%s  %shttp://%s:%d/%s\n", colorGreen, colorReset, colorBold, colorReset, colorPurple, ip, port, colorReset)
	}

}
