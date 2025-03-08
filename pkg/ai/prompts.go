package ai

const (
	default_prompt = `简化以下由三个破折号分隔的Kubernetes错误信息，
	错误内容：--- %s ---。
	请以分步形式提供最可能的解决方案，字符数不超过280。
	输出格式：
	错误信息: {此处解释错误}
	解决方案: {此处分步说明解决方案}
	`

	prom_conf_prompt = `简化以下由三个破折号分隔的Prometheus错误信息，
	错误内容：--- %s ---。
	该错误发生在验证Prometheus配置文件时。
	请参考Prometheus文档（如适用），提供分步修复指导及建议。
	输出格式如下（字符数不超过300）：
	错误信息: {此处解释错误}
	解决方案: {此处分步说明解决方案}
	`

	prom_relabel_prompt = `
	请使用以下语言生成响应：%s
	配置列表形式如下：
	job_name:
	{Prometheus job_name}
	relabel_configs:
	{Prometheus重标记配置}
	kubernetes_sd_configs:
	{Prometheus服务发现配置}
	---
	%s
	---
	请为每个job_name描述其匹配的Kubernetes服务和Pod标签、
	命名空间、端口及容器。
	返回消息：
	已发现并解析Prometheus抓取配置。
	确保被监控目标运行时有以下至少一个标签集：
	按以下格式为每个job生成报告：
	- 任务: {job_name}
	  - 服务标签:
	    - {服务标签列表}
	  - Pod标签:
	    - {Pod标签列表}
	  - 命名空间:
	    - {命名空间列表}
	  - 端口:
	    - {端口列表}
	  - 容器:
	    - {容器名称列表}
	`
	kyverno_prompt = `简化以下由三个破折号分隔的Kyverno告警信息，
	告警内容：--- %s ---。
	请提供最可能的kubectl命令解决方案。
	输出格式要求（仅显示解决方案命令）：
	错误信息: {此处解释错误}
	解决方案: {kubectl command}
	`
)

var PromptMap = map[string]string{
	"default":                       default_prompt,
	"PrometheusConfigValidate":      prom_conf_prompt,
	"PrometheusConfigRelabelReport": prom_relabel_prompt,
	"PolicyReport":                  kyverno_prompt,
	"ClusterPolicyReport":           kyverno_prompt,
}
