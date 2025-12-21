package backend

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/plugins"
	"k8s.io/klog/v2"
)

// RegisterRoutes 注册Demo插件的后端路由
func RegisterRoutes(api *gin.RouterGroup) {
	sg := plugins.NewSecuredGroup(api, "demo")
	// 列表
	sg.GET("/items", plugins.AccessRoles, List, "user")
	// 新增
	sg.POST("/items", plugins.AccessPlatformAdmin, Create)
	// 更新
	sg.POST("/items/:id", plugins.AccessPlatformAdmin, Update)
	// 删除
	sg.POST("/remove/items/:id", plugins.AccessPlatformAdmin, Delete)

	klog.V(6).Infof("注册插件路由: %s", "/items/:id")
}

// List 返回演示列表数据
func List(c *gin.Context) {
	klog.V(6).Infof("获取演示列表")

	params := dao.BuildParams(c)
	m := &Item{}
	items, total, err := dao.GenericQuery(params, m)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}

// Create 新增演示项
func Create(c *gin.Context) {
	var req Item
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	klog.V(6).Infof("新增演示项请求，名称=%s", req.Name)
	params := dao.BuildParams(c)
	err := dao.GenericSave(params, &req)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, req)
}

// Update 更新演示项
func Update(c *gin.Context) {

	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	var req Item
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	klog.V(6).Infof("更新演示项请求，ID=%d", id64)
	req.ID = uint(id64)
	params := dao.BuildParams(c)
	err = dao.GenericSave(params, &req)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

// Delete 删除演示项
func Delete(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	klog.V(6).Infof("删除演示项请求，ID=%d", id64)
	params := dao.BuildParams(c)
	err = dao.GenericDelete(params, &Item{}, []int64{int64(id64)})
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}
