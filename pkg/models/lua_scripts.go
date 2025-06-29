package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/constants"
	"gorm.io/gorm"
)

// InspectionLuaScript 表示一条 Lua 脚本的元数据及内容
// 包含脚本名称、描述、分组、版本、类型和脚本内容等信息
// 用于存储和管理自定义 Lua 脚本
type InspectionLuaScript struct {
	ID          uint                    `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	Name        string                  `json:"name"`                                   // 脚本名称，主键
	Description string                  `json:"description"`                            // 脚本描述
	Group       string                  `json:"group"`                                  // 分组
	Version     string                  `json:"version"`                                // 版本
	Kind        string                  `json:"kind"`                                   // 类型
	ScriptType  constants.LuaScriptType `json:"script_type"`                            // 脚本类型 内置/自定义
	Script      string                  `gorm:"type:text" json:"script"`                // 脚本内容
	ScriptCode  string                  `gorm:"uniqueIndex;size:64" json:"script_code"` // 脚本唯一标识码，每个脚本唯一
	CreatedAt   time.Time               `json:"created_at,omitempty" gorm:"<-:create"`
	UpdatedAt   time.Time               `json:"updated_at,omitempty"` // Automatically managed by GORM for update time

}

// List 返回符合条件的 InspectionLuaScript 列表及总数
func (c *InspectionLuaScript) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*InspectionLuaScript, int64, error) {
	return dao.GenericQuery(params, c, queryFuncs...)
}

// Save 保存或更新 InspectionLuaScript 实例
func (c *InspectionLuaScript) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, c, queryFuncs...)
}

// Delete 根据指定 ID 删除 InspectionLuaScript 实例
func (c *InspectionLuaScript) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, c, utils.ToInt64Slice(ids), queryFuncs...)
}

// GetOne 获取单个 InspectionLuaScript 实例
func (c *InspectionLuaScript) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*InspectionLuaScript, error) {
	return dao.GenericGetOne(params, c, queryFuncs...)
}

// InspectionLuaScriptBuiltinVersion 用于记录内置脚本的版本号
// 只会有一条记录，key 固定为 builtin_lua_scripts
// 用于判断是否需要更新内置脚本
type InspectionLuaScriptBuiltinVersion struct {
	Key       string    `gorm:"primaryKey;size:64" json:"key"` // 固定为 builtin_lua_scripts
	Version   string    `json:"version"`                       // 版本号
	UpdatedAt time.Time `json:"updated_at"`
}

// GetBuiltinLuaScriptsVersion 获取数据库中记录的内置脚本版本
func GetBuiltinLuaScriptsVersion(db *gorm.DB) (string, error) {
	record := &InspectionLuaScriptBuiltinVersion{}
	err := db.First(record, "`key` = ?", "builtin_lua_scripts").Error
	if err != nil {
		return "", err
	}
	return record.Version, nil
}

// SetBuiltinLuaScriptsVersion 设置数据库中内置脚本的版本号
func SetBuiltinLuaScriptsVersion(db *gorm.DB, version string) error {
	record := &InspectionLuaScriptBuiltinVersion{
		Key:       "builtin_lua_scripts",
		Version:   version,
		UpdatedAt: time.Now(),
	}
	return db.Save(record).Error
}
