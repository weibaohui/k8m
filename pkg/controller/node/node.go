package node

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
)

func Drain(c *gin.Context) {
	name := c.Param("name")
	ctx := c.Request.Context()
	err := kom.DefaultCluster().WithContext(ctx).Resource(&v1.Node{}).Name(name).
		Ctl().Node().Drain()
	amis.WriteJsonErrorOrOK(c, err)
}
func Cordon(c *gin.Context) {
	name := c.Param("name")
	ctx := c.Request.Context()
	err := kom.DefaultCluster().WithContext(ctx).Resource(&v1.Node{}).Name(name).
		Ctl().Node().Cordon()
	amis.WriteJsonErrorOrOK(c, err)
}
func Usage(c *gin.Context) {
	name := c.Param("name")
	ctx := c.Request.Context()
	usage := kom.DefaultCluster().WithContext(ctx).Resource(&v1.Node{}).Name(name).
		Ctl().Node().ResourceUsage()
	data, err := convertToTableData(usage)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, data)
}

// 临时结构体，用于存储每一行数据
type ResourceUsageRow struct {
	ResourceType    string `json:"resourceType"`
	Total           string `json:"total"`
	Request         string `json:"request"`
	RequestFraction string `json:"requestFraction"`
	Limit           string `json:"limit"`
	LimitFraction   string `json:"limitFraction"`
}

func convertToTableData(result *kom.ResourceUsageResult) ([]*ResourceUsageRow, error) {
	var tableData []*ResourceUsageRow

	// 遍历资源类型（CPU、内存等），并生成表格行
	for _, resourceType := range []v1.ResourceName{v1.ResourceCPU, v1.ResourceMemory} {
		// 创建一行数据
		quantity := result.Allocatable[resourceType]
		req := result.Requests[resourceType]
		lit := result.Limits[resourceType]
		row := ResourceUsageRow{
			ResourceType:    string(resourceType),
			Total:           quantity.String(),
			Request:         req.String(),
			RequestFraction: fmt.Sprintf("%.2f", result.UsageFractions[resourceType].RequestFraction),
			Limit:           lit.String(),
			LimitFraction:   fmt.Sprintf("%.2f", result.UsageFractions[resourceType].LimitFraction),
		}

		// 将行加入表格数据
		tableData = append(tableData, &row)
	}

	// 如果存储资源需要处理，可以按类似方式扩展
	quantity := result.Allocatable[v1.ResourceEphemeralStorage]
	storageLimit := result.Limits[v1.ResourceEphemeralStorage]
	storageReq := result.Requests[v1.ResourceEphemeralStorage]
	row := &ResourceUsageRow{
		ResourceType:    "存储",
		Total:           quantity.String(),
		Request:         storageReq.String(),
		RequestFraction: fmt.Sprintf("%.2f", result.UsageFractions[v1.ResourceEphemeralStorage].RequestFraction),
		Limit:           storageLimit.String(),
		LimitFraction:   fmt.Sprintf("%.2f", result.UsageFractions[v1.ResourceEphemeralStorage].LimitFraction),
	}
	tableData = append(tableData, row)
	return tableData, nil
}
func UnCordon(c *gin.Context) {
	name := c.Param("name")
	ctx := c.Request.Context()
	err := kom.DefaultCluster().WithContext(ctx).Resource(&v1.Node{}).Name(name).
		Ctl().Node().UnCordon()
	amis.WriteJsonErrorOrOK(c, err)
}
