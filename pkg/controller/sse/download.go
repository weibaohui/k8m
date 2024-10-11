package sse

import (
	"fmt"
	"io"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/utils/amis"
	v1 "k8s.io/api/core/v1"
)

func DownloadLog(c *gin.Context, opt *v1.PodLogOptions, stream io.ReadCloser) {
	defer func() {
		if err := stream.Close(); err != nil {
			// 处理关闭流时的错误
			log.Printf("stream close error:%v", err)
		}
	}()

	name := fmt.Sprintf("%s.log", opt.Container)
	// 设置响应头信息，指定文件下载
	c.Writer.Header().Set("Content-Disposition", "attachment; filename="+name)
	c.Writer.Header().Set("Content-Type", "text/plain")

	// 将日志直接写入响应流
	_, err := io.Copy(c.Writer, stream)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
}
