package param

import (
	"fmt"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/flag"
)

// Version 获取版本号
func Version(c *gin.Context) {

	cfg := flag.Init()
	amis.WriteJsonData(c, gin.H{
		"version":   cfg.Version,
		"gitCommit": cfg.GitCommit,
		"gitTag":    cfg.GitTag,
		"gitRepo":   cfg.GitRepo,
		"buildDate": cfg.BuildDate,
		"goVersion": runtime.Version(),
		"compiler":  runtime.Compiler,
		"platform":  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	})
}
