package login

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/models"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/totp"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/flag"
	"github.com/weibaohui/k8m/pkg/service"
	"k8s.io/klog/v2"
)

var ErrorUerPassword = errors.New("用户名密码错误")

// LoginRequest 用户结构体
type LoginRequest struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	LoginType int    `json:"loginType"` // 0: 普通登录, 1: LDAP登录
	Code      string `json:"code"`
}

// 验证用户名和密码
// 1、从cfg中获取用户名，先判断是不是admin，是进行密码比对.必须启用临时管理员配置才进行这一步
// 2、从DB中获取用户名密码
func LoginByPassword(c *gin.Context) {
	var req LoginRequest
	errorInfo := gin.H{"message": "用户名密码错误或用户被禁用"}
	if err := c.ShouldBindJSON(&req); err != nil {
		klog.Errorf("LoginByPassword %v", err.Error())
		c.JSON(http.StatusUnauthorized, errorInfo)
		return
	}

	// 初始化配置
	cfg := flag.Init()

	// 对密码进行解密
	decrypt, err := utils.AesDecrypt(req.Password)
	if err != nil {
		klog.Errorf("LoginByPassword %v", err.Error())
		c.JSON(http.StatusUnauthorized, errorInfo)
		return
	}

	// LDAP登录判断
	if req.LoginType == 1 {
		if err := handleLDAPLogin(c, req.Username, string(decrypt), req.Code, cfg); err != nil {
			// 前端处理登录状态码，不要修改
			c.JSON(http.StatusUnauthorized, errorInfo)
			return
		}
		return
	}

	if req.Username == cfg.AdminUserName && cfg.EnableTempAdmin {
		// cfg 用户名密码
		if string(decrypt) != cfg.AdminPassword {
			// 前端处理登录状态码，不要修改
			c.JSON(http.StatusUnauthorized, errorInfo)
			return
		}
		// Admin用户不需要2FA验证
		token, _ := service.UserService().GenerateJWTToken(req.Username, []string{constants.RolePlatformAdmin}, nil, time.Hour*24)
		c.JSON(http.StatusOK, gin.H{"token": token})
		return
	} else {
		// DB 用户名密码
		list, err := service.UserService().List()
		if err != nil {
			klog.Errorf("LoginByPassword %v", err.Error())
			c.JSON(http.StatusUnauthorized, errorInfo)
			return
		}
		for _, v := range list {
			if v.Username == req.Username {

				if v.Disabled {
					klog.Errorf("用户[%s]被禁用", v.Username)
					c.JSON(http.StatusUnauthorized, errorInfo)
					return
				}

				// password base64解密
				// 前端密码解密的值，加上盐值，重新计算
				decryptPsw, err := utils.AesEncrypt([]byte(fmt.Sprintf("%s%s", string(decrypt), v.Salt)))
				if err != nil {
					klog.Errorf("LoginByPassword %v", err.Error())
					c.JSON(http.StatusUnauthorized, errorInfo)
					return
				}
				dbPsw, err := base64.StdEncoding.DecodeString(v.Password)
				if err != nil {
					klog.Errorf("LoginByPassword %v", err.Error())
					c.JSON(http.StatusUnauthorized, errorInfo)
					return
				}
				if !bytes.Equal(dbPsw, decryptPsw) {
					// 前端处理登录状态码，不要修改
					c.JSON(http.StatusUnauthorized, errorInfo)
					return
				}

				// 检查是否启用了2FA
				if v.TwoFAEnabled {
					// 如果启用了2FA但未提供验证码
					if req.Code == "" {
						c.JSON(http.StatusUnauthorized, gin.H{"message": "请输入2FA验证码"})
						return
					}
					// 验证2FA代码
					if !totp.ValidateCode(v.TwoFASecret, req.Code) {
						c.JSON(http.StatusUnauthorized, gin.H{"message": "2FA验证码错误"})
						return
					}
				}

				token, _ := service.UserService().GenerateJWTTokenByUserName(v.Username, 24*time.Hour)
				c.JSON(http.StatusOK, gin.H{"token": token})
				return
			}
		}
	}

	c.JSON(http.StatusUnauthorized, errorInfo)
}

// handleLDAPLogin 处理LDAP登录流程
func handleLDAPLogin(c *gin.Context, username, password, code string, cfg *flag.Config) error {
	// 1. LDAP认证
	_, err := service.UserService().LoginWithLdap(username, password, cfg)
	if err != nil {
		klog.Errorf("LDAP登录失败: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"message": "LDAP登录验证失败"})
		return err
	}

	// 2. 检查或创建用户
	if err := service.UserService().CheckAndCreateUser(username, "ldap"); err != nil {
		klog.Errorf("创建/检查LDAP用户失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "系统错误"})
		return err
	}

	// 3. 获取用户信息
	user, err := getUserInfo(username)
	if err != nil {
		klog.Errorf("获取用户信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "系统错误"})
		return err
	}

	// 4. 验证2FA
	if err := validateTwoFA(user, code, c); err != nil {
		return err
	}

	// 5. 生成token
	token, _ := service.UserService().GenerateJWTTokenByUserName(username, 24*time.Hour)
	c.JSON(http.StatusOK, gin.H{"token": token})
	return nil
}

// getUserInfo 获取用户信息
func getUserInfo(username string) (*models.User, error) {
	params := &dao.Params{}
	user := &models.User{}
	queryFunc := func(db *gorm.DB) *gorm.DB {
		return db.Where("username = ?", username)
	}
	return user.GetOne(params, queryFunc)
}

// validateTwoFA 验证2FA
func validateTwoFA(user *models.User, code string, c *gin.Context) error {
	if user != nil && user.TwoFAEnabled {
		if code == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "请输入2FA验证码"})
			return errors.New("2FA验证码未提供")
		}
		if !totp.ValidateCode(user.TwoFASecret, code) {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "2FA验证码错误"})
			return errors.New("2FA验证码错误")
		}
	}
	return nil
}
