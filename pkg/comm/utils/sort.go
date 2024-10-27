package utils

import (
	"sort"
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
