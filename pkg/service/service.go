package service

var localPodService = &podService{}
var localChatService = &chatService{}
var localNodeService = &nodeService{}
var localDeploymentService = &deployService{}
var localClusterService = &clusterService{
	clusterConfigs: []*clusterConfig{},
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
