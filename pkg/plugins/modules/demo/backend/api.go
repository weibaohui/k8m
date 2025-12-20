package backend

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"k8s.io/klog/v2"
)

// RegisterRoutes 注册Demo插件的后端路由
func RegisterRoutes(api *gin.RouterGroup) {
	grp := api.Group("/plugins/demo")
	// 页面Schema
	grp.GET("/page", Page)
	// 列表
	grp.GET("/items", List)
	// 新增
	grp.POST("/items", Create)
	// 更新
	grp.PUT("/items/:id", Update)
	// 删除
	grp.DELETE("/items/:id", Delete)
}

// Page 返回AMIS页面Schema
func Page(c *gin.Context) {
	schema := gin.H{
		"type":  "page",
		"title": "Demo 列表",
		"body": []any{
			gin.H{
				"type": "crud",
				"api":  "get:/api/plugins/demo/items",
				"columns": []any{
					gin.H{"name": "id", "label": "ID"},
					gin.H{"name": "name", "label": "名称"},
					gin.H{"name": "description", "label": "描述"},
				},
				"headerToolbar": []any{
					gin.H{
						"type":       "button",
						"actionType": "dialog",
						"label":      "新增",
						"level":      "primary",
						"dialog": gin.H{
							"title": "新增",
							"body": gin.H{
								"type": "form",
								"api":  "post:/api/plugins/demo/items",
								"controls": []any{
									gin.H{"type": "text", "name": "name", "label": "名称", "required": true},
									gin.H{"type": "textarea", "name": "description", "label": "描述"},
								},
							},
						},
					},
				},
				"itemActions": []any{
					gin.H{
						"type":       "button",
						"actionType": "dialog",
						"label":      "编辑",
						"dialog": gin.H{
							"title": "编辑",
							"body": gin.H{
								"type": "form",
								"api":  "put:/api/plugins/demo/items/${id}",
								"controls": []any{
									gin.H{"type": "text", "name": "name", "label": "名称", "required": true},
									gin.H{"type": "textarea", "name": "description", "label": "描述"},
								},
							},
						},
					},
					gin.H{
						"type":        "button",
						"actionType":  "ajax",
						"label":       "删除",
						"level":       "danger",
						"confirmText": "确认删除该项？",
						"api":         "delete:/api/plugins/demo/items/${id}",
					},
				},
			},
		},
	}
	amis.WriteJsonData(c, schema)
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

