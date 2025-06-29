package models

import (
	"fmt"
	"strings"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/flag"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

func init() {

	err := AutoMigrate()
	if err != nil {
		klog.Errorf("数据库迁移失败: %v", err.Error())
	}
	klog.V(4).Info("数据库自动迁移完成")

	_ = FixClusterName()
	_ = AddInnerMCPServer()
	_ = FixRoleName()
	_ = InitConfigTable()
	_ = InitConditionTable()
	_ = FixClusterAuthorizationTypeName()
	_ = AddInnerAdminUserGroup()
	_ = AddInnerAdminUser()
	_ = MigrateAIModel()
	_ = AddBuiltinLuaScripts()
}
func AutoMigrate() error {

	var errs []error
	// 添加需要迁移的所有模型

	if err := dao.DB().AutoMigrate(&CustomTemplate{}); err != nil {
		errs = append(errs, err)
	}
	if err := dao.DB().AutoMigrate(&KubeConfig{}); err != nil {
		errs = append(errs, err)
	}
	if err := dao.DB().AutoMigrate(&User{}); err != nil {
		errs = append(errs, err)
	}
	if err := dao.DB().AutoMigrate(&ClusterUserRole{}); err != nil {
		errs = append(errs, err)
	}
	if err := dao.DB().AutoMigrate(&OperationLog{}); err != nil {
		errs = append(errs, err)
	}
	if err := dao.DB().AutoMigrate(&ShellLog{}); err != nil {
		errs = append(errs, err)
	}
	if err := dao.DB().AutoMigrate(&HelmRepository{}); err != nil {
		errs = append(errs, err)
	}
	// MYSQL 下需要单独处理 content字段为LONGTEXT，pg、sqlite不需要处理
	if dao.DB().Migrator().HasTable(&HelmRepository{}) && dao.DB().Dialector.Name() == "mysql" {
		dao.DB().Exec("ALTER TABLE helm_repositories MODIFY COLUMN content LONGTEXT")
	}

	if err := dao.DB().AutoMigrate(&HelmChart{}); err != nil {
		errs = append(errs, err)
	}
	if err := dao.DB().AutoMigrate(&UserGroup{}); err != nil {
		errs = append(errs, err)
	}
	if err := dao.DB().AutoMigrate(&MCPServerConfig{}); err != nil {
		errs = append(errs, err)
	}
	if err := dao.DB().AutoMigrate(&MCPTool{}); err != nil {
		errs = append(errs, err)
	}
	if err := dao.DB().AutoMigrate(&Config{}); err != nil {
		errs = append(errs, err)
	}
	if err := dao.DB().AutoMigrate(&ApiKey{}); err != nil {
		errs = append(errs, err)
	}
	if err := dao.DB().AutoMigrate(&ConditionReverse{}); err != nil {
		errs = append(errs, err)
	}
	if err := dao.DB().AutoMigrate(&MCPToolLog{}); err != nil {
		errs = append(errs, err)
	}
	if err := dao.DB().AutoMigrate(&McpKey{}); err != nil {
		errs = append(errs, err)
	}
	if err := dao.DB().AutoMigrate(&SSOConfig{}); err != nil {
		errs = append(errs, err)
	}
	if err := dao.DB().AutoMigrate(&AIModelConfig{}); err != nil {
		errs = append(errs, err)
	}
	if err := dao.DB().AutoMigrate(&InspectionCheckEvent{}); err != nil {
		errs = append(errs, err)
	}
	if err := dao.DB().AutoMigrate(&InspectionRecord{}); err != nil {
		errs = append(errs, err)
	}
	if err := dao.DB().AutoMigrate(&InspectionSchedule{}); err != nil {
		errs = append(errs, err)
	}
	if err := dao.DB().AutoMigrate(&InspectionScriptResult{}); err != nil {
		errs = append(errs, err)
	}
	if err := dao.DB().AutoMigrate(&InspectionLuaScript{}); err != nil {
		errs = append(errs, err)
	}
	if err := dao.DB().AutoMigrate(&InspectionLuaScriptBuiltinVersion{}); err != nil {
		errs = append(errs, err)
	}
	if err := dao.DB().AutoMigrate(&WebhookReceiver{}); err != nil {
		errs = append(errs, err)
	}
	// 删除 user 表 name 字段，已弃用
	if dao.DB().Migrator().HasColumn(&User{}, "Role") {
		if err := dao.DB().Migrator().DropColumn(&User{}, "Role"); err != nil {
			// 判断是不是check that column/key exists
			// 不存在，不用删除，那么不用报错误
			if !strings.Contains(err.Error(), "check that column/key exists") {
				errs = append(errs, err)
			}
		}
	}

	// 打印所有非nil的错误
	for _, err := range errs {
		if err != nil {
			klog.Errorf("数据库迁移报错: %v", err.Error())
		}
	}

	return nil
}
func AddBuiltinLuaScripts() error {
	// 检查数据库中记录的内置脚本版本
	db := dao.DB()
	version, err := GetBuiltinLuaScriptsVersion(db)
	if err == nil {
		// 有记录，判断是否需要更新
		if version == BuiltinLuaScriptsVersion {
			// 版本一致，无需更新
			return nil
		}
	}
	// 版本不一致或无记录，先删除所有内置脚本
	if err := db.Where("script_type = ?", constants.LuaScriptTypeBuiltin).Delete(&InspectionLuaScript{}).Error; err != nil {
		klog.Errorf("删除旧内置巡检脚本失败: %v", err)
		return err
	}
	// 插入最新内置脚本
	if err := db.CreateInBatches(BuiltinLuaScripts, 100).Error; err != nil {
		klog.Errorf("插入内置巡检脚本失败: %v", err)
		return err
	}
	// 更新版本号
	if err := SetBuiltinLuaScriptsVersion(db, BuiltinLuaScriptsVersion); err != nil {
		klog.Errorf("更新内置脚本版本号失败: %v", err)
		return err
	}
	return nil
}
func FixRoleName() error {
	// 将用户组表中角色进行统一，除了平台管理员以外，都更新为普通用户guest
	err := dao.DB().Model(&UserGroup{}).Where("role != ?", "platform_admin").Update("role", "guest").Error
	if err != nil {
		klog.Errorf("更新用户组表中角色失败: %v", err)
		return err
	}

	return nil
}
func FixClusterAuthorizationTypeName() error {
	// 将用户组表中角色进行统一，除了平台管理员以外，都更新为普通用户guest
	err := dao.DB().Model(&ClusterUserRole{}).Where("authorization_type = '' or authorization_type is null").Update("authorization_type", "user").Error
	if err != nil {
		klog.Errorf("更新用户组表中角色失败: %v", err)
		return err
	}

	return nil
}
func FixClusterName() error {
	// 将display_name为空的记录更新为cluster字段
	result := dao.DB().Model(&KubeConfig{}).Where("display_name = ?", "").Update("display_name", gorm.Expr("cluster"))
	if result.Error != nil {
		klog.Errorf("更新cluster_name失败: %v", result.Error)
		return result.Error
	}
	return nil
}

// AddInnerMCPServer 检查并初始化名为 "k8m" 的内部 MCP 服务器配置，不存在则创建，已存在则更新其 URL。
func AddInnerMCPServer() error {
	// 检查是否存在名为k8m的记录
	var count int64
	if err := dao.DB().Model(&MCPServerConfig{}).Where("name = ?", "k8m").Count(&count).Error; err != nil {
		klog.Errorf("查询MCP服务器配置失败: %v", err)
		return err
	}
	cfg := flag.Init()
	// 如果不存在，添加默认的内部MCP服务器配置
	if count == 0 {
		config := &MCPServerConfig{
			Name:    "k8m",
			URL:     fmt.Sprintf("http://localhost:%d/mcp/k8m/sse", cfg.Port),
			Enabled: false,
		}
		if err := dao.DB().Create(config).Error; err != nil {
			klog.Errorf("添加内部MCP服务器配置失败: %v", err)
			return err
		}
		klog.V(4).Info("成功添加内部MCP服务器配置")
	} else {
		klog.V(4).Info("内部MCP服务器配置已存在")
		dao.DB().Model(&MCPServerConfig{}).Select("url").
			Where("name =?", "k8m").
			Update("url", fmt.Sprintf("http://localhost:%d/mcp/k8m/sse", cfg.Port))
	}

	return nil
}
func InitConfigTable() error {
	var count int64
	if err := dao.DB().Model(&Config{}).Count(&count).Error; err != nil {
		klog.Errorf("查询配置表: %v", err)
		return err
	}
	if count == 0 {
		config := &Config{
			PrintConfig: false,
			EnableAI:    true,
			AnySelect:   true,
			LoginType:   "password",
		}
		if err := dao.DB().Create(config).Error; err != nil {
			klog.Errorf("初始化配置表失败: %v", err)
			return err
		}
		klog.V(4).Info("成功初始化配置表")
	}

	return nil
}

func InitConditionTable() error {
	var count int64
	if err := dao.DB().Model(&ConditionReverse{}).Count(&count).Error; err != nil {
		klog.Errorf("查询翻转指标配置表: %v", err)
		return err
	}
	if count == 0 {
		// 初始化需要翻转的指标
		conditions := []ConditionReverse{
			{Name: "Pressure", Enabled: true},
			{Name: "Unavailable", Enabled: true},
			{Name: "Problem", Enabled: true},
			{Name: "Error", Enabled: true},
			{Name: "Slow", Enabled: true},
		}
		if err := dao.DB().Create(&conditions).Error; err != nil {
			klog.Errorf("初始化翻转指标配置失败: %v", err)
			return err
		}

		klog.V(4).Info("成功初始化翻转指标配置表")
	}

	return nil
}

// AddInnerAdminUser 添加内置管理员账户
func AddInnerAdminUser() error {
	// 检查是否存在名为k8m的记录
	var count int64
	if err := dao.DB().Model(&User{}).Count(&count).Error; err != nil {
		klog.Errorf("统计用户数错误: %v", err)
		return err
	}
	if count > 0 {
		klog.V(4).Info("已存在用户，不再添加默认管理员用户")
		return nil
	}
	if err := dao.DB().Model(&User{}).Where("username = ?", "k8m").Count(&count).Error; err != nil {
		klog.Errorf("查看k8m默认用户是否存在，发生错误: %v", err)
		return err
	}
	// 如果不存在，添加默认的一个默认的平台管理员账户
	// 用户名为: k8m
	// 密码为: k8m
	if count == 0 {
		config := &User{
			Username:   "k8m",
			Salt:       "grfi92rq",
			Password:   "8RGCXWw6IzgKDPyeFKt6Kw==",
			GroupNames: "平台管理员组",
		}
		if err := dao.DB().Create(config).Error; err != nil {
			klog.Errorf("添加默认平台管理员账户失败: %v", err)
			return err
		}
		klog.V(4).Info("成功添加默认平台管理员账户")
	} else {
		klog.V(4).Info("默认平台管理员k8m账户已存在")
	}

	return nil
}

// AddInnerAdminUserGroup 添加内置管理员账户组
func AddInnerAdminUserGroup() error {
	// 检查是否存在名为 平台管理员组 的内置管理员账户组的记录
	var count int64
	if err := dao.DB().Model(&UserGroup{}).Where("group_name = ?", "平台管理员组").Count(&count).Error; err != nil {
		klog.Errorf("已存在内置 平台管理员组 角色: %v", err)
		return err
	}
	// 如果不存在，添加默认的内部MCP服务器配置
	if count == 0 {
		config := &UserGroup{
			GroupName: "平台管理员组",
			Role:      "platform_admin",
		}
		if err := dao.DB().Create(config).Error; err != nil {
			klog.Errorf("添加默认平台管理员组失败: %v", err)
			return err
		}
		klog.V(4).Info("成功添加默认平台管理员组")
	} else {
		klog.V(4).Info("默认平台管理员组已存在")
	}

	return nil
}

func MigrateAIModel() error {
	// 将旧表 config 中的 AI 配置字段批量搬迁到新表 ai_model_configs
	// 先看AIModelConfigs是否有数据，如果有数据，则停止迁移
	model := &AIModelConfig{}
	_, count, err := model.List(nil)
	if err != nil {
		klog.Errorf("查询新表 ai_model_configs 失败: %v", err)
	}
	if count > 0 {
		klog.V(4).Info("新表 ai_model_configs 已有数据，不再进行迁移")
		// 不需要进行迁移
		return nil
	}

	if !dao.DB().Migrator().HasColumn(&Config{}, "api_key") {
		// 不需要进行迁移
		klog.Infof("参数表config 无老版本API_KEY相关配置,无需进行迁移")
		return nil
	}

	row := dao.DB().Raw("select api_key,api_model,api_url,temperature,top_p from configs limit 1").Row()

	err = row.Scan(&model.ApiKey, &model.ApiModel, &model.ApiURL, &model.Temperature, &model.TopP)
	if err != nil {
		// 不需要进行迁移
		klog.Infof("查询旧表 config 失败: %v,不再进行迁移", err)
		return nil
	}

	// 检查这几个 字段是否为空，如果为空，则不进行迁移
	if model.ApiKey == "" && model.ApiModel == "" && model.ApiURL == "" {
		klog.V(4).Info("旧表 config 中的 AI 配置字段为空，不再进行迁移")
		// 不需要进行迁移
		return nil
	}

	err = model.Save(nil)
	if err != nil {
		klog.Errorf("保存新表 ai_model_configs 失败: %v", err)
		return err
	}

	// 更新config表，记录ModelID
	dao.DB().Model(&Config{}).Update("model_id", model.ID)
	return nil
}
