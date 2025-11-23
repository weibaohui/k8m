package utils

import (
	"encoding/json"

	"k8s.io/klog/v2"
)

// ToJSON 将任意结构体转换为格式化的 JSON 字符串
func ToJSON(v any) string {
	jsonData, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		klog.V(6).Infof("Error converting to JSON: %v", err)
	}
	return string(jsonData)
}

// ToJSONCompact 将任意结构体转换为紧凑的 JSON 字符串
// 用于机器处理场景，输出不进行格式美化，尽可能减少字符数。
func ToJSONCompact(v any) string {
	jsonData, err := json.Marshal(v)
	if err != nil {
		klog.V(6).Infof("转换为紧凑JSON失败: %v", err)
	}
	return string(jsonData)
}

// DeepCopy 函数接受任何类型的参数并返回其深复制的副本
func DeepCopy[T any](src T) (T, error) {
	var dst T
	data, err := json.Marshal(src)
	if err != nil {
		return dst, err
	}
	err = json.Unmarshal(data, &dst)
	return dst, err
}
