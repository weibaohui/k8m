package service

import (
	"fmt"
	"sync"

	"github.com/robfig/cron/v3"
	utils2 "github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog/v2"
)

type nodeService struct {
	// 存储节点标签的map，key为集群ID，value为该集群下所有节点的标签map
	nodeLabels map[string][]*NodeLabels
	// 用于保护map的并发访问
	lock sync.RWMutex
}

// NodeLabels 定义结构体，
type NodeLabels struct {
	ClusterName string            // 集群名称
	NodeName    string            // 节点名称
	Labels      map[string]string // 标签
}

// UpdateNodeLabels 更新节点的标签
func (n *nodeService) UpdateNodeLabels(selectedCluster string, nodeName string, labels map[string]string) {
	n.lock.Lock()
	defer n.lock.Unlock()
	if n.nodeLabels == nil {
		n.nodeLabels = make(map[string][]*NodeLabels)
	}
	// 查找是否已存在该节点的标签
	var found bool
	if nodeList, ok := n.nodeLabels[selectedCluster]; ok {
		for _, node := range nodeList {
			if node.NodeName == nodeName {
				node.Labels = labels
				found = true
				break
			}
		}
	} else {
		n.nodeLabels[selectedCluster] = make([]*NodeLabels, 0)
	}
	// 如果节点不存在，则添加新节点
	if !found {
		n.nodeLabels[selectedCluster] = append(n.nodeLabels[selectedCluster], &NodeLabels{
			ClusterName: selectedCluster,
			NodeName:    nodeName,
			Labels:      labels,
		})
	}
}

// DeleteNodeLabels 删除节点的标签
func (n *nodeService) DeleteNodeLabels(selectedCluster string, nodeName string) {
	n.lock.Lock()
	defer n.lock.Unlock()
	if nodeList, ok := n.nodeLabels[selectedCluster]; ok {
		for i, node := range nodeList {
			if node.NodeName == nodeName {
				// 从切片中删除该节点
				n.nodeLabels[selectedCluster] = append(nodeList[:i], nodeList[i+1:]...)
				break
			}
		}
	}
}

// GetNodeLabels 获取节点的标签
func (n *nodeService) GetNodeLabels(selectedCluster string, nodeName string) map[string]string {
	n.lock.RLock()
	defer n.lock.RUnlock()
	if nodeList, ok := n.nodeLabels[selectedCluster]; ok {
		for _, node := range nodeList {
			if node.NodeName == nodeName {
				return node.Labels
			}
		}
	}
	return nil
}

// GetAllNodeLabels 获取所有节点的标签
func (n *nodeService) GetAllNodeLabels(selectedCluster string) map[string]map[string]string {
	n.lock.RLock()
	defer n.lock.RUnlock()
	// 创建一个新的map来返回，避免直接返回内部map
	result := make(map[string]map[string]string)
	if nodeList, ok := n.nodeLabels[selectedCluster]; ok {
		for _, node := range nodeList {
			labels := make(map[string]string)
			for k, v := range node.Labels {
				labels[k] = v
			}
			result[node.NodeName] = labels
		}
	}
	return result
}

// GetUniqueLabels 获取所有节点标签的唯一集合
func (n *nodeService) GetUniqueLabels(selectedCluster string) map[string]string {
	n.lock.RLock()
	defer n.lock.RUnlock()
	// 创建一个新的map来存储唯一的标签
	uniqueLabels := make(map[string]string)
	// 遍历所有节点的标签
	if nodeList, ok := n.nodeLabels[selectedCluster]; ok {
		for _, node := range nodeList {
			// 将每个节点的标签添加到唯一标签集合中
			for k, v := range node.Labels {
				labelKey := fmt.Sprintf("%s=%s", k, v)
				uniqueLabels[labelKey] = labelKey
			}
		}
	}
	return uniqueLabels
}

func (n *nodeService) watchSingleCluster(selectedCluster string) watch.Interface {
	// watch Node资源的变更
	var watcher watch.Interface
	var node v1.Node
	ctx := utils2.GetContextWithAdmin()
	err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&node).Watch(&watcher).Error
	if err != nil {
		klog.Errorf("%s 创建Node监听器失败 %v", selectedCluster, err)
		return nil
	}
	go func() {
		klog.V(6).Infof("%s start watch node", selectedCluster)
		defer watcher.Stop()
		for event := range watcher.ResultChan() {
			err = kom.Cluster(selectedCluster).WithContext(ctx).Tools().ConvertRuntimeObjectToTypedObject(event.Object, &node)
			if err != nil {
				klog.V(6).Infof("%s 无法将对象转换为 *v1.Node 类型: %v", selectedCluster, err)
				return
			}
			// 处理事件
			switch event.Type {
			case watch.Added:
				// 新增节点时，保存节点标签
				n.UpdateNodeLabels(selectedCluster, node.Name, node.Labels)
				klog.V(6).Infof("%s 添加Node [ %s ] 标签数量: %d\n", selectedCluster, node.Name, len(node.Labels))
			case watch.Modified:
				// 修改节点时，更新节点标签
				n.UpdateNodeLabels(selectedCluster, node.Name, node.Labels)
				klog.V(6).Infof("%s 修改Node [ %s ] 标签数量: %d\n", selectedCluster, node.Name, len(node.Labels))
			case watch.Deleted:
				// 删除节点时，删除节点标签
				n.DeleteNodeLabels(selectedCluster, node.Name)
				klog.V(6).Infof("%s 删除Node [ %s ]\n", selectedCluster, node.Name)
			}
		}
	}()

	return watcher
}

func (n *nodeService) Watch() {
	// 设置一个定时器，后台不断更新storageClass状态
	inst := cron.New()
	_, err := inst.AddFunc("@every 1m", func() {
		// 延迟启动cron
		clusters := ClusterService().ConnectedClusters()
		for _, cluster := range clusters {
			selectedCluster := ClusterService().ClusterID(cluster)
			n.SyncNodeStatus(selectedCluster)
			if !cluster.GetClusterWatchStatus("node") {
				selectedCluster := ClusterService().ClusterID(cluster)
				watcher := n.watchSingleCluster(selectedCluster)
				cluster.SetClusterWatchStarted("node", watcher)
			}
			klog.V(6).Infof("执行定时更新Node状态%s", selectedCluster)
		}
	})
	if err != nil {
		klog.Errorf("新增Node定时任务报错: %v\n", err)
	}
	inst.Start()
	klog.V(6).Infof("新增 Node  状态定时更新任务【@every 5m】\n")
}
