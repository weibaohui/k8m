package chat

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/htpl"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/controller/sse"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
	"k8s.io/klog/v2"
)

type Controller struct {
}

func RegisterChatRoutes(ai *gin.RouterGroup) {
	ctrl := &Controller{}
	ai.GET("/chat/event", ctrl.Event)
	ai.GET("/chat/log", ctrl.Log)
	ai.GET("/chat/cron", ctrl.Cron)
	ai.GET("/chat/describe", ctrl.Describe)
	ai.GET("/chat/resource", ctrl.Resource)
	ai.GET("/chat/any_question", ctrl.AnyQuestion)
	ai.GET("/chat/any_selection", ctrl.AnySelection)
	ai.GET("/chat/example", ctrl.Example)
	ai.GET("/chat/example/field", ctrl.FieldExample)
	ai.GET("/chat/ws_chatgpt", ctrl.GPTShell)
	ai.GET("/chat/ws_chatgpt/history", ctrl.History)
	ai.GET("/chat/ws_chatgpt/history/reset", ctrl.Reset)
	ai.GET("/chat/k8s_gpt/resource", ctrl.K8sGPTResource)
}

type ResourceData struct {
	// 资源版本
	Version string `form:"version"`
	// 资源类型
	Kind string `form:"kind"`
	// 资源组
	Group string `form:"group"`
	// 资源描述
	Describe string `form:"describe"`
	// 定时任务
	Cron string `form:"cron"`
	// 日志
	Data      string `form:"data"`
	Field     string `form:"field"`
	Name      string `form:"name"`
	Namespace string `form:"namespace"`
	// 事件
	Note                string `form:"note"`
	Source              string `form:"source"`
	Reason              string `form:"reason"`
	ReportingController string `form:"reportingController"`
	Type                string `form:"type"`
	RegardingKind       string `form:"regardingKind"`
	// AnyQuestion 任意提问
	Question string `form:"question"`
}

func handleRequest(c *gin.Context, promptFunc func(data any) string) {
	if !service.AIService().IsEnabled() {
		amis.WriteJsonData(c, gin.H{
			"result": "请先配置开启ChatGPT功能",
		})
		return
	}

	var data ResourceData
	err := c.ShouldBindQuery(&data)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	ctxInst := amis.GetContextWithUser(c)

	prompt := promptFunc(data)

	stream, err := service.ChatService().GetChatStreamWithoutHistory(ctxInst, prompt)
	if err != nil {
		klog.V(2).Infof("Error Stream chat request:%v\n\n", err)
		return
	}
	sse.WriteWebSocketChatCompletionStream(c, stream)
}

// renderTemplate 通用的模板处理函数
// templateStr: 模板字符串
// contextBuilder: 根据ResourceData构建上下文的函数
func renderTemplate(templateStr string, data any, contextBuilder func(ResourceData) map[string]any) string {
	d, ok := data.(ResourceData)
	if !ok {
		klog.V(6).Infof("Error: data is not ResourceData type")
		return ""
	}
	eng := htpl.NewEngine()
	// 解析模板
	tpl, err := eng.ParseString(templateStr)
	if err != nil {
		klog.V(6).Infof("Error Parse template:%v\n\n", err)
		return ""
	}

	ctx := contextBuilder(d)

	// 渲染模板
	result, err := tpl.Render(ctx)
	if err != nil {
		klog.V(6).Infof("Error Render template:%v\n\n", err)
		return ""
	}
	return result
}

// @Summary 分析K8s事件
// @Security BearerAuth
// @Param note query string false "事件备注"
// @Param source query string false "事件来源"
// @Param reason query string false "事件原因"
// @Param type query string false "事件类型"
// @Param regardingKind query string false "相关资源类型"
// @Success 200 {object} string
// @Router /ai/chat/event [get]
func (cc *Controller) Event(c *gin.Context) {

	handleRequest(c, func(data any) string {
		// 从数据库获取prompt模板
		templateStr := getPromptWithFallback(c.Request.Context(), constants.AIPromptTypeEvent)

		return renderTemplate(templateStr, data, func(d ResourceData) map[string]any {
			return map[string]any{
				"Note":          d.Note,
				"Source":        d.Source,
				"Reason":        d.Reason,
				"Type":          d.Type,
				"RegardingKind": d.RegardingKind,
			}
		})
	})
}

// @Summary 分析K8s资源描述
// @Security BearerAuth
// @Param group query string false "资源组"
// @Param version query string false "资源版本"
// @Param kind query string false "资源类型"
// @Param name query string false "资源名称"
// @Param namespace query string false "命名空间"
// @Success 200 {object} string
// @Router /ai/chat/describe [get]
func (cc *Controller) Describe(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	var data ResourceData
	err := c.ShouldBindQuery(&data)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	cluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	var describe []byte
	kom.Cluster(cluster).WithContext(ctx).GVK(data.Group, data.Version, data.Kind).
		Name(data.Name).
		Namespace(data.Namespace).
		Describe(&describe)

	handleRequest(c, func(data any) string {
		// 从数据库获取prompt模板
		templateStr := getPromptWithFallback(c.Request.Context(), constants.AIPromptTypeDescribe)

		return renderTemplate(templateStr, data, func(d ResourceData) map[string]any {
			return map[string]any{
				"Group":        d.Group,
				"Kind":         d.Kind,
				"DescribeInfo": string(describe),
			}
		})
	})
}

// @Summary 获取K8s资源使用示例
// @Security BearerAuth
// @Param group query string false "资源组"
// @Param version query string false "资源版本"
// @Param kind query string false "资源类型"
// @Success 200 {object} string
// @Router /ai/chat/example [get]
func (cc *Controller) Example(c *gin.Context) {
	handleRequest(c, func(data any) string {
		// 从数据库获取prompt模板
		templateStr := getPromptWithFallback(c.Request.Context(), constants.AIPromptTypeExample)

		return renderTemplate(templateStr, data, func(d ResourceData) map[string]any {
			return map[string]any{
				"Kind":    d.Kind,
				"Group":   d.Group,
				"Version": d.Version,
			}
		})
	})
}

// @Summary 获取K8s资源字段示例
// @Security BearerAuth
// @Param group query string false "资源组"
// @Param version query string false "资源版本"
// @Param kind query string false "资源类型"
// @Param field query string false "字段名称"
// @Success 200 {object} string
// @Router /ai/chat/example/field [get]
func (cc *Controller) FieldExample(c *gin.Context) {
	handleRequest(c, func(data any) string {
		// 从数据库获取prompt模板
		templateStr := getPromptWithFallback(c.Request.Context(), constants.AIPromptTypeFieldExample)

		return renderTemplate(templateStr, data, func(d ResourceData) map[string]any {
			return map[string]any{
				"Kind":    d.Kind,
				"Group":   d.Group,
				"Version": d.Version,
				"Field":   d.Field,
			}
		})
	})
}

// @Summary 获取K8s资源使用指南
// @Security BearerAuth
// @Param group query string false "资源组"
// @Param version query string false "资源版本"
// @Param kind query string false "资源类型"
// @Success 200 {object} string
// @Router /ai/chat/resource [get]
func (cc *Controller) Resource(c *gin.Context) {
	handleRequest(c, func(data any) string {
		// 从数据库获取prompt模板
		templateStr := getPromptWithFallback(c.Request.Context(), constants.AIPromptTypeResource)

		return renderTemplate(templateStr, data, func(d ResourceData) map[string]any {
			return map[string]any{
				"Kind":    d.Kind,
				"Group":   d.Group,
				"Version": d.Version,
			}
		})
	})
}

// @Summary K8s错误信息分析
// @Security BearerAuth
// @Param data query string false "错误内容"
// @Param name query string false "资源名称"
// @Param kind query string false "资源类型"
// @Param field query string false "相关字段"
// @Success 200 {object} string
// @Router /ai/chat/k8s_gpt/resource [get]
func (cc *Controller) K8sGPTResource(c *gin.Context) {
	handleRequest(c, func(data any) string {
		// 从数据库获取prompt模板
		templateStr := getPromptWithFallback(c.Request.Context(), constants.AIPromptTypeK8sGPTResource)

		return renderTemplate(templateStr, data, func(d ResourceData) map[string]any {
			return map[string]any{
				"Data":  d.Data,
				"Name":  d.Name,
				"Kind":  d.Kind,
				"Field": d.Field,
			}
		})
	})
}

// @Summary 解释选择内容
// @Security BearerAuth
// @Param question query string false "要解释的内容"
// @Success 200 {object} string
// @Router /ai/chat/any_selection [get]
func (cc *Controller) AnySelection(c *gin.Context) {
	handleRequest(c, func(data any) string {
		// 从数据库获取prompt模板
		templateStr := getPromptWithFallback(c.Request.Context(), constants.AIPromptTypeAnySelection)

		return renderTemplate(templateStr, data, func(d ResourceData) map[string]any {
			return map[string]any{
				"Question": d.Question,
			}
		})
	})
}

// @Summary 回答K8s相关问题
// @Security BearerAuth
// @Param group query string false "资源组"
// @Param version query string false "资源版本"
// @Param kind query string false "资源类型"
// @Param question query string false "问题内容"
// @Success 200 {object} string
// @Router /ai/chat/any_question [get]
func (cc *Controller) AnyQuestion(c *gin.Context) {
	handleRequest(c, func(data any) string {
		// 从数据库获取prompt模板
		templateStr := getPromptWithFallback(c.Request.Context(), constants.AIPromptTypeAnyQuestion)

		return renderTemplate(templateStr, data, func(d ResourceData) map[string]any {
			return map[string]any{
				"Kind":     d.Kind,
				"Group":    d.Group,
				"Version":  d.Version,
				"Question": d.Question,
			}
		})
	})
}

// @Summary 分析Cron表达式
// @Security BearerAuth
// @Param cron query string false "Cron表达式"
// @Success 200 {object} string
// @Router /ai/chat/cron [get]
func (cc *Controller) Cron(c *gin.Context) {
	handleRequest(c, func(data any) string {
		// 从数据库获取prompt模板
		templateStr := getPromptWithFallback(c.Request.Context(), constants.AIPromptTypeCron)

		return renderTemplate(templateStr, data, func(d ResourceData) map[string]any {
			return map[string]any{
				"Cron": d.Cron,
			}
		})
	})
}

// @Summary 分析日志
// @Security BearerAuth
// @Param data query string false "日志内容"
// @Success 200 {object} string
// @Router /ai/chat/log [get]
func (cc *Controller) Log(c *gin.Context) {
	handleRequest(c, func(data any) string {
		// 从数据库获取prompt模板
		templateStr := getPromptWithFallback(c.Request.Context(), constants.AIPromptTypeLog)

		return renderTemplate(templateStr, data, func(d ResourceData) map[string]any {
			return map[string]any{
				"Data": utils.ToJSON(d.Data),
			}
		})
	})
}

// getPromptWithFallback 根据提示类型从数据库获取模板，若失败则回退到内置模板。
// 参数说明：
// - ctx: 请求上下文，用于数据库或服务调用的上下文传递。
// - promptType: 提示词类型常量（如 constants.AIPromptTypeEvent 等）。
// 返回值：
// - 模板字符串，如果数据库查询失败则返回内置模板内容。
func getPromptWithFallback(ctx context.Context, promptType constants.AIPromptType) string {
	templateStr, err := service.PromptService().GetPrompt(ctx, promptType)
	if err != nil {
		klog.Errorf("获取%s prompt模板失败: %v", promptType, err)
		// 如果获取失败，使用内置模板
		templateStr = models.GetBuiltinPromptContent(promptType)
	}
	return templateStr
}
