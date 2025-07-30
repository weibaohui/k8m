package dynamic

import (
	"fmt"
)

type yamlRequest struct {
	Yaml string `json:"yaml" binding:"required"`
}

// 返回资源类型对应的路径
func getResourcePaths(kind string) ([]string, error) {
	switch kind {
	case "CloneSet":
		return []string{"spec", "template", "spec"}, nil
	case "Deployment", "DaemonSet", "StatefulSet", "ReplicaSet", "Job":
		return []string{"spec", "template", "spec"}, nil
	case "CronJob":
		return []string{"spec", "jobTemplate", "spec", "template", "spec"}, nil
	case "Pod":
		return []string{"spec"}, nil
	default:
		return nil, fmt.Errorf("unsupported resource kind: %s", kind)
	}
}
