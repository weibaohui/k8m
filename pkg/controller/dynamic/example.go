package dynamic

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/kubectl"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
)

func Example(c *gin.Context) {
	item := v1.Deployment{}
	err := kubectl.Init().
		WithContext(c.Request.Context()).
		Resource(&item).
		Namespace("default").
		Name("ci-755702-codexxx").
		Get(&item).Error
	if err != nil {
		klog.Errorf("Deployment Get(&item) error :%v", err)
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
	createItem := v1.Deployment{

		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deploy",
			Namespace: "default",
		},
		Spec: v1.DeploymentSpec{
			Replicas: utils.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "test",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "test",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test",
							Image: "nginx:1.14.2",
						},
					},
				},
			},
		},
	}
	err = kubectl.Init().
		WithContext(c.Request.Context()).
		Resource(&createItem).
		Create(&createItem).Error
	if err != nil {
		klog.Errorf("Deployment Create(&item) error :%v", err)
	}

}
