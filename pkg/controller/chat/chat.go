package chat

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/russross/blackfriday/v2"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/controller/sse"
	"github.com/weibaohui/k8m/pkg/service"
	"k8s.io/klog/v2"
)

func markdownToHTML(md string) string {
	html := blackfriday.Run([]byte(md))
	return string(html)
}
func Chat(c *gin.Context) {
	q := c.Query("q")
	chatService := service.ChatService()
	result := chatService.Chat(q)
	amis.WriteJsonData(c, result)
}
func Event(c *gin.Context) {
	chatService := service.ChatService()
	if !chatService.IsEnabled() {
		amis.WriteJsonData(c, gin.H{
			"result": "请先配置开启ChatGPT功能",
		})
		return
	}
	var event struct {
		Note                string `json:"note"`
		Source              string `json:"source"`
		Reason              string `json:"reason"`
		ReportingController string `json:"reportingController"`
		Type                string `json:"type"`
		RegardingKind       string `json:"kind"`
	}
	err := c.ShouldBindJSON(&event)
	if err != nil {
		amis.WriteJsonError(c, err)
	}

	prompt := fmt.Sprintf("请你作为k8s专家，对下面的Event做出分析:\n%s", utils.ToJSON(event))

	result := chatService.Chat(prompt)
	result = markdownToHTML(result)
	amis.WriteJsonData(c, gin.H{
		"result": result,
	})
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
		Describe string `json:"describe"`
		Kind     string `json:"kind"`
		Group    string `json:"group"`
	}

	err := c.ShouldBindJSON(&data)
	if err != nil {
		amis.WriteJsonError(c, err)
	}

	prompt := fmt.Sprintf(
		`
		我们一起合作了k8m这个k8s AI dashboard,我们要一起面对用户，解决用户的问题。
		用户正在查看关于%s %s 资源的Describe 信息。
		请你作为kubernetes k8s 技术专家，给出分析结论，如果有问题，请指出问题，并给出可能得解决方案。
		\n注意：
		\n1、你我之间只进行这一轮交互，后面不要再问问题了。
		\n2、请你在给出答案前反思下回答是否逻辑正确，如有问题请先修正，再返回。
		\n3、请不要向我提问，也不要向我确认信息，请不要让我检查markdown格式，不要让我确认markdown格式是否正确。
		\n4、我会将你的回答直接返回给用户，所以不要将你我之间的沟通信息暴露给用户。
		\n5、我们是一个整体，共同面对用户，我们不分你我，在用户面前，你就是我，我就是你。我们使用一个身份，为用户解答疑惑。
		\n6、我们对用户称呼为您，涉及对用户的表述，都用‘您’进行替换。
		\n7、在行文上，让用户觉得咱们是一个人，不要让用户看出你、我、用户是三个独立个体。
		\n\nDescribe信息如下：%s`,
		data.Group, data.Kind, data.Describe)

	result := chatService.Chat(prompt)
	amis.WriteJsonData(c, gin.H{
		"result": result,
	})
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
		Version string `json:"version"`
		Kind    string `json:"kind"`
		Group   string `json:"group"`
	}

	err := c.ShouldBindJSON(&data)
	if err != nil {
		amis.WriteJsonError(c, err)
	}

	prompt := fmt.Sprintf(
		`
		我们一起合作了k8m这个k8s AI dashboard,我们要一起面对用户，解决用户的问题。
		当前用户正在访问的k8s资源信息为Kind=%s,Gropu=%s,version=%s。
		\n请你作为kubernetes k8s 技术专家，给用户一份关于这个k8s资源的使用指南。
		要求包括资源说明、使用场景、最佳实践、典型示例、常见问题等你认为对用户有帮助的信息。
		\n注意：
		\n1、你我之间只进行这一轮交互，后面不要再问问题了。
		\n2、请你在给出答案前反思下回答是否逻辑正确，如有问题请先修正，再返回。回答要直接，不要加入上下衔接、开篇语气词、结尾语气词等啰嗦的信息。
		\n3、请不要向我提问，也不要向我确认信息，请不要让我检查markdown格式，不要让我确认markdown格式是否正确
		\n4、我会将你的回答直接返回给用户，所以不要将你我之间的沟通信息暴露给用户
		\n5、我们是一个整体，共同面对用户，我们不分你我，在用户面前，你就是我，我就是你。我们使用一个身份，为用户解答疑惑。
		\n6、我们对用户称呼为您，涉及对用户的表述，都用‘您’进行替换。
		\n7、在行文上，让用户觉得咱们是一个人，不要让用户看出你、我、用户是三个独立个体。`,
		data.Group, data.Kind, data.Version)

	result := chatService.Chat(prompt)
	amis.WriteJsonData(c, gin.H{
		"result": result,
	})
}
func SSEDescribe(c *gin.Context) {

	chatService := service.ChatService()
	if !chatService.IsEnabled() {
		amis.WriteJsonData(c, gin.H{
			"result": "请先配置开启ChatGPT功能",
		})
		return
	}
	var data struct {
		Describe string `json:"describe"`
		Kind     string `json:"kind"`
		Group    string `json:"group"`
	}

	err := c.ShouldBindJSON(&data)
	if err != nil {
		amis.WriteJsonError(c, err)
	}

	prompt := fmt.Sprintf("请你作为kubernetes k8s 运维专家，对下面 %s %s 资源的Describe 信息 分析。请给出分析结论，如果有问题，请指出问题，并给出可能得解决方案:\n%s\n。格式要求：请使用文本格式，不要使用markdown格式。请保留换行符等保证基本的格式", data.Group, data.Kind, data.Describe)

	stream, err := chatService.GetChatStream(prompt)
	if err != nil {
		klog.V(2).Infof("Error Ssemaking request:%v\n\n", err)
		return
	}
	sse.WriteSSEChatCompletionStream(c, stream)
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
		Cron string `json:"cron"`
	}

	err := c.ShouldBindJSON(&data)
	if err != nil {
		amis.WriteJsonError(c, err)
	}

	prompt := fmt.Sprintf(
		`我们一起合作了k8m这个k8s AI dashboard,我们要一起面对用户，解决用户的问题。
		\n当前用户正在浏览k8s cronjob 中的schedule 表达式：%s。
		\n请你作为k8s技术专家，对 %s 这个表达式进行分析。
		要求包括表达式说明、使用场景、常见问题等你认为对用户有帮助的信息。
		\n注意：
		\n1、你我之间只进行这一轮交互，后面不要再问问题了。
		\n2、请你在给出答案前反思下回答是否逻辑正确，如有问题请先修正，再返回。回答要直接，不要加入上下衔接、开篇语气词、结尾语气词等啰嗦的信息。
		\n3、请不要向我提问，也不要向我确认信息，请不要让我检查markdown格式，不要让我确认markdown格式是否正确
		\n4、我会将你的回答直接返回给用户，所以不要将你我之间的沟通信息暴露给用户
		\n5、我们是一个整体，共同面对用户，我们不分你我，在用户面前，你就是我，我就是你。我们使用一个身份，为用户解答疑惑
		\n6、我们对用户称呼为您，涉及对用户的表述，都用‘您’进行替换，
		\n7、在行文上，让用户觉得咱们是一个人，不要让用户看出你、我、用户是三个独立个体。`,
		data.Cron, data.Cron)

	result := chatService.Chat(prompt)
	result = markdownToHTML(result)
	amis.WriteJsonData(c, gin.H{
		"result": result,
	})
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
		Data []string `json:"data"`
	}

	err := c.ShouldBindJSON(&data)
	if err != nil {
		amis.WriteJsonError(c, err)
	}

	prompt := fmt.Sprintf("请你作为k8s、Devops、软件工程专家，对下面的Log做出分析:\n%s", utils.ToJSON(data))

	result := chatService.Chat(prompt)
	result = markdownToHTML(result)
	amis.WriteJsonData(c, gin.H{
		"result": result,
	})
}
