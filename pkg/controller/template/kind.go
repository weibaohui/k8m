package template

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"gorm.io/gorm"
)

func ListKind(c *gin.Context) {
	params := dao.BuildParams(c)
	m := &models.CustomTemplate{}
	params.PerPage = 1000
	items, total, err := m.List(params, func(db *gorm.DB) *gorm.DB {
		return db.Select("Kind").Distinct("Kind")
	})
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}
