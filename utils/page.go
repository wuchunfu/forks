package utils

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseGitHubNum(num_str string) (int, error) {
	num_str = strings.TrimSpace(num_str)
	if num_str == "" {
		return 0, fmt.Errorf("stars string is empty")
	}

	// 检查是否有后缀
	suffix := num_str[len(num_str)-1]
	numStr := num_str[:len(num_str)-1]

	var multiplier int
	switch suffix {
	case 'k', 'K':
		multiplier = 1000
	case 'm', 'M':
		multiplier = 1000000
	default:
		multiplier = 1
		numStr = num_str
	}

	// 尝试将数字部分转换为浮点数
	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number format: %s", num_str)
	}

	// 计算最终的星星数量
	stars := int(num * float64(multiplier))
	return stars, nil
}
