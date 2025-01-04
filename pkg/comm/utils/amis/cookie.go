package amis

import (
	"github.com/gin-gonic/gin"
)

func GetDefaultCluster(c *gin.Context) string {
	defaultCluster, err := c.Cookie("defaultCluster")
	if err != nil {
		return "InCluster"
	}
	return defaultCluster
}
