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
	"k8s.io/klog/v2"
)

func ImagePullSecretOptionList(c *gin.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	group := c.Param("group")
	kind := c.Param("kind")
	version := c.Param("version")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var item *unstructured.Unstructured
	err = kom.Cluster(selectedCluster).WithContext(ctx).
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
	if len(imagePullSecrets) == 0 {
		amis.WriteJsonData(c, gin.H{
			"options": make([]map[string]string, 0),
		})
		return
	}
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
	if len(secretsList) == 0 {
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

func ContainerResourcesInfo(c *gin.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	group := c.Param("group")
	kind := c.Param("kind")
	version := c.Param("version")
	containerName := c.Param("container_name")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var item *unstructured.Unstructured
	err = kom.Cluster(selectedCluster).WithContext(ctx).
		CRD(group, version, kind).
		Namespace(ns).
		Name(name).Get(&item).Error

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	requestCPU, limitCPU, requestMemory, limitMemory, err := getContainerResourcesInfoByName(item, containerName)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, gin.H{
		"name":           containerName,
		"request_cpu":    requestCPU,
		"limit_cpu":      limitCPU,
		"request_memory": requestMemory,
		"limit_memory":   limitMemory,
	})

}

func getContainerResourcesInfoByName(item *unstructured.Unstructured, containerName string) (string, string, string, string, error) {
	// 获取资源类型
	kind := item.GetKind()

	// 根据资源类型获取 containers 的路径
	resourcePaths, err := getResourcePaths(kind)
	if err != nil {
		return "", "", "", "", err
	}
	containersPath := append(resourcePaths, "containers")

	// 获取嵌套字段
	containers, found, err := unstructured.NestedSlice(item.Object, containersPath...)
	if err != nil {
		return "", "", "", "", fmt.Errorf("error getting containers: %w", err)
	}
	if !found {
		return "", "", "", "", fmt.Errorf("containers field not found")
	}

	// 遍历 containers 列表
	for _, container := range containers {
		// 断言 container 类型为 map[string]interface{}
		containerMap, ok := container.(map[string]interface{})
		if !ok {
			return "", "", "", "", fmt.Errorf("unexpected container format")
		}

		// 获取容器的 name
		name, _, err := unstructured.NestedString(containerMap, "name")
		if err != nil {
			return "", "", "", "", fmt.Errorf("error getting container name: %w", err)
		}

		// 如果 name 匹配目标容器名，则获取其 image
		if name == containerName {
			// 获取 resources
			resourcesMap, found, err := unstructured.NestedMap(containerMap, "resources")
			if err != nil {
				return "", "", "", "", fmt.Errorf("error getting container resources: %w", err)
			}
			if !found {
				return "", "", "", "", nil // 如果没有 resources 字段，返回空字符串
			}

			// 获取 requests
			requestsMap, found, err := unstructured.NestedMap(resourcesMap, "requests")
			if err != nil {
				return "", "", "", "", fmt.Errorf("error getting container requests: %w", err)
			}
			if !found {
				requestsMap = make(map[string]interface{}) // 如果没有 requests 字段，初始化为空 map
			}

			// 获取 limits
			limitsMap, found, err := unstructured.NestedMap(resourcesMap, "limits")
			if err != nil {
				return "", "", "", "", fmt.Errorf("error getting container limits: %w", err)
			}
			if !found {
				limitsMap = make(map[string]interface{}) // 如果没有 limits 字段，初始化为空 map
			}

			// 获取 request CPU
			requestCPU, _, err := unstructured.NestedString(requestsMap, "cpu")
			if err != nil {
				return "", "", "", "", fmt.Errorf("error getting container request cpu: %w", err)
			}

			// 获取 request 内存
			requestMemory, _, err := unstructured.NestedString(requestsMap, "memory")
			if err != nil {
				return "", "", "", "", fmt.Errorf("error getting container request memory: %w", err)
			}

			// 获取 limit CPU
			limitCPU, _, err := unstructured.NestedString(limitsMap, "cpu")
			if err != nil {
				return "", "", "", "", fmt.Errorf("error getting container limit cpu: %w", err)
			}

			// 获取 limit 内存
			limitMemory, _, err := unstructured.NestedString(limitsMap, "memory")
			if err != nil {
				return "", "", "", "", fmt.Errorf("error getting container limit memory: %w", err)
			}

			return requestCPU, limitCPU, requestMemory, limitMemory, nil
		}
	}

	// 如果未找到匹配的容器名
	return "", "", "", "", fmt.Errorf("container with name %q not found", containerName)
}

// 资源信息结构体
// json
// {"container_name":"my-container","request_cpu":"1","request_memory":"1","request_memory":"1Gi","limit_memory":"1Gi"}
type resourceInfo struct {
	ContainerName string `json:"container_name"`
	RequestCpu    string `json:"request_cpu"`
	LimitCpu      string `json:"limit_cpu"`
	RequestMemory string `json:"request_memory"`
	LimitMemory   string `json:"limit_memory"`
}

func UpdateResources(c *gin.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	group := c.Param("group")
	kind := c.Param("kind")
	version := c.Param("version")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var info resourceInfo

	if err := c.ShouldBindJSON(&info); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	patchData, err := generateResourcePatch(kind, info)
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

// 生成资源patch数据
func generateResourcePatch(kind string, info resourceInfo) (map[string]interface{}, error) {
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

	// 构造资源请求和限制
	resources := make(map[string]interface{})

	// 设置请求资源
	requests := make(map[string]string)
	if info.RequestCpu != "" {
		requests["cpu"] = info.RequestCpu
	}
	if info.RequestMemory != "" {
		requests["memory"] = info.RequestMemory
	}
	if len(requests) > 0 {
		resources["requests"] = requests
	}

	// 设置限制资源
	limits := make(map[string]string)
	if info.LimitCpu != "" {
		limits["cpu"] = info.LimitCpu
	}
	if info.LimitMemory != "" {
		limits["memory"] = info.LimitMemory
	}
	if len(limits) > 0 {
		resources["limits"] = limits
	}

	// 构造容器数组
	current["containers"] = []map[string]interface{}{
		{
			"name":      info.ContainerName,
			"resources": resources,
		},
	}

	return patch, nil
}

func ContainerInfo(c *gin.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	group := c.Param("group")
	kind := c.Param("kind")
	version := c.Param("version")
	containerName := c.Param("container_name")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var item *unstructured.Unstructured
	err = kom.Cluster(selectedCluster).WithContext(ctx).
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

// 获取container的环境变量信息
func ContainerEnvInfo(c *gin.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	group := c.Param("group")
	kind := c.Param("kind")
	version := c.Param("version")
	containerName := c.Param("container_name")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var item *unstructured.Unstructured
	err = kom.Cluster(selectedCluster).WithContext(ctx).
		CRD(group, version, kind).
		Namespace(ns).
		Name(name).Get(&item).Error

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	// 返回格式为
	envVars, err := getContainerEnvVarsByName(item, containerName)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, gin.H{
		"envs": envVars,
	})

}

type ContainerEnv struct {
	ContainerName string            `json:"container_name"`
	Envs          map[string]string `json:"envs"`
}

// 更新container的环境变量信息
func UpdateContainerEnv(c *gin.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	group := c.Param("group")
	kind := c.Param("kind")
	version := c.Param("version")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var info ContainerEnv
	if err := c.ShouldBindJSON(&info); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	patchData, err := generateEnvPatch(kind, info)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	patchJSON := utils.ToJSON(patchData)
	var patch interface{}
	err = kom.Cluster(selectedCluster).
		WithContext(ctx).
		CRD(group, version, kind).
		Namespace(ns).Name(name).
		Patch(&patch, types.StrategicMergePatchType, patchJSON).Error
	amis.WriteJsonErrorOrOK(c, err)
}

// 验证环境变量名称是否符合Kubernetes规范
func isValidEnvVarName(name string) bool {
	if len(name) == 0 {
		return false
	}
	if name[0] >= '0' && name[0] <= '9' {
		return false
	}

	for _, c := range name {
		if !((c >= 'a' && c <= 'z') ||
			(c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') ||
			c == '_' || c == '-' || c == '.') {
			return false
		}
	}
	return true
}

func generateEnvPatch(kind string, info ContainerEnv) (map[string]interface{}, error) {
	// 获取资源路径
	paths, err := getResourcePaths(kind)
	if err != nil {
		return nil, err
	}
	// 动态构造 patch 数据
	patch := make(map[string]interface{})
	// 验证环境变量名称是否符合规范
	for key := range info.Envs {
		if !isValidEnvVarName(key) {
			return nil, fmt.Errorf("环境变量名'%s'无效: 只能包含字母、数字、_、-或.，且不能以数字开头", key)
		}
	}

	current := patch

	// 按层级动态生成嵌套结构
	for _, path := range paths {
		if _, exists := current[path]; !exists {
			current[path] = make(map[string]interface{})
		}
		current = current[path].(map[string]interface{})
	}

	// 构造环境变量数组
	envArray := make([]map[string]string, 0)
	for key, value := range info.Envs {
		// 验证key是否合法

		envArray = append(envArray, map[string]string{
			"name":  key,
			"value": value,
		})
	}
	// 如果没有环境变量，设置envarray为nil
	if len(envArray) == 0 {
		envArray = nil
	}
	// 构造容器数组
	current["containers"] = []map[string]interface{}{
		{
			"name": info.ContainerName,
			"env":  envArray,
		},
	}

	return patch, nil
}

// 获取container的env信息
func getContainerEnvVarsByName(item *unstructured.Unstructured, containerName string) (map[string]string, error) {
	// 获取资源类型
	kind := item.GetKind()

	// 根据资源类型获取 containers 的路径
	resourcePaths, err := getResourcePaths(kind)
	if err != nil {
		return nil, err
	}
	containersPath := append(resourcePaths, "containers")

	// 获取嵌套字段
	containers, found, err := unstructured.NestedSlice(item.Object, containersPath...)
	if err != nil {
		return nil, fmt.Errorf("error getting containers: %w", err)
	}
	if !found {
		return nil, fmt.Errorf("containers field not found")
	}
	// 遍历 containers 列表
	for _, container := range containers {
		// 断言 container 类型为 map[string]interface{}
		containerMap, ok := container.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected container format")
		}
		// 获取容器的 name
		name, _, err := unstructured.NestedString(containerMap, "name")
		if err != nil {
			return nil, fmt.Errorf("error getting container name: %w", err)
		}
		// 如果 name 匹配目标容器名，则获取其 image
		if name == containerName {
			// 获取 env 字段
			env, found, err := unstructured.NestedSlice(containerMap, "env")
			if err != nil {
				return nil, fmt.Errorf("error getting container env: %w", err)
			}
			// 如果未找到 env 字段，则返回空列表
			if !found {
				return make(map[string]string), nil
			}
			// 遍历 env 列表
			envVars := make(map[string]string)
			for _, envVar := range env {
				// 断言 envVar 类型为 map[string]interface{}
				envVarMap, ok := envVar.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("unexpected envVar format")
				}
				// 获取 envVar 的 name 和 value
				name, _, err := unstructured.NestedString(envVarMap, "name")
				if err != nil {
					return nil, fmt.Errorf("error getting envVar name: %w", err)
				}
				value, _, err := unstructured.NestedString(envVarMap, "value")
				if err != nil {
					return nil, fmt.Errorf("error getting envVar value: %w", err)
				}
				// 将 name 和 value 组合成 map 并添加到结果列表中
				envVars[name] = value
			}
			return envVars, nil
		}
	}
	// 如果未找到匹配的容器名
	return nil, fmt.Errorf("container with name %q not found", containerName)
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
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

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

// 接口 获取容器健康检查信息
func ContainerHealthChecksInfo(c *gin.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	group := c.Param("group")
	kind := c.Param("kind")
	version := c.Param("version")
	containerName := c.Param("container_name")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var item *unstructured.Unstructured
	err = kom.Cluster(selectedCluster).WithContext(ctx).
		CRD(group, version, kind).
		Namespace(ns).
		Name(name).Get(&item).Error

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	healthChecks, err := getContainerHealthChecksByName(item, containerName)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonData(c, gin.H{
		"container_name":  containerName,
		"readiness_probe": healthChecks["readinessProbe"],
		"liveness_probe":  healthChecks["livenessProbe"],
	})
}

// 获取容器健康检查信息
func getContainerHealthChecksByName(item *unstructured.Unstructured, containerName string) (map[string]interface{}, error) {
	// 获取资源类型
	kind := item.GetKind()

	// 根据资源类型获取 containers 的路径
	resourcePaths, err := getResourcePaths(kind)
	if err != nil {
		return nil, err
	}
	containersPath := append(resourcePaths, "containers")

	// 获取嵌套字段
	containers, found, err := unstructured.NestedSlice(item.Object, containersPath...)
	if err != nil {
		return nil, fmt.Errorf("error getting containers: %w", err)
	}
	if !found {
		return nil, fmt.Errorf("containers field not found")
	}

	// 遍历 containers 列表
	for _, container := range containers {
		containerMap, ok := container.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected container format")
		}

		name, _, err := unstructured.NestedString(containerMap, "name")
		if err != nil {
			return nil, fmt.Errorf("error getting container name: %w", err)
		}

		if name == containerName {
			result := make(map[string]interface{})

			// 获取就绪检查
			if readinessProbe, found, _ := unstructured.NestedMap(containerMap, "readinessProbe"); found {
				result["readinessProbe"] = readinessProbe
			}

			// 获取存活检查
			if livenessProbe, found, _ := unstructured.NestedMap(containerMap, "livenessProbe"); found {
				result["livenessProbe"] = livenessProbe
			}

			return result, nil
		}
	}

	return nil, fmt.Errorf("container with name %q not found", containerName)
}

// 健康检查配置结构体
type HealthCheckInfo struct {
	ContainerName  string                 `json:"container_name"`
	LivenessType   string                 `json:"liveness_type"`
	ReadinessType  string                 `json:"readiness_type"`
	ReadinessProbe map[string]interface{} `json:"readiness_probe,omitempty"`
	LivenessProbe  map[string]interface{} `json:"liveness_probe,omitempty"`
}

// 接口 更新容器健康检查
func UpdateHealthChecks(c *gin.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	group := c.Param("group")
	kind := c.Param("kind")
	version := c.Param("version")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var info HealthCheckInfo
	if err := c.ShouldBindJSON(&info); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	patchData, err := generateHealthCheckPatch(kind, info)
	klog.V(6).Infof("UpdateHealthChecks Patch JSON :\n%s\n", patchData)
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
		Patch(&item, types.StrategicMergePatchType, patchJSON).Error // 指定 Patch 类型
	amis.WriteJsonErrorOrOK(c, err)
}

// 生成健康检查的 patch 数据
func generateHealthCheckPatch(kind string, info HealthCheckInfo) (map[string]interface{}, error) {
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
	// 判断健康检查类型
	if info.LivenessType == "httpGet" {
		info.LivenessProbe["exec"] = nil
		info.LivenessProbe["tcpSocket"] = nil
	} else if info.LivenessType == "exec" {
		info.LivenessProbe["httpGet"] = nil
		info.LivenessProbe["tcpSocket"] = nil
	} else if info.LivenessType == "tcpSocket" {
		info.LivenessProbe["httpGet"] = nil
		info.LivenessProbe["exec"] = nil
	} else {
		info.LivenessProbe = nil
	}
	if info.ReadinessType == "httpGet" {
		info.ReadinessProbe["exec"] = nil
		info.ReadinessProbe["tcpSocket"] = nil
	} else if info.ReadinessType == "exec" {
		info.ReadinessProbe["httpGet"] = nil
		info.ReadinessProbe["tcpSocket"] = nil
	} else if info.ReadinessType == "tcpSocket" {
		info.ReadinessProbe["httpGet"] = nil
		info.ReadinessProbe["exec"] = nil
	} else {
		info.ReadinessProbe = nil
	}

	// 构造容器数组
	current["containers"] = []map[string]interface{}{
		{
			"name":           info.ContainerName,
			"livenessProbe":  info.LivenessProbe,
			"readinessProbe": info.ReadinessProbe,
		},
	}

	return patch, nil
}

// 返回资源类型对应的路径
func getResourcePaths(kind string) ([]string, error) {
	switch kind {
	case "CloneSet":
		return []string{"spec", "template", "spec"}, nil
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
