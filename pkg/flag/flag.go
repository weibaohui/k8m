package flag

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/spf13/pflag"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
)

var config *Config
var once sync.Once

type Config struct {
	Port              int    // gin 监听端口
	KubeConfig        string // KUBECONFIG文件路径
	ApiKey            string // OPENAI_API_KEY
	ApiURL            string // OPENAI_API_URL
	ApiModel          string // OPENAI_MODEL
	Debug             bool   // 调试模式，同步修改所有的debug模式
	LogV              int    // klog的日志级别klog.V(this)
	InCluster         bool   // 是否集群内模式
	LoginType         string // password,oauth,token,.. 登录方式，默认为password
	AdminUserName     string // 管理员用户名
	AdminPassword     string // 管理员密码
	JwtTokenSecret    string // JWT token secret
	NodeShellImage    string // nodeShell 镜像
	KubectlShellImage string // kubectlShell 镜像
	SqlitePath        string // sqlite 数据库路径
}

func Init() *Config {
	once.Do(func() {
		config = &Config{}
		config.InitFlags()
	})
	return config
}

func (c *Config) InitFlags() {
	// 如果有其他类似的引用，请参考下面的方式进行整合
	// 初始化klog
	klog.InitFlags(nil)
	// 将Go的flag绑定到pflag
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	// 环境变量绑定
	// 默认端口为3618
	defaultPort := getEnvAsInt("PORT", 3618)

	// 默认kubeconfig为~/.kube/config
	defaultKubeConfig := getEnv("KUBECONFIG", filepath.Join(homedir.HomeDir(), ".kube", "config"))

	// 默认apiKey为环境变量OPENAI_API_KEY/OPENAI_API_URL/
	defaultApiKey := getEnv("OPENAI_API_KEY", "")
	defaultApiURL := getEnv("OPENAI_API_URL", "")
	defaultModel := getEnv("OPENAI_MODEL", "Qwen/Qwen2.5-Coder-7B-Instruct")

	// 默认登录方式为password
	defaultLoginType := getEnv("LOGIN_TYPE", "password")
	defaultAdminUserName := getEnv("ADMIN_USERNAME", "admin")
	defaultAdminPassword := getEnv("ADMIN_PASSWORD", "123456")

	// 默认debug为false
	defaultDebug := getEnvAsBool("DEBUG", false)

	// jwt token secret
	defaultJwtTokenSecret := getEnv("JWT_TOKEN_SECRET", "your-secret-key")

	// nodeShell 镜像
	defaultNodeShellImage := getEnv("NODE_SHELL_IMAGE", "alpine:latest")

	// kubectlShell 镜像
	// bitnami/kubectl:latest
	defaultKubectlShellImage := getEnv("KUBECTL_SHELL_IMAGE", "bitnami/kubectl:latest")
	// 输出日志的级别
	defaultLogV := getEnv("LOG_V", "2")

	// sqlite数据库文件路径
	defaultSqlitePath := getEnv("SQLITE_PATH", "/data/data.db")

	pflag.BoolVarP(&c.Debug, "debug", "d", defaultDebug, "调试模式")
	pflag.IntVarP(&c.Port, "port", "p", defaultPort, "监听端口")
	pflag.StringVarP(&c.ApiKey, "chatgpt-key", "k", defaultApiKey, "大模型的自定义API Key")
	pflag.StringVarP(&c.ApiURL, "chatgpt-url", "u", defaultApiURL, "大模型的自定义API URL")
	pflag.StringVarP(&c.ApiModel, "chatgpt-model", "m", defaultModel, "大模型的自定义模型名称")
	pflag.StringVarP(&c.KubeConfig, "kubeconfig", "c", defaultKubeConfig, "kubeconfig文件路径")
	pflag.StringVar(&c.LoginType, "login-type", defaultLoginType, "登录方式，password, oauth, token等,default is password")
	pflag.StringVar(&c.AdminUserName, "admin-username", defaultAdminUserName, "管理员用户名")
	pflag.StringVar(&c.AdminPassword, "admin-password", defaultAdminPassword, "管理员密码")
	pflag.StringVar(&c.JwtTokenSecret, "jwt-token-secret", defaultJwtTokenSecret, "登录后生成JWT token 使用的Secret")
	pflag.StringVar(&c.NodeShellImage, "node-shell-image", defaultNodeShellImage, "NodeShell 镜像。 默认为 alpine:latest，必须包含nsenter命令")
	pflag.StringVar(&c.KubectlShellImage, "kubectl-shell-image", defaultKubectlShellImage, "Kubectl Shell 镜像。默认为 bitnami/kubectl:latest，必须包含kubectl命令")
	pflag.IntVar(&c.LogV, "log-v", 2, "klog的日志级别klog.V(2)")
	pflag.StringVar(&c.SqlitePath, "sqlite-path", defaultSqlitePath, "sqlite数据库文件路径")

	// 检查是否设置了 --v 参数
	if vFlag := pflag.Lookup("v"); vFlag == nil || vFlag.Value.String() == "0" {
		// 如果没有设置，手动将 --v 设置为 环境变量值
		_ = flag.Set("v", defaultLogV)
	}
	pflag.Parse()

}

// getEnv 读取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt 读取环境变量，如果不存在则返回默认值
func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBool 获取环境变量的布尔值，支持 "true"/"false"（大小写不敏感）和 "1"/"0"，否则返回默认值
func getEnvAsBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
