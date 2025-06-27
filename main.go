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
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/weibaohui/k8m/pkg/cb"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/controller/admin/cluster"
	"github.com/weibaohui/k8m/pkg/controller/admin/config"
	"github.com/weibaohui/k8m/pkg/controller/admin/inspection"
	"github.com/weibaohui/k8m/pkg/controller/admin/mcp"
	"github.com/weibaohui/k8m/pkg/controller/admin/user"
	"github.com/weibaohui/k8m/pkg/controller/chat"
	"github.com/weibaohui/k8m/pkg/controller/cm"
	"github.com/weibaohui/k8m/pkg/controller/cronjob"
	"github.com/weibaohui/k8m/pkg/controller/deploy"
	"github.com/weibaohui/k8m/pkg/controller/doc"
	"github.com/weibaohui/k8m/pkg/controller/ds"
	"github.com/weibaohui/k8m/pkg/controller/dynamic"
	"github.com/weibaohui/k8m/pkg/controller/gatewayapi"
	"github.com/weibaohui/k8m/pkg/controller/helm"
	"github.com/weibaohui/k8m/pkg/controller/ingressclass"
	"github.com/weibaohui/k8m/pkg/controller/k8sgpt"
	"github.com/weibaohui/k8m/pkg/controller/log"
	"github.com/weibaohui/k8m/pkg/controller/login"
	"github.com/weibaohui/k8m/pkg/controller/node"
	"github.com/weibaohui/k8m/pkg/controller/ns"
	"github.com/weibaohui/k8m/pkg/controller/param"
	"github.com/weibaohui/k8m/pkg/controller/pod"
	"github.com/weibaohui/k8m/pkg/controller/rs"
	"github.com/weibaohui/k8m/pkg/controller/sso"
	"github.com/weibaohui/k8m/pkg/controller/storageclass"
	"github.com/weibaohui/k8m/pkg/controller/sts"
	"github.com/weibaohui/k8m/pkg/controller/svc"
	"github.com/weibaohui/k8m/pkg/controller/template"
	"github.com/weibaohui/k8m/pkg/controller/user/apikey"
	"github.com/weibaohui/k8m/pkg/controller/user/mcpkey"
	"github.com/weibaohui/k8m/pkg/controller/user/profile"
	"github.com/weibaohui/k8m/pkg/flag"
	"github.com/weibaohui/k8m/pkg/lua"
	"github.com/weibaohui/k8m/pkg/middleware"
	_ "github.com/weibaohui/k8m/pkg/models" // 注册模型
	"github.com/weibaohui/k8m/pkg/service"
	_ "github.com/weibaohui/k8m/swagger"
	"github.com/weibaohui/kom/callbacks"
	"k8s.io/klog/v2"
)

//go:embed ui/dist/*
var embeddedFiles embed.FS
var Version string
var GitCommit string
var GitTag string
var GitRepo string
var InnerModel = "Qwen/Qwen2.5-7B-Instruct"
var InnerApiKey string
var InnerApiUrl string
var BuildDate string

// Init 完成服务的初始化，包括加载配置、设置版本信息、初始化 AI 服务、注册集群及其回调，并启动资源监控。
func Init() {
	// 初始化配置
	cfg := flag.Init()
	// 从数据库中更新配置
	err := service.ConfigService().UpdateFlagFromDBConfig()
	if err != nil {
		klog.Errorf("加载数据库内配置信息失败 error: %v", err)
	}
	cfg.Version = Version
	cfg.GitCommit = GitCommit
	cfg.GitTag = GitTag
	cfg.GitRepo = GitRepo
	cfg.BuildDate = BuildDate
	cfg.ShowConfigInfo()

	// 打印版本和 Git commit 信息
	klog.V(2).Infof("版本: %s\n", Version)
	klog.V(2).Infof("Git Commit: %s\n", GitCommit)
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化ChatService
	service.AIService().SetVars(InnerApiKey, InnerApiUrl, InnerModel)

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

		// 启动时是否自动连接集群
		if cfg.ConnectCluster {
			// 调用 AllClusters 方法获取所有集群
			clusters := service.ClusterService().AllClusters()
			// 遍历集群，进行连接
			for _, clusterInfo := range clusters {
				klog.Infof("连接集群:%s", clusterInfo.ClusterID)
				service.ClusterService().Connect(clusterInfo.ClusterID)
			}
		}
		// 打印集群连接信息
		klog.Infof("处理%d个集群，其中%d个集群已连接", len(service.ClusterService().AllClusters()), len(service.ClusterService().ConnectedClusters()))

	}()

	// 启动watch
	go func() {
		service.McpService().Init()
		service.ClusterService().DelayStartFunc(func() {
			service.PodService().Watch()
			service.NodeService().Watch()
			service.PVCService().Watch()
			service.PVService().Watch()
			service.IngressService().Watch()
			service.McpService().Start()
			// 启动集群巡检
			lua.InitClusterInspection()
		})
	}()

}

// main 启动并运行 Kubernetes 管理服务，完成配置初始化、集群注册与资源监控，配置 Gin 路由和中间件，挂载前端静态资源，并提供认证、集群与资源管理、AI 聊天、用户与平台管理等丰富的 HTTP API 接口。
func main() {
	Init()

	r := gin.Default()

	cfg := flag.Init()

	// 开启Recovery中间件
	if !cfg.Debug {
		r.Use(middleware.CustomRecovery())
	}

	if cfg.Debug {
		// Debug 模式 注册 pprof 路由
		pprof.Register(r)
	}
	r.Use(cors.Default())
	r.Use(gzip.Gzip(gzip.BestCompression))
	r.Use(middleware.SetCacheHeaders())
	r.Use(middleware.AuthMiddleware())
	r.Use(middleware.EnsureSelectedClusterMiddleware())

	r.MaxMultipartMemory = 100 << 20 // 100 MiB

	// 挂载子目录
	pagesFS, _ := fs.Sub(embeddedFiles, "ui/dist/pages")
	r.StaticFS("/public/pages", http.FS(pagesFS))
	assetsFS, _ := fs.Sub(embeddedFiles, "ui/dist/assets")
	r.StaticFS("/assets", http.FS(assetsFS))
	monacoFS, _ := fs.Sub(embeddedFiles, "ui/dist/monacoeditorwork")
	r.StaticFS("/monacoeditorwork", http.FS(monacoFS))

	r.GET("/favicon.ico", func(c *gin.Context) {
		favicon, _ := embeddedFiles.ReadFile("ui/dist/favicon.ico")
		c.Data(http.StatusOK, "image/x-icon", favicon)
	})

	// MCP Server
	sseServer := GetMcpSSEServer("/mcp/k8m/")
	r.GET("/mcp/k8m/sse", adapt(sseServer.SSEHandler))
	r.POST("/mcp/k8m/sse", adapt(sseServer.SSEHandler))
	r.POST("/mcp/k8m/message", adapt(sseServer.MessageHandler))
	r.GET("/mcp/k8m/:key/sse", adapt(sseServer.SSEHandler))
	r.POST("/mcp/k8m/:key/sse", adapt(sseServer.SSEHandler))
	r.POST("/mcp/k8m/:key/message", adapt(sseServer.MessageHandler))

	// @title           k8m API
	// @version         1.0
	// @securityDefinitions.apikey BearerAuth
	// @in header
	// @name Authorization
	// @description 请输入以 `Bearer ` 开头的 Token，例：Bearer xxxxxxxx。未列出接口请参考前端调用方法。
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	auth := r.Group("/auth")
	{
		auth.POST("/login", login.LoginByPassword)
		auth.GET("/sso/config", sso.GetSSOConfig)
		auth.GET("/oidc/:name/sso", sso.GetAuthCodeURL)
		auth.GET("/oidc/:name/callback", sso.HandleCallback)
		auth.GET("/ldap/config", sso.GetLdapEnabled)
	}

	// 公共参数
	params := r.Group("/params", middleware.AuthMiddleware())
	{
		// 获取当前登录用户的角色，登录即可
		params.GET("/user/role", param.UserRole)
		// 获取某个配置项
		params.GET("/config/:key", param.Config)
		// 获取当前登录用户的集群列表,下拉列表
		params.GET("/cluster/option_list", param.ClusterOptionList)
		// 获取当前登录用户的集群列表,table列表
		params.GET("/cluster/all", param.ClusterTableList)
		// 获取当前软件版本信息
		params.GET("/version", param.Version)
		// 获取helm 仓库列表
		params.GET("/helm/repo/option_list", param.RepoOptionList)

		// 获取翻转显示的指标列表
		params.GET("/condition/reverse/list", param.Conditions)

	}
	ai := r.Group("/ai", middleware.AuthMiddleware())
	{

		// chatgpt
		ai.GET("/chat/event", chat.Event)
		ai.GET("/chat/log", chat.Log)
		ai.GET("/chat/cron", chat.Cron)
		ai.GET("/chat/describe", chat.Describe)
		ai.GET("/chat/resource", chat.Resource)
		ai.GET("/chat/any_question", chat.AnyQuestion)
		ai.GET("/chat/any_selection", chat.AnySelection)
		ai.GET("/chat/example", chat.Example)
		ai.GET("/chat/example/field", chat.FieldExample)
		ai.GET("/chat/ws_chatgpt", chat.GPTShell)
		ai.GET("/chat/ws_chatgpt/history", chat.History)
		ai.GET("/chat/ws_chatgpt/history/reset", chat.Reset)
		ai.GET("/chat/k8s_gpt/resource", chat.K8sGPTResource)

	}
	api := r.Group("/k8s/cluster/:cluster", middleware.AuthMiddleware())
	{
		// dynamic
		api.POST("/yaml/apply", dynamic.Apply)
		api.POST("/yaml/upload", dynamic.UploadFile)
		api.POST("/yaml/delete", dynamic.Delete)
		// CRD
		api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name", dynamic.Fetch)                         // CRD
		api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/json", dynamic.FetchJson)                // CRD
		api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/event", dynamic.Event)                   // CRD
		api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/hpa", dynamic.HPA)                       // CRD
		api.POST("/:kind/group/:group/version/:version/ns/:ns/name/:name/scale/replica/:replica", dynamic.Scale) // CRD
		api.POST("/:kind/group/:group/version/:version/remove/ns/:ns/name/:name", dynamic.Remove)                // CRD
		api.POST("/:kind/group/:group/version/:version/batch/remove", dynamic.BatchRemove)                       // CRD
		api.POST("/:kind/group/:group/version/:version/force_remove", dynamic.BatchForceRemove)                  // CRD
		api.POST("/:kind/group/:group/version/:version/update/ns/:ns/name/:name", dynamic.Save)                  // CRD       // CRD
		api.POST("/:kind/group/:group/version/:version/describe/ns/:ns/name/:name", dynamic.Describe)            // CRD
		api.POST("/:kind/group/:group/version/:version/list/ns/:ns", dynamic.List)                               // CRD
		api.POST("/:kind/group/:group/version/:version/list/ns/", dynamic.List)                                  // CRD
		api.POST("/:kind/group/:group/version/:version/list", dynamic.List)
		api.POST("/:kind/group/:group/version/:version/update_labels/ns/:ns/name/:name", dynamic.UpdateLabels)           // CRD
		api.GET("/:kind/group/:group/version/:version/annotations/ns/:ns/name/:name", dynamic.ListAnnotations)           // CRD
		api.POST("/:kind/group/:group/version/:version/update_annotations/ns/:ns/name/:name", dynamic.UpdateAnnotations) // CRD
		api.GET("/crd/group/option_list", dynamic.GroupOptionList)
		api.GET("/crd/kind/option_list", dynamic.KindOptionList)
		api.GET("/crd/status", dynamic.CRDStatus)
		// Container 信息
		api.GET("/:kind/group/:group/version/:version/container_info/ns/:ns/name/:name/container/:container_name", dynamic.ContainerInfo)
		api.GET("/:kind/group/:group/version/:version/container_resources_info/ns/:ns/name/:name/container/:container_name", dynamic.ContainerResourcesInfo)
		api.GET("/:kind/group/:group/version/:version/image_pull_secrets/ns/:ns/name/:name", dynamic.ImagePullSecretOptionList)
		api.GET("/:kind/group/:group/version/:version/container_health_checks/ns/:ns/name/:name/container/:container_name", dynamic.ContainerHealthChecksInfo)
		api.GET("/:kind/group/:group/version/:version/container_env/ns/:ns/name/:name/container/:container_name", dynamic.ContainerEnvInfo)

		api.POST("/:kind/group/:group/version/:version/update_image/ns/:ns/name/:name", dynamic.UpdateImageTag)
		api.POST("/:kind/group/:group/version/:version/update_resources/ns/:ns/name/:name", dynamic.UpdateResources)
		api.POST("/:kind/group/:group/version/:version/update_health_checks/ns/:ns/name/:name", dynamic.UpdateHealthChecks)
		api.POST("/:kind/group/:group/version/:version/update_env/ns/:ns/name/:name", dynamic.UpdateContainerEnv)

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
		api.GET("/pod/top/ns/:ns/list", pod.TopList)

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
		api.POST("/deploy/create", deploy.Create)
		// p8s svc
		api.POST("/service/create", svc.Create)
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
		api.GET("/node/top/list", node.TopList)
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
		api.POST("/configmap/ns/:ns/name/:name/:key/update_configmap", cm.Update)
		api.POST("/configmap/create", cm.Create)
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

		// k8s gateway_class
		api.GET("/gateway_class/option_list", gatewayapi.GatewayClassOptionList)

		// doc
		api.GET("/doc/gvk/:api_version/:kind", doc.Doc)
		api.GET("/doc/kind/:kind/group/:group/version/:version", doc.Doc)
		api.POST("/doc/detail", doc.Detail)

		api.GET("/k8s_gpt/kind/:kind/run", k8sgpt.ResourceRunAnalysis)
		api.POST("/k8s_gpt/cluster/:user_cluster/run", k8sgpt.ClusterRunAnalysis)
		api.GET("/k8s_gpt/cluster/:user_cluster/result", k8sgpt.GetClusterRunAnalysisResult)
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
		// Pod 端口转发
		api.POST("/pod/port_forward/ns/:ns/name/:name/container/:container_name/pod_port/:pod_port/local_port/:local_port/start", pod.StartPortForward)
		api.POST("/pod/port_forward/ns/:ns/name/:name/container/:container_name/pod_port/:pod_port/stop", pod.StopPortForward)
		api.GET("/pod/port_forward/ns/:ns/name/:name/port/list", pod.PortForwardList)

		api.GET("/helm/release/list", helm.ListRelease)
		api.GET("/helm/release/ns/:ns/name/:name/history/list", helm.ListReleaseHistory)
		api.POST("/helm/release/:release/repo/:repo/chart/:chart/version/:version/install", helm.InstallRelease)
		api.POST("/helm/release/ns/:ns/name/:name/uninstall", helm.UninstallRelease)
		api.POST("/helm/release/batch/uninstall", helm.BatchUninstallRelease)
		api.POST("/helm/release/upgrade", helm.UpgradeRelease)
		api.GET("/helm/chart/list", helm.ListChart)
		api.GET("/helm/repo/:repo/chart/:chart/versions", helm.ChartVersionOptionList)
		api.GET("/helm/repo/:repo/chart/:chart/version/:version/values", helm.GetChartValue)
		// helm
		api.GET("/helm/repo/list", helm.ListRepo)
		api.POST("/helm/repo/delete/:ids", helm.DeleteRepo)
		api.POST("/helm/repo/update_index", helm.UpdateReposIndex)
		api.POST("/helm/repo/save", helm.AddOrUpdateRepo)

	}

	mgm := r.Group("/mgm", middleware.AuthMiddleware())
	{

		mgm.GET("/custom/template/kind/list", template.ListKind)
		mgm.GET("/custom/template/list", template.ListTemplate)
		mgm.POST("/custom/template/save", template.SaveTemplate)
		mgm.POST("/custom/template/delete/:ids", template.DeleteTemplate)

		// user profile 用户自助操作
		mgm.GET("/user/profile", profile.Profile)
		mgm.GET("/user/profile/cluster/permissions/list", profile.ListUserPermissions)
		mgm.POST("/user/profile/update_psw", profile.UpdatePsw)
		// user profile 2FA 用户自助操作
		mgm.POST("/user/profile/2fa/generate", profile.Generate2FASecret)
		mgm.POST("/user/profile/2fa/disable", profile.Disable2FA)
		mgm.POST("/user/profile/2fa/enable", profile.Enable2FA)

		// API密钥管理
		mgm.GET("/user/profile/apikeys/list", apikey.List)
		mgm.POST("/user/profile/apikeys/create", apikey.Create)
		mgm.POST("/user/profile/apikeys/delete/:id", apikey.Delete)

		// MCP密钥管理
		mgm.GET("/user/profile/mcpkeys/list", mcpkey.List)
		mgm.POST("/user/profile/mcpkeys/create", mcpkey.Create)
		mgm.POST("/user/profile/mcpkeys/delete/:id", mcpkey.Delete)

		// log
		mgm.GET("/log/shell/list", log.ListShell)
		mgm.GET("/log/operation/list", log.ListOperation)
		// 集群连接
		cluster.RegisterUserClusterRoutes(mgm)

	}

	admin := r.Group("/admin", middleware.PlatformAuthMiddleware())
	{
		// condition
		admin.GET("/condition/list", config.ConditionList)
		admin.POST("/condition/save", config.ConditionSave)
		admin.POST("/condition/delete/:ids", config.ConditionDelete)
		// 指标翻转状态修改
		admin.POST("/condition/save/id/:id/status/:status", config.ConditionQuickSave)

		// SSO 配置
		admin.GET("/config/sso/list", config.SSOConfigList)
		admin.POST("/config/sso/save", config.SSOConfigSave)
		admin.POST("/config/sso/delete/:ids", config.SSOConfigDelete)
		admin.POST("/config/sso/save/id/:id/status/:enabled", config.SSOConfigQuickSave)

		// 集群巡检定时任务
		inspection.RegisterAdminScheduleRoutes(admin)
		// 集群巡检记录
		inspection.RegisterAdminRecordRoutes(admin)
		// 集群巡检脚本lua脚本管理
		inspection.RegisterAdminLuaScriptRoutes(admin)
		// 集群巡检webhook管理
		inspection.RegisterAdminWebhookRoutes(admin)

		// 平台参数配置
		admin.GET("/config/all", config.GetConfig)
		admin.POST("/config/update", config.UpdateConfig)

		// 大模型列表管理
		config.RegisterAIModelConfigRoutes(admin)
		mcp.RegisterMCPServerRoutes(admin)
		mcp.RegisterMCPToolRoutes(admin)
		// 集群授权相关
		user.RegisterClusterPermissionRoutes(admin)
		// 用户管理相关
		user.RegisterAdminUserRoutes(admin)
		// 用户组管理相关
		user.RegisterAdminUserGroupRoutes(admin)
		// 管理集群、纳管\解除纳管\扫描
		cluster.RegisterAdminClusterRoutes(admin)

	}

	showBootInfo(Version, cfg.Port)
	err := r.Run(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))
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
