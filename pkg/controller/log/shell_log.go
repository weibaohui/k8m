package log

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"gorm.io/gorm"
)

type Controller struct{}

func RegisterLogRoutes(mgm *gin.RouterGroup) {
	ctrl := &Controller{}
	mgm.GET("/log/shell/list", ctrl.ListShell)
	mgm.GET("/log/operation/list", ctrl.ListOperation)
	mgm.GET("/log/global/list", ctrl.ListGlobalLog)
}

// @Summary Shell日志列表
// @Description 获取所有Shell操作日志
// @Security BearerAuth
// @Success 200 {object} string
// @Router /mgm/log/shell/list [get]
func (lc *Controller) ListShell(c *gin.Context) {
	params := dao.BuildParams(c)
	m := &models.ShellLog{}

	// 处理时间范围查询
	var queryFuncs []func(*gorm.DB) *gorm.DB
	if createdAtRange, exists := params.Queries["created_at_range"]; exists && createdAtRange != "" {
		// 解析时间范围参数，格式为 "startTime,endTime"
		timeRange := fmt.Sprintf("%v", createdAtRange)
		if strings.Contains(timeRange, ",") {
			times := strings.Split(timeRange, ",")
			if len(times) == 2 {
				startTimeStr := strings.TrimSpace(times[0])
				endTimeStr := strings.TrimSpace(times[1])
				
				queryFuncs = append(queryFuncs, func(db *gorm.DB) *gorm.DB {
					if startTimeStr != "" {
						if startTime, err := time.Parse("2006-01-02 15:04:05", startTimeStr); err == nil {
							db = db.Where("created_at >= ?", startTime)
						}
					}
					if endTimeStr != "" {
						if endTime, err := time.Parse("2006-01-02 15:04:05", endTimeStr); err == nil {
							db = db.Where("created_at <= ?", endTime)
						}
					}
					return db
				})
			}
		}
		// 从查询参数中移除 created_at_range，避免在通用查询中被处理
		delete(params.Queries, "created_at_range")
	}

	items, total, err := m.List(params, queryFuncs...)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}

// @Summary 操作日志列表
// @Description 获取所有操作日志
// @Security BearerAuth
// @Success 200 {object} string
// @Router /mgm/log/operation/list [get]
func (lc *Controller) ListOperation(c *gin.Context) {
	params := dao.BuildParams(c)
	m := &models.OperationLog{}

	// 处理时间范围查询
	var queryFuncs []func(*gorm.DB) *gorm.DB
	if createdAtRange, exists := params.Queries["created_at_range"]; exists && createdAtRange != "" {
		// 解析时间范围参数，格式为 "startTime,endTime"
		timeRange := fmt.Sprintf("%v", createdAtRange)
		if strings.Contains(timeRange, ",") {
			times := strings.Split(timeRange, ",")
			if len(times) == 2 {
				startTimeStr := strings.TrimSpace(times[0])
				endTimeStr := strings.TrimSpace(times[1])
				
				queryFuncs = append(queryFuncs, func(db *gorm.DB) *gorm.DB {
					if startTimeStr != "" {
						if startTime, err := time.Parse("2006-01-02 15:04:05", startTimeStr); err == nil {
							db = db.Where("created_at >= ?", startTime)
						}
					}
					if endTimeStr != "" {
						if endTime, err := time.Parse("2006-01-02 15:04:05", endTimeStr); err == nil {
							db = db.Where("created_at <= ?", endTime)
						}
					}
					return db
				})
			}
		}
		// 从查询参数中移除 created_at_range，避免在通用查询中被处理
		delete(params.Queries, "created_at_range")
	}

	items, total, err := m.List(params, queryFuncs...)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}
