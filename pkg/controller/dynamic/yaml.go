package dynamic

import (
	"fmt"
	"io"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/response"
	"github.com/weibaohui/kom/kom"
)

type YamlController struct{}

func RegisterYamlRoutes(api chi.Router) {
	ctrl := &YamlController{}
	api.Post("/yaml/apply", response.Adapter(ctrl.Apply))
	api.Post("/yaml/upload", response.Adapter(ctrl.UploadFile))
	api.Post("/yaml/delete", response.Adapter(ctrl.Delete))
}

// @Summary 上传YAML文件并应用
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param file formData file true "YAML文件"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/yaml/upload [post]
func (yc *YamlController) UploadFile(c *response.Context) {
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	ctx := amis.GetContextWithUser(c)
	// 获取上传的文件
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

// @Summary 应用YAML配置
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param body body yamlRequest true "YAML配置请求"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/yaml/apply [post]
func (yc *YamlController) Apply(c *response.Context) {
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

// @Summary 删除YAML配置
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param body body yamlRequest true "YAML配置请求"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/yaml/delete [post]
func (yc *YamlController) Delete(c *response.Context) {
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
