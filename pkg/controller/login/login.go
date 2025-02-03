package login

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/weibaohui/k8m/pkg/flag"
)

// 定义 JWT 密钥
// todo 作为参数项
var jwtSecret = []byte("your-secret-key")

// LoginRequest 用户结构体
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// 生成 Token
func generateToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(), // 24小时有效
	})
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
	// 这里可以替换为数据库验证
	if req.Username == cfg.AdminUserName && req.Password == cfg.AdminPassword {
		token, _ := generateToken(req.Username)
		c.JSON(http.StatusOK, gin.H{"token": token})
		return
	}
	c.JSON(http.StatusUnauthorized, gin.H{"message": "用户名或密码错误"})
}
