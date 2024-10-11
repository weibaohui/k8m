package utils

import (
	"testing"
)

// 测试 ExtractRepoName 方法
func TestExtractRepoName(t *testing.T) {
	parser := &GitURLParser{}

	tests := []struct {
		url      string
		expected string
	}{
		{"http://gitlab.sd.devops.cmit.cloud:32766/TYSFRZGLPT.TYYHGLPT/uum-new-server.git", "uum-new-server"},
		{"https://github.com/owner/repo.git", "repo"},
		{"git@gitcode.com:Cangjie/CangjieCommunity.git", "CangjieCommunity"},
	}

	for _, tt := range tests {
		result := parser.ExtractRepoName(tt.url)
		if result != tt.expected {
			t.Errorf("ExtractRepoName(%s) = %s; want %s", tt.url, result, tt.expected)
		}
	}
}

// 测试 ExtractProjectID 方法
func TestExtractProjectID(t *testing.T) {
	parser := &GitURLParser{}

	tests := []struct {
		url      string
		expected string
	}{
		{"http://gitlab.sd.devops.cmit.cloud:32766/TYSFRZGLPT.TYYHGLPT/uum-new-server.git", "TYSFRZGLPT.TYYHGLPT"},
		{"https://github.com/owner/repo.git", "owner"},
		{"git@gitcode.com:Cangjie/CangjieCommunity.git", "Cangjie"},
	}

	for _, tt := range tests {
		result := parser.ExtractProjectID(tt.url)
		if result != tt.expected {
			t.Errorf("ExtractProjectID(%s) = %s; want %s", tt.url, result, tt.expected)
		}
	}
}
