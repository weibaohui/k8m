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

// AnyIn 判断A数组中的元素，是否有任意一个存在于B数组中
// a := []string{"apple", "kiwi"}
// b := []string{"apple", "banana", "cherry"}
//
// fmt.Println(AnyIn(a, b)) // true，因为"apple"在b中
//
// a2 := []string{"grape", "kiwi"}
// fmt.Println(AnyIn(a2, b)) // false，因为没有元素在b中
func AnyIn(a, b []string) bool {
	bSet := make(map[string]struct{}, len(b))
	for _, item := range b {
		bSet[item] = struct{}{}
	}

	for _, item := range a {
		if _, ok := bSet[item]; ok {
			return true
		}
	}
	return false
}
