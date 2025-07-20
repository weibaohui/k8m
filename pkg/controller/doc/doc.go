package doc

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
)

type Controller struct{}

func RegisterRoutes(api *gin.RouterGroup) {
	ctrl := &Controller{}
	api.GET("/doc/gvk/:api_version/:kind", ctrl.Doc)
	api.GET("/doc/kind/:kind/group/:group/version/:version", ctrl.Doc)
	api.POST("/doc/detail", ctrl.Detail)
}

// @Summary 获取Kubernetes资源文档信息
// @Security BearerAuth
// @Param api_version path string true "API版本(base64编码)"
// @Param kind path string true "资源类型"
// @Param kind path string true "资源类型"
// @Param group path string true "API组"
// @Param version path string true "API版本"
// @Success 200 {object} string
// @Router /doc/kind/{kind}/group/{group}/version/{version} [get]
func (cc *Controller) Doc(c *gin.Context) {
	kind := c.Param("kind")
	apiVersion := c.Param("api_version")
	group := c.Param("group")
	version := c.Param("version")
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	ctx := amis.GetContextWithUser(c)

	// apiVersion 有可能包含xxx.com/v1 类似，所以需要处理
	// 前端使用了base64Encode，这里需要反向解析处理
	if apiVersion != "" {
		apiVersion, err = utils.DecodeBase64(apiVersion)
		if err != nil {
			amis.WriteJsonError(c, err)
			return
		}
	}
	if apiVersion == "" {
		// 没有传递apiVersion
		apiVersion = fmt.Sprintf("%s/%s", group, version)
	}

	docs := kom.Cluster(selectedCluster).WithContext(ctx).Status().Docs()
	node := docs.FetchByGVK(apiVersion, kind)

	amis.WriteJsonData(c, gin.H{
		"options": []interface{}{
			node,
		},
	})
}

type DetailReq struct {
	Description string `json:"description"`
	Translate   string `json:"translate"`
}

// @Summary 获取文档详情(含翻译)
// @Security BearerAuth
// @Param request body DetailReq true "请求体，包含description字段"
// @Success 200 {object} DetailReq
// @Router /doc/detail [post]
func (cc *Controller) Detail(c *gin.Context) {
	detail := &DetailReq{}
	err := c.ShouldBindBodyWithJSON(&detail)
	if err != nil {
		amis.WriteJsonError(c, err)
	}
	if detail.Description != "" {
		q := fmt.Sprintf("请翻译下面的语句，注意直接给出翻译内容，不要解释。待翻译内如如下：\n\n%s", detail.Description)
		chatService := service.ChatService()
		result := chatService.Chat(c, q)
		detail.Translate = result
	}

	amis.WriteJsonData(c, detail)
}
