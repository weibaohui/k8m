package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/robfig/cron/v3"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/flag"
	"github.com/weibaohui/kom/kom"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

type clusterService struct {
	clusterConfigs        []*ClusterConfig // 文件名+context名称 -> 集群配置
	AggregateDelaySeconds int              // 聚合延迟时间
}

type ClusterConfig struct {
	FileName                     string                         `json:"fileName,omitempty"`                     // kubeconfig 文件名称
	ContextName                  string                         `json:"contextName,omitempty"`                  // context名称
	ClusterName                  string                         `json:"clusterName,omitempty"`                  // 集群名称
	Server                       string                         `json:"server,omitempty"`                       // 集群地址
	ServerVersion                string                         `json:"serverVersion,omitempty"`                // 通过这个值来判断集群是否可用
	UserName                     string                         `json:"userName,omitempty"`                     // 用户名
	Namespace                    string                         `json:"namespace,omitempty"`                    // kubeconfig 限制Namespace
	Err                          string                         `json:"err,omitempty"`                          // 连接错误信息
	NodeStatusAggregated         bool                           `json:"nodeStatusAggregated,omitempty"`         // 是否已聚合节点状态
	PodStatusAggregated          bool                           `json:"podStatusAggregated,omitempty"`          // 是否已聚合容器组状态
	StorageClassStatusAggregated bool                           `json:"storageClassStatusAggregated,omitempty"` // 是否已聚合容器组状态
	IsInCluster                  bool                           `json:"isInCluster,omitempty"`                  // 是否为集群内运行获取到的配置
	watchStatus                  map[string]*clusterWatchStatus // watch 类型为key，比如pod,deploy,node,pvc,sc
	restConfig                   *rest.Config                   // 直连rest.Config
	kubeConfig                   []byte                         // 集群配置.kubeconfig原始文件内容
	watchStatusLock              sync.RWMutex                   // watch状态读写锁
}

// 记录每个集群的watch 启动情况
// watch 有多种类型，需要记录
type clusterWatchStatus struct {
	WatchType   string          `json:"watchType,omitempty"`
	Started     bool            `json:"started,omitempty"`
	StartedTime time.Time       `json:"startedTime,omitempty"`
	Watcher     watch.Interface `json:"-"`
}

// SetClusterWatchStarted 设置集群Watch启动状态
func (c *ClusterConfig) SetClusterWatchStarted(watchType string, watcher watch.Interface) {
	c.watchStatusLock.Lock()
	defer c.watchStatusLock.Unlock()
	c.watchStatus[watchType] = &clusterWatchStatus{
		WatchType:   watchType,
		Started:     true,
		StartedTime: time.Now(),
		Watcher:     watcher,
	}
}

// GetClusterWatchStatus 获取集群Watch状态
func (c *ClusterConfig) GetClusterWatchStatus(watchType string) bool {
	watcher := c.watchStatus[watchType]
	if watcher == nil {
		return false
	}
	return watcher.Started
}

// GetClusterID 根据ClusterConfig，按照 文件名+context名称 获取clusterID
func (c *ClusterConfig) GetClusterID() string {
	if c.IsInCluster {
		return "InCluster"
	}
	return fmt.Sprintf("%s/%s", c.FileName, c.ContextName)
}

// ClusterID 根据ClusterConfig，按照 文件名+context名称 获取clusterID
func (c *clusterService) ClusterID(clusterConfig *ClusterConfig) string {
	return clusterConfig.GetClusterID()
}

// GetClusterByID 获取ClusterConfig
func (c *clusterService) GetClusterByID(id string) *ClusterConfig {
	if id == "" {
		return nil
	}
	if id == "InCluster" {
		// InCluster 并没有使用ClusterConfig
		predicate := func(index int, item *ClusterConfig) bool {
			return item.IsInCluster
		}
		if v, ok := slice.FindBy(c.clusterConfigs, predicate); ok {
			return v
		}
	}
	// 解析selectedCluster
	clusterID := strings.Split(id, "/")
	if len(clusterID) != 2 {
		return nil
	}
	fileName := clusterID[0]
	contextName := clusterID[1]
	for _, clusterConfig := range c.clusterConfigs {
		if clusterConfig.FileName == fileName && clusterConfig.ContextName == contextName {
			return clusterConfig
		}
	}
	return nil
}

// IsConnected 判断集群是否连接
func (c *clusterService) IsConnected(selectedCluster string) bool {
	cluster := c.GetClusterByID(selectedCluster)
	connected := cluster.ServerVersion != ""
	return connected
}

func (c *clusterService) DelayStartFunc(f func()) {
	// 延迟启动cron
	// 设置一次性任务的执行时间，例如 5 秒后执行
	schedule := utils.DelayStartSchedule(c.AggregateDelaySeconds)
	cronInstance := cron.New()
	_, err := cronInstance.AddFunc(schedule, f)
	if err != nil {
		klog.Errorf("延迟方法注册失败%v", err)
		return
	}
	cronInstance.Start()
	klog.V(6).Infof("延迟启动cron %ds: %s", c.AggregateDelaySeconds, schedule)
}

// Reconnect 重新连接集群
func (c *clusterService) Reconnect(fileName string, contextName string) {
	// 先清除原来的状态
	for _, clusterConfig := range c.clusterConfigs {
		if clusterConfig.FileName == fileName && clusterConfig.ContextName == contextName {
			clusterConfig.ServerVersion = ""
			clusterConfig.restConfig = nil
			clusterConfig.Err = ""
			c.RegisterCluster(clusterConfig)
		}
	}
}

// Disconnect 断开连接
func (c *clusterService) Disconnect(fileName string, contextName string) {
	// 先清除原来的状态
	cc := c.GetClusterByID(fmt.Sprintf("%s/%s", fileName, contextName))
	cc.ServerVersion = ""
	cc.restConfig = nil
	cc.Err = ""
	for _, v := range cc.watchStatus {
		if v.Watcher != nil {
			v.Watcher.Stop()
			klog.V(6).Infof("%s 停止 Watch  %s", cc.ClusterName, v.WatchType)
		}
	}
}

// Scan 扫描集群
func (c *clusterService) Scan() {
	cfg := flag.Init()
	c.ScanClustersInDir(cfg.KubeConfig)
}

// AllClusters 获取所有集群
func (c *clusterService) AllClusters() []*ClusterConfig {
	return c.clusterConfigs
}

// ConnectedClusters 获取已连接的集群
func (c *clusterService) ConnectedClusters() []*ClusterConfig {
	connected := slice.Filter(c.AllClusters(), func(index int, item *ClusterConfig) bool {
		return item.ServerVersion != ""
	})
	return connected
}

// FirstClusterID 获取第一个集群ID
func (c *clusterService) FirstClusterID() string {
	clusters := c.ConnectedClusters()
	var selectedCluster string
	if len(clusters) > 0 {
		cluster := clusters[0]
		selectedCluster = c.ClusterID(cluster)
	}
	return selectedCluster
}

// RegisterClustersByPath 根据kubeconfig地址注册集群
func (c *clusterService) RegisterClustersByPath(filePath string) {
	// 如果c.clusterConfigs为空，则返回
	if len(c.clusterConfigs) == 0 {
		klog.V(6).Infof("clusterConfigs为空，不进行注册")
		return
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		klog.V(6).Infof("读取文件[%s]失败: %v", filePath, err)
		return
	}

	config, err := clientcmd.Load(content)
	if err != nil {
		klog.V(6).Infof("解析文件[%s]失败: %v", filePath, err)
	}
	contextName := config.CurrentContext

	fileName := filepath.Base(filePath)
	c.Reconnect(fileName, contextName)
}

// ScanClustersInDir 扫描文件夹下的kubeconfig文件，仅扫描形成列表但是不注册集群
func (c *clusterService) ScanClustersInDir(path string) {
	// 1. 通过kubeconfig文件，找到所在目录
	dir := filepath.Dir(path)

	// 2. 通过所在目录，找到同目录下的所有文件
	files, err := os.ReadDir(dir)
	if err != nil {
		klog.V(6).Infof("读取文件夹[%s]失败: %v", dir, err)
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
			klog.V(6).Infof("读取文件[%s]失败: %v", filePath, err)
			continue
		}

		config, err := clientcmd.Load(content)
		if err != nil {
			klog.V(6).Infof("解析文件[%s]失败: %v", filePath, err)
			continue // 解析失败，跳过该文件
		}
		for contextName, _ := range config.Contexts {
			context := config.Contexts[contextName]
			cluster := config.Clusters[context.Cluster]

			clusterConfig := &ClusterConfig{
				FileName:    file.Name(),
				ContextName: contextName,
				UserName:    context.AuthInfo,
				ClusterName: context.Cluster,
				Namespace:   context.Namespace,
				kubeConfig:  content,
				watchStatus: make(map[string]*clusterWatchStatus),
			}
			clusterConfig.Server = cluster.Server
			c.AddToClusterList(clusterConfig)
		}
	}

}

func (c *clusterService) AddToClusterList(clusterConfig *ClusterConfig) {
	// 判断是否已经存在
	if c.GetClusterByID(clusterConfig.GetClusterID()) != nil {
		return
	}
	c.clusterConfigs = append(c.clusterConfigs, clusterConfig)
}

// Deprecated
// RegisterClustersInDir 注册集群,扫描文件夹下的kubeconfig文件，注册集群
func (c *clusterService) RegisterClustersInDir(path string) {
	// 1. 通过kubeconfig文件，找到所在目录
	dir := filepath.Dir(path)

	// 2. 通过所在目录，找到同目录下的所有文件
	files, err := os.ReadDir(dir)
	if err != nil {
		klog.V(6).Infof("读取文件夹[%s]失败: %v", dir, err)
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
			klog.V(6).Infof("读取文件[%s]失败: %v", filePath, err)
			continue
		}

		config, err := clientcmd.Load(content)
		if err != nil {
			klog.V(6).Infof("解析文件[%s]失败: %v", filePath, err)
			continue // 解析失败，跳过该文件
		}
		for contextName, _ := range config.Contexts {
			context := config.Contexts[contextName]
			cluster := config.Clusters[context.Cluster]

			clusterConfig := &ClusterConfig{
				FileName:    file.Name(),
				ContextName: contextName,
				UserName:    context.AuthInfo,
				ClusterName: context.Cluster,
				Namespace:   context.Namespace,
				kubeConfig:  content,
				watchStatus: make(map[string]*clusterWatchStatus),
			}
			clusterConfig.Server = cluster.Server
			c.AddToClusterList(clusterConfig)
		}
	}

	// 注册
	for _, clusterConfig := range c.clusterConfigs {
		// 改为只注册CurrentContext的这个
		c.RegisterCluster(clusterConfig)
	}
	// 打印serverVersion
	for _, clusterConfig := range c.clusterConfigs {
		klog.V(6).Infof("ServerVersion: %s/%s: %s[%s] using user: %s", clusterConfig.FileName, clusterConfig.ContextName, clusterConfig.ServerVersion, clusterConfig.Server, clusterConfig.UserName)
	}
}

// RegisterCluster 从已扫描的集群列表中注册指定的某个集群
func (c *clusterService) RegisterCluster(clusterConfig *ClusterConfig) {
	clusterID := clusterConfig.GetClusterID()
	// 先检查连接是否可以直连，如果可以直连，则直接注册
	if c.CheckCluster(clusterConfig) {
		if clusterConfig.IsInCluster {
			// InCluster模式
			_, err := kom.Clusters().RegisterInCluster()
			if err != nil {
				klog.V(6).Infof("注册集群[%s]失败: %v", clusterID, err)
				return
			}
		} else {
			// 集群外模式
			_, err := kom.Clusters().RegisterByConfigWithID(clusterConfig.restConfig, clusterID)
			if err != nil {
				klog.V(6).Infof("注册集群[%s]失败: %v", clusterID, err)
				return
			}
		}
		klog.V(6).Infof("成功注册集群: %s", clusterID)

	}

}

// CheckCluster 校验集群是否可连接，并更新状态
func (c *clusterService) CheckCluster(config *ClusterConfig) bool {
	var restConfig *rest.Config
	var err error

	if config.IsInCluster {
		// 集群内模式
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			klog.V(6).Infof("获取InCluster的配置失败: %v", err)
			config.Err = err.Error()
			return false
		}
	} else {
		// 集群外模式
		lines := strings.Split(string(config.kubeConfig), "\n")
		for i, line := range lines {
			if strings.HasPrefix(line, "current-context:") {
				lines[i] = "current-context: " + config.ContextName
			}
		}
		bytes := []byte(strings.Join(lines, "\n"))

		restConfig, err = clientcmd.RESTConfigFromKubeConfig(bytes)
		if err != nil {
			klog.V(6).Infof("解析rest.Config错误 %s: %v", config.GetClusterID(), err)
			config.Err = err.Error()
			return false
		}
	}

	// 校验集群是否可连接
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		klog.V(6).Infof("创建clientset失败 %s: %v", config.GetClusterID(), err)
		config.Err = err.Error()
		return false
	}

	// 尝试获取集群版本以验证连接
	info, err := clientset.ServerVersion()
	if err != nil {
		klog.V(6).Infof("连接集群失败 %s: %v", config.GetClusterID(), err)
		config.Err = err.Error()
		return false
	}
	klog.V(6).Infof("成功连接集群 %s", config.GetClusterID())
	// 可以连接的放到数组中记录
	config.ServerVersion = info.GitVersion
	config.restConfig = restConfig
	return true
}

// RegisterInCluster 将InCluster的配置注册到集群列表中
func (c *clusterService) RegisterInCluster() {
	cfg := flag.Init()

	// 获取InCluster的配置
	config, err := rest.InClusterConfig()
	if err != nil {
		klog.V(6).Infof("获取InCluster的配置失败: %v", err)
		cfg.InCluster = false
		return
	}
	cfg.InCluster = true

	// 3. 生成 ClusterConfig
	clusterConfig := &ClusterConfig{
		ClusterName: "kubernetes", // InCluster 模式没有 context, 设定默认名称
		FileName:    "InCluster",
		ContextName: "InCluster",
		Server:      config.Host,
		IsInCluster: true,
		restConfig:  config,
		watchStatus: make(map[string]*clusterWatchStatus),
	}

	c.AddToClusterList(clusterConfig)
	c.RegisterCluster(clusterConfig)
}
