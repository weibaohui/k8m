package models

import (
	"time"
)

type Config struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	// Port              int       `json:"port,omitempty"`
	// MCPServerPort     int       `json:"mcp_server_port,omitempty"`
	// KubeConfig        string    `json:"kube_config,omitempty"`
	ApiKey   string `json:"api_key,omitempty"`
	ApiURL   string `json:"api_url,omitempty"`
	ApiModel string `json:"api_model,omitempty"`
	// Debug             bool      `gorm:"default:false" json:"debug"`
	// LogV              int       `json:"log_v,omitempty"`
	// InCluster         bool      `gorm:"default:true" json:"in_cluster"`
	LoginType         string `json:"login_type,omitempty"`
	JwtTokenSecret    string `json:"jwt_token_secret,omitempty"`
	NodeShellImage    string `json:"node_shell_image,omitempty"`
	KubectlShellImage string `json:"kubectl_shell_image,omitempty"`
	ImagePullTimeout  int    `gorm:"default:30" json:"image_pull_timeout,omitempty"` // 镜像拉取超时时间（秒）
	// SqlitePath        string    `json:"sqlite_path,omitempty"`
	AnySelect       bool `gorm:"default:true" json:"any_select"`
	PrintConfig     bool `json:"print_config"`
	EnableAI        bool `gorm:"default:true" json:"enable_ai"` // 是否启用AI功能，默认开启
	UseBuiltInModel bool `gorm:"default:true" json:"use_built_in_model"`
	// ConnectCluster  bool      `json:"connect_cluster"`      // 启动集群是是否自动连接现有集群，默认关闭
	CreatedAt time.Time `json:"created_at,omitempty"` // Automatically managed by GORM for creation time
	UpdatedAt time.Time `json:"updated_at,omitempty"` // Automatically managed by GORM for update time
}
