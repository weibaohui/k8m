package dynamic

import (
	"fmt"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
)

type YamlController struct{}

func RegisterYamlRoutes(api *gin.RouterGroup) {
	ctrl := &YamlController{}
	api.POST("/yaml/apply", ctrl.Apply)
	api.POST("/yaml/upload", ctrl.UploadFile)
	api.POST("/yaml/delete", ctrl.Delete)
}

func (yc *YamlController) UploadFile(c *gin.Context) {
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

func (yc *YamlController) Apply(c *gin.Context) {
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
	amis.WriteJsonData(c, gin.H{
		"result": result,
	})

}
func (yc *YamlController) Delete(c *gin.Context) {
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
	amis.WriteJsonData(c, gin.H{
		"result": result,
	})
}
