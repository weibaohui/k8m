package login

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/weibaohui/k8m/pkg/comm/utils"
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
}

// 生成 Token
func generateToken(username, role string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		constants.JwtUserName: username,
		constants.JwtUserRole: role,
		"exp":                 time.Now().Add(24 * time.Hour).Unix(), // 24小时有效
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
		token, _ := generateToken(req.Username, models.RolePlatformAdmin)
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

				token, _ := generateToken(v.Username, v.Role)
				c.JSON(http.StatusOK, gin.H{"token": token})
				return
			}
		}
	}

	c.JSON(http.StatusUnauthorized, errorInfo)
}
