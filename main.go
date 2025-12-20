package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/weibaohui/k8m/pkg/cb"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/controller/admin/ai_prompt"
	"github.com/weibaohui/k8m/pkg/controller/admin/cluster"
	"github.com/weibaohui/k8m/pkg/controller/admin/config"
	"github.com/weibaohui/k8m/pkg/controller/admin/event"
	"github.com/weibaohui/k8m/pkg/controller/admin/inspection"
	"github.com/weibaohui/k8m/pkg/controller/admin/mcp"
	"github.com/weibaohui/k8m/pkg/controller/admin/menu"
	"github.com/weibaohui/k8m/pkg/controller/admin/user"
	"github.com/weibaohui/k8m/pkg/controller/chat"
	"github.com/weibaohui/k8m/pkg/controller/cluster_status"
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
	"github.com/weibaohui/k8m/pkg/eventhandler"
	"github.com/weibaohui/k8m/pkg/flag"
	helm2 "github.com/weibaohui/k8m/pkg/helm"
	"github.com/weibaohui/k8m/pkg/leader"
	"github.com/weibaohui/k8m/pkg/lease"
	"github.com/weibaohui/k8m/pkg/lua"
	"github.com/weibaohui/k8m/pkg/middleware"
	_ "github.com/weibaohui/k8m/pkg/models" // 注册模型
	"github.com/weibaohui/k8m/pkg/plugins"
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

	// 启动watch和定时任务（仅在成为Leader时执行）
	go func() {
		service.McpService().Init()
		// 初始化 Lease 同步（监听器与后续 Leader 清理）
		cfg := flag.Init()
		leaseOpts := lease.Options{
			Namespace:                 cfg.LeaseNamespace,
			LeaseDurationSeconds:      cfg.LeaseDurationSeconds,
			LeaseRenewIntervalSeconds: cfg.LeaseRenewIntervalSeconds,
			ResyncPeriod:              30 * time.Second,
			ClusterID:                 cfg.HostClusterID,
		}
		leaseCtx := context.Background()
		if err = service.LeaseManager().Init(leaseCtx, leaseOpts); err == nil {
			err = service.LeaseManager().StartWatcher(leaseCtx, service.ClusterService().Connect, service.ClusterService().Disconnect)
			if err != nil {
				klog.Errorf("启动 Lease 管理器监听器失败: %v", err)
			}
			// 启动 Lease 过期清理（Leader）
			err = service.LeaseManager().StartLeaderCleanup(leaseCtx)
			if err != nil {
				klog.Errorf("启动 Lease 管理器过期清理失败: %v", err)
			}
		} else {
			klog.Errorf("初始化 Lease 管理器失败: %v", err)
		}

		service.ClusterService().DelayStartFunc(func() {
			service.PodService().Watch()
			service.NodeService().Watch()
			service.PVCService().Watch()
			service.PVService().Watch()
			service.IngressService().Watch()
			service.McpService().Start()

			// 启动Leader选举，成功后再启动定时任务
			leaderCfg := leader.Config{
				LockName:      "k8m-leader-lock",
				LeaseDuration: 60 * time.Second, // 增加到60秒
				RenewDeadline: 50 * time.Second, // 增加到50秒
				RetryPeriod:   10 * time.Second, // 增加到10秒
				OnStartedLeading: func(ctx context.Context) {
					klog.V(2).Infof("[leader] 成为Leader，启动定时任务（集群巡检、Helm仓库更新）")
					lua.InitClusterInspection()
					// 启动helm 更新repo定时任务
					helm2.StartUpdateHelmRepoInBackground()
					// leader 启动对event的webhook处理
					eventhandler.StartEventForwardingWatch()
				},
				OnStoppedLeading: func() {
					klog.V(2).Infof("[leader] 不再是Leader，停止定时任务（集群巡检、Helm仓库更新）")
					// 停止集群巡检任务
					lua.StopClusterInspection()
					// 停止helm更新任务
					helm2.StopUpdateHelmRepoInBackground()
					// leader 启动对event的webhook处理
					eventhandler.StopEventForwardingWatch()
				},
			}

			// 使用后台context
			ctx := context.Background()
			if err := leader.Run(ctx, leaderCfg); err != nil {
				klog.Errorf("[leader] Leader选举失败: %v", err)
			}
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
	// @description 请输入以 `Bearer ` 开头的 Token，例：Bearer xxxxxxxx。未列出接口请参考前端调用方法。Token在个人中心-API密钥菜单下申请。
	r.GET("/swagger/*any", func(c *gin.Context) {
		if cfg.EnableSwagger {
			ginSwagger.WrapHandler(swaggerFiles.Handler)(c)
		} else {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Swagger documentation is disabled",
				"message": "Swagger文档已被禁用，请联系管理员启用",
			})
		}
	})

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
		login.RegisterLoginRoutes(auth)
		sso.RegisterAuthRoutes(auth)
	}

	// 公共参数
	params := r.Group("/params", middleware.AuthMiddleware())
	{
		param.RegisterParamRoutes(params)
	}
	ai := r.Group("/ai", middleware.AuthMiddleware())
	{
		chat.RegisterChatRoutes(ai)
	}

	// 初始化插件管理器并启动（集中注册与默认启用策略在 Start 中完成）
	mgr := plugins.NewManager()
	mgr.Start()

	api := r.Group("/k8s/cluster/:cluster", middleware.AuthMiddleware())
	{

		// cluster
		cluster_status.RegisterClusterRoutes(api)
		// CRD status
		dynamic.RegisterCRDRoutes(api)

		// CRD action
		dynamic.RegisterActionRoutes(api)

		dynamic.RegisterMetadataRoutes(api)
		// Container 信息
		dynamic.RegisterContainerRoutes(api)
		// 节点亲和性
		dynamic.RegisterNodeAffinityRoutes(api)
		// Pod亲和性
		dynamic.RegisterPodAffinityRoutes(api)
		// Pod反亲和性
		dynamic.RegisterPodAntiAffinityRoutes(api)
		// 容忍度
		dynamic.RegisterTolerationRoutes(api)

		// Pod关联资源
		dynamic.RegisterPodLinkRoutes(api)
		// k8s pod
		pod.RegisterLabelRoutes(api)
		pod.RegisterLogRoutes(api)
		pod.RegisterXtermRoutes(api)
		// pod 文件浏览上传下载
		pod.RegisterPodFileRoutes(api)
		// Pod 资源使用情况
		pod.RegisterResourceRoutes(api)
		// Pod 端口转发
		pod.RegisterPortRoutes(api)

		// k8s deploy
		deploy.RegisterActionRoutes(api)
		// p8s svc
		svc.RegisterActionRoutes(api)
		// k8s node
		node.RegisterActionRoutes(api)
		// 资源情况
		node.RegisterResourceRoutes(api)
		// 节点污点
		node.RegisterTaintRoutes(api)
		// label等基础信息
		node.RegisterMetadataRoutes(api)
		node.RegisterShellRoutes(api)
		// k8s ns
		ns.RegisterRoutes(api)
		// yaml
		dynamic.RegisterYamlRoutes(api)

		// k8s sts
		sts.RegisterRoutes(api)
		// k8s ds
		ds.RegisterRoutes(api)

		// k8s rs
		rs.RegisterRoutes(api)
		// k8s configmap
		cm.RegisterRoutes(api)
		// k8s cronjob
		cronjob.RegisterRoutes(api)
		// k8s storage_class
		storageclass.RegisterRoutes(api)
		// k8s ingress_class
		ingressclass.RegisterRoutes(api)
		// k8s gateway_class
		gatewayapi.RegisterRoutes(api)
		// doc
		doc.RegisterRoutes(api)
		k8sgpt.RegisterRoutes(api)
		// helm release
		helm.RegisterHelmReleaseRoutes(api)

		// 插件路由注册交由 Manager 统一处理
		mgr.RegisterRoutes(api)
	}

	mgm := r.Group("/mgm", middleware.AuthMiddleware())
	{
		template.RegisterTemplateRoutes(mgm)
		// user profile 用户自助操作
		profile.RegisterProfileRoutes(mgm)
		// API密钥管理
		apikey.RegisterAPIKeysRoutes(mgm)
		// MCP密钥管理
		mcpkey.RegisterMCPKeysRoutes(mgm)
		// log
		log.RegisterLogRoutes(mgm)
		// 集群连接
		cluster.RegisterUserClusterRoutes(mgm)
		// helm chart
		helm.RegisterHelmChartRoutes(mgm)
	}

	admin := r.Group("/admin", middleware.PlatformAuthMiddleware())
	{
		// condition
		config.RegisterConditionRoutes(admin)
		// sso
		config.RegisterSSOConfigRoutes(admin)
		// ldap
		config.RegisterLdapConfigRoutes(admin)
		// 平台参数配置
		config.RegisterConfigRoutes(admin)
		// 大模型列表管理
		config.RegisterAIModelConfigRoutes(admin)
		// AI提示词管理
		ai_prompt.RegisterAdminAIPromptRoutes(admin)
		// 集群巡检定时任务
		inspection.RegisterAdminScheduleRoutes(admin)
		// K8s事件转发配置
		event.RegisterAdminEventRoutes(admin)
		// 集群巡检记录
		inspection.RegisterAdminRecordRoutes(admin)
		// 集群巡检脚本lua脚本管理
		inspection.RegisterAdminLuaScriptRoutes(admin)
		// 集群巡检webhook管理
		inspection.RegisterAdminWebhookRoutes(admin)
		// MCP配置
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
		// helm Repo 操作
		helm.RegisterHelmRepoRoutes(admin)

		// 菜单自定义
		menu.RegisterAdminMenuRoutes(admin)

		// 插件路由注册交由 Manager 统一处理
		mgr.RegisterAdminRoutes(admin)
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
