package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/flag"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := flag.Init()
		claims, err := utils.GetJWTClaims(c, cfg.JwtTokenSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
			c.Abort()
			return
		}

		c.Set(constants.JwtUserName, claims[constants.JwtUserName])
		c.Set(constants.JwtUserRole, claims[constants.JwtUserRole])
		c.Set(constants.JwtClusters, claims[constants.JwtClusters])
		c.Next()
	}
}
