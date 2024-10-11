package utils

import (
	"encoding/json"
	"log"
)

// JSONUtils 提供通用的 JSON 操作方法
type JSONUtils struct{}

// ToJSON 将任意结构体转换为格式化的 JSON 字符串
func (j *JSONUtils) ToJSON(v interface{}) string {
	jsonData, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Fatalf("Error converting to JSON: %v", err)
	}
	return string(jsonData)
}
