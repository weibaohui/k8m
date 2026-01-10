package plugins

// Status 插件状态
type Status int

const (
	// StatusUninstalled 未安装
	StatusUninstalled Status = iota
	// StatusInstalled 已安装未启用
	StatusInstalled
	// StatusEnabled 已启用（配置级别，插件已启用但未运行）
	StatusEnabled
	// StatusRunning 运行中（运行时级别，插件正在运行）
	StatusRunning
	// StatusStopped 已停止（运行时级别，插件已停止但仍然是启用状态）
	StatusStopped
	// StatusDisabled 已禁用（配置级别，插件被禁用）
	StatusDisabled
)

// statusToCN 状态转中文字符串
func statusToCN(s Status) string {
	switch s {
	case StatusUninstalled:
		return "未安装"
	case StatusInstalled:
		return "已安装"
	case StatusEnabled:
		return "已启用"
	case StatusRunning:
		return "运行中"
	case StatusStopped:
		return "已停止"
	case StatusDisabled:
		return "已禁用"
	default:
		return "未知"
	}
}

// statusToString 状态转字符串
func statusToString(s Status) string {
	switch s {
	case StatusUninstalled:
		return "uninstalled"
	case StatusInstalled:
		return "installed"
	case StatusEnabled:
		return "enabled"
	case StatusRunning:
		return "running"
	case StatusStopped:
		return "stopped"
	case StatusDisabled:
		return "disabled"
	default:
		return "unknown"
	}
}

// statusFromString 字符串转状态
func statusFromString(s string) Status {
	switch s {
	case "uninstalled":
		return StatusUninstalled
	case "installed":
		return StatusInstalled
	case "enabled":
		return StatusEnabled
	case "running":
		return StatusRunning
	case "stopped":
		return StatusStopped
	case "disabled":
		return StatusDisabled
	default:
		return StatusUninstalled
	}
}
