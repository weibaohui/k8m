package param

import (
	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"gorm.io/gorm"
)

func RepoOptionList(c *gin.Context) {
	params := dao.BuildParams(c)
	params.OrderBy = "name"
	m := &models.HelmRepository{}
	items, _, err := m.List(params, func(db *gorm.DB) *gorm.DB {
		return db.Distinct("name")
	})
	if err != nil {
		amis.WriteJsonData(c, gin.H{
			"options": make([]map[string]string, 0),
		})
		return
	}
	var names []map[string]string
	for _, n := range items {
		names = append(names, map[string]string{
			"label": n.Name,
			"value": n.Name,
		})
	}
	slice.SortBy(names, func(a, b map[string]string) bool {
		return a["label"] < b["label"]
	})
	amis.WriteJsonData(c, gin.H{
		"options": names,
	})
}
