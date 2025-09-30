package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/constants"
	"gorm.io/gorm"
)

// AIPrompt AI提示词模型
// 用于存储和管理AI提示词的配置信息
type AIPrompt struct {
	ID          uint                   `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	Name        string                 `json:"name" gorm:"size:100;not null"`             // 提示词名称
	Description string                 `json:"description" gorm:"size:500"`               // 提示词描述
	PromptType  constants.AIPromptType `json:"prompt_type" gorm:"size:50;not null;index"` // 提示词类型
	Content     string                 `json:"content" gorm:"type:text;not null"`         // 提示词内容
	IsBuiltin   bool                   `json:"is_builtin" gorm:"default:false;index"`     // 是否为内置提示词
	IsEnabled   bool                   `json:"is_enabled" gorm:"default:false;index"`     // 是否启用
	CreatedAt   time.Time              `json:"created_at,omitempty" gorm:"<-:create"`
	UpdatedAt   time.Time              `json:"updated_at,omitempty"`
}

// TableName 表名
func (AIPrompt) TableName() string {
	return "ai_prompts"
}

// List 获取AI提示词列表
func (m *AIPrompt) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*AIPrompt, int64, error) {
	return dao.GenericQuery(params, m, queryFuncs...)
}

// Save 保存AI提示词
func (m *AIPrompt) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, m, queryFuncs...)
}

// Delete 删除AI提示词
func (m *AIPrompt) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, m, utils.ToInt64Slice(ids), queryFuncs...)
}

// GetOne 获取单个AI提示词
func (m *AIPrompt) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*AIPrompt, error) {
	return dao.GenericGetOne(params, m, queryFuncs...)
}

// BuiltinAIPrompts 内置AI提示词
var BuiltinAIPrompts = []AIPrompt{
	{
		Name:        "K8s事件分析",
		Description: "分析Kubernetes事件信息，提供问题诊断和解决建议",
		PromptType:  constants.AIPromptTypeEvent,
		Content: `请你作为k8s专家，对下面的Event做出分析:
				note:   ${Note},
				source: ${Source},
				reason: ${Reason},
				type:   ${Type},
				kind:   ${RegardingKind},
`,
		IsBuiltin: true,
		IsEnabled: true,
	},
	{
		Name:        "K8s资源描述分析",
		Description: "分析Kubernetes资源的describe信息，识别问题并提供解决方案",
		PromptType:  constants.AIPromptTypeDescribe,
		Content: `我正在查看关于k8s ${Group} ${Kind} 资源的Describe (kubectl describe )信息。
		请你作为kubernetes k8s 技术专家，对这个describe的文本进行分析。
		\n 请给出分析结论，如果有问题，请指出问题，并给出可能得解决方案。
		\n注意：
		\n0、使用中文进行回答。
		\n1、你我之间只进行这一轮交互，后面不要再问问题了。
		\n2、请你在给出答案前反思下回答是否逻辑正确，如有问题请先修正，再返回。回答要直接，不要加入上下衔接、开篇语气词、结尾语气词等啰嗦的信息。
		\n3、请不要向我提问，也不要向我确认信息，请不要让我检查markdown格式，不要让我确认markdown格式是否正确。
		\n\nDescribe信息如下：${DescribeInfo}`,
		IsBuiltin: true,
		IsEnabled: true,
	},
	{
		Name:        "K8s资源使用示例",
		Description: "提供Kubernetes资源的详细使用指南和YAML示例",
		PromptType:  constants.AIPromptTypeExample,
		Content: `我正在浏览k8s资源管理页面，资源定义Kind=${Kind},Group=${Group},version=${Version}。
		\n请你作为kubernetes k8s 技术专家，给我一份关于这个k8s资源的使用指南。
		\n要求包括资源说明、使用场景（举例说明）、最佳实践、典型示例（配合前面的场景举例，编写yaml文件，每一行yaml都增加简体中文注释）、关键字段及其含义、常见问题、官方文档链接、引用文档链接等你认为对我有帮助的信息。
		\n最后给出一份关于这个资源的yaml样例。
		\n要求先假设一个简单场景、一个复杂场景。1、分别概要介绍这两个场景，2、为这两个场景书写yaml文件，每一行yaml都增加简体中文注释。
		\n注意：
		\n0、使用中文进行回答。
		\n1、你我之间只进行这一轮交互，后面不要再问问题了。
		\n2、请你在给出答案前反思下回答是否逻辑正确，如有问题请先修正，再返回。回答要直接，不要加入上下衔接、开篇语气词、结尾语气词等啰嗦的信息。
		\n3、请不要向我提问，也不要向我确认信息，请不要让我检查markdown格式，不要让我确认markdown格式是否正确`,
		IsBuiltin: true,
		IsEnabled: true,
	},
	{
		Name:        "K8s字段使用示例",
		Description: "详细解释Kubernetes资源中特定字段的用法和示例",
		PromptType:  constants.AIPromptTypeFieldExample,
		Content: `我正在浏览k8s资源管理页面，资源定义Kind=${Kind},Group=${Group},version=${Version}。
		\n请你作为kubernetes k8s 技术专家，给出一份关于  ${Field}  这个具体字段的使用场景。请在回答中使用 "该字段" 代替这个具体的字段。
		请详细解释该字段的含义、用法、并给出一个假设的使用场景，为这个场景书写yaml文件，每一行yaml都增加简体中文注释。
		\n注意：
		\n0、使用中文进行回答。
		\n1、你我之间只进行这一轮交互，后面不要再问问题了。
		\n2、请你在给出答案前反思下回答是否逻辑正确，如有问题请先修正，再返回。回答要直接，不要加入上下衔接、开篇语气词、结尾语气词等啰嗦的信息。
		\n3、请不要向我提问，也不要向我确认信息，请不要让我检查markdown格式，不要让我确认markdown格式是否正确`,
		IsBuiltin: true,
		IsEnabled: true,
	},
	{
		Name:        "K8s资源使用指南",
		Description: "提供Kubernetes资源的完整使用指南",
		PromptType:  constants.AIPromptTypeResource,
		Content: `我正在浏览k8s资源管理页面，资源定义Kind=${Kind},Group=${Group},version=${Version}。
		\n请你作为kubernetes k8s 技术专家，给我一份关于这个k8s资源的使用指南。
		要求包括资源说明、使用场景（举例说明）、最佳实践、典型示例（配合前面的场景举例，编写yaml文件，每一行yaml都增加简体中文注释）、关键字段及其含义、常见问题、官方文档链接、引用文档链接等你认为对我有帮助的信息。
		\n注意：
		\n0、使用中文进行回答。
		\n1、你我之间只进行这一轮交互，后面不要再问问题了。
		\n2、请你在给出答案前反思下回答是否逻辑正确，如有问题请先修正，再返回。回答要直接，不要加入上下衔接、开篇语气词、结尾语气词等啰嗦的信息。
		\n3、请不要向我提问，也不要向我确认信息，请不要让我检查markdown格式，不要让我确认markdown格式是否正确`,
		IsBuiltin: true,
		IsEnabled: true,
	},
	{
		Name:        "K8s错误信息分析",
		Description: "分析Kubernetes错误信息并提供解决方案",
		PromptType:  constants.AIPromptTypeK8sGPTResource,
		Content: `简化以下由三个破折号分隔的Kubernetes错误信息，
	错误内容：--- ${Data} ---。
	资源名称：--- ${Name} ---。
	资源类型：--- ${Kind} ---。
	相关字段k8s官方文档解释：--- ${Field} ---。
	请以分步形式提供最可能的解决方案，字符数不超过280。
	输出格式：
	错误信息: {此处解释错误}
	解决方案: {此处分步说明解决方案}
		\n注意：
		\n0、使用中文进行回答。
		\n1、你我之间只进行这一轮交互，后面不要再问问题了。
		\n2、请你在给出答案前反思下回答是否逻辑正确，如有问题请先修正，再返回。回答要直接，不要加入上下衔接、开篇语气词、结尾语气词等啰嗦的信息。
		\n3、请不要向我提问，也不要向我确认信息，请不要让我检查markdown格式，不要让我确认markdown格式是否正确`,
		IsBuiltin: true,
		IsEnabled: true,
	},

	{
		Name:        "K8s问题解答",
		Description: "回答Kubernetes相关的任意问题",
		PromptType:  constants.AIPromptTypeAnyQuestion,
		Content: `我正在浏览k8s资源管理页面，资源定义Kind=${Kind},Group=${Group},version=${Version}。
		\n请你作为kubernetes k8s 技术专家，请你详细解释下我的疑问： %s 。
		\n要求包括关键名词解释、作用、典型示例（以场景举例，编写yaml文件，每一行yaml都增加简体中文注释）、关键字段及其含义、常见问题、官方文档链接、引用文档链接等你认为对我有帮助的信息。
		\n注意：
		\n0、使用中文进行回答。
		\n1、你我之间只进行这一轮交互，后面不要再问问题了。
		\n2、请你在给出答案前反思下回答是否逻辑正确，如有问题请先修正，再返回。回答要直接，不要加入上下衔接、开篇语气词、结尾语气词等啰嗦的信息。
		\n3、请不要向我提问，也不要向我确认信息，请不要让我检查markdown格式，不要让我确认markdown格式是否正确`,
		IsBuiltin: true,
		IsEnabled: true,
	},
	{
		Name:        "Cron表达式分析",
		Description: "分析和解释Cron表达式的含义",
		PromptType:  constants.AIPromptTypeCron,
		Content: `我正在查看k8s cronjob 中的schedule 表达式：${Cron}。
		\n请你作为k8s技术专家，对 ${Cron} 这个表达式进行分析，给出详细的解释。
		\n注意：
		\n0、使用中文进行回答。
		\n1、你我之间只进行这一轮交互，后面不要再问问题了。
		\n2、请你在给出答案前反思下回答是否逻辑正确，如有问题请先修正，再返回。回答要直接，不要加入上下衔接、开篇语气词、结尾语气词等啰嗦的信息。
		\n3、请不要向我提问，也不要向我确认信息，请不要让我检查markdown格式，不要让我确认markdown格式是否正确`,
		IsBuiltin: true,
		IsEnabled: true,
	},
	{
		Name:        "日志分析",
		Description: "分析应用程序或系统日志",
		PromptType:  constants.AIPromptTypeLog,
		Content: `请你作为k8s、Devops、软件工程专家，对下面的Log做出分析:
		\n${Data}
		\n请提供：
		\n1. 日志级别和类型分析
		\n2. 关键信息提取
		\n3. 问题识别和诊断
		\n4. 解决建议和后续行动
		\n注意：
		\n- 使用中文进行回答
		\n- 回答要直接，不要加入啰嗦的信息
		\n- 不要向我提问或确认信息`,
		IsBuiltin: true,
		IsEnabled: true,
	},
}
