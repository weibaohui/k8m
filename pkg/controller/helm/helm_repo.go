package helm

import (
	"strings"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"gorm.io/gorm"
)

type RepoController struct {
}

func RegisterHelmRepoRoutes(admin *gin.RouterGroup) {
	ctrl := &RepoController{}
	// helm
	admin.GET("/helm/repo/list", ctrl.List)
	admin.POST("/helm/repo/delete/:ids", ctrl.Delete)
	admin.POST("/helm/repo/update_index", ctrl.UpdateReposIndex)
	admin.POST("/helm/repo/save", ctrl.Save)
}

// @Summary Helm仓库列表
// @Description 获取所有Helm仓库信息
// @Security BearerAuth
// @Success 200 {object} string
// @Router /admin/helm/repo/list [get]
func (r *RepoController) List(c *gin.Context) {
	// 从数据库查询列表
	params := dao.BuildParams(c)
	m := &models.HelmRepository{}
	items, total, err := m.List(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}

// Save 添加或更新Helm仓库
// @Summary 添加或更新Helm仓库
// @Description 添加或更新一个Helm仓库信息
// @Security BearerAuth
// @Param repo body models.HelmRepository true "Helm仓库信息"
// @Success 200 {object} string "操作成功"
// @Router /admin/helm/repo/save [post]
func (r *RepoController) Save(c *gin.Context) {
	var repo models.HelmRepository
	if err := c.ShouldBindJSON(&repo); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	h, err := getHelmWithNoCluster()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	if err = h.AddOrUpdateRepo(&repo); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

// @Summary Helm仓库选项列表
// @Description 获取所有Helm仓库名称，用于下拉选项
// @Security BearerAuth
// @Success 200 {object} string
// @Router /admin/helm/repo/option_list [get]
func RepoOptionList(c *gin.Context) {
	params := dao.BuildParams(c)
	m := &models.HelmRepository{}
	items, _, err := m.List(params, func(db *gorm.DB) *gorm.DB {
		return db.Distinct("name")
	})
	if err != nil {
		amis.WriteJsonData(c, gin.H{
			"options": make([]map[string]string, 0),
		})
		return
	}
	var names []map[string]string
	for _, n := range items {
		names = append(names, map[string]string{
			"label": n.Name,
			"value": n.Name,
		})
	}
	slice.SortBy(names, func(a, b map[string]string) bool {
		return a["label"] < b["label"]
	})
	amis.WriteJsonData(c, gin.H{
		"options": names,
	})
}

// @Summary 删除Helm仓库
// @Description 删除一个或多个Helm仓库
// @Security BearerAuth
// @Param ids path string true "要删除的仓库ID，多个用逗号分隔"
// @Success 200 {object} string "操作成功"
// @Router /admin/helm/repo/delete/{ids} [post]
func (r *RepoController) Delete(c *gin.Context) {
	ids := c.Param("ids")

	h, err := getHelmWithNoCluster()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	m := &models.HelmRepository{}
	list, _, err := m.List(dao.BuildDefaultParams(), func(db *gorm.DB) *gorm.DB {
		return db.Where("id in ?", strings.Split(ids, ","))
	})
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	for _, repository := range list {
		err = h.RemoveRepo(repository.Name)
	}

	// 删除
	dao.DB().Where("id in ?", strings.Split(ids, ",")).Delete(&models.HelmRepository{})

	dao.DB().Where("repository_id in ?", strings.Split(ids, ",")).Delete(&models.HelmChart{})

	amis.WriteJsonOK(c)
}

// @Summary 更新Helm仓库索引
// @Description 更新指定Helm仓库的索引信息
// @Security BearerAuth
// @Param ids body string true "要更新索引的仓库ID，多个用逗号分隔"
// @Success 200 {object} string "操作成功"
// @Router /admin/helm/repo/update_index [post]
func (r *RepoController) UpdateReposIndex(c *gin.Context) {
	var req struct {
		IDs string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	h, err := getHelmWithNoCluster()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	go h.UpdateReposIndex(req.IDs)
	amis.WriteJsonOK(c)
}
