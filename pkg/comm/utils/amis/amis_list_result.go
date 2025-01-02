package amis

import (
	"github.com/gin-gonic/gin"
)

type ListResponse[T any] struct {
	Count int64 `json:"count"`
	Rows  []T   `json:"rows"`
}

func WriteJsonList[T any](c *gin.Context, data []T) {
	if len(data) > 0 {
		c.JSON(200, gin.H{
			"status": 0,
			"msg":    "success",
			"data": ListResponse[T]{
				Count: int64(len(data)),
				Rows:  data,
			},
		})
	} else {
		c.JSON(200, gin.H{
			"status": 0,
			"msg":    "无数据",
			"data": ListResponse[T]{
				Count: 0,
				Rows:  []T{},
			},
		})
	}
}

func WriteJsonListWithTotal[T any](c *gin.Context, total int64, data []T) {
	if len(data) > 0 {
		c.JSON(200, gin.H{
			"status": 0,
			"msg":    "success",
			"data": ListResponse[T]{
				Count: total,
				Rows:  data,
			},
		})
	} else {
		c.JSON(200, gin.H{
			"status": 0,
			"msg":    "无数据",
			"data": ListResponse[T]{
				Count: total,
				Rows:  []T{},
			},
		})
	}

}

func WriteJsonListWithError[T any](c *gin.Context, data []T, err error) {
	if err != nil {

		c.JSON(200, gin.H{
			"status": 0,
			"msg":    "无数据",
			"data": ListResponse[T]{
				Count: 0,
				Rows:  []T{},
			},
		})
		return
	}
	WriteJsonList(c, data)
}
func WriteJsonListTotalWithError[T any](c *gin.Context, total int64, data []T, err error) {
	if err != nil {

		c.JSON(200, gin.H{
			"status": 0,
			"msg":    "无数据",
			"data": ListResponse[T]{
				Count: 0,
				Rows:  []T{},
			},
		})
		return
	}
	WriteJsonListWithTotal(c, total, data)
}
