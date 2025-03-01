package helm

import (
	"strings"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/helm"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/service"
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
	var repoEntry repo.Entry
	if err := c.ShouldBindJSON(&repoEntry); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	h, err := getHelm(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	if err := h.AddOrUpdateRepo(&repoEntry); err != nil {
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
func ListChart(c *gin.Context) {
	// 从数据库查询列表
	params := dao.BuildParams(c)
	m := &models.HelmChart{}
	items, total, err := m.List(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}

func DeleteRepo(c *gin.Context) {
	ids := c.Param("ids")
	// 删除
	dao.DB().Where("id in ?", strings.Split(ids, ",")).Delete(&models.HelmRepository{})
	dao.DB().Where("repository_id in ?", strings.Split(ids, ",")).Delete(&models.HelmChart{})

	amis.WriteJsonOK(c)
}
func UpdateReposIndex(c *gin.Context) {
	var req struct {
		IDs string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	h, err := getHelm(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	go h.UpdateReposIndex(req.IDs)
	amis.WriteJsonOK(c)
}

// ListReleaseHistory 获取Release的历史版本
func ListReleaseHistory(c *gin.Context) {
	releaseName := c.Param("name")
	h, err := getHelm(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	history, err := h.GetReleaseHistory(releaseName)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, history)
}

func getHelm(c *gin.Context) (helm.Helm, error) {
	selectedCluster := amis.GetSelectedCluster(c)
	restConfig := service.ClusterService().GetClusterByID(selectedCluster).GetRestConfig()
	h, err := helm.New(restConfig)
	return h, err
}

// InstallRelease 安装Helm Release
func InstallRelease(c *gin.Context) {
	var req struct {
		ReleaseName string   `json:"release_name"`
		RepoName    string   `json:"repo_name"`
		ChartName   string   `json:"chart_name"`
		Version     string   `json:"version"`
		Values      []string `json:"values,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	h, err := getHelm(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	if err := h.InstallRelease(req.ReleaseName, req.RepoName, req.ChartName, req.Version, req.Values...); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

// UninstallRelease 卸载Helm Release
func UninstallRelease(c *gin.Context) {
	releaseName := c.Param("name")
	h, err := getHelm(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	if err := h.UninstallRelease(releaseName); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

// UpgradeRelease 升级Helm Release
func UpgradeRelease(c *gin.Context) {
	var req struct {
		ReleaseName string `json:"release_name"`
		RepoName    string `json:"repo_name"`
		Version     string `json:"version"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	h, err := getHelm(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	if err := h.UpgradeRelease(req.ReleaseName, req.RepoName, req.Version); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

// GetChartValue 获取Chart的值
func GetChartValue(c *gin.Context) {
	chartName := c.Param("chart")
	version := c.Param("version")

	selectedCluster := amis.GetSelectedCluster(c)
	restConfig := service.ClusterService().GetClusterByID(selectedCluster).GetRestConfig()
	h, err := helm.New(restConfig)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	value, err := h.GetChartValue(chartName, version)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, value)
}

// GetChartVersions 获取Chart的版本列表
func GetChartVersions(c *gin.Context) {
	chartName := c.Param("chart")

	h, err := getHelm(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	versions, err := h.GetChartVersions(chartName)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, versions)
}
