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

type info struct {
	ContainerName string `json:"containerName,omitempty"`
	PodName       string `json:"podName,omitempty"`
	Namespace     string `json:"namespace,omitempty"`
	IsDir         bool   `json:"isDir,omitempty"`
	Path          string `json:"path,omitempty"`
	FileContext   string `json:"fileContext,omitempty"`
	FileName      string `json:"fileName,omitempty"`
}

// FileListHandler  处理获取文件列表的 HTTP 请求
func FileListHandler(c *gin.Context) {
	info := &info{}
	err := c.ShouldBindBodyWithJSON(info)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	pf := kubectl.PodFile{
		Namespace:     info.Namespace,
		PodName:       info.PodName,
		ContainerName: info.ContainerName,
	}

	if info.Path == "" {
		info.Path = "/"
	}
	// 获取文件列表
	nodes, err := pf.GetFileList(info.Path)
	if err != nil {
		log.Printf("Error getting file list: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	amis.WriteJsonList(c, nodes)
}

// ShowFileHandler 处理下载文件的 HTTP 请求
func ShowFileHandler(c *gin.Context) {
	info := &info{}
	err := c.ShouldBindBodyWithJSON(info)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	pf := kubectl.PodFile{
		Namespace:     info.Namespace,
		PodName:       info.PodName,
		ContainerName: info.ContainerName,
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
	info := &info{}
	err := c.ShouldBindBodyWithJSON(info)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	pf := kubectl.PodFile{
		Namespace:     info.Namespace,
		PodName:       info.PodName,
		ContainerName: info.ContainerName,
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

// DownloadFileHandler 处理下载文件的 HTTP 请求
func DownloadFileHandler(c *gin.Context) {
	info := &info{}
	err := c.ShouldBindBodyWithJSON(info)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	pf := kubectl.PodFile{
		Namespace:     info.Namespace,
		PodName:       info.PodName,
		ContainerName: info.ContainerName,
	}
	// 从容器中下载文件
	fileContent, err := pf.DownloadFile(info.Path)
	if err != nil {
		log.Printf("Error downloading file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 设置响应头，指定文件名和类型
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(info.Path)))
	c.Data(http.StatusOK, "application/octet-stream", fileContent)
}

// UploadFileHandler 处理上传文件的 HTTP 请求
func UploadFileHandler(c *gin.Context) {
	info := &info{}

	info.ContainerName = c.PostForm("containerName")
	info.Namespace = c.PostForm("namespace")
	info.PodName = c.PostForm("podName")
	info.Path = c.PostForm("path")
	info.FileName = c.PostForm("fileName")

	if info.FileName == "" {
		amis.WriteJsonError(c, fmt.Errorf("文件名不能为空"))
		return
	}
	if info.Path == "" {
		amis.WriteJsonError(c, fmt.Errorf("路径不能为空"))
		return
	}
	// 删除path最后一个/后面的内容，取当前选中文件的父级路径

	pf := kubectl.PodFile{
		Namespace:     info.Namespace,
		PodName:       info.PodName,
		ContainerName: info.ContainerName,
	}

	// 获取上传的文件
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		log.Printf("Error retrieving file: %v", err)
		amis.WriteJsonError(c, err)
		return
	}
	defer file.Close()

	savePath := fmt.Sprintf("%s/%s", info.Path, info.FileName)
	log.Printf("存储文件路径%s", savePath)
	// 上传文件
	if err := pf.UploadFile(savePath, file); err != nil {
		log.Printf("Error uploading file: %v", err)
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonData(c, gin.H{
		"value": "/#",
	})
}
