package amis

import (
	"github.com/gin-gonic/gin"
)

func GetselectedCluster(c *gin.Context) string {
	selectedCluster, err := c.Cookie("selectedCluster")
	if err != nil {
		return "InCluster"
	}
	return selectedCluster
}
