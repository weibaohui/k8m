package mcpkey

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/service"
	"gorm.io/gorm"
)

// Create 处理创建新的MCP密钥的HTTP请求。
// 从请求中解析描述信息，获取当前用户，生成有效期为10年的JWT令牌，并创建包含该信息的MCP密钥记录保存到数据库。
// 失败时返回JSON格式的错误响应，成功时返回操作成功的JSON响应。
func Create(c *gin.Context) {
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
	}

	// 保存到数据库
	if err := mcpKey.Save(params); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOK(c)
}

// List 获取MCP密钥列表
func List(c *gin.Context) {
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

func Delete(c *gin.Context) {
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
