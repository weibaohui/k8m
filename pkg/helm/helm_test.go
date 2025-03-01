package helm

import (
	"testing"

	"helm.sh/helm/v3/pkg/repo"
)

func TestClient_AddOrUpdateRepo(t *testing.T) {
	helm, err := New()
	if err != nil {
		t.Logf("new helm error: %v", err)
	}
	err = helm.AddOrUpdateRepo(&repo.Entry{
		Name: "bitnami",
		URL:  "https://charts.bitnami.com/bitnami",
	})
	if err != nil {
		t.Logf("helm.AddOrUpdateRepo error: %v", err)
	}
	err = helm.InstallRelease("haproxy-r", "bitnami", "haproxy", "2.2.11")
	if err != nil {
		t.Logf("helm.InstallRelease error: %v", err)
	}
	history, err := helm.GetReleaseHistory("haproxy-r")
	if err != nil {
		t.Logf("helm.GetReleaseHistory error: %v", err)
		return
	}
	for _, v := range history {
		t.Logf("release: %s, version: %d, status: %s", v.Name, v.Version, v.Info.Status.String())
	}

}
