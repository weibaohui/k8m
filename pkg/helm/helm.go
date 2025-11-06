package helm

import (
	"sync"

	"github.com/robfig/cron/v3"
	"github.com/weibaohui/k8m/pkg/flag"
	"github.com/weibaohui/k8m/pkg/models"
	"k8s.io/klog/v2"
)

// 添加全局变量来存储 cron 实例和 mutex
var (
	helmCron *cron.Cron
	helmMu   sync.Mutex
)

type Helm interface {
	AddOrUpdateRepo(repo *models.HelmRepository) error
	GetReleaseHistory(ns, releaseName string) ([]*models.ReleaseHistory, error)
	InstallRelease(ns, releaseName, repoName, chartName, version string, values ...string) error
	UninstallRelease(ns, releaseName string) error
	UpgradeRelease(ns, name string, values ...string) error
	GetChartValue(repoName, chartName, version string) (string, error)
	GetChartVersions(repoName, chartName string) ([]string, error)
	UpdateReposIndex(ids string)
	GetReleaseList() ([]*models.Release, error)
	GetReleaseNote(ns, name string) (string, error)
	GetReleaseNoteWithRevision(ns, name, revision string) (string, error)
	GetReleaseValues(ns, name string) (string, error)
	GetReleaseValuesWithRevision(ns, name, revision string) (string, error)
	RemoveRepo(repoName string) error
	GetRepoList() ([]*RepoVO, error)
}

func StartUpdateHelmRepoInBackground() {
	cfg := flag.Init()
	cn := cfg.HelmUpdateCron

	if cn == "" {
		klog.V(6).Infof(" HelmUpdateCron 表达式 为空，跳过定时任务")
		return
	}
	if _, err := cron.ParseStandard(cfg.HelmUpdateCron); err != nil {
		klog.Errorf("非法的 HelmUpdateCron 表达式 %q: %v", cfg.HelmUpdateCron, err)
		return
	}

	helmMu.Lock()
	defer helmMu.Unlock()

	// 如果已经有 cron 实例在运行，先停止它
	if helmCron != nil {
		helmCron.Stop()
	}

	inst := cron.New()
	_, err := inst.AddFunc(cn, func() {
		h := NewBackgroundHelmCmd("helm")
		h.ReAddMissingRepo()
		h.UpdateAllReposIndex()
	})
	if err != nil {
		klog.Errorf("新增Helm更新定时任务失败: %v", err)
		return
	}
	helmCron = inst
	inst.Start()
	klog.V(6).Infof("新增 Helm 更新定时任务 %s", cn)
}

// StopUpdateHelmRepoInBackground 停止 Helm 更新定时任务
func StopUpdateHelmRepoInBackground() {
	helmMu.Lock()
	defer helmMu.Unlock()

	if helmCron != nil {
		klog.V(6).Infof("停止 Helm 更新定时任务")
		helmCron.Stop()
		helmCron = nil
	}
}
