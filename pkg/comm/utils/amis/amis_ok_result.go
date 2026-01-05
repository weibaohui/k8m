package amis

import (
	"github.com/weibaohui/k8m/pkg/response"
)

func WriteJsonOK(c *response.Context) {
	c.JSON(200, response.H{
		"status": 0,
		"msg":    "success",
	})
}
func WriteJsonOKMsg(c *response.Context, msg string) {
	c.JSON(200, response.H{
		"status": 0,
		"msg":    msg,
	})
}
func WriteJsonError(c *response.Context, err error) {
	c.JSON(200, response.H{
		"status": 1,
		"msg":    err.Error(),
	})
}
func WriteJsonErrorOrOK(c *response.Context, err error) {
	if err == nil {
		WriteJsonOK(c)
		return
	}

	WriteJsonError(c, err)
}

func WriteJsonData[T any](c *response.Context, data T) {
	c.JSON(200, response.H{
		"status": 0,
		"msg":    "success",
		"data":   data,
	})
}
