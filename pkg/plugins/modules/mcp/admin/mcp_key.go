package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/plugins/modules/mcp/models"
)

type KeyController struct {
}

func (r *KeyController) List(c *gin.Context) {
	params := dao.BuildParams(c)
	m := &models.McpKey{}
	items, total, err := m.List(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}

func (r *KeyController) Save(c *gin.Context) {
	var key models.McpKey
	if err := c.ShouldBindJSON(&key); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	if key.ID == 0 {
		if err := dao.DB().Create(&key).Error; err != nil {
			amis.WriteJsonError(c, err)
			return
		}
	} else {
		if err := dao.DB().Save(&key).Error; err != nil {
			amis.WriteJsonError(c, err)
			return
		}
	}
	amis.WriteJsonOK(c)
}

func (r *KeyController) Delete(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	if err := dao.DB().Where("id in ?", req.IDs).Delete(&models.McpKey{}).Error; err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOK(c)
}

func (r *KeyController) MyList(c *gin.Context) {
	username := amis.GetLoginUser(c)
	m := &models.McpKey{}
	items, total, err := m.ListByUser(dao.DB(), username)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}

func (r *KeyController) GenKey(c *gin.Context) {
	username := amis.GetLoginUser(c)
	key := &models.McpKey{
		Username: username,
		McpKey:   generateRandomKey(),
	}
	if err := dao.DB().Create(&key).Error; err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, key)
}

func (r *KeyController) RefreshJWT(c *gin.Context) {
	amis.WriteJsonOK(c)
}

func generateRandomKey() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}
