package constants

// LuaScriptType 脚本类型
type LuaScriptType string

const (
	LuaScriptTypeBuiltin LuaScriptType = "Builtin" // 内置
	LuaScriptTypeCustom  LuaScriptType = "Custom"  // 自定义
)
