package webhook

// RegisterAllAdapters registers all platform adapters.
func RegisterAllAdapters() {
	RegisterAdapter("feishu", &FeishuAdapter{})
	RegisterAdapter("dingtalk", &DingtalkAdapter{})
	RegisterAdapter("wechat", &WechatAdapter{})
	RegisterAdapter("default", &DefaultAdapter{})
	// Future adapters can be registered here
}

// RegisterAllSenders registers all legacy senders for backward compatibility.
// Deprecated: Use RegisterAllAdapters instead.
// Note: Legacy sender implementations have been removed. This function is kept for API compatibility.
func RegisterAllSenders() {
	// Legacy sender implementations have been removed in favor of the new adapter architecture.
	// This function is kept empty to maintain backward compatibility.
	// Use RegisterAllAdapters() and the new WebhookClient instead.
}

func init() {
	RegisterAllAdapters()
	RegisterAllSenders() // Keep for backward compatibility
}
