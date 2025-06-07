package utils

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetParentId 从 gin.Context 中解析用户信息并获取 parent_id
func GetParentId(c *gin.Context) (int, error) {
	// 获取用户信息
	userInfo, exists := c.Get("userInfo")
	userId := c.GetInt("uid") // 获取用户ID
	//fmt.Println("userInfo:", userInfo)
	if !exists {
		return 0, fmt.Errorf("用户信息缺失")
	}

	// 类型断言为 map[string]string
	userInfoMap, ok := userInfo.(map[string]string)
	if !ok {
		return 0, fmt.Errorf("用户信息格式错误")
	}

	// 从 map 中获取 parent_id
	parentIdStr, ok := userInfoMap["parentId"]
	//fmt.Println("parentIdStr:", parentIdStr)
	if !ok || parentIdStr == "" || parentIdStr == "0" {
		// 如果 parentId 不存在、为空或等于 "0"，赋值为 userId
		return userId, nil
	}

	// 将 parent_id 转换为整数
	parentId, err := strconv.Atoi(parentIdStr)
	if err != nil {
		return 0, fmt.Errorf("invalid parent_id: %s", parentIdStr)
	}

	return parentId, nil
}
