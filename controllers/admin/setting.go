package admin

import (
	"fmt"
	"nasa-go-admin/inout"
	"nasa-go-admin/services/admin_service"
	"nasa-go-admin/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

var SettingService = &admin_service.SettingService{}

// GetSetting
// @Summary 获取系统参数配置列表
func GetSetting(c *gin.Context) {

	var params inout.GetArticleListReq

	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	data, err := SettingService.GetSetting(c, params)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	Resp.Succ(c, data)

}

// GetSettingDetail
// @Summary 获取系统参数配置详情
func GetSettingDetail(c *gin.Context) {

	idStr := c.Query("id")
	if idStr == "" {
		utils.Err(c, utils.ErrCodeInvalidParams, utils.NewError("id不能为空"))
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Err(c, utils.ErrCodeInvalidParams, err)
		return
	}

	data, err := SettingService.GetSettingDetail(c, id)
	if err != nil {
		utils.Err(c, utils.ErrCodeInternalError, err)
		return
	}

	Resp.Succ(c, data)

}

// UpdateSetting
func UpdateSetting(c *gin.Context) {
	var params inout.UpdateSettingReq
	if err := c.ShouldBindJSON(&params); err != nil {
		utils.Err(c, utils.ErrCodeInvalidParams, err)
		return
	}
	data, err := SettingService.UpdateSetting(c, params)
	if err != nil {
		utils.Err(c, utils.ErrCodeInternalError, err)
		return
	}
	Resp.Succ(c, data)
}

// AddSetting
func AddSetting(c *gin.Context) {
	var params inout.SettingReq
	if err := c.ShouldBind(&params); err != nil {
		utils.Err(c, utils.ErrCodeInvalidParams, err)
		return
	}
	data, err := SettingService.AddSetting(c, params)
	if err != nil {
		utils.Err(c, utils.ErrCodeInternalError, err)
		return
	}
	Resp.Succ(c, data)
}

// DeleteSetting
func DeleteSetting(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		utils.Err(c, utils.ErrCodeInvalidParams, utils.NewError("id不能为空"))
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Err(c, utils.ErrCodeInvalidParams, err)
		return
	}

	err = SettingService.DeleteSetting(c, id)
	if err != nil {
		utils.Err(c, utils.ErrCodeInternalError, err)
		return
	}

	Resp.Succ(c, nil)

}

// GetQueueLogList
// @Summary 获取队列消息日志记录列表
func GetQueueLogList(c *gin.Context) {
	// 获取分页参数
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	// 转换为整数
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	// 调用服务
	settingService := admin_service.SettingService{}
	result, err := settingService.GetQueueLogList(page, pageSize)

	if err != nil {
		c.JSON(200, gin.H{
			"code": 50001,
			"msg":  fmt.Sprintf("获取队列状态失败: %v", err),
		})
		return
	}

	c.JSON(200, gin.H{
		"code": 0,
		"msg":  "成功",
		"data": result,
	})
}
