package utils

import (
	"regexp"
	"strings"
)

// GitURLParser 包含解析 Git URL 的方法
type GitURLParser struct{}

// ValidateGitURL 校验是否是符合 Git 协议的 URL
func (p *GitURLParser) ValidateGitURL(gitURL string) bool {
	// 匹配 http(s) 协议的 Git URL 或 git@ 协议的 Git URL
	regex := `^(https?://|git@[\w\.]+:).*\.git$`
	match, _ := regexp.MatchString(regex, gitURL)
	return match
}

// ExtractRepoName 提取最后的 repo 名称
func (p *GitURLParser) ExtractRepoName(gitURL string) string {
	if strings.HasPrefix(gitURL, "http") || strings.HasPrefix(gitURL, "https") {
		// 对 http 和 https 协议的处理
		parts := strings.Split(gitURL, "/")
		return strings.TrimSuffix(parts[len(parts)-1], ".git")
	} else if strings.HasPrefix(gitURL, "git@") {
		// 对 git 协议的处理
		parts := strings.Split(gitURL, ":")
		repoParts := strings.Split(parts[len(parts)-1], "/")
		return strings.TrimSuffix(repoParts[len(repoParts)-1], ".git")
	}
	return ""
}

// ExtractProjectID 提取前面的 project ID
func (p *GitURLParser) ExtractProjectID(gitURL string) string {
	if strings.HasPrefix(gitURL, "http") || strings.HasPrefix(gitURL, "https") {
		// 对 http 和 https 协议的处理
		parts := strings.Split(gitURL, "/")
		return parts[len(parts)-2]
	} else if strings.HasPrefix(gitURL, "git@") {
		// 对 git 协议的处理
		parts := strings.Split(gitURL, ":")
		projectParts := strings.Split(parts[len(parts)-1], "/")
		return projectParts[0]
	}
	return ""
}
