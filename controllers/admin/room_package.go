package admin

import (
	"net/http"
	"strconv"
	"time"

	"nasa-go-admin/inout"
	"nasa-go-admin/pkg/response"
	"nasa-go-admin/services/app_service"

	"github.com/gin-gonic/gin"
)

type RoomPackageController struct {
	roomService *app_service.RoomService
}

func NewRoomPackageController() *RoomPackageController {
	return &RoomPackageController{
		roomService: &app_service.RoomService{},
	}
}

// CreatePackage 创建套餐
func (rpc *RoomPackageController) CreatePackage(c *gin.Context) {
	var req inout.CreatePackageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	// 获取当前用户ID (从JWT中)
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "未授权", "无法获取用户信息")
		return
	}

	pkg, err := rpc.roomService.CreatePackage(&req, userID.(int))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "创建套餐失败", err.Error())
		return
	}

	response.Success(c, gin.H{
		"data":    pkg,
		"message": "创建套餐成功",
	})
}

// UpdatePackage 更新套餐
func (rpc *RoomPackageController) UpdatePackage(c *gin.Context) {
	var req inout.UpdatePackageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	pkg, err := rpc.roomService.UpdatePackage(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "更新套餐失败", err.Error())
		return
	}

	response.Success(c, gin.H{
		"data":    pkg,
		"message": "更新套餐成功",
	})
}

// GetPackageList 获取套餐列表
func (rpc *RoomPackageController) GetPackageList(c *gin.Context) {
	var req inout.PackageListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	result, err := rpc.roomService.GetPackageList(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取套餐列表失败", err.Error())
		return
	}

	response.Success(c, result)
}

// DeletePackage 删除套餐
func (rpc *RoomPackageController) DeletePackage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", "套餐ID必须是数字")
		return
	}

	if err := rpc.roomService.DeletePackage(id); err != nil {
		response.Error(c, http.StatusInternalServerError, "删除套餐失败", err.Error())
		return
	}

	response.Success(c, gin.H{
		"message": "删除套餐成功",
	})
}

// CreatePackageRule 创建套餐规则
func (rpc *RoomPackageController) CreatePackageRule(c *gin.Context) {
	var req inout.CreatePackageRuleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	rule, err := rpc.roomService.CreatePackageRule(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "创建套餐规则失败", err.Error())
		return
	}

	response.Success(c, gin.H{
		"data":    rule,
		"message": "创建套餐规则成功",
	})
}

// UpdatePackageRule 更新套餐规则
func (rpc *RoomPackageController) UpdatePackageRule(c *gin.Context) {
	var req inout.UpdatePackageRuleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	rule, err := rpc.roomService.UpdatePackageRule(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "更新套餐规则失败", err.Error())
		return
	}

	response.Success(c, gin.H{
		"data":    rule,
		"message": "更新套餐规则成功",
	})
}

// GetPackageRuleList 获取套餐规则列表
func (rpc *RoomPackageController) GetPackageRuleList(c *gin.Context) {
	packageIDStr := c.Query("package_id")
	if packageIDStr == "" {
		response.Error(c, http.StatusBadRequest, "参数错误", "套餐ID不能为空")
		return
	}

	packageID, err := strconv.Atoi(packageIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", "套餐ID必须是数字")
		return
	}

	rules, err := rpc.roomService.GetPackageRuleList(packageID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取套餐规则列表失败", err.Error())
		return
	}

	response.Success(c, rules)
}

// DeletePackageRule 删除套餐规则
func (rpc *RoomPackageController) DeletePackageRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", "规则ID必须是数字")
		return
	}

	if err := rpc.roomService.DeletePackageRule(id); err != nil {
		response.Error(c, http.StatusInternalServerError, "删除套餐规则失败", err.Error())
		return
	}

	response.Success(c, gin.H{
		"message": "删除套餐规则成功",
	})
}

// GetSpecialDateList 获取特殊日期列表
func (rpc *RoomPackageController) GetSpecialDateList(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	dates, err := rpc.roomService.GetSpecialDateList(page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取特殊日期列表失败", err.Error())
		return
	}

	response.Success(c, dates)
}

// CreateSpecialDate 创建特殊日期
func (rpc *RoomPackageController) CreateSpecialDate(c *gin.Context) {
	var req struct {
		Date        string `json:"date" binding:"required"`
		DateType    string `json:"date_type" binding:"required,oneof=holiday festival special"`
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	// 解析日期
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", "日期格式错误，应为YYYY-MM-DD")
		return
	}

	specialDate, err := rpc.roomService.CreateSpecialDate(date, req.DateType, req.Name, req.Description)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "创建特殊日期失败", err.Error())
		return
	}

	response.Success(c, gin.H{
		"data":    specialDate,
		"message": "创建特殊日期成功",
	})
}

// DeleteSpecialDate 删除特殊日期
func (rpc *RoomPackageController) DeleteSpecialDate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", "特殊日期ID必须是数字")
		return
	}

	if err := rpc.roomService.DeleteSpecialDate(id); err != nil {
		response.Error(c, http.StatusInternalServerError, "删除特殊日期失败", err.Error())
		return
	}

	response.Success(c, gin.H{
		"message": "删除特殊日期成功",
	})
}

// CalculatePrice 价格计算接口
func (rpc *RoomPackageController) CalculatePrice(c *gin.Context) {
	var req struct {
		RoomID    int    `json:"room_id" binding:"required"`
		StartTime string `json:"start_time" binding:"required"`
		Hours     int    `json:"hours" binding:"required,min=1,max=24"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	// 解析开始时间
	startTime, err := time.Parse("2006-01-02 15:04:05", req.StartTime)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", "开始时间格式错误")
		return
	}

	endTime := startTime.Add(time.Duration(req.Hours) * time.Hour)

	// 计算价格
	finalPrice, appliedRules, err := rpc.roomService.CalculatePriceWithPackage(req.RoomID, startTime, endTime)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "价格计算失败", err.Error())
		return
	}

	result := gin.H{
		"room_id":       req.RoomID,
		"start_time":    req.StartTime,
		"hours":         req.Hours,
		"final_price":   finalPrice,
		"applied_rules": appliedRules,
	}

	response.Success(c, result)
}

// ========== 用户端接口 ==========

// GetRoomPackages 获取房间可用套餐
func (rpc *RoomPackageController) GetRoomPackages(c *gin.Context) {
	var req inout.GetRoomPackagesReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	resp, err := rpc.roomService.GetRoomPackages(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取套餐失败", err.Error())
		return
	}

	response.Success(c, resp)
}

// BookingPricePreview 预订价格预览
func (rpc *RoomPackageController) BookingPricePreview(c *gin.Context) {
	var req inout.BookingPricePreviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	resp, err := rpc.roomService.BookingPricePreview(&req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "价格计算失败", err.Error())
		return
	}

	response.Success(c, resp)
}
