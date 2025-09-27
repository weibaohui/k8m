package cluster

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/service"
	komaws "github.com/weibaohui/kom/kom/aws"
	"gorm.io/gorm"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

// @Summary 保存KubeConfig
// @Description 保存KubeConfig配置到数据库
// @Security BearerAuth
// @Success 200 {object} string
// @Router /admin/cluster/kubeconfig/save [post]
func (a *Controller) SaveKubeConfig(c *gin.Context) {

	params := dao.BuildParams(c)
	m := models.KubeConfig{}
	err := c.ShouldBindJSON(&m)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 清理显示名称，将特殊字符替换为连字符，支持中文字符
	m.DisplayName = strings.NewReplacer("/", "-", "\\", "-", " ", "-").Replace(strings.TrimSpace(m.DisplayName))

	if m.DisplayName == "" {
		m.DisplayName = m.Cluster
	}
	// 因先删除再创建
	// 可能会更新地址的kubeconfig

	config, err := clientcmd.Load([]byte(m.Content))
	if err != nil {
		klog.V(6).Infof("解析 集群 [%s]失败: %v", m.Server, err)
		return
	}
	index := 0
	total := len(config.Contexts)
	for contextName, _ := range config.Contexts {
		index += 1
		context := config.Contexts[contextName]
		cluster := config.Clusters[context.Cluster]

		kc := &models.KubeConfig{
			Cluster:   context.Cluster,
			Server:    cluster.Server,
			User:      context.AuthInfo,
			Namespace: context.Namespace,
			Content:   m.Content,
		}

		kc.DisplayName = m.DisplayName
		// 大于1个，则名称加序列号
		if total != 1 {
			kc.DisplayName = fmt.Sprintf("%s-%d", m.DisplayName, index)
		}

		if list, _, err := kc.List(params); err == nil && list != nil {
			for _, item := range list {
				_ = kc.Delete(params, fmt.Sprintf("%d", item.ID))
			}
		}

		err = kc.Save(params)
		if err != nil {
			klog.V(6).Infof("保存 集群 [%s]失败: %v", m.Server, err)
			amis.WriteJsonError(c, err)
			return
		}

	}

	// 执行一下扫描
	service.ClusterService().ScanClustersInDB()
	// 初始化本项目中的回调
	amis.WriteJsonOK(c)
}

// @Summary 删除KubeConfig
// @Description 从数据库中删除KubeConfig配置
// @Security BearerAuth
// @Success 200 {object} string
// @Router /admin/cluster/kubeconfig/remove [post]
func (a *Controller) RemoveKubeConfig(c *gin.Context) {

	params := dao.BuildParams(c)
	m := models.KubeConfig{}
	err := c.ShouldBindJSON(&m)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	if list, _, err := m.List(params); err == nil && list != nil {
		for _, item := range list {
			_ = m.Delete(params, fmt.Sprintf("%d", item.ID))
		}
	}

	// 执行一下扫描
	service.ClusterService().ScanClustersInDB()

	amis.WriteJsonOK(c)
}

// SaveAWSEKSCluster
// @Summary 保存AWS EKS集群配置
// @Description 保存AWS EKS集群配置到数据库并注册集群
// @Security BearerAuth
// @Param request body object true "AWS EKS配置信息"
// @Success 200 {object} string "保存成功"
// @Router /admin/cluster/aws/save [post]
func (a *Controller) SaveAWSEKSCluster(c *gin.Context) {
	params := dao.BuildParams(c)

	// 定义请求结构体
	type AWSEKSRequest struct {
		AccessKey   string `json:"accessKey" binding:"required"`
		SecretKey   string `json:"secretKey" binding:"required"`
		Region      string `json:"region" binding:"required"`
		ClusterName string `json:"clusterName" binding:"required"`
		DisplayName string `json:"displayName"`
	}

	var req AWSEKSRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		klog.V(6).Infof("绑定AWS EKS请求参数失败: %v", err)
		amis.WriteJsonError(c, err)
		return
	}
	req.DisplayName = strings.NewReplacer("/", "-", "\\", "-", " ", "-").Replace(strings.TrimSpace(req.DisplayName))

	// 如果没有提供显示名称，使用集群名称
	if req.DisplayName == "" {
		req.DisplayName = req.ClusterName
	}

	// 构造AWS EKS配置
	eksConfig := &komaws.EKSAuthConfig{
		AccessKey:       req.AccessKey,
		SecretAccessKey: req.SecretKey,
		Region:          req.Region,
		ClusterName:     req.ClusterName,
	}

	kg := komaws.NewKubeconfigGenerator()
	kcs, err := kg.GenerateFromAWS(eksConfig)
	if err != nil {
		klog.V(6).Infof("生成AWS EKS集群kubeconfig配置失败: %v", err)
		amis.WriteJsonError(c, fmt.Errorf("生成AWS EKS集群kubeconfig配置失败: %w", err))
		return
	}
	config, err := clientcmd.Load([]byte(kcs))
	if err != nil {
		klog.V(6).Infof("解析 AWS EKS集群kubeconfig配置失败: %v", err)
		return
	}

	index := 0
	total := len(config.Contexts)
	for contextName, _ := range config.Contexts {
		index += 1
		context := config.Contexts[contextName]
		cluster := config.Clusters[context.Cluster]

		kc := &models.KubeConfig{
			Cluster:         context.Cluster,
			Server:          cluster.Server,
			User:            context.AuthInfo,
			Namespace:       context.Namespace,
			Content:         kcs,
			AccessKey:       req.AccessKey,
			SecretAccessKey: req.SecretKey,
			Region:          req.Region,
			ClusterName:     req.ClusterName,
			IsAWSEKS:        true,
		}

		kc.DisplayName = req.DisplayName
		// 大于1个，则名称加序列号
		if total != 1 {
			kc.DisplayName = fmt.Sprintf("%s-%d", req.DisplayName, index)
		}

		if list, _, err := kc.List(params); err == nil && list != nil {
			for _, item := range list {
				_ = kc.Delete(params, fmt.Sprintf("%d", item.ID))
			}
		}

		err = kc.Save(params)
		if err != nil {
			klog.V(6).Infof("保存 AWS EKS集群 [%s]失败: %v", cluster.Server, err)
			amis.WriteJsonError(c, err)
			return
		}

	}

	klog.V(4).Infof("成功保存AWS EKS集群配置: %s [%s/%s]", req.DisplayName, req.Region, req.ClusterName)

	// 执行一下扫描
	service.ClusterService().ScanClustersInDB()

	amis.WriteJsonOKMsg(c, "AWS EKS集群纳管成功")
}

// SaveTokenCluster
// @Summary 保存Token方式集群配置
// @Description 保存Token方式集群配置到数据库并注册集群
// @Security BearerAuth
// @Param request body object true "Token集群配置信息"
// @Success 200 {object} string "保存成功"
// @Router /admin/cluster/token/save [post]
func (a *Controller) SaveTokenCluster(c *gin.Context) {
	params := dao.BuildParams(c)

	// 定义请求结构体
	type TokenClusterRequest struct {
		Server      string `json:"server" binding:"required"`      // Kubernetes API服务器地址
		Token       string `json:"token" binding:"required"`       // 访问Token
		CaData      string `json:"caData"`                         // CA证书数据（可选）
		DisplayName string `json:"displayName" binding:"required"` // 显示名称
		Description string `json:"description"`                    // 集群描述（可选）
	}

	var req TokenClusterRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		klog.V(6).Infof("绑定Token集群请求参数失败: %v", err)
		amis.WriteJsonError(c, err)
		return
	}

	// 清理和验证输入
	req.DisplayName = strings.NewReplacer("/", "-", "\\", "-", " ", "-").Replace(strings.TrimSpace(req.DisplayName))
	req.Server = strings.TrimSpace(req.Server)
	req.Token = strings.TrimSpace(req.Token)
	req.CaData = strings.TrimSpace(req.CaData)
	req.Description = strings.TrimSpace(req.Description)

	// 验证服务器地址格式
	if !strings.HasPrefix(req.Server, "https://") && !strings.HasPrefix(req.Server, "http://") {
		amis.WriteJsonError(c, fmt.Errorf("服务器地址必须以 https:// 或 http:// 开头"))
		return
	}

	// 构造kubeconfig内容
	kubeconfigContent := fmt.Sprintf(`apiVersion: v1
kind: Config
clusters:
- cluster:
    server: %s`, req.Server)

	if req.CaData != "" {
		kubeconfigContent += fmt.Sprintf(`
    certificate-authority-data: %s`, req.CaData)
	} else {
		kubeconfigContent += `
    insecure-skip-tls-verify: true`
	}

	kubeconfigContent += fmt.Sprintf(`
  name: %s
contexts:
- context:
    cluster: %s
    user: %s
  name: %s
current-context: %s
users:
- name: %s
  user:
    token: %s
`, req.DisplayName, req.DisplayName, req.DisplayName, req.DisplayName, req.DisplayName, req.DisplayName, req.Token)

	// 解析kubeconfig以验证格式
	config, err := clientcmd.Load([]byte(kubeconfigContent))
	if err != nil {
		klog.V(6).Infof("解析Token集群kubeconfig配置失败: %v", err)
		amis.WriteJsonError(c, fmt.Errorf("生成的kubeconfig配置无效: %w", err))
		return
	}

	// 检查是否已存在相同的集群配置
	existingKc := &models.KubeConfig{}
	if existing, err := existingKc.GetOne(params, func(db *gorm.DB) *gorm.DB {
		return db.Where("server = ? AND display_name = ?", req.Server, req.DisplayName)
	}); err == nil && existing != nil {
		// 删除已存在的配置
		_ = existingKc.Delete(params, fmt.Sprintf("%d", existing.ID))
	}

	// 从解析的config中获取信息
	contextName := config.CurrentContext
	context := config.Contexts[contextName]
	cluster := config.Clusters[context.Cluster]

	// 创建KubeConfig记录
	kc := &models.KubeConfig{
		Cluster:     context.Cluster,
		Server:      cluster.Server,
		User:        context.AuthInfo,
		Namespace:   context.Namespace,
		Content:     kubeconfigContent,
		DisplayName: req.DisplayName,
		Token:       req.Token,
		CACert:      req.CaData,
		IsAWSEKS:    false, // 明确标识这不是AWS EKS集群
	}

	// 保存到数据库
	err = kc.Save(params)
	if err != nil {
		klog.V(6).Infof("保存Token集群 [%s]失败: %v", req.Server, err)
		amis.WriteJsonError(c, err)
		return
	}

	klog.V(4).Infof("成功保存Token集群配置: %s [%s]", req.DisplayName, req.Server)

	// 执行集群扫描以注册新集群
	service.ClusterService().ScanClustersInDB()

	amis.WriteJsonOKMsg(c, "Token方式集群纳管成功")
}
