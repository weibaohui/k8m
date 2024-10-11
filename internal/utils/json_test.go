package utils

import (
	"testing"
)

// TestToJSON 测试 ToJSON 方法
func TestToJSON(t *testing.T) {
	// 实例化 JSONUtils
	jsonUtils := &JSONUtils{}

	// 测试数据结构体
	type CodeRepoInfo struct {
		CodeRepoURL       string `json:"code_repo_url"`
		CodeBranch        string `json:"code_branch"`
		CodeLanguage      string `json:"code_language"`
		BuildTools        string `json:"build_tools"`
		CiResultImageRepo string `json:"ci_result_image_repo"`
	}

	data := CodeRepoInfo{
		CodeRepoURL:       "http://gitee.com/zhangsan/server.git",
		CodeBranch:        "master",
		CodeLanguage:      "Java",
		BuildTools:        "maven:3.9-jdk8",
		CiResultImageRepo: "uum_images",
	}

	// 期望的 JSON 输出
	expected := `{
  "code_repo_url": "http://gitee.com/zhangsan/server.git",
  "code_branch": "master",
  "code_language": "Java",
  "build_tools": "maven:3.9-jdk8",
  "ci_result_image_repo": "uum_images"
}`

	// 调用 ToJSON 方法
	jsonString := jsonUtils.ToJSON(data)

	// 校验输出是否符合期望
	if jsonString != expected {
		t.Errorf("ToJSON() = %s; want %s", jsonString, expected)
	}
}
