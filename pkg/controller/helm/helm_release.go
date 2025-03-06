package helm

import (
	"fmt"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/klog/v2"
)

// ListReleaseHistory 获取Release的历史版本
func ListReleaseHistory(c *gin.Context) {
	releaseName := c.Param("name")
	ns := c.Param("ns")
	h, err := getHelm(c, ns)
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
func ListRelease(c *gin.Context) {
	ns := c.Param("ns")
	h, err := getHelm(c, ns)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	list, err := h.GetReleaseList()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	slice.SortBy(list, func(i, j *release.Release) bool {
		return i.Info.LastDeployed.After(j.Info.LastDeployed)
	})
	if list == nil {
		list = make([]*release.Release, 0)
	}
	amis.WriteJsonData(c, list)
}

// InstallRelease 安装Helm Release
func InstallRelease(c *gin.Context) {

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
	_, _, err := handleCommonLogic(c, "InstallRelease", releaseName, req.Namespace, repoName)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	h, err := getHelm(c, req.Namespace)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	if releaseName == "" {
		releaseName = fmt.Sprintf("%s-%d", chartName, utils.RandNDigitInt(8))
	}
	go func() {
		if err := h.InstallRelease(req.Namespace, releaseName, repoName, chartName, version, req.Values); err != nil {
			klog.Errorf("install %s/%s error %v", req.Namespace, releaseName, err)
		}
	}()

	amis.WriteJsonOKMsg(c, "正在安装中，界面显示可能有延迟")
}

// UninstallRelease 卸载Helm Release
func UninstallRelease(c *gin.Context) {
	releaseName := c.Param("name")
	ns := c.Param("ns")

	// 检查权限
	_, _, err := handleCommonLogic(c, "UninstallRelease", releaseName, ns, "")
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	h, err := getHelm(c, ns)
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

func BatchUninstallRelease(c *gin.Context) {
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
		h, err := getHelm(c, ns)
		if err != nil {
			amis.WriteJsonError(c, err)
			return
		}

		// 检查权限
		_, _, err = handleCommonLogic(c, "BatchUninstallRelease", name, ns, "")
		if err != nil {
			amis.WriteJsonError(c, err)
			return
		}
		x := h.UninstallRelease(name)
		if x != nil {
			klog.V(6).Infof("batch remove %s/%s error %v", ns, name, x)
			err = x
		}
	}

	amis.WriteJsonOK(c)
}

// UpgradeRelease 升级Helm Release
func UpgradeRelease(c *gin.Context) {

	var req struct {
		ReleaseName string `json:"release_name,omitempty"`
		RepoName    string `json:"repo_name,omitempty"`
		Version     string `json:"version,omitempty"`
		Values      string `json:"values,omitempty"`
		Namespace   string `json:"namespace,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 检查权限
	_, _, err := handleCommonLogic(c, "UpgradeRelease", req.ReleaseName, req.Namespace, req.RepoName)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	h, err := getHelm(c, req.Namespace)

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	if err := h.UpgradeRelease(req.ReleaseName, req.RepoName, req.Version, req.Values); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}
