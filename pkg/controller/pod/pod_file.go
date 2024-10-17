package pod

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/kubectl"
	"github.com/weibaohui/k8m/internal/utils"
	"github.com/weibaohui/k8m/internal/utils/amis"
)

// FileListHandler  处理获取文件列表的 HTTP 请求
func FileListHandler(c *gin.Context) {
	pf := kubectl.PodFile{
		Namespace:     c.Query("namespace"),
		PodName:       c.Query("podName"),
		ContainerName: c.Query("containerName"),
	}
	pf.Namespace = "default"
	pf.PodName = "nginx-deployment-7484bcf4c5-4jh7m"
	pf.ContainerName = "nginx"
	path := c.Query("path")

	if path == "" {
		path = "/"
	}
	// 获取文件列表
	nodes, err := pf.GetFileList(path)
	if err != nil {
		log.Printf("Error getting file list: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	amis.WriteJsonList(c, nodes)
}

// ShowFileHandler 处理下载文件的 HTTP 请求
func ShowFileHandler(c *gin.Context) {
	pf := kubectl.PodFile{
		Namespace:     c.Query("namespace"),
		PodName:       c.Query("podName"),
		ContainerName: c.Query("containerName"),
	}
	pf.Namespace = "default"
	pf.PodName = "nginx-deployment-7484bcf4c5-4jh7m"
	pf.ContainerName = "nginx"

	info := &kubectl.PodFileNode{}
	err := c.ShouldBindBodyWithJSON(info)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	if info.Path == "" {
		amis.WriteJsonOK(c)
		return
	}
	if info.IsDir {
		amis.WriteJsonOK(c)
		return
	}
	// 从容器中下载文件
	fileContent, err := pf.DownloadFile(info.Path)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonData(c, gin.H{
		"content": fileContent,
	})
}
func SaveFileHandler(c *gin.Context) {
	pf := kubectl.PodFile{
		Namespace:     c.Query("namespace"),
		PodName:       c.Query("podName"),
		ContainerName: c.Query("containerName"),
	}
	pf.Namespace = "default"
	pf.PodName = "nginx-deployment-7484bcf4c5-4jh7m"
	pf.ContainerName = "nginx"

	info := &kubectl.PodFileNode{}
	err := c.ShouldBindBodyWithJSON(info)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	if info.Path == "" {
		amis.WriteJsonOK(c)
		return
	}
	if info.IsDir {
		amis.WriteJsonOK(c)
		return
	}

	context, err := utils.DecodeBase64(info.FileContext)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	// 上传文件
	if err := pf.SaveFile(info.Path, context); err != nil {
		log.Printf("Error uploading file: %v", err)
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOK(c)
}

// downloadFileHandler 处理下载文件的 HTTP 请求
func downloadFileHandler(c *gin.Context) {
	pf := kubectl.PodFile{
		Namespace:     c.Query("namespace"),
		PodName:       c.Query("podName"),
		ContainerName: c.Query("containerName"),
	}
	filePath := c.Query("filePath")

	// 从容器中下载文件
	fileContent, err := pf.DownloadFile(filePath)
	if err != nil {
		log.Printf("Error downloading file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 设置响应头，指定文件名和类型
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(filePath)))
	c.Data(http.StatusOK, "application/octet-stream", fileContent)
}

// uploadFileHandler 处理上传文件的 HTTP 请求
func uploadFileHandler(c *gin.Context) {
	pf := kubectl.PodFile{
		Namespace:     c.Query("namespace"),
		PodName:       c.Query("podName"),
		ContainerName: c.Query("containerName"),
	}
	destPath := c.Query("destPath")

	// 解析表单
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		log.Printf("Error parsing form: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid form"})
		return
	}

	// 获取上传的文件
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		log.Printf("Error retrieving file: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "File not found"})
		return
	}
	defer file.Close()

	// 上传文件
	if err := pf.UploadFile(destPath, file); err != nil {
		log.Printf("Error uploading file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully"})
}
