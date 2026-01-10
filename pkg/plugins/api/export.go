package api

func InitNoopService() {
	initAINoop()
	initWebhookNoop()
}

// AIChatService 返回当前生效的 AIChat 实现，始终非 nil。
func AIChatService() AIChat {
	return aiChatVal.Load().(*aiChatHolder).chat
}

// AIConfigService 返回当前生效的 AIConfig 实现，始终非 nil。
func AIConfigService() AIConfig {
	return aiConfigVal.Load().(*aiConfigHolder).cfg
}

// WebhookService 中文函数注释：返回当前生效的 Webhook 实现，始终非 nil。
func WebhookService() Webhook {
	return webhookVal.Load().(*webhookHolder).svc
}
