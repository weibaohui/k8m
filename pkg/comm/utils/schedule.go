package utils

import (
	"fmt"
	"time"
)

// 延迟启动DelayStartSchedule
// 设置一次性任务的执行时间，例如 5 秒后执行
// 返回 cron 表达式
func DelayStartSchedule(seconds int) string {
	// 使用 cron 表达式，精确到秒需要扩展格式
	fromNow := time.Now().Add(time.Duration(seconds) * time.Second)
	schedule := fmt.Sprintf("%d %d %d %d %d",
		fromNow.Second(),
		fromNow.Minute(),
		fromNow.Hour(),
		fromNow.Day(),
		fromNow.Month())
	return schedule
}
