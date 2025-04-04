package login

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/totp"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/flag"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/service"
	"k8s.io/klog/v2"
)

var ErrorUerPassword = errors.New("用户名密码错误")

// LoginRequest 用户结构体
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Code     string `json:"code"`
}

// 生成 Token
func generateToken(username string, roles []string, clusters []string) (string, error) {
	role := constants.JwtUserRole
	name := constants.JwtUserName
	cst := constants.JwtClusters
	var token = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		name:  username,
		role:  strings.Join(roles, ","),
		cst:   strings.Join(clusters, ","),
		"exp": time.Now().Add(24 * time.Hour).Unix(), // 24小时有效
	})
	cfg := flag.Init()
	var jwtSecret = []byte(cfg.JwtTokenSecret)
	return token.SignedString(jwtSecret)
}

func LoginByPassword(c *gin.Context) {
	var req LoginRequest
	errorInfo := gin.H{"message": "用户名或密码错误"}
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

	// 验证用户名和密码
	// 1、从cfg中获取用户名，先判断是不是admin，是进行密码比对
	// 2、从DB中获取用户名密码

	if req.Username == cfg.AdminUserName {
		// cfg 用户名密码
		if string(decrypt) != cfg.AdminPassword {
			// 前端处理登录状态码，不要修改
			c.JSON(http.StatusUnauthorized, errorInfo)
			return
		}
		// Admin用户不需要2FA验证
		token, _ := generateToken(req.Username, []string{models.RolePlatformAdmin}, []string{})
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
				var ugList []models.UserGroup
				err = dao.DB().Model(&models.UserGroup{}).Where("group_name in ?", strings.Split(v.GroupNames, ",")).Distinct("role").Find(&ugList).Error
				if err != nil {
					c.JSON(http.StatusUnauthorized, errorInfo)
					return
				}
				// 查询所有的用户组，判断用户组的角色
				// 形成一个用户组对应的角色列表
				var roles []string
				for _, ug := range ugList {
					roles = append(roles, ug.Role)
				}

				//查询用户对应的集群
				clusters, _ := service.UserService().GetClusters(v.Username)
				token, _ := generateToken(v.Username, roles, clusters)
				c.JSON(http.StatusOK, gin.H{"token": token})
				return
			}
		}
	}

	c.JSON(http.StatusUnauthorized, errorInfo)
}
