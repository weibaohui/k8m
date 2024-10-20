package ns

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/kubectl"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func OptionList(c *gin.Context) {
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "nginx-jpswn",
		},
	}
	err := kubectl.Init().SQLGet(&pod)
	if err != nil {
		klog.Errorf("k8s.First(&pod) error :%v", err)
	}
	fmt.Println(pod.Spec.Containers[0].Image)
	fmt.Println(pod.Spec.Containers[0].Image)
	fmt.Println(pod.Spec.Containers[0].Image)
	fmt.Println(pod.Spec.Containers[0].Image)
	fmt.Println(pod.Spec.Containers[0].Image)
	fmt.Println(pod.Spec.Containers[0].Image)

	ctx := c.Request.Context()
	namespace, err := kubectl.Init().ListNamespace(ctx)
	if err != nil {
		amis.WriteJsonError(c, err)
	}
	var list []map[string]string
	for _, ns := range namespace {
		list = append(list, map[string]string{
			"label": ns.Name,
			"value": ns.Name,
		})
	}
	amis.WriteJsonData(c, gin.H{
		"options": list,
	})
}
