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
		Get(&item).Error
	if err != nil {
		klog.Errorf("k8s.First(&item) error :%v", err)
	}
	fmt.Printf("Get Item %s\n", item.Spec.Template.Spec.Containers[0].Image)
	var crontab unstructured.Unstructured
	err = kubectl.Init().
		WithContext(c.Request.Context()).
		CRD("stable.example.com", "v1", "CronTab").
		Namespace("default").
		Name("my-new-cron-object").
		Get(&crontab).Error
	if err != nil {
		fmt.Printf("Fill %v\n", err)
	}
	json := utils.ToJSON(crontab)
	fmt.Printf("crontab json %s\n", json)

	var items []v1.Deployment
	err = kubectl.Init().
		WithContext(c.Request.Context()).
		Resource(&item).
		Namespace("default").
		List(&items).Error
	if err != nil {
		fmt.Printf("List Error %v\n", err)
	}
	fmt.Printf("List Deployment count %d\n", len(items))
	for _, d := range items {
		fmt.Printf("List Deployment Items foreach %s\n", d.Spec.Template.Spec.Containers[0].Image)
	}

	var crontabList []unstructured.Unstructured
	err = kubectl.Init().
		WithContext(c.Request.Context()).
		CRD("stable.example.com", "v1", "CronTab").
		Namespace("default").
		List(&crontabList).Error
	fmt.Printf("List crd crontabList count %d\n", len(crontabList))
	for _, d := range crontabList {
		fmt.Printf("List Deployment Items foreach %s\n", d.GetName())
	}
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
