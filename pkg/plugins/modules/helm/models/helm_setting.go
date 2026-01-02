package models

import (
	"errors"
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"gorm.io/gorm"
)

type HelmSetting struct {
	ID             uint   `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	HelmCachePath  string `json:"helm_cache_path"`
	HelmUpdateCron string `json:"helm_update_cron"`

	CreatedAt time.Time `json:"created_at,omitempty" gorm:"<-:create"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

func (HelmSetting) TableName() string {
	return "helm_settings"
}

func DefaultHelmSetting() *HelmSetting {

	// 	defaultHelmCachePath := getEnv("HELM_CACHE_PATH", "/tmp/helm-cache")
	// defaultHelmUpdateCron := getEnv("HELM_UPDATE_CRON", "0 */6 * * *")

	return &HelmSetting{
		HelmCachePath:  "/tmp/helm-cache",
		HelmUpdateCron: "0 */6 * * *",
	}
}

func GetOrCreateHelmSetting() (*HelmSetting, error) {
	db := dao.DB()
	var s HelmSetting
	if err := db.Order("id asc").First(&s).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			def := DefaultHelmSetting()
			if cErr := db.Create(def).Error; cErr != nil {
				return nil, cErr
			}
			return def, nil
		}
		return nil, err
	}
	return &s, nil
}

func UpdateHelmSetting(in *HelmSetting) (*HelmSetting, error) {
	if in == nil {
		return nil, nil
	}
	cur, err := GetOrCreateHelmSetting()
	if err != nil {
		return nil, err
	}

	cur.HelmCachePath = in.HelmCachePath
	cur.HelmUpdateCron = in.HelmUpdateCron

	if err := dao.DB().Save(cur).Error; err != nil {
		return nil, err
	}
	return cur, nil
}
