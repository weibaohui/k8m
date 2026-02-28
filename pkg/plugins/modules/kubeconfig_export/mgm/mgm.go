package mgm

import (
	"fmt"
	"strings"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/response"
	"gorm.io/gorm"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

// ListTemplates 获取 kubeconfig 模板列表
func ListTemplates(c *response.Context) {
	klog.V(6).Infof("获取 kubeconfig 模板列表")

	params := dao.BuildParams(c)
	kc := &models.KubeConfig{}
	items, total, err := kc.List(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 只返回必要的信息，不返回敏感的 AccessKey 和 SecretAccessKey
	result := make([]map[string]any, 0)
	for _, item := range items {
		result = append(result, map[string]any{
			"id":           item.ID,
			"server":       item.Server,
			"user":         item.User,
			"cluster":      item.Cluster,
			"namespace":    item.Namespace,
			"display_name": item.DisplayName,
			"is_aws_eks":   item.IsAWSEKS,
			"cluster_name": item.ClusterName,
			"region":       item.Region,
			"proxy_url":    item.ProxyURL,
			"timeout":      item.Timeout,
			"qps":          item.QPS,
			"burst":        item.Burst,
			"created_at":   item.CreatedAt,
			"updated_at":   item.UpdatedAt,
		})
	}

	amis.WriteJsonListWithTotal(c, total, result)
}

// GetClusterKubeconfig 获取指定集群的 kubeconfig
func GetClusterKubeconfig(c *response.Context) {
	clusterID := c.Param("clusterID")
	if clusterID == "" {
		amis.WriteJsonError(c, fmt.Errorf("集群ID不能为空"))
		return
	}

	klog.V(6).Infof("获取集群 %s 的 kubeconfig", clusterID)

	// 查询数据库中的 kubeconfig（这里需要实现根据 clusterID 查询的逻辑）
	// 暂时返回空结果
	amis.WriteJsonError(c, fmt.Errorf("功能暂未实现"))

	klog.V(6).Infof("成功获取集群 %s 的 kubeconfig 信息", clusterID)
}

// GetKubeConfigByID 根据 ID 获取 kubeconfig（用于导出）
func GetKubeConfigByID(c *response.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		amis.WriteJsonError(c, fmt.Errorf("kubeconfig ID不能为空"))
		return
	}

	klog.V(6).Infof("获取 kubeconfig ID %s", idStr)

	params := dao.BuildParams(c)
	kc := &models.KubeConfig{}
	kubeConfig, err := kc.GetOne(params, func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", idStr)
	})
	if err != nil || kubeConfig == nil {
		amis.WriteJsonError(c, fmt.Errorf("kubeconfig 不存在"))
		return
	}

	// 返回完整的 kubeconfig 内容
	amis.WriteJsonData(c, map[string]any{
		"id":           kubeConfig.ID,
		"server":       kubeConfig.Server,
		"user":         kubeConfig.User,
		"cluster":      kubeConfig.Cluster,
		"namespace":    kubeConfig.Namespace,
		"display_name": kubeConfig.DisplayName,
		"content":      kubeConfig.Content,
	})

	klog.V(6).Infof("成功获取 kubeconfig ID %s 的内容", idStr)
}

// ExportKubeConfig 导出 kubeconfig 文件
func ExportKubeConfig(c *response.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		amis.WriteJsonError(c, fmt.Errorf("kubeconfig ID不能为空"))
		return
	}

	klog.V(6).Infof("导出 kubeconfig ID %s", idStr)

	params := dao.BuildParams(c)
	kc := &models.KubeConfig{}
	kubeConfig, err := kc.GetOne(params, func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", idStr)
	})
	if err != nil || kubeConfig == nil {
		amis.WriteJsonError(c, fmt.Errorf("kubeconfig 不存在"))
		return
	}

	// 获取请求参数（从 URL 查询参数获取）
	namespace := c.Query("namespace")
	role := c.Query("role")
	_ = c.Query("description") // description 参数暂未使用

	// 解析 kubeconfig
	config, err := clientcmd.Load([]byte(kubeConfig.Content))
	if err != nil {
		klog.V(6).Infof("解析 kubeconfig 失败: %v", err)
		amis.WriteJsonError(c, err)
		return
	}

	// 如果指定了 namespace，添加到 context
	if namespace != "" {
		currentContext := config.Contexts[config.CurrentContext]
		if currentContext != nil {
			currentContext.Namespace = namespace
		}
	}

	// 生成导出的 kubeconfig
	exportedKubeConfig, err := clientcmd.Write(*config)
	if err != nil {
		klog.V(6).Infof("导出 kubeconfig 失败: %v", err)
		amis.WriteJsonError(c, err)
		return
	}

	// 设置文件名
	filename := kubeConfig.DisplayName
	if filename == "" {
		filename = kubeConfig.Cluster
	}
	if namespace != "" {
		filename += "-" + namespace
	}
	if role != "" {
		filename += "-" + role
	}
	// 清理文件名，替换特殊字符
	filename = strings.ReplaceAll(filename, "/", "-")
	filename = strings.ReplaceAll(filename, "\\", "-")
	filename += ".yaml"

	// 设置响应头
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Header("Content-Type", "application/octet-stream")

	// 写入文件内容
	c.Data(200, "application/octet-stream", exportedKubeConfig)

	klog.V(6).Infof("成功导出 kubeconfig ID %s 到文件: %s, namespace=%s, role=%s", idStr, filename, namespace, role)
}