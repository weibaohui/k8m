package mcp

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/service"
	"k8s.io/klog/v2"
)

type MCPServerController struct {
}

// RegisterMCPServerRoutes 注册路由
func RegisterMCPServerRoutes(admin *gin.RouterGroup) {
	ctrl := &MCPServerController{}
	admin.GET("/mcp/list", ctrl.ServerList)
	admin.POST("/mcp/connect/:name", ctrl.Connect)
	admin.POST("/mcp/delete", ctrl.Delete)
	admin.POST("/mcp/save", ctrl.AddOrUpdate)
	admin.POST("/mcp/save/id/:id/status/:status", ctrl.QuickSave)
	admin.GET("/mcp/log/list", ctrl.MCPLogList)
}

// @Summary 获取MCP服务器列表
// @Security BearerAuth
// @Success 200 {object} string
// @Router /admin/mcp/list [get]
func (m *MCPServerController) ServerList(c *gin.Context) {
	params := dao.BuildParams(c)
	var mcpServer models.MCPServerConfig
	list, count, err := mcpServer.List(params)
	amis.WriteJsonListTotalWithError(c, count, list, err)
}

// @Summary 连接指定MCP服务器
// @Security BearerAuth
// @Param name path string true "MCP服务器名称"
// @Success 200 {object} string
// @Router /admin/mcp/connect/{name} [post]
func (m *MCPServerController) Connect(c *gin.Context) {
	name := c.Param("name")
	err := service.McpService().Host().ConnectServer(c.Request.Context(), name)
	amis.WriteJsonErrorOrOK(c, err)
}

// @Summary 删除MCP服务器
// @Security BearerAuth
// @Success 200 {object} string
// @Router /admin/mcp/delete [post]
func (m *MCPServerController) Delete(c *gin.Context) {
	var req struct {
		IDs []int `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var servers []models.MCPServerConfig
	dao.DB().Where("id in?", req.IDs).Find(&servers)
	// 删除
	dao.DB().Where("id in ?", req.IDs).Delete(&models.MCPServerConfig{})
	for _, server := range servers {
		service.McpService().RemoveServer(server)
	}
	amis.WriteJsonOK(c)
}

// @Summary 创建或更新MCP服务器
// @Security BearerAuth
// @Success 200 {object} string
// @Router /admin/mcp/save [post]
func (m *MCPServerController) AddOrUpdate(c *gin.Context) {
	params := dao.BuildParams(c)

	var entity models.MCPServerConfig
	if err := c.ShouldBindJSON(&entity); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	err := entity.Save(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	service.McpService().UpdateServer(entity)
	removeTools(entity)
	addTools(params, entity)

	amis.WriteJsonErrorOrOK(c, err)
}

// @Summary 快速保存MCP服务器状态
// @Security BearerAuth
// @Param id path int true "服务器ID"
// @Param status path string true "状态，例如：true、false"
// @Success 200 {object} string
// @Router /admin/mcp/save/id/{id}/status/{status} [post]
func (m *MCPServerController) QuickSave(c *gin.Context) {
	id := c.Param("id")
	status := c.Param("status")
	params := dao.BuildParams(c)

	var entity models.MCPServerConfig
	err := dao.DB().Where("id = ?", id).First(&entity).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	if status == "true" {
		entity.Enabled = true
	} else {
		entity.Enabled = false
	}
	err = entity.Save(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	removeTools(entity)
	service.McpService().UpdateServer(entity)
	if status == "true" {
		addTools(params, entity)
	}

	amis.WriteJsonErrorOrOK(c, err)
}

// @Summary 获取MCP服务器日志列表
// @Security BearerAuth
// @Success 200 {object} string
// @Router /admin/mcp/log/list [get]
func (m *MCPServerController) MCPLogList(c *gin.Context) {
	params := dao.BuildParams(c)
	var tool models.MCPToolLog
	list, count, err := tool.List(params)
	amis.WriteJsonListTotalWithError(c, count, list, err)
}

func addTools(params *dao.Params, entity models.MCPServerConfig) bool {
	// 获取Tools列表
	if tools, err := service.McpService().GetTools(entity); err == nil {
		for _, tool := range tools {
			mt := models.MCPTool{
				ServerName:  entity.Name,
				Name:        tool.Name,
				Description: tool.Description,
				InputSchema: utils.ToJSON(tool.InputSchema),
				Enabled:     true,
			}
			err = mt.Save(params)
			if err != nil {
				klog.V(6).Infof("保存工具失败:[%s/%s] %v\n", entity.Name, tool.Name, err)
				return true
			}
		}

	}
	return false
}

func removeTools(entity models.MCPServerConfig) {
	dao.DB().Where("server_name = ?", entity.Name).Delete(&models.MCPTool{})
}
