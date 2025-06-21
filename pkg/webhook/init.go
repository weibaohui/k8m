package webhook

// RegisterAllSenders 集中注册所有平台 Sender
func RegisterAllSenders() {
	RegisterSender("feishu", &FeishuSender{})
	RegisterSender("default", &DefaultSender{})
	// 未来可在此注册更多 Sender
}

func init() {
	RegisterAllSenders()
}
