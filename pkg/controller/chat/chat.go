package chat

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/controller/sse"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
	"k8s.io/klog/v2"
)

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

func handleRequest(c *gin.Context, promptFunc func(data interface{}) string) {
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
func Event(c *gin.Context) {
	handleRequest(c, func(data interface{}) string {
		d := data.(ResourceData)
		return fmt.Sprintf("请你作为k8s专家，对下面的Event做出分析:\n%s", utils.ToJSON(gin.H{
			"note":   d.Note,
			"source": d.Source,
			"reason": d.Reason,
			"type":   d.Type,
			"kind":   d.RegardingKind,
		}))
	})
}

// Describe TODO 改为不要传Describe内容，比较大，传个名称过来，从后台Describe一下即可
func Describe(c *gin.Context) {
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

	handleRequest(c, func(data interface{}) string {
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

func Example(c *gin.Context) {
	handleRequest(c, func(data interface{}) string {
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
func FieldExample(c *gin.Context) {
	handleRequest(c, func(data interface{}) string {
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
func Resource(c *gin.Context) {
	handleRequest(c, func(data interface{}) string {
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
func K8sGPTResource(c *gin.Context) {
	handleRequest(c, func(data interface{}) string {
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
func AnySelection(c *gin.Context) {
	handleRequest(c, func(data interface{}) string {
		d := data.(ResourceData)
		return fmt.Sprintf(
			`
		\n请你作为kubernetes k8s 技术专家，请你详细解释下面的文字： %s 。
		\n注意：
		\n0、使用中文进行回答。
		\n1、你我之间只进行这一轮交互，后面不要再问问题了。
		\n2、请你在给出答案前反思下回答是否逻辑正确，如有问题请先修正，再返回。回答要直接，不要加入上下衔接、开篇语气词、结尾语气词等啰嗦的信息。
		\n3、请不要向我提问，也不要向我确认信息，请不要让我检查markdown格式，不要让我确认markdown格式是否正确`,
			d.Question)
	})
}
func AnyQuestion(c *gin.Context) {
	handleRequest(c, func(data interface{}) string {
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

func Cron(c *gin.Context) {
	handleRequest(c, func(data interface{}) string {
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

func Log(c *gin.Context) {
	handleRequest(c, func(data interface{}) string {
		d := data.(ResourceData)
		return fmt.Sprintf("请你作为k8s、Devops、软件工程专家，对下面的Log做出分析:\n%s", utils.ToJSON(d.Data))
	})
}
