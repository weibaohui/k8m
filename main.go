package main

import (
	"embed"
	"flag"
	"io/fs"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/kubectl"
	"github.com/weibaohui/k8m/pkg/controller/chat"
	"github.com/weibaohui/k8m/pkg/controller/deploy"
	"github.com/weibaohui/k8m/pkg/controller/doc"
	"github.com/weibaohui/k8m/pkg/controller/dynamic"
	"github.com/weibaohui/k8m/pkg/controller/pod"
	"k8s.io/klog/v2"
)

//go:embed assets
var embeddedFiles embed.FS
var Version string
var GitCommit string

func main() {
	// 初始化 klog，解析命令行参数
	klog.InitFlags(nil)
	_ = flag.Set("v", "2") // 设置日志级别为 2，等同于运行时使用 --v=2
	flag.Parse()
	defer klog.Flush()

	// 打印版本和 Git commit 信息
	klog.V(2).Infof("版本: %s\n", Version)
	klog.V(2).Infof("Git Commit: %s\n", GitCommit)
	_ = kubectl.NewDocs()
	r := gin.Default()

	r.Use(cors.Default())
	r.Use(gzip.Gzip(gzip.DefaultCompression))

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

	api := r.Group("/k8s")
	{
		// dynamic
		api.POST("/yaml/apply", dynamic.Apply)
		api.POST("/yaml/delete", dynamic.Delete)
		api.GET("/:kind/list", dynamic.List)
		api.GET("/:kind/list/ns/:ns", dynamic.List)
		api.POST("/:kind/remove/ns/:ns/name/:name", dynamic.Remove)
		api.POST("/:kind/remove/ns/:ns/names", dynamic.BatchRemove)
		api.POST("/:kind/update/ns/:ns/name/:name", dynamic.Save)
		api.GET("/:kind/ns/:ns/name/:name", dynamic.Fetch)
		// CRD
		api.GET("/:kind/group/:group/ns/:ns/name/:name", dynamic.Fetch)          // CRD
		api.POST("/:kind/group/:group/remove/ns/:ns/name/:name", dynamic.Remove) // CRD
		api.POST("/:kind/group/:group/remove/ns/:ns/names", dynamic.BatchRemove) // CRD
		api.POST("/:kind/group/:group/update/ns/:ns/name/:name", dynamic.Save)   // CRD
		api.GET("/:kind/group/:group/list/ns/:ns", dynamic.List)                 // CRD
		api.GET("/:kind/group/:group/list", dynamic.List)                        // CRD
		// k8s pod
		api.GET("/pod/logs/sse/ns/:ns/pod_name/:pod_name/container/:container_name", pod.StreamLogs)
		api.GET("/pod/logs/download/ns/:ns/pod_name/:pod_name/container/:container_name", pod.DownloadLogs)
		// k8s deploy
		api.POST("/deploy/restart/ns/:ns/name/:name", deploy.Restart)
		api.POST("/deploy/update/ns/:ns/name/:name/container/:container_name/tag/:tag", deploy.UpdateImageTag)

		// 其他 API 路由
		api.GET("/doc/:kind", doc.Doc)
		api.GET("/doc/gvk/:api_version/:kind", doc.Doc)
		api.POST("/doc/detail", doc.Detail)

		// chatgpt
		api.POST("/chat", chat.Chat)
		api.GET("/chat/sse", chat.Sse)

		// pod 文件浏览上传下载
		api.POST("/file/list", pod.FileListHandler)
		api.POST("/file/show", pod.ShowFileHandler)
		api.POST("/file/save", pod.SaveFileHandler)
		api.POST("/file/download", pod.DownloadFileHandler)
		api.POST("/file/upload", pod.UploadFileHandler)

		// k8s pod
		// http://127.0.0.1:3618/k8s/doc/gvk/stable.example.com%2Fv1/CronTab

	}

	err := r.Run(":3618")
	if err != nil {
		log.Fatalf("Error %v", err)
	}
	// listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
