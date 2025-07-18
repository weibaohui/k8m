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

// ListReleaseHistory 获取Release的历史版本
func ListReleaseHistory(c *gin.Context) {
	releaseName := c.Param("name")
	ns := c.Param("ns")
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
func ListRelease(c *gin.Context) {
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

// UninstallRelease 卸载Helm Release
func UninstallRelease(c *gin.Context) {
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

// GetReleaseNote 获取ReleaseNote
func GetReleaseNote(c *gin.Context) {
	releaseName := c.Param("name")
	ns := c.Param("ns")
	revision := c.Param("revision")

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

// GetReleaseValues 获取安装yaml
func GetReleaseValues(c *gin.Context) {
	releaseName := c.Param("name")
	ns := c.Param("ns")
	revision := c.Param("revision")

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

// UpgradeRelease 升级Helm Release
func UpgradeRelease(c *gin.Context) {

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
