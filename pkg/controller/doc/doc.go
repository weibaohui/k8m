package doc

import (
	"fmt"

	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/plugins/api"
	"github.com/weibaohui/k8m/pkg/response"
	"github.com/weibaohui/kom/kom"
)

type Controller struct{}

// RegisterRoutes 注册路由

func RegisterRoutes(r chi.Router) {
	ctrl := &Controller{}
	r.Get("/doc/gvk/{api_version}/{kind}", response.Adapter(ctrl.Doc))
	r.Get("/doc/kind/{kind}/group/{group}/version/{version}", response.Adapter(ctrl.Doc))
	r.Post("/doc/detail", response.Adapter(ctrl.Detail))
}

// @Summary 获取Kubernetes资源文档信息
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Param group path string true "API组"
// @Param version path string true "API版本"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/doc/kind/{kind}/group/{group}/version/{version} [get]
func (cc *Controller) Doc(c *response.Context) {
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

	amis.WriteJsonData(c, response.H{
		"options": []any{
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
// @Param cluster query string true "集群名称"
// @Param request body DetailReq true "请求体，包含description字段"
// @Success 200 {object} DetailReq
// @Router /k8s/cluster/{cluster}/doc/detail [post]
func (cc *Controller) Detail(c *response.Context) {
	detail := &DetailReq{}
	err := c.ShouldBindJSON(&detail)
	if err != nil {
		amis.WriteJsonError(c, err)
	}
	if detail.Description != "" {
		q := fmt.Sprintf("请翻译下面的语句，注意直接给出翻译内容，不要解释。待翻译内如如下：\n\n%s", detail.Description)
		ctxInst := amis.GetContextWithUser(c)
		ai := api.AIChatService()
		if result, err := ai.Chat(ctxInst, q); err == nil {
			detail.Translate = result
		}
	}

	amis.WriteJsonData(c, detail)
}
