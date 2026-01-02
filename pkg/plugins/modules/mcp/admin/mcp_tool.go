package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/plugins/modules/mcp/models"
	"gorm.io/gorm"
)

type ToolController struct {
}

// @Summary 获取指定MCP服务器的工具列表
// @Security BearerAuth
// @Param name path string true "MCP服务器名称"
// @Success 200 {object} string
// @Router /admin/mcp/server/{name}/tools/list [get]
func (m *ToolController) List(c *gin.Context) {
	name := c.Param("name")
	params := dao.BuildParams(c)
	params.PerPage = 10000
	var tool models.MCPTool
	list, _, err := tool.List(params, func(db *gorm.DB) *gorm.DB {
		return db.Where("server_name=?", name).Order("name asc")
	})
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonList(c, list)
}

// @Summary 快速更新MCP工具状态
// @Security BearerAuth
// @Param id path int true "工具ID"
// @Param status path string true "状态，例如：true、false"
// @Success 200 {object} string
// @Router /admin/mcp/tool/save/id/{id}/status/{status} [post]
func (m *ToolController) QuickSave(c *gin.Context) {
	id := c.Param("id")
	status := c.Param("status")

	var entity models.MCPTool
	entity.ID = utils.ToUInt(id)

	if status == "true" {
		entity.Enabled = true
	} else {
		entity.Enabled = false
	}
	err := dao.DB().Model(&entity).Select("enabled").Updates(entity).Error

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonErrorOrOK(c, err)
}
