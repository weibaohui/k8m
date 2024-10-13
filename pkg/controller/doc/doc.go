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
