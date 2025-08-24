package utils

import (
	"math"
	"testing"
)

func TestRandIntOverflow(t *testing.T) {
	// 测试接近 int 最大值的情况，确保不会溢出
	tests := []struct {
		name string
		min  int
		max  int
	}{
		{
			name: "正常范围",
			min:  1,
			max:  100,
		},
		{
			name: "大数值范围",
			min:  math.MaxInt32 - 1000,
			max:  math.MaxInt32,
		},
		{
			name: "负数范围",
			min:  -100,
			max:  -1,
		},
		{
			name: "跨零范围",
			min:  -50,
			max:  50,
		},
		{
			name: "相等值",
			min:  42,
			max:  42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RandInt(tt.min, tt.max)

			// 验证结果在预期范围内
			if result < tt.min || result > tt.max {
				t.Errorf("RandInt(%d, %d) = %d, 超出范围 [%d, %d]",
					tt.min, tt.max, result, tt.min, tt.max)
			}
		})
	}
}

func TestRandNDigitIntOverflow(t *testing.T) {
	// 测试不同位数的随机数生成
	tests := []struct {
		name   string
		digits int
		minVal int
		maxVal int
	}{
		{
			name:   "1位数",
			digits: 1,
			minVal: 1,
			maxVal: 9,
		},
		{
			name:   "2位数",
			digits: 2,
			minVal: 10,
			maxVal: 99,
		},
		{
			name:   "3位数",
			digits: 3,
			minVal: 100,
			maxVal: 999,
		},
		{
			name:   "5位数",
			digits: 5,
			minVal: 10000,
			maxVal: 99999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RandNDigitInt(tt.digits)

			// 验证结果在预期范围内
			if result < tt.minVal || result > tt.maxVal {
				t.Errorf("RandNDigitInt(%d) = %d, 超出范围 [%d, %d]",
					tt.digits, result, tt.minVal, tt.maxVal)
			}
		})
	}
}

func TestRandNLengthString(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"空字符串", 0},
		{"单字符", 1},
		{"短字符串", 5},
		{"长字符串", 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RandNLengthString(tt.length)

			if len(result) != tt.length {
				t.Errorf("RandNLengthString(%d) 长度 = %d, 期望 %d",
					tt.length, len(result), tt.length)
			}

			// 验证字符串只包含有效字符
			for _, char := range result {
				found := false
				for _, validChar := range charset {
					if char == validChar {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("RandNLengthString(%d) 包含无效字符: %c", tt.length, char)
				}
			}
		})
	}
}

// 边界测试
func TestEdgeCases(t *testing.T) {
	// 测试零或负数输入
	t.Run("RandNDigitInt零输入", func(t *testing.T) {
		result := RandNDigitInt(0)
		if result != 0 {
			t.Errorf("RandNDigitInt(0) = %d, 期望 0", result)
		}
	})

	t.Run("RandNDigitInt负数输入", func(t *testing.T) {
		result := RandNDigitInt(-5)
		if result != 0 {
			t.Errorf("RandNDigitInt(-5) = %d, 期望 0", result)
		}
	})

	t.Run("RandInt颠倒的min/max", func(t *testing.T) {
		result := RandInt(100, 50)
		if result < 50 || result > 100 {
			t.Errorf("RandInt(100, 50) = %d, 超出范围 [50, 100]", result)
		}
	})
}
