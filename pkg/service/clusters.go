package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/robfig/cron/v3"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/flag"
	"github.com/weibaohui/k8m/pkg/k8sgpt/analysis"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/kom/kom"
	komaws "github.com/weibaohui/kom/kom/aws"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

type clusterService struct {
	clusterConfigs        []*ClusterConfig                    // 文件名+context名称 -> 集群配置
	AggregateDelaySeconds int                                 // 聚合延迟时间
	callbackRegisterFunc  func(cluster *ClusterConfig) func() // 用来注册回调参数的回调方法
	// 心跳管理
	heartbeatCancel           sync.Map // 心跳取消函数，改为sync.Map
	HeartbeatIntervalSeconds  int      // 心跳间隔秒数，默认30
	HeartbeatFailureThreshold int      // 心跳失败阈值，默认3

	// 自动重连管理
	reconnectCancel             sync.Map // 自动重连取消函数，改为sync.Map
	ReconnectMaxIntervalSeconds int      // 自动重连最大退避秒数，默认3600
	MaxRetryAttempts            int      // 最大重试次数，默认100次

}

func newClusterService() *clusterService {
	cfg := flag.Init()
	// Service.ClusterService()使用了init启动，那么会优先于main函数中的执行逻辑（config update from db）
	// 导致return 实例的时候，使用的是cfg中的默认值
	// 因此我们在下面加载下数据库中的配置，确保在后台管理界面中设置的值，是生效的
	_ = ConfigService().UpdateFlagFromDBConfig()
	return &clusterService{
		clusterConfigs:              []*ClusterConfig{},
		AggregateDelaySeconds:       61,
		HeartbeatIntervalSeconds:    cfg.HeartbeatIntervalSeconds,
		HeartbeatFailureThreshold:   cfg.HeartbeatFailureThreshold,
		ReconnectMaxIntervalSeconds: cfg.ReconnectMaxIntervalSeconds,
		MaxRetryAttempts:            cfg.MaxRetryAttempts,
	}
}

func (c *clusterService) UpdateHeartbeatSettings() {
	cfg := flag.Init()

	c.HeartbeatIntervalSeconds = cfg.HeartbeatIntervalSeconds
	c.HeartbeatFailureThreshold = cfg.HeartbeatFailureThreshold
	c.ReconnectMaxIntervalSeconds = cfg.ReconnectMaxIntervalSeconds
	c.MaxRetryAttempts = cfg.MaxRetryAttempts
	klog.V(4).Infof("更新集群心跳和重连配置：心跳间隔 %d 秒，心跳失败阈值 %d，重连最大间隔 %d 秒，最大重试次数 %d",
		c.HeartbeatIntervalSeconds, c.HeartbeatFailureThreshold, c.ReconnectMaxIntervalSeconds, c.MaxRetryAttempts)
}

func (c *clusterService) SetRegisterCallbackFunc(callback func(cluster *ClusterConfig) func()) {
	c.callbackRegisterFunc = callback
}

type ClusterConfig struct {
	ClusterID               string                         `json:"cluster_id,omitempty"`        // 自动生成，不要赋值
	ClusterIDBase64         string                         `json:"cluster_id_base64,omitempty"` // 自动生成，不要赋值
	FileName                string                         `json:"fileName,omitempty"`          // kubeconfig 文件名称
	ContextName             string                         `json:"contextName,omitempty"`       // context名称
	ClusterName             string                         `json:"clusterName,omitempty"`       // 集群名称
	Server                  string                         `json:"server,omitempty"`            // 集群地址
	ServerVersion           string                         `json:"serverVersion,omitempty"`     // 通过这个值来判断集群是否可用
	HeartbeatHistory        []HeartbeatRecord              `json:"heartbeat_history,omitempty"`
	UserName                string                         `json:"userName,omitempty"`                // 用户名
	Namespace               string                         `json:"namespace,omitempty"`               // kubeconfig 限制Namespace
	Err                     string                         `json:"err,omitempty"`                     // 连接错误信息
	NodeStatusAggregated    bool                           `json:"nodeStatusAggregated,omitempty"`    // 是否已聚合节点状态
	PodStatusAggregated     bool                           `json:"podStatusAggregated,omitempty"`     // 是否已聚合容器组状态
	PVCStatusAggregated     bool                           `json:"pvcStatusAggregated,omitempty"`     // 是否已聚合pcv状态
	PVStatusAggregated      bool                           `json:"pvStatusAggregated,omitempty"`      // 是否已聚合pv状态
	IngressStatusAggregated bool                           `json:"ingressStatusAggregated,omitempty"` // 是否已聚合ingress状态
	ClusterConnectStatus    constants.ClusterConnectStatus `json:"clusterConnectStatus,omitempty"`    // 集群连接状态
	IsInCluster             bool                           `json:"isInCluster,omitempty"`             // 是否为集群内运行获取到的配置
	watchStatus             sync.Map                       // watch 类型为key，比如pod,deploy,node,pvc,sc
	restConfig              *rest.Config                   // 直连rest.Config
	kubeConfig              []byte                         // 集群配置.kubeconfig原始文件内容
	Source                  ClusterConfigSource            `json:"source,omitempty"`                 // 配置文件来源
	K8sGPTProblemsCount     int                            `json:"k8s_gpt_problems_count,omitempty"` // k8sGPT 扫描结果
	K8sGPTProblemsResult    *analysis.ResultWithStatus     `json:"k8s_gpt_problems,omitempty"`       // k8sGPT 扫描结果
	NotAfter                *time.Time                     `json:"not_after,omitempty"`
	AWSConfig               *komaws.EKSAuthConfig          `json:"aws_config,omitempty"` // AWS EKS配置信息
	IsAWSEKS                bool                           `json:"is_aws_eks,omitempty"` // 标识是否为AWS EKS集群

	// kom 集群注册配置项
	DBID     uint    `json:"id,omitempty"`        // 数据库ID
	ProxyURL string  `json:"proxy_url,omitempty"` // HTTP 代理，例如 http://127.0.0.1:7890
	Timeout  int     `json:"timeout,omitempty"`   // 请求超时时间，单位为秒，默认为 30 秒
	QPS      float32 `json:"qps,omitempty"`       // 每秒查询数限制，默认为 200
	Burst    int     `json:"burst,omitempty"`     // 突发请求数限制，默认为 2000
}
type ClusterConfigSource string

var ClusterConfigSourceFile ClusterConfigSource = "File"
var ClusterConfigSourceDB ClusterConfigSource = "DB"
var ClusterConfigSourceInCluster ClusterConfigSource = "InCluster"
var ClusterConfigSourceAWS ClusterConfigSource = "AWS"

// 记录每个集群的watch 启动情况
// watch 有多种类型，需要记录
type clusterWatchStatus struct {
	WatchType   string          `json:"watchType,omitempty"`
	Started     bool            `json:"started,omitempty"`
	StartedTime time.Time       `json:"startedTime,omitempty"`
	Watcher     watch.Interface `json:"-"`
}

// HeartbeatRecord 心跳结果记录条目
// 中文说明：index 为在当前窗口中的位置（1..N），success 表示本次心跳是否成功，time 为发生时间（本地时区）
type HeartbeatRecord struct {
	Index   int    `json:"index"`
	Success bool   `json:"success"`
	Time    string `json:"time"`
}

// appendHeartbeatRecord 追加一条心跳记录，并裁剪为阈值长度
// 中文函数注释：将成功/失败结果与本地时间写入 HeartbeatHistory，保持长度不超过阈值
func (c *clusterService) appendHeartbeatRecord(cluster *ClusterConfig, success bool, ts time.Time) {
	if cluster == nil {
		return
	}
	if cluster.HeartbeatHistory == nil {
		cluster.HeartbeatHistory = make([]HeartbeatRecord, 0)
	}
	// 追加一条记录
	rec := HeartbeatRecord{
		Success: success,
		Time:    ts.Local().Format("2006-01-02 15:04:05"),
	}
	cluster.HeartbeatHistory = append(cluster.HeartbeatHistory, rec)
	// 裁剪为最近阈值条目
	threshold := c.HeartbeatFailureThreshold
	if threshold <= 0 {
		threshold = 3 // 兜底：默认 3 次
	}
	if len(cluster.HeartbeatHistory) > threshold {
		cluster.HeartbeatHistory = cluster.HeartbeatHistory[len(cluster.HeartbeatHistory)-threshold:]
	}
	// 重新标注窗口内的序号为 1..N
	for i := range cluster.HeartbeatHistory {
		cluster.HeartbeatHistory[i].Index = i + 1
	}
}

// SetClusterWatchStarted 设置集群Watch启动状态
func (c *ClusterConfig) SetClusterWatchStarted(watchType string, watcher watch.Interface) {
	c.watchStatus.Store(watchType, &clusterWatchStatus{
		WatchType:   watchType,
		Started:     true,
		StartedTime: time.Now(),
		Watcher:     watcher,
	})
}

// GetClusterWatchStatus 获取集群Watch状态
func (c *ClusterConfig) GetClusterWatchStatus(watchType string) bool {
	if value, ok := c.watchStatus.Load(watchType); ok {
		if watcher, ok := value.(*clusterWatchStatus); ok {
			return watcher.Started
		}
	}
	return false
}
func (c *ClusterConfig) GetKubeconfig() string {
	return string(c.kubeConfig)
}

// GetClusterID 根据ClusterConfig，按照 文件名+context名称 获取clusterID
func (c *ClusterConfig) GetClusterID() string {

	// 原有逻辑
	id := fmt.Sprintf("%s/%s", c.FileName, c.ContextName)
	if c.IsInCluster {
		id = "InCluster"
	}
	if id == "InCluster/InCluster" {
		id = "InCluster"
	}
	c.ClusterID = id
	c.ClusterIDBase64 = base64.StdEncoding.EncodeToString([]byte(id))
	return id
}

func (c *ClusterConfig) GetRestConfig() *rest.Config {
	return c.restConfig
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
	// 第一个/前面的字符是fileName。其他的都是contextName
	// 有可能会出现多个/，如config/aws-x-x-x/demo
	slashIndex := strings.Index(id, "/")
	if slashIndex == -1 {
		return nil
	}
	fileName := id[:slashIndex]
	contextName := id[slashIndex+1:]
	for _, clusterConfig := range c.clusterConfigs {
		if clusterConfig.FileName == fileName && clusterConfig.ContextName == contextName {
			return clusterConfig
		}
	}
	return nil
}

// GetCertificateExpiry 获取集群证书的过期时间
func (c *ClusterConfig) GetCertificateExpiry() time.Time {
	// 检查 kubeConfig 是否为空
	if len(c.kubeConfig) == 0 {
		klog.V(8).Infof("设置NotAfter, 集群[%s] kubeConfig为空", c.ClusterID)
		return time.Time{}
	}

	config, err := clientcmd.Load(c.kubeConfig)
	if err != nil {
		klog.V(8).Infof("设置NotAfter, 解析文件[%s]失败: %v", c.ClusterID, err)
		return time.Time{}
	}

	// 检查 config 是否为空
	if config == nil {
		klog.V(8).Infof("设置NotAfter, 集群[%s] config为空", c.ClusterID)
		return time.Time{}
	}

	// 检查 CurrentContext 是否为空
	if config.CurrentContext == "" {
		klog.V(8).Infof("设置NotAfter, 集群[%s] CurrentContext为空", c.ClusterID)
		return time.Time{}
	}

	// 检查 Contexts 是否为空
	if config.Contexts == nil {
		klog.V(8).Infof("设置NotAfter, 集群[%s] Contexts为空", c.ClusterID)
		return time.Time{}
	}

	// 检查当前 context 是否存在
	currentContext, contextExists := config.Contexts[config.CurrentContext]
	if !contextExists || currentContext == nil {
		klog.V(8).Infof("设置NotAfter, 集群[%s] 当前context[%s]不存在", c.ClusterID, config.CurrentContext)
		return time.Time{}
	}

	// 检查 AuthInfos 是否为空
	if config.AuthInfos == nil {
		klog.V(8).Infof("设置NotAfter, 集群[%s] AuthInfos为空", c.ClusterID)
		return time.Time{}
	}

	// 检查 AuthInfo 名称是否为空
	if currentContext.AuthInfo == "" {
		klog.V(8).Infof("设置NotAfter, 集群[%s] AuthInfo名称为空", c.ClusterID)
		return time.Time{}
	}

	// 获取 authInfo
	authInfo, exists := config.AuthInfos[currentContext.AuthInfo]
	if !exists || authInfo == nil {
		klog.V(8).Infof("设置NotAfter, 集群[%s] authInfo[%s]不存在", c.ClusterID, currentContext.AuthInfo)
		return time.Time{}
	}

	// 检查证书数据是否为空
	if len(authInfo.ClientCertificateData) == 0 {
		klog.V(8).Infof("设置NotAfter, 集群[%s] ClientCertificateData为空", c.ClusterID)
		return time.Time{}
	}

	// 解析证书
	cert, err := utils.ParseCertificate(authInfo.ClientCertificateData)
	if err != nil {
		klog.V(8).Infof("设置NotAfter, 集群[%s]解析证书失败: %v", c.ClusterID, err)
		return time.Time{}
	}

	// 检查证书是否为空
	if cert == nil {
		klog.V(8).Infof("设置NotAfter, 集群[%s]解析出的证书为空", c.ClusterID)
		return time.Time{}
	}

	return cert.NotAfter.Local()
}

// IsConnected 判断集群是否连接
func (c *clusterService) IsConnected(selectedCluster string) bool {
	cluster := c.GetClusterByID(selectedCluster)
	if cluster == nil {
		return false
	}
	if cluster.ClusterConnectStatus == "" {
		return false
	}
	// 加强语义：必须已成功获取过 ServerVersion 才认为“已连接”
	connected := cluster.ClusterConnectStatus == constants.ClusterConnectStatusConnected && cluster.ServerVersion != ""
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

// Connect 重新连接集群
// 中文函数注释：尝试连接指定集群。仅在集群不在"已连接"或"连接中"状态时执行实际连接操作。
// 连接过程会先清理旧的连接资源，再尝试重新注册。注意此函数不负责重试逻辑，重试由上层的自动重连循环处理。
func (c *clusterService) Connect(clusterID string) {
	klog.V(4).Infof("连接集群 %s 开始", clusterID)

	cc := c.GetClusterByID(clusterID)
	if cc == nil {
		klog.V(4).Infof("集群[%s] 不存在，无法连接", clusterID)
		return
	}

	// 只有当集群不是"已连接"或"连接中"状态时，才执行连接操作
	if !(cc.ClusterConnectStatus == constants.ClusterConnectStatusConnected ||
		cc.ClusterConnectStatus == constants.ClusterConnectStatusConnecting) {
		klog.V(4).Infof("集群[%s] 当前状态为[%s]，开始连接操作", clusterID, cc.ClusterConnectStatus)

		// 清理连接资源，但不停止自动重连（第二个参数为 false）
		c.disconnectWithOption(clusterID, false)

		// 更新状态为"连接中"
		cc.ClusterConnectStatus = constants.ClusterConnectStatusConnecting

		// 尝试注册集群
		_, err := c.RegisterCluster(cc)
		if err != nil {
			klog.V(4).Infof("集群[%s] 连接失败: %v，等待下一次重试", clusterID, err)
			// 注意：这里不设置状态，让上层重试循环继续工作
			if cc.ClusterConnectStatus == constants.ClusterConnectStatusConnecting {
				cc.ClusterConnectStatus = constants.ClusterConnectStatusFailed
			}
		} else {
			klog.V(4).Infof("集群[%s] 连接成功", clusterID)
		}
	} else {
		klog.V(4).Infof("集群[%s] 当前状态为[%s]，跳过连接操作", clusterID, cc.ClusterConnectStatus)
	}

	klog.V(4).Infof("连接集群 %s 完毕", clusterID)
}

// disconnectWithOption 断开连接（可选是否停止自动重连）
// 中文函数注释：幂等清理指定集群的连接状态与资源；当 stopReconnect 为 true 时，连同自动重连循环一并停止，
// 为 false 时仅做资源清理以便在自动重连循环内使用，避免自我取消导致循环中断。
func (c *clusterService) disconnectWithOption(clusterID string, stopReconnect bool) {
	klog.V(6).Infof("Disconnect 开始清理集群 %s 原始信息（停止自动重连：%t）", clusterID, stopReconnect)

	cc := c.GetClusterByID(clusterID)
	if cc == nil {
		return
	}
	// 停止心跳
	c.StopHeartbeat(clusterID)
	// 根据需要停止自动重连
	if stopReconnect {
		c.StopReconnect(clusterID)
	}
	// 清理本地状态
	cc.ServerVersion = ""
	cc.restConfig = nil
	cc.Err = ""
	cc.ClusterConnectStatus = constants.ClusterConnectStatusDisconnected
	cc.watchStatus.Range(func(key, value interface{}) bool {
		if v, ok := value.(*clusterWatchStatus); ok {
			if v.Watcher != nil {
				v.Watcher.Stop()
				klog.V(6).Infof("%s 停止 Watch  %s", cc.ClusterName, v.WatchType)
			}
		}
		return true
	})
	// 从kom解除
	kom.Clusters().RemoveClusterById(clusterID)
	klog.V(6).Infof("Disconnect 完成清理集群 %s", clusterID)
}

// Disconnect 断开连接
// 中文函数注释：幂等清理指定集群的连接状态与资源，并停止自动重连；用于外部显式断开场景。
func (c *clusterService) Disconnect(clusterID string) {
	c.disconnectWithOption(clusterID, true)
}

// Scan 扫描集群
func (c *clusterService) Scan() {
	cfg := flag.Init()
	c.ScanClustersInDir(cfg.KubeConfig)

	c.ScanClustersInDB()
}

// AllClusters 获取所有集群
func (c *clusterService) AllClusters() []*ClusterConfig {
	return c.clusterConfigs
}

// ConnectedClusters 获取已连接的集群
func (c *clusterService) ConnectedClusters() []*ClusterConfig {
	connected := slice.Filter(c.AllClusters(), func(index int, item *ClusterConfig) bool {
		return item.ClusterConnectStatus == constants.ClusterConnectStatusConnected
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
	// 处理路径中的 ~ 符号
	expandedPath, err := utils.ExpandHomePath(filePath)
	if err != nil {
		klog.V(6).Infof("展开路径失败: %v", err)
		return
	}
	filePath = expandedPath
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
	c.Connect(fmt.Sprintf("%s/%s", fileName, contextName))
}

// ScanClustersInDir 扫描文件夹下的kubeconfig文件，仅扫描形成列表但是不注册集群
func (c *clusterService) ScanClustersInDir(path string) {
	// 处理路径中的 ~ 符号
	expandedPath, err := utils.ExpandHomePath(path)
	if err != nil {
		klog.V(6).Infof("展开路径失败: %v", err)
		return
	}
	path = expandedPath

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
		for contextName := range config.Contexts {
			context := config.Contexts[contextName]
			cluster := config.Clusters[context.Cluster]

			clusterConfig := &ClusterConfig{
				FileName:             file.Name(),
				ContextName:          contextName,
				ClusterID:            fmt.Sprintf("%s/%s", file.Name(), contextName),
				UserName:             context.AuthInfo,
				ClusterName:          context.Cluster,
				Namespace:            context.Namespace,
				kubeConfig:           content,
				ClusterConnectStatus: constants.ClusterConnectStatusDisconnected,
				Source:               ClusterConfigSourceFile,
			}
			clusterConfig.Server = cluster.Server
			c.AddToClusterList(clusterConfig)
		}
	}

}
func (c *clusterService) ScanClustersInDB() {

	var list []*models.KubeConfig
	err := dao.DB().Model(&models.KubeConfig{}).Find(&list).Error
	if err != nil {
		klog.Errorf("查询集群失败: %v", err)
		return
	}

	for i, cc := range c.clusterConfigs {
		if cc.Source == ClusterConfigSourceDB || cc.Source == ClusterConfigSourceAWS {
			// 查一下list中是否存在
			filter := slice.Filter(list, func(index int, item *models.KubeConfig) bool {
				if item.Server == cc.Server && item.User == cc.UserName && item.Cluster == cc.ClusterName {
					return true
				}
				return false
			})
			if len(filter) == 0 {
				// 在数据库中也不存在
				// 从list中删除
				// 删除前先断开连接，避免watcher泄露
				c.Disconnect(cc.ClusterID)
				c.clusterConfigs = slice.DeleteAt(c.clusterConfigs, i)
			}
		}
	}

	// 2. 处理数据库中的配置
	for _, item := range list {
		config, err := clientcmd.Load([]byte(item.Content))
		if err != nil {
			klog.V(6).Infof("解析集群 [%s]失败: %v", item.Server, err)
			continue
		}

		// 检查每个context
		for contextName := range config.Contexts {
			context := config.Contexts[contextName]
			cluster := config.Clusters[context.Cluster]

			if context.AuthInfo == item.User {
				// 检查是否已存在该配置
				exists := false
				for _, cc := range c.clusterConfigs {
					if (cc.FileName == string(ClusterConfigSourceAWS)) && cc.Server == cluster.Server && cc.ContextName == contextName {
						exists = true
						break
					}
				}

				// 如果不存在，添加新配置
				if !exists {
					clusterConfig := &ClusterConfig{
						ContextName:          contextName,
						UserName:             context.AuthInfo,
						ClusterName:          context.Cluster,
						Namespace:            context.Namespace,
						kubeConfig:           []byte(item.Content),
						ClusterConnectStatus: constants.ClusterConnectStatusDisconnected,
						Server:               cluster.Server,
						Source:               ClusterConfigSourceDB,
						// 从数据库中读取 kom 配置项
						ProxyURL: item.ProxyURL,
						Timeout:  item.Timeout,
						QPS:      item.QPS,
						Burst:    item.Burst,
						DBID:     item.ID,
					}
					if item.DisplayName != "" {
						clusterConfig.FileName = item.DisplayName
					} else {
						clusterConfig.FileName = fmt.Sprintf("%d-%s", item.ID, contextName)
					}

					// aws 单独处理
					if item.IsAWSEKS {
						clusterConfig.Source = ClusterConfigSourceAWS
						eksConfig := &komaws.EKSAuthConfig{
							AccessKey:       item.AccessKey,
							SecretAccessKey: item.SecretAccessKey,
							Region:          item.Region,
							ClusterName:     item.ClusterName,
						}
						clusterConfig.AWSConfig = eksConfig
						clusterConfig.IsAWSEKS = true
						clusterConfig.FileName = string(ClusterConfigSourceAWS)
					}
					clusterConfig.Server = cluster.Server
					c.AddToClusterList(clusterConfig)
				}

			}

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
		for contextName := range config.Contexts {
			context := config.Contexts[contextName]
			cluster := config.Clusters[context.Cluster]

			clusterConfig := &ClusterConfig{
				FileName:             file.Name(),
				ContextName:          contextName,
				UserName:             context.AuthInfo,
				ClusterName:          context.Cluster,
				Namespace:            context.Namespace,
				kubeConfig:           content,
				ClusterConnectStatus: constants.ClusterConnectStatusDisconnected,
				Source:               ClusterConfigSourceFile,
			}
			clusterConfig.Server = cluster.Server
			c.AddToClusterList(clusterConfig)
		}
	}

	// 注册
	for _, clusterConfig := range c.clusterConfigs {
		// 改为只注册CurrentContext的这个
		_, _ = c.RegisterCluster(clusterConfig)
	}
	// 打印serverVersion
	for _, clusterConfig := range c.clusterConfigs {
		klog.V(6).Infof("ServerVersion: %s/%s: %s[%s] using user: %s", clusterConfig.FileName, clusterConfig.ContextName, clusterConfig.ServerVersion, clusterConfig.Server, clusterConfig.UserName)
	}
}

// RegisterCluster 从已扫描的集群列表中注册指定的某个集群
// 负责单次注册尝试，不处理重试逻辑。连接失败时返回错误，
// 由调用方（Connect 方法）负责设置状态为 Failed 以配合上层的重试机制。
// 致命错误（如配置错误）会直接在本方法中设置为失败状态。
func (c *clusterService) RegisterCluster(clusterConfig *ClusterConfig) (bool, error) {
	if clusterConfig == nil {
		return false, fmt.Errorf("集群配置为空")
	}

	clusterID := clusterConfig.GetClusterID()
	klog.V(6).Infof("开始注册集群 %s [来源：%s]", clusterID, clusterConfig.Source)

	// AWS EKS 集群处理
	if clusterConfig.IsAWSEKS {
		if clusterConfig.AWSConfig == nil {
			err := fmt.Errorf("AWS EKS 集群[%s]缺少 AWSConfig", clusterID)
			clusterConfig.ClusterConnectStatus = constants.ClusterConnectStatusFailed // 配置错误属于致命错误
			clusterConfig.Err = err.Error()
			return false, err
		}

		// 构建注册选项
		opts := c.buildRegisterOptions(clusterConfig)
		// 使用带 ID 的注册确保幂等
		if _, err := kom.Clusters().RegisterAWSClusterWithID(clusterConfig.AWSConfig, clusterID, opts...); err != nil {
			klog.V(4).Infof("注册 AWS 集群[%s]失败: %v", clusterID, err)
			clusterConfig.Err = err.Error()
			return false, err // 保持"连接中"状态
		}

		// 注册成功后校验连通性
		if err := c.LoadRestConfig(clusterConfig); err != nil {
			clusterConfig.Err = err.Error()
			return false, err // 保持"连接中"状态
		}
	} else {
		// 非 AWS 集群处理
		if err := c.LoadRestConfig(clusterConfig); err != nil {
			clusterConfig.Err = err.Error()
			return false, err // 保持"连接中"状态
		}

		if clusterConfig.IsInCluster {
			// InCluster 模式
			if _, err := kom.Clusters().RegisterInCluster(); err != nil {
				klog.V(4).Infof("注册集群[%s]失败: %v", clusterID, err)
				clusterConfig.Err = err.Error()
				return false, err // 保持"连接中"状态
			}
		} else {
			// 集群外模式
			opts := c.buildRegisterOptions(clusterConfig)
			if _, err := kom.Clusters().RegisterByConfigWithID(clusterConfig.restConfig, clusterID, opts...); err != nil {
				klog.V(4).Infof("注册集群[%s]失败: %v", clusterID, err)
				clusterConfig.Err = err.Error()
				return false, err // 保持"连接中"状态
			}
		}
	}

	// 所有注册步骤成功完成，更新状态
	klog.V(4).Infof("成功注册集群: %s [%s]", clusterID, clusterConfig.Server)
	clusterConfig.ClusterConnectStatus = constants.ClusterConnectStatusConnected
	clusterConfig.Err = "" // 清除错误信息

	// 启动心跳监测
	c.StartHeartbeat(clusterID)

	// 执行回调注册
	if c.callbackRegisterFunc != nil {
		c.callbackRegisterFunc(clusterConfig)
	}

	return true, nil
}

// LoadRestConfig 校验集群是否可连接，并更新状态
func (c *clusterService) LoadRestConfig(config *ClusterConfig) error {
	var restConfig *rest.Config
	var err error
	if config.IsInCluster {
		// 集群内模式（严格处理错误，避免误判为连通）
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			klog.V(6).Infof("加载 InCluster 配置失败 %s: %v", config.GetClusterID(), err)
			config.Err = err.Error()
			config.ClusterConnectStatus = constants.ClusterConnectStatusFailed
			return err
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

	}
	config.restConfig = restConfig

	// 校验集群是否可连接
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		klog.V(6).Infof("创建clientset失败 %s: %v", config.GetClusterID(), err)
		config.Err = err.Error()
		config.ClusterConnectStatus = constants.ClusterConnectStatusFailed
		return err
	}

	// 尝试获取集群版本以验证连接
	info, err := clientset.ServerVersion()
	if err != nil {
		klog.V(6).Infof("连接集群失败 %s: %v", config.GetClusterID(), err)
		config.Err = err.Error()
		config.ClusterConnectStatus = constants.ClusterConnectStatusFailed
		return err
	}
	klog.V(6).Infof("LoadRestConfig 获取集群 版本成功 %s", config.GetClusterID())
	config.ServerVersion = info.GitVersion
	return err
}

// RegisterInCluster 将InCluster的配置注册到集群列表中
func (c *clusterService) RegisterInCluster() {

	// 获取InCluster的配置
	config, err := rest.InClusterConfig()
	if err != nil {
		cfg := flag.Init()
		cfg.InCluster = false
		klog.Errorf("获取InCluster的配置失败,InCluster模式关闭.错误：%v", err)
		return
	}

	// 3. 生成 ClusterConfig
	clusterConfig := &ClusterConfig{
		ClusterName:          "kubernetes", // InCluster 模式没有 context, 设定默认名称
		FileName:             "InCluster",
		ContextName:          "InCluster",
		ClusterID:            "InCluster",
		Server:               config.Host,
		IsInCluster:          true,
		restConfig:           config,
		ClusterConnectStatus: constants.ClusterConnectStatusDisconnected,
		Source:               ClusterConfigSourceInCluster,
	}

	c.AddToClusterList(clusterConfig)
	_, _ = c.RegisterCluster(clusterConfig)
}

func (c *ClusterConfig) SetClusterScanStatus(result *analysis.ResultWithStatus) {
	c.K8sGPTProblemsCount = result.Problems
	c.K8sGPTProblemsResult = result

}
func (c *ClusterConfig) GetClusterScanResult() *analysis.ResultWithStatus {

	return c.K8sGPTProblemsResult

}

// =========================== 集群服务增强方法 ===========================

// validateAWSConfig 验证AWS配置
func (c *clusterService) validateAWSConfig(config *komaws.EKSAuthConfig) error {
	if config == nil {
		return fmt.Errorf("AWS配置不能为空")
	}

	if config.AccessKey == "" {
		return fmt.Errorf("AWS Access Key不能为空")
	}

	if config.SecretAccessKey == "" {
		return fmt.Errorf("AWS Secret Access Key不能为空")
	}

	if config.Region == "" {
		return fmt.Errorf("AWS区域不能为空")
	}

	if config.ClusterName == "" {
		return fmt.Errorf("EKS集群名称不能为空")
	}

	return nil
}

// RegisterAWSEKSCluster 注册AWS EKS集群
func (c *clusterService) RegisterAWSEKSCluster(config *komaws.EKSAuthConfig) (*ClusterConfig, error) {
	// 参数验证
	if err := c.validateAWSConfig(config); err != nil {
		return nil, fmt.Errorf("AWS配置验证失败: %w", err)
	}

	// 将项目内部的AWSEKSConfig转换为kom库要求的kom.aws.EKSAuthConfig
	eksAuthConfig := &komaws.EKSAuthConfig{
		AccessKey:       config.AccessKey,
		SecretAccessKey: config.SecretAccessKey,
		Region:          config.Region,
		ClusterName:     config.ClusterName,
	}

	kg := komaws.NewKubeconfigGenerator()
	content, err := kg.GenerateFromAWS(eksAuthConfig)
	if err != nil {
		return nil, fmt.Errorf("生成AWS EKS集群配置文件失败: %w", err)
	}
	kubeconfig, err := clientcmd.Load([]byte(content))
	if err != nil {
		klog.V(6).Infof("解析 AWS EKS集群kubeconfig配置失败: %v", err)
		return nil, fmt.Errorf("生成AWS EKS集群配置文件失败: %w", err)
	}

	// 只取第一个context
	var contextName string
	var clusterConfig *ClusterConfig
	for name := range kubeconfig.Contexts {
		contextName = name
		break
	}
	if contextName != "" {
		context := kubeconfig.Contexts[contextName]
		cluster := kubeconfig.Clusters[context.Cluster]

		clusterConfig = &ClusterConfig{
			FileName:             string(ClusterConfigSourceAWS),
			ContextName:          contextName,
			ClusterName:          context.Cluster,
			Namespace:            context.Namespace,
			Server:               cluster.Server,
			kubeConfig:           []byte(content),
			ClusterConnectStatus: constants.ClusterConnectStatusDisconnected,
			Source:               ClusterConfigSourceAWS,
			IsAWSEKS:             true,
			AWSConfig:            config,
		}
		clusterID := clusterConfig.GetClusterID()

		// 构建注册选项（使用默认值，因为此方法没有传入 kom 配置项）
		opts := c.buildRegisterOptions(clusterConfig)
		// 使用kom统一的AWS EKS集群注册方法
		_, err = kom.Clusters().RegisterAWSClusterWithID(eksAuthConfig, clusterID, opts...)
		if err != nil {
			clusterConfig.ClusterConnectStatus = constants.ClusterConnectStatusFailed
			clusterConfig.Err = err.Error()
			return nil, fmt.Errorf("注册AWS EKS集群失败: %w", err)
		}

		// 添加到集群列表
		c.AddToClusterList(clusterConfig)
		
		// 校验连通性并设置ServerVersion
		if err := c.LoadRestConfig(clusterConfig); err != nil {
			clusterConfig.ClusterConnectStatus = constants.ClusterConnectStatusFailed
			clusterConfig.Err = err.Error()
			return nil, fmt.Errorf("AWS EKS集群连通性校验失败: %w", err)
		}
		
		clusterConfig.ClusterConnectStatus = constants.ClusterConnectStatusConnected
		clusterConfig.Err = ""
		klog.V(4).Infof("成功注册AWS EKS集群: %s [%s]", config.ClusterName, clusterID)
	}

	return clusterConfig, nil
}

// buildRegisterOptions 根据 ClusterConfig 构建 kom 注册选项
func (c *clusterService) buildRegisterOptions(clusterConfig *ClusterConfig) []kom.RegisterOption {
	klog.V(6).Infof("开始构建集群 %s 的注册选项配置", clusterConfig.ClusterID)
	var opts []kom.RegisterOption

	// 设置代理
	if clusterConfig.ProxyURL != "" {
		klog.V(6).Infof("设置集群 %s 代理URL: %s", clusterConfig.ClusterID, clusterConfig.ProxyURL)
		opts = append(opts, kom.RegisterProxyURL(clusterConfig.ProxyURL))
	} else {
		klog.V(6).Infof("集群 %s 未设置代理URL", clusterConfig.ClusterID)
	}

	// 设置超时时间
	if clusterConfig.Timeout > 0 {
		klog.V(6).Infof("设置集群 %s 超时时间: %d 秒", clusterConfig.ClusterID, clusterConfig.Timeout)
		opts = append(opts, kom.RegisterTimeout(time.Duration(clusterConfig.Timeout)*time.Second))
	} else {
		klog.V(6).Infof("集群 %s 使用默认超时时间", clusterConfig.ClusterID)
	}

	// 设置 QPS
	if clusterConfig.QPS > 0 {
		klog.V(6).Infof("设置集群 %s QPS 限制: %.2f", clusterConfig.ClusterID, clusterConfig.QPS)
		opts = append(opts, kom.RegisterQPS(clusterConfig.QPS))
	} else {
		klog.V(6).Infof("集群 %s 使用默认 QPS 限制", clusterConfig.ClusterID)
	}

	// 设置 Burst
	if clusterConfig.Burst > 0 {
		klog.V(6).Infof("设置集群 %s Burst 限制: %d", clusterConfig.ClusterID, clusterConfig.Burst)
		opts = append(opts, kom.RegisterBurst(clusterConfig.Burst))
	} else {
		klog.V(6).Infof("集群 %s 使用默认 Burst 限制", clusterConfig.ClusterID)
	}

	klog.V(6).Infof("集群 %s 注册选项配置完成，共配置 %d 个选项", clusterConfig.ClusterID, len(opts))
	return opts
}

// UpdateClusterConfig 更新已加载集群的配置参数
// @Description 根据数据库ID更新已加载集群的ProxyURL、Timeout、QPS、Burst配置，并重新注册已连接的集群
// @Param dbID 数据库中的集群配置ID
// @Param proxyURL HTTP代理URL
// @Param timeout 请求超时时间（秒）
// @Param qps 每秒查询数限制
// @Param burst 突发请求数限制
func (c *clusterService) UpdateClusterConfig(dbID uint, proxyURL string, timeout int, qps float32, burst int) error {
	klog.V(6).Infof("开始更新集群配置，数据库ID: %d", dbID)

	// 查找对应的集群配置
	var targetCluster *ClusterConfig
	for _, cluster := range c.clusterConfigs {
		if cluster.DBID == dbID {
			targetCluster = cluster
			break
		}
	}

	if targetCluster == nil {
		klog.V(4).Infof("未找到数据库ID为 %d 的集群配置", dbID)
		return fmt.Errorf("未找到数据库ID为 %d 的集群配置", dbID)
	}

	klog.V(6).Infof("找到集群配置: %s [%s]", targetCluster.ClusterID, targetCluster.Server)

	// 记录原始配置用于日志
	oldProxyURL := targetCluster.ProxyURL
	oldTimeout := targetCluster.Timeout
	oldQPS := targetCluster.QPS
	oldBurst := targetCluster.Burst

	// 更新配置参数
	targetCluster.ProxyURL = proxyURL
	targetCluster.Timeout = timeout
	targetCluster.QPS = qps
	targetCluster.Burst = burst

	klog.V(6).Infof("集群 %s 配置更新: ProxyURL [%s->%s], Timeout [%d->%d], QPS [%.2f->%.2f], Burst [%d->%d]",
		targetCluster.ClusterID, oldProxyURL, proxyURL, oldTimeout, timeout, oldQPS, qps, oldBurst, burst)

	// 如果集群已连接，需要重新注册以应用新配置
	if targetCluster.ClusterConnectStatus == constants.ClusterConnectStatusConnected {
		klog.V(6).Infof("集群 %s 已连接，开始重新注册以应用新配置", targetCluster.ClusterID)

		// 重新连接，这会使用新的配置参数
		go func() {
			time.Sleep(200 * time.Millisecond) // 稍微延迟一下再重连
			c.Connect(targetCluster.ClusterID)
		}()

		klog.V(4).Infof("集群 %s 配置更新完成，已启动重新连接", targetCluster.ClusterID)
	} else {
		klog.V(6).Infof("集群 %s 未连接，配置更新完成，下次连接时将使用新配置", targetCluster.ClusterID)
	}

	return nil
}

// StartHeartbeat 启动心跳任务
// @Description 周期性检测集群连通性并记录心跳历史；当心跳失败次数达到阈值时，自动取消当前心跳、清理历史并执行重连。
// @Param clusterID 集群ID
func (c *clusterService) StartHeartbeat(clusterID string) {
	// 初始化心跳配置默认值
	if c.HeartbeatIntervalSeconds <= 0 {
		c.HeartbeatIntervalSeconds = 30
	}
	if c.HeartbeatFailureThreshold <= 0 {
		c.HeartbeatFailureThreshold = 3
	}

	// 如果已有心跳，先停止
	if cancelInterface, ok := c.heartbeatCancel.Load(clusterID); ok {
		if cancel, ok := cancelInterface.(context.CancelFunc); ok && cancel != nil {
			cancel()
		}
		c.heartbeatCancel.Delete(clusterID)
	}

	cluster := c.GetClusterByID(clusterID)
	if cluster == nil {
		klog.V(6).Infof("启动心跳失败：未找到集群 %s", clusterID)
		return
	}
	// 仅在已连接时启动心跳
	if cluster.ClusterConnectStatus != constants.ClusterConnectStatusConnected {
		klog.V(6).Infof("集群 %s 非已连接状态，心跳不启动", clusterID)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	c.heartbeatCancel.Store(clusterID, cancel)

	interval := time.Duration(c.HeartbeatIntervalSeconds) * time.Second
	ticker := time.NewTicker(interval)
	klog.V(6).Infof("集群 %s 心跳启动，间隔 %ds，失败阈值 %d，自动重连最大退避秒数 %ds", clusterID, c.HeartbeatIntervalSeconds, c.HeartbeatFailureThreshold, c.ReconnectMaxIntervalSeconds)

	go func() {
		defer ticker.Stop()
		failureCount := 0
		for {
			select {
			case <-ctx.Done():
				klog.V(6).Infof("集群 %s 心跳已停止", clusterID)
				return
			case <-ticker.C:
				// 若集群不再是已连接状态，则停止心跳
				if cluster.ClusterConnectStatus != constants.ClusterConnectStatusConnected {
					klog.V(6).Infof("集群 %s 心跳检测：状态已非已连接，停止心跳", clusterID)

					cancel()
					return
				}
				// restConfig 必须存在
				if cluster.restConfig == nil {
					failureCount++
					klog.V(6).Infof("集群 %s 心跳检测失败：restConfig 不存在（累计失败 %d）", clusterID, failureCount)
					// 记录本次心跳失败
					c.appendHeartbeatRecord(cluster, false, time.Now())
				} else {
					clientset, err := kubernetes.NewForConfig(cluster.restConfig)
					if err != nil {
						failureCount++
						klog.V(6).Infof("集群 %s 创建 clientset 失败：%v（累计失败 %d）", clusterID, err, failureCount)
						// 记录本次心跳失败
						c.appendHeartbeatRecord(cluster, false, time.Now())
					} else {
						sv, err := clientset.Discovery().ServerVersion()
						if err != nil {
							failureCount++
							klog.V(6).Infof("集群 %s 心跳检测读取版本失败：%v（累计失败 %d）", clusterID, err, failureCount)
							cluster.Err = err.Error()
							// 记录本次心跳失败
							c.appendHeartbeatRecord(cluster, false, time.Now())
						} else {
							// 成功，重置失败计数并同步版本
							failureCount = 0
							if sv != nil {
								cluster.ServerVersion = sv.GitVersion
							}
							klog.V(6).Infof("集群 %s 心跳检测成功，当前版本：%s", clusterID, cluster.ServerVersion)
							// 记录本次心跳成功
							c.appendHeartbeatRecord(cluster, true, time.Now())
						}
					}
				}

				if failureCount >= c.HeartbeatFailureThreshold {
					// 达到失败阈值，切换为断开并停止心跳，并启动独立的自动重连循环
					cluster.ClusterConnectStatus = constants.ClusterConnectStatusDisconnected
					klog.V(6).Infof("集群 %s 心跳连续失败达到阈值，状态切换为未连接，启动自动重连循环", clusterID)

					// 停止当前心跳循环
					cancel()

					// 启动自动重连循环（退避重试直到成功或被停止）
					c.StartReconnect(clusterID)
					return
				}
			}
		}
	}()
}

// StopHeartbeat 停止指定集群的心跳任务
// @Description 若心跳存在则停止并清理取消函数。
func (c *clusterService) StopHeartbeat(clusterID string) {
	if cancelInterface, ok := c.heartbeatCancel.Load(clusterID); ok {
		if cancel, ok := cancelInterface.(context.CancelFunc); ok && cancel != nil {
			cancel()
		}
		c.heartbeatCancel.Delete(clusterID)
		klog.V(6).Infof("集群 %s 心跳任务已停止", clusterID)
	}
}

// StartReconnect 启动指定集群的自动重连循环
// 中文函数注释：当集群处于不可用状态时，周期性地执行“先断开清理、再尝试连接”，并采用指数退避策略（最大退避秒数可配置）；
// 当检测到集群成功连接后，自动结束重连循环。日志均为中文，便于观察。
func (c *clusterService) StartReconnect(clusterID string) {
	// 初始化重连配置默认值
	if c.ReconnectMaxIntervalSeconds <= 0 {
		c.ReconnectMaxIntervalSeconds = 3600
	}

	// 若已有自动重连任务，先停止
	if cancelInterface, ok := c.reconnectCancel.Load(clusterID); ok {
		if cancel, ok := cancelInterface.(context.CancelFunc); ok && cancel != nil {
			cancel()
		}
		c.reconnectCancel.Delete(clusterID)
	}

	ctx, cancel := context.WithCancel(context.Background())
	c.reconnectCancel.Store(clusterID, cancel)

	klog.V(6).Infof("集群 %s 自动重连循环启动（最大退避 %ds，最大重试次数 %d）", clusterID, c.ReconnectMaxIntervalSeconds, c.MaxRetryAttempts)

	go func(id string) {
		attempt := 0
		backoff := 1 // 初始退避秒数
		maxIntervalSeconds := c.ReconnectMaxIntervalSeconds
		maxRetryAttempts := c.MaxRetryAttempts
		// 如果最大重试次数小于等于0，则设置默认值100
		if maxRetryAttempts <= 0 {
			maxRetryAttempts = 100
		}

		for {
			select {
			case <-ctx.Done():
				klog.V(6).Infof("集群 %s 自动重连循环已停止", id)
				return
			default:
			}

			// 若已连接则结束重连
			if c.IsConnected(id) {
				klog.V(6).Infof("集群 %s 已连接，自动重连循环结束", id)
				cancel()
				c.reconnectCancel.Delete(id)
				return
			}

			attempt++
			// 检查是否超过最大重试次数
			if attempt > maxRetryAttempts {
				klog.V(6).Infof("集群 %s 自动重连已达到最大重试次数 %d，停止重连", id, maxRetryAttempts)
				cancel()
				c.reconnectCancel.Delete(id)
				return
			}

			klog.V(6).Infof("集群 %s 自动重连第 %d 次尝试：先断开清理后重连", id, attempt)

			// 尝试连接
			c.Connect(id)

			// 若连接成功，结束重连循环
			if c.IsConnected(id) {
				klog.V(6).Infof("集群 %s 自动重连成功", id)
				cancel()
				c.reconnectCancel.Delete(id)
				return
			}

			// 指数退避，封顶
			backoff = min(backoff*2, maxIntervalSeconds)
			klog.V(6).Infof("集群 %s 自动重连失败，%ds 后重试", id, backoff)
			time.Sleep(time.Duration(backoff) * time.Second)
		}
	}(clusterID)
}

// StopReconnect 停止指定集群的自动重连循环
// 中文函数注释：若自动重连循环存在则停止，并清理取消函数。
func (c *clusterService) StopReconnect(clusterID string) {
	if cancelInterface, ok := c.reconnectCancel.Load(clusterID); ok {
		if cancel, ok := cancelInterface.(context.CancelFunc); ok && cancel != nil {
			cancel()
		}
		c.reconnectCancel.Delete(clusterID)
		klog.V(6).Infof("集群 %s 自动重连循环已停止", clusterID)
	}
}
