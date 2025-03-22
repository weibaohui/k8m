package mcp

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/service"
)

func List(c *gin.Context) {
	servers := service.McpService().Host().ListServers()
	amis.WriteJsonData(c, servers)
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
	// 检查权限
	_, _, err := handleCommonLogic(c, "Delete", utils.ToJSON(req.IDs), "")
	if err != nil {
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
	// 检查权限
	_, _, err := handleCommonLogic(c, "AddOrUpdateRepo", entity.Name, entity.URL)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	err = entity.Save(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	service.McpService().UpdateServer(entity)
	amis.WriteJsonErrorOrOK(c, err)
}

func handleCommonLogic(c *gin.Context, action string, name, url string) (string, string, error) {
	ctx := amis.GetContextWithUser(c)
	username := fmt.Sprintf("%s", ctx.Value(constants.JwtUserName))
	role := fmt.Sprintf("%s", ctx.Value(constants.JwtUserRole))

	log := models.OperationLog{
		Action:       action,
		Cluster:      "",
		Kind:         "MCP",
		Name:         fmt.Sprintf("%s[%s]", name, url),
		Namespace:    "",
		UserName:     username,
		Group:        "",
		Role:         role,
		ActionResult: "success",
	}

	var err error
	if role == models.RoleClusterReadonly {
		err = fmt.Errorf("非管理员不能%s资源", action)
	}
	if err != nil {
		log.ActionResult = err.Error()
	}
	go func() {
		time.Sleep(1 * time.Second)
		service.OperationLogService().Add(&log)
	}()
	return username, role, err
}
