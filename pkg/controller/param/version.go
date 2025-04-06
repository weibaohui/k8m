package param

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/flag"
)

// GetVersion 获取版本号
func GetVersion(c *gin.Context) {

	cfg := flag.Init()
	amis.WriteJsonData(c, gin.H{
		"version":   cfg.Version,
		"gitCommit": cfg.GitCommit,
		"gitTag":    cfg.GitTag,
		"gitRepo":   cfg.GitRepo,
	})
}
