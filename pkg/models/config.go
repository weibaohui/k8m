package models

import (
	"time"
)

type Config struct {
	ID                uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	Port              int       `json:"port,omitempty"`
	MCPServerPort     int       `json:"mcp_server_port,omitempty"`
	KubeConfig        string    `json:"kube_config,omitempty"`
	ApiKey            string    `json:"api_key,omitempty"`
	ApiURL            string    `json:"api_url,omitempty"`
	ApiModel          string    `json:"api_model,omitempty"`
	Debug             bool      `json:"debug,omitempty"`
	LogV              int       `json:"log_v,omitempty"`
	InCluster         bool      `json:"in_cluster,omitempty"`
	LoginType         string    `json:"login_type,omitempty"`
	JwtTokenSecret    string    `json:"jwt_token_secret,omitempty"`
	NodeShellImage    string    `json:"node_shell_image,omitempty"`
	KubectlShellImage string    `json:"kubectl_shell_image,omitempty"`
	SqlitePath        string    `json:"sqlite_path,omitempty"`
	AnySelect         bool      `json:"any_select,omitempty"`
	PrintConfig       bool      `json:"print_config,omitempty"`
	EnableAI          bool      `json:"enable_ai,omitempty"`       // 是否启用AI功能，默认开启
	ConnectCluster    bool      `json:"connect_cluster,omitempty"` // 启动集群是是否自动连接现有集群，默认关闭
	CreatedAt         time.Time `json:"created_at,omitempty"`      // Automatically managed by GORM for creation time
	UpdatedAt         time.Time `json:"updated_at,omitempty"`      // Automatically managed by GORM for update time
}
