package service

import (
    "sync"

    "github.com/weibaohui/k8m/pkg/lease"
    "k8s.io/client-go/rest"
)

var localPodService = &podService{
	podLabels: make(map[string][]*PodLabels),
}
var localChatService = &chatService{}
var localNodeService = &nodeService{
	nodeLabels: make(map[string][]*NodeLabels),
}
var localDeploymentService = &deployService{}

var localClusterService = newClusterService()
var localStorageClassService = &storageClassService{}
var localIngressClassService = &ingressClassService{}
var localPVCService = &pvcService{
	CountList: []*pvcCount{},
}
var localPVService = &pvService{
	CountList: []*pvCount{},
}
var localIngressService = &ingressService{
	CountList: []*ingressCount{},
}
var localUserService = &userService{
	cacheKeys: sync.Map{},
}
var localOperationLogService = NewOperationLogService()
var localShellLogService = &shellLogService{}
var localAiService = &aiService{}
var localMcpService = &mcpService{}
var localPromptService = &promptService{}
var localLeaseManager = lease.NewManager()

// init 中文函数注释：在 service 初始化时向 lease 包注入 ClusterID → RestConfig 的解析器，避免循环引入。
func init() {
    lease.GetRestConfigByClusterID = func(clusterID string) *rest.Config {
        c := localClusterService.GetClusterByID(clusterID)
        if c == nil {
            return nil
        }
        return c.GetRestConfig()
    }
}

func PromptService() *promptService {
	return localPromptService
}

func ChatService() *chatService {
	return localChatService
}
func DeploymentService() *deployService {
	return localDeploymentService
}
func PodService() *podService {
	return localPodService
}
func NodeService() *nodeService {
	return localNodeService
}
func ClusterService() *clusterService {
	return localClusterService
}
func StorageClassService() *storageClassService {
	return localStorageClassService
}
func IngressClassService() *ingressClassService {
	return localIngressClassService
}
func PVCService() *pvcService {
	return localPVCService
}
func PVService() *pvService {
	return localPVService
}
func IngressService() *ingressService {
	return localIngressService
}

func UserService() *userService {
	return localUserService
}

func OperationLogService() *operationLogService {
	return localOperationLogService
}
func ShellLogService() *shellLogService {
	return localShellLogService
}
func AIService() *aiService {
	return localAiService

}

func McpService() *mcpService {

    return localMcpService
}

func ConfigService() *configService {
    return NewConfigService()
}

func LeaseManager() lease.Manager {
    return localLeaseManager
}
