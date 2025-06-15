package constants

// LuaScriptType 脚本类型
type LuaScriptType string

const (
	LuaScriptTypeBuiltin LuaScriptType = "Builtin" // 内置
	LuaScriptTypeCustom  LuaScriptType = "Custom"  // 自定义
)

type LuaEventStatus string

const (
	LuaEventStatusNormal LuaEventStatus = "正常" // 正常
	LuaEventStatusFailed LuaEventStatus = "失败" // 失败
)
