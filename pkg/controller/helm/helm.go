package helm

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/helm"
	"github.com/weibaohui/k8m/pkg/service"
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
	amis.WriteJsonData(c, list)
}

func getHelm(c *gin.Context, namespace string) (helm.Helm, error) {
	// if namespace == "" {
	// 	namespace = "default"
	// }
	selectedCluster := amis.GetSelectedCluster(c)
	restConfig := service.ClusterService().GetClusterByID(selectedCluster).GetRestConfig()
	h, err := helm.New(restConfig, namespace)
	return h, err
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

	h, err := getHelm(c, req.Namespace)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	klog.V(0).Infof("values: \n%s", req.Values)

	if releaseName == "" {
		releaseName = fmt.Sprintf("%s-%d", chartName, utils.RandNDigitInt(8))
	}
	if err = h.InstallRelease(req.Namespace, releaseName, repoName, chartName, version, req.Values); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

// UninstallRelease 卸载Helm Release
func UninstallRelease(c *gin.Context) {
	releaseName := c.Param("name")
	ns := c.Param("ns")
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

// UpgradeRelease 升级Helm Release
func UpgradeRelease(c *gin.Context) {
	ns := c.Param("ns")

	var req struct {
		ReleaseName string `json:"release_name"`
		RepoName    string `json:"repo_name"`
		Version     string `json:"version"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	h, err := getHelm(c, ns)

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
