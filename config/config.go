package config

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/pflag"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
)

var config *Config

type Config struct {
	Port       int
	KubeConfig string
	ApiKey     string
	ApiURL     string
	Debug      bool
}

func Init() *Config {
	if config == nil {
		config = &Config{}
		config.InitFlags()
	}
	return config
}

func (c *Config) InitFlags() {
	// 如果有其他类似的引用，请参考下面的方式进行整合
	// 初始化klog
	klog.InitFlags(nil)
	// 将Go的flag绑定到pflag
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	// 环境变量绑定
	defaultPort := 3618
	if envPort := os.Getenv("PORT"); envPort != "" {
		defaultPort, _ = strconv.Atoi(envPort)
	}
	defaultKubeConfig := os.Getenv("KUBECONFIG")
	if defaultKubeConfig == "" {
		defaultKubeConfig = filepath.Join(homedir.HomeDir(), ".kube", "config")
	}

	defaultApiKey := os.Getenv("OPENAI_API_KEY")
	defaultApiURL := os.Getenv("OPENAI_API_URL")
	defaultDebug := false
	if debug := os.Getenv("GIN_MODE"); debug == "release" {
		// GIN_MODE=release
		// 关闭debug模式
		defaultDebug = false
	}

	pflag.BoolVarP(&c.Debug, "debug", "d", defaultDebug, "Debug mode,same as GIN_MODE")
	pflag.IntVarP(&c.Port, "port", "p", defaultPort, "Port for the server to listen on")
	pflag.StringVarP(&c.ApiKey, "chatgpt-key", "k", defaultApiKey, "API Key for ChatGPT")
	pflag.StringVarP(&c.ApiURL, "chatgpt-url", "u", defaultApiURL, "API URL for ChatGPT")
	pflag.StringVarP(&c.KubeConfig, "kubeconfig", "c", defaultKubeConfig, "Absolute path to the kubeConfig file")

	// 检查是否设置了 --v 参数
	if vFlag := pflag.Lookup("v"); vFlag == nil || vFlag.Value.String() == "0" {
		// 如果没有设置，手动将 --v 设置为 2
		_ = flag.Set("v", "2")
	}
	pflag.Parse()

}
