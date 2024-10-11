package dynamic

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/kubectl"
	"github.com/weibaohui/k8m/internal/utils/amis"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func List(c *gin.Context) {
	ns := c.Param("ns")
	group := c.Param("group")
	kind := c.Param("kind")
	var list []unstructured.Unstructured
	var err error
	if rt, err := kubectl.ParseResourceType(kind); err == nil {
		// 内置资源
		list, err = kubectl.Init().ListResources(rt, ns)
	} else {
		// CRD 类型资源
		if crd, err := kubectl.Init().GetCRD(kind, group); err == nil {
			list, err = kubectl.Init().ListCRD(crd, ns)
		}
	}

	amis.WriteJsonListWithError(c, list, err)
}
func Fetch(c *gin.Context) {
	var ns = c.Param("ns")
	var name = c.Param("name")
	kind := c.Param("kind")
	group := c.Param("group")

	var obj *unstructured.Unstructured
	var err error
	rt, err := kubectl.ParseResourceType(kind)
	if err != nil {
		// CRD 类型资源
		if crd, err := kubectl.Init().GetCRD(kind, group); err == nil {
			obj, err = kubectl.Init().FetchCRD(crd, ns, name)
			if err != nil {
				amis.WriteJsonError(c, err)
				return
			}
		}
	} else {
		obj, err = kubectl.Init().GetResource(rt, ns, name)
		if err != nil {
			amis.WriteJsonError(c, err)
			return
		}
	}

	yamlStr, err := kubectl.Init().ConvertUnstructuredToYAML(obj)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, gin.H{
		"yaml": yamlStr,
	})
}
func Remove(c *gin.Context) {
	var ns = c.Param("ns")
	var name = c.Param("name")
	kind := c.Param("kind")
	group := c.Param("group")

	err := removeSingle(kind, group, ns, name)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)

}
func removeSingle(kind, group, ns, name string) error {
	rt, err := kubectl.ParseResourceType(kind)
	if err != nil {
		// CRD 类型资源
		if crd, err := kubectl.Init().GetCRD(kind, group); err == nil {
			err = kubectl.Init().RemoveCRD(crd, ns, name)
			if err != nil {
				return err
			}
		}
	} else {
		// 内置资源类型
		err = kubectl.Init().RemoveResource(rt, ns, name)
		if err != nil {
			return err
		}
	}
	return err
	// todo 校验是否有权限删除，ns为为本人名字开头

}

// NamesPayload 定义结构体以匹配批量删除 JSON 结构
type NamesPayload struct {
	Names []string `json:"names"`
}

func BatchRemove(c *gin.Context) {
	var ns = c.Param("ns")
	kind := c.Param("kind")
	group := c.Param("group")

	// 初始化结构体实例
	var payload NamesPayload

	// 反序列化 JSON 数据到结构体
	if err := c.ShouldBindJSON(&payload); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	for _, name := range payload.Names {
		_ = removeSingle(kind, group, ns, name)
	}
	amis.WriteJsonOK(c)
}

type ApplyYAMLRequest struct {
	YAML string `json:"yaml" binding:"required"`
}

func Save(c *gin.Context) {
	var ns = c.Param("ns")
	var name = c.Param("name")
	kind := c.Param("kind")
	group := c.Param("group")

	var req ApplyYAMLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	yamlStr := req.YAML

	// 解析 YAML 到 Unstructured 对象
	var obj unstructured.Unstructured
	if err := yaml.Unmarshal([]byte(yamlStr), &obj.Object); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	obj.SetName(name)
	obj.SetNamespace(ns)

	rt, err := kubectl.ParseResourceType(kind)
	if err != nil {
		// CRD 类型资源
		if crd, err := kubectl.Init().GetCRD(kind, group); err == nil {
			_, err = kubectl.Init().UpdateCRD(crd, &obj)
			if err != nil {
				amis.WriteJsonError(c, err)
				return
			}
		}
	} else {
		_, err = kubectl.Init().UpdateResource(rt, ns, &obj)
		if err != nil {
			amis.WriteJsonError(c, err)
			return
		}
	}

	amis.WriteJsonOK(c)
	// todo 做一个机制，限制每个人的可操作ns，只能是自己权限下的ns,
	// todo 给资源增加label标签 ，后续按ns、标签进行过滤
}

func Apply(c *gin.Context) {
	var req ApplyYAMLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	yamlStr := req.YAML
	result := kubectl.Init().ApplyYAML(yamlStr)
	amis.WriteJsonData(c, gin.H{
		"result": result,
	})
	// todo 校验是否有权限创建ns，ns名称必须为本人名字开头

}
func Delete(c *gin.Context) {
	var req ApplyYAMLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	yamlStr := req.YAML
	result := kubectl.Init().DeleteYAML(yamlStr)
	amis.WriteJsonData(c, gin.H{
		"result": result,
	})
	// todo 校验是否有权限删除，label中owner是否为本人名字开头
}
