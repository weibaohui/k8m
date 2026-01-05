package config

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-ldap/ldap/v3"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/response"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

type LdapConfigController struct {
}

// RegisterLdapConfigRoutes 注册路由
// 从 gin 切换到 chi，使用 chi.Router 替代 gin.RouterGroup
func RegisterLdapConfigRoutes(r chi.Router) {
	ctrl := &LdapConfigController{}
	// ldap 配置
	r.Get("/config/ldap/list", response.Adapter(ctrl.LDAPConfigList))
	r.Get("/config/ldap/{id}", response.Adapter(ctrl.LDAPConfigDetail))
	r.Post("/config/ldap/save", response.Adapter(ctrl.LDAPConfigSave))
	r.Post("/config/ldap/delete/{ids}", response.Adapter(ctrl.LDAPConfigDelete))
	r.Post("/config/ldap/save/id/{id}/status/{enabled}", response.Adapter(ctrl.LDAPConfigQuickSave))
	r.Post("/config/ldap/test_connect", response.Adapter(ctrl.LDAPConfigTestConnect))
}

// LDAP配置列表
func (lc *LdapConfigController) LDAPConfigList(c *response.Context) {
	params := dao.BuildParams(c)
	m := &models.LDAPConfig{}

	items, total, err := m.List(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 隐藏密码字段
	for _, item := range items {
		item.BindPassword = ""
	}

	amis.WriteJsonListWithTotal(c, total, items)
}

// 获取单个LDAP配置
func (lc *LdapConfigController) LDAPConfigDetail(c *response.Context) {
	id := c.Param("id")
	params := dao.BuildParams(c)
	m := &models.LDAPConfig{}
	conf, err := m.GetOne(params, func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	})
	if err != nil || conf == nil {
		c.JSON(http.StatusNotFound, response.H{"status": 1, "msg": "未找到配置"})
		return
	}
	c.JSON(http.StatusOK, response.H{"status": 0, "msg": "ok", "data": conf})
}

// 保存LDAP配置（新建/编辑）
func (lc *LdapConfigController) LDAPConfigSave(c *response.Context) {
	params := dao.BuildParams(c)
	m := models.LDAPConfig{}
	err := c.ShouldBindJSON(&m)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 判断是新建还是编辑
	if m.ID > 0 {
		// 编辑，若未填写密码则保留原密码
		var old models.LDAPConfig
		if err := dao.DB().Where("id = ?", m.ID).First(&old).Error; err == nil {
			if m.BindPassword == "" {
				// 保留原加密密码
				m.BindPassword = old.BindPassword
			} else if m.BindPassword != old.BindPassword {
				// 仅当填写新密码时才加密
				encrypted, err := utils.AesEncrypt([]byte(m.BindPassword))
				if err != nil {
					c.JSON(http.StatusInternalServerError, response.H{"status": 1, "msg": "密码加密失败"})
					return
				}
				m.BindPassword = base64.StdEncoding.EncodeToString(encrypted)
			}
		}
	} else {
		// 新增配置时也要对密码进行加密
		if m.BindPassword != "" {
			encrypted, err := utils.AesEncrypt([]byte(m.BindPassword))
			if err != nil {
				c.JSON(http.StatusInternalServerError, response.H{"status": 1, "msg": "密码加密失败"})
				return
			}
			m.BindPassword = base64.StdEncoding.EncodeToString(encrypted)
		}
	}

	// 保存数据库，仅更新指定字段，避免覆盖其他字段
	err = m.Save(params, func(db *gorm.DB) *gorm.DB {
		return db.Select([]string{"name", "host", "port", "bind_dn", "bind_password", "base_dn", "user_filter", "login2_auth_close", "default_group", "enabled"})
	})
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, response.H{
		"id": m.ID,
	})

}

// 删除LDAP配置（支持批量）
func (lc *LdapConfigController) LDAPConfigDelete(c *response.Context) {
	ids := c.Param("ids")
	params := dao.BuildParams(c)
	m := &models.LDAPConfig{}

	err := m.Delete(params, ids)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

// LDAPConfigQuickSave 快速保存启用状态
func (lc *LdapConfigController) LDAPConfigQuickSave(c *response.Context) {
	id := c.Param("id")
	enabled := c.Param("enabled")

	var entity models.LDAPConfig
	entity.ID = utils.ToUInt(id)

	if enabled == "true" {
		entity.Enabled = true
	} else {
		entity.Enabled = false
	}
	err := dao.DB().Model(&entity).Select("enabled").Updates(entity).Error

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonErrorOrOK(c, err)
}

// 获取ldap的enabled状态
func (lc *LdapConfigController) GetLdapConfig(c *response.Context) {
	// 查询是否有启用的LDAP配置
	var ldapConfig models.LDAPConfig
	err := dao.DB().Where("enabled = ?", true).Order("id desc").Limit(1).Find(&ldapConfig).Error

	// 构造前端需要的响应���式
	response := response.H{
		"enabled": ldapConfig.ID > 0, // 如果找���启用的配置，则ID > 0
	}

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, response)
}

// 测试LDAP连接
func (lc *LdapConfigController) LDAPConfigTestConnect(c *response.Context) {
	type Req struct {
		Host         string `json:"host"`
		Port         int    `json:"port"`
		BindDN       string `json:"bind_dn"`
		BindPassword string `json:"bind_password"`
		BaseDN       string `json:"base_dn"`
		UserFilter   string `json:"user_filter"`
	}
	var req Req
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.H{"status": 1, "msg": "参数错误"})
		return
	}

	addr := fmt.Sprintf("%s:%d", req.Host, req.Port)
	conn, err := ldap.Dial("tcp", addr)
	if err != nil {
		klog.Errorf("连接LDAP服务器失败: %v", err)
		c.JSON(http.StatusOK, response.H{"status": 1, "msg": fmt.Sprintf("连接LDAP服务器失败: %v", err)})
		return
	}
	defer conn.Close()

	// 先尝试用明文密码绑定
	if err := conn.Bind(req.BindDN, req.BindPassword); err == nil {
		c.JSON(http.StatusOK, response.H{"status": 0, "msg": "连接成功"})
		return
	}

	// 如果明文失败，再尝试解密（兼容编辑后密文场景）
	decryptedPwd := req.BindPassword
	if req.BindPassword != "" {
		if plain, err := utils.AesDecrypt(req.BindPassword); err == nil {
			decryptedPwd = string(plain)
			if err := conn.Bind(req.BindDN, decryptedPwd); err == nil {
				c.JSON(http.StatusOK, response.H{"status": 0, "msg": "连接成功"})
				return
			}
		}
	}

	klog.Errorf("管理员账号或密码错误")
	c.JSON(http.StatusOK, response.H{"status": 1, "msg": "管理员账号或密码错误"})
}
