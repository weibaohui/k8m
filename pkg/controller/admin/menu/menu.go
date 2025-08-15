package menu

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"gorm.io/gorm"
)

type AdminMenuController struct {
}

// AdminMenu 用于菜单相关接口
// 路由注册函数
func RegisterAdminMenuRoutes(admin *gin.RouterGroup) {

	ctrl := AdminMenuController{}
	// menu 平台管理员可操作，管理菜单
	admin.GET("/menu/list", ctrl.List)
	admin.GET("/menu/history", ctrl.History)
	admin.POST("/menu/save", ctrl.Save)
	admin.POST("/menu/delete/:ids", ctrl.Delete)

}

// @Summary 获取菜单列表
// @Description 获取所有菜单版本信息
// @Security BearerAuth
// @Success 200 {object} []models.Menu
// @Router /admin/menu/list [get]
func (a *AdminMenuController) List(c *gin.Context) {
	params := dao.BuildParams(c)
	m := &models.Menu{}
	items, _, err := m.List(params, func(db *gorm.DB) *gorm.DB {
		return db
	})
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, items)
}

// @Summary 保存菜单
// @Description 新增或更新菜单（每次操作生成新版本）
// @Security BearerAuth
// @Accept json
// @Param data body models.Menu true "菜单内容"
// @Success 200 {object} map[string]interface{}
// @Router /admin/menu/save [post]
func (a *AdminMenuController) Save(c *gin.Context) {
	params := dao.BuildParams(c)
	m := &models.Menu{}
	if err := c.ShouldBind(&m); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	err := m.Save(params, func(db *gorm.DB) *gorm.DB {
		return db
	})
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

// @Summary 获取菜单历史记录
// @Description 获取菜单修改历史记录，按时间倒序排列
// @Security BearerAuth
// @Success 200 {object} []models.Menu
// @Router /admin/menu/history [get]
func (a *AdminMenuController) History(c *gin.Context) {
	params := dao.BuildParams(c)
	m := &models.Menu{}
	params.PerPage = 100000
	items, _, err := m.List(params, func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC") // 按创建时间倒序排序
	})
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, items)
}

// @Summary 删除菜单
// @Description 根据ID批量删除菜单版本
// @Security BearerAuth
// @Param ids path string true "菜单ID，多个用逗号分隔"
// @Success 200 {object} string
// @Router /admin/menu/delete/{ids} [post]
func (a *AdminMenuController) Delete(c *gin.Context) {
	ids := c.Param("ids")
	params := dao.BuildParams(c)
	m := &models.Menu{}

	err := m.Delete(params, ids)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}
