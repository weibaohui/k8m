package pod

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom/poder"
	"k8s.io/klog/v2"
)

type info struct {
	ContainerName string `json:"containerName,omitempty"`
	PodName       string `json:"podName,omitempty"`
	Namespace     string `json:"namespace,omitempty"`
	IsDir         bool   `json:"isDir,omitempty"`
	Path          string `json:"path,omitempty"`
	FileContext   string `json:"fileContext,omitempty"`
	FileName      string `json:"fileName,omitempty"`
	Size          int64  `json:"size,omitempty"`
	FileType      string `json:"type,omitempty"` // 只有file类型可以查、下载
}

// FileList  处理获取文件列表的 HTTP 请求
func FileList(c *gin.Context) {
	info := &info{}
	err := c.ShouldBindBodyWithJSON(info)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	poder := poder.Instance().
		WithContext(c.Request.Context()).
		Namespace(info.Namespace).
		Name(info.PodName).
		ContainerName(info.ContainerName)

	if info.Path == "" {
		info.Path = "/"
	}
	// 获取文件列表
	nodes, err := poder.GetFileList(info.Path)
	if err != nil {
		amis.WriteJsonError(c, fmt.Errorf("获取文件列表失败,容器内没有shell或者没有ls命令"))
		return
	}
	amis.WriteJsonList(c, nodes)
}

// ShowFile 处理下载文件的 HTTP 请求
func ShowFile(c *gin.Context) {
	info := &info{}
	err := c.ShouldBindBodyWithJSON(info)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	poder := poder.Instance().
		WithContext(c.Request.Context()).
		Namespace(info.Namespace).
		Name(info.PodName).
		ContainerName(info.ContainerName)
	if info.FileType != "" && info.FileType != "file" && info.FileType != "directory" {
		amis.WriteJsonError(c, fmt.Errorf("无法查看%s类型文件", info.FileType))
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
	fileContent, err := poder.DownloadFile(info.Path)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	isText, err := utils.IsTextFile(fileContent)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	if !isText {
		amis.WriteJsonError(c, fmt.Errorf("%s包含非文本内容，请下载后查看", info.Path))
		return
	}

	amis.WriteJsonData(c, gin.H{
		"content": fileContent,
	})
}
func SaveFile(c *gin.Context) {
	info := &info{}
	err := c.ShouldBindBodyWithJSON(info)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	poder := poder.Instance().
		WithContext(c.Request.Context()).
		Namespace(info.Namespace).
		Name(info.PodName).
		ContainerName(info.ContainerName)

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
	if err := poder.SaveFile(info.Path, context); err != nil {
		klog.V(2).Infof("Error uploading file: %v", err)
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOK(c)
}

// DownloadFile 处理下载文件的 HTTP 请求
func DownloadFile(c *gin.Context) {
	info := &info{}
	err := c.ShouldBindBodyWithJSON(info)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	poder := poder.Instance().
		WithContext(c.Request.Context()).
		Namespace(info.Namespace).
		Name(info.PodName).
		ContainerName(info.ContainerName)
	// 从容器中下载文件
	fileContent, err := poder.DownloadFile(info.Path)
	if err != nil {
		klog.V(2).Infof("Error downloading file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 设置响应头，指定文件名和类型
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(info.Path)))
	c.Data(http.StatusOK, "application/octet-stream", fileContent)
}

// UploadFile 处理上传文件的 HTTP 请求
func UploadFile(c *gin.Context) {
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
	// 替换FileName中非法字符
	info.FileName = utils.SanitizeFileName(info.FileName)

	poder := poder.Instance().
		WithContext(c.Request.Context()).
		Namespace(info.Namespace).
		Name(info.PodName).
		ContainerName(info.ContainerName)

	// 获取上传的文件
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		klog.V(2).Infof("Error retrieving file: %v", err)
		amis.WriteJsonError(c, err)
		return
	}
	defer file.Close()

	savePath := fmt.Sprintf("%s/%s", info.Path, info.FileName)
	// klog.V(2).Infof("存储文件路径%s", savePath)
	// 上传文件
	if err := poder.UploadFile(savePath, file); err != nil {
		klog.V(2).Infof("Error uploading file: %v", err)
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonData(c, gin.H{
		"value": "/#",
	})
}
