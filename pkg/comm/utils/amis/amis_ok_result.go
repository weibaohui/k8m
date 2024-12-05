package amis

import (
	"github.com/gin-gonic/gin"
)

func WriteJsonOK(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": 0,
		"msg":    "success",
	})
}
func WriteJsonOKMsg(c *gin.Context, msg string) {
	c.JSON(200, gin.H{
		"status": 0,
		"msg":    msg,
	})
}
func WriteJsonError(c *gin.Context, err error) {
	c.JSON(200, gin.H{
		"status": 1,
		"msg":    err.Error(),
	})
}
func WriteJsonErrorOrOK(c *gin.Context, err error) {
	if err == nil {
		WriteJsonOK(c)
		return
	}

	WriteJsonError(c, err)
}

func WriteJsonData[T any](c *gin.Context, data T) {
	c.JSON(200, gin.H{
		"status": 0,
		"msg":    "success",
		"data":   data,
	})
}
