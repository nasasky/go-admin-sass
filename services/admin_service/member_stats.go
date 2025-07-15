package admin_service

import (
	"fmt"
	"math"
	"nasa-go-admin/db"
	"nasa-go-admin/inout"
	"nasa-go-admin/model/admin_model"
	"time"

	"github.com/gin-gonic/gin"
)

type MemberStatsService struct{}

// GetMemberStats 获取会员统计数据
func (s *MemberStatsService) GetMemberStats(c *gin.Context, params inout.GetMemberStatsReq) (*inout.MemberStatsResp, error) {
	// 解析时间范围
	startDate, endDate, err := s.parseDateRange(params)
	if err != nil {
		return nil, fmt.Errorf("解析日期范围失败: %w", err)
	}

	// 获取统计数据
	chartData, err := s.generateChartData(startDate, endDate, params.Type)
	if err != nil {
		return nil, fmt.Errorf("生成图表数据失败: %w", err)
	}

	// 生成汇总信息
	summary, err := s.generateSummary(chartData, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("生成汇总信息失败: %w", err)
	}

	// 生成趋势信息
	trendInfo := s.generateTrendInfo(chartData)

	// 构建响应
	resp := &inout.MemberStatsResp{
		Type: params.Type,
		DateRange: &inout.DateRangeInfo{
			StartDate:   startDate.Format("2006-01-02"),
			EndDate:     endDate.Format("2006-01-02"),
			TotalDays:   int(endDate.Sub(startDate).Hours()/24) + 1,
			CurrentDate: time.Now().Format("2006-01-02"),
		},
		Summary:   summary,
		ChartData: chartData,
		TrendInfo: trendInfo,
		UpdatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	return resp, nil
}

// parseDateRange 解析日期范围
func (s *MemberStatsService) parseDateRange(params inout.GetMemberStatsReq) (time.Time, time.Time, error) {
	var startDate, endDate time.Time
	var err error

	// 如果提供了具体的开始和结束日期
	if params.StartDate != "" && params.EndDate != "" {
		startDate, err = time.Parse("2006-01-02", params.StartDate)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("开始日期格式错误: %w", err)
		}

		endDate, err = time.Parse("2006-01-02", params.EndDate)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("结束日期格式错误: %w", err)
		}
	} else {
		// 根据Days参数计算日期范围
		days := params.Days
		if days == 0 {
			// 根据统计类型设置默认天数
			switch params.Type {
			case "daily":
				days = 30 // 最近30天
			case "weekly":
				days = 84 // 最近12周
			case "monthly":
				days = 365 // 最近12个月
			default:
				days = 30
			}
		}

		endDate = time.Now().Truncate(24 * time.Hour)
		startDate = endDate.AddDate(0, 0, -(days - 1))
	}

	// 确保开始日期不晚于结束日期
	if startDate.After(endDate) {
		return time.Time{}, time.Time{}, fmt.Errorf("开始日期不能晚于结束日期")
	}

	return startDate, endDate, nil
}

// generateChartData 生成图表数据
func (s *MemberStatsService) generateChartData(startDate, endDate time.Time, statsType string) ([]inout.MemberStatsChartPoint, error) {
	switch statsType {
	case "daily":
		return s.generateDailyChartData(startDate, endDate)
	case "weekly":
		return s.generateWeeklyChartData(startDate, endDate)
	case "monthly":
		return s.generateMonthlyChartData(startDate, endDate)
	default:
		return nil, fmt.Errorf("不支持的统计类型: %s", statsType)
	}
}

// generateDailyChartData 生成每日图表数据
func (s *MemberStatsService) generateDailyChartData(startDate, endDate time.Time) ([]inout.MemberStatsChartPoint, error) {
	var chartData []inout.MemberStatsChartPoint

	// 获取总会员数（到结束日期为止）
	var totalMembersAtEnd int64
	err := db.Dao.Model(&admin_model.Member{}).
		Where("DATE(create_time) <= ?", endDate.Format("2006-01-02")).
		Count(&totalMembersAtEnd).Error
	if err != nil {
		return nil, fmt.Errorf("查询总会员数失败: %w", err)
	}

	// 按日期查询每日新增会员数
	type DailyStats struct {
		Date       string `json:"date"`
		NewMembers int64  `json:"new_members"`
	}

	var dailyStats []DailyStats
	err = db.Dao.Model(&admin_model.Member{}).
		Select("DATE(create_time) as date, COUNT(*) as new_members").
		Where("DATE(create_time) BETWEEN ? AND ?", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")).
		Group("DATE(create_time)").
		Order("date ASC").
		Find(&dailyStats).Error
	if err != nil {
		return nil, fmt.Errorf("查询每日新增会员数失败: %w", err)
	}

	// 创建日期到新增会员数的映射
	statsMap := make(map[string]int64)
	for _, stat := range dailyStats {
		statsMap[stat.Date] = stat.NewMembers
	}

	// 生成完整的日期序列
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		newMembers := int(statsMap[dateStr])

		// 计算当日的累计会员数（往前推算）
		var totalAtThisDay int64
		err := db.Dao.Model(&admin_model.Member{}).
			Where("DATE(create_time) <= ?", dateStr).
			Count(&totalAtThisDay).Error
		if err != nil {
			return nil, fmt.Errorf("查询累计会员数失败: %w", err)
		}

		point := inout.MemberStatsChartPoint{
			Date:          dateStr,
			NewMembers:    newMembers,
			TotalMembers:  int(totalAtThisDay),
			DayOfWeek:     s.getDayOfWeekName(d.Weekday()),
			FormattedDate: d.Format("01月02日"),
		}
		chartData = append(chartData, point)
	}

	return chartData, nil
}

// generateWeeklyChartData 生成每周图表数据
func (s *MemberStatsService) generateWeeklyChartData(startDate, endDate time.Time) ([]inout.MemberStatsChartPoint, error) {
	var chartData []inout.MemberStatsChartPoint

	// 调整到周一开始
	for startDate.Weekday() != time.Monday {
		startDate = startDate.AddDate(0, 0, -1)
	}

	// 按周统计
	for weekStart := startDate; weekStart.Before(endDate) || weekStart.Equal(endDate); weekStart = weekStart.AddDate(0, 0, 7) {
		weekEnd := weekStart.AddDate(0, 0, 6)
		if weekEnd.After(endDate) {
			weekEnd = endDate
		}

		// 查询这一周的新增会员数
		var newMembers int64
		err := db.Dao.Model(&admin_model.Member{}).
			Where("DATE(create_time) BETWEEN ? AND ?", weekStart.Format("2006-01-02"), weekEnd.Format("2006-01-02")).
			Count(&newMembers).Error
		if err != nil {
			return nil, fmt.Errorf("查询周新增会员数失败: %w", err)
		}

		// 查询到这一周结束时的累计会员数
		var totalMembers int64
		err = db.Dao.Model(&admin_model.Member{}).
			Where("DATE(create_time) <= ?", weekEnd.Format("2006-01-02")).
			Count(&totalMembers).Error
		if err != nil {
			return nil, fmt.Errorf("查询累计会员数失败: %w", err)
		}

		point := inout.MemberStatsChartPoint{
			Date:          weekStart.Format("2006-01-02"),
			NewMembers:    int(newMembers),
			TotalMembers:  int(totalMembers),
			DayOfWeek:     "Monday",
			FormattedDate: fmt.Sprintf("%s - %s", weekStart.Format("01/02"), weekEnd.Format("01/02")),
		}
		chartData = append(chartData, point)
	}

	return chartData, nil
}

// generateMonthlyChartData 生成每月图表数据
func (s *MemberStatsService) generateMonthlyChartData(startDate, endDate time.Time) ([]inout.MemberStatsChartPoint, error) {
	var chartData []inout.MemberStatsChartPoint

	// 调整到月初
	startDate = time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, startDate.Location())

	// 按月统计
	for monthStart := startDate; monthStart.Before(endDate) || monthStart.Equal(endDate); monthStart = monthStart.AddDate(0, 1, 0) {
		monthEnd := monthStart.AddDate(0, 1, -1)
		if monthEnd.After(endDate) {
			monthEnd = endDate
		}

		// 查询这个月的新增会员数
		var newMembers int64
		err := db.Dao.Model(&admin_model.Member{}).
			Where("DATE(create_time) BETWEEN ? AND ?", monthStart.Format("2006-01-02"), monthEnd.Format("2006-01-02")).
			Count(&newMembers).Error
		if err != nil {
			return nil, fmt.Errorf("查询月新增会员数失败: %w", err)
		}

		// 查询到这个月结束时的累计会员数
		var totalMembers int64
		err = db.Dao.Model(&admin_model.Member{}).
			Where("DATE(create_time) <= ?", monthEnd.Format("2006-01-02")).
			Count(&totalMembers).Error
		if err != nil {
			return nil, fmt.Errorf("查询累计会员数失败: %w", err)
		}

		point := inout.MemberStatsChartPoint{
			Date:          monthStart.Format("2006-01-02"),
			NewMembers:    int(newMembers),
			TotalMembers:  int(totalMembers),
			DayOfWeek:     "",
			FormattedDate: monthStart.Format("2006年01月"),
		}
		chartData = append(chartData, point)
	}

	return chartData, nil
}

// generateSummary 生成汇总信息
func (s *MemberStatsService) generateSummary(chartData []inout.MemberStatsChartPoint, startDate, endDate time.Time) (*inout.MemberStatsSummary, error) {
	if len(chartData) == 0 {
		return &inout.MemberStatsSummary{}, nil
	}

	// 计算总的新增会员数
	totalNewMembers := 0
	maxNewMembers := 0
	maxNewMembersDay := ""

	for _, point := range chartData {
		totalNewMembers += point.NewMembers
		if point.NewMembers > maxNewMembers {
			maxNewMembers = point.NewMembers
			maxNewMembersDay = point.Date
		}
	}

	// 计算日均新增会员数
	days := len(chartData)
	avgDailyNewMembers := float64(totalNewMembers) / float64(days)

	// 计算增长率
	var growthRate float64
	if len(chartData) >= 2 {
		firstTotal := chartData[0].TotalMembers - chartData[0].NewMembers // 开始时的总数
		lastTotal := chartData[len(chartData)-1].TotalMembers
		if firstTotal > 0 {
			growthRate = (float64(lastTotal-firstTotal) / float64(firstTotal)) * 100
		}
	}

	// 取最后一个点的累计会员数作为总会员数
	totalMembers := 0
	if len(chartData) > 0 {
		totalMembers = chartData[len(chartData)-1].TotalMembers
	}

	summary := &inout.MemberStatsSummary{
		TotalMembers:        totalMembers,
		NewMembersInPeriod:  totalNewMembers,
		AvgDailyNewMembers:  math.Round(avgDailyNewMembers*100) / 100,
		PeakNewMembersDay:   maxNewMembersDay,
		PeakNewMembersCount: maxNewMembers,
		GrowthRate:          math.Round(growthRate*100) / 100,
	}

	return summary, nil
}

// generateTrendInfo 生成趋势信息
func (s *MemberStatsService) generateTrendInfo(chartData []inout.MemberStatsChartPoint) *inout.MemberTrendInfo {
	if len(chartData) < 2 {
		return &inout.MemberTrendInfo{
			Trend:     "stable",
			TrendDesc: "数据不足，无法分析趋势",
		}
	}

	// 计算最近几天的趋势
	dataLength := len(chartData)
	recentCount := minInt(7, dataLength/2) // 取最近几天或数据的一半
	if recentCount < 2 {
		recentCount = 2
	}

	// 计算前半段和后半段的平均值
	midPoint := dataLength - recentCount
	firstHalfAvg := 0.0
	secondHalfAvg := 0.0

	for i := 0; i < midPoint; i++ {
		firstHalfAvg += float64(chartData[i].NewMembers)
	}
	firstHalfAvg /= float64(midPoint)

	for i := midPoint; i < dataLength; i++ {
		secondHalfAvg += float64(chartData[i].NewMembers)
	}
	secondHalfAvg /= float64(recentCount)

	// 计算趋势
	var trend string
	var trendPercent float64
	var trendDesc string

	if firstHalfAvg > 0 {
		trendPercent = ((secondHalfAvg - firstHalfAvg) / firstHalfAvg) * 100
	}

	if math.Abs(trendPercent) < 5 {
		trend = "stable"
		trendDesc = "会员增长相对稳定"
	} else if trendPercent > 0 {
		trend = "up"
		trendDesc = fmt.Sprintf("会员增长呈上升趋势，增长%.1f%%", trendPercent)
	} else {
		trend = "down"
		trendDesc = fmt.Sprintf("会员增长呈下降趋势，下降%.1f%%", math.Abs(trendPercent))
	}

	comparedToPrev := "与上一周期相比"
	if trendPercent > 0 {
		comparedToPrev = fmt.Sprintf("比上一周期增长了%.1f%%", trendPercent)
	} else if trendPercent < 0 {
		comparedToPrev = fmt.Sprintf("比上一周期下降了%.1f%%", math.Abs(trendPercent))
	} else {
		comparedToPrev = "与上一周期基本持平"
	}

	return &inout.MemberTrendInfo{
		Trend:          trend,
		TrendPercent:   math.Round(trendPercent*100) / 100,
		TrendDesc:      trendDesc,
		ComparedToPrev: comparedToPrev,
	}
}

// getDayOfWeekName 获取星期几的中文名称
func (s *MemberStatsService) getDayOfWeekName(weekday time.Weekday) string {
	switch weekday {
	case time.Sunday:
		return "周日"
	case time.Monday:
		return "周一"
	case time.Tuesday:
		return "周二"
	case time.Wednesday:
		return "周三"
	case time.Thursday:
		return "周四"
	case time.Friday:
		return "周五"
	case time.Saturday:
		return "周六"
	default:
		return "未知"
	}
}

// minInt 返回两个整数的最小值
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
