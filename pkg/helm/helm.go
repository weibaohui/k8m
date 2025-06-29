package helm

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/models"
	"gorm.io/gorm"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

type Helm interface {
	AddOrUpdateRepo(repoEntry *repo.Entry) error
	GetReleaseHistory(releaseName string) ([]*release.Release, error)
	InstallRelease(namespace, releaseName, repoName, chartName, version string, values ...string) error
	UninstallRelease(releaseName string) error
	UpgradeRelease(releaseName, repoName, targetVersion string, values ...string) error
	GetChartValue(repoName, chartName, version string) (string, error)
	GetChartVersions(repoName string, chartName string) ([]string, error)
	UpdateReposIndex(ids string)
	GetReleaseList() ([]*release.Release, error)
}

type Client struct {
	setting *cli.EnvSettings
	ac      *action.Configuration
	getter  *RESTClientGetterImpl
	driver  string
}

type Option func(client *Client)

// New  Helm Interface
// 此处namespace 决定了release 记录信息写入哪个命名空间
func New(restConfig *rest.Config, namespace string, options ...Option) (Helm, error) {
	h := Client{
		setting: cli.New(),
		driver:  "secret",
	}

	for _, op := range options {
		op(&h)
	}

	var ac action.Configuration
	g := h.setting.RESTClientGetter()

	h.getter = NewRESTClientGetterImpl(restConfig)
	if h.getter != nil {
		g = h.getter
	}

	// 指定命名空间
	err := ac.Init(g, namespace, h.driver, debug)
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
	klog.V(6).Infof("[%s] get release on target cluster", releaseName)

	// use HELM_NAMESPACE find release
	hc := action.NewHistory(c.ac)

	releases, err := hc.Run(releaseName)
	if err != nil {
		if errors.Is(err, driver.ErrReleaseNotFound) {
			return releases, nil
		}
		klog.Errorf("[%s] 1history client run error: %v", releaseName, err)
		return nil, err
	}
	klog.V(6).Infof("[%s] history releases: %+v", releaseName, releases)
	klog.V(6).Infof(" history releases: %d", len(releases))
	return releases, nil
}
func (c *Client) GetReleaseList() ([]*release.Release, error) {
	// 创建 List 对象
	listAction := action.NewList(c.ac)
	listAction.AllNamespaces = true
	// 添加状态掩码过滤
	listAction.StateMask = action.ListDeployed | action.ListUninstalled | action.ListFailed | action.ListSuperseded | action.ListUninstalling | action.ListPendingInstall | action.ListPendingUpgrade | action.ListPendingRollback | action.ListUnknown
	listAction.All = true

	// 获取 Release 列表
	releases, err := listAction.Run()
	if err != nil {
		klog.V(6).Infof("Failed to list releases: %v", err)
		return nil, err
	}
	return releases, nil
}

// InstallRelease install release
func (c *Client) InstallRelease(namespace, releaseName, repoName, chartName, version string, values ...string) error {
	klog.V(6).Infof("install release, name: %s, version: %s, chartName: %s", releaseName, version, chartName)
	klog.V(6).Infof("helm repository cache path: %s", c.setting.RepositoryCache)

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
	ic.Namespace = namespace
	client, _ := registry.NewClient()
	ic.SetRegistryClient(client)
	// 安装时，写入到release的info.Description中。
	ic.Description = repoName
	chartReq, err := c.getChart(repoName, chartName, version, &ic.ChartPathOptions)
	if err != nil {
		return fmt.Errorf("[%s] get chart error: %v", releaseName, err)
	}

	// 4. 加载默认 values.yaml
	defaultValues, err := chartutil.CoalesceValues(chartReq, nil)
	if err != nil {
		return fmt.Errorf("failed to coalesce default values: %v", err)
	}

	finalValues := defaultValues

	if len(values) != 0 {
		// 5. 加载自定义 values.yaml
		customValues, err := ParseValuesYaml(values[0])
		if err != nil {
			return fmt.Errorf("failed to parse custom values: %v", err)
		}

		// 6. 合并默认值 + 自定义值
		finalValues = chartutil.CoalesceTables(customValues, defaultValues.AsMap())

	}
	klog.V(6).Infof("values: \n%s", finalValues)
	if _, err = ic.Run(chartReq, finalValues); err != nil {
		return fmt.Errorf("[%s] install error: %v", releaseName, err)
	}

	klog.V(6).Infof("[%s] release install success", releaseName)

	return nil
}
func ParseValuesYaml(data string) (map[string]interface{}, error) {

	var result map[string]interface{}
	err := yaml.Unmarshal([]byte(data), &result)
	return result, err
}

// getChart get chart
func (c *Client) getChart(repoName, chartName, version string, chartPathOptions *action.ChartPathOptions) (*chart.Chart, error) {
	var (
		lc  *chart.Chart
		err error
	)
	klog.V(6).Infof("LocalPath=%s/%s-%s.tgz", c.setting.RepositoryCache, chartName, version)
	localPath := fmt.Sprintf("%s/%s-%s.tgz", c.setting.RepositoryCache, chartName, version)
	if _, err = os.Stat(localPath); err == nil {
		lc, err = loader.Load(localPath)
		if err == nil && lc != nil {
			// 找到本地缓存
			klog.V(6).Infof("使用[%s/%s-%s]本地缓存 %s", repoName, chartName, version, localPath)
			return lc, nil
		}
		if err != nil {
			klog.V(6).Infof("未找到[%s/%s-%s]本地缓存 %s", repoName, chartName, version, localPath)
		}
	}

	// 创建HelmRepository对象
	helmRepo := &models.HelmRepository{
		Name: repoName,
	}
	helm, err := helmRepo.GetOne(nil, func(db *gorm.DB) *gorm.DB {
		return db.Where("name = ?", repoName).First(helmRepo)
	})
	if err != nil {
		return nil, err
	}

	// 读取repo 元信息
	// 解析 YAML 文件
	var index repo.IndexFile
	err = yaml.Unmarshal([]byte(helm.Content), &index)
	if err != nil {
		return nil, err
	}

	var chartURL string
	if cv, ok := index.Entries[chartName]; ok {
		if item, ok := slice.FindBy(cv, func(index int, item *repo.ChartVersion) bool {
			return item.Version == version
		}); ok {
			if len(item.URLs) > 0 {
				chartURL = item.URLs[0]
			}
		}
	}
	klog.V(6).Infof("chartURL  %s", chartURL)
	filepath, err := chartPathOptions.LocateChart(chartURL, c.setting)
	if err != nil {
		return nil, fmt.Errorf("定位Chart %s 失败: %v。请尝试更新缓存", chartURL, err)
	}
	klog.V(6).Infof("使用[%s/%s] 在线地址 %s", repoName, chartName, chartURL)

	klog.V(6).Infof("chart filepath  %s", filepath)
	lc, err = loader.Load(filepath)

	if err != nil {
		return nil, fmt.Errorf("load chart path options error: %v", err)
	}

	return lc, nil
}

// GetValuesYaml 提取 values.yaml 文件内容
func GetValuesYaml(c *chart.Chart) string {
	for _, file := range c.Raw {
		if file.Name == "values.yaml" {
			return string(file.Data)
		}
	}
	return ""
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

	klog.V(6).Infof("[%s] uninstall release success", releaseName)

	return nil
}

// UpgradeRelease upgrade release version
func (c *Client) UpgradeRelease(releaseName, repoName, targetVersion string, values ...string) error {
	// use HELM_NAMESPACE find release
	uc := action.NewUpgrade(c.ac)

	r, err := c.GetReleaseHistory(releaseName)
	if err != nil {
		return err
	}

	if len(r) == 0 {
		return fmt.Errorf("[%s] release doesn't install", releaseName)
	}

	version := r[len(r)-1]
	// 同版本更新参数不能阻止
	// if version.Chart.Metadata.Version == targetVersion {
	// 	return fmt.Errorf("[%s] version %s already installed", releaseName, version.Chart.Metadata.Version)
	// }
	uc.Version = targetVersion
	uc.Namespace = version.Namespace
	client, _ := registry.NewClient()
	uc.SetRegistryClient(client)
	uc.Description = repoName
	chartName := version.Chart.Name()
	chartReq, err := c.getChart(repoName, chartName, targetVersion, &uc.ChartPathOptions)
	if err != nil {
		return fmt.Errorf("[%s] get chart error: %v", releaseName, err)
	}

	// 4. 加载默认 values.yaml
	defaultValues, err := chartutil.CoalesceValues(chartReq, nil)
	if err != nil {
		return fmt.Errorf("failed to coalesce default values: %v", err)
	}

	finalValues := defaultValues

	if len(values) != 0 {
		// 5. 加载自定义 values.yaml
		customValues, err := ParseValuesYaml(values[0])
		if err != nil {
			return fmt.Errorf("failed to parse custom values: %v", err)
		}

		// 6. 合并默认值 + 自定义值
		finalValues = chartutil.CoalesceTables(customValues, defaultValues.AsMap())

	}

	if _, err = uc.Run(releaseName, chartReq, finalValues); err != nil {
		return fmt.Errorf("[%s] release upgrade from version %s to %s error: %v", releaseName,
			version.Chart.Metadata.Version, targetVersion, err)
	}

	klog.V(6).Infof("[%s] release upgrade from version %s to %s success", releaseName,
		version.Chart.Metadata.Version, targetVersion)

	return nil
}

// AddOrUpdateRepo Add or update repo from repo config
func (c *Client) AddOrUpdateRepo(repoEntry *repo.Entry) error {
	klog.V(6).Infof("load repo info: %+v", repoEntry)

	// 创建HelmRepository对象
	helmRepo := &models.HelmRepository{
		Name:     repoEntry.Name,
		URL:      repoEntry.URL,
		Username: repoEntry.Username,
		Password: repoEntry.Password,
	}
	// 判断该名称、URL的仓库是否存在
	// 检查是否存在相同名称和URL的仓库
	if id, err := helmRepo.GetIDByNameAndURL(nil); err == nil && id > 0 {
		helmRepo.ID = id
	} else {
		// 第一次创建，先保存到数据库
		if err = helmRepo.Save(nil); err != nil {
			return fmt.Errorf("save helm repository to database error: %v", err)
		}
	}

	err := c.updateRepoIndex(repoEntry, helmRepo)
	if err != nil {
		klog.V(6).Infof("update repo info error: %v", err)
		return err
	}
	klog.V(6).Infof("[%s] helm repository saved to database successfully", repoEntry.Name)
	return nil
}

func (c *Client) updateRepoIndex(repoEntry *repo.Entry, helmRepo *models.HelmRepository) error {
	cr, err := repo.NewChartRepository(repoEntry, getter.All(c.setting))
	if err != nil {
		return err
	}

	klog.V(6).Infof("[%s] start download index file", repoEntry.Name)
	indexFilePath, err := cr.DownloadIndexFile()
	if err != nil {
		return fmt.Errorf("[%s] download index file error: %v", repoEntry.Name, err)
	}
	klog.V(6).Infof("Index file = %s", indexFilePath)

	// 将索引文件加载到content字段中
	file, err := os.ReadFile(indexFilePath)
	if err != nil {
		return fmt.Errorf("[%s] read index file error: %v", repoEntry.Name, err)
	}
	helmRepo.Content = string(file)

	// 读取repo 元信息
	// 解析 YAML 文件
	var index repo.IndexFile

	if err = yaml.Unmarshal(file, &index); err == nil {
		helmRepo.Generated = fmt.Sprintf("%s", index.Generated)
	}

	// 保存到数据库
	if err = helmRepo.UpdateContent(nil); err != nil {
		return fmt.Errorf("update helm repository Content to database error: %v", err)
	}

	// 清空数据库中对应的chart repo
	dao.DB().Where("repository_id = ?", helmRepo.ID).Delete(models.HelmChart{})
	// 对index 提取ChartVersions
	for chartName, versionList := range index.Entries {

		if len(versionList) == 0 {
			continue
		}
		slice.SortBy(versionList, func(a *repo.ChartVersion, b *repo.ChartVersion) bool {
			return a.Created.After(b.Created)
		})

		ct := versionList[0]
		m := models.HelmChart{
			RepositoryID:   helmRepo.ID,
			RepositoryName: helmRepo.Name,
			Name:           chartName,
			LatestVersion:  ct.Version,
			Description:    ct.Description,
			Home:           ct.Home,
			Icon:           ct.Icon,
			Keywords:       strings.Join(ct.Keywords, ","),
			KubeVersion:    ct.KubeVersion,
			AppVersion:     ct.AppVersion,
			Deprecated:     ct.Deprecated,
		}
		if len(ct.Sources) > 0 {
			m.Sources = ct.Sources[0]
		}
		err = m.Save(nil)
		if err != nil {
			klog.V(6).Infof("[%s] save helm chart to database error: %v", chartName, err)
		}

	}

	return nil
}

func (c *Client) UpdateReposIndex(ids string) {
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
		err = c.updateRepoIndex(repoEntry, item)
		if err != nil {
			klog.V(6).Infof("update helm repository info error: %v", err)
		}

	}

}
func (c *Client) GetChartValue(repoName, chartName, version string) (string, error) {
	ic := action.NewInstall(c.ac)
	ic.Version = version
	client, _ := registry.NewClient()
	ic.SetRegistryClient(client)
	chartReq, err := c.getChart(repoName, chartName, version, &ic.ChartPathOptions)
	if err != nil {
		return "", fmt.Errorf("[%s/%s] get chart error: %v", repoName, chartName, err)
	}
	// 3. 获取 values.yaml
	values := GetValuesYaml(chartReq)

	return values, nil
}

// GetChartVersions 获取chart的版本
func (c *Client) GetChartVersions(repoName string, chartName string) ([]string, error) {
	var rp models.HelmRepository
	err := dao.DB().Where("name=?", repoName).First(&rp).Error
	if err != nil {
		return nil, err
	}

	// 解析 YAML 文件
	var index repo.IndexFile
	err = yaml.Unmarshal([]byte(rp.Content), &index)
	if err != nil {
		return nil, err
	}

	// 查找 haproxy 的所有版本
	var versions []string
	if chartEntries, ok := index.Entries[chartName]; ok {
		for _, entry := range chartEntries {
			versions = append(versions, entry.Version)
		}
	}

	return versions, nil
}

func debug(format string, v ...interface{}) {
	klog.V(6).Infof(format, v...)
}
