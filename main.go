package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"

	"github.com/fatih/color"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/cb"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/controller/chat"
	"github.com/weibaohui/k8m/pkg/controller/cluster"
	"github.com/weibaohui/k8m/pkg/controller/cm"
	"github.com/weibaohui/k8m/pkg/controller/cronjob"
	"github.com/weibaohui/k8m/pkg/controller/deploy"
	"github.com/weibaohui/k8m/pkg/controller/doc"
	"github.com/weibaohui/k8m/pkg/controller/ds"
	"github.com/weibaohui/k8m/pkg/controller/dynamic"
	"github.com/weibaohui/k8m/pkg/controller/helm"
	"github.com/weibaohui/k8m/pkg/controller/ingressclass"
	"github.com/weibaohui/k8m/pkg/controller/k8sgpt"
	"github.com/weibaohui/k8m/pkg/controller/kubeconfig"
	"github.com/weibaohui/k8m/pkg/controller/log"
	"github.com/weibaohui/k8m/pkg/controller/login"
	"github.com/weibaohui/k8m/pkg/controller/mcp"
	"github.com/weibaohui/k8m/pkg/controller/node"
	"github.com/weibaohui/k8m/pkg/controller/ns"
	"github.com/weibaohui/k8m/pkg/controller/pod"
	"github.com/weibaohui/k8m/pkg/controller/rs"
	"github.com/weibaohui/k8m/pkg/controller/storageclass"
	"github.com/weibaohui/k8m/pkg/controller/sts"
	"github.com/weibaohui/k8m/pkg/controller/template"
	"github.com/weibaohui/k8m/pkg/controller/user"
	"github.com/weibaohui/k8m/pkg/flag"
	"github.com/weibaohui/k8m/pkg/middleware"
	_ "github.com/weibaohui/k8m/pkg/models" // 注册模型
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/callbacks"
	"k8s.io/klog/v2"
)

//go:embed ui/dist/*
var embeddedFiles embed.FS
var Version string
var GitCommit string

var Model = "Qwen/Qwen2.5-7B-Instruct"
var ApiKey string
var ApiUrl string

func Init() {
	// 初始化配置
	cfg := flag.Init()

	// 打印版本和 Git commit 信息
	klog.V(2).Infof("版本: %s\n", Version)
	klog.V(2).Infof("Git Commit: %s\n", GitCommit)
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化ChatService
	service.AIService().SetVars(ApiKey, ApiUrl, Model)

	go func() {
		// 初始化kom
		// 先注册回调，后面集群连接后，需要执行回调
		callbacks.RegisterInit()

		// 先把自定义钩子注册登记
		service.ClusterService().SetRegisterCallbackFunc(cb.RegisterDefaultCallbacks)

		if cfg.InCluster {
			klog.V(6).Infof("启用InCluster模式，自动注册纳管宿主集群")
			// 注册InCluster集群
			service.ClusterService().RegisterInCluster()
		} else {
			klog.V(6).Infof("未启用InCluster模式，忽略宿主集群纳管")
		}

		// 再注册其他集群
		service.ClusterService().ScanClustersInDB()
		service.ClusterService().ScanClustersInDir(cfg.KubeConfig)
		service.ClusterService().RegisterClustersByPath(cfg.KubeConfig)
		// 打印集群连接信息
		klog.Infof("处理%d个集群，其中%d个集群已连接", len(service.ClusterService().AllClusters()), len(service.ClusterService().ConnectedClusters()))

	}()

	// 启动watch
	go func() {
		service.ClusterService().DelayStartFunc(func() {
			service.PodService().Watch()
			service.NodeService().Watch()
			service.PVCService().Watch()
			service.PVService().Watch()
			service.IngressService().Watch()
		})
		service.McpService().Init()
	}()

}

func main() {
	Init()

	r := gin.Default()

	cfg := flag.Init()
	if !cfg.Debug {
		// debug 模式可以崩溃
		r.Use(middleware.CustomRecovery())
	}
	r.Use(cors.Default())
	r.Use(gzip.Gzip(gzip.BestCompression))
	r.Use(middleware.SetCacheHeaders())
	r.Use(middleware.EnsureSelectedClusterMiddleware())

	r.MaxMultipartMemory = 100 << 20 // 100 MiB

	// 挂载子目录
	pagesFS, _ := fs.Sub(embeddedFiles, "ui/dist/pages")
	r.StaticFS("/public/pages", http.FS(pagesFS))
	assetsFS, _ := fs.Sub(embeddedFiles, "ui/dist/assets")
	r.StaticFS("/assets", http.FS(assetsFS))
	monacoFS, _ := fs.Sub(embeddedFiles, "ui/dist/monacoeditorwork")
	r.StaticFS("/monacoeditorwork", http.FS(monacoFS))

	// 直接返回 index.html
	r.GET("/", func(c *gin.Context) {
		index, err := embeddedFiles.ReadFile("ui/dist/index.html") // 这里路径必须匹配
		if err != nil {
			c.String(http.StatusInternalServerError, "Internal Server Error")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", index)
	})
	// 处理 favicon.ico 请求
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	auth := r.Group("/auth")
	{
		auth.POST("/login", login.LoginByPassword)
	}

	api := r.Group("/k8s", middleware.AuthMiddleware())
	{
		// dynamic
		api.POST("/yaml/apply", dynamic.Apply)
		api.POST("/yaml/upload", dynamic.UploadFile)
		api.POST("/yaml/delete", dynamic.Delete)
		// CRD
		api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name", dynamic.Fetch)              // CRD
		api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/json", dynamic.FetchJson)     // CRD
		api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/event", dynamic.Event)        // CRD
		api.POST("/:kind/group/:group/version/:version/remove/ns/:ns/name/:name", dynamic.Remove)     // CRD
		api.POST("/:kind/group/:group/version/:version/batch/remove", dynamic.BatchRemove)            // CRD
		api.POST("/:kind/group/:group/version/:version/force_remove", dynamic.BatchForceRemove)       // CRD
		api.POST("/:kind/group/:group/version/:version/update/ns/:ns/name/:name", dynamic.Save)       // CRD
		api.POST("/:kind/group/:group/version/:version/describe/ns/:ns/name/:name", dynamic.Describe) // CRD
		api.POST("/:kind/group/:group/version/:version/list/ns/:ns", dynamic.List)                    // CRD
		api.POST("/:kind/group/:group/version/:version/list/ns/", dynamic.List)                       // CRD
		api.POST("/:kind/group/:group/version/:version/list", dynamic.List)
		api.POST("/:kind/group/:group/version/:version/update_labels/ns/:ns/name/:name", dynamic.UpdateLabels)           // CRD
		api.GET("/:kind/group/:group/version/:version/annotations/ns/:ns/name/:name", dynamic.ListAnnotations)           // CRD
		api.POST("/:kind/group/:group/version/:version/update_annotations/ns/:ns/name/:name", dynamic.UpdateAnnotations) // CRD
		api.GET("/crd/group/option_list", dynamic.GroupOptionList)
		api.GET("/crd/kind/option_list", dynamic.KindOptionList)
		// Container 信息
		api.GET("/:kind/group/:group/version/:version/container_info/ns/:ns/name/:name/container/:container_name", dynamic.ContainerInfo)
		api.GET("/:kind/group/:group/version/:version/image_pull_secrets/ns/:ns/name/:name", dynamic.ImagePullSecretOptionList)
		api.POST("/:kind/group/:group/version/:version/update_image/ns/:ns/name/:name", dynamic.UpdateImageTag)
		// 节点亲和性
		api.POST("/:kind/group/:group/version/:version/update_node_affinity/ns/:ns/name/:name", dynamic.UpdateNodeAffinity)
		api.POST("/:kind/group/:group/version/:version/delete_node_affinity/ns/:ns/name/:name", dynamic.DeleteNodeAffinity)
		api.POST("/:kind/group/:group/version/:version/add_node_affinity/ns/:ns/name/:name", dynamic.AddNodeAffinity)
		api.GET("/:kind/group/:group/version/:version/list_node_affinity/ns/:ns/name/:name", dynamic.ListNodeAffinity)
		// Pod亲和性
		api.POST("/:kind/group/:group/version/:version/update_pod_affinity/ns/:ns/name/:name", dynamic.UpdatePodAffinity)
		api.POST("/:kind/group/:group/version/:version/delete_pod_affinity/ns/:ns/name/:name", dynamic.DeletePodAffinity)
		api.POST("/:kind/group/:group/version/:version/add_pod_affinity/ns/:ns/name/:name", dynamic.AddPodAffinity)
		api.GET("/:kind/group/:group/version/:version/list_pod_affinity/ns/:ns/name/:name", dynamic.ListPodAffinity)
		// Pod反亲和性
		api.POST("/:kind/group/:group/version/:version/update_pod_anti_affinity/ns/:ns/name/:name", dynamic.UpdatePodAntiAffinity)
		api.POST("/:kind/group/:group/version/:version/delete_pod_anti_affinity/ns/:ns/name/:name", dynamic.DeletePodAntiAffinity)
		api.POST("/:kind/group/:group/version/:version/add_pod_anti_affinity/ns/:ns/name/:name", dynamic.AddPodAntiAffinity)
		api.GET("/:kind/group/:group/version/:version/list_pod_anti_affinity/ns/:ns/name/:name", dynamic.ListPodAntiAffinity)
		// 容忍度
		api.POST("/:kind/group/:group/version/:version/update_tolerations/ns/:ns/name/:name", dynamic.UpdateTolerations)
		api.POST("/:kind/group/:group/version/:version/delete_tolerations/ns/:ns/name/:name", dynamic.DeleteTolerations)
		api.POST("/:kind/group/:group/version/:version/add_tolerations/ns/:ns/name/:name", dynamic.AddTolerations)
		api.GET("/:kind/group/:group/version/:version/list_tolerations/ns/:ns/name/:name", dynamic.ListTolerations)

		// Pod关联资源
		api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/links/services", dynamic.LinksServices)
		api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/links/endpoints", dynamic.LinksEndpoints)
		api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/links/pvc", dynamic.LinksPVC)
		api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/links/pv", dynamic.LinksPV)
		api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/links/ingress", dynamic.LinksIngress)
		api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/links/env", dynamic.LinksEnv)
		api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/links/envFromPod", dynamic.LinksEnvFromPod)
		api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/links/configmap", dynamic.LinksConfigMap)
		api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/links/secret", dynamic.LinksSecret)
		api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/links/node", dynamic.LinksNode)
		api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/links/pod", dynamic.LinksPod)

		// k8s pod
		api.GET("/pod/logs/sse/ns/:ns/pod_name/:pod_name/container/:container_name", pod.StreamLogs)
		api.GET("/pod/logs/download/ns/:ns/pod_name/:pod_name/container/:container_name", pod.DownloadLogs)
		api.GET("/pod/xterm/ns/:ns/pod_name/:pod_name", pod.Xterm)
		// k8s deploy
		api.POST("/deploy/ns/:ns/name/:name/restart", deploy.Restart)
		api.POST("/deploy/batch/restart", deploy.BatchRestart)
		api.POST("/deploy/batch/stop", deploy.BatchStop)
		api.POST("/deploy/batch/restore", deploy.BatchRestore)
		api.POST("/deploy/ns/:ns/name/:name/revision/:revision/rollout/undo", deploy.Undo)
		api.GET("/deploy/ns/:ns/name/:name/rollout/history", deploy.History)
		api.GET("/deploy/ns/:ns/name/:name/revision/:revision/rollout/history", deploy.HistoryRevisionDiff)
		api.POST("/deploy/ns/:ns/name/:name/rollout/pause", deploy.Pause)
		api.POST("/deploy/ns/:ns/name/:name/rollout/resume", deploy.Resume)
		api.POST("/deploy/ns/:ns/name/:name/scale/replica/:replica", deploy.Scale)
		api.GET("/deploy/ns/:ns/name/:name/events/all", deploy.Event)
		api.GET("/deploy/ns/:ns/name/:name/hpa", deploy.HPA)

		// k8s node
		api.POST("/node/drain/name/:name", node.Drain)
		api.POST("/node/cordon/name/:name", node.Cordon)
		api.POST("/node/uncordon/name/:name", node.UnCordon)
		api.GET("/node/usage/name/:name", node.Usage)
		api.POST("/node/batch/drain", node.BatchDrain)
		api.POST("/node/batch/cordon", node.BatchCordon)
		api.POST("/node/batch/uncordon", node.BatchUnCordon)
		api.GET("/node/name/option_list", node.NameOptionList)
		api.GET("/node/labels/list", node.AllLabelList)
		api.GET("/node/labels/unique_labels", node.UniqueLabels)
		api.GET("/node/taints/list", node.AllTaintList)
		api.POST("/node/name/:node_name/create_node_shell", node.CreateNodeShell)
		api.POST("/node/name/:node_name/cluster_id/:cluster_id/create_kubectl_shell", node.CreateKubectlShell)

		// 节点污点
		api.POST("/node/update_taints/name/:name", node.UpdateTaint)
		api.POST("/node/delete_taints/name/:name", node.DeleteTaint)
		api.POST("/node/add_taints/name/:name", node.AddTaint)
		api.GET("/node/list_taints/name/:name", node.ListTaint)

		// k8s ns
		api.GET("/ns/option_list", ns.OptionList)
		api.POST("/ResourceQuota/create", ns.CreateResourceQuota)
		api.POST("/LimitRange/create", ns.CreateLimitRange)

		// k8s cluster
		api.GET("/cluster/all", cluster.List)
		api.POST("/cluster/scan", cluster.Scan)
		api.GET("/cluster/option_list", cluster.OptionList)
		api.GET("/cluster/file/option_list", cluster.FileOptionList)
		api.POST("/cluster/reconnect/fileName/:fileName/contextName/:contextName", cluster.Reconnect)
		api.POST("/cluster/disconnect/fileName/:fileName/contextName/:contextName", cluster.Disconnect)
		api.POST("/cluster/setDefault/fileName/:fileName/contextName/:contextName", cluster.SetDefault)
		api.POST("/cluster/setDefault/full_name/:fileName/:contextName", cluster.SetDefault)
		api.POST("/cluster/setDefault/full_name/InCluster", cluster.SetDefaultInCluster)
		api.POST("/cluster/kubeconfig/save", kubeconfig.Save)
		api.POST("/cluster/kubeconfig/remove", kubeconfig.Remove)

		// k8s sts
		api.POST("/statefulset/ns/:ns/name/:name/revision/:revision/rollout/undo", sts.Undo)
		api.GET("/statefulset/ns/:ns/name/:name/rollout/history", sts.History)
		api.POST("/statefulset/ns/:ns/name/:name/restart", sts.Restart)
		api.POST("/statefulset/batch/restart", sts.BatchRestart)
		api.POST("/statefulset/batch/stop", sts.BatchStop)
		api.POST("/statefulset/batch/restore", sts.BatchRestore)
		api.POST("/statefulset/ns/:ns/name/:name/scale/replica/:replica", sts.Scale)
		api.GET("/statefulset/ns/:ns/name/:name/hpa", sts.HPA)

		// k8s ds
		api.POST("/daemonset/ns/:ns/name/:name/revision/:revision/rollout/undo", ds.Undo)
		api.GET("/daemonset/ns/:ns/name/:name/rollout/history", ds.History)
		api.POST("/daemonset/ns/:ns/name/:name/restart", ds.Restart)
		api.POST("/daemonset/batch/restart", ds.BatchRestart)
		api.POST("/daemonset/batch/stop", ds.BatchStop)
		api.POST("/daemonset/batch/restore", ds.BatchRestore)

		// k8s rs
		api.POST("/replicaset/ns/:ns/name/:name/restart", rs.Restart)
		api.POST("/replicaset/batch/restart", rs.BatchRestart)
		api.POST("/replicaset/batch/stop", rs.BatchStop)
		api.POST("/replicaset/batch/restore", rs.BatchRestore)
		api.GET("/replicaset/ns/:ns/name/:name/events/all", rs.Event)
		api.GET("/replicaset/ns/:ns/name/:name/hpa", rs.HPA)

		// k8s configmap
		api.POST("/configmap/ns/:ns/name/:name/import", cm.Import)

		// k8s cronjob
		api.POST("/cronjob/pause/ns/:ns/name/:name", cronjob.Pause)
		api.POST("/cronjob/resume/ns/:ns/name/:name", cronjob.Resume)
		api.POST("/cronjob/batch/resume", cronjob.BatchResume)
		api.POST("/cronjob/batch/pause", cronjob.BatchPause)

		// k8s storage_class
		api.POST("/storage_class/set_default/name/:name", storageclass.SetDefault)
		api.GET("/storage_class/option_list", storageclass.OptionList)
		// k8s ingress_class
		api.POST("/ingress_class/set_default/name/:name", ingressclass.SetDefault)
		api.GET("/ingress_class/option_list", ingressclass.OptionList)

		// doc
		api.GET("/doc/gvk/:api_version/:kind", doc.Doc)
		api.GET("/doc/kind/:kind/group/:group/version/:version", doc.Doc)
		api.POST("/doc/detail", doc.Detail)

		// chatgpt
		api.GET("/chat/event", chat.Event)
		api.GET("/chat/log", chat.Log)
		api.GET("/chat/cron", chat.Cron)
		api.GET("/chat/describe", chat.Describe)
		api.GET("/chat/resource", chat.Resource)
		api.GET("/chat/any_question", chat.AnyQuestion)
		api.GET("/chat/any_selection", chat.AnySelection)
		api.GET("/chat/example", chat.Example)
		api.GET("/chat/example/field", chat.FieldExample)
		api.GET("/chat/ws_chatgpt", chat.GPTShell)
		api.GET("/chat/k8s_gpt/resource", chat.K8sGPTResource)

		api.GET("/k8s_gpt/kind/:kind/run", k8sgpt.ResourceRunAnalysis)
		api.GET("/k8s_gpt/cluster/:cluster/run", k8sgpt.ClusterRunAnalysis)
		api.GET("/k8s_gpt/var", k8sgpt.GetFields)

		// pod 文件浏览上传下载
		api.POST("/file/list", pod.FileList)
		api.POST("/file/show", pod.ShowFile)
		api.POST("/file/save", pod.SaveFile)
		api.GET("/file/download", pod.DownloadFile)
		api.POST("/file/upload", pod.UploadFile)
		api.POST("/file/delete", pod.DeleteFile)
		// Pod 资源使用情况
		api.GET("/pod/usage/ns/:ns/name/:name", pod.Usage)
		api.GET("/pod/labels/unique_labels", pod.UniqueLabels)

	}

	mgm := r.Group("/mgm", middleware.AuthMiddleware())
	{
		// 2FA
		mgm.POST("/user/2fa/generate/:id", user.Generate2FASecret)
		mgm.POST("/user/2fa/disable/:id", user.Disable2FA)
		mgm.POST("/user/2fa/enable/:id", user.Enable2FA)

		mgm.GET("/custom/template/kind/list", template.ListKind)
		mgm.GET("/custom/template/list", template.ListTemplate)
		mgm.POST("/custom/template/save", template.SaveTemplate)
		mgm.POST("/custom/template/delete/:ids", template.DeleteTemplate)

		// user
		mgm.GET("/user/list", user.List)
		mgm.POST("/user/save", user.Save)
		mgm.POST("/user/delete/:ids", user.Delete)
		mgm.POST("/user/update_psw/:id", user.UpdatePsw)

		// user_group
		mgm.GET("/user_group/list", user.ListUserGroup)
		mgm.POST("/user_group/save", user.SaveUserGroup)
		mgm.POST("/user_group/delete/:ids", user.DeleteUserGroup)
		mgm.GET("/user_group/option_list", user.GroupOptionList)

		// log
		mgm.GET("/log/shell/list", log.ListShell)
		mgm.GET("/log/operation/list", log.ListOperation)

		// helm
		mgm.GET("/helm/repo/list", helm.ListRepo)
		mgm.POST("/helm/repo/delete/:ids", helm.DeleteRepo)
		mgm.POST("/helm/repo/update_index", helm.UpdateReposIndex)
		mgm.POST("/helm/repo/save", helm.AddOrUpdateRepo)
		mgm.GET("/helm/repo/option_list", helm.RepoOptionList)
		mgm.GET("/helm/repo/:repo/chart/:chart/versions", helm.ChartVersionOptionList)
		mgm.GET("/helm/repo/:repo/chart/:chart/version/:version/values", helm.GetChartValue)
		mgm.GET("/helm/chart/list", helm.ListChart)

		mgm.GET("/helm/release/list", helm.ListRelease)
		mgm.GET("/helm/release/ns/:ns/name/:name/history/list", helm.ListReleaseHistory)
		mgm.POST("/helm/release/:release/repo/:repo/chart/:chart/version/:version/install", helm.InstallRelease)
		mgm.POST("/helm/release/ns/:ns/name/:name/uninstall", helm.UninstallRelease)
		mgm.POST("/helm/release/batch/uninstall", helm.BatchUninstallRelease)
		mgm.POST("/helm/release/upgrade", helm.UpgradeRelease)

		// mcp
		mgm.GET("/mcp/list", mcp.List)

	}

	showBootInfo(Version, flag.Init().Port)
	err := r.Run(fmt.Sprintf(":%d", flag.Init().Port))
	if err != nil {
		klog.Fatalf("Error %v", err)
	}
}

func showBootInfo(version string, port int) {

	// 获取本机所有 IP 地址
	ips, err := utils.GetLocalIPs()
	if err != nil {
		klog.Fatalf("获取本机 IP 失败: %v", err)
		os.Exit(1)
	}
	// 打印 Vite 风格的启动信息
	color.Green("k8m %s  启动成功", version)
	fmt.Printf("%s  ", color.GreenString("➜"))
	fmt.Printf("%s    ", color.New(color.Bold).Sprint("Local:"))
	fmt.Printf("%s\n", color.MagentaString("http://localhost:%d/", port))

	for _, ip := range ips {
		fmt.Printf("%s  ", color.GreenString("➜"))
		fmt.Printf("%s  ", color.New(color.Bold).Sprint("Network:"))
		fmt.Printf("%s\n", color.MagentaString("http://%s:%d/", ip, port))
	}

}
