package cm

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
)

type info struct {
	FileName string `json:"fileName,omitempty"`
}

// Import 处理上传文件的 HTTP 请求
func Import(c *gin.Context) {
	info := &info{}
	ns := c.Param("ns")
	name := c.Param("name")
	selectedCluster := amis.GetselectedCluster(c)

	info.FileName = c.PostForm("fileName")

	// 替换FileName中非法字符
	info.FileName = utils.SanitizeFileName(info.FileName)

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		amis.WriteJsonError(c, fmt.Errorf("error retrieving file: %v", err))
		return
	}

	// 保存上传文件
	tempFilePath, err := saveUploadedFile(file)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	defer os.Remove(tempFilePath) // 请求结束时删除临时文件

	var cm *v1.ConfigMap
	err = kom.Cluster(selectedCluster).Resource(&v1.ConfigMap{}).Name(name).Namespace(ns).Get(&cm).Error
	if err != nil {
		amis.WriteJsonError(c, fmt.Errorf("error retrieving configmap: %v", err))
		return
	}

	data := cm.Data
	bytes, err := os.ReadFile(tempFilePath)
	if err != nil {
		amis.WriteJsonError(c, fmt.Errorf("error reading file: %v", err))
		return
	}
	data[info.FileName] = string(bytes)
	err = kom.Cluster(selectedCluster).Resource(cm).Update(cm).Error
	if err != nil {
		amis.WriteJsonError(c, fmt.Errorf("error updating configmap: %v", err))
		return
	}
	amis.WriteJsonData(c, gin.H{
		"value": "/#",
	})
}

// saveUploadedFile 保存上传文件并返回临时文件路径
func saveUploadedFile(file *multipart.FileHeader) (string, error) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "upload-*")
	if err != nil {
		return "", fmt.Errorf("error creating temp directory: %v", err)
	}

	// 使用原始文件名生成临时文件路径
	tempFilePath := filepath.Join(tempDir, file.Filename)

	// 创建并保存文件
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		return "", fmt.Errorf("error creating temp file: %v", err)
	}
	defer tempFile.Close()

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("无法打开上传文件: %v", err)
	}
	defer src.Close()

	if _, err := io.Copy(tempFile, src); err != nil {
		return "", fmt.Errorf("无法写入临时文件: %v", err)
	}

	return tempFilePath, nil
}
