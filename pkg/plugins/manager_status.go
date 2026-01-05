package plugins

// Status 插件状态
type Status int

const (
	// StatusDiscovered 已发现
	StatusDiscovered Status = iota
	// StatusInstalled 已安装未启用
	StatusInstalled
	// StatusEnabled 已启用
	StatusEnabled
	// StatusDisabled 已禁用
	StatusDisabled
)

// statusToCN 状态转中文字符串
func statusToCN(s Status) string {
	switch s {
	case StatusDiscovered:
		return "已发现"
	case StatusInstalled:
		return "已安装"
	case StatusEnabled:
		return "已启用"
	case StatusDisabled:
		return "已禁用"
	default:
		return "未知"
	}
}

// statusToString 状态转字符串
func statusToString(s Status) string {
	switch s {
	case StatusDiscovered:
		return "discovered"
	case StatusInstalled:
		return "installed"
	case StatusEnabled:
		return "enabled"
	case StatusDisabled:
		return "disabled"
	default:
		return "unknown"
	}
}

// statusFromString 字符串转状态
func statusFromString(s string) Status {
	switch s {
	case "discovered":
		return StatusDiscovered
	case "installed":
		return StatusInstalled
	case "enabled":
		return StatusEnabled
	case "disabled":
		return StatusDisabled
	default:
		return StatusDiscovered
	}
}
