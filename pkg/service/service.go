package service

import "sync"

var localPodService = &podService{
	podLabels: make(map[string][]*PodLabels),
}
var localChatService = &chatService{}
var localNodeService = &nodeService{
	nodeLabels: make(map[string][]*NodeLabels),
}
var localDeploymentService = &deployService{}
var localClusterService = &clusterService{
	clusterConfigs:        []*ClusterConfig{},
	AggregateDelaySeconds: 61, // 没有秒级支持，所以大于1分钟
}
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
