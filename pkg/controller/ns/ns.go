package ns

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/kubectl"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
)

func TestDel(c *gin.Context) {
	item := v1.Deployment{}
	err := kubectl.Init().
		WithContext(c.Request.Context()).
		Resource(&item).
		Namespace("default").
		Name("ci-755702-codexxx").
		Get(&item)
	if err != nil {
		klog.Errorf("k8s.First(&item) error :%v", err)
	}
	fmt.Println(item.Spec.Template.Spec.Containers[0].Image)
	fmt.Println(item.Spec.Template.Spec.Containers[0].Image)
	fmt.Println(item.Spec.Template.Spec.Containers[0].Image)
	fmt.Println(item.Spec.Template.Spec.Containers[0].Image)
	fmt.Println(item.Spec.Template.Spec.Containers[0].Image)
	var crontab unstructured.Unstructured
	err = kubectl.Init().
		WithContext(c.Request.Context()).
		CRD("stable.example.com", "v1", "CronTab").
		Namespace("default").
		Name("my-new-cron-object").
		Unstructured().
		Fill(&crontab)
	if err != nil {
		fmt.Printf("Fill %v\n", err)
	}
	json := utils.ToJSON(crontab)
	fmt.Println(json)
}
func OptionList(c *gin.Context) {
	TestDel(c)

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
