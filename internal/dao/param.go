package dao

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/constants"
)

// Params 用于处理分页和排序参数，以及上下文中的用户信息
type Params struct {
	OrderBy  string                 // 排序字段
	OrderDir string                 // 排序方向
	Page     int                    // 当前页
	PerPage  int                    // 每页数量
	Queries  map[string]interface{} // 动态查询条件
	UserName string                 // 登录用户名

}

// BuildDefaultParams 从 gin.Context 中获取默认的分页和排序参数
func BuildDefaultParams() *Params {
	// 返回 Params 结构体
	return &Params{
		PerPage: 1000000,
	}
}

// BuildParams 从 gin.Context 中获取分页和排序参数
func BuildParams(c *gin.Context) *Params {
	// 获取排序字段，默认为 "id"
	orderBy := c.DefaultQuery("orderBy", "id")

	// 获取排序方向，默认为 "asc"
	orderDir := c.DefaultQuery("orderDir", "desc")

	// 获取当前页码，默认为 1
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1 // 确保页码大于 0
	}

	// 获取每页显示条数，默认为 20
	perPageStr := c.DefaultQuery("perPage", "20")
	perPage, err := strconv.Atoi(perPageStr)
	if err != nil || perPage < 1 {
		perPage = 15 // 确保每页数量大于 0
	}

	// 构建动态查询条件
	queries := make(map[string]interface{})
	// 遍历所有查询参数
	for key, values := range c.Request.URL.Query() {
		// values 是 []string，我们只取第一个值
		if len(values) > 0 {
			queries[key] = values[0] // 只存储第一个值
		}
	}

	userName := c.GetString(constants.JwtUserName)

	// 返回 Params 结构体
	return &Params{
		OrderBy:  orderBy,
		OrderDir: orderDir,
		Page:     page,
		PerPage:  perPage,
		Queries:  queries,
		UserName: userName,
	}
}
