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

func ServerList(c *gin.Context) {
	params := dao.BuildParams(c)
	var mcpServer models.MCPServerConfig
	list, count, err := mcpServer.List(params)
	amis.WriteJsonListTotalWithError(c, count, list, err)
}
func Connect(c *gin.Context) {
	name := c.Param("name")
	err := service.McpService().Host().ConnectServer(c.Request.Context(), name)
	amis.WriteJsonErrorOrOK(c, err)
}

func Delete(c *gin.Context) {
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
func AddOrUpdate(c *gin.Context) {
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
func QuickSave(c *gin.Context) {
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

func removeTools(entity models.MCPServerConfig) {
	dao.DB().Where("server_name = ?", entity.Name).Delete(&models.MCPTool{})
}
