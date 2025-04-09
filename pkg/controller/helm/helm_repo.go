package helm

import (
	"strings"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"gorm.io/gorm"
	"helm.sh/helm/v3/pkg/repo"
)

func ListRepo(c *gin.Context) {
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

// AddOrUpdateRepo 添加或更新Helm仓库
func AddOrUpdateRepo(c *gin.Context) {
	ns := c.Param("ns")

	// 检查权限
	_, _, err := handleCommonLogic(c, "AddOrUpdateRepo", "", ns, "")
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var repoEntry repo.Entry
	if err = c.ShouldBindJSON(&repoEntry); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	h, err := getHelm(c, ns)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	if err = h.AddOrUpdateRepo(&repoEntry); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

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

func DeleteRepo(c *gin.Context) {
	ids := c.Param("ids")

	// 检查权限
	_, _, err := handleCommonLogic(c, "DeleteRepo", ids, "", "")
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 删除
	dao.DB().Where("id in ?", strings.Split(ids, ",")).Delete(&models.HelmRepository{})
	dao.DB().Where("repository_id in ?", strings.Split(ids, ",")).Delete(&models.HelmChart{})

	amis.WriteJsonOK(c)
}
func UpdateReposIndex(c *gin.Context) {
	ns := c.Param("ns")

	var req struct {
		IDs string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 检查权限
	_, _, err := handleCommonLogic(c, "UpdateReposIndex", req.IDs, "", "")
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	h, err := getHelm(c, ns)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	go h.UpdateReposIndex(req.IDs)
	amis.WriteJsonOK(c)
}
