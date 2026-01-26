package controller

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/plugins/modules/ai/service"
	"github.com/weibaohui/k8m/pkg/plugins/modules/yaml_editor/models"
	"github.com/weibaohui/k8m/pkg/response"
	"github.com/weibaohui/kom/kom"
)

type Controller struct{}

type yamlRequest struct {
	Yaml string `json:"yaml"`
}

type aiGenerateRequest struct {
	Prompt string `json:"prompt"`
}

// AIGenerate 使用 AI 生成 Kubernetes YAML 配置
func (yc *Controller) AIGenerate(c *response.Context) {
	var req aiGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, fmt.Errorf("提取prompt错误。\n %v", err))
		return
	}

	// 构建 AI 提示词
	prompt := fmt.Sprintf(`你是一个 Kubernetes 专家。请根据以下描述生成准确、完整的 Kubernetes YAML 配置。

要求：
1. 只返回 YAML 代码，不要包含任何解释、注释或其他文本
2. YAML 格式必须正确，缩进使用 2 个空格
3. 包含所有必需的字段（apiVersion, kind, metadata, spec 等）
4. 如果描述涉及多个资源，请使用 YAML 文档分隔符 "---" 分隔
5. 确保资源名称和标签符合 Kubernetes 命名规范

用户描述：%s

请直接返回 YAML 代码：`, req.Prompt)

	ctx := context.Background()
	result, err := service.GetChatService().ChatWithCtxNoHistory(ctx, prompt)
	if err != nil {
		amis.WriteJsonError(c, fmt.Errorf("AI 生成失败：%v", err))
		return
	}

	// 清理可能存在的 markdown 代码块标记
	result = strings.TrimSpace(result)
	result = strings.TrimPrefix(result, "```yaml")
	result = strings.TrimPrefix(result, "```")
	result = strings.TrimSuffix(result, "```")
	result = strings.TrimSpace(result)

	amis.WriteJsonData(c, response.H{
		"yaml": result,
	})
}

func (yc *Controller) UploadFile(c *response.Context) {
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	ctx := amis.GetContextWithUser(c)
	file, err := c.FormFile("file")
	if err != nil {
		amis.WriteJsonError(c, fmt.Errorf("获取上传的文件错误。\n %v", err))
		return
	}
	src, err := file.Open()
	if err != nil {
		amis.WriteJsonError(c, fmt.Errorf("打开上传的文件错误。\n %v", err))
		return
	}
	defer src.Close()
	yamlBytes, err := io.ReadAll(src)
	if err != nil {
		amis.WriteJsonError(c, fmt.Errorf("读取上传的文件内容错误。\n %v", err))
		return
	}
	yamlStr := string(yamlBytes)
	result := kom.Cluster(selectedCluster).WithContext(ctx).Applier().Apply(yamlStr)
	amis.WriteJsonOKMsg(c, strings.Join(result, "\n"))
}

func (yc *Controller) Apply(c *response.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var req yamlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, fmt.Errorf("提取yaml错误。\n %v", err))
		return
	}
	yamlStr := req.Yaml
	result := kom.Cluster(selectedCluster).WithContext(ctx).Applier().Apply(yamlStr)
	amis.WriteJsonData(c, response.H{
		"result": result,
	})
}

func (yc *Controller) Delete(c *response.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var req yamlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	yamlStr := req.Yaml
	result := kom.Cluster(selectedCluster).WithContext(ctx).Applier().Delete(yamlStr)
	amis.WriteJsonData(c, response.H{
		"result": result,
	})
}

func (t *Controller) List(c *response.Context) {
	params := dao.BuildParams(c)
	m := &models.Template{}

	items, total, err := m.List(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}

func (t *Controller) Save(c *response.Context) {
	params := dao.BuildParams(c)
	m := models.Template{}
	err := c.ShouldBindJSON(&m)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	if m.Kind == "" {
		m.Kind = "未分类"
	}

	err = m.Save(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, response.H{
		"id": m.ID,
	})
}

func (t *Controller) DeleteTemplate(c *response.Context) {
	ids := c.Param("ids")
	params := dao.BuildParams(c)
	m := &models.Template{}
	err := m.Delete(params, ids)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

// ListKind 获取模板分类列表，用于左侧导航栏渲染。
// 返回格式与 AI 提示词类型接口保持一致：{ data: { options: [{label,value}, ...] } }。
func (t *Controller) ListKind(c *response.Context) {
	var kinds []struct {
		Kind string `json:"kind"`
	}

	err := dao.DB().Model(&models.Template{}).
		Select("kind").
		Where("kind != ?", "").
		Group("kind").
		Order("kind ASC").
		Scan(&kinds).Error

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	options := make([]map[string]string, 0, len(kinds)+1)
	options = append(options, map[string]string{
		"label": "全部",
		"value": "all",
	})
	for _, k := range kinds {
		if k.Kind == "" {
			continue
		}
		options = append(options, map[string]string{
			"label": k.Kind,
			"value": k.Kind,
		})
	}
	amis.WriteJsonData(c, response.H{
		"options": options,
	})
}
