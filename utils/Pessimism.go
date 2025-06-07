package utils

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ParsePessimism 解析 Pessimism 字段
func ParsePessimism(c *gin.Context, pessimismStr string) ([]int, error) {
	var pessimism []int
	if pessimismStr != "" {
		pessimismItems := strings.Split(pessimismStr, ",")
		for _, p := range pessimismItems {
			pInt, err := strconv.Atoi(p)
			if err != nil {
				return nil, err
			}
			pessimism = append(pessimism, pInt)
		}
	}
	return pessimism, nil
}
