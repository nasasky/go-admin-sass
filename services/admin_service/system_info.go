package admin_service

import (
	"errors"
	"nasa-go-admin/db"
	"nasa-go-admin/inout"
	"nasa-go-admin/model/admin_model"
	"nasa-go-admin/utils"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SystemInfoService struct{}

// AddSystemInfo 添加系统信息
func (s *SystemInfoService) AddSystemInfo(c *gin.Context, params inout.SystemInfoReq) error {
	parentId, err := utils.GetParentId(c)
	if err != nil {
		return err
	}

	// 使用 idx_tenant_status 索引检查是否已存在启用的系统信息
	var count int64
	err = db.Dao.Model(&admin_model.SystemInfo{}).
		Where("tenants_id = ? AND status = ?", parentId, 1).
		Count(&count).Error
	if err != nil {
		return err
	}

	if count > 0 && params.Status == 1 {
		return errors.New("已存在启用的系统信息，请先禁用其他系统信息")
	}

	systemInfo := admin_model.SystemInfo{
		SystemName:  params.SystemName,
		SystemTitle: params.SystemTitle,
		IcpNumber:   params.IcpNumber,
		Copyright:   params.Copyright,
		Status:      params.Status,
		TenantsId:   parentId,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}

	return db.Dao.Create(&systemInfo).Error
}

// UpdateSystemInfo 更新系统信息
func (s *SystemInfoService) UpdateSystemInfo(c *gin.Context, params inout.UpdateSystemInfoReq) error {
	parentId, err := utils.GetParentId(c)
	if err != nil {
		return err
	}

	// 使用 idx_tenant_status 索引检查其他启用的系统信息
	if params.Status == 1 {
		var count int64
		err = db.Dao.Model(&admin_model.SystemInfo{}).
			Where("tenants_id = ? AND status = ? AND id != ?", parentId, 1, params.Id).
			Count(&count).Error
		if err != nil {
			return err
		}

		if count > 0 {
			return errors.New("已存在启用的系统信息，请先禁用其他系统信息")
		}
	}

	// 直接更新所有字段，允许空字符串更新
	updates := map[string]interface{}{
		"system_name":  params.SystemName,
		"system_title": params.SystemTitle,
		"icp_number":   params.IcpNumber,
		"copyright":    params.Copyright,
		"status":       params.Status,
		"update_time":  time.Now(),
	}

	// 使用主键和租户ID索引进行更新
	result := db.Dao.Model(&admin_model.SystemInfo{}).
		Where("id = ? AND tenants_id = ?", params.Id, parentId).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("未找到要更新的系统信息或无权限更新")
	}

	return nil
}

// GetSystemInfo 获取系统信息
func (s *SystemInfoService) GetSystemInfo(c *gin.Context) (*inout.SystemInfoResponse, error) {
	// 获取第一条系统信息记录
	var systemInfo admin_model.SystemInfo
	err := db.Dao.Order("id ASC").First(&systemInfo).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &inout.SystemInfoResponse{
		Id:          systemInfo.Id,
		SystemName:  systemInfo.SystemName,
		SystemTitle: systemInfo.SystemTitle,
		IcpNumber:   systemInfo.IcpNumber,
		Copyright:   systemInfo.Copyright,
		Status:      systemInfo.Status,
		CreateTime:  systemInfo.CreateTime,
		UpdateTime:  systemInfo.UpdateTime,
	}, nil
}

// GetSystemInfoList 获取系统信息列表
func (s *SystemInfoService) GetSystemInfoList(c *gin.Context, params inout.GetSystemInfoListReq) (*inout.SystemInfoListResponse, error) {
	parentId, err := utils.GetParentId(c)
	if err != nil {
		return nil, err
	}

	// 设置默认分页参数
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}

	query := db.Dao.Model(&admin_model.SystemInfo{}).Where("tenants_id = ?", parentId)

	// 使用 idx_system_name 索引进行模糊搜索
	if params.Search != "" {
		query = query.Where("system_name LIKE ?", "%"+params.Search+"%")
	}

	// 获取总数
	var total int64
	err = query.Count(&total).Error
	if err != nil {
		return nil, err
	}

	// 根据排序字段选择合适的索引
	switch params.OrderBy {
	case "create_time":
		// 使用 idx_tenant_create_time 索引
		query = query.Order("create_time DESC")
	case "update_time":
		// 使用 idx_tenant_update_time 索引
		query = query.Order("update_time DESC")
	default:
		// 默认使用 id 排序
		query = query.Order("id DESC")
	}

	// 计算偏移量
	offset := (params.Page - 1) * params.PageSize

	var systemInfos []admin_model.SystemInfo
	err = query.Offset(offset).Limit(params.PageSize).Find(&systemInfos).Error
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	items := make([]inout.SystemInfoResponse, len(systemInfos))
	for i, info := range systemInfos {
		items[i] = inout.SystemInfoResponse{
			Id:          info.Id,
			SystemName:  info.SystemName,
			SystemTitle: info.SystemTitle,
			IcpNumber:   info.IcpNumber,
			Copyright:   info.Copyright,
			Status:      info.Status,
			CreateTime:  info.CreateTime,
			UpdateTime:  info.UpdateTime,
		}
	}

	return &inout.SystemInfoListResponse{
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
		Items:    items,
	}, nil
}

// GetFirstSystemInfo 获取第一条系统信息记录（不需要验证，公开接口）
func (s *SystemInfoService) GetFirstSystemInfo() (*inout.SystemInfoResponse, error) {
	var systemInfo admin_model.SystemInfo
	err := db.Dao.Model(&admin_model.SystemInfo{}).
		Order("id ASC").
		First(&systemInfo).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &inout.SystemInfoResponse{
		Id:          systemInfo.Id,
		SystemName:  systemInfo.SystemName,
		SystemTitle: systemInfo.SystemTitle,
		IcpNumber:   systemInfo.IcpNumber,
		Copyright:   systemInfo.Copyright,
		Status:      systemInfo.Status,
		CreateTime:  systemInfo.CreateTime,
		UpdateTime:  systemInfo.UpdateTime,
	}, nil
}
