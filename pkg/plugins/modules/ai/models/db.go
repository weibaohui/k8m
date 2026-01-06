package models

import (
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/models"
	"k8s.io/klog/v2"
)

func InitDB() error {
	return dao.DB().AutoMigrate(&AIModelConfig{}, &AIPrompt{}, &AIRunConfig{})
}

func UpgradeDB(fromVersion string, toVersion string) error {
	klog.V(6).Infof("开始升级 AI 插件数据库：从版本 %s 到版本 %s", fromVersion, toVersion)
	if err := dao.DB().AutoMigrate(&AIModelConfig{}, &AIPrompt{}, &AIRunConfig{}); err != nil {
		klog.V(6).Infof("自动迁移 AI 插件数据库失败: %v", err)
		return err
	}

	klog.V(6).Infof("升级 AI 插件数据库完成")
	return nil
}

func DropDB() error {
	db := dao.DB()
	if db.Migrator().HasTable(&AIModelConfig{}) {
		if err := db.Migrator().DropTable(&AIModelConfig{}); err != nil {
			klog.V(6).Infof("删除 AI Model Config 表失败: %v", err)
			return err
		}
	}
	if db.Migrator().HasTable(&AIPrompt{}) {
		if err := db.Migrator().DropTable(&AIPrompt{}); err != nil {
			klog.V(6).Infof("删除 AI Prompt 表失败: %v", err)
			return err
		}
	}
	if db.Migrator().HasTable(&AIRunConfig{}) {
		if err := db.Migrator().DropTable(&AIRunConfig{}); err != nil {
			klog.V(6).Infof("删除 AI Run Config 表失败: %v", err)
			return err
		}
	}
	klog.V(6).Infof("已删除 AI 插件表及数据")
	return nil
}

func InitBuiltinAIPrompts() error {
	var count int64
	dao.DB().Model(&AIPrompt{}).Where("is_builtin = ?", true).Count(&count)
	if count > 0 {
		return nil
	}

	for _, prompt := range BuiltinAIPrompts {
		var existing AIPrompt
		err := dao.DB().Where("name = ? AND prompt_type = ?", prompt.Name, prompt.PromptType).First(&existing).Error
		if err == nil {
			continue
		}

		if err := dao.DB().Create(&prompt).Error; err != nil {
			return err
		}
	}

	return nil
}

func MigrateAIModel() error {
	model := &AIModelConfig{}
	_, count, err := model.List(nil)
	if err != nil {
		klog.Errorf("查询新表 ai_model_configs 失败: %v", err)
	}
	if count > 0 {
		klog.V(4).Info("新表 ai_model_configs 已有数据，不再进行迁移")
		return nil
	}

	if !dao.DB().Migrator().HasColumn(&models.Config{}, "api_key") {
		klog.Infof("参数表config 无老版本API_KEY相关配置,无需进行迁移")
		return nil
	}

	row := dao.DB().Raw("select api_key,api_model,api_url,temperature,top_p from configs limit 1").Row()

	err = row.Scan(&model.ApiKey, &model.ApiModel, &model.ApiURL, &model.Temperature, &model.TopP)
	if err != nil {
		klog.Infof("查询旧表 config 失败: %v,不再进行迁移", err)
		return nil
	}

	if model.ApiKey == "" && model.ApiModel == "" && model.ApiURL == "" {
		klog.V(4).Info("旧表 config 中的 AI 配置字段为空，不再进行迁移")
		return nil
	}

	err = model.Save(nil)
	if err != nil {
		klog.Errorf("保存新表 ai_model_configs 失败: %v", err)
		return err
	}

	dao.DB().Model(&models.Config{}).Update("model_id", model.ID)
	return nil
}
