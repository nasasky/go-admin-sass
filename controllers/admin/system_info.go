package admin

import (
	"nasa-go-admin/inout"
	"nasa-go-admin/services/admin_service"

	"github.com/gin-gonic/gin"
)

var systemInfoService = &admin_service.SystemInfoService{}

// AddSystemInfo 添加系统信息
func AddSystemInfo(c *gin.Context) {
	var params inout.SystemInfoReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	err := systemInfoService.AddSystemInfo(c, params)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	Resp.Succ(c, nil)
}

// UpdateSystemInfo 更新系统信息
func UpdateSystemInfo(c *gin.Context) {
	var params inout.UpdateSystemInfoReq
	if err := c.ShouldBindJSON(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	err := systemInfoService.UpdateSystemInfo(c, params)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	Resp.Succ(c, nil)
}

// GetSystemInfo 获取当前启用的系统信息
func GetSystemInfo(c *gin.Context) {
	info, err := systemInfoService.GetSystemInfo(c)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	Resp.Succ(c, info)
}

// GetSystemInfoList 获取系统信息列表
func GetSystemInfoList(c *gin.Context) {
	var params inout.GetSystemInfoListReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	list, err := systemInfoService.GetSystemInfoList(c, params)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	Resp.Succ(c, list)
}

// GetFirstSystemInfo 获取第一条系统信息记录（公开接口，不需要验证）
func GetFirstSystemInfo(c *gin.Context) {
	info, err := systemInfoService.GetFirstSystemInfo()
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	Resp.Succ(c, info)
}
