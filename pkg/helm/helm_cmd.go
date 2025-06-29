package helm

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"gorm.io/gorm"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
	"k8s.io/klog/v2"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/models"
	"gopkg.in/yaml.v2"
)

// HelmCmd 通过调用 helm 二进制命令实现 Helm 接口
// 注意：部分方法需解析 helm 命令输出，部分功能可能与 SDK 实现略有差异

type HelmCmd struct {
	HelmBin      string // helm 二进制路径，默认 "helm"
	repoCacheDir string
}

func NewHelmCmd(helmBin string) *HelmCmd {
	if helmBin == "" {
		helmBin = "helm"
	}
	homeDir := getHomeDir()
	repoCacheDir := fmt.Sprintf("%s/.cache/helm", homeDir)

	return &HelmCmd{HelmBin: helmBin, repoCacheDir: repoCacheDir}
}

func (h *HelmCmd) runAndLog(cmd *exec.Cmd) ([]byte, error) {
	fmt.Printf("[helm-cmd] exec: %s\n", strings.Join(cmd.Args, " "))
	out, err := cmd.CombinedOutput()
	fmt.Printf("[helm-cmd] result: %s\n", string(out))
	return out, err
}

func (h *HelmCmd) AddOrUpdateRepo(repoEntry *repo.Entry) error {
	// 1. 先执行数据库操作，保存 HelmRepository 信息
	// 2. 再执行 helm repo add/update

	// 创建HelmRepository对象
	helmRepo := &models.HelmRepository{
		Name:     repoEntry.Name,
		URL:      repoEntry.URL,
		Username: repoEntry.Username,
		Password: repoEntry.Password,
	}
	// 判断该名称、URL的仓库是否存在
	if id, err := helmRepo.GetIDByNameAndURL(nil); err == nil && id > 0 {
		helmRepo.ID = id
	} else {
		// 第一次创建，先保存到数据库
		if err = helmRepo.Save(nil); err != nil {
			return fmt.Errorf("save helm repository to database error: %v", err)
		}
	}

	// 3. helm repo add
	// 设置 HELM_CACHE_HOME 环境变量，保证 index 文件写入到指定目录

	os.Setenv("HELM_CACHE_HOME", h.repoCacheDir)

	args := []string{"repo", "add", repoEntry.Name, repoEntry.URL}
	if repoEntry.Username != "" {
		args = append(args, "--username", repoEntry.Username)
	}
	if repoEntry.Password != "" {
		args = append(args, "--password", repoEntry.Password)
	}
	cmd := exec.Command(h.HelmBin, args...)
	cmd.Env = append(os.Environ(), "HELM_CACHE_HOME="+h.repoCacheDir)
	out, err := h.runAndLog(cmd)
	if err != nil && !strings.Contains(string(out), "already exists") {
		return fmt.Errorf("helm repo add failed: %v, output: %s", err, string(out))
	}
	err2, done := h.updateRepoByName(repoEntry, helmRepo)
	if done {
		return err2
	}
	return nil
}

func (h *HelmCmd) updateRepoByName(repoEntry *repo.Entry, helmRepo *models.HelmRepository) (error, bool) {
	// 4. helm repo update
	cmd := exec.Command(h.HelmBin, "repo", "update", repoEntry.Name)
	cmd.Env = append(os.Environ(), "HELM_CACHE_HOME="+h.repoCacheDir)
	out, err := h.runAndLog(cmd)
	if err != nil {
		return fmt.Errorf("helm repo update failed: %v, output: %s", err, string(out)), true
	}

	// 5. helm repo index 文件分析，记录所有chart到数据库
	cachePath := fmt.Sprintf("%s/repository/%s-index.yaml", h.repoCacheDir, repoEntry.Name)
	indexData, err := os.ReadFile(cachePath)
	if err != nil {
		fmt.Printf("[helm-cmd] warn: read repo index file failed: %v\n", err)
		return nil, true // 不阻断主流程
	}
	var index repo.IndexFile
	if err := yaml.Unmarshal(indexData, &index); err != nil {
		fmt.Printf("[helm-cmd] warn: unmarshal repo index file failed: %v\n", err)
		return nil, true
	}
	// 清空数据库中对应的chart repo
	dao.DB().Where("repository_name = ?", repoEntry.Name).Delete(models.HelmChart{})
	for chartName, versionList := range index.Entries {
		if len(versionList) == 0 {
			continue
		}
		slice.SortBy(versionList, func(a *repo.ChartVersion, b *repo.ChartVersion) bool {
			return a.Created.After(b.Created)
		})
		ct := versionList[0]
		if ct.Metadata != nil {
			m := models.HelmChart{
				RepositoryID:   helmRepo.ID,
				RepositoryName: helmRepo.Name,
				Name:           chartName,
			}
			if ct.Version != "" {
				m.LatestVersion = ct.Version
			}
			if ct.Description != "" {
				m.Description = ct.Description
			}
			if ct.Home != "" {
				m.Home = ct.Home
			}
			if ct.Icon != "" {
				m.Icon = ct.Icon
			}
			if ct.KubeVersion != "" {
				m.KubeVersion = ct.KubeVersion
			}
			if ct.AppVersion != "" {
				m.AppVersion = ct.AppVersion
			}
			m.Deprecated = ct.Deprecated
			if len(ct.Keywords) > 0 {
				m.Keywords = strings.Join(ct.Keywords, ",")
			}
			if len(ct.Sources) > 0 {
				m.Sources = ct.Sources[0]
			}
			err = m.Save(nil)
			if err != nil {
				fmt.Printf("[helm-cmd] warn: save helm chart to database error: %v\n", err)
			}
		}
	}
	return nil, false
}

func getHomeDir() string {
	home, _ := os.UserHomeDir()
	return home
}

func (h *HelmCmd) GetReleaseHistory(releaseName string) ([]*release.Release, error) {
	cmd := exec.Command(h.HelmBin, "history", releaseName, "-o", "json")
	out, err := h.runAndLog(cmd)
	if err != nil {
		return nil, fmt.Errorf("helm history failed: %v, output: %s", err, string(out))
	}
	var releases []*release.Release
	if err := json.Unmarshal(out, &releases); err != nil {
		return nil, fmt.Errorf("unmarshal helm history output failed: %v, output: %s", err, string(out))
	}
	return releases, nil
}

func (h *HelmCmd) InstallRelease(namespace, releaseName, repoName, chartName, version string, values ...string) error {
	chartRef := fmt.Sprintf("%s/%s", repoName, chartName)
	args := []string{"install", releaseName, chartRef, "--namespace", namespace, "--version", version, "--create-namespace"}
	if len(values) > 0 && values[0] != "" {
		args = append(args, "-f", "-")
		cmd := exec.Command(h.HelmBin, args...)
		cmd.Stdin = strings.NewReader(values[0])
		out, err := h.runAndLog(cmd)
		if err != nil {
			return fmt.Errorf("helm install failed: %v, output: %s", err, string(out))
		}
		return nil
	}
	cmd := exec.Command(h.HelmBin, args...)
	out, err := h.runAndLog(cmd)
	if err != nil {
		return fmt.Errorf("helm install failed: %v, output: %s", err, string(out))
	}
	return nil
}

func (h *HelmCmd) UninstallRelease(releaseName string) error {
	cmd := exec.Command(h.HelmBin, "uninstall", releaseName)
	out, err := h.runAndLog(cmd)
	if err != nil {
		return fmt.Errorf("helm uninstall failed: %v, output: %s", err, string(out))
	}
	return nil
}

func (h *HelmCmd) UpgradeRelease(releaseName, repoName, targetVersion string, values ...string) error {
	chartRef := fmt.Sprintf("%s/%s", repoName, releaseName)
	args := []string{"upgrade", releaseName, chartRef, "--version", targetVersion}
	if len(values) > 0 && values[0] != "" {
		args = append(args, "-f", "-")
		cmd := exec.Command(h.HelmBin, args...)
		cmd.Stdin = strings.NewReader(values[0])
		out, err := h.runAndLog(cmd)
		if err != nil {
			return fmt.Errorf("helm upgrade failed: %v, output: %s", err, string(out))
		}
		return nil
	}
	cmd := exec.Command(h.HelmBin, args...)
	out, err := h.runAndLog(cmd)
	if err != nil {
		return fmt.Errorf("helm upgrade failed: %v, output: %s", err, string(out))
	}
	return nil
}

func (h *HelmCmd) GetChartValue(repoName, chartName, version string) (string, error) {
	// helm show values repo/chart --version x.x.x
	chartRef := fmt.Sprintf("%s/%s", repoName, chartName)
	args := []string{"show", "values", chartRef, "--version", version}
	cmd := exec.Command(h.HelmBin, args...)
	out, err := h.runAndLog(cmd)
	if err != nil {
		return "", fmt.Errorf("helm show values failed: %v, output: %s", err, string(out))
	}
	return string(out), nil
}

func (h *HelmCmd) GetChartVersions(repoName string, chartName string) ([]string, error) {
	// helm search repo repo/chart -o json
	chartRef := fmt.Sprintf("%s/%s", repoName, chartName)
	args := []string{"search", "repo", chartRef, "-o", "json"}
	cmd := exec.Command(h.HelmBin, args...)
	out, err := h.runAndLog(cmd)
	if err != nil {
		return nil, fmt.Errorf("helm search repo failed: %v, output: %s", err, string(out))
	}
	var result []struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("unmarshal helm search output failed: %v, output: %s", err, string(out))
	}
	var versions []string
	for _, r := range result {
		versions = append(versions, r.Version)
	}
	return versions, nil
}

func (h *HelmCmd) UpdateReposIndex(ids string) {
	// 解析ids为数组
	idsArray := strings.Split(ids, ",")
	m := models.HelmRepository{}
	list, _, err := m.List(nil, func(db *gorm.DB) *gorm.DB {
		return db.Where("id in ?", idsArray).Find(&m)
	})
	if err != nil {
		klog.V(6).Infof("get helm repository list error: %v", err)
		return
	}

	// 遍历ids，更新每个repo的index

	for _, item := range list {

		repoEntry := &repo.Entry{
			Name:                  item.Name,
			URL:                   item.URL,
			Username:              item.Username,
			Password:              item.Password,
			CAFile:                item.CAFile,
			CertFile:              item.CertFile,
			KeyFile:               item.KeyFile,
			InsecureSkipTLSverify: item.InsecureSkipTLSverify,
			PassCredentialsAll:    item.PassCredentialsAll,
		}
		h.updateRepoByName(repoEntry, item)

	}

}

func (h *HelmCmd) GetReleaseList() ([]*release.Release, error) {
	cmd := exec.Command(h.HelmBin, "list", "-A", "-o", "json")
	out, err := h.runAndLog(cmd)
	if err != nil {
		return nil, fmt.Errorf("helm list failed: %v, output: %s", err, string(out))
	}
	var releases []*release.Release
	if err := json.Unmarshal(out, &releases); err != nil {
		return nil, fmt.Errorf("unmarshal helm list output failed: %v, output: %s", err, string(out))
	}
	return releases, nil
}
