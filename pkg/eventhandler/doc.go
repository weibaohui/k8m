// Package eventhandler 提供Kubernetes事件处理功能
//
// 该包实现了灵活的Kubernetes事件处理模块，支持：
// - 事件监听和过滤
// - 规则匹配（支持反向选择）
// - 异步事件处理
// - Webhook推送
// - 配置热更新
// - 多数据库支持
//
// 使用示例：
//
//	// 加载配置（使用flag.Init()）
//	cfg := config.LoadConfigFromFlags()
//	
//	// 创建存储层（使用GORM和dao.DB()）
//	eventStore, err := store.NewStore()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	
//	// 创建Watcher和Worker
//	eventWatcher := watcher.NewEventWatcher(client, eventStore, cfg)
//	eventWorker := worker.NewEventWorker(eventStore, cfg)
//	
//	// 启动服务
//	if err := eventWatcher.Start(); err != nil {
//	    log.Fatal(err)
//	}
//	if err := eventWorker.Start(); err != nil {
//	    log.Fatal(err)
//	}
package eventhandler