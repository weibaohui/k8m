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
	"github.com/weibaohui/k8m/pkg/controller/cm"
	"github.com/weibaohui/k8m/pkg/controller/cronjob"
	"github.com/weibaohui/k8m/pkg/controller/deploy"
	"github.com/weibaohui/k8m/pkg/controller/doc"
	"github.com/weibaohui/k8m/pkg/controller/ds"
	"github.com/weibaohui/k8m/pkg/controller/dynamic"
	"github.com/weibaohui/k8m/pkg/controller/node"
	"github.com/weibaohui/k8m/pkg/controller/ns"
	"github.com/weibaohui/k8m/pkg/controller/pod"
	"github.com/weibaohui/k8m/pkg/controller/rs"
	"github.com/weibaohui/k8m/pkg/controller/sts"
	"github.com/weibaohui/k8m/pkg/flag"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/callbacks"
	"github.com/weibaohui/kom/kom"
	"k8s.io/klog/v2"
)

//go:embed assets
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

	// 初始化kom
	// 先注册回调，后面集群连接后，需要执行回调
	callbacks.RegisterInit()
	// 首先尝试读取 in-cluster 配置
	_, err := kom.Clusters().RegisterInCluster()
	if err != nil {
		klog.Errorf("InCluster集群初始化失败%v，下面尝试使用kubeconfig文件初始化", err)
		// 初始化kubectl 连接
		_, err = kom.Clusters().RegisterByPathWithID(cfg.KubeConfig, "default")
		if err != nil {
			klog.Fatalf("InCluster集群初始化失败、外部集群初始化失败%v", err)
		}
	}

	kom.Clusters().Show()

	// 初始化本项目中的回调
	cb.RegisterCallback()

	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化ChatService
	service.ChatService().SetVars(ApiKey, ApiUrl, Model)
}

func main() {
	Init()

	r := gin.Default()

	r.Use(cors.Default())
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.MaxMultipartMemory = 100 << 20 // 100 MiB
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	// 获取嵌入的静态文件的子文件系统
	publicFS, _ := fs.Sub(embeddedFiles, "assets/public")
	pagesFS, _ := fs.Sub(embeddedFiles, "assets/pages")
	r.StaticFS("/public", http.FS(publicFS))
	r.StaticFS("/pages", http.FS(pagesFS))
	// 可选：提供根路由的 HTML 文件
	// 假设有一个 index.html 文件在 static 目录下
	r.GET("/index.html", func(c *gin.Context) {
		index, err := embeddedFiles.ReadFile("assets/index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Internal Server Error")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", index)
	})
	// 设置根路径路由
	r.GET("/", func(c *gin.Context) {
		// 使用 HTTP 302 重定向
		c.Redirect(http.StatusFound, "/index.html")
	})
	api := r.Group("/k8s")
	{
		// dynamic
		api.POST("/yaml/apply", dynamic.Apply)
		api.POST("/yaml/upload", dynamic.UploadFile)
		api.POST("/yaml/delete", dynamic.Delete)
		// CRD
		api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name", dynamic.Fetch)              // CRD
		api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/event", dynamic.Event)        // CRD
		api.POST("/:kind/group/:group/version/:version/remove/ns/:ns/name/:name", dynamic.Remove)     // CRD
		api.POST("/:kind/group/:group/version/:version/remove/ns/:ns/names", dynamic.BatchRemove)     // CRD
		api.POST("/:kind/group/:group/version/:version/update/ns/:ns/name/:name", dynamic.Save)       // CRD
		api.POST("/:kind/group/:group/version/:version/describe/ns/:ns/name/:name", dynamic.Describe) // CRD
		api.GET("/:kind/group/:group/version/:version/list/ns/:ns", dynamic.List)                     // CRD
		api.GET("/:kind/group/:group/version/:version/list", dynamic.List)                            // CRD
		// k8s pod
		api.GET("/pod/logs/sse/ns/:ns/pod_name/:pod_name/container/:container_name", pod.StreamLogs)
		api.GET("/pod/logs/download/ns/:ns/pod_name/:pod_name/container/:container_name", pod.DownloadLogs)
		api.POST("/pod/exec/ns/:ns/pod_name/:pod_name/container/:container_name", pod.Exec)
		// k8s deploy
		api.POST("/deploy/restart/ns/:ns/name/:name", deploy.Restart)
		api.POST("/deploy/update/ns/:ns/name/:name/container/:container_name/tag/:tag", deploy.UpdateImageTag)
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

		// k8s ns
		api.GET("/ns/option_list", ns.OptionList)

		// k8s sts
		api.POST("/statefulset/rollout/undo/ns/:ns/name/:name/revision/:revision", sts.Undo)
		api.GET("/statefulset/rollout/history/ns/:ns/name/:name", sts.History)
		api.POST("/statefulset/restart/ns/:ns/name/:name", sts.Restart)
		api.POST("/statefulset/scale/ns/:ns/name/:name/replica/:replica", sts.Scale)

		// k8s ds
		api.POST("/daemonset/rollout/undo/ns/:ns/name/:name/revision/:revision", ds.Undo)
		api.GET("/daemonset/rollout/history/ns/:ns/name/:name", ds.History)
		api.POST("/daemonset/restart/ns/:ns/name/:name", ds.Restart)

		// k8s rs
		api.POST("/replicaset/restart/ns/:ns/name/:name", rs.Restart)
		// k8s configmap
		api.POST("/configmap/ns/:ns/name/:name/import", cm.Import)

		// k8s cronjob
		api.POST("/cronjob/pause/ns/:ns/name/:name", cronjob.Pause)
		api.POST("/cronjob/resume/ns/:ns/name/:name", cronjob.Resume)
		// doc
		api.GET("/doc/:kind", doc.Doc)
		api.GET("/doc/gvk/:api_version/:kind", doc.Doc)
		api.POST("/doc/detail", doc.Detail)

		// chatgpt
		api.POST("/chat", chat.Chat)
		api.POST("/chat/event", chat.Event)
		api.POST("/chat/log", chat.Log)
		api.POST("/chat/cron", chat.Cron)
		api.POST("/chat/describe", chat.Describe)
		api.POST("/chat/resource", chat.Resource)
		api.POST("/chat/sse/describe", chat.SSEDescribe)

		// pod 文件浏览上传下载
		api.POST("/file/list", pod.FileList)
		api.POST("/file/show", pod.ShowFile)
		api.POST("/file/save", pod.SaveFile)
		api.POST("/file/download", pod.DownloadFile)
		api.POST("/file/upload", pod.UploadFile)
		api.POST("/file/delete", pod.DeleteFile)
		// Pod 资源使用情况
		api.GET("/pod/usage/ns/:ns/name/:name", pod.Usage)

	}

	klog.Infof("启动成功，k8m服务监听0.0.0.0:%d", flag.Init().Port)
	err := r.Run(fmt.Sprintf(":%d", flag.Init().Port))
	if err != nil {
		klog.Fatalf("Error %v", err)
	}
	// listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
