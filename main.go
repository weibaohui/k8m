package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/cb"
	"github.com/weibaohui/k8m/pkg/controller/chat"
	"github.com/weibaohui/k8m/pkg/controller/cluster"
	"github.com/weibaohui/k8m/pkg/controller/cm"
	"github.com/weibaohui/k8m/pkg/controller/cronjob"
	"github.com/weibaohui/k8m/pkg/controller/deploy"
	"github.com/weibaohui/k8m/pkg/controller/doc"
	"github.com/weibaohui/k8m/pkg/controller/ds"
	"github.com/weibaohui/k8m/pkg/controller/dynamic"
	"github.com/weibaohui/k8m/pkg/controller/login"
	"github.com/weibaohui/k8m/pkg/controller/menu"
	"github.com/weibaohui/k8m/pkg/controller/node"
	"github.com/weibaohui/k8m/pkg/controller/ns"
	"github.com/weibaohui/k8m/pkg/controller/pod"
	"github.com/weibaohui/k8m/pkg/controller/rs"
	"github.com/weibaohui/k8m/pkg/controller/storage"
	"github.com/weibaohui/k8m/pkg/controller/sts"
	"github.com/weibaohui/k8m/pkg/flag"
	"github.com/weibaohui/k8m/pkg/middleware"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/callbacks"
	"github.com/weibaohui/kom/kom"
	"k8s.io/klog/v2"
)

//go:embed ui/dist/*
var embeddedFiles embed.FS
var Version string
var GitCommit string

var Model = "Qwen/Qwen2.5-Coder-7B-Instruct"
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
	service.ChatService().SetVars(ApiKey, ApiUrl, Model)

	go func() {
		// 初始化kom
		// 先注册回调，后面集群连接后，需要执行回调
		callbacks.RegisterInit()
		// 首先尝试读取 in-cluster 配置
		_, err := kom.Clusters().RegisterInCluster()
		if err != nil {
			klog.Errorf("InCluster集群初始化失败%v，下面尝试使用kubeconfig文件初始化，并扫描同文件夹下其他配置文件", err)
			// 初始化kubectl 连接
			service.ClusterService().ScanClustersInPath(cfg.KubeConfig)
			service.ClusterService().RegisterClustersByPath(cfg.KubeConfig)
			klog.Infof("处理%d个集群", len(service.ClusterService().AllClusters()))
			klog.Infof("已连接%d个集群", len(service.ClusterService().ConnectedClusters()))
		} else {
			cfg.InCluster = true
			klog.Infof("启用InCluster 模式。k8m当前运行在宿主集群内部")
		}

		kom.Clusters().Show()

		// 初始化本项目中的回调
		cb.RegisterCallback()
	}()

	// 启动watch
	go func() {
		service.ClusterService().DelayStartFunc(func() {
			service.PodService().Watch()
			service.NodeService().Watch()
		})
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
	r.Use(gzip.Gzip(gzip.DefaultCompression))
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
		api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/links/services", pod.LinksServices)
		api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/links/endpoints", pod.LinksEndpoints)
		api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/links/pvc", pod.LinksPVC)
		api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/links/pv", pod.LinksPV)
		api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/links/ingress", pod.LinksIngress)
		api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/links/env", pod.LinksEnv)
		api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/links/envFromPod", pod.LinksEnvFromPod)
		api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/links/configmap", pod.LinksConfigMap)
		api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/links/secret", pod.LinksSecret)
		api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/links/node", pod.LinksNode)

		// k8s pod
		api.GET("/pod/logs/sse/ns/:ns/pod_name/:pod_name/container/:container_name", pod.StreamLogs)
		api.GET("/pod/logs/download/ns/:ns/pod_name/:pod_name/container/:container_name", pod.DownloadLogs)
		api.POST("/pod/exec/ns/:ns/pod_name/:pod_name/container/:container_name", pod.Exec)
		api.GET("/pod/wsexec/ns/:ns/pod_name/:pod_name/container/:container_name", pod.WsExec)
		// k8s deploy
		api.POST("/:kind/group/:group/version/:version/restart/ns/:ns/name/:name", deploy.Restart)
		api.POST("/:kind/group/:group/version/:version/batch/restart", deploy.BatchRestart)
		api.POST("/deploy/batch/stop", deploy.BatchStop)
		api.POST("/deploy/batch/restore", deploy.BatchRestore)
		api.POST("/deploy/rollout/undo/ns/:ns/name/:name/revision/:revision", deploy.Undo)
		api.GET("/deploy/rollout/history/ns/:ns/name/:name", deploy.History)
		api.POST("/deploy/rollout/pause/ns/:ns/name/:name", deploy.Pause)
		api.POST("/deploy/rollout/resume/ns/:ns/name/:name", deploy.Resume)
		api.POST("/deploy/scale/ns/:ns/name/:name/replica/:replica", deploy.Scale)

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
		api.GET("/node/taints/list", node.AllTaintList)

		// 节点污点
		api.POST("/node/update_taints/name/:name", node.UpdateTaint)
		api.POST("/node/delete_taints/name/:name", node.DeleteTaint)
		api.POST("/node/add_taints/name/:name", node.AddTaint)
		api.GET("/node/list_taints/name/:name", node.ListTaint)

		// k8s ns
		api.GET("/ns/option_list", ns.OptionList)

		// k8s storageclass
		api.GET("/storageclass/option_list", storage.OptionList)

		// k8s cluster
		api.GET("/cluster/all", cluster.List)
		api.POST("/cluster/scan", cluster.Scan)
		api.GET("/cluster/option_list", cluster.OptionList)
		api.GET("/cluster/file/option_list", cluster.FileOptionList)
		api.POST("/cluster/reconnect/fileName/:fileName/contextName/:contextName", cluster.Reconnect)
		api.POST("/cluster/setDefault/fileName/:fileName/contextName/:contextName", cluster.SetDefault)
		api.POST("/cluster/setDefault/full_name/:fileName/:contextName", cluster.SetDefault)
		// k8s sts
		api.POST("/statefulset/rollout/undo/ns/:ns/name/:name/revision/:revision", sts.Undo)
		api.GET("/statefulset/rollout/history/ns/:ns/name/:name", sts.History)
		api.POST("/statefulset/restart/ns/:ns/name/:name", sts.Restart)
		api.POST("/statefulset/batch/restart", sts.BatchRestart)
		api.POST("/statefulset/batch/stop", sts.BatchStop)
		api.POST("/statefulset/batch/restore", sts.BatchRestore)
		api.POST("/statefulset/scale/ns/:ns/name/:name/replica/:replica", sts.Scale)

		// k8s ds
		api.POST("/daemonset/rollout/undo/ns/:ns/name/:name/revision/:revision", ds.Undo)
		api.GET("/daemonset/rollout/history/ns/:ns/name/:name", ds.History)
		api.POST("/daemonset/restart/ns/:ns/name/:name", ds.Restart)
		api.POST("/daemonset/batch/restart", ds.BatchRestart)
		api.POST("/daemonset/batch/stop", ds.BatchStop)
		api.POST("/daemonset/batch/restore", ds.BatchRestore)

		// k8s rs
		api.POST("/replicaset/restart/ns/:ns/name/:name", rs.Restart)
		api.POST("/replicaset/batch/restart", rs.BatchRestart)
		api.POST("/replicaset/batch/stop", rs.BatchStop)
		api.POST("/replicaset/batch/restore", rs.BatchRestore)
		// k8s configmap
		api.POST("/configmap/ns/:ns/name/:name/import", cm.Import)

		// k8s cronjob
		api.POST("/cronjob/pause/ns/:ns/name/:name", cronjob.Pause)
		api.POST("/cronjob/resume/ns/:ns/name/:name", cronjob.Resume)
		api.POST("/cronjob/batch/resume", cronjob.BatchResume)
		api.POST("/cronjob/batch/pause", cronjob.BatchPause)

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
		api.GET("/chat/example", chat.Example)
		api.GET("/chat/example/field", chat.FieldExample)
		api.GET("/chat/ask", chat.Ask)

		// pod 文件浏览上传下载
		api.POST("/file/list", pod.FileList)
		api.POST("/file/show", pod.ShowFile)
		api.POST("/file/save", pod.SaveFile)
		api.POST("/file/download", pod.DownloadFile)
		api.POST("/file/upload", pod.UploadFile)
		api.POST("/file/delete", pod.DeleteFile)
		// Pod 资源使用情况
		api.GET("/pod/usage/ns/:ns/name/:name", pod.Usage)
		// Pod 关联资源
		api.GET("/pod/ns/:ns/name/:name/links/services", pod.LinksServices)
		api.GET("/pod/ns/:ns/name/:name/links/endpoints", pod.LinksEndpoints)
		api.GET("/pod/ns/:ns/name/:name/links/pvc", pod.LinksPVC)
		api.GET("/pod/ns/:ns/name/:name/links/pv", pod.LinksPV)
		api.GET("/pod/ns/:ns/name/:name/links/ingress", pod.LinksIngress)
		api.GET("/pod/ns/:ns/name/:name/links/env", pod.LinksEnv)
		api.GET("/pod/ns/:ns/name/:name/links/envFromPod", pod.LinksEnvFromPod)
		api.GET("/pod/ns/:ns/name/:name/links/configmap", pod.LinksConfigMap)
		api.GET("/pod/ns/:ns/name/:name/links/secret", pod.LinksSecret)
		api.GET("/pod/ns/:ns/name/:name/links/node", pod.LinksNode)

		// 菜单
		api.GET("/menu/list", menu.List)
	}

	klog.Infof("启动成功，开始监听0.0.0.0:%d", flag.Init().Port)
	err := r.Run(fmt.Sprintf(":%d", flag.Init().Port))
	if err != nil {
		klog.Fatalf("Error %v", err)
	}
	// listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
