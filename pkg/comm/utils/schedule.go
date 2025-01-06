package utils

import (
	"fmt"
	"time"
)

// 延迟启动DelayStartSchedule
// 设置一次性任务的执行时间，例如 5 秒后执行
// 返回 cron 表达式
func DelayStartSchedule(delaySeconds int) string {
	// 获取当前时间
	now := time.Now()

	// 计算延迟后的时间
	targetTime := now.Add(time.Duration(delaySeconds) * time.Second)

	// 提取目标时间的分钟、小时、天、月份和星期字段
	minute := targetTime.Minute()
	hour := targetTime.Hour()
	day := targetTime.Day()
	month := int(targetTime.Month())
	weekday := int(targetTime.Weekday())

	// 返回 Cron 表达式
	cronExpression := fmt.Sprintf("%d %d %d %d %d ", minute, hour, day, month, weekday)
	return cronExpression
}
