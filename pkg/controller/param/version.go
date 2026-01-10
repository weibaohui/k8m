package param

import (
	"fmt"
	"os"
	"runtime"

	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/flag"
	"github.com/weibaohui/k8m/pkg/response"
)

// Version 获取版本号
// @Summary 获取版本信息
// @Description 获取当前软件的版本及构建信息
// @Security BearerAuth
// @Success 200 {object} string
// @Router /params/version [get]
func (pc *Controller) Version(c *response.Context) {

	podName := os.Getenv("POD_NAME")
	namespace := os.Getenv("POD_NAMESPACE")
	podIP := os.Getenv("POD_IP")

	cfg := flag.Init()
	amis.WriteJsonData(c, response.H{
		"version":   cfg.Version,
		"gitCommit": cfg.GitCommit,
		"gitTag":    cfg.GitTag,
		"gitRepo":   cfg.GitRepo,
		"buildDate": cfg.BuildDate,
		"goVersion": runtime.Version(),
		"compiler":  runtime.Compiler,
		"platform":  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		"dbType":    cfg.DBDriver,
		"podName":   podName,
		"namespace": namespace,
		"podIP":     podIP,
	})
}
