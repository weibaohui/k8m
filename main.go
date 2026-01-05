package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"

	"github.com/fatih/color"
	"github.com/go-chi/chi/v5"
	cmiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	_ "github.com/swaggo/http-swagger" // 导入 swagger 文档
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/weibaohui/k8m/pkg/cb"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/controller/admin/ai_prompt"
	"github.com/weibaohui/k8m/pkg/controller/admin/cluster"
	"github.com/weibaohui/k8m/pkg/controller/admin/config"
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
	"github.com/weibaohui/k8m/pkg/controller/user/profile"
	"github.com/weibaohui/k8m/pkg/flag"
	"github.com/weibaohui/k8m/pkg/middleware"
	_ "github.com/weibaohui/k8m/pkg/models" // 注册模型
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	_ "github.com/weibaohui/k8m/pkg/plugins/modules/registrar" // 注册插件集中器
	"github.com/weibaohui/k8m/pkg/response"
	"github.com/weibaohui/k8m/pkg/service"
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

	// 启动watch和定时任务
	go func() {

		service.ClusterService().DelayStartFunc(func() {
			service.PodService().Watch()
			service.NodeService().Watch()
			service.PVCService().Watch()
			service.PVService().Watch()
			service.IngressService().Watch()
		})
	}()

}

// main 启动并运行 Kubernetes 管理服务，完成配置初始化、集群注册与资源监控，配置 Chi 路由和中间件，挂载前端静态资源，并提供认证、集群与资源管理、AI 聊天、用户与平台管理等丰富的 HTTP API 接口。
func buildRouter(mgr *plugins.Manager, r chi.Router) http.Handler {
	cfg := flag.Init()

	if !cfg.Debug {
		r.Use(middleware.CustomRecovery())
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(middleware.AuthMiddleware())

	r.Use(middleware.EnsureSelectedClusterMiddleware())

	pagesFS, _ := fs.Sub(embeddedFiles, "ui/dist/pages")
	r.Handle("/public/pages/*", http.StripPrefix("/public/pages", http.FileServer(http.FS(pagesFS))))
	assetsFS, _ := fs.Sub(embeddedFiles, "ui/dist/assets")
	r.Handle("/assets/*", http.StripPrefix("/assets", http.FileServer(http.FS(assetsFS))))
	monacoFS, _ := fs.Sub(embeddedFiles, "ui/dist/monacoeditorwork")
	r.Handle("/monacoeditorwork/*", http.StripPrefix("/monacoeditorwork", http.FileServer(http.FS(monacoFS))))

	if cfg.Debug {
		r.Mount("/debug", cmiddleware.Profiler())
	}

	r.Get("/favicon.ico", response.Adapter(func(c *response.Context) {
		favicon, _ := embeddedFiles.ReadFile("ui/dist/favicon.ico")
		c.Data(http.StatusOK, "image/x-icon", favicon)
	}))

	// @title           k8m API
	// @version         1.0
	// @securityDefinitions.apikey BearerAuth
	// @in header
	// @name Authorization
	// @description 请输入以 `Bearer ` 开头的 Token，例：Bearer xxxxxxxx。未列出接口请参考前端调用方法。Token在个人中心-API密钥菜单下申请。
	r.Get("/swagger/*", func(w http.ResponseWriter, r *http.Request) {
		if mgr.IsEnabled(modules.PluginNameSwagger) {
			httpSwagger.Handler().ServeHTTP(w, r)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error":"Swagger documentation is disabled","message":"Swagger文档已被禁用，请联系管理员启用"}`))
		}
	})

	r.Get("/", response.Adapter(func(c *response.Context) {
		index, err := embeddedFiles.ReadFile("ui/dist/index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Internal Server Error")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", index)
	}))

	r.Get("/ping", response.Adapter(func(c *response.Context) {
		c.JSON(http.StatusOK, response.H{
			"message": "pong",
		})
	}))
	r.Get("/healthz", response.Adapter(func(c *response.Context) {
		c.JSON(http.StatusOK, response.H{"status": "ok"})
	}))

	r.Route("/auth", func(auth chi.Router) {
		login.RegisterLoginRoutes(auth)
		sso.RegisterAuthRoutes(auth)
	})

	r.Route("/", func(root chi.Router) {
		mgr.RegisterRootRoutes(root)
	})

	r.Route("/params", func(params chi.Router) {
		params.Use(middleware.AuthMiddleware())
		param.RegisterParamRoutes(params)
		mgr.RegisterParamRoutes(params)
	})
	r.Route("/ai", func(ai chi.Router) {
		ai.Use(middleware.AuthMiddleware())
		chat.RegisterChatRoutes(ai)
	})

	r.Get("/health/ready", response.Adapter(func(c *response.Context) {
		if !mgr.IsEnabled(modules.PluginNameLeader) {
			c.Status(http.StatusOK)
			return
		}
		if service.LeaderService().IsCurrentLeader() {
			c.Status(http.StatusOK)
		} else {
			c.Status(http.StatusServiceUnavailable)
		}
	}))

	r.Route("/k8s/cluster/{cluster}", func(api chi.Router) {
		api.Use(middleware.EnsureSelectedClusterMiddleware())

		cluster_status.RegisterClusterRoutes(api)
		dynamic.RegisterCRDRoutes(api)
		dynamic.RegisterActionRoutes(api)
		dynamic.RegisterMetadataRoutes(api)
		dynamic.RegisterContainerRoutes(api)
		dynamic.RegisterNodeAffinityRoutes(api)
		dynamic.RegisterPodAffinityRoutes(api)
		dynamic.RegisterPodAntiAffinityRoutes(api)
		dynamic.RegisterTolerationRoutes(api)
		dynamic.RegisterPodLinkRoutes(api)
		pod.RegisterLabelRoutes(api)
		pod.RegisterLogRoutes(api)
		pod.RegisterXtermRoutes(api)
		pod.RegisterPodFileRoutes(api)
		pod.RegisterResourceRoutes(api)
		pod.RegisterPortRoutes(api)
		deploy.RegisterActionRoutes(api)
		svc.RegisterActionRoutes(api)
		node.RegisterActionRoutes(api)
		node.RegisterResourceRoutes(api)
		node.RegisterTaintRoutes(api)
		node.RegisterMetadataRoutes(api)
		node.RegisterShellRoutes(api)
		ns.RegisterRoutes(api)
		dynamic.RegisterYamlRoutes(api)
		sts.RegisterRoutes(api)
		ds.RegisterRoutes(api)
		rs.RegisterRoutes(api)
		cm.RegisterRoutes(api)
		cronjob.RegisterRoutes(api)
		storageclass.RegisterRoutes(api)
		ingressclass.RegisterRoutes(api)
		gatewayapi.RegisterRoutes(api)
		doc.RegisterRoutes(api)
		k8sgpt.RegisterRoutes(api)
		mgr.RegisterClusterRoutes(api)
	})

	r.Route("/mgm", func(mgm chi.Router) {
		template.RegisterTemplateRoutes(mgm)
		profile.RegisterProfileRoutes(mgm)
		log.RegisterLogRoutes(mgm)
		cluster.RegisterUserClusterRoutes(mgm)
		mgr.RegisterManagementRoutes(mgm)
	})

	r.Route("/admin", func(admin chi.Router) {
		admin.Use(middleware.PlatformAuthMiddleware())
		config.RegisterConditionRoutes(admin)
		config.RegisterSSOConfigRoutes(admin)
		config.RegisterLdapConfigRoutes(admin)
		config.RegisterConfigRoutes(admin)
		config.RegisterAIModelConfigRoutes(admin)
		ai_prompt.RegisterAdminAIPromptRoutes(admin)
		user.RegisterClusterPermissionRoutes(admin)
		user.RegisterAdminUserRoutes(admin)
		user.RegisterAdminUserGroupRoutes(admin)
		cluster.RegisterAdminClusterRoutes(admin)
		menu.RegisterAdminMenuRoutes(admin)
		mgr.RegisterAdminRoutes(admin)
		mgr.RegisterPluginAdminRoutes(admin)
	})

	return r
}

func main() {
	Init()

	mgr := plugins.ManagerInstance()
	mgr.SetRouterBuilder(func(r chi.Router) http.Handler {
		return buildRouter(mgr, r)
	})
	mgr.SetEngine(chi.NewRouter())
	mgr.Start()

	initialRouter := chi.NewRouter()
	ah := plugins.NewAtomicHandler(buildRouter(mgr, initialRouter))
	mgr.SetAtomicHandler(ah)

	cfg := flag.Init()
	showBootInfo(Version, cfg.Port)
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port), ah)
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
