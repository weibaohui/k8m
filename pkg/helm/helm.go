package helm

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	clivalues "helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
	"helm.sh/helm/v3/pkg/storage/driver"
	"helm.sh/helm/v3/pkg/strvals"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

type Helm interface {
	AddOrUpdateRepo(repoEntry *repo.Entry) error
	GetReleaseHistory(releaseName string) ([]*release.Release, error)
	InstallRelease(releaseName, chartName, version string, values ...string) error
	UninstallRelease(releaseName string) error
	UpgradeRelease(releaseName, localRepoName, targetVersion string) error
}

type Client struct {
	setting *cli.EnvSettings
	ac      *action.Configuration
	getter  *RESTClientGetterImpl
	driver  string
}

type Option func(client *Client)

// New  Helm Interface
func New(options ...Option) (Helm, error) {
	h := Client{
		setting: cli.New(),
		driver:  "secret",
	}

	for _, op := range options {
		op(&h)
	}

	var ac action.Configuration
	g := h.setting.RESTClientGetter()

	if h.getter != nil {
		g = h.getter
	}

	err := ac.Init(g, h.setting.Namespace(), h.driver, debug)
	if err != nil {
		return nil, err
	}

	h.ac = &ac

	return &h, nil
}

// WithRESTClientGetter with custom rest client getter, use rest.Config to visit Kubernetes
func WithRESTClientGetter(getter *RESTClientGetterImpl) Option {
	return func(client *Client) {
		client.getter = getter
	}
}

// GetReleaseHistory check release installed or not
func (c *Client) GetReleaseHistory(releaseName string) ([]*release.Release, error) {
	klog.V(0).Infof("[%s] get release on target cluster", releaseName)

	// use HELM_NAMESPACE find release
	hc := action.NewHistory(c.ac)

	releases, err := hc.Run(releaseName)
	if err != nil {
		if err == driver.ErrReleaseNotFound {
			return releases, nil
		}
		klog.Errorf("[%s] history client run error: %v", releaseName, err)
		return nil, err
	}

	return releases, nil
}

// InstallRelease install release
func (c *Client) InstallRelease(releaseName, chartName, version string, values ...string) error {
	klog.V(0).Infof("install release, name: %s, version: %s, chartName: %s", releaseName, version, chartName)
	klog.V(0).Infof("helm repository cache path: %s", c.setting.RepositoryCache)

	if res, err := c.GetReleaseHistory(releaseName); err != nil {
		return err
	} else {
		if len(res) != 0 {
			return fmt.Errorf("[%s] release already exist on target cluster, version: %s",
				releaseName, res[len(res)-1].Chart.Metadata.Version)
		}
	}

	ic := action.NewInstall(c.ac)

	ic.ReleaseName = releaseName
	ic.Version = version
	ic.Namespace = "default" // todo 传参

	chartReq, err := c.getChart(chartName, version, &ic.ChartPathOptions)
	if err != nil {
		return fmt.Errorf("[%s] get chart error: %v", releaseName, err)
	}

	var vals map[string]interface{}

	// if values setting, merge values to vals
	if len(values) != 0 {
		cvOptions := &clivalues.Options{}
		vals, err = cvOptions.MergeValues(getter.All(c.setting))
		if err != nil {
			return err
		}

		if err = strvals.ParseInto(values[0], vals); err != nil {
			return err
		}

	}

	if _, err = ic.Run(chartReq, vals); err != nil {
		return fmt.Errorf("[%s] install error: %v", releaseName, err)
	}

	klog.V(0).Infof("[%s] release install success", releaseName)

	return nil
}

// getChart get chart
func (c *Client) getChart(chartName, version string, chartPathOptions *action.ChartPathOptions) (*chart.Chart, error) {
	var (
		lc  *chart.Chart
		err error
	)
	option, err := chartPathOptions.LocateChart(chartName, c.setting)
	if err != nil {
		return nil, fmt.Errorf("located charts %s error: %v", chartName, err)
	}

	lc, err = loader.Load(option)

	if err != nil {
		return nil, fmt.Errorf("load chart path options error: %v", err)
	}

	return lc, nil
}

// UninstallRelease uninstall release which deployed
func (c *Client) UninstallRelease(releaseName string) error {
	// use HELM_NAMESPACE find release
	uc := action.NewUninstall(c.ac)

	resp, err := uc.Run(releaseName)
	if resp != nil {
		klog.V(6).Infof("[%s] uninstall release %+v,response: %v", releaseName, resp.Release, resp.Info)
	}
	if err != nil {
		return fmt.Errorf("[%s] run uninstall client error: %v", releaseName, err)
	}

	klog.V(0).Infof("[%s] uninstall release success", releaseName)

	return nil
}

// UpgradeRelease upgrade release version
func (c *Client) UpgradeRelease(releaseName, localRepoName, targetVersion string) error {
	// use HELM_NAMESPACE find release
	uc := action.NewUpgrade(c.ac)
	r, err := c.GetReleaseHistory(releaseName)
	if err != nil {
		return err
	}

	if len(r) == 0 {
		return fmt.Errorf("[%s] release doesn't install", releaseName)
	}

	if r[len(r)-1].Chart.Metadata.Version == targetVersion {
		return fmt.Errorf("[%s] version %s already installed", releaseName, r[len(r)-1].Chart.Metadata.Version)
	}

	uc.Version = targetVersion

	chartName := fmt.Sprintf("%s/%s", localRepoName, r[len(r)-1].Chart.Name())
	chartReq, err := c.getChart(chartName, targetVersion, &uc.ChartPathOptions)
	if err != nil {
		return fmt.Errorf("[%s] get chart error: %v", releaseName, err)
	}

	if _, err = uc.Run(releaseName, chartReq, nil); err != nil {
		return fmt.Errorf("[%s] release upgrade from version %s to %s error: %v", releaseName,
			r[len(r)-1].Chart.Metadata.Version, targetVersion, err)
	}

	klog.V(0).Infof("[%s] release upgrade from version %s to %s success", releaseName,
		r[len(r)-1].Chart.Metadata.Version, targetVersion)

	return nil
}

// AddOrUpdateRepo Add or update repo from repo config
func (c *Client) AddOrUpdateRepo(repoEntry *repo.Entry) error {
	klog.V(0).Infof("load repo info: %+v", repoEntry)

	rfPath := c.setting.RepositoryConfig

	if _, err := os.Stat(rfPath); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		err := os.MkdirAll(filepath.Dir(rfPath), os.ModePerm)
		if err != nil {
			return err
		}
	}

	rfContent, err := os.ReadFile(rfPath)
	if err != nil {
		return fmt.Errorf("repo file read error, path: %s, error: %v", rfPath, err)
	}

	rf := repo.File{}

	if err = yaml.Unmarshal(rfContent, &rf); err != nil {
		return err
	}

	klog.V(0).Infof("load repo file: %+v", rf)

	isNewRepo := true

	// if has repo already exists, tip and update repo.
	if rf.Has(repoEntry.Name) {
		klog.V(0).Infof("[%s] repo already exists", repoEntry.Name)
		isNewRepo = false
	}

	cr, err := repo.NewChartRepository(repoEntry, getter.All(c.setting))
	if err != nil {
		return err
	}

	klog.V(0).Infof("[%s] start download index file", repoEntry.Name)
	indexFilePath, err := cr.DownloadIndexFile()
	if err != nil {
		return fmt.Errorf("[%s] download index file error: %v", repoEntry.Name, err)
	}
	klog.V(0).Infof("Index file = %s", indexFilePath)

	if !isNewRepo {
		klog.V(0).Infof("[%s] repo update success, path: %s", repoEntry.Name, rfPath)
		return nil
	}

	// Update new repo to repo config file.
	rf.Update(repoEntry)
	if err := rf.WriteFile(c.setting.RepositoryConfig, 0644); err != nil {
		return fmt.Errorf("write repo file %s error: %v", rfPath, err)
	}

	klog.V(0).Infof("change repo success, path: %s", rfPath)

	return nil
}

// 获取chart的版本号，TODO
func (c *Client) GetVersions(err error, indexFilePath string) error {
	// 读取 index.yaml 文件
	file, err := os.ReadFile(indexFilePath)
	if err != nil {
		log.Fatalf("Error opening index file: %v", err)
	}

	// 解析 YAML 文件
	var index repo.IndexFile

	err = yaml.Unmarshal(file, &index)
	if err != nil {
		return err
	}

	// 查找 haproxy 的所有版本
	chartName := "haproxy"
	versions := []string{}
	if chartEntries, ok := index.Entries[chartName]; ok {
		for _, entry := range chartEntries {
			versions = append(versions, entry.Version)
		}
	}

	// 输出结果
	if len(versions) == 0 {
		fmt.Printf("No versions found for chart %s\n", chartName)
	} else {
		fmt.Printf("Available versions for %s:\n", chartName)
		for _, v := range versions {
			fmt.Println(v)
		}
	}
	return nil
}

func debug(format string, v ...interface{}) {
	klog.V(0).Infof(format, v...)
}
