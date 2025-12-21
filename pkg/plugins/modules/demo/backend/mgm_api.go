package backend

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/plugins"
	"k8s.io/klog/v2"
)

// RegisterMgmRoutes 注册Demo插件的管理类（mgm）路由
func RegisterMgmRoutes(mgm *gin.RouterGroup) {
	g := mgm.Group("/plugins/demo")
	// 列表
	g.GET("/items", List)
	// 新增
	g.POST("/items", Create)
	// 更新
	g.POST("/items/:id", Update)
	// 删除
	g.POST("/remove/items/:id", Delete)

	klog.V(6).Infof("注册插件管理路由(mgm): %s", "/items/:id")
}

// Create 新增演示项
func Create(c *gin.Context) {
	// 平台管理员校验
	ok, err := plugins.EnsureUserIsPlatformAdmin(c)
	if !ok {
		amis.WriteJsonError(c, err)
		return
	}
	var req Item
	if err = c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	klog.V(6).Infof("新增演示项请求，名称=%s", req.Name)
	params := dao.BuildParams(c)
	err = dao.GenericSave(params, &req)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, req)
}

// Update 更新演示项
func Update(c *gin.Context) {
	// 平台管理员校验
	ok, err := plugins.EnsureUserIsPlatformAdmin(c)
	if !ok {
		amis.WriteJsonError(c, err)
		return
	}

	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	var req Item
	if err = c.ShouldBindJSON(&req); err != nil {
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
	// 平台管理员校验
	ok, err := plugins.EnsureUserIsPlatformAdmin(c)
	if !ok {
		amis.WriteJsonError(c, err)
		return
	}
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
