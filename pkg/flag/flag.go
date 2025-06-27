package flag

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
)

var config *Config
var once sync.Once

type Config struct {
	Port                 int     // gin 监听端口
	Host                 string  // gin 监听地址
	KubeConfig           string  // KUBECONFIG文件路径
	ApiKey               string  // OPENAI_API_KEY
	ApiURL               string  // OPENAI_API_URL
	ApiModel             string  // OPENAI_MODEL
	Debug                bool    // 调试模式，同步修改所有的debug模式
	LogV                 int     // klog的日志级别klog.V(this)
	InCluster            bool    // 是否集群内模式
	LoginType            string  // password,oauth,token,.. 登录方式，默认为password
	EnableTempAdmin      bool    // 是否启用临时管理员账户配置
	AdminUserName        string  // 管理员用户名，启用临时管理员账户配置后生效
	AdminPassword        string  // 管理员密码，启用临时管理员账户配置后生效
	JwtTokenSecret       string  // JWT token secret
	NodeShellImage       string  // nodeShell 镜像
	KubectlShellImage    string  // kubectlShell 镜像
	ImagePullTimeout     int     // 镜像拉取超时时间（秒）
	PrintConfig          bool    // 是否打印配置信息
	Version              string  // 版本号，由编译时自动注入
	GitCommit            string  // git commit, 由编译时自动注入
	GitTag               string  // git tag, 由编译时自动注入
	GitRepo              string  // git仓库地址, 由编译时自动注入
	BuildDate            string  // 编译时间, 由编译时自动注入
	EnableAI             bool    // 是否启用AI功能，默认开启
	ConnectCluster       bool    // 启动程序后，是否自动连接发现的集群，默认关闭
	UseBuiltInModel      bool    // 是否使用内置大模型参数，默认开启
	ProductName          string  // 产品名称，默认为K8M
	ResourceCacheTimeout int     // 资源缓存时间（秒）
	Temperature          float32 // 模型温度
	TopP                 float32 //  模型topP参数
	MaxIterations        int32   //  模型自动对话的最大轮数
	MaxHistory           int32   //  模型对话上下文历史记录数
	AnySelect            bool    // 是否开启任意选择，默认开启
	DBDriver             string  // 数据库驱动类型: sqlite、mysql、postgresql等
	SqlitePath           string  // sqlite 数据库路径
	// MySQL 配置
	MysqlHost      string // mysql 主机
	MysqlPort      int    // mysql 端口
	MysqlUser      string // mysql 用户名
	MysqlPassword  string // mysql 密码
	MysqlDatabase  string // mysql 数据库名
	MysqlCharset   string // mysql 字符集
	MysqlCollation string // mysql 排序规则
	MysqlQuery     string // mysql 额外参数
	MysqlLogMode   bool   // mysql 日志模式
	// PostgreSQL 配置
	PgHost     string // postgres 主机
	PgPort     int    // postgres 端口
	PgUser     string // postgres 用户名
	PgPassword string // postgres 密码
	PgDatabase string // postgres 数据库名
	PgSSLMode  string // postgres sslmode
	PgTimeZone string // postgres 时区
	PgLogMode  bool   // postgres 日志模式
	Think      bool   // AI 是否开启思考过程输出，true 时显示思考过程，建议生产环境开启
	// LDAP配置
	LdapEnabled         bool   // 是否使用SSL连接LDAP服务器
	LdapHost            string // LDAP服务器地址
	LdapPort            string // LDAP服务器端口
	LdapUsername        string // LDAP用户名
	LdapPassword        string // LDAP密码
	LdapBaseDN          string // LDAP基础DN
	LdapBindUserDN      string // LDAP绑定用户DN
	LdapAnonymousQuery  int    // 是否允许匿名查询LDAP
	LdapUserField       string // LDAP用户字段
	LdapLogin2AuthClose bool   // LDAP登录后是否关闭认证

}

func Init() *Config {
	once.Do(func() {
		config = &Config{}
		loadEnv()
		config.InitFlags()

	})
	return config
}
func (c *Config) ShowConfigInfo() {
	// 根据PrintConfig决定是否打印配置信息
	if c.PrintConfig {
		klog.Infof("配置加载顺序:1.启动参数->2.环境变量->3.数据库参数设置（界面配置）,后加载的配置覆盖前面的配置")
		klog.Infof("已开启配置信息打印选项.\n%s:\n %+v\n%s\n", color.RedString("↓↓↓↓↓↓生产环境请务必关闭↓↓↓↓↓↓"), utils.ToJSON(config), color.RedString("↑↑↑↑↑生产环境请务必关闭↑↑↑↑↑↑"))
		c.ShowConfigCloseMethod()
	}
}
func (c *Config) ShowConfigCloseMethod() {
	klog.Infof("关闭打印选项方法：\n1. %s\n2. %s \n3. %s  \n", color.RedString("平台管理-参数设置-打印配置，选择关闭"), color.RedString("启动参数 --print-config = false"), color.RedString("env PRINT_CONFIG=false"))
}
func loadEnv() {
	env := os.Getenv("K8M_ENV")
	if env == "" {
		// 默认开发环境加载".env.dev.local"
		env = "dev"
	}
	// 依次加载并覆盖
	if err := godotenv.Overload(".env", ".env."+env+".local"); err != nil {
		klog.Warningf("Error loading .env file: %v", err)
	}
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

	// 默认监听地址为0.0.0.0
	defaultHost := getEnv("HOST", "0.0.0.0")

	// 默认kubeconfig为~/.kube/config
	defaultKubeConfig := getEnv("KUBECONFIG", filepath.Join(homedir.HomeDir(), ".kube", "config"))

	// 默认apiKey为环境变量OPENAI_API_KEY/OPENAI_API_URL/
	defaultApiKey := getEnv("OPENAI_API_KEY", "")
	defaultApiURL := getEnv("OPENAI_API_URL", "")
	defaultModel := getEnv("OPENAI_MODEL", "Qwen/Qwen2.5-7B-Instruct")

	// 默认登录方式为password
	defaultLoginType := getEnv("LOGIN_TYPE", "password")
	defaultAdminUserName := getEnv("ADMIN_USERNAME", "")
	defaultAdminPassword := getEnv("ADMIN_PASSWORD", "")

	// 默认debug为false
	defaultDebug := getEnvAsBool("DEBUG", false)

	defaultInCluster := getEnvAsBool("IN_CLUSTER", true)

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
	defaultSqlitePath := getEnv("SQLITE_PATH", "./data/k8m.db")

	// 默认开启任意选择
	defaultAnySelect := getEnvAsBool("ANY_SELECT", true)

	// 默认不打印配置
	defaultPrintConfig := getEnvAsBool("PRINT_CONFIG", false)
	// 默认开启AI功能
	defaultEnableAI := getEnvAsBool("ENABLE_AI", true)
	// 默认关闭启动连接集群
	defaultConnectCluster := getEnvAsBool("CONNECT_CLUSTER", false)
	// 默认使用内置大模型参数
	defaultUseBuiltInModel := getEnvAsBool("USE_BUILTIN_MODEL", true)
	// 默认不启用临时管理员账户配置
	defaultEnableTempAdmin := getEnvAsBool("ENABLE_TEMP_ADMIN", false)

	// 默认镜像拉取超时时间为30秒
	defaultImagePullTimeout := getEnvAsInt("IMAGE_PULL_TIMEOUT", 30)

	// 默认产品名称为K8M
	defaultProductName := getEnv("PRODUCT_NAME", "K8M")

	// 默认资源缓存时间为60秒
	defaultResourceCacheTimeout := getEnvAsInt("RESOURCE_CACHE_TIMEOUT", 60)

	// 默认模型自动对话的最大轮数为10
	defaultMaxIterations := getEnvAsInt32("MAX_ITERATIONS", 10)

	// MySQL 配置默认值
	defaultMysqlHost := getEnv("MYSQL_HOST", "127.0.0.1")
	defaultMysqlPort := getEnvAsInt("MYSQL_PORT", 3306)
	defaultMysqlUser := getEnv("MYSQL_USER", "root")
	defaultMysqlPassword := getEnv("MYSQL_PASSWORD", "")
	defaultMysqlDatabase := getEnv("MYSQL_DATABASE", "k8m")
	defaultMysqlCharset := getEnv("MYSQL_CHARSET", "utf8mb4")
	defaultMysqlCollation := getEnv("MYSQL_COLLATION", "utf8mb4_general_ci")
	defaultMysqlQuery := getEnv("MYSQL_QUERY", "parseTime=True&loc=Local")
	defaultMysqlLogMode := getEnvAsBool("MYSQL_LOGMODE", false)

	// Postgres 配置默认值
	defaultPgHost := getEnv("PG_HOST", "127.0.0.1")
	defaultPgPort := getEnvAsInt("PG_PORT", 5432)
	defaultPgUser := getEnv("PG_USER", "postgres")
	defaultPgPassword := getEnv("PG_PASSWORD", "")
	defaultPgDatabase := getEnv("PG_DATABASE", "k8m")
	defaultPgSSLMode := getEnv("PG_SSLMODE", "disable")
	defaultPgTimeZone := getEnv("PG_TIMEZONE", "Asia/Shanghai")
	defaultPgLogMode := getEnvAsBool("PG_LOGMODE", false)
	// ldap 配置默认值
	defaultLdapEnabled := getEnvAsBool("LDAP_ENABLED", false)
	defaultLdapHost := getEnv("LDAP_HOST", "")
	defaultLdapPort := getEnv("LDAP_PORT", "389")
	defaultLdapUsername := getEnv("LDAP_USERNAME", "")
	defaultLdapPassword := getEnv("LDAP_PASSWORD", "")
	defaultLdapBaseDN := getEnv("LDAP_BASEDN", "")
	defaultLdapBindUserDN := getEnv("LDAP_BINDUSERDN", "")
	defaultLdapAnonymousQuery := getEnvAsInt("LDAP_ANONYMOUSQUERY", 0)
	defaultLdapUserField := getEnv("LDAP_USERFIELD", "sAMAccountName")
	defaultLdapLogin2AuthClose := getEnvAsBool("LDAP_LOGIN2AUTHCLOSE", true)

	// 默认AI关闭思考过程输出为false
	defaultThink := getEnvAsBool("THINK", false)

	// 参数配置
	pflag.BoolVarP(&c.Debug, "debug", "d", defaultDebug, "调试模式")
	pflag.IntVarP(&c.Port, "port", "p", defaultPort, "监听端口,默认3618")
	pflag.StringVarP(&c.Host, "host", "h", defaultHost, "监听地址,默认0.0.0.0")

	pflag.StringVar(&c.ProductName, "product-name", defaultProductName, "产品名称，默认为K8M")

	pflag.StringVar(&c.LoginType, "login-type", defaultLoginType, "登录方式，password, oauth, token等,default is password")
	pflag.StringVar(&c.JwtTokenSecret, "jwt-token-secret", defaultJwtTokenSecret, "登录后生成JWT token 使用的Secret")

	// 临时管理员账户配置
	pflag.BoolVar(&c.EnableTempAdmin, "enable-temp-admin", defaultEnableTempAdmin, "是否启用临时管理员账户配置，默认关闭")
	pflag.StringVar(&c.AdminUserName, "admin-username", defaultAdminUserName, "管理员用户名，启用临时管理员账户配置后生效")
	pflag.StringVar(&c.AdminPassword, "admin-password", defaultAdminPassword, "管理员密码，启用临时管理员账户配置后生效")

	// k8s 集群配置
	pflag.StringVarP(&c.KubeConfig, "kubeconfig", "c", defaultKubeConfig, "kubeconfig文件路径")
	pflag.StringVar(&c.NodeShellImage, "node-shell-image", defaultNodeShellImage, "NodeShell 镜像。 默认为 alpine:latest，必须包含nsenter命令")
	pflag.StringVar(&c.KubectlShellImage, "kubectl-shell-image", defaultKubectlShellImage, "Kubectl Shell 镜像。默认为 bitnami/kubectl:latest，必须包含kubectl命令")
	pflag.IntVar(&c.ImagePullTimeout, "image-pull-timeout", defaultImagePullTimeout, "镜像拉取超时时间（秒），默认30秒")
	pflag.BoolVar(&c.InCluster, "in-cluster", defaultInCluster, "是否自动注册纳管宿主集群，默认启用")
	pflag.BoolVar(&c.ConnectCluster, "connect-cluster", defaultConnectCluster, "启动程序后，是否自动连接发现的集群，默认关闭  ")
	pflag.IntVar(&c.ResourceCacheTimeout, "resource-cache-timeout", defaultResourceCacheTimeout, "资源缓存时间（秒），默认60秒")

	// AI配置
	pflag.BoolVar(&c.EnableAI, "enable-ai", defaultEnableAI, "是否启用AI功能，默认开启")
	pflag.BoolVar(&c.AnySelect, "any-select", defaultAnySelect, "是否开启任意选择，默认开启")
	pflag.BoolVar(&c.Think, "think", defaultThink, "AI是否开启思考过程输出，true时显示思考过程，建议生产环境开启")
	pflag.Int32Var(&c.MaxIterations, "max-iterations", defaultMaxIterations, "模型自动对话的最大轮数，默认10轮")
	pflag.BoolVar(&c.UseBuiltInModel, "use-builtin-model", defaultUseBuiltInModel, "是否使用内置大模型参数，默认开启")
	pflag.StringVarP(&c.ApiKey, "chatgpt-key", "k", defaultApiKey, "大模型的自定义API Key")
	pflag.StringVarP(&c.ApiURL, "chatgpt-url", "u", defaultApiURL, "大模型的自定义API URL")
	pflag.StringVarP(&c.ApiModel, "chatgpt-model", "m", defaultModel, "大模型的自定义模型名称")

	// 数据库配置
	pflag.StringVar(&c.DBDriver, "db-driver", getEnv("DB_DRIVER", "sqlite"), "数据库驱动类型: sqlite、mysql、postgresql等")
	// 数据库-sqlite
	pflag.StringVar(&c.SqlitePath, "sqlite-path", defaultSqlitePath, "sqlite数据库文件路径，默认./data/k8m.db")
	// 数据库-mysql
	pflag.StringVar(&c.MysqlHost, "mysql-host", defaultMysqlHost, "MySQL主机地址")
	pflag.IntVar(&c.MysqlPort, "mysql-port", defaultMysqlPort, "MySQL端口")
	pflag.StringVar(&c.MysqlUser, "mysql-user", defaultMysqlUser, "MySQL用户名")
	pflag.StringVar(&c.MysqlPassword, "mysql-password", defaultMysqlPassword, "MySQL密码")
	pflag.StringVar(&c.MysqlDatabase, "mysql-database", defaultMysqlDatabase, "MySQL数据库名")
	pflag.StringVar(&c.MysqlCharset, "mysql-charset", defaultMysqlCharset, "MySQL字符集")
	pflag.StringVar(&c.MysqlCollation, "mysql-collation", defaultMysqlCollation, "MySQL排序规则")
	pflag.StringVar(&c.MysqlQuery, "mysql-query", defaultMysqlQuery, "MySQL连接额外参数")
	pflag.BoolVar(&c.MysqlLogMode, "mysql-logmode", defaultMysqlLogMode, "MySQL日志模式")
	// 数据库-postgresql
	pflag.StringVar(&c.PgHost, "pg-host", defaultPgHost, "PostgreSQL主机地址")
	pflag.IntVar(&c.PgPort, "pg-port", defaultPgPort, "PostgreSQL端口")
	pflag.StringVar(&c.PgUser, "pg-user", defaultPgUser, "PostgreSQL用户名")
	pflag.StringVar(&c.PgPassword, "pg-password", defaultPgPassword, "PostgreSQL密码")
	pflag.StringVar(&c.PgDatabase, "pg-database", defaultPgDatabase, "PostgreSQL数据库名")
	pflag.StringVar(&c.PgSSLMode, "pg-sslmode", defaultPgSSLMode, "PostgreSQL SSL模式")
	pflag.StringVar(&c.PgTimeZone, "pg-timezone", defaultPgTimeZone, "PostgreSQL时区")
	pflag.BoolVar(&c.PgLogMode, "pg-logmode", defaultPgLogMode, "PostgreSQL日志模式")

	// ldap配置
	pflag.BoolVar(&c.LdapEnabled, "ldap-enabled", defaultLdapEnabled, "是否使用启用LDAP登录")
	pflag.StringVar(&c.LdapHost, "ldap-host", defaultLdapHost, "LDAP服务器地址")
	pflag.StringVar(&c.LdapPort, "ldap-port", defaultLdapPort, "LDAP服务器端口")
	pflag.StringVar(&c.LdapUsername, "ldap-username", defaultLdapUsername, "LDAP用户名")
	pflag.StringVar(&c.LdapPassword, "ldap-password", defaultLdapPassword, "LDAP密码")
	pflag.StringVar(&c.LdapBaseDN, "ldap-basedn", defaultLdapBaseDN, "LDAP基础DN")
	pflag.StringVar(&c.LdapBindUserDN, "ldap-binduserdn", defaultLdapBindUserDN, "LDAP绑定用户DN")
	pflag.IntVar(&c.LdapAnonymousQuery, "ldap-anonymousquery", defaultLdapAnonymousQuery, "是否允许匿名查询LDAP，0表示不允许，1表示允许")
	pflag.StringVar(&c.LdapUserField, "ldap-userfield", defaultLdapUserField, "LDAP用户字段，默认为sAMAccountName")
	pflag.BoolVar(&c.LdapLogin2AuthClose, "ldap-login2authclose", defaultLdapLogin2AuthClose, "LDAP登录后是否关闭认证，默认开启")

	// 其他配置-打印配置信息
	pflag.BoolVar(&c.PrintConfig, "print-config", defaultPrintConfig, "是否打印配置信息，默认关闭")

	// 检查是否设置了 --v 参数
	pflag.IntVar(&c.LogV, "log-v", 2, "klog的日志级别klog.V(2)")
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

// getEnvAsBool 返回指定环境变量的布尔值，支持 "true"/"false"（不区分大小写）和 "1"/"0"，若未设置或解析失败则返回默认值。
func getEnvAsBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// getEnvAsInt32 返回指定环境变量的 int32 类型值，不存在或解析失败时返回默认值。
func getEnvAsInt32(key string, defaultValue int32) int32 {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.ParseInt(value, 10, 32); err == nil {
			return int32(intValue)
		}
	}
	return defaultValue
}
