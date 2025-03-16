package dynamic

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
)

func ImagePullSecretOptionList(c *gin.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	group := c.Param("group")
	kind := c.Param("kind")
	version := c.Param("version")
	ctx := amis.GetContextWithUser(c)
	selectedCluster := amis.GetSelectedCluster(c)

	var item *unstructured.Unstructured
	err := kom.Cluster(selectedCluster).WithContext(ctx).
		CRD(group, version, kind).
		Namespace(ns).
		Name(name).Get(&item).Error

	if err != nil {
		amis.WriteJsonData(c, gin.H{
			"options": make([]map[string]string, 0),
		})
		return
	}
	imagePullSecrets, _ := getImagePullSecrets(item)

	// 从Secret中寻找镜像拉取密钥
	// 获取list
	var secretsList []*v1.Secret
	err = kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.Secret{}).
		Namespace(ns).
		Where(fmt.Sprintf("type = '%s' or type = '%s' ", v1.SecretTypeDockerConfigJson, v1.SecretTypeDockercfg)).
		List(&secretsList).Error
	if err != nil {
		amis.WriteJsonData(c, gin.H{
			"options": make([]map[string]string, 0),
		})
		return
	}
	var options []map[string]string
	for _, s := range secretsList {
		options = append(options, map[string]string{
			"label": s.Name,
			"value": s.Name,
		})
	}

	amis.WriteJsonData(c, gin.H{
		"options": options,
		"value":   strings.Join(imagePullSecrets, ","),
	})
}

func ContainerInfo(c *gin.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	group := c.Param("group")
	kind := c.Param("kind")
	version := c.Param("version")
	containerName := c.Param("container_name")
	ctx := amis.GetContextWithUser(c)
	selectedCluster := amis.GetSelectedCluster(c)

	var item *unstructured.Unstructured
	err := kom.Cluster(selectedCluster).WithContext(ctx).
		CRD(group, version, kind).
		Namespace(ns).
		Name(name).Get(&item).Error

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	imageFullName, imagePullPolicy, err := getContainerImageByName(item, containerName)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	image, tag := utils.GetImageNameAndTag(imageFullName)
	amis.WriteJsonData(c, gin.H{
		"name":              containerName,
		"image":             image,
		"tag":               tag,
		"image_pull_policy": imagePullPolicy,
	})

}

// 获取 imagePullSecrets 列表
func getImagePullSecrets(item *unstructured.Unstructured) ([]string, error) {
	// 获取资源类型
	kind := item.GetKind()

	// 根据资源类型获取 imagePullSecrets 的路径
	resourcePaths, err := getResourcePaths(kind)
	if err != nil {
		return nil, err
	}
	imagePullSecretsPath := append(resourcePaths, "imagePullSecrets")

	// 获取嵌套字段
	secrets, found, err := unstructured.NestedSlice(item.Object, imagePullSecretsPath...)
	if err != nil {
		return nil, fmt.Errorf("error getting imagePullSecrets: %w", err)
	}
	if !found {
		return nil, fmt.Errorf("imagePullSecrets field not found for kind %q", kind)
	}

	// 提取 secret name 列表
	var secretNames []string
	for _, secret := range secrets {
		secretMap, ok := secret.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected imagePullSecret format")
		}

		name, found, err := unstructured.NestedString(secretMap, "name")
		if err != nil {
			return nil, fmt.Errorf("error getting imagePullSecret name: %w", err)
		}
		if found {
			secretNames = append(secretNames, name)
		}
	}

	return secretNames, nil
}
func getContainerImageByName(item *unstructured.Unstructured, containerName string) (string, string, error) {
	// 获取资源类型
	kind := item.GetKind()

	// 根据资源类型获取 containers 的路径
	resourcePaths, err := getResourcePaths(kind)
	if err != nil {
		return "", "", err
	}
	containersPath := append(resourcePaths, "containers")

	// 获取嵌套字段
	containers, found, err := unstructured.NestedSlice(item.Object, containersPath...)
	if err != nil {
		return "", "", fmt.Errorf("error getting containers: %w", err)
	}
	if !found {
		return "", "", fmt.Errorf("containers field not found")
	}

	// 遍历 containers 列表
	for _, container := range containers {
		// 断言 container 类型为 map[string]interface{}
		containerMap, ok := container.(map[string]interface{})
		if !ok {
			return "", "", fmt.Errorf("unexpected container format")
		}

		// 获取容器的 name
		name, _, err := unstructured.NestedString(containerMap, "name")
		if err != nil {
			return "", "", fmt.Errorf("error getting container name: %w", err)
		}

		// 如果 name 匹配目标容器名，则获取其 image
		if name == containerName {
			image, _, err := unstructured.NestedString(containerMap, "image")
			if err != nil {
				return "", "", fmt.Errorf("error getting container image: %w", err)
			}
			imagePullPolicy, _, err := unstructured.NestedString(containerMap, "imagePullPolicy")
			if err != nil {
				return "", "", fmt.Errorf("error getting container imagePullPolicy: %w", err)
			}
			return image, imagePullPolicy, nil
		}
	}

	// 如果未找到匹配的容器名
	return "", "", fmt.Errorf("container with name %q not found", containerName)
}

// json
// {"container_name":"my-container","image":"my-image","name":"my-container","tag":"sss1","image_pull_secrets":"myregistrykey"}
type imageInfo struct {
	ContainerName    string `json:"container_name"`
	Image            string `json:"image"`
	Tag              string `json:"tag"`
	ImagePullSecrets string `json:"image_pull_secrets"`
	ImagePullPolicy  string `json:"image_pull_policy"`
}

func UpdateImageTag(c *gin.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	group := c.Param("group")
	kind := c.Param("kind")
	version := c.Param("version")
	ctx := amis.GetContextWithUser(c)
	selectedCluster := amis.GetSelectedCluster(c)

	var info imageInfo

	if err := c.ShouldBindJSON(&info); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	patchData, err := generateDynamicPatch(kind, info)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	patchJSON := utils.ToJSON(patchData)
	var item interface{}
	err = kom.Cluster(selectedCluster).
		WithContext(ctx).
		CRD(group, version, kind).
		Namespace(ns).Name(name).
		Patch(&item, types.StrategicMergePatchType, patchJSON).Error
	amis.WriteJsonErrorOrOK(c, err)
}

// 生成动态的 patch 数据
func generateDynamicPatch(kind string, info imageInfo) (map[string]interface{}, error) {
	// 获取资源路径
	paths, err := getResourcePaths(kind)
	if err != nil {
		return nil, err
	}

	// 动态构造 patch 数据
	patch := make(map[string]interface{})
	current := patch

	// 按层级动态生成嵌套结构
	for _, path := range paths {
		if _, exists := current[path]; !exists {
			current[path] = make(map[string]interface{})
		}
		current = current[path].(map[string]interface{})
	}

	// 构造 `imagePullSecrets`
	if info.ImagePullSecrets == "" {
		current["imagePullSecrets"] = nil // 删除字段
	} else {
		secretNames := strings.Split(info.ImagePullSecrets, ",")
		imagePullSecrets := make([]map[string]string, 0, len(secretNames))
		for _, name := range secretNames {
			imagePullSecrets = append(imagePullSecrets, map[string]string{"name": name})
		}
		current["imagePullSecrets"] = imagePullSecrets
	}

	// 构造 `containers`
	current["containers"] = []map[string]string{
		{
			"name":            info.ContainerName,
			"image":           fmt.Sprintf("%s:%s", info.Image, info.Tag),
			"imagePullPolicy": info.ImagePullPolicy,
		},
	}

	return patch, nil
}

// 返回资源类型对应的路径
func getResourcePaths(kind string) ([]string, error) {
	switch kind {
	case "Deployment", "DaemonSet", "StatefulSet", "ReplicaSet", "Job":
		return []string{"spec", "template", "spec"}, nil
	case "CronJob":
		return []string{"spec", "jobTemplate", "spec", "template", "spec"}, nil
	case "Pod":
		return []string{"spec"}, nil
	default:
		return nil, fmt.Errorf("unsupported resource kind: %s", kind)
	}
}
