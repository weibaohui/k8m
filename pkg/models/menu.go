package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

// Menu represents a versioned menu structure.
// Each save operation creates a new record with an incremented version.
type Menu struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	MenuData  string    `gorm:"type:json" json:"menu_data,omitempty"`
	Version   int       `gorm:"index;unique" json:"version,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// List retrieves a list of menu versions based on parameters.
func (m *Menu) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*Menu, int64, error) {
	return dao.GenericQuery(params, m, queryFuncs...)
}

func (c *Menu) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	c.Version = c.Version + 1
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
