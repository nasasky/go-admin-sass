package admin_service

import (
	"context"
	"fmt"
	"nasa-go-admin/db"
	"nasa-go-admin/inout"
	"nasa-go-admin/model/admin_model"
	"nasa-go-admin/mongodb"
	"nasa-go-admin/utils"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SystemService 系统服务
type SystemService struct{}

// GetSystemLog 从 MongoDB 中读取所有日志数据
func (s *SystemService) GetSystemLog(req inout.GetSystemLogReq, collectionName string) (interface{}, error) {
	collection := mongodb.GetCollection(collectionName, "logs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 构建查询条件
	filter := bson.M{}
	if req.Keyword != "" {
		// 扩展关键词搜索，支持多个字段
		filter["$or"] = []bson.M{
			{"username": bson.M{"$regex": req.Keyword, "$options": "i"}},
			{"path": bson.M{"$regex": req.Keyword, "$options": "i"}},
			{"method": bson.M{"$regex": req.Keyword, "$options": "i"}},
			{"client_ip": bson.M{"$regex": req.Keyword, "$options": "i"}},
			{"user_agent": bson.M{"$regex": req.Keyword, "$options": "i"}},
			{"response_message": bson.M{"$regex": req.Keyword, "$options": "i"}},
		}
	}

	// 设置默认分页参数
	req.Page = max(req.Page, 1)
	req.PageSize = max(req.PageSize, 10)

	// 设置分页选项
	findOptions := options.Find()
	findOptions.SetSkip(int64((req.Page - 1) * req.PageSize))
	findOptions.SetLimit(int64(req.PageSize))
	findOptions.SetSort(bson.D{{"timestamp", -1}})

	// 查询日志数据
	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []bson.M
	if err = cursor.All(ctx, &logs); err != nil {
		return nil, err
	}

	// 获取总数
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}

	// 添加一些统计信息
	stats := s.calculateLogStats(logs)

	// 格式化时间字段
	formattedLogs := utils.FormatTimeFieldsForResponse(logs)

	// 构造返回结果
	response := map[string]interface{}{
		"total":      total,
		"page":       req.Page,
		"page_size":  req.PageSize,
		"items":      formattedLogs,
		"stats":      stats,
		"query_time": time.Now().Format("2006-01-02 15:04:05"),
		"database":   collectionName,
		"timezone":   utils.GetTimeZoneInfo(),
	}

	return response, nil
}

// calculateLogStats 计算日志统计信息
func (s *SystemService) calculateLogStats(logs []bson.M) map[string]interface{} {
	stats := map[string]interface{}{
		"total_requests":   len(logs),
		"unique_users":     0,
		"unique_ips":       0,
		"methods":          make(map[string]int),
		"status_codes":     make(map[int]int),
		"paths":            make(map[string]int),
		"avg_latency_ms":   0.0,
		"total_latency_ms": 0.0,
		"max_latency_ms":   0.0,
		"min_latency_ms":   99999.0,
		"error_count":      0,
		"success_count":    0,
	}

	uniqueUsers := make(map[string]bool)
	uniqueIPs := make(map[string]bool)
	methods := make(map[string]int)
	statusCodes := make(map[int]int)
	paths := make(map[string]int)

	var totalLatency float64 = 0
	var maxLatency float64 = 0
	var minLatency float64 = 99999
	var errorCount int = 0
	var successCount int = 0

	for _, log := range logs {
		// 统计用户
		if userID, ok := log["user_id"].(string); ok && userID != "" {
			uniqueUsers[userID] = true
		}

		// 统计IP
		if clientIP, ok := log["client_ip"].(string); ok && clientIP != "" {
			uniqueIPs[clientIP] = true
		}

		// 统计HTTP方法
		if method, ok := log["method"].(string); ok {
			methods[method]++
		}

		// 统计状态码
		if statusCode, ok := log["status_code"].(int32); ok {
			statusCodes[int(statusCode)]++
			if int(statusCode) >= 400 {
				errorCount++
			} else {
				successCount++
			}
		} else if statusCode, ok := log["status_code"].(int); ok {
			statusCodes[statusCode]++
			if statusCode >= 400 {
				errorCount++
			} else {
				successCount++
			}
		}

		// 统计路径
		if path, ok := log["path"].(string); ok {
			paths[path]++
		}

		// 统计延迟
		if latencyMs, ok := log["latency_ms"].(int64); ok {
			latency := float64(latencyMs)
			totalLatency += latency
			if latency > maxLatency {
				maxLatency = latency
			}
			if latency < minLatency {
				minLatency = latency
			}
		} else if latencyMs, ok := log["latency_ms"].(int32); ok {
			latency := float64(latencyMs)
			totalLatency += latency
			if latency > maxLatency {
				maxLatency = latency
			}
			if latency < minLatency {
				minLatency = latency
			}
		}
	}

	// 计算平均延迟
	var avgLatency float64 = 0
	if len(logs) > 0 {
		avgLatency = totalLatency / float64(len(logs))
	}

	stats["unique_users"] = len(uniqueUsers)
	stats["unique_ips"] = len(uniqueIPs)
	stats["methods"] = methods
	stats["status_codes"] = statusCodes
	stats["top_paths"] = s.getTopPaths(paths, 10)
	stats["avg_latency_ms"] = avgLatency
	stats["total_latency_ms"] = totalLatency
	stats["max_latency_ms"] = maxLatency
	if minLatency == 99999 {
		stats["min_latency_ms"] = 0
	} else {
		stats["min_latency_ms"] = minLatency
	}
	stats["error_count"] = errorCount
	stats["success_count"] = successCount

	if len(logs) > 0 {
		stats["success_rate"] = float64(successCount) / float64(len(logs)) * 100
		stats["error_rate"] = float64(errorCount) / float64(len(logs)) * 100
	}

	return stats
}

// getTopPaths 获取访问最多的路径
func (s *SystemService) getTopPaths(paths map[string]int, limit int) []map[string]interface{} {
	type pathCount struct {
		Path  string
		Count int
	}

	var pathCounts []pathCount
	for path, count := range paths {
		pathCounts = append(pathCounts, pathCount{Path: path, Count: count})
	}

	// 简单排序（冒泡排序）
	for i := 0; i < len(pathCounts)-1; i++ {
		for j := 0; j < len(pathCounts)-i-1; j++ {
			if pathCounts[j].Count < pathCounts[j+1].Count {
				pathCounts[j], pathCounts[j+1] = pathCounts[j+1], pathCounts[j]
			}
		}
	}

	// 取前N个
	if len(pathCounts) > limit {
		pathCounts = pathCounts[:limit]
	}

	var result []map[string]interface{}
	for _, pc := range pathCounts {
		result = append(result, map[string]interface{}{
			"path":  pc.Path,
			"count": pc.Count,
		})
	}

	return result
}

// AddDictType 添加字典类型
func (s *SystemService) AddDictType(c *gin.Context, params inout.AddDictTypeReq) error {
	// 设置默认值
	if params.IsLock == "" { // 如果 IsLock 是 string 类型
		params.IsLock = "T" // 设置为字符串 "1"
	}
	if params.IsShow == "" { // 如果 IsShow 是 string 类型
		params.IsShow = "T" // 设置为字符串 "1"

	}
	if params.DelFlag == "" { // 如果 Type 是 string 类型
		params.DelFlag = "F" // 设置为字符串 "1"
	}

	// 构建字典类型对象
	dictType := admin_model.DictType{
		TypeName:   params.TypeName,
		TypeCode:   params.TypeCode,
		IsLock:     params.IsLock,
		IsShow:     params.IsShow,
		Type:       params.Type,
		DelFlag:    params.DelFlag,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
		Remark:     params.Remark,
	}

	// 插入到数据库
	err := db.Dao.Create(&dictType).Error
	if err != nil {
		return err
	}
	return nil
}

// AddDictValue 添加字典值
func (s *SystemService) AddDictValue(c *gin.Context, params inout.AddDictValueReq) error {
	// 设置默认值
	if params.IsLock == "" { // 如果 IsLock 是 string 类型
		params.IsLock = "T" // 设置为字符串 "1"
	}
	if params.IsShow == "" { // 如果 IsShow 是 string 类型
		params.IsShow = "T" // 设置为字符串 "1"

	}
	if params.DelFlag == "" { // 如果 Type 是 string 类型
		params.DelFlag = "F" // 设置为字符串 "1"
	}

	// 构建字典值对象
	dictValue := admin_model.AddDictDetail{
		SysDictTypeId:     params.Id,
		CodeName:          params.CodeName,
		Code:              params.Code,
		Alias:             params.Alias,
		CallbackShowStyle: params.CallbackShowStyle,
		IsLock:            params.IsLock,
		IsShow:            params.IsShow,
		CreateTime:        time.Now(),
		UpdateTime:        time.Now(),
		Remark:            params.Remark,
	}

	if params.Id > 0 {
		dictValue.Id = params.Id
	}

	err := db.Dao.Create(&dictValue).Error
	if err != nil {
		return err
	}
	return nil
}

// 获取系统字典参数类型列表
func (s *SystemService) GetDictTypeList(c *gin.Context, params inout.ListpageReq) (interface{}, error) {
	var data []admin_model.DictType
	var total int64

	// 设置默认分页参数
	params.Page = max(params.Page, 1)
	params.PageSize = max(params.PageSize, 10)

	// 构建查询
	query := db.Dao.Model(&admin_model.DictType{}).Order("create_time DESC")

	// 计算偏移量
	offset := (params.Page - 1) * params.PageSize

	// 执行查询
	err := query.Count(&total).Offset(offset).Limit(params.PageSize).Find(&data).Error
	if err != nil {
		return nil, err
	}

	// 格式化数据
	formattedData := formatDictTypeData(data)

	response := admin_model.SettingType{
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
		Items:    formattedData,
	}
	return response, nil

}

func (s *SystemService) GetDictDetailData(c *gin.Context, params inout.DicteReq) (interface{}, error) {
	var data []admin_model.DictDetail
	var total int64
	// 设置默认分页参数
	params.Page = max(params.Page, 1)
	params.PageSize = max(params.PageSize, 10)
	// 构建查询
	query := db.Dao.Model(&admin_model.DictDetail{})
	// 添加ID过滤条件
	if params.Id > 0 {
		query = query.Where("sys_dict_type_id = ?", params.Id)
	}
	// 计算偏移量
	offset := (params.Page - 1) * params.PageSize
	// 执行查询
	err := query.Count(&total).Offset(offset).Limit(params.PageSize).Find(&data).Error
	if err != nil {
		return nil, err
	}
	// 格式化数据
	formattedData := []map[string]interface{}{}
	if len(data) > 0 {
		formattedData = formatDictTypeDataDetail(data)
	}
	response := map[string]interface{}{
		"total":     total,
		"page":      params.Page,
		"page_size": params.PageSize,
		"items":     formattedData,
	}
	return response, nil

}

// 格式化字典类型数据
func formatDictTypeData(data []admin_model.DictType) []admin_model.DictTypeResp {
	var resp []admin_model.DictTypeResp
	for _, item := range data {
		resp = append(resp, admin_model.DictTypeResp{
			Id:         item.Id,
			TypeName:   item.TypeName,
			TypeCode:   item.TypeCode,
			IsLock:     item.IsLock,
			IsShow:     item.IsShow,
			Type:       item.Type,
			CreateTime: utils.FormatTime2(item.CreateTime),
			UpdateTime: utils.FormatTime2(item.UpdateTime),
			Remark:     item.Remark,
		})
	}
	return resp
}

// 格式化字典详情数据并处理时间格式
func formatDictTypeDataDetail(data []admin_model.DictDetail) []map[string]interface{} {
	var resp []map[string]interface{}
	for _, item := range data {
		dictMap := map[string]interface{}{
			"id":                  item.Id,
			"sys_dict_type_id":    item.SysDictTypeId,
			"code_name":           item.CodeName,
			"sort":                item.Sort,
			"del_flag":            item.DelFlag,
			"is_lock":             item.IsLock,
			"is_show":             item.IsShow,
			"code":                item.Code,
			"callback_show_style": item.CallbackShowStyle,
			"create_time":         utils.FormatTime2(item.CreateTime),
			"update_time":         utils.FormatTime2(item.UpdateTime),
			"remark":              item.Remark,
		}
		resp = append(resp, dictMap)
	}
	return resp
}

// GetAllDictType 获取所有字典类型及其对应的字典值
func (s *SystemService) GetAllDictType(c *gin.Context) (interface{}, error) {
	// 查询所有字典类型
	var dictTypes []admin_model.DictType
	err := db.Dao.Model(&admin_model.DictType{}).Find(&dictTypes).Error
	if err != nil {
		return nil, fmt.Errorf("查询字典类型失败: %w", err)
	}

	// 构建返回结果
	result := make(map[string][]map[string]interface{})

	for _, dictType := range dictTypes {
		// 查询每个字典类型对应的字典值
		var dictDetails []admin_model.DictDetail
		err := db.Dao.Model(&admin_model.DictDetail{}).
			Where("sys_dict_type_id = ?", dictType.Id).
			Find(&dictDetails).Error
		if err != nil {
			return nil, fmt.Errorf("查询字典详情失败: %w", err)
		}

		// 构建字典值列表
		detailList := make([]map[string]interface{}, 0)
		for _, detail := range dictDetails {
			detailList = append(detailList, map[string]interface{}{
				"id":                detail.Id,
				"sysDictTypeId":     detail.SysDictTypeId,
				"codeName":          detail.CodeName,
				"callbackShowStyle": detail.CallbackShowStyle,
				"sort":              detail.Sort,
				"code":              detail.Code,
				"is_lock":           detail.IsLock,
				"is_show":           detail.IsShow,
				"sysDictTypeCode":   dictType.TypeCode,
				"sysDictTypeName":   dictType.TypeName,
				"create_time":       utils.FormatTime2(detail.CreateTime),
				"update_time":       utils.FormatTime2(detail.UpdateTime),
				"remark":            detail.Remark,
			})
		}

		// 将字典类型和对应的字典值加入结果
		result[dictType.TypeCode] = detailList
	}

	return result, nil
}
