package helm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"helm.sh/helm/v3/pkg/release"
	"k8s.io/klog/v2"
)

const (
	ActionInstall = iota
	ActionUpgrade
	ActionUninstall
)

var (
	defaultTimeOut   = 90 * time.Second
	defaultCheckTime = 20 * time.Second
)

type Manager struct {
	Charts        []*ChartSpec
	HelmClient    Helm
	LocalRepoName string
	TimeOut       time.Duration
}

type ChartSpec struct {
	ReleaseName string
	ChartName   string
	Version     string
	Action      int
	// overwrite values.yaml from chart, only use for install action
	// like: key=value,key.nKey.value; you can use ParseValues convert map to values
	Values string
}

func (m *Manager) Execute() error {
	for _, chart := range m.Charts {
		// check release spec result
		isCompleted, err := m.checkDeployed(chart)
		if err != nil {
			return err
		}
		// release is spec status, skip execute
		if isCompleted {
			continue
		}
		// execute action
		if err = m.ActionExecute(chart); err != nil {
			return err
		}

		if err = m.controller(chart); err != nil {
			return err
		}

		<-time.After(defaultCheckTime)
	}
	return nil
}

func (m *Manager) controller(chart *ChartSpec) error {
	timeOut := defaultTimeOut
	if m.TimeOut != 0 {
		timeOut = m.TimeOut
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("[%s] check status time out", chart.ReleaseName)
		default:
			res, err := m.checkDeployed(chart)
			if err != nil {
				return err
			}
			if res {
				return nil
			}
			continue
		}
	}
}

func (m *Manager) ActionExecute(chart *ChartSpec) error {
	switch chart.Action {
	case ActionInstall:
		return m.HelmClient.InstallRelease(chart.ReleaseName, m.transChartName(chart.ChartName), chart.Version,
			chart.Values)
	case ActionUpgrade:
		return m.HelmClient.UpgradeRelease(chart.ReleaseName, m.LocalRepoName, chart.Version)
	case ActionUninstall:
		return m.HelmClient.UninstallRelease(chart.ReleaseName)
	default:
		return fmt.Errorf("not supported action")
	}
}

// checkDeployed check release deployed
func (m *Manager) checkDeployed(chart *ChartSpec) (bool, error) {
	klog.V(6).Infof("[%s] check release status", chart.ChartName)
	// Get release deploy history
	releases, err := m.HelmClient.GetReleaseHistory(chart.ReleaseName)
	if err != nil {
		return false, err
	}

	var lr *release.Release

	// release had exist on target cluster
	if len(releases) != 0 {
		lr = releases[len(releases)-1]
	} else {
		if chart.Action == ActionUninstall {
			klog.V(6).Infof("[%s] check chart uninstall success", chart.ChartName)
			return true, nil
		}
	}

	if lr == nil {
		return false, nil
	}

	if lr.Chart.Metadata.Version != chart.Version {
		klog.V(6).Infof("[%s] installed version %s, spce: %s", chart.ReleaseName, lr.Chart.Metadata.Version, chart.Version)
	}

	// check status by action type
	switch chart.Action {
	case ActionInstall:
		if lr.Info.Status != "deployed" {
			klog.V(6).Infof("[%s] check status is %s", chart.ChartName, lr.Info.Status)
			return false, nil
		}
		klog.V(6).Infof("[%s] check chart install success", chart.ChartName)
		return true, nil
	case ActionUpgrade:
		if lr.Info.Status != "deployed" {
			klog.V(6).Infof("[%s] check status is %s", chart.ChartName, lr.Info.Status)
			return false, nil
		}
		klog.V(6).Infof("[%s] check chart upgrade success", chart.ChartName)
		return true, nil
	case ActionUninstall:
		return false, nil
	}

	return true, nil
}

func (m *Manager) transChartName(chartName string) string {
	return fmt.Sprintf("%s/%s", m.LocalRepoName, chartName)
}

func ParseValues(preValues map[string]string) string {
	values := make([]string, 0)
	for k, v := range preValues {
		values = append(values, fmt.Sprintf("%s=%s", k, v))
	}

	return strings.Join(values, ",")
}
