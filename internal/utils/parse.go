package utils

import (
	"strconv"
	"strings"
)

func ParseCommaSeparatedToMap[T int | int64](input string) map[T]bool {
	resultMap := make(map[T]bool)
	if input == "" {
		return resultMap
	}

	parts := strings.Split(input, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if val, err := strconv.Atoi(part); err == nil {
			resultMap[T(val)] = true
		}
	}
	return resultMap
}
