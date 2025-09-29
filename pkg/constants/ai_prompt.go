package constants

// AIPromptType AI提示词类型
type AIPromptType string

const (
	// AIPromptTypeEvent 事件分析类型
	AIPromptTypeEvent AIPromptType = "Event"
	
	// AIPromptTypeDescribe 资源描述分析类型
	AIPromptTypeDescribe AIPromptType = "Describe"
	
	// AIPromptTypeExample 示例类型
	AIPromptTypeExample AIPromptType = "Example"
	
	// AIPromptTypeFieldExample 字段示例类型
	AIPromptTypeFieldExample AIPromptType = "FieldExample"
	
	// AIPromptTypeResource 资源类型
	AIPromptTypeResource AIPromptType = "Resource"
	
	// AIPromptTypeK8sGPTResource K8sGPT资源类型
	AIPromptTypeK8sGPTResource AIPromptType = "K8sGPTResource"
	
	// AIPromptTypeAnySelection 任意选择类型
	AIPromptTypeAnySelection AIPromptType = "AnySelection"
	
	// AIPromptTypeAnyQuestion 任意问题类型
	AIPromptTypeAnyQuestion AIPromptType = "AnyQuestion"
	
	// AIPromptTypeCron Cron表达式类型
	AIPromptTypeCron AIPromptType = "Cron"
	
	// AIPromptTypeLog 日志分析类型
	AIPromptTypeLog AIPromptType = "Log"
)

// AIPromptCategory AI提示词分类
type AIPromptCategory string

const (
	// AIPromptCategoryDiagnosis 诊断分析分类
	AIPromptCategoryDiagnosis AIPromptCategory = "诊断分析"
	
	// AIPromptCategoryGuide 使用指南分类
	AIPromptCategoryGuide AIPromptCategory = "使用指南"
	
	// AIPromptCategoryError 错误诊断分类
	AIPromptCategoryError AIPromptCategory = "错误诊断"
	
	// AIPromptCategoryGeneral 通用问答分类
	AIPromptCategoryGeneral AIPromptCategory = "通用问答"
	
	// AIPromptCategoryTool 工具分析分类
	AIPromptCategoryTool AIPromptCategory = "工具分析"
)