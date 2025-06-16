package webhooksender

func init() {
	RegisterSender("feishu", &FeishuSender{})
}
