package helm

import (
	"fmt"
	"time"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"k8s.io/klog/v2"
)

type ReleaseController struct {
}

func RegisterHelmReleaseRoutes(api *gin.RouterGroup) {
	ctrl := &ReleaseController{}

	api.GET("/helm/release/list", ctrl.ListRelease)
	api.GET("/helm/release/ns/:ns/name/:name/history/list", ctrl.ListReleaseHistory)
	api.POST("/helm/release/:release/repo/:repo/chart/:chart/version/:version/install", ctrl.InstallRelease)
	api.POST("/helm/release/ns/:ns/name/:name/uninstall", ctrl.UninstallRelease)
	api.GET("/helm/release/ns/:ns/name/:name/revision/:revision/values", ctrl.GetReleaseValues)
	api.GET("/helm/release/ns/:ns/name/:name/revision/:revision/notes", ctrl.GetReleaseNote)
	api.GET("/helm/release/ns/:ns/name/:name/revision/:revision/install_log", ctrl.GetReleaseInstallLog)
	api.POST("/helm/release/batch/uninstall", ctrl.BatchUninstallRelease)
	api.POST("/helm/release/upgrade", ctrl.UpgradeRelease)

}

// @Summary 获取Release的历史版本
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param ns path string true "命名空间"
// @Param name path string true "Release名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/helm/release/ns/{ns}/name/{name}/history/list [get]
func (hr *ReleaseController) ListReleaseHistory(c *gin.Context) {
	releaseName := c.Param("name")
	ns := c.Param("ns")

	// 检查权限
	_, _, err := handleCommonLogic(c, "list", releaseName, ns, "")
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	h, err := getHelm(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	history, err := h.GetReleaseHistory(ns, releaseName)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, history)
}

// @Summary 获取Release列表
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/helm/release/list [get]
func (hr *ReleaseController) ListRelease(c *gin.Context) {
	// 检查权限
	_, _, err := handleCommonLogic(c, "list", "", "", "")
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	h, err := getHelm(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	list, err := h.GetReleaseList()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	slice.SortBy(list, func(i, j *models.Release) bool {

		it, err := time.Parse("2006-01-02 15:04:05.000000 -0700 MST", i.Updated)
		if err != nil {
			return false
		}
		jt, err := time.Parse("2006-01-02 15:04:05.000000 -0700 MST", j.Updated)
		if err != nil {
			return false
		}
		return it.Before(jt)
	})
	if list == nil {
		list = make([]*models.Release, 0)
	}
	amis.WriteJsonData(c, list)
}

// @Summary 安装Helm Release
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param release path string true "Release名称"
// @Param repo path string true "仓库名称"
// @Param chart path string true "Chart名称"
// @Param version path string true "版本号"
// @Param body body object true "安装参数"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/helm/release/{release}/repo/{repo}/chart/{chart}/version/{version}/install [post]
func (hr *ReleaseController) InstallRelease(c *gin.Context) {

	releaseName := c.Param("release")
	repoName := c.Param("repo")
	chartName := c.Param("chart")
	version := c.Param("version")

	var req struct {
		Values    string `json:"values,omitempty"`
		Namespace string `json:"ns,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 检查权限
	_, _, err := handleCommonLogic(c, "create", releaseName, req.Namespace, repoName)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	h, err := getHelm(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	if releaseName == "" {
		releaseName = fmt.Sprintf("%s-%d", chartName, utils.RandNDigitInt(8))
	}
	err = h.InstallRelease(req.Namespace, releaseName, repoName, chartName, version, req.Values)
	if err != nil {
		klog.Errorf("install %s/%s error %v", req.Namespace, releaseName, err)
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOKMsg(c, "正在安装中，界面显示可能有延迟")
}

// @Summary 卸载Helm Release
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param ns path string true "命名空间"
// @Param name path string true "Release名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/helm/release/ns/{ns}/name/{name}/uninstall [post]
func (hr *ReleaseController) UninstallRelease(c *gin.Context) {
	releaseName := c.Param("name")
	ns := c.Param("ns")

	// 检查权限
	_, _, err := handleCommonLogic(c, "delete", releaseName, ns, "")
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	h, err := getHelm(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	if err := h.UninstallRelease(ns, releaseName); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

// @Summary 获取ReleaseNote
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param ns path string true "命名空间"
// @Param name path string true "Release名称"
// @Param revision path string true "版本号"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/helm/release/ns/{ns}/name/{name}/revision/{revision}/notes [get]
func (hr *ReleaseController) GetReleaseNote(c *gin.Context) {
	releaseName := c.Param("name")
	ns := c.Param("ns")
	revision := c.Param("revision")

	// 检查权限
	_, _, err := handleCommonLogic(c, "get", releaseName, ns, "")
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	h, err := getHelm(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	note, err := h.GetReleaseNoteWithRevision(ns, releaseName, revision)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, gin.H{
		"note": note,
	})
}

// @Summary 获取Release安装Log
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param ns path string true "命名空间"
// @Param name path string true "Release名称"
// @Param revision path string true "版本号"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/helm/release/ns/{ns}/name/{name}/revision/{revision}/install_log [get]
func (hr *ReleaseController) GetReleaseInstallLog(c *gin.Context) {
	releaseName := c.Param("name")
	ns := c.Param("ns")

	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 检查权限
	_, _, err = handleCommonLogic(c, "get", releaseName, ns, "")
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	rr, err := models.GetHelmReleaseByNsAndReleaseName(ns, releaseName, selectedCluster)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, gin.H{
		"result": rr.Result,
	})
}

// @Summary 获取安装yaml
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param ns path string true "命名空间"
// @Param name path string true "Release名称"
// @Param revision path string true "版本号"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/helm/release/ns/{ns}/name/{name}/revision/{revision}/values [get]
func (hr *ReleaseController) GetReleaseValues(c *gin.Context) {
	releaseName := c.Param("name")
	ns := c.Param("ns")
	revision := c.Param("revision")
	// 检查权限
	_, _, err := handleCommonLogic(c, "get", releaseName, ns, "")
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	h, err := getHelm(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	ret := ""
	if revision == "" {
		ret, err = h.GetReleaseValues(ns, releaseName)
		if err != nil {
			amis.WriteJsonError(c, err)
			return
		}
	} else {
		ret, err = h.GetReleaseValuesWithRevision(ns, releaseName, revision)
		if err != nil {
			amis.WriteJsonError(c, err)
			return
		}
	}

	amis.WriteJsonData(c, ret)
}

// @Summary 批量卸载Helm Release
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param body body object true "批量卸载参数"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/helm/release/batch/uninstall [post]
func (hr *ReleaseController) BatchUninstallRelease(c *gin.Context) {
	var req struct {
		Names      []string `json:"name_list"`
		Namespaces []string `json:"ns_list"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	for i := 0; i < len(req.Names); i++ {
		name := req.Names[i]
		ns := req.Namespaces[i]
		h, err := getHelm(c)
		if err != nil {
			amis.WriteJsonError(c, err)
			return
		}

		// 检查权限
		_, _, err = handleCommonLogic(c, "delete", name, ns, "")
		if err != nil {
			amis.WriteJsonError(c, err)
			return
		}
		x := h.UninstallRelease(ns, name)
		if x != nil {
			klog.V(6).Infof("batch remove %s/%s error %v", ns, name, x)
			err = x
		}
	}

	amis.WriteJsonOK(c)
}

// @Summary 升级Helm Release
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param body body object true "升级参数"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/helm/release/upgrade [post]
func (hr *ReleaseController) UpgradeRelease(c *gin.Context) {

	var req struct {
		Name      string `json:"name,omitempty"`
		Namespace string `json:"namespace,omitempty"`
		Values    string `json:"values,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 检查权限
	_, _, err := handleCommonLogic(c, "update", req.Name, req.Namespace, "")
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	h, err := getHelm(c)

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	if err := h.UpgradeRelease(req.Namespace, req.Name, req.Values); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}
