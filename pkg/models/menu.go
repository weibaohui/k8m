package models

import (
	"encoding/json"
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

// Menu represents a versioned menu structure.
// Each save operation creates a new record with an incremented version.
type Menu struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	MenuData  string    `gorm:"type:text" json:"-"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// MenuDataJSON 用于处理JSON序列化和反序列化
type MenuDataJSON struct {
	ID        uint        `json:"id,omitempty"`
	MenuData  interface{} `json:"menu_data,omitempty"`
	CreatedAt time.Time   `json:"created_at,omitempty"`
	UpdatedAt time.Time   `json:"updated_at,omitempty"`
}

// MarshalJSON 自定义JSON序列化方法
// 将存储在数据库中的字符串转换为JSON对象
func (m *Menu) MarshalJSON() ([]byte, error) {
	var menuData interface{}
	if m.MenuData != "" {
		if err := json.Unmarshal([]byte(m.MenuData), &menuData); err != nil {
			// 如果解析失败，直接返回字符串
			menuData = m.MenuData
		}
	}
	
	menuJSON := MenuDataJSON{
		ID:        m.ID,
		MenuData:  menuData,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
	
	return json.Marshal(menuJSON)
}

// UnmarshalJSON 自定义JSON反序列化方法
// 将接收到的JSON数据转换为字符串存储
func (m *Menu) UnmarshalJSON(data []byte) error {
	var menuJSON MenuDataJSON
	if err := json.Unmarshal(data, &menuJSON); err != nil {
		return err
	}
	
	m.ID = menuJSON.ID
	m.CreatedAt = menuJSON.CreatedAt
	m.UpdatedAt = menuJSON.UpdatedAt
	
	// 将menu_data转换为JSON字符串存储
	if menuJSON.MenuData != nil {
		menuDataBytes, err := json.Marshal(menuJSON.MenuData)
		if err != nil {
			return err
		}
		m.MenuData = string(menuDataBytes)
	}
	
	return nil
}

// List retrieves a list of menu versions based on parameters.
func (m *Menu) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*Menu, int64, error) {
	return dao.GenericQuery(params, m, queryFuncs...)
}

func (c *Menu) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, c, queryFuncs...)
}

// Delete removes menu records by their IDs.
func (m *Menu) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, m, utils.ToInt64Slice(ids), queryFuncs...)
}

// GetOne retrieves a single menu record by its ID.
func (m *Menu) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*Menu, error) {
	return dao.GenericGetOne(params, m, queryFuncs...)
}

// GetLatest retrieves the latest version of the menu.
func (m *Menu) GetLatest(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*Menu, error) {
	var menu Menu
	db := dao.DB()
	for _, f := range queryFuncs {
		db = f(db)
	}
	err := db.Order("version desc").First(&menu).Error
	if err != nil {
		return nil, err
	}
	return &menu, nil
}

// DeleteByID removes a single menu record by its ID.
func (m *Menu) DeleteByID(id int) error {
	return dao.DB().Delete(&Menu{}, id).Error
}
