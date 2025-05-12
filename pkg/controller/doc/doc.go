package doc

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
)

func Doc(c *gin.Context) {
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

func Detail(c *gin.Context) {
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
