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
	docs := kom.DefaultCluster().Status().Docs()

	// apiVersion 有可能包含xxx.com/v1 类似，所以需要处理
	// 前端使用了base64Encode，这里需要反向解析处理
	apiVersion, _ = utils.DecodeBase64(apiVersion)
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
		chatService := service.ChatService{}
		result := chatService.Chat(q)
		detail.Translate = result
	}

	amis.WriteJsonData(c, detail)
}
