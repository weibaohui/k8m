package param

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
)

// @Summary 翻转指标列表
// @Description 获取所有启用的翻转显示指标名称
// @Security BearerAuth
// @Success 200 {object} string
// @Router /params/condition/reverse/list [get]
func (pc *Controller) Conditions(c *gin.Context) {
	var list []*models.ConditionReverse
	err := dao.DB().Model(&models.ConditionReverse{}).
		Select("name").
		Where("enabled = ?", true).
		Find(&list).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	var names []string
	for _, item := range list {
		names = append(names, item.Name)
	}
	amis.WriteJsonData(c, names)
}
