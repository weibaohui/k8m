package chat

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/htpl"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/controller/sse"
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
		eng := htpl.NewEngine()
		// 解析模板
		tpl, err := eng.ParseString(`请你作为k8s专家，对下面的Event做出分析:\n
				note:   ${Note},
				source: ${Source},
				reason: ${Reason},
				type:   ${Type},
				kind:   ${RegardingKind},
		\n`)
		if err != nil {
			klog.V(6).Infof("Error Parse template:%v\n\n", err)
			return ""
		}
		d := data.(ResourceData)
		ctx := map[string]any{
			"Note":          d.Note,
			"Source":        d.Source,
			"Reason":        d.Reason,
			"Type":          d.Type,
			"RegardingKind": d.RegardingKind,
		}
		// 渲染模板
		result, err := tpl.Render(ctx)
		if err != nil {
			klog.V(6).Infof("Error Render template:%v\n\n", err)
			return ""
		}
		klog.V(4).Infof("Render template:%s\n\n", result)
		klog.V(4).Infof("Render template:%s\n\n", result)
		klog.V(4).Infof("Render template:%s\n\n", result)
		klog.V(4).Infof("Render template:%s\n\n", result)
		klog.V(4).Infof("Render template:%s\n\n", result)
		return fmt.Sprintf(result, utils.ToJSON(gin.H{
			"note":   d.Note,
			"source": d.Source,
			"reason": d.Reason,
			"type":   d.Type,
			"kind":   d.RegardingKind,
		}))
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
		d := data.(ResourceData)
		return fmt.Sprintf(
			`
		我正在查看关于k8s %s %s 资源的Describe (kubectl describe )信息。
		请你作为kubernetes k8s 技术专家，对这个describe的文本进行分析。
		\n 请给出分析结论，如果有问题，请指出问题，并给出可能得解决方案。
		\n注意：
		\n0、使用中文进行回答。
		\n1、你我之间只进行这一轮交互，后面不要再问问题了。
		\n2、请你在给出答案前反思下回答是否逻辑正确，如有问题请先修正，再返回。回答要直接，不要加入上下衔接、开篇语气词、结尾语气词等啰嗦的信息。
		\n3、请不要向我提问，也不要向我确认信息，请不要让我检查markdown格式，不要让我确认markdown格式是否正确。
		\n\nDescribe信息如下：%s`,
			d.Group, d.Kind, string(describe))
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
		d := data.(ResourceData)
		return fmt.Sprintf(
			`
		我正在浏览k8s资源管理页面，资源定义Kind=%s,Gropu=%s,version=%s。
		\n请你作为kubernetes k8s 技术专家，给我一份关于这个k8s资源的使用指南。
		\n要求包括资源说明、使用场景（举例说明）、最佳实践、典型示例（配合前面的场景举例，编写yaml文件，每一行yaml都增加简体中文注释）、关键字段及其含义、常见问题、官方文档链接、引用文档链接等你认为对我有帮助的信息。
		\n最后给出一份关于这个资源的yaml样例。
		\n要求先假设一个简单场景、一个复杂场景。1、分别概要介绍这两个场景，2、为这两个场景书写yaml文件，每一行yaml都增加简体中文注释。
		\n注意：
		\n0、使用中文进行回答。
		\n1、你我之间只进行这一轮交互，后面不要再问问题了。
		\n2、请你在给出答案前反思下回答是否逻辑正确，如有问题请先修正，再返回。回答要直接，不要加入上下衔接、开篇语气词、结尾语气词等啰嗦的信息。
		\n3、请不要向我提问，也不要向我确认信息，请不要让我检查markdown格式，不要让我确认markdown格式是否正确`,
			d.Group, d.Kind, d.Version)
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
		d := data.(ResourceData)
		return fmt.Sprintf(
			`
		我正在浏览k8s资源管理页面，资源定义Kind=%s,Gropu=%s,version=%s。
		\n请你作为kubernetes k8s 技术专家，给出一份关于  %s  这个具体字段的使用场景。请在回答中使用 “该字段” 代替这个具体的字段。
		请详细解释该字段的含义、用法、并给出一个假设的使用场景，为这个场景书写yaml文件，每一行yaml都增加简体中文注释。
		\n注意：
		\n0、使用中文进行回答。
		\n1、你我之间只进行这一轮交互，后面不要再问问题了。
		\n2、请你在给出答案前反思下回答是否逻辑正确，如有问题请先修正，再返回。回答要直接，不要加入上下衔接、开篇语气词、结尾语气词等啰嗦的信息。
		\n3、请不要向我提问，也不要向我确认信息，请不要让我检查markdown格式，不要让我确认markdown格式是否正确`,
			d.Group, d.Kind, d.Version, d.Field)
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
		d := data.(ResourceData)
		return fmt.Sprintf(
			`
		我正在浏览k8s资源管理页面，资源定义Kind=%s,Gropu=%s,version=%s。
		\n请你作为kubernetes k8s 技术专家，给我一份关于这个k8s资源的使用指南。
		要求包括资源说明、使用场景（举例说明）、最佳实践、典型示例（配合前面的场景举例，编写yaml文件，每一行yaml都增加简体中文注释）、关键字段及其含义、常见问题、官方文档链接、引用文档链接等你认为对我有帮助的信息。
		\n注意：
		\n0、使用中文进行回答。
		\n1、你我之间只进行这一轮交互，后面不要再问问题了。
		\n2、请你在给出答案前反思下回答是否逻辑正确，如有问题请先修正，再返回。回答要直接，不要加入上下衔接、开篇语气词、结尾语气词等啰嗦的信息。
		\n3、请不要向我提问，也不要向我确认信息，请不要让我检查markdown格式，不要让我确认markdown格式是否正确`,
			d.Group, d.Kind, d.Version)
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
		d := data.(ResourceData)
		return fmt.Sprintf(
			`
			简化以下由三个破折号分隔的Kubernetes错误信息，
	错误内容：--- %s ---。
	资源名称：--- %s ---。
	资源类型：--- %s ---。
	相关字段k8s官方文档解释：--- %s ---。
	请以分步形式提供最可能的解决方案，字符数不超过280。
	输出格式：
	错误信息: {此处解释错误}
	解决方案: {此处分步说明解决方案}
		\n注意：
		\n0、使用中文进行回答。
		\n1、你我之间只进行这一轮交互，后面不要再问问题了。
		\n2、请你在给出答案前反思下回答是否逻辑正确，如有问题请先修正，再返回。回答要直接，不要加入上下衔接、开篇语气词、结尾语气词等啰嗦的信息。
		\n3、请不要向我提问，也不要向我确认信息，请不要让我检查markdown格式，不要让我确认markdown格式是否正确`,
			d.Data, d.Name, d.Kind, d.Field)
	})
}

// @Summary 解释选择内容
// @Security BearerAuth
// @Param question query string false "要解释的内容"
// @Success 200 {object} string
// @Router /ai/chat/any_selection [get]
func (cc *Controller) AnySelection(c *gin.Context) {
	prompt, err := service.PromptService().GetPrompt(c.Request.Context(), constants.AIPromptTypeAnySelection)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	klog.V(4).Infof("prompt: %s", prompt)
	klog.V(4).Infof("prompt: %s", prompt)
	klog.V(4).Infof("prompt: %s", prompt)
	handleRequest(c, func(data any) string {
		d := data.(ResourceData)
		return fmt.Sprintf(prompt, d.Question)
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
		d := data.(ResourceData)
		return fmt.Sprintf(
			`
		我正在浏览k8s资源管理页面，资源定义Kind=%s,Gropu=%s,version=%s。
		\n请你作为kubernetes k8s 技术专家，请你详细解释下我的疑问： %s 。
		要求包括关键名词解释、作用、典型示例（以场景举例，编写yaml文件，每一行yaml都增加简体中文注释）、关键字段及其含义、常见问题、官方文档链接、引用文档链接等你认为对我有帮助的信息。
		\n注意：
		\n0、使用中文进行回答。
		\n1、你我之间只进行这一轮交互，后面不要再问问题了。
		\n2、请你在给出答案前反思下回答是否逻辑正确，如有问题请先修正，再返回。回答要直接，不要加入上下衔接、开篇语气词、结尾语气词等啰嗦的信息。
		\n3、请不要向我提问，也不要向我确认信息，请不要让我检查markdown格式，不要让我确认markdown格式是否正确`,
			d.Group, d.Kind, d.Version, d.Question)
	})
}

// @Summary 分析Cron表达式
// @Security BearerAuth
// @Param cron query string false "Cron表达式"
// @Success 200 {object} string
// @Router /ai/chat/cron [get]
func (cc *Controller) Cron(c *gin.Context) {
	handleRequest(c, func(data any) string {
		d := data.(ResourceData)
		return fmt.Sprintf(
			`我正在查看k8s cronjob 中的schedule 表达式：%s。
		\n请你作为k8s技术专家，对 %s 这个表达式进行分析，给出详细的解释。
		\n注意：
		\n0、使用中文进行回答。
		\n1、你我之间只进行这一轮交互，后面不要再问问题了。
		\n2、请你在给出答案前反思下回答是否逻辑正确，如有问题请先修正，再返回。回答要直接，不要加入上下衔接、开篇语气词、结尾语气词等啰嗦的信息。
		\n3、请不要向我提问，也不要向我确认信息，请不要让我检查markdown格式，不要让我确认markdown格式是否正确`,
			d.Cron, d.Cron)
	})
}

// @Summary 分析日志
// @Security BearerAuth
// @Param data query string false "日志内容"
// @Success 200 {object} string
// @Router /ai/chat/log [get]
func (cc *Controller) Log(c *gin.Context) {
	handleRequest(c, func(data any) string {
		d := data.(ResourceData)
		return fmt.Sprintf("请你作为k8s、Devops、软件工程专家，对下面的Log做出分析:\n%s", utils.ToJSON(d.Data))
	})
}
