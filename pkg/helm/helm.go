package helm

import (
	"github.com/weibaohui/k8m/pkg/models"
	"helm.sh/helm/v3/pkg/repo"
)

type Helm interface {
	AddOrUpdateRepo(repoEntry *repo.Entry) error
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
}
