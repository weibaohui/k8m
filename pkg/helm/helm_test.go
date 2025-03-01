package helm

// func Test_AddOrUpdateRepo(t *testing.T) {
// 	helm, err := New()
// 	if err != nil {
// 		t.Logf("new helm error: %v", err)
// 	}
// 	err = helm.AddOrUpdateRepo(&repo.Entry{
// 		Name: "bitnami",
// 		URL:  "https://charts.bitnami.com/bitnami",
// 	})
// 	if err != nil {
// 		t.Logf("helm.AddOrUpdateRepo error: %v", err)
// 	}
//
// }
//
// func Test_InstallChart(t *testing.T) {
// 	helm, err := New()
// 	if err != nil {
// 		t.Logf("new helm error: %v", err)
// 	}
// 	err = helm.InstallRelease("haproxy-r", "bitnami", "haproxy", "2.2.11")
// 	if err != nil {
// 		t.Logf("helm.InstallRelease error: %v", err)
// 	}
//
// }

// func Test_GetReleaseHistory1(t *testing.T) {
// 	helm, err := New("")
// 	if err != nil {
// 		t.Logf("new helm error: %v", err)
// 	}
// 	history, err := helm.GetReleaseHistory("haproxy-r")
// 	klog.V(0).Infof("history len: %d", len(history))
// 	if err != nil {
// 		t.Logf("helm.GetReleaseHistory error: %v", err)
// 	}
// 	for _, v := range history {
// 		klog.V(0).Infof("release: %s, version: %d, status: %s", v.Name, v.Version, v.Info.Status.String())
// 	}
// }
