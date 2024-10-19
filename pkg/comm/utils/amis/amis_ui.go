package amis

import (
	"fmt"
)

// Option 定义选项结构体
type Option struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

func ArrayToOptions[T any](array []T) []Option {
	options := make([]Option, len(array))

	for i, item := range array {
		str := fmt.Sprintf("%v", item)
		options[i] = Option{Label: str, Value: str}
	}
	return options
}
func MapToOptions[T any](m map[string]T) []Option {
	var options []Option
	for k, v := range m {
		options = append(options, Option{
			Label: k,
			Value: fmt.Sprintf("%v", v),
		})
	}
	return options
}
