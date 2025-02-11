package pod

import (
	"time"

	"github.com/gin-gonic/gin"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type LogQueryParams struct {
	Follow       bool   `form:"follow"`
	Previous     bool   `form:"previous"`
	SinceSeconds *int64 `form:"sinceSeconds"`
	SinceTime    string `form:"sinceTime"` // RFC3339 格式时间字符串
	Timestamps   bool   `form:"timestamps"`
	TailLines    *int64 `form:"tailLines"`
}

func BindPodLogOptions(c *gin.Context, containerName string) (*v1.PodLogOptions, error) {
	var params LogQueryParams

	// 绑定查询参数到自定义结构体
	if err := c.BindQuery(&params); err != nil {
		return nil, err
	}

	// 解析 sinceTime 字符串为 metav1.Time
	var sinceTime *metav1.Time
	if params.SinceTime != "" && params.SinceTime != "undefined" {
		parsedTime, err := time.Parse("2006-01-02 15:04:05", params.SinceTime)
		if err != nil {
			return nil, err
		}
		sinceTime = &metav1.Time{Time: parsedTime}
	}

	// 构造 v1.PodLogOptions 对象
	logOpt := &v1.PodLogOptions{
		Container:    containerName,
		Follow:       params.Follow,
		Previous:     params.Previous,
		SinceSeconds: params.SinceSeconds,
		SinceTime:    sinceTime,
		Timestamps:   params.Timestamps,
		TailLines:    params.TailLines,
	}
	return logOpt, nil
}
