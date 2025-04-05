package utils

// RemoveEmptyLines 删除字符串切片中的空行
func RemoveEmptyLines(lines []string) []string {
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		if line != "" {
			result = append(result, line)
		}
	}
	return result
}

// AllIn 判断A数组中的元素，是否都在B数组中
// a := []string{"apple", "banana"}
// b := []string{"apple", "banana", "cherry"}
//
// fmt.Println(AllIn(a, b)) // true
//
// a2 := []string{"apple", "kiwi"}
// fmt.Println(AllIn(a2, b)) // false
func AllIn(a, b []string) bool {
	bSet := make(map[string]struct{}, len(b))
	for _, item := range b {
		bSet[item] = struct{}{}
	}

	for _, item := range a {
		if _, ok := bSet[item]; !ok {
			return false
		}
	}
	return true
}
