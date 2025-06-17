package inspection

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/webhooksender"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

func Push(c *gin.Context) {
	recordIDStr := c.Param("id")
	if recordIDStr == "" {
		amis.WriteJsonError(c, fmt.Errorf("缺少 record_id 参数"))
		return
	}
	recordID := utils.ToUInt(recordIDStr)
	if recordID == 0 {
		amis.WriteJsonError(c, fmt.Errorf("record_id 参数无效"))
		return
	}

	// 1. 查询 InspectionRecord
	record := &models.InspectionRecord{}
	record, err := record.GetOne(nil, func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", recordID)
	})
	if err != nil {
		amis.WriteJsonError(c, fmt.Errorf("未找到对应的巡检记录: %v", err))
		return
	}

	// 查询webhooks
	hookModel := &models.WebhookReceiver{}
	hooks, _, err := hookModel.List(dao.BuildDefaultParams())
	if err != nil {
		amis.WriteJsonError(c, fmt.Errorf("查询webhooks失败: %v", err))
		return
	}
	var results []webhooksender.SendResult
	for _, hook := range hooks {
		if hook.Platform == "feishu" {
			receiver := webhooksender.NewFeishuReceiver(hook.TargetURL, hook.SignSecret)
			ret := webhooksender.PushEvent(record.AISummary, []*webhooksender.WebhookReceiver{
				receiver,
			})
			results = append(results, ret...)
		}
	}

	for _, result := range results {
		klog.Infof("Push event: %v", result)
	}
	amis.WriteJsonOK(c)
}
