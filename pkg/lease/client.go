package lease

import (
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

// GetRestConfigByClusterID 中文函数注释：可选注入的解析器，根据 ClusterID 返回对应的 RestConfig。
// 通过在外部（service 包）设置该回调，避免 lease 包直接依赖 service 包，打破循环引入。
var GetRestConfigByClusterID func(clusterID string) *rest.Config

// GetClientset 获取 Kubernetes ClientSet
// 优先顺序：
// 1. 指定 ClusterID（外部集群）
// 2. InCluster 配置
// 3. 本地 KUBECONFIG 或 ~/.kube/config
// 返回值第二个布尔含义：是否存在可用集群（用于决定是否进行选举）
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
		klog.V(6).Infof("[leader] InCluster 创建 ClientSet 失败：%v，继续尝试本地 kubeconfig", cErr)
	}

	// Fallback 到本地 kubeconfig
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = os.ExpandEnv("$HOME/.kube/config")
	}
	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		klog.V(6).Infof("[leader] 本地 kubeconfig 加载失败：%v", err)
		// 无可用集群，返回非错误以便降级为本地 Leader
		return nil, false, nil
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.V(6).Infof("[leader] 本地 kubeconfig 创建 ClientSet 失败：%v", err)
		return nil, false, nil
	}
	return clientset, true, nil
}

// DetectNamespace 中文函数注释：自动检测当前运行的命名空间；InCluster 读取 SA 文件，否则返回 default。
func DetectNamespace() string {
	data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err == nil {
		return string(data)
	}
	return "default"
}
