package pod

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
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
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	info := &info{}
	err = c.ShouldBindBodyWithJSON(info)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	ctx := amis.GetContextWithUser(c)
	poder := kom.Cluster(selectedCluster).WithContext(ctx).
		Namespace(info.Namespace).
		Name(info.PodName).Ctl().Pod().
		ContainerName(info.ContainerName)

	if info.Path == "" {
		info.Path = "/"
	}
	// 获取文件列表
	nodes, err := poder.ListAllFiles(info.Path)
	if err != nil {
		amis.WriteJsonError(c, fmt.Errorf("获取文件列表失败,容器内没有shell或者没有ls命令"))
		return
	}
	// 作为文件树，应该去掉. .. 两个条目
	nodes = slice.Filter(nodes, func(index int, item *kom.FileInfo) bool {
		return item.Name != "." && item.Name != ".."
	})
	amis.WriteJsonList(c, nodes)
}

// ShowFile 处理下载文件的 HTTP 请求
func ShowFile(c *gin.Context) {
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	info := &info{}
	err = c.ShouldBindBodyWithJSON(info)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	ctx := amis.GetContextWithUser(c)
	poder := kom.Cluster(selectedCluster).WithContext(ctx).
		Namespace(info.Namespace).
		Name(info.PodName).Ctl().Pod().
		ContainerName(info.ContainerName)
	if info.FileType != "" && info.FileType != "file" && info.FileType != "directory" {
		amis.WriteJsonError(c, fmt.Errorf("无法查看%s类型文件", info.FileType))
		return
	}
	if info.Path == "" {
		amis.WriteJsonError(c, fmt.Errorf("路径不能为空"))
		return
	}
	if info.IsDir {
		amis.WriteJsonError(c, fmt.Errorf("无法保存目录"))
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
		"content": string(fileContent),
	})
}
func SaveFile(c *gin.Context) {
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	info := &info{}
	err = c.ShouldBindBodyWithJSON(info)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	klog.V(6).Infof("info \n%v\n", utils.ToJSON(info))

	ctx := amis.GetContextWithUser(c)
	poder := kom.Cluster(selectedCluster).WithContext(ctx).
		Namespace(info.Namespace).
		Name(info.PodName).Ctl().Pod().
		ContainerName(info.ContainerName)

	if info.Path == "" {
		amis.WriteJsonError(c, fmt.Errorf("路径不能为空"))
		return
	}
	if info.IsDir {
		amis.WriteJsonError(c, fmt.Errorf("无法保存目录"))
		return
	}

	// 上传文件
	if err := poder.SaveFile(info.Path, info.FileContext); err != nil {
		klog.V(6).Infof("Error uploading file: %v", err)
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOK(c)
}

func DownloadFile(c *gin.Context) {
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	info := &info{}
	info.PodName = c.Query("podName")
	info.Path = c.Query("path")
	info.ContainerName = c.Query("containerName")
	info.Namespace = c.Query("namespace")

	ctx := amis.GetContextWithUser(c)
	poder := kom.Cluster(selectedCluster).WithContext(ctx).
		Namespace(info.Namespace).
		Name(info.PodName).Ctl().Pod().
		ContainerName(info.ContainerName)

	// 从容器中下载文件
	var fileContent []byte
	var finalFileName string
	if c.Query("type") == "tar" {
		fileContent, err = poder.DownloadTarFile(info.Path)
		// 从路径中提取文件名作为下载时的文件名，并添加.tar后缀
		fileName := filepath.Base(info.Path)
		fileNameWithoutExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))
		finalFileName = fileNameWithoutExt + ".tar"
	} else {
		fileContent, err = poder.DownloadFile(info.Path)
		finalFileName = filepath.Base(info.Path)
	}
	if err != nil {
		klog.V(6).Infof("下载文件错误: %v", err)
		amis.WriteJsonError(c, err)
		return
	}
	// 设置响应头，指定文件名和类型
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", finalFileName))
	c.Data(http.StatusOK, "application/octet-stream", fileContent)

}

// UploadFile 处理上传文件的 HTTP 请求
func UploadFile(c *gin.Context) {
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	info := &info{}

	info.ContainerName = c.PostForm("containerName")
	info.Namespace = c.PostForm("namespace")
	info.PodName = c.PostForm("podName")
	info.Path = c.PostForm("path")
	info.FileName = c.PostForm("fileName")

	if info.FileName == "" {
		amis.WriteJsonData(c, gin.H{
			"file": gin.H{
				"uid":    -1,
				"name":   info.FileName,
				"status": "error",
				"error":  "文件名不能为空",
			},
		})
		return
	}
	if info.Path == "" {
		amis.WriteJsonData(c, gin.H{
			"file": gin.H{
				"uid":    -1,
				"name":   info.FileName,
				"status": "error",
				"error":  "路径不能为空",
			},
		})
		return
	}
	// 替换FileName中非法字符
	info.FileName = utils.SanitizeFileName(info.FileName)

	ctx := amis.GetContextWithUser(c)
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		amis.WriteJsonData(c, gin.H{
			"file": gin.H{
				"uid":    -1,
				"name":   info.FileName,
				"status": "error",
				"error":  "获取上传文件错误",
			},
		})
		return
	}

	// 保存上传文件
	tempFilePath, err := saveUploadedFile(file)
	if err != nil {
		amis.WriteJsonData(c, gin.H{
			"file": gin.H{
				"uid":    -1,
				"name":   info.FileName,
				"status": "error",
				"error":  err.Error(),
			},
		})
		return
	}
	defer os.Remove(tempFilePath) // 请求结束时删除临时文件

	// 上传文件到 Pod 中
	if err := uploadToPod(ctx, selectedCluster, info, tempFilePath); err != nil {
		amis.WriteJsonData(c, gin.H{
			"file": gin.H{
				"uid":    -1,
				"name":   info.FileName,
				"status": "error",
				"error":  err.Error(),
			},
		})
		return
	}

	// 	{
	//    uid: 'uid',      // 文件唯一标识，建议设置为负数，防止和内部产生的 id 冲突
	//    name: 'xx.png',   // 文件名
	//    status: 'done' | 'uploading' | 'error' | 'removed' , //  beforeUpload 拦截的文件没有 status 状态属性
	//    response: '{"status": "success"}', // 服务端响应内容
	//    linkProps: '{"download": "image"}', // 下载链接额外的 HTML 属性
	// }
	amis.WriteJsonData(c, gin.H{
		"file": gin.H{
			"uid":    -1,
			"name":   info.FileName,
			"status": "done",
		},
	})

}
func DeleteFile(c *gin.Context) {
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	info := &info{}
	err = c.ShouldBindBodyWithJSON(info)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	ctx := amis.GetContextWithUser(c)
	poder := kom.Cluster(selectedCluster).WithContext(ctx).
		Namespace(info.Namespace).
		Name(info.PodName).Ctl().Pod().
		ContainerName(info.ContainerName)
	// 从容器中下载文件
	result, err := poder.DeleteFile(info.Path)
	if err != nil {
		klog.V(6).Infof("删除文件错误: %v", err)
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOKMsg(c, "删除成功"+string(result))
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
		return "", fmt.Errorf("打开上传文件错误: %v", err)
	}
	defer src.Close()

	if _, err := io.Copy(tempFile, src); err != nil {
		return "", fmt.Errorf("无法写入临时文件: %v", err)
	}

	return tempFilePath, nil
}

// uploadToPod 上传文件到 Pod
func uploadToPod(ctx context.Context, selectedCluster string, info *info, tempFilePath string) error {

	poder := kom.Cluster(selectedCluster).WithContext(ctx).
		Namespace(info.Namespace).
		Name(info.PodName).Ctl().Pod().
		ContainerName(info.ContainerName)

	openTmpFile, err := os.Open(tempFilePath)
	if err != nil {
		return fmt.Errorf("打开上传临时文件错误: %v", err)
	}
	defer openTmpFile.Close()

	// 上传文件到 Pod 中
	if err := poder.UploadFile(info.Path, openTmpFile); err != nil {
		return fmt.Errorf("上传文件到Pod中错误: %v", err)
	}

	return nil
}
