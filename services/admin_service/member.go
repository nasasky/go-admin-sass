package admin_service

import (
	"nasa-go-admin/db"
	"nasa-go-admin/inout"
	"nasa-go-admin/model/admin_model"

	"github.com/gin-gonic/gin"
)

type MemberService struct{}

func (s *MemberService) GetMemberList(c *gin.Context, params inout.ListpageReq) (interface{}, error) {
	var data []admin_model.Member
	var total int64
	// 设置默认分页参数
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}
	// 计算偏移量
	offset := (params.Page - 1) * params.PageSize

	// 初始化查询构建器
	query := db.Dao.Model(&admin_model.Member{})

	// 添加搜索条件
	if params.Search != "" {
		// 使用 LIKE 进行模糊搜索 - 用户名和手机号
		searchTerm := "%" + params.Search + "%"
		query = query.Where("username LIKE ? OR phone LIKE ? OR nick_name LIKE ?",
			searchTerm, searchTerm, searchTerm)
	}

	// 查询总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, err
	}

	// 查询分页数据
	err = query.Offset(offset).Limit(params.PageSize).Find(&data).Error
	if err != nil {
		return nil, err
	}
	// 格式化时间字段

	//格式化数据
	formattedData := formMemberData(data)
	// 构建响应
	response := inout.ListMemberpageResp{
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
		Items:    formattedData,
	}
	return response, nil
}

func formMemberData(data []admin_model.Member) []inout.MemberListItem {
	formattedData := make([]inout.MemberListItem, len(data))
	for i, item := range data {
		formattedData[i] = inout.MemberListItem{
			Id:         item.Id,
			UserName:   item.UserName,
			NickName:   item.NickName,
			Avatar:     item.Avatar,
			Phone:      item.Phone,
			CreateTime: formatTime(item.CreateTime),
			UpdateTime: formatTime(item.CreateTime),
		}
	}
	return formattedData
}
