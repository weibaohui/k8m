package ai_prompt

import (
	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/models"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

// AdminAIPromptController AI提示词管理控制器
// 提供AI提示词的增删改查功能
type AdminAIPromptController struct {
}

// RegisterAdminAIPromptRoutes 注册AI提示词管理路由
func RegisterAdminAIPromptRoutes(admin *gin.RouterGroup) {
	ctrl := &AdminAIPromptController{}
	admin.GET("/ai_prompt/list", ctrl.AIPromptList)
	admin.POST("/ai_prompt/delete/:ids", ctrl.AIPromptDelete)
	admin.POST("/ai_prompt/save", ctrl.AIPromptSave)
	admin.POST("/ai_prompt/load", ctrl.AIPromptLoad)
	admin.GET("/ai_prompt/option_list", ctrl.AIPromptOptionList)
	admin.GET("/ai_prompt/types", ctrl.AIPromptTypes)
	admin.GET("/ai_prompt/categories", ctrl.AIPromptCategories)
	admin.GET("/ai_prompt/category_list", ctrl.AIPromptCategories) // 添加category_list路由
	admin.POST("/ai_prompt/toggle/:id", ctrl.AIPromptToggle)       // 添加启用/禁用路由
}

// @Summary 获取AI提示词列表
// @Security BearerAuth
// @Success 200 {object} string
// @Router /admin/ai_prompt/list [get]
func (s *AdminAIPromptController) AIPromptList(c *gin.Context) {
	params := dao.BuildParams(c)
	m := &models.AIPrompt{}

	// 构建查询函数，支持按类型筛选
	var queryFuncs []func(*gorm.DB) *gorm.DB
	
	// 检查是否有类型筛选参数
	if promptType := c.Query("prompt_type"); promptType != "" {
		queryFuncs = append(queryFuncs, func(db *gorm.DB) *gorm.DB {
			return db.Where("prompt_type = ?", promptType)
		})
		// 从params.Queries中删除prompt_type，避免重复筛选
		delete(params.Queries, "prompt_type")
	}

	items, total, err := m.List(params, queryFuncs...)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}

// @Summary 保存AI提示词
// @Security BearerAuth
// @Success 200 {object} string
// @Router /admin/ai_prompt/save [post]
func (s *AdminAIPromptController) AIPromptSave(c *gin.Context) {
	params := dao.BuildParams(c)
	m := models.AIPrompt{}
	err := c.ShouldBindJSON(&m)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 如果是新增且未指定是否内置，默认为自定义
	if m.ID == 0 && !m.IsBuiltin {
		m.IsBuiltin = false
	}

	err = m.Save(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOK(c)
}

// @Summary 删除AI提示词
// @Security BearerAuth
// @Param ids path string true "提示词ID，多个用逗号分隔"
// @Success 200 {object} string
// @Router /admin/ai_prompt/delete/{ids} [post]
func (s *AdminAIPromptController) AIPromptDelete(c *gin.Context) {
	ids := c.Param("ids")
	params := dao.BuildParams(c)
	params.UserName = ""

	m := &models.AIPrompt{}

	// 只允许删除非内置的提示词
	queryFunc := func(db *gorm.DB) *gorm.DB {
		return db.Where("is_builtin = ?", false)
	}

	err := m.Delete(params, ids, queryFunc)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

// @Summary 获取AI提示词选项列表
// @Security BearerAuth
// @Success 200 {object} string
// @Router /admin/ai_prompt/option_list [get]
func (s *AdminAIPromptController) AIPromptOptionList(c *gin.Context) {
	m := models.AIPrompt{}
	params := dao.BuildParams(c)
	params.PerPage = 100000

	// 只获取启用的提示词
	queryFunc := func(db *gorm.DB) *gorm.DB {
		return db.Where("is_enabled = ?", true)
	}

	list, _, err := m.List(params, queryFunc)

	if err != nil {
		amis.WriteJsonData(c, gin.H{
			"options": make([]map[string]string, 0),
		})
		return
	}

	var prompts []map[string]string
	for _, n := range list {
		prompts = append(prompts, map[string]string{
			"label":       n.Name,
			"value":       n.PromptCode,
			"prompt_code": n.PromptCode,
			"name":        n.Name,
			"description": n.Description,
			"prompt_type": string(n.PromptType),
			"category":    string(n.Category),
		})
	}
	slice.SortBy(prompts, func(a, b map[string]string) bool {
		return a["label"] < b["label"]
	})
	amis.WriteJsonData(c, gin.H{
		"options": prompts,
	})
}

// @Summary 加载内置AI提示词
// @Security BearerAuth
// @Success 200 {object} string
// @Router /admin/ai_prompt/load [post]
func (s *AdminAIPromptController) AIPromptLoad(c *gin.Context) {
	// 删除后，重新插入内置提示词
	err := dao.DB().Model(&models.AIPrompt{}).Where("is_builtin = ?", true).Delete(&models.AIPrompt{}).Error
	if err != nil {
		klog.Errorf("删除内置AI提示词失败: %v", err)
		amis.WriteJsonError(c, err)
		return
	}
	err = dao.DB().Model(&models.AIPrompt{}).CreateInBatches(models.BuiltinAIPrompts, 100).Error
	if err != nil {
		klog.Errorf("插入内置AI提示词失败: %v", err)
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

// @Summary 获取AI提示词类型列表
// @Security BearerAuth
// @Success 200 {object} string
// @Router /admin/ai_prompt/types [get]
func (s *AdminAIPromptController) AIPromptTypes(c *gin.Context) {
	types := []map[string]string{
		{"label": "事件分析", "value": string(constants.AIPromptTypeEvent)},
		{"label": "资源描述", "value": string(constants.AIPromptTypeDescribe)},
		{"label": "示例说明", "value": string(constants.AIPromptTypeExample)},
		{"label": "字段示例", "value": string(constants.AIPromptTypeFieldExample)},
		{"label": "资源分析", "value": string(constants.AIPromptTypeResource)},
		{"label": "K8sGPT资源", "value": string(constants.AIPromptTypeK8sGPTResource)},
		{"label": "任意选择", "value": string(constants.AIPromptTypeAnySelection)},
		{"label": "任意问题", "value": string(constants.AIPromptTypeAnyQuestion)},
		{"label": "定时任务", "value": string(constants.AIPromptTypeCron)},
		{"label": "日志分析", "value": string(constants.AIPromptTypeLog)},
	}
	amis.WriteJsonData(c, gin.H{
		"options": types,
	})
}

// @Summary 获取AI提示词分类列表
// @Security BearerAuth
// @Success 200 {object} string
// @Router /admin/ai_prompt/categories [get]
func (s *AdminAIPromptController) AIPromptCategories(c *gin.Context) {
	categories := []map[string]string{
		{"label": "诊断分析", "value": string(constants.AIPromptCategoryDiagnosis)},
		{"label": "操作指南", "value": string(constants.AIPromptCategoryGuide)},
		{"label": "错误处理", "value": string(constants.AIPromptCategoryError)},
		{"label": "通用功能", "value": string(constants.AIPromptCategoryGeneral)},
		{"label": "工具辅助", "value": string(constants.AIPromptCategoryTool)},
	}
	amis.WriteJsonData(c, gin.H{
		"options": categories,
	})
}

// @Summary 启用/禁用AI提示词
// @Security BearerAuth
// @Param id path string true "提示词ID"
// @Success 200 {object} string
// @Router /admin/ai_prompt/toggle/{id} [post]
func (s *AdminAIPromptController) AIPromptToggle(c *gin.Context) {
	id := c.Param("id")
	params := dao.BuildParams(c)
	
	// 获取当前提示词
	m := &models.AIPrompt{}
	currentPrompt, err := m.GetOne(params, func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	})
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 如果要启用此提示词，需要先禁用同类型的其他提示词
	if !currentPrompt.IsEnabled {
		// 禁用同类型的其他提示词
		err = dao.DB().Model(&models.AIPrompt{}).
			Where("prompt_type = ? AND id != ? AND is_enabled = ?", currentPrompt.PromptType, id, true).
			Update("is_enabled", false).Error
		if err != nil {
			klog.Errorf("禁用同类型提示词失败: %v", err)
			amis.WriteJsonError(c, err)
			return
		}
	}

	// 切换当前提示词的启用状态
	newStatus := !currentPrompt.IsEnabled
	err = dao.DB().Model(&models.AIPrompt{}).Where("id = ?", id).Update("is_enabled", newStatus).Error
	if err != nil {
		klog.Errorf("更新提示词状态失败: %v", err)
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOK(c)
}