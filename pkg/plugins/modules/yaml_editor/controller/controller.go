package controller

import (
	"fmt"
	"io"
	"strings"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/plugins/modules/yaml_editor/models"
	"github.com/weibaohui/k8m/pkg/response"
	"github.com/weibaohui/kom/kom"
)

type Controller struct{}

type yamlRequest struct {
	Yaml string `json:"yaml"`
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
