package doc

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/kubectl"
	"github.com/weibaohui/k8m/internal/utils/amis"
)

func Doc(c *gin.Context) {
	kind := c.Param("kind")
	docs := kubectl.NewDocs()
	node := docs.Fetch(kind)

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
	// 考虑增加AI翻译
	detail := &DetailReq{}
	err := c.ShouldBindBodyWithJSON(&detail)
	if err != nil {
		amis.WriteJsonError(c, err)
	}
	detail.Translate = detail.Description
	amis.WriteJsonData(c, detail)
}
