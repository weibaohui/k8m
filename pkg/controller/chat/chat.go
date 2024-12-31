package chat

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/controller/sse"
	"github.com/weibaohui/k8m/pkg/service"
	"k8s.io/klog/v2"
)

func Event(c *gin.Context) {
	chatService := service.ChatService()
	if !chatService.IsEnabled() {
		amis.WriteJsonData(c, gin.H{
			"result": "请先配置开启ChatGPT功能",
		})
		return
	}
	var data struct {
		Note                string `form:"note"`
		Source              string `form:"source"`
		Reason              string `form:"reason"`
		ReportingController string `form:"reportingController"`
		Type                string `form:"type"`
		RegardingKind       string `form:"kind"`
	}
	err := c.ShouldBindQuery(&data)
	if err != nil {
		amis.WriteJsonError(c, err)
	}

	prompt := fmt.Sprintf("请你作为k8s专家，对下面的Event做出分析:\n%s", utils.ToJSON(data))

	stream, err := chatService.GetChatStream(prompt)
	if err != nil {
		klog.V(2).Infof("Error Stream chat request:%v\n\n", err)
		return
	}
	sse.WriteWebSocketChatCompletionStream(c, stream)
}

func Describe(c *gin.Context) {
	chatService := service.ChatService()
	if !chatService.IsEnabled() {
		amis.WriteJsonData(c, gin.H{
			"result": "请先配置开启ChatGPT功能",
		})
		return
	}
	var data struct {
		Describe string `form:"describe"`
		Kind     string `form:"kind"`
		Group    string `form:"group"`
	}

	err := c.ShouldBindQuery(&data)
	if err != nil {
		amis.WriteJsonError(c, err)
	}

	prompt := fmt.Sprintf(
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
		data.Group, data.Kind, data.Describe)

	stream, err := chatService.GetChatStream(prompt)
	if err != nil {
		klog.V(2).Infof("Error Stream chat request:%v\n\n", err)
		return
	}
	sse.WriteWebSocketChatCompletionStream(c, stream)
}
func Example(c *gin.Context) {
	chatService := service.ChatService()
	if !chatService.IsEnabled() {
		amis.WriteJsonData(c, gin.H{
			"result": "请先配置开启ChatGPT功能",
		})
		return
	}
	var data struct {
		Version string `form:"version" `
		Kind    string `form:"kind"`
		Group   string `form:"group"`
	}

	err := c.ShouldBindQuery(&data)
	if err != nil {
		amis.WriteJsonError(c, err)
	}

	prompt := fmt.Sprintf(
		`
		我正在浏览k8s资源管理页面，资源定义Kind=%s,Gropu=%s,version=%s。
		\n请你作为kubernetes k8s 技术专家，给出一份关于这个资源的yaml样例。
		要求先假设一个简单场景、一个复杂场景。1、分别概要介绍这两个场景，2、为这两个场景书写yaml文件，每一行yaml都增加简体中文注释。
		\n注意：
		\n0、使用中文进行回答。
		\n1、你我之间只进行这一轮交互，后面不要再问问题了。
		\n2、请你在给出答案前反思下回答是否逻辑正确，如有问题请先修正，再返回。回答要直接，不要加入上下衔接、开篇语气词、结尾语气词等啰嗦的信息。
		\n3、请不要向我提问，也不要向我确认信息，请不要让我检查markdown格式，不要让我确认markdown格式是否正确`,
		data.Group, data.Kind, data.Version)

	stream, err := chatService.GetChatStream(prompt)
	if err != nil {
		klog.V(2).Infof("Error Stream chat request:%v\n\n", err)
		return
	}
	sse.WriteWebSocketChatCompletionStream(c, stream)
}
func Resource(c *gin.Context) {
	chatService := service.ChatService()
	if !chatService.IsEnabled() {
		amis.WriteJsonData(c, gin.H{
			"result": "请先配置开启ChatGPT功能",
		})
		return
	}
	var data struct {
		Version string `form:"version" `
		Kind    string `form:"kind"`
		Group   string `form:"group"`
	}

	err := c.ShouldBindQuery(&data)
	if err != nil {
		amis.WriteJsonError(c, err)
	}

	prompt := fmt.Sprintf(
		`
		我正在浏览k8s资源管理页面，资源定义Kind=%s,Gropu=%s,version=%s。
		\n请你作为kubernetes k8s 技术专家，给我一份关于这个k8s资源的使用指南。
		要求包括资源说明、使用场景（举例说明）、最佳实践、典型示例（配合前面的场景举例，编写yaml文件，每一行yaml都增加简体中文注释）、关键字段及其含义、常见问题、官方文档链接、引用文档链接等你认为对我有帮助的信息。
		\n注意：
		\n0、使用中文进行回答。
		\n1、你我之间只进行这一轮交互，后面不要再问问题了。
		\n2、请你在给出答案前反思下回答是否逻辑正确，如有问题请先修正，再返回。回答要直接，不要加入上下衔接、开篇语气词、结尾语气词等啰嗦的信息。
		\n3、请不要向我提问，也不要向我确认信息，请不要让我检查markdown格式，不要让我确认markdown格式是否正确`,
		data.Group, data.Kind, data.Version)

	stream, err := chatService.GetChatStream(prompt)
	if err != nil {
		klog.V(2).Infof("Error Stream chat request:%v\n\n", err)
		return
	}
	sse.WriteWebSocketChatCompletionStream(c, stream)
}
func SSEDescribe(c *gin.Context) {

	chatService := service.ChatService()
	if !chatService.IsEnabled() {
		amis.WriteJsonData(c, gin.H{
			"result": "请先配置开启ChatGPT功能",
		})
		return
	}
	txt := c.Query("txt")
	prompt := fmt.Sprintf("请你作为kubernetes k8s 技术专家，回答下面的问题%s\n。请先复述问题，再列出提纲，再写出思路，最后给出样例", txt)

	stream, err := chatService.GetChatStream(prompt)
	if err != nil {
		klog.V(2).Infof("Error Stream chat request:%v\n\n", err)
		return
	}
	sse.WriteWebSocketChatCompletionStream(c, stream)
}

func Cron(c *gin.Context) {
	chatService := service.ChatService()
	if !chatService.IsEnabled() {
		amis.WriteJsonData(c, gin.H{
			"result": "请先配置开启ChatGPT功能",
		})
		return
	}
	var data struct {
		Cron string `form:"cron"`
	}
	err := c.ShouldBindQuery(&data)
	if err != nil {
		amis.WriteJsonError(c, err)
	}

	prompt := fmt.Sprintf(
		`我正在查看k8s cronjob 中的schedule 表达式：%s。
		\n请你作为k8s技术专家，对 %s 这个表达式进行分析，给出详细的解释。
		\n注意：
		\n0、使用中文进行回答。
		\n1、你我之间只进行这一轮交互，后面不要再问问题了。
		\n2、请你在给出答案前反思下回答是否逻辑正确，如有问题请先修正，再返回。回答要直接，不要加入上下衔接、开篇语气词、结尾语气词等啰嗦的信息。
		\n3、请不要向我提问，也不要向我确认信息，请不要让我检查markdown格式，不要让我确认markdown格式是否正确`,
		data.Cron, data.Cron)

	stream, err := chatService.GetChatStream(prompt)
	if err != nil {
		klog.V(2).Infof("Error Stream chat request:%v\n\n", err)
		return
	}
	sse.WriteWebSocketChatCompletionStream(c, stream)
}
func Log(c *gin.Context) {
	chatService := service.ChatService()
	if !chatService.IsEnabled() {
		amis.WriteJsonData(c, gin.H{
			"result": "请先配置开启ChatGPT功能",
		})
		return
	}
	var data struct {
		Data []string `form:"data"`
	}

	err := c.ShouldBindQuery(&data)
	if err != nil {
		amis.WriteJsonError(c, err)
	}

	prompt := fmt.Sprintf("请你作为k8s、Devops、软件工程专家，对下面的Log做出分析:\n%s", utils.ToJSON(data))

	stream, err := chatService.GetChatStream(prompt)
	if err != nil {
		klog.V(2).Infof("Error Stream chat request:%v\n\n", err)
		return
	}
	sse.WriteWebSocketChatCompletionStream(c, stream)
}
