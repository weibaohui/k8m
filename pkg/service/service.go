package service

var localPodService = &podService{}
var localChatService = &chatService{}
var localNodeService = &nodeService{}
var localDeploymentService = &deployService{}

func ChatService() *chatService {
	return localChatService
}
func DeploymentService() *deployService {
	return localDeploymentService
}
func PodService() *podService {
	return newPodService()
}
func NodeService() *nodeService {
	return newNodeService()
}
