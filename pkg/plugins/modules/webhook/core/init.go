package core

// RegisterAllAdapters registers all platform adapters.
func RegisterAllAdapters() {
	RegisterAdapter("feishu", &FeishuAdapter{})
	RegisterAdapter("dingtalk", &DingtalkAdapter{})
	RegisterAdapter("wechat", &WechatAdapter{})
	RegisterAdapter("default", &DefaultAdapter{})
	// Future adapters can be registered here
}

func init() {
	RegisterAllAdapters()
}
