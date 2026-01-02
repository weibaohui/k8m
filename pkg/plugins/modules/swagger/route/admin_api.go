package route

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/plugins/modules/swagger/models"
	"k8s.io/klog/v2"
)

func RegisterAdminRoutes(admin *gin.RouterGroup) {
	g := admin.Group("/plugins/swagger")
	g.GET("/setting/get", GetSetting)
	g.POST("/setting/update", UpdateSetting)
	klog.V(6).Infof("注册Swagger插件管理路由(admin)")
}

func GetSetting(c *gin.Context) {
	cfg, err := models.GetConfig()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, gin.H{
		"enable_swagger": cfg.Enabled,
	})
}

func UpdateSetting(c *gin.Context) {
	var req struct {
		EnableSwagger bool `json:"enable_swagger"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"message": "请求参数无效",
		})
		return
	}

	if err := models.UpdateConfig(req.EnableSwagger); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOK(c)
}
