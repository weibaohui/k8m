package cm

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type info struct {
	FileName string `json:"fileName,omitempty"`
}

// Import 处理上传文件的 HTTP 请求
func Import(c *gin.Context) {
	info := &info{}
	ns := c.Param("ns")
	name := c.Param("name")
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	ctx := amis.GetContextWithUser(c)
	info.FileName = c.PostForm("fileName")

	// 替换FileName中非法字符
	info.FileName = utils.SanitizeFileName(info.FileName)

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		amis.WriteJsonError(c, fmt.Errorf("获取上传的文件错误: %v", err))
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
	err = kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.ConfigMap{}).Name(name).Namespace(ns).Get(&cm).Error
	if err != nil {
		amis.WriteJsonError(c, fmt.Errorf("获取configmap错误: %v", err))
		return
	}
	data := cm.Data
	bytes, err := os.ReadFile(tempFilePath)
	if err != nil {
		amis.WriteJsonError(c, fmt.Errorf("读取文件错误: %v", err))
		return
	}
	if data == nil {
		data = make(map[string]string)
	}
	data[info.FileName] = string(bytes)
	cm.Data = data
	err = kom.Cluster(selectedCluster).WithContext(ctx).Resource(cm).Update(cm).Error
	if err != nil {
		amis.WriteJsonError(c, fmt.Errorf("更新configmap错误: %v", err))
		return
	}
	amis.WriteJsonData(c, gin.H{
		"value": "/#",
	})
}

// Update 更新配置文件
func Update(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	key := c.Param("key")
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	ctx := amis.GetContextWithUser(c)
	// 解析JSON请求体
	var requestBody struct {
		Content string `json:"update_configmap"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		amis.WriteJsonError(c, fmt.Errorf("解析请求体错误: %v", err))
		return
	}
	// 判断content是否==${value}
	if requestBody.Content == "${value}" {
		amis.WriteJsonError(c, fmt.Errorf("内容未发生变化或为${value}"))
		return
	}
	var cm *v1.ConfigMap
	err = kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.ConfigMap{}).Name(name).Namespace(ns).Get(&cm).Error
	if err != nil {
		amis.WriteJsonError(c, fmt.Errorf("获取configmap错误: %v", err))
		return
	}
	// 判断对应key是否存在
	if _, exists := cm.Data[key]; !exists {
		amis.WriteJsonError(c, fmt.Errorf("文件 %s 不存在", key))
		return
	}

	// 更新ConfigMap数据
	if cm.Data == nil {
		cm.Data = make(map[string]string)
	}
	// 替换\r\n为\n
	cm.Data[key] = strings.ReplaceAll(requestBody.Content, "\r\n", "\n")

	// 更新到Kubernetes
	err = kom.Cluster(selectedCluster).
		WithContext(ctx).
		Resource(cm).
		Namespace(ns).
		Update(cm).Error

	amis.WriteJsonErrorOrOK(c, err)
}

// Create 创建configmap接口
func Create(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 解析请求体
	var requestBody struct {
		Metadata struct {
			Namespace string            `json:"namespace"`
			Name      string            `json:"name"`
			Labels    map[string]string `json:"labels,omitempty"`
		} `json:"metadata"`
		Data map[string]interface{} `json:"data"` // 修改为 interface{} 类型
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		amis.WriteJsonError(c, fmt.Errorf("解析请求体错误: %v", err))
		return
	}

	// 判断是否存在同名ConfigMap
	var existingCM v1.ConfigMap
	err = kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.ConfigMap{}).Name(requestBody.Metadata.Name).Namespace(requestBody.Metadata.Namespace).Get(&existingCM).Error
	if err == nil {
		amis.WriteJsonError(c, fmt.Errorf("ConfigMap %s 已存在", requestBody.Metadata.Name))
		return
	}

	// 处理数据：转换所有值为字符串，并替换\r\n为\n
	data := make(map[string]string)
	for key, value := range requestBody.Data {
		var strValue string
		switch v := value.(type) {
		case string:
			strValue = v
		default:
			// 非字符串类型转换为JSON字符串
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				amis.WriteJsonError(c, fmt.Errorf("转换数据为字符串失败: %v", err))
				return
			}
			strValue = string(jsonBytes)
		}
		// 替换换行符
		data[key] = strings.ReplaceAll(strValue, "\r\n", "\n")
	}

	// 创建ConfigMap对象
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      requestBody.Metadata.Name,
			Namespace: requestBody.Metadata.Namespace,
			Labels:    requestBody.Metadata.Labels,
			Annotations: map[string]string{
				"currentVersion": "1",
				"description":    "",
				"originName":     requestBody.Metadata.Name,
			},
		},
		Data: data, // 使用处理后的数据
	}

	// 创建到Kubernetes
	err = kom.Cluster(selectedCluster).
		WithContext(ctx).
		Resource(cm).
		Namespace(requestBody.Metadata.Namespace).
		Create(cm).Error

	amis.WriteJsonErrorOrOK(c, err)
}

// saveUploadedFile 保存上传文件并返回临时文件路径
func saveUploadedFile(file *multipart.FileHeader) (string, error) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "upload-*")
	if err != nil {
		return "", fmt.Errorf("创建临时目录错误: %v", err)
	}

	// 使用原始文件名生成临时文件路径
	tempFilePath := filepath.Join(tempDir, file.Filename)

	// 创建并保存文件
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		return "", fmt.Errorf("创建临时文件错误: %v", err)
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
