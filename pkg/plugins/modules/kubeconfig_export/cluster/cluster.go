package cluster

import (
	"fmt"
	"strings"

	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/response"
	"github.com/weibaohui/k8m/pkg/service"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

// GenerateRequest 生成 kubeconfig 的请求参数
type GenerateRequest struct {
	Namespace   string `json:"namespace,omitempty"`   // 限制的 namespace
	Duration    int    `json:"duration,omitempty"`    // 有效期（天）
	Role        string `json:"role,omitempty"`        // 角色：admin, edit, view
	Description string `json:"description,omitempty"` // 描述
}

// Generate 为当前集群生成 kubeconfig
func Generate(c *response.Context) {
	// 获取集群ID
	clusterID := c.Param("clusterID")
	if clusterID == "" {
		amis.WriteJsonError(c, fmt.Errorf("集群ID不能为空"))
		return
	}

	// 获取请求参数
	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		klog.V(6).Infof("解析生成kubeconfig请求参数失败: %v", err)
		amis.WriteJsonError(c, err)
		return
	}

	// 获取集群配置
	clusterConfig := service.ClusterService().GetClusterByID(clusterID)
	if clusterConfig == nil {
		amis.WriteJsonError(c, fmt.Errorf("集群不存在"))
		return
	}

	// 获取原始 kubeconfig 内容
	kubeConfigContent := clusterConfig.GetKubeconfig()
	if kubeConfigContent == "" {
		amis.WriteJsonError(c, fmt.Errorf("集群 kubeconfig 为空"))
		return
	}

	// 解析 kubeconfig
	config, err := clientcmd.Load([]byte(kubeConfigContent))
	if err != nil {
		klog.V(6).Infof("解析集群 kubeconfig 失败: %v", err)
		amis.WriteJsonError(c, err)
		return
	}

	// 如果指定了 namespace，添加到 context
	if req.Namespace != "" {
		currentContext := config.Contexts[config.CurrentContext]
		if currentContext != nil {
			currentContext.Namespace = req.Namespace
		}
	}

	// 生成新的 kubeconfig
	newKubeConfig, err := clientcmd.Write(*config)
	if err != nil {
		klog.V(6).Infof("生成新的 kubeconfig 失败: %v", err)
		amis.WriteJsonError(c, err)
		return
	}

	// 返回生成的 kubeconfig
	amis.WriteJsonData(c, map[string]interface{}{
		"cluster_id":  clusterID,
		"cluster_name": clusterConfig.ClusterName,
		"server":       clusterConfig.Server,
		"namespace":    req.Namespace,
		"role":         req.Role,
		"kubeconfig":   string(newKubeConfig),
	})

	klog.V(6).Infof("成功为集群 %s 生成 kubeconfig", clusterID)
}

// ExportRequest 导出 kubeconfig 的请求参数
type ExportRequest struct {
	Format      string `json:"format,omitempty"`      // 格式：yaml, json
	Namespace   string `json:"namespace,omitempty"`   // 限制的 namespace
	Duration    int    `json:"duration,omitempty"`    // 有效期（天）
	Role        string `json:"role,omitempty"`        // 角色
	Description string `json:"description,omitempty"` // 描述
}

// sanitizeFilename 清理文件名，移除可能破坏响应头的字符
func sanitizeFilename(input string) string {
	// 移除控制字符（ASCII < 0x20，不包括 CR LF）和控制字符 DEL
	var result strings.Builder
	for _, r := range input {
		// 允许字母、数字、中文、空格、连字符、下划线、点
	// 移除引号、分号、反斜杠等可能导致头注入的字符
		if r >= 32 && r < 127 {
			// ASCII 可打印字符
			if r == '"' || r == ';' || r == '\\' || r == '/' {
				continue // 移除特殊字符
			}
			result.WriteRune(r)
		} else if r > 127 {
			// 允许中文字符（Unicode 大于 127）
			result.WriteRune(r)
		}
	}
	resultStr := result.String()
	if resultStr == "" {
		resultStr = "kubeconfig"
	}
	return resultStr
}

// Export 导出当前集群的 kubeconfig
func Export(c *response.Context) {
	// 获取集群ID
	clusterID := c.Param("clusterID")
	if clusterID == "" {
		amis.WriteJsonError(c, fmt.Errorf("集群ID不能为空"))
		return
	}

	// 获取请求参数
	var req ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果解析失败，使用默认值
		req = ExportRequest{
			Format: "yaml",
		}
	}

	// 获取集群配置
	clusterConfig := service.ClusterService().GetClusterByID(clusterID)
	if clusterConfig == nil {
		amis.WriteJsonError(c, fmt.Errorf("集群不存在"))
		return
	}

	// 获取原始 kubeconfig 内容
	kubeConfigContent := clusterConfig.GetKubeconfig()
	if kubeConfigContent == "" {
		amis.WriteJsonError(c, fmt.Errorf("集群 kubeconfig 为空"))
		return
	}

	// 解析 kubeconfig
	config, err := clientcmd.Load([]byte(kubeConfigContent))
	if err != nil {
		klog.V(6).Infof("解析集群 kubeconfig 失败: %v", err)
		amis.WriteJsonError(c, err)
		return
	}

	// 如果指定了 namespace，添加到 context
	if req.Namespace != "" {
		currentContext := config.Contexts[config.CurrentContext]
		if currentContext != nil {
			currentContext.Namespace = req.Namespace
		}
	}

	// 生成导出的 kubeconfig
	exportedKubeConfig, err := clientcmd.Write(*config)
	if err != nil {
		klog.V(6).Infof("导出 kubeconfig 失败: %v", err)
		amis.WriteJsonError(c, err)
		return
	}

	// 设置文件名（清理用户输入，防止头注入）
	filename := sanitizeFilename(clusterConfig.ClusterName)
	if req.Namespace != "" {
		sanitizedNS := sanitizeFilename(req.Namespace)
		if sanitizedNS != "" {
			filename += "-" + sanitizedNS
		}
	}
	if req.Role != "" {
		sanitizedRole := sanitizeFilename(req.Role)
		if sanitizedRole != "" {
			filename += "-" + sanitizedRole
		}
	}
	filename += ".yaml"

	// 设置响应头
	c.Writer.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Writer.Header().Set("Content-Type", "application/octet-stream")

	// 写入文件内容
	c.Writer.WriteHeader(200)
	_, _ = c.Writer.Write(exportedKubeConfig)

	klog.V(6).Infof("成功导出集群 %s 的 kubeconfig 到文件: %s", clusterID, filename)
}

// GetKubeConfigByDBID 根据数据库中的 kubeconfig ID 获取 kubeconfig
func GetKubeConfigByDBID(c *response.Context) {
	// 获取 kubeconfig ID
	idStr := c.Param("id")
	if idStr == "" {
		amis.WriteJsonError(c, fmt.Errorf("kubeconfig ID不能为空"))
		return
	}

	// 获取请求参数
	var req ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req = ExportRequest{
			Format: "yaml",
		}
	}

	// 查询数据库中的 kubeconfig
	// 注意：这个函数暂时未实现，因为需要从数据库查询 kubeconfig
	// 未来可以根据实际需求实现
	amis.WriteJsonError(c, fmt.Errorf("功能暂未实现"))

	klog.V(6).Infof("请求导出 kubeconfig ID %s", idStr)
}