package helm

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/weibaohui/k8m/pkg/flag"
	"github.com/weibaohui/k8m/pkg/service"
	"gorm.io/gorm"
	"k8s.io/klog/v2"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/models"
	"sigs.k8s.io/yaml"
)

// HelmCmd 通过调用 helm 二进制命令实现 Helm 接口
// 注意：部分方法需解析 helm 命令输出，部分功能可能与 SDK 实现略有差异

type HelmCmd struct {
	HelmBin        string // helm 二进制路径，默认 "helm"
	repoCacheDir   string //
	clusterID      string
	kubeconfigPath string
	token          string
	caFile         string
	apiServer      string
	cluster        *service.ClusterConfig
}

// NewBackgroundHelmCmd 独立后台执行Helm命令
func NewBackgroundHelmCmd(helmBin string) *HelmCmd {
	if helmBin == "" {
		helmBin = "helm"
	}

	cfg := flag.Init()

	h := &HelmCmd{
		HelmBin:      helmBin,
		repoCacheDir: cfg.HelmCachePath,
	}

	// 确保目录存在
	if err := os.MkdirAll(h.repoCacheDir, 0755); err != nil {
		klog.V(6).Infof("[helm-cmd] warn: create repo cache dir failed: %v", err)
	}
	return h
}
func NewHelmCmd(helmBin string, clusterID string, cluster *service.ClusterConfig) *HelmCmd {

	if helmBin == "" {
		helmBin = "helm"
	}

	cfg := flag.Init()

	h := &HelmCmd{
		HelmBin:      helmBin,
		repoCacheDir: cfg.HelmCachePath,
		clusterID:    clusterID,
		cluster:      cluster,
	}

	// 确保目录存在
	if err := os.MkdirAll(h.repoCacheDir, 0755); err != nil {
		klog.V(6).Infof("[helm-cmd] warn: create repo cache dir failed: %v", err)
	}
	// 将kubeconfig 字符串 存放到临时目录
	// 每次都固定格式，<cluster_name>-kubeconfig.yaml
	encodedClusterID := base64.URLEncoding.EncodeToString([]byte(clusterID))
	kubeconfigPath := fmt.Sprintf("%s/%s-kubeconfig.yaml", h.repoCacheDir, encodedClusterID)
	kubeconfig := cluster.GetKubeconfig()
	if err := os.WriteFile(kubeconfigPath, []byte(kubeconfig), 0644); err != nil {
		klog.V(6).Infof("[helm-cmd] warn: write kubeconfig to file failed: %v", err)
	}
	h.kubeconfigPath = kubeconfigPath

	if cluster.IsInCluster {
		h.fillK8sToken()
	}
	return h
}

// runAndLog 执行 helm 命令并输出日志，支持 shell 特性和可选 stdin
func (h *HelmCmd) runAndLog(args []string, stdin string) ([]byte, error) {

	if h.cluster != nil && h.cluster.IsInCluster {
		accessArgs := []string{
			"--kube-token", h.token,
			"--kube-apiserver", h.apiServer,
			"--kube-ca-file", h.caFile,
		}
		// 参数只有一个，直接追加访问参数
		args = append(args, accessArgs...)
	}
	// 最后一个参数都加上 2>/dev/null
	args = append(args, "2>/dev/null")

	cmdStr := h.HelmBin + " " + strings.Join(args, " ")
	klog.V(6).Infof("[helm-cmd] exec: %s\n", cmdStr)

	cmd := exec.Command("sh", "-c", cmdStr)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("%s=%s", "HELM_CACHE_HOME", h.repoCacheDir),
	)

	if h.cluster != nil && !h.cluster.IsInCluster {
		// 不在集群内
		cmd.Env = append(cmd.Env,
			fmt.Sprintf("%s=%s", "KUBECONFIG", h.kubeconfigPath),
		)
	}

	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if stdout.Len() > 0 {
		klog.V(6).Infof("[helm-cmd] stdout: %s\n", stdout.String())
	}
	if stderr.Len() > 0 {
		klog.V(6).Infof("[helm-cmd] stderr: %s\n", stderr.String())
		err = fmt.Errorf("%s", stderr.String())
	}
	return stdout.Bytes(), err
}

// AddOrUpdateRepo 添加或更新 Helm 仓库
func (h *HelmCmd) AddOrUpdateRepo(helmRepo *models.HelmRepository) error {
	// 1. 先执行数据库操作，保存 HelmRepository 信息
	// 2. 再执行 helm repo add/update

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
	args := []string{"repo", "add", helmRepo.Name, helmRepo.URL}
	if helmRepo.Username != "" {
		args = append(args, "--username", helmRepo.Username)
	}
	if helmRepo.Password != "" {
		args = append(args, "--password", helmRepo.Password)
	}
	out, err := h.runAndLog(args, "")
	if err != nil && !strings.Contains(string(out), "already exists") {
		return fmt.Errorf("helm repo add failed: %v, output: %s", err, string(out))
	}
	_, err = h.updateRepoByName(helmRepo)
	if err != nil {
		return err
	}
	return nil
}

func (h *HelmCmd) updateRepoByName(helmRepo *models.HelmRepository) (bool, error) {
	// 4. helm repo update
	out, err := h.runAndLog([]string{"repo", "update", helmRepo.Name}, "")
	if err != nil {
		return false, fmt.Errorf("helm repo update failed: %v, output: %s", err, string(out))
	}

	// 5. helm repo index 文件分析，记录所有chart到数据库
	cachePath := fmt.Sprintf("%s/repository/%s-index.yaml", h.repoCacheDir, helmRepo.Name)
	indexData, err := os.ReadFile(cachePath)
	if err != nil {
		klog.V(6).Infof("[helm-cmd] warn: read repo index file failed: %v\n", err)
		return true, nil // 不阻断主流程
	}
	var index IndexFile

	if err := yaml.Unmarshal(indexData, &index); err != nil {
		klog.V(6).Infof("[helm-cmd] warn: unmarshal repo index file failed: %v\n", err)
		return true, nil
	}
	// 清空数据库中对应的chart repo
	dao.DB().Where("repository_id = ?", helmRepo.ID).Delete(models.HelmChart{})
	for chartName, versionList := range index.Entries {
		if len(versionList) == 0 {
			continue
		}
		slice.SortBy(versionList, func(a *ChartVersion, b *ChartVersion) bool {
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
				klog.V(6).Infof("[helm-cmd] warn: save helm chart to database error: %v\n", err)
			}
		}
	}

	if !index.Generated.IsZero() {
		// 更新索引时间
		helmRepo.Generated = index.Generated.Format(time.DateTime)
		_ = helmRepo.Save(nil)
	}

	return false, nil
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
	// 删除数据库HelmRelease
	if err == nil {
		_ = models.DeleteHelmReleaseByNsAndReleaseName(namespace, releaseName, h.clusterID) // 忽略错误
	}
	if err != nil {
		return fmt.Errorf("helm uninstall failed: %v, output: %s", err, string(out))
	}
	return nil
}

func (h *HelmCmd) InstallRelease(namespace, releaseName, repoName, chartName, version string, values ...string) error {
	chartRef := fmt.Sprintf("%s/%s", repoName, chartName)
	args := []string{"install", releaseName, chartRef, "--namespace", namespace, "--version", version, "--create-namespace"}
	stdin := ""
	if len(values) > 0 && values[0] != "" {
		args = append(args, "-f", "-")
		stdin = values[0]
	}
	ob, _ := h.runAndLog(args, stdin)

	// 安装成功后记录到数据库
	release := &models.HelmRelease{
		Cluster:      h.clusterID,
		ReleaseName:  releaseName,
		RepoName:     repoName,
		Namespace:    namespace,
		ChartName:    chartName,
		ChartVersion: version,
		Values:       stdin,
		Status:       "installed",
		Result:       string(ob),
	}

	_ = release.Save(nil) // 忽略错误，防止影响主流程

	return nil
}
func (h *HelmCmd) UpgradeRelease(ns, name string, values ...string) error {
	// helm upgrade <release-name> <chart-path-or-name>
	hr, err := models.GetHelmReleaseByNsAndReleaseName(ns, name, h.clusterID)
	if err != nil {
		return fmt.Errorf("get repoName from db failed: %v", err)
	}
	args := []string{"upgrade", name, fmt.Sprintf("%s/%s", hr.RepoName, hr.ChartName), "--namespace", ns, "--version", hr.ChartVersion}
	stdin := ""
	if len(values) > 0 && values[0] != "" {
		args = append(args, "-f", "-")
		stdin = values[0]
	}
	out, err := h.runAndLog(args, stdin)

	// 更新values、result
	hr.Result = string(out)
	hr.Values = stdin
	_ = hr.Save(nil) // 忽略错误，防止影响主流程

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
	args := []string{"search", "repo", chartRef, "-o", "json", "--versions"}
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
		_, _ = h.updateRepoByName(item)
	}

}
func (h *HelmCmd) UpdateAllReposIndex() {
	m := models.HelmRepository{}
	list, _, err := m.List(nil)
	if err != nil {
		klog.V(6).Infof("get helm repository list error: %v", err)
		return
	}
	for _, item := range list {
		_, _ = h.updateRepoByName(item)
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

func (h *HelmCmd) GetRepoList() ([]*RepoVO, error) {
	out, err := h.runAndLog([]string{"repo", "list", "-o", "json"}, "")
	if err != nil {
		return nil, fmt.Errorf("helm repo list failed: %v, output: %s", err, string(out))
	}
	var list []*RepoVO
	if err := json.Unmarshal(out, &list); err != nil {
		return nil, fmt.Errorf("unmarshal helm repo list output failed: %v, output: %s", err, string(out))
	}
	return list, nil
}
func (h *HelmCmd) GetReleaseNote(ns string, name string) (string, error) {
	out, err := h.runAndLog([]string{"get", "notes", name, "-n", ns}, "")
	if err != nil {
		return "", fmt.Errorf("helm get  notes failed: %v, output: %s", err, string(out))
	}
	return string(out), nil
}
func (h *HelmCmd) GetReleaseNoteWithRevision(ns string, name string, revision string) (string, error) {
	out, err := h.runAndLog([]string{"get", "notes", name, "-n", ns, "--revision", revision}, "")
	if err != nil {
		return "", fmt.Errorf("helm get  notes failed: %v, output: %s", err, string(out))
	}
	return string(out), nil
}
func (h *HelmCmd) GetReleaseValues(ns string, name string) (string, error) {
	out, err := h.runAndLog([]string{"get", "values", name, "-n", ns, "--all", "-o", "yaml"}, "")
	if err != nil {
		return "", fmt.Errorf("helm get values failed: %v, output: %s", err, string(out))
	}
	return string(out), nil
}
func (h *HelmCmd) GetReleaseValuesWithRevision(ns string, name string, revision string) (string, error) {
	out, err := h.runAndLog([]string{"get", "values", name, "-n", ns, "--revision", revision, "-o", "yaml"}, "")
	if err != nil {
		return "", fmt.Errorf("helm get values failed: %v, output: %s", err, string(out))
	}
	return string(out), nil
}

func (h *HelmCmd) RemoveRepo(repoName string) error {
	out, err := h.runAndLog([]string{"repo", "remove", repoName}, "")
	if err != nil {
		return fmt.Errorf("helm remove repo failed: %v, output: %s", err, string(out))
	}
	return nil
}

// InCluster 模式填充参数
func (h *HelmCmd) fillK8sToken() {
	// 获取 ServiceAccount token
	tokenBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		klog.V(6).Infof("failed to read token: %v", err)
	}
	token := strings.TrimSpace(string(tokenBytes))

	// 获取 CA 证书路径
	caFile := "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"

	// 获取 API Server 地址
	host := os.Getenv("KUBERNETES_SERVICE_HOST")
	port := os.Getenv("KUBERNETES_SERVICE_PORT")
	apiServer := fmt.Sprintf("https://%s:%s", host, port)

	// "helm", "list",
	// 	"--kube-token", token,
	// 	"--kube-apiserver", apiServer,
	// 	"--kube-ca-file", caFile,
	// 	"--namespace", "default",
	h.token = token
	h.caFile = caFile
	h.apiServer = apiServer
}

// ReAddMissingRepo 重新添加丢失的Repo，比如在容器环境中重启了。
func (h *HelmCmd) ReAddMissingRepo() {
	m := models.HelmRepository{}
	list, _, err := m.List(nil)
	if err != nil {
		klog.V(6).Infof("get helm repository list error: %v", err)
		return
	}
	repoList, err := h.GetRepoList()
	if err != nil {
		klog.V(6).Infof("get helm repository list error: %v", err)
		return
	}
	var repos []string
	for _, vo := range repoList {
		repos = append(repos, vo.Name)
	}
	for _, item := range list {
		if !slice.Contain(repos, item.Name) {
			klog.V(6).Infof("helm repository adding %s", item.Name)
			_ = h.AddOrUpdateRepo(item)
		}
	}
}
