package admin

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/plugins/modules/mcp_runtime/models"
	"github.com/weibaohui/k8m/pkg/response"
	"github.com/weibaohui/k8m/pkg/service"
	"gorm.io/gorm"
)

type KeyController struct{}

func RegisterMCPKeysRoutes(mgm chi.Router) {
	ctrl := &KeyController{}
	mgm.Get("/user/profile/mcp_keys/list", response.Adapter(ctrl.List))
	mgm.Post("/user/profile/mcp_keys/create", response.Adapter(ctrl.Create))
	mgm.Post("/user/profile/mcp_keys/delete/{id}", response.Adapter(ctrl.Delete))
}

// Create 处理创建新的MCP密钥的HTTP请求。
// 从请求中解析描述信息，获取当前用户，生成有效期为10年的JWT令牌，并创建包含该信息的MCP密钥记录保存到数据库。
// 失败时返回JSON格式的错误响应，成功时返回操作成功的JSON响应。
// @Summary 创建MCP密钥
// @Description 为当前用户创建一个新的MCP密钥（10年有效期）
// @Security BearerAuth
// @Param description body string false "密钥描述"
// @Success 200 {object} string "操作成功"
// @Router /mgm/user/profile/mcp_keys/create [post]
func (mc *KeyController) Create(c *response.Context) {
	params := dao.BuildParams(c)

	var req struct {
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 从JWT中获取用户信息
	username := c.GetString(constants.JwtUserName)

	jwt, err := service.UserService().GenerateJWTTokenOnlyUserName(username, time.Hour*24*365*10)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 生成MCP密钥
	mcpKey := &models.McpKey{
		Username:    username,
		McpKey:      utils.RandNLengthString(8),
		Jwt:         jwt,
		Description: req.Description,
		LastUsedAt:  time.Now(),
	}

	// 保存到数据库
	if err := mcpKey.Save(params); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOK(c)
}

// List 获取MCP密钥列表
// @Summary 获取MCP密钥列表
// @Description 获取当前用户的所有MCP密钥
// @Security BearerAuth
// @Success 200 {object} string
// @Router /mgm/user/profile/mcp_keys/list [get]
func (mc *KeyController) List(c *response.Context) {
	username := c.GetString(constants.JwtUserName)
	params := dao.BuildParams(c)

	mcpKey := &models.McpKey{}
	list, _, err := mcpKey.List(params, func(db *gorm.DB) *gorm.DB {
		return db.Where("username = ?", username)
	})

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonData(c, list)
}

// @Summary 删除MCP密钥
// @Description 删除指定ID的MCP密钥
// @Security BearerAuth
// @Param id path string true "MCP密钥ID"
// @Success 200 {object} string "操作成功"
// @Router /mgm/user/profile/mcp_keys/delete/{id} [post]
func (mc *KeyController) Delete(c *response.Context) {
	id := c.Param("id")
	params := dao.BuildParams(c)

	mcpKey := &models.McpKey{}

	err := mcpKey.Delete(params, id)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}
