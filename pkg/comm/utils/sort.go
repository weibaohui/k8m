package utils

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func SortByLastTimestamp(items []unstructured.Unstructured) []unstructured.Unstructured {
	sort.Slice(items, func(i, j int) bool {
		tsI, foundI, _ := unstructured.NestedString(items[i].Object, "lastTimestamp")
		tsJ, foundJ, _ := unstructured.NestedString(items[j].Object, "lastTimestamp")

		if foundI && foundJ {
			timeI, errI := time.Parse(time.RFC3339, tsI)
			timeJ, errJ := time.Parse(time.RFC3339, tsJ)
			if errI == nil && errJ == nil {
				return timeI.After(timeJ) // 倒序排列
			}
		}
		return foundI // 如果 I 有，J 没有，I 排在前面
	})
	return items
}

// ParseVersion 函数用于将版本号字符串解析为数字切片
func ParseVersion(version string) []int {
	parts := strings.Split(version, ".")
	result := make([]int, len(parts))
	for i, part := range parts {
		num, _ := strconv.Atoi(part)
		result[i] = num
	}
	return result
}

// CompareVersions 函数用于比较两个版本号的大小
// Example:
//
//	sort.Slice(versions, func(i, j int) bool {
//			return utils.CompareVersions(versions[i], versions[j])
//		})
func CompareVersions(v1, v2 string) bool {
	parts1 := ParseVersion(v1)
	parts2 := ParseVersion(v2)
	for i := 0; i < len(parts1) && i < len(parts2); i++ {
		if parts1[i] > parts2[i] {
			return true
		} else if parts1[i] < parts2[i] {
			return false
		}
	}
	return len(parts1) > len(parts2)
}
