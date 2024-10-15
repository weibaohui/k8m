package doc

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/kubectl"
	"github.com/weibaohui/k8m/internal/utils"
	"github.com/weibaohui/k8m/internal/utils/amis"
)

func Doc(c *gin.Context) {
	kind := c.Param("kind")
	apiVersion := c.Param("api_version")
	docs := kubectl.NewDocs()

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
	// TODO 考虑增加AI翻译
	detail := &DetailReq{}
	err := c.ShouldBindBodyWithJSON(&detail)
	if err != nil {
		amis.WriteJsonError(c, err)
	}
	// detail.Translate = detail.Description
	amis.WriteJsonData(c, detail)
}
