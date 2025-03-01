package helm

import (
	"strings"

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
	releaseName := c.Param("release")
	repoName := c.Param("repo")
	chartName := c.Param("chart")
	version := c.Param("version")

	var req struct {
		Values string `json:"values,omitempty"`
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

	// values 中有#注释
	// 逐行对values 进行处理，去掉#注释
	values := strings.Split(req.Values, "\n")
	for i, v := range values {
		if strings.HasPrefix(strings.TrimSpace(v), "#") {
			values[i] = ""
		}
	}
	// 删除空行
	values = utils.RemoveEmptyLines(values)

	req.Values = strings.Join(values, "\n")
	klog.V(0).Infof("values: \n%s", req.Values)
	if err = h.InstallRelease(releaseName, repoName, chartName, version, req.Values); err != nil {
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
