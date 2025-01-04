package service

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/weibaohui/k8m/pkg/flag"
	"github.com/weibaohui/kom/kom"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

type clusterService struct {
	clusterConfigs []*clusterConfig // 文件名+context名称 -> 集群配置
}

func (c *clusterService) Reconnect(fileName string, contextName string) {
	c.RegisterCluster(fileName, contextName)
}

func (c *clusterService) Scan() {
	c.clusterConfigs = []*clusterConfig{}
	cfg := flag.Init()
	c.ListClustersInPath(cfg.KubeConfig)
}

func (c *clusterService) AllClusters() []*clusterConfig {
	return c.clusterConfigs
}

type clusterConfig struct {
	FileName      string       `json:"fileName,omitempty"`      // kubeconfig 文件名称
	ContextName   string       `json:"contextName,omitempty"`   // context名称
	ClusterName   string       `json:"clusterName,omitempty"`   // 集群名称
	Server        string       `json:"server,omitempty"`        // 集群地址
	ServerVersion string       `json:"serverVersion,omitempty"` // 通过这个值来判断集群是否可用
	UserName      string       `json:"userName,omitempty"`      // 用户名
	restConfig    *rest.Config // 直连rest.Config
	kubeConfig    []byte       // 集群配置.kubeconfig原始文件内容
	Err           string       `json:"err,omitempty"` // 连接错误信息
}

func (c *clusterService) ListClustersInPath(path string) {
	// 1. 通过kubeconfig文件，找到所在目录
	dir := filepath.Dir(path)

	// 2. 通过所在目录，找到同目录下的所有文件
	files, err := os.ReadDir(dir)
	if err != nil {
		klog.V(6).Infof("Error reading directory: %v", err)
		return
	}

	// 3. 检查每个文件是否为有效的kubeconfig文件

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(dir, file.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			klog.V(6).Infof("Error reading file: %v", err)
			continue
		}

		config, err := clientcmd.Load(content)
		if err != nil {
			continue // 解析失败，跳过该文件
		}
		for contextName, _ := range config.Contexts {
			clusterConfig := &clusterConfig{
				FileName:    file.Name(),
				ContextName: contextName,
				UserName:    config.Contexts[contextName].AuthInfo,
				ClusterName: config.Contexts[contextName].Cluster,
				kubeConfig:  content,
			}
			clusterConfig.Server = config.Clusters[contextName].Server
			c.clusterConfigs = append(c.clusterConfigs, clusterConfig)
		}
	}

	// 注册
	for _, clusterConfig := range c.clusterConfigs {
		c.RegisterCluster(clusterConfig.FileName, clusterConfig.ContextName)
	}
	// 打印serverVersion
	for _, clusterConfig := range c.clusterConfigs {
		klog.V(6).Infof("ServerVersion: %s/%s: %s[%s] using user: %s", clusterConfig.FileName, clusterConfig.ContextName, clusterConfig.ServerVersion, clusterConfig.Server, clusterConfig.UserName)
	}
}

// RegisterCluster 注册集群
func (c *clusterService) RegisterCluster(fileName string, contextName string) {

	for _, clusterConfig := range c.clusterConfigs {
		if clusterConfig.FileName == fileName && clusterConfig.ContextName == contextName {

			// 定义集群ID
			clusterID := fileName + "/" + contextName
			// 先检查连接是否可以直连，如果可以直连，则直接注册
			if c.CheckCluster(fileName, contextName) {
				_, err := kom.Clusters().RegisterByConfigWithID(clusterConfig.restConfig, clusterID)
				if err != nil {
					klog.V(6).Infof("Error registering cluster: %s/%s: %v", fileName, contextName, err)
					continue
				}
				klog.V(6).Infof("Successfully registered cluster: %s/%s", fileName, contextName)
			}
		}
	}
}

// CheckCluster 校验集群是否可连接，并更新状态
func (c *clusterService) CheckCluster(fileName string, contextName string) bool {
	for i := range c.clusterConfigs {
		config := c.clusterConfigs[i]
		if config.FileName == fileName && config.ContextName == contextName {
			lines := strings.Split(string(config.kubeConfig), "\n")
			for i, line := range lines {
				if strings.HasPrefix(line, "current-context:") {
					lines[i] = "current-context: " + contextName
				}
			}
			bytes := []byte(strings.Join(lines, "\n"))

			restConfig, err := clientcmd.RESTConfigFromKubeConfig(bytes)
			if err != nil {
				klog.V(6).Infof("Error creating rest.Config for context %s/%s: %v", fileName, contextName, err)
				config.Err = err.Error()
				return false
			}

			// 校验集群是否可连接
			clientset, err := kubernetes.NewForConfig(restConfig)
			if err != nil {
				klog.V(6).Infof("Error creating clientset for context %s/%s: %v", fileName, contextName, err)
				config.Err = err.Error()
				return false
			}

			// 尝试获取集群版本以验证连接
			info, err := clientset.ServerVersion()
			if err != nil {
				klog.V(6).Infof("Error connecting to cluster for context %s/%s: %v", fileName, contextName, err)
				config.Err = err.Error()
				return false
			}
			klog.V(6).Infof("Successfully connected to cluster for context %s/%s", fileName, contextName)
			// 可以连接的放到数组中记录
			config.ServerVersion = info.GitVersion
			config.restConfig = restConfig
			return true
		}
	}
	return false
}
