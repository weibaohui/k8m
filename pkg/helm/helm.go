package helm

import (
	"github.com/weibaohui/k8m/pkg/models"
	"helm.sh/helm/v3/pkg/repo"
)
 
type Helm interface {
	AddOrUpdateRepo(repoEntry *repo.Entry) error
	GetReleaseHistory(namespace string, releaseName string) ([]*models.ReleaseHistory, error)
	InstallRelease(namespace, releaseName, repoName, chartName, version string, values ...string) error
	UninstallRelease(namespace string, releaseName string) error
	UpgradeRelease(ns string, name string, values ...string) error
	GetChartValue(repoName, chartName, version string) (string, error)
	GetChartVersions(repoName string, chartName string) ([]string, error)
	UpdateReposIndex(ids string)
	GetReleaseList() ([]*models.Release, error)
	GetReleaseNote(ns string, name string) (string, error)
	GetReleaseValues(ns string, name string) (string, error)
	GetReleaseValuesWithRevision(ns string, name string, revision string) (string, error)
}
