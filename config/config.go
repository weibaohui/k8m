package config

import (
	"flag"
	"path/filepath"

	"github.com/spf13/pflag"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
)

var config *Config

type Config struct {
	Port       int
	Kubeconfig string
	APIKey     string
	APIURL     string
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
	// todo 获取默认值，从ENV中获取
	// 初始化klog
	klog.InitFlags(nil)
	// 将Go的flag绑定到pflag
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	pflag.BoolVarP(&c.Debug, "debug", "d", false, "Debug mode")
	pflag.IntVarP(&c.Port, "port", "p", 3618, "Port for the server to listen on")
	pflag.StringVarP(&c.APIKey, "chatgpt-key", "k", "", "API Key for ChatGPT")
	pflag.StringVarP(&c.APIURL, "chatgpt-url", "u", "", "API URL for ChatGPT")
	pflag.StringVarP(&c.Kubeconfig, "kubeconfig", "c", filepath.Join(homedir.HomeDir(), ".kube", "config"), "Absolute path to the kubeConfig file")

	// 检查是否设置了 --v 参数
	if vFlag := pflag.Lookup("v"); vFlag == nil || vFlag.Value.String() == "0" {
		// 如果没有设置，手动将 --v 设置为 2
		_ = flag.Set("v", "2")
	}
	pflag.Parse()

}
