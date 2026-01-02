package service

var localMcpService = &mcpService{}

func McpService() *mcpService {

	return localMcpService
}
