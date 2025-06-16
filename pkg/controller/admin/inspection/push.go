package inspection

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/webhooksender"
	"k8s.io/klog/v2"
)

func Push(c *gin.Context) {
	event := webhooksender.InspectionCheckEvent{
		ID:          1,
		RecordID:    2,
		EventStatus: "失败",
		EventMsg:    "关联Pod数为0",
		Extra:       "",
		ScriptName:  "检测Service Pod关联数",
		Kind:        "Service",
		CheckDesc:   "检测Service Pod关联数",
		Cluster:     "podman",
		Namespace:   "kube-system",
		Name:        "istio-gateway",
	}
	receiver := webhooksender.NewFeishuReceiver("https://open.feishu.cn/open-apis/bot/v2/hook/7e484f0c-0a0d-4b78-99fd-886f3e62bb8c", "JQMQdkqqMEEwoU96XG27qb")
	results := webhooksender.PushEvent(&event, []*webhooksender.WebhookReceiver{
		receiver,
	})
	for _, result := range results {
		klog.Infof("Push event: %v", result)
	}
	amis.WriteJsonOK(c)
}
