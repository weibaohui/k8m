package login

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/flag"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/service"
)

var secretKey = "secret-key-16-ok"

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
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误"})
		return
	}
	// 初始化配置
	cfg := flag.Init()

	// 对密码进行解密
	decrypt, err := AesDecrypt(req.Password)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 验证用户名和密码
	// 1、从cfg中获取用户名，先判断是不是admin，是进行密码比对
	// 2、从DB中获取用户名密码

	if req.Username == cfg.AdminUserName {
		// cfg 用户名密码
		if string(decrypt) != cfg.AdminPassword {
			// 前端处理登录状态码，不要修改
			c.JSON(http.StatusUnauthorized, gin.H{"message": "用户名或密码错误"})
			return
		}
		token, _ := generateToken(req.Username, models.RoleAdmin)
		c.JSON(http.StatusOK, gin.H{"token": token})
		return
	} else {
		// DB 用户名密码
		list, err := service.UserService().List()
		if err != nil {
			amis.WriteJsonError(c, err)
			return
		}
		for _, v := range list {
			if v.Username == req.Username {
				decryptDBPsw, err := AesDecrypt(v.Password)
				if err != nil {
					amis.WriteJsonError(c, err)
					return
				}
				if string(decrypt) != string(decryptDBPsw) {
					// 前端处理登录状态码，不要修改
					c.JSON(http.StatusUnauthorized, gin.H{"message": "用户名或密码错误"})
					return
				}
				token, _ := generateToken(v.Username, v.Role)
				c.JSON(http.StatusOK, gin.H{"token": token})
				return
			}
		}
	}

	c.JSON(http.StatusUnauthorized, gin.H{"message": "用户名或密码错误"})
}

// pkcs7Padding 填充
func pkcs7Padding(data []byte, blockSize int) []byte {
	// 判断缺少几位长度。最少1，最多 blockSize
	padding := blockSize - len(data)%blockSize
	// 补足位数。把切片[]byte{byte(padding)}复制padding个
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// pkcs7UnPadding 填充的反向操作
func pkcs7UnPadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("加密字符串错误！")
	}
	// 获取填充的个数
	unPadding := int(data[length-1])
	return data[:(length - unPadding)], nil
}

// AesEncrypt 加密
func AesEncrypt(data []byte) ([]byte, error) {
	key := []byte(secretKey)
	// 创建加密实例
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	// 判断加密块的大小
	blockSize := block.BlockSize()
	// 填充
	encryptBytes := pkcs7Padding(data, blockSize)
	// 初始化加密数据接收切片
	crypted := make([]byte, len(encryptBytes))
	// 使用cbc加密模式
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	// 执行加密
	blockMode.CryptBlocks(crypted, encryptBytes)
	return crypted, nil
}

// AesDecrypt 解密
func AesDecrypt(ciphertextBase64 string) ([]byte, error) {

	// 解码 Base64 密文
	data, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return nil, err
	}

	key := []byte(secretKey)
	// 创建实例
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	// 获取块的大小
	blockSize := block.BlockSize()
	// 使用cbc
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	// 初始化解密数据接收切片
	crypted := make([]byte, len(data))
	// 执行解密
	blockMode.CryptBlocks(crypted, data)
	// 去除填充
	crypted, err = pkcs7UnPadding(crypted)
	if err != nil {
		return nil, err
	}
	return crypted, nil
}
