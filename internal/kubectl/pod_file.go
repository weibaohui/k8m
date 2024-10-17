package kubectl

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"strings"

	"k8s.io/client-go/tools/remotecommand"
)

// PodFileNode 文件节点结构
type PodFileNode struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // file or directory
	Permissions string `json:"permissions"`
	Size        int64  `json:"size"`
	ModTime     string `json:"modTime"`
	Path        string `json:"path"`         // 存储路径
	IsDir       bool   `json:"isDir"`        // 指示是否
	FileContext string `json:"file_context"` // 文件内容，保存文件时作为DTO字段使用，读取时不加载
}

// PodFile 应用配置结构
type PodFile struct {
	Namespace     string
	PodName       string
	ContainerName string
}

// GetFileList  获取容器中指定路径的文件和目录列表
func (p *PodFile) GetFileList(path string) ([]*PodFileNode, error) {
	cmd := []string{"ls", "-l", path}
	log.Println("GetFileList", cmd)
	req := kubectl.client.CoreV1().RESTClient().
		Get().
		Namespace(p.Namespace).
		Resource("pods").
		Name(p.PodName).
		SubResource("exec").
		Param("container", p.ContainerName).
		Param("command", cmd[0]).
		Param("command", cmd[1]).
		Param("command", cmd[2]).
		Param("tty", "false").
		Param("stdin", "false").
		Param("stdout", "true").
		Param("stderr", "true")

	executor, err := remotecommand.NewSPDYExecutor(kubectl.config, "POST", req.URL())
	if err != nil {
		return nil, fmt.Errorf("Error creating executor: %v", err)
	}

	var stdout bytes.Buffer
	err = executor.Stream(remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: os.Stderr,
	})

	if err != nil {
		return nil, fmt.Errorf("Error executing command: %v", err)
	}

	s := stdout.String()
	log.Printf("输出%s", s)
	return p.parseFileList(path, s), nil
}

// parseFileList 解析输出并生成 PodFileNode 列表
func (p *PodFile) parseFileList(path, output string) []*PodFileNode {
	var nodes []*PodFileNode
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 9 {
			continue // 不完整的行
		}

		permissions := parts[0]
		name := parts[8]
		size := int64(0)
		modTime := strings.Join(parts[5:8], " ")

		// 判断文件类型
		fileType := "file"
		if permissions[0] == 'd' {
			fileType = "directory"
		}

		// 封装成 PodFileNode
		node := PodFileNode{
			Path:        fmt.Sprintf("/%s", name),
			Name:        name,
			Type:        fileType,
			Permissions: permissions,
			Size:        size,
			ModTime:     modTime,
			IsDir:       fileType == "directory",
		}
		if path != "/" {
			node.Path = fmt.Sprintf("%s/%s", path, name)
		}
		nodes = append(nodes, &node)
	}

	return nodes
}

// downloadFile 从指定容器下载文件
func (p *PodFile) DownloadFile(filePath string) ([]byte, error) {
	cmd := []string{"cat", filePath}

	req := kubectl.client.CoreV1().RESTClient().
		Get().
		Namespace(p.Namespace).
		Resource("pods").
		Name(p.PodName).
		SubResource("exec").
		Param("container", p.ContainerName).
		Param("command", cmd[0]).
		Param("command", cmd[1]).
		Param("tty", "false").
		Param("stdin", "false").
		Param("stdout", "true").
		Param("stderr", "true")

	executor, err := remotecommand.NewSPDYExecutor(kubectl.config, "POST", req.URL())
	if err != nil {
		return nil, fmt.Errorf("Error creating executor: %v", err)
	}

	var stdout bytes.Buffer
	err = executor.Stream(remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: os.Stderr,
	})

	if err != nil {
		return nil, fmt.Errorf("Error executing command: %v", err)
	}

	return stdout.Bytes(), nil
}

// uploadFile 将文件上传到指定容器
func (p *PodFile) UploadFile(destPath string, file multipart.File) error {
	// 创建临时文件
	tempFile, err := os.CreateTemp("", "upload-*")
	if err != nil {
		return fmt.Errorf("Error creating temp file: %v", err)
	}
	defer os.Remove(tempFile.Name()) // 确保临时文件在函数结束时被删除

	// 将上传的文件内容写入临时文件
	_, err = io.Copy(tempFile, file)
	if err != nil {
		return fmt.Errorf("Error writing to temp file: %v", err)
	}

	// 确保文件关闭
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("Error closing temp file: %v", err)
	}

	// 使用 kubectl cp 命令将文件复制到容器内
	cmd := []string{"cp", tempFile.Name(), fmt.Sprintf("%s/%s:%s", p.Namespace, p.PodName, destPath)}

	req := kubectl.client.CoreV1().RESTClient().
		Post().
		Namespace(p.Namespace).
		Resource("pods").
		Name(p.PodName).
		SubResource("exec").
		Param("container", p.ContainerName).
		Param("command", cmd[0]).
		Param("command", cmd[1]).
		Param("command", cmd[2]).
		Param("command", cmd[3]).
		Param("tty", "false").
		Param("stdin", "false").
		Param("stdout", "true").
		Param("stderr", "true")

	executor, err := remotecommand.NewSPDYExecutor(kubectl.config, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("Error creating executor: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err = executor.Stream(remotecommand.StreamOptions{
		Stdin:  bytes.NewReader([]byte{}),
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if err != nil {
		return fmt.Errorf("Error executing command: %v: %s", err, stderr.String())
	}

	return nil
}

func (p *PodFile) SaveFile(path string, context string) error {

	// 创建临时文件
	tempFile, err := os.CreateTemp("", "upload-*")
	if err != nil {
		return fmt.Errorf("error creating temp file: %v", err)
	}
	defer os.Remove(tempFile.Name()) // 确保临时文件在函数结束时被删除

	// 将上传的文件内容写入临时文件
	_, err = io.WriteString(tempFile, context)
	if err != nil {
		return fmt.Errorf("error writing to temp file: %v", err)
	}

	// 确保文件关闭
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("error closing temp file: %v", err)
	}

	cmd := []string{"sh", "-c", fmt.Sprintf("cat > %s", path)}

	req := kubectl.client.CoreV1().RESTClient().
		Post().
		Namespace(p.Namespace).
		Resource("pods").
		Name(p.PodName).
		SubResource("exec").
		Param("container", p.ContainerName).
		Param("tty", "false").
		Param("command", cmd[0]).
		Param("command", cmd[1]).
		Param("command", cmd[2]).
		Param("stdin", "true").
		Param("stdout", "true").
		Param("stderr", "true")

	executor, err := remotecommand.NewSPDYExecutor(kubectl.config, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("error creating executor: %v", err)
	}

	// 打开本地文件进行传输
	file, err := os.Open(tempFile.Name())
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()
	var stdout, stderr bytes.Buffer
	err = executor.Stream(remotecommand.StreamOptions{
		Stdin:  file,
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if err != nil {
		return fmt.Errorf("error executing command: %v: %s", err, stderr.String())
	}

	return nil
}
