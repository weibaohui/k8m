package lease

import (
    "fmt"
    "os"

    "github.com/weibaohui/k8m/pkg/comm/utils"
)

// GenerateInstanceID 中文函数注释：生成当前实例的唯一身份标识，规则为 hostname-随机3位。
func GenerateInstanceID() string {
    id, _ := os.Hostname()
    return fmt.Sprintf("%s-%s", id, utils.RandNLengthString(3))
}

