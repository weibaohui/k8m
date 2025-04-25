package ns

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func CreateLimitRange(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var data struct {
		Name     string `json:"name"`
		Metadata struct {
			Namespace string `json:"namespace"`
		} `json:"metadata"`
		Spec struct {
			Limits []struct {
				Type           string            `json:"type"`
				Default        map[string]string `json:"default"`
				DefaultRequest map[string]string `json:"defaultRequest"`
				Max            map[string]string `json:"max"`
				Min            map[string]string `json:"min"`
			} `json:"limits"`
		} `json:"spec"`
	}

	if err := c.ShouldBindJSON(&data); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	limitRange := &v1.LimitRange{
		ObjectMeta: metav1.ObjectMeta{
			Name:      data.Name,
			Namespace: data.Metadata.Namespace,
		},
		Spec: v1.LimitRangeSpec{
			Limits: []v1.LimitRangeItem{},
		},
	}

	// 处理limits资源
	for _, limit := range data.Spec.Limits {
		limitItem := v1.LimitRangeItem{
			Type:           v1.LimitType(limit.Type),
			Default:        make(v1.ResourceList),
			DefaultRequest: make(v1.ResourceList),
			Max:            make(v1.ResourceList),
			Min:            make(v1.ResourceList),
		}

		// 处理Default资源
		// 当类型为Pod时，不处理Default资源
		if string(limitItem.Type) != string(v1.LimitTypePod) {
			for name, value := range limit.Default {
				// 检查是否为小数值
				if name == "cpu" || name == "memory" {
					if utils.IsDecimal(value) {
						amis.WriteJsonError(c, fmt.Errorf("资源值不能为小数，请使用整数值: %s=%s", name, value))
						return
					}
				}

				if name == "cpu" {
					value = fmt.Sprintf("%sm", value)
				}
				if name == "memory" {
					value = fmt.Sprintf("%sMi", value)
				}
				quantity, err := resource.ParseQuantity(value)
				if err != nil {
					amis.WriteJsonError(c, err)
					return
				}
				limitItem.Default[v1.ResourceName(name)] = quantity
			}

			// 处理DefaultRequest资源
			for name, value := range limit.DefaultRequest {
				// 检查是否为小数值
				if name == "cpu" || name == "memory" {
					if utils.IsDecimal(value) {
						amis.WriteJsonError(c, fmt.Errorf("资源值不能为小数，请使用整数值: %s=%s", name, value))
						return
					}
				}

				if name == "cpu" {
					value = fmt.Sprintf("%sm", value)
				}
				if name == "memory" {
					value = fmt.Sprintf("%sMi", value)
				}
				quantity, err := resource.ParseQuantity(value)
				if err != nil {
					amis.WriteJsonError(c, err)
					return
				}
				limitItem.DefaultRequest[v1.ResourceName(name)] = quantity
			}
		}

		// 处理Max资源
		for name, value := range limit.Max {
			// 检查是否为小数值
			if name == "cpu" || name == "memory" {
				if utils.IsDecimal(value) {
					amis.WriteJsonError(c, fmt.Errorf("资源值不能为小数，请使用整数值: %s=%s", name, value))
					return
				}
			}

			if name == "cpu" {
				value = fmt.Sprintf("%sm", value)
			}
			if name == "memory" {
				value = fmt.Sprintf("%sMi", value)
			}
			quantity, err := resource.ParseQuantity(value)
			if err != nil {
				amis.WriteJsonError(c, err)
				return
			}
			limitItem.Max[v1.ResourceName(name)] = quantity
		}

		// 处理Min资源
		for name, value := range limit.Min {
			// 检查是否为小数值
			if name == "cpu" || name == "memory" {
				if utils.IsDecimal(value) {
					amis.WriteJsonError(c, fmt.Errorf("资源值不能为小数，请使用整数值: %s=%s", name, value))
					return
				}
			}

			if name == "cpu" {
				value = fmt.Sprintf("%sm", value)
			}
			if name == "memory" {
				value = fmt.Sprintf("%sMi", value)
			}
			quantity, err := resource.ParseQuantity(value)
			if err != nil {
				amis.WriteJsonError(c, err)
				return
			}
			limitItem.Min[v1.ResourceName(name)] = quantity
		}

		limitRange.Spec.Limits = append(limitRange.Spec.Limits, limitItem)
	}
	klog.Infof("limitRange: %v", utils.ToJSON(limitRange))
	err = kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(limitRange).
		Name(data.Name).
		Namespace(data.Metadata.Namespace).
		Create(&limitRange).Error

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOK(c)
}
