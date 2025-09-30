package models

import (
	"github.com/weibaohui/k8m/pkg/constants"
)

// BuiltinAIPromptsVersion 统一管理所有内置AI提示词的版本号
const BuiltinAIPromptsVersion = "v1"

// BuiltinAIPromptsExtended 扩展的内置AI提示词列表
var BuiltinAIPromptsExtended = []AIPrompt{
	{
		Name:        "事件分析提示词",
		Description: "用于分析Kubernetes事件的AI提示词",
		PromptType:  constants.AIPromptTypeEvent,
		Content: `你是一个Kubernetes专家，请分析以下事件信息：

事件类型：{{.Type}}
事件原因：{{.Reason}}
事件消息：{{.Message}}
涉及对象：{{.InvolvedObject.Kind}}/{{.InvolvedObject.Name}}
命名空间：{{.InvolvedObject.Namespace}}
发生时间：{{.FirstTimestamp}}

请提供：
1. 事件的严重程度评估
2. 可能的原因分析
3. 具体的解决建议
4. 预防措施

请用中文回答，并提供具体的kubectl命令示例。`,
		IsBuiltin: true,
		IsEnabled: true,
	},
	{
		Name:        "资源描述分析",
		Description: "用于分析Kubernetes资源描述信息的AI提示词",
		PromptType:  constants.AIPromptTypeDescribe,
		Content: `你是一个Kubernetes专家，请分析以下资源的详细信息：

资源类型：{{.Kind}}
资源名称：{{.Name}}
命名空间：{{.Namespace}}
资源状态：{{.Status}}
资源配置：
{{.Spec}}

事件信息：
{{.Events}}

请提供：
1. 资源当前状态的健康评估
2. 配置是否合理的分析
3. 发现的问题及其影响
4. 优化建议和最佳实践
5. 相关的故障排查步骤

请用中文回答，并提供具体的kubectl命令和YAML配置示例。`,
		IsBuiltin: true,
		IsEnabled: true,
	},
	{
		Name:        "配置示例生成",
		Description: "用于生成Kubernetes资源配置示例的AI提示词",
		PromptType:  constants.AIPromptTypeExample,
		Content: `你是一个Kubernetes专家，请根据以下需求生成配置示例：

资源类型：{{.ResourceType}}
应用名称：{{.AppName}}
命名空间：{{.Namespace}}
特殊要求：{{.Requirements}}

请提供：
1. 完整的YAML配置文件
2. 配置中每个重要字段的说明
3. 部署和验证的kubectl命令
4. 相关的最佳实践建议
5. 常见的配置陷阱和注意事项

请确保配置符合生产环境的安全和性能要求，用中文提供详细说明。`,
		IsBuiltin: true,
		IsEnabled: true,
	},
	{
		Name:        "字段配置指导",
		Description: "用于指导特定字段配置的AI提示词",
		PromptType:  constants.AIPromptTypeFieldExample,
		Content: `你是一个Kubernetes专家，请为以下字段提供配置指导：

资源类型：{{.ResourceType}}
字段路径：{{.FieldPath}}
字段描述：{{.FieldDescription}}
当前值：{{.CurrentValue}}
使用场景：{{.UseCase}}

请提供：
1. 该字段的详细说明和作用
2. 不同场景下的推荐配置值
3. 配置示例和最佳实践
4. 常见的配置错误和解决方法
5. 与其他字段的关联关系

请用中文回答，并提供具体的配置示例。`,
		IsBuiltin: true,
		IsEnabled: true,
	},
	{
		Name:        "资源状态分析",
		Description: "用于分析Kubernetes资源状态的AI提示词",
		PromptType:  constants.AIPromptTypeResource,
		Content: `你是一个Kubernetes专家，请分析以下资源的状态信息：

资源类型：{{.Kind}}
资源名称：{{.Name}}
命名空间：{{.Namespace}}
当前状态：{{.Status}}
期望状态：{{.DesiredState}}
资源年龄：{{.Age}}
标签：{{.Labels}}
注解：{{.Annotations}}

请提供：
1. 资源状态的健康评估
2. 状态异常的原因分析
3. 性能和资源使用情况评估
4. 安全配置检查
5. 维护和优化建议

请用中文回答，并提供具体的诊断和修复命令。`,
		IsBuiltin: true,
		IsEnabled: true,
	},
	{
		Name:        "K8sGPT资源分析",
		Description: "用于K8sGPT风格的资源分析AI提示词",
		PromptType:  constants.AIPromptTypeK8sGPTResource,
		Content: `作为Kubernetes诊断专家，请分析以下资源问题：

问题类型：{{.ProblemType}}
资源信息：{{.ResourceInfo}}
错误详情：{{.ErrorDetails}}
相关日志：{{.Logs}}
集群环境：{{.ClusterInfo}}

请按照以下格式提供分析：

🔍 **问题诊断**
- 问题根本原因
- 影响范围评估

🛠️ **解决方案**
- 立即修复步骤
- 长期优化建议

📋 **验证步骤**
- 修复后的验证命令
- 监控指标检查

⚠️ **预防措施**
- 避免类似问题的配置建议
- 监控和告警设置

请用中文回答，提供具体可执行的命令和配置。`,
		IsBuiltin: true,
		IsEnabled: true,
	},
	{
		Name:        "任意选择分析",
		Description: "用于分析用户选择的任意内容的AI提示词",
		PromptType:  constants.AIPromptTypeAnySelection,
		Content: `你是一个Kubernetes专家，请分析用户选择的以下内容：

选择内容：
{{.SelectedContent}}

上下文信息：{{.Context}}
用户意图：{{.UserIntent}}

请提供：
1. 对选择内容的详细解释
2. 相关的Kubernetes概念说明
3. 可能存在的问题或改进点
4. 相关的最佳实践建议
5. 进一步的学习资源推荐

请用中文回答，并根据内容类型提供相应的示例和命令。`,
		IsBuiltin: true,
		IsEnabled: true,
	},
	{
		Name:        "任意问题解答",
		Description: "用于回答用户任意Kubernetes问题的AI提示词",
		PromptType:  constants.AIPromptTypeAnyQuestion,
		Content: `你是一个经验丰富的Kubernetes专家，请回答以下问题：

用户问题：{{.Question}}
相关上下文：{{.Context}}
用户技能水平：{{.UserLevel}}

请提供：
1. 问题的直接答案
2. 详细的技术解释
3. 实际的操作示例
4. 相关的最佳实践
5. 可能的替代方案
6. 进阶学习建议

请用中文回答，根据用户技能水平调整回答的深度和复杂度。如果涉及具体操作，请提供完整的kubectl命令和YAML配置示例。`,
		IsBuiltin: true,
		IsEnabled: true,
	},
	{
		Name:        "CronJob分析",
		Description: "用于分析CronJob配置和状态的AI提示词",
		PromptType:  constants.AIPromptTypeCron,
		Content: `你是一个Kubernetes专家，请分析以下CronJob的配置和状态：

CronJob名称：{{.Name}}
命名空间：{{.Namespace}}
调度表达式：{{.Schedule}}
暂停状态：{{.Suspend}}
并发策略：{{.ConcurrencyPolicy}}
成功历史限制：{{.SuccessfulJobsHistoryLimit}}
失败历史限制：{{.FailedJobsHistoryLimit}}
最后调度时间：{{.LastScheduleTime}}
活跃Job数量：{{.ActiveJobs}}

请提供：
1. 调度表达式的解释和验证
2. 配置参数的合理性分析
3. 性能和资源优化建议
4. 监控和告警配置建议
5. 故障排查指导

请用中文回答，并提供相关的kubectl命令示例。`,
		IsBuiltin: true,
		IsEnabled: true,
	},
	{
		Name:        "日志分析",
		Description: "用于分析Kubernetes组件日志的AI提示词",
		PromptType:  constants.AIPromptTypeLog,
		Content: `你是一个Kubernetes专家，请分析以下日志信息：

日志来源：{{.Source}}
时间范围：{{.TimeRange}}
日志级别：{{.LogLevel}}
日志内容：
{{.LogContent}}

相关上下文：{{.Context}}

请提供：
1. 日志中关键信息的提取和解释
2. 错误和警告信息的分析
3. 性能指标的评估
4. 潜在问题的识别
5. 具体的解决建议和操作步骤
6. 日志监控和告警配置建议

请用中文回答，重点关注异常模式和性能瓶颈，提供具体的诊断和修复命令。`,
		IsBuiltin: true,
		IsEnabled: true,
	},
}
