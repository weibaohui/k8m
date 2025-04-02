package utils

import (
	"strings"
)

// UpdateImageName
// 检查镜像名称是否包含前缀，只能是harbor.power.sd.k9s.space
// 是否已harbor.power.sd.k9s.space开头，是，没问题。
// 否，检查是否有其他前缀，删除，替换为harbor.power.sd.k9s.space
// 没有，增加harbor.power.sd.k9s.space前缀
// 检查镜像名称是否已以指定前缀开头
func UpdateImageName(imageName string, imagePrefix string) string {

	// 如果镜像名称已以指定前缀开头，直接返回
	if strings.HasPrefix(imageName, imagePrefix) {
		return imageName
	}

	// 检查是否是 docker.io/library 前缀的特殊情况
	if strings.HasPrefix(imageName, "docker.io/library/") {
		// 替换为指定前缀
		return imagePrefix + imageName[len("docker.io/library"):]
	}

	// 拆分镜像名称路径，查找第一个斜杠位置
	slashIndex := strings.Index(imageName, "/")
	if slashIndex == -1 {
		// 没有前缀，直接添加指定的前缀
		imageName = imagePrefix + "/" + imageName
		return imageName
	}

	// 分析并确保路径中保留多层级结构
	parts := strings.Split(imageName, "/")
	if len(parts) > 1 && !strings.Contains(parts[0], ".") {
		// 如果第一部分不是域名，则添加前缀
		imageName = imagePrefix + "/" + imageName
	} else {
		// 如果是域名或没有多层路径，直接附加前缀
		imageName = imagePrefix + "/" + strings.Join(parts, "/")
	}
	imageName = strings.TrimSpace(imageName)
	return imageName
}

// 获取镜像名称及Tag
func GetImageNameAndTag(imageName string) (string, string) {
	// 拆分镜像名称路径，查找最后一个冒号位置
	//harbor.sdibt.com:5000/public/nginx:1.02

	colonIndex := strings.LastIndex(imageName, ":")
	if colonIndex == -1 {
		// 没有冒号，默认为无Tag
		return imageName, "latest"
	}

	// 获取镜像名称和Tag
	imageNameWithoutTag := imageName[:colonIndex]
	tag := imageName[colonIndex+1:]
	return imageNameWithoutTag, tag
}
