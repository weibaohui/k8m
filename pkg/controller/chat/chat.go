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

	prompt := fmt.Sprintf("请你作为kubernetes k8s 运维专家，对下面 %s %s 资源的Describe 信息 分析。请给出分析结论，如果有问题，请指出问题，并给出可能得解决方案:\n%s", data.Group, data.Kind, data.Describe)

	result := chatService.Chat(prompt)
	result = markdownToHTML(result)
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

	prompt := fmt.Sprintf("请你作为Cron表达式专家，对下面的Cron表达式进行分析:\n%s", data.Cron)

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
