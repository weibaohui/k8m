package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

type Template struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	Name      string    `gorm:"index" json:"name,omitempty"`
	Content   string    `gorm:"type:text" json:"content,omitempty"`
	Kind      string    `gorm:"index" json:"kind,omitempty"`
	Cluster   string    `gorm:"index" json:"cluster,omitempty"`
	IsGlobal  bool      `gorm:"index" json:"is_global,omitempty"`
	CreatedBy string    `gorm:"index" json:"created_by,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty" gorm:"<-:create"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

func (t *Template) TableName() string {
	return "yaml_editor_templates"
}

func InitDB() error {
	return dao.DB().AutoMigrate(&Template{})
}

func DropDB() error {
	return dao.DB().Migrator().DropTable(&Template{})
}

func UpgradeDB(fromVersion, toVersion string) error {
	return nil
}

func (t *Template) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*Template, int64, error) {
	return dao.GenericQuery(params, t, queryFuncs...)
}

func (t *Template) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, t, queryFuncs...)
}

func (t *Template) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, t, utils.ToInt64Slice(ids), queryFuncs...)
}

func (t *Template) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*Template, error) {
	return dao.GenericGetOne(params, t, queryFuncs...)
}
