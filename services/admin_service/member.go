package admin_service

import (
	"fmt"
	"nasa-go-admin/db"
	"nasa-go-admin/inout"
	"nasa-go-admin/model/admin_model"
	"nasa-go-admin/utils"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type MemberService struct{}

func (s *MemberService) GetMemberList(c *gin.Context, params inout.GetMemberListReq) (interface{}, error) {
	var data []admin_model.Member
	var total int64
	
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}
	
	offset := (params.Page - 1) * params.PageSize

	query := db.Dao.Model(&admin_model.Member{})

	s.buildSearchConditions(query, params)

	err := query.Count(&total).Error
	if err != nil {
		return nil, err
	}

	err = query.Order("create_time DESC").Offset(offset).Limit(params.PageSize).Find(&data).Error
	if err != nil {
		return nil, err
	}

	formattedData := s.formMemberData(data)
	
	response := inout.ListMemberpageResp{
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
		Items:    formattedData,
	}
	return response, nil
}

// ExportMemberList 导出会员列表为Excel格式（修复版）
func (s *MemberService) ExportMemberList(c *gin.Context, params inout.ExportMemberListReq) ([]byte, string, error) {
	var data []admin_model.Member

	query := db.Dao.Model(&admin_model.Member{})

	s.buildExportSearchConditions(query, params)

	err := query.Order("create_time DESC").Find(&data).Error
	if err != nil {
		return nil, "", err
	}

	// 创建Excel文件
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("关闭Excel文件失败: %v\n", err)
		}
	}()

	// 使用英文工作表名称避免兼容性问题
	sheetName := "Members"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, "", err
	}

	// 删除默认的Sheet1
	err = f.DeleteSheet("Sheet1")
	if err != nil {
		fmt.Printf("删除默认工作表失败: %v\n", err)
	}

	// 设置表头
	headers := []string{"序号", "会员ID", "用户名", "昵称", "手机号", "地址", "注册时间", "更新时间"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	// 简化表头样式
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 11,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"DDDDDD"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return nil, "", err
	}

	// 应用表头样式
	for i := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	// 填充数据并清理内容
	for idx, item := range data {
		row := idx + 2
		
		// 使用数据清理函数处理所有字符串字段
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), idx+1)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), item.Id)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), s.cleanText(item.UserName))
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), s.cleanText(item.NickName))
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), s.cleanText(item.Phone))
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), s.cleanText(item.Address))
		
		// 处理时间字段，避免空时间导致的问题
		createTime := s.formatTimeForExcel(item.CreateTime)
		updateTime := s.formatTimeForExcel(item.UpdateTime)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), createTime)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), updateTime)
	}

	// 设置列宽
	columnWidths := map[string]float64{
		"A": 8,   // 序号
		"B": 12,  // 会员ID
		"C": 15,  // 用户名
		"D": 15,  // 昵称
		"E": 18,  // 手机号
		"F": 25,  // 地址
		"G": 22,  // 注册时间
		"H": 22,  // 更新时间
	}
	
	for col, width := range columnWidths {
		err = f.SetColWidth(sheetName, col, col, width)
		if err != nil {
			return nil, "", err
		}
	}

	// 设置简单的边框（避免复杂样式）
	if len(data) > 0 {
		borderStyle, err := f.NewStyle(&excelize.Style{
			Border: []excelize.Border{
				{Type: "left", Color: "000000", Style: 1},
				{Type: "top", Color: "000000", Style: 1},
				{Type: "bottom", Color: "000000", Style: 1},
				{Type: "right", Color: "000000", Style: 1},
			},
		})
		if err == nil {
			// 只给数据区域设置边框
			for row := 2; row <= len(data)+1; row++ {
				for col := 0; col < len(headers); col++ {
					cell := fmt.Sprintf("%c%d", 'A'+col, row)
					f.SetCellStyle(sheetName, cell, cell, borderStyle)
				}
			}
		}
	}

	// 冻结首行
	err = f.SetPanes(sheetName, &excelize.Panes{
		Freeze: true,
		Split:  false,
		XSplit: 0,
		YSplit: 1,
	})
	if err != nil {
		fmt.Printf("设置冻结行失败: %v\n", err)
	}

	// 设置活动工作表
	f.SetActiveSheet(index)

	// 保存到字节数组
	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, "", err
	}

	// 生成文件名（使用英文）
	filename := fmt.Sprintf("members_%s.xlsx", time.Now().Format("20060102_150405"))

	return buffer.Bytes(), filename, nil
}

// cleanText 清理文本内容，移除可能导致Excel问题的字符
func (s *MemberService) cleanText(text string) string {
	if text == "" {
		return ""
	}
	
	// 移除控制字符和不可见字符
	reg := regexp.MustCompile(`[\x00-\x1F\x7F]`)
	cleaned := reg.ReplaceAllString(text, "")
	
	// 替换可能有问题的字符
	cleaned = strings.ReplaceAll(cleaned, "\r\n", " ")
	cleaned = strings.ReplaceAll(cleaned, "\n", " ")
	cleaned = strings.ReplaceAll(cleaned, "\r", " ")
	cleaned = strings.ReplaceAll(cleaned, "\t", " ")
	
	// 去除前后空白
	cleaned = strings.TrimSpace(cleaned)
	
	// 限制长度避免单元格内容过长
	if len(cleaned) > 32767 { // Excel单元格最大长度
		cleaned = cleaned[:32767]
	}
	
	return cleaned
}

// formatTimeForExcel 格式化时间用于Excel导出
func (s *MemberService) formatTimeForExcel(t time.Time) string {
	// 检查是否为零值时间
	if t.IsZero() {
		return ""
	}
	
	// 使用标准格式，避免时间解析问题
	return t.Format("2006-01-02 15:04:05")
}

func (s *MemberService) buildExportSearchConditions(query *gorm.DB, params inout.ExportMemberListReq) {
	var conditions []string
	var values []interface{}

	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		conditions = append(conditions, "(username LIKE ? OR phone LIKE ? OR nick_name LIKE ?)")
		values = append(values, searchTerm, searchTerm, searchTerm)
	}

	if params.Name != "" {
		nameTerm := "%" + params.Name + "%"
		conditions = append(conditions, "nick_name LIKE ?")
		values = append(values, nameTerm)
	}

	if params.Phone != "" {
		phoneTerm := "%" + params.Phone + "%"
		conditions = append(conditions, "phone LIKE ?")
		values = append(values, phoneTerm)
	}

	if params.Username != "" {
		usernameTerm := "%" + params.Username + "%"
		conditions = append(conditions, "username LIKE ?")
		values = append(values, usernameTerm)
	}

	if params.StartDate != "" || params.EndDate != "" {
		if params.StartDate != "" && params.EndDate != "" {
			conditions = append(conditions, "DATE(create_time) BETWEEN ? AND ?")
			values = append(values, params.StartDate, params.EndDate)
		} else if params.StartDate != "" {
			conditions = append(conditions, "DATE(create_time) >= ?")
			values = append(values, params.StartDate)
		} else {
			conditions = append(conditions, "DATE(create_time) <= ?")
			values = append(values, params.EndDate)
		}
	}

	if len(conditions) > 0 {
		whereClause := strings.Join(conditions, " AND ")
		query.Where(whereClause, values...)
	}
}

func (s *MemberService) buildSearchConditions(query *gorm.DB, params inout.GetMemberListReq) {
	var conditions []string
	var values []interface{}

	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		conditions = append(conditions, "(username LIKE ? OR phone LIKE ? OR nick_name LIKE ?)")
		values = append(values, searchTerm, searchTerm, searchTerm)
	}

	if params.Name != "" {
		nameTerm := "%" + params.Name + "%"
		conditions = append(conditions, "nick_name LIKE ?")
		values = append(values, nameTerm)
	}

	if params.Phone != "" {
		phoneTerm := "%" + params.Phone + "%"
		conditions = append(conditions, "phone LIKE ?")
		values = append(values, phoneTerm)
	}

	if params.Username != "" {
		usernameTerm := "%" + params.Username + "%"
		conditions = append(conditions, "username LIKE ?")
		values = append(values, usernameTerm)
	}

	if len(conditions) > 0 {
		whereClause := strings.Join(conditions, " AND ")
		query.Where(whereClause, values...)
	}
}

func (s *MemberService) formMemberData(data []admin_model.Member) []inout.MemberListItem {
	formattedData := make([]inout.MemberListItem, len(data))
	for i, item := range data {
		formattedData[i] = inout.MemberListItem{
			Id:         item.Id,
			UserName:   item.UserName,
			NickName:   item.NickName,
			Avatar:     item.Avatar,
			Phone:      item.Phone,
			Address:    item.Address,
			CreateTime: utils.FormatTime2(item.CreateTime),
			UpdateTime: utils.FormatTime2(item.UpdateTime),
		}
	}
	return formattedData
}
