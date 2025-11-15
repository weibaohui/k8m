package utils

import (
	"fmt"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

// GetRestConfigByClusterID 中文函数注释：可选注入的解析器，根据 ClusterID 返回对应的 RestConfig。
// 通过在外部（service 包）设置该回调，避免 utils 包直接依赖具体的集群服务，便于 lease、leader 等模块复用。
var GetRestConfigByClusterID func(clusterID string) *rest.Config

// GetClientSet 中文函数注释：获取 Kubernetes ClientSet。
// 优先顺序：
// 1. 指定 ClusterID（外部集群）
// 2. InCluster 配置
// 3. 本地 KUBECONFIG 或 ~/.kube/config
// 返回值第二个布尔含义：是否存在可用集群（用于决定是否进行选举/租约相关操作）。
func GetClientSet(clusterID string) (*kubernetes.Clientset, bool, error) {
	// 如果提供了 ClusterID，优先尝试使用指定集群的配置
	if clusterID != "" {
		klog.V(6).Infof("[leader] 尝试使用指定的 ClusterID 获取配置：%s", clusterID)
		// 通过注入的回调解析 ClusterID → RestConfig，避免直接依赖 service 包
		if GetRestConfigByClusterID != nil {
			restCfg := GetRestConfigByClusterID(clusterID)
			if restCfg != nil {
				clientset, err := kubernetes.NewForConfig(restCfg)
				if err == nil {
					klog.V(6).Infof("[leader] 已使用指定集群的配置创建 ClientSet：%s", clusterID)
					// 找到可用集群
					return clientset, true, nil
				}
				klog.V(6).Infof("[leader] 使用指定集群创建 ClientSet 失败：%v，继续尝试 InCluster 配置", err)
			} else {
				klog.V(6).Infof("[leader] 指定集群未提供 RestConfig：%s，继续尝试 InCluster 配置", clusterID)
			}
		} else {
			klog.V(6).Infof("[leader] 未注入 ClusterID 解析器，继续尝试 InCluster 配置")
		}
	}

	// 尝试 InCluster 模式
	config, err := rest.InClusterConfig()
	if err == nil {
		clientset, cErr := kubernetes.NewForConfig(config)
		if cErr == nil {
			return clientset, true, nil
		}
		klog.V(6).Infof("InCluster 创建 ClientSet 失败：%v，继续尝试本地 kubeconfig", cErr)
	}

	klog.V(6).Infof("未指定宿主集群ID、未检测到InCluster模式，获取 ClientSet 失败：%v", err)
	return nil, false, nil
}

// DetectNamespace 中文函数注释：自动检测当前运行的命名空间；InCluster 读取 SA 文件，否则返回 default。
func DetectNamespace() string {
	data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err == nil {
		return string(data)
	}
	return "default"
}

// GenerateInstanceID 生成当前实例的唯一身份标识，规则为 hostname-随机3位。
func GenerateInstanceID() string {
	id, err := os.Hostname()
	if err != nil || id == "" {
		id = "unknown-host"
	}
	return fmt.Sprintf("%s-%s", id, RandNLengthString(3))
}
