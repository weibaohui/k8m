package helm

import (
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
)

type ChartController struct {
}

func RegisterHelmChartRoutes(mgm *gin.RouterGroup) {
	ctrl := &ChartController{}
	mgm.GET("/helm/repo/:repo/chart/:chart/versions", ctrl.ChartVersionOptionList)
	mgm.GET("/helm/repo/:repo/chart/:chart/version/:version/values", ctrl.GetChartValue)
	mgm.GET("/helm/chart/list", ctrl.ListChart)
}

func (hc *ChartController) ListChart(c *gin.Context) {
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

// GetChartValue 获取Chart的值
func (hc *ChartController) GetChartValue(c *gin.Context) {
	repoName := c.Param("repo")
	chartName := c.Param("chart")
	version := c.Param("version")
	h, err := getHelmWithNoCluster()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	value, err := h.GetChartValue(repoName, chartName, version)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonData(c, gin.H{
		"yaml": value,
	})
}

// ChartVersionOptionList 获取Chart的版本列表
func (hc *ChartController) ChartVersionOptionList(c *gin.Context) {
	repoName := c.Param("repo")
	chartName := c.Param("chart")
	h, err := getHelmWithNoCluster()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	versions, err := h.GetChartVersions(repoName, chartName)
	if err != nil {
		amis.WriteJsonData(c, gin.H{
			"options": make([]map[string]string, 0),
		})
		return
	}

	sort.Slice(versions, func(i, j int) bool {
		return utils.CompareVersions(versions[i], versions[j])
	})

	amis.WriteJsonData(c, gin.H{
		"options": versions,
	})

}
