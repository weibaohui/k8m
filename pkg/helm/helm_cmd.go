package helm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"gorm.io/gorm"
	"helm.sh/helm/v3/pkg/repo"
	"k8s.io/klog/v2"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/models"
	"sigs.k8s.io/yaml"
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

// runAndLog 执行 helm 命令并输出日志，支持 shell 特性和可选 stdin
func (h *HelmCmd) runAndLog(args []string, stdin string) ([]byte, error) {
	cmdStr := h.HelmBin + " " + strings.Join(args, " ")
	fmt.Printf("[helm-cmd] exec: %s\n", cmdStr)
	// cmd := exec.Command(h.HelmBin, args...)

	cmd := exec.Command("sh", "-c", cmdStr)
	cmd.Env = []string{fmt.Sprintf("%s=%s", "HELM_CACHE_HOME", h.repoCacheDir)}

	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if stdout.Len() > 0 {
		fmt.Printf("[helm-cmd] stdout: %s\n", stdout.String())
	}
	if stderr.Len() > 0 {
		fmt.Printf("[helm-cmd] stderr: %s\n", stderr.String())
		err = fmt.Errorf(stderr.String())
	}
	return stdout.Bytes(), err
}

// AddOrUpdateRepo 添加或更新 Helm 仓库
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
	args := []string{"repo", "add", repoEntry.Name, repoEntry.URL}
	if repoEntry.Username != "" {
		args = append(args, "--username", repoEntry.Username)
	}
	if repoEntry.Password != "" {
		args = append(args, "--password", repoEntry.Password)
	}
	out, err := h.runAndLog(args, "")
	if err != nil && !strings.Contains(string(out), "already exists") {
		return fmt.Errorf("helm repo add failed: %v, output: %s", err, string(out))
	}
	_, err = h.updateRepoByName(repoEntry, helmRepo)
	if err != nil {
		return err
	}
	return nil
}

func (h *HelmCmd) updateRepoByName(repoEntry *repo.Entry, helmRepo *models.HelmRepository) (bool, error) {
	// 4. helm repo update
	out, err := h.runAndLog([]string{"repo", "update", repoEntry.Name}, "")
	if err != nil {
		return false, fmt.Errorf("helm repo update failed: %v, output: %s", err, string(out))
	}

	// 5. helm repo index 文件分析，记录所有chart到数据库
	cachePath := fmt.Sprintf("%s/repository/%s-index.yaml", h.repoCacheDir, repoEntry.Name)
	indexData, err := os.ReadFile(cachePath)
	if err != nil {
		fmt.Printf("[helm-cmd] warn: read repo index file failed: %v\n", err)
		return true, nil // 不阻断主流程
	}
	var index repo.IndexFile

	if err := yaml.Unmarshal(indexData, &index); err != nil {
		fmt.Printf("[helm-cmd] warn: unmarshal repo index file failed: %v\n", err)
		return true, nil
	}
	// 清空数据库中对应的chart repo
	dao.DB().Where("repository_id = ?", helmRepo.ID).Delete(models.HelmChart{})
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
	return false, nil
}

func getHomeDir() string {
	home, _ := os.UserHomeDir()
	return home
}

func (h *HelmCmd) GetReleaseHistory(namespace string, releaseName string) ([]*models.ReleaseHistory, error) {
	out, err := h.runAndLog([]string{"history", releaseName, "-n", namespace, "-o", "json"}, "")
	if err != nil {
		return nil, fmt.Errorf("helm history failed: %v, output: %s", err, string(out))
	}
	var releases []*models.ReleaseHistory
	if err := json.Unmarshal(out, &releases); err != nil {
		return nil, fmt.Errorf("unmarshal helm history output failed: %v, output: %s", err, string(out))
	}
	return releases, nil
}

func (h *HelmCmd) UninstallRelease(namespace string, releaseName string) error {
	out, err := h.runAndLog([]string{"uninstall", releaseName, "-n", namespace}, "")
	if err != nil {
		return fmt.Errorf("helm uninstall failed: %v, output: %s", err, string(out))
	}
	return nil
}

func (h *HelmCmd) InstallRelease(namespace, releaseName, repoName, chartName, version string, values ...string) error {
	chartRef := fmt.Sprintf("%s/%s", repoName, chartName)
	args := []string{"install", releaseName, chartRef, "--namespace", namespace, "--version", version, "--create-namespace"}
	klog.Infof("安装参数 %d =\n %s", len(values), values)
	stdin := ""
	if len(values) > 0 && values[0] != "" {
		args = append(args, "-f", "-")
		stdin = values[0]
	}
	_, err := h.runAndLog(args, stdin)
	if err != nil {
		return err
	}
	return nil
}
func (h *HelmCmd) UpgradeRelease(releaseName, repoName, targetVersion string, values ...string) error {
	chartRef := fmt.Sprintf("%s/%s", repoName, releaseName)
	args := []string{"upgrade", releaseName, chartRef, "--version", targetVersion}
	stdin := ""
	if len(values) > 0 && values[0] != "" {
		args = append(args, "-f", "-")
		stdin = values[0]
	}
	out, err := h.runAndLog(args, stdin)
	if err != nil {
		return fmt.Errorf("helm upgrade failed: %v, output: %s", err, string(out))
	}
	return nil
}

func (h *HelmCmd) GetChartValue(repoName, chartName, version string) (string, error) {
	chartRef := fmt.Sprintf("%s/%s", repoName, chartName)
	args := []string{"show", "values", chartRef, "--version", version}
	out, err := h.runAndLog(args, "")
	if err != nil {
		return "", fmt.Errorf("helm show values failed: %v, output: %s", err, string(out))
	}
	return string(out), nil
}

func (h *HelmCmd) GetChartVersions(repoName string, chartName string) ([]string, error) {
	chartRef := fmt.Sprintf("%s/%s", repoName, chartName)
	args := []string{"search", "repo", chartRef, "-o", "json", "--versions", "2>/dev/null"}
	out, err := h.runAndLog(args, "")
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
	klog.Infof("helm search result: %v", versions)
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

func (h *HelmCmd) GetReleaseList() ([]*models.Release, error) {
	out, err := h.runAndLog([]string{"list", "-A", "-o", "json"}, "")
	if err != nil {
		return nil, fmt.Errorf("helm list failed: %v, output: %s", err, string(out))
	}
	var releases []*models.Release
	if err := json.Unmarshal(out, &releases); err != nil {
		return nil, fmt.Errorf("unmarshal helm list output failed: %v, output: %s", err, string(out))
	}
	return releases, nil
}
func (h *HelmCmd) GetReleaseNote(ns string, name string) (string, error) {
	out, err := h.runAndLog([]string{"get", "notes", name, "-n", ns, "-o", "json"}, "")
	if err != nil {
		return "", fmt.Errorf("helm get  notes failed: %v, output: %s", err, string(out))
	}
	return string(out), nil
}
func (h *HelmCmd) GetReleaseValues(ns string, name string) (string, error) {
	out, err := h.runAndLog([]string{"get", "values", name, "-n", ns, "-o", "json"}, "")
	if err != nil {
		return "", fmt.Errorf("helm get values failed: %v, output: %s", err, string(out))
	}
	return string(out), nil
}
func (h *HelmCmd) GetReleaseValuesWithRevision(ns string, name string, revision string) (string, error) {
	out, err := h.runAndLog([]string{"get", "values", name, "-n", ns, "--revision", revision, "-o", "json"}, "")
	if err != nil {
		return "", fmt.Errorf("helm get values failed: %v, output: %s", err, string(out))
	}
	return string(out), nil
}
