package admin_service

import (
	"context"
	"encoding/json"
	"fmt"
	"nasa-go-admin/db"
	"nasa-go-admin/inout"
	"nasa-go-admin/model/admin_model"
	"nasa-go-admin/mongodb"
	"nasa-go-admin/redis"
	"nasa-go-admin/utils"
	"sort"
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

	// 关键词搜索（保持原有功能）
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

	// 用户名搜索
	if req.Username != "" {
		filter["username"] = bson.M{"$regex": req.Username, "$options": "i"}
	}

	// 日期时间范围搜索
	if req.StartDate != "" || req.EndDate != "" {
		timeFilter := bson.M{}
		if req.StartDate != "" {
			timeFilter["$gte"] = req.StartDate
		}
		if req.EndDate != "" {
			timeFilter["$lte"] = req.EndDate
		}
		filter["timestamp"] = timeFilter
	}

	// HTTP方法过滤
	if req.Method != "" {
		filter["method"] = bson.M{"$regex": req.Method, "$options": "i"}
	}

	// 状态码过滤
	if req.StatusCode > 0 {
		filter["status_code"] = req.StatusCode
	}

	// 客户端IP过滤
	if req.ClientIP != "" {
		filter["client_ip"] = bson.M{"$regex": req.ClientIP, "$options": "i"}
	}

	// 请求路径过滤
	if req.Path != "" {
		filter["path"] = bson.M{"$regex": req.Path, "$options": "i"}
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
		"filters": map[string]interface{}{
			"keyword":     req.Keyword,
			"username":    req.Username,
			"start_date":  req.StartDate,
			"end_date":    req.EndDate,
			"method":      req.Method,
			"status_code": req.StatusCode,
			"client_ip":   req.ClientIP,
			"path":        req.Path,
		},
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
	// 1. 尝试从缓存获取数据
	cacheKey := "system:dict:all"
	redisClient := redis.GetClient()
	if redisClient != nil {
		cachedData, err := redisClient.Get(context.Background(), cacheKey).Result()
		if err == nil {
			var result map[string][]map[string]interface{}
			if err := json.Unmarshal([]byte(cachedData), &result); err == nil {
				return result, nil
			}
		}
	}

	// 2. 使用一次JOIN查询获取所有数据
	var dictTypes []admin_model.DictType
	var dictDetails []admin_model.DictDetail

	// 2.1 首先获取所有字典类型
	if err := db.Dao.Model(&admin_model.DictType{}).
		Where("del_flag != ?", "T").
		Find(&dictTypes).Error; err != nil {
		return nil, fmt.Errorf("查询字典类型失败: %w", err)
	}

	// 2.2 如果有字典类型，一次性获取所有相关的字典值
	if len(dictTypes) > 0 {
		var typeIds []int
		for _, dt := range dictTypes {
			typeIds = append(typeIds, dt.Id)
		}

		if err := db.Dao.Model(&admin_model.DictDetail{}).
			Where("sys_dict_type_id IN ?", typeIds).
			Where("del_flag != ?", "T").
			Find(&dictDetails).Error; err != nil {
			return nil, fmt.Errorf("查询字典详情失败: %w", err)
		}
	}

	// 3. 构建结果Map
	result := make(map[string][]map[string]interface{})

	// 3.1 创建字典类型的快速查找map
	typeMap := make(map[int]admin_model.DictType)
	for _, dt := range dictTypes {
		typeMap[dt.Id] = dt
	}

	// 3.2 按字典类型分组字典值
	detailMap := make(map[int][]admin_model.DictDetail)
	for _, detail := range dictDetails {
		detailMap[detail.SysDictTypeId] = append(detailMap[detail.SysDictTypeId], detail)
	}

	// 3.3 构建最终结果
	for _, dictType := range dictTypes {
		details := detailMap[dictType.Id]
		detailList := make([]map[string]interface{}, 0, len(details))

		for _, detail := range details {
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

		result[dictType.TypeCode] = detailList
	}

	// 4. 将结果存入缓存
	if jsonData, err := json.Marshal(result); err == nil && redisClient != nil {
		// 设置缓存，有效期5分钟
		redisClient.Set(context.Background(), cacheKey, string(jsonData), 5*time.Minute)
	}

	return result, nil
}

// ClearSystemLog 清空指定数据库的日志数据
func (s *SystemService) ClearSystemLog(collectionName string) (interface{}, error) {
	collection := mongodb.GetCollection(collectionName, "logs")
	if collection == nil {
		return nil, fmt.Errorf("MongoDB collection 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 获取清空前的记录数
	totalBefore, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("获取记录数失败: %w", err)
	}

	// 清空所有日志数据
	result, err := collection.DeleteMany(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("清空日志失败: %w", err)
	}

	// 记录清空操作到系统日志
	clearLog := map[string]interface{}{
		"operation":     "clear_logs",
		"database":      collectionName,
		"cleared_count": result.DeletedCount,
		"timestamp":     utils.GetCurrentTimeForMongo(),
		"operator":      "admin_system",
	}

	// 将清空操作记录到系统操作日志中
	operationCollection := mongodb.GetCollection("admin_log_db", "logs")
	if operationCollection != nil {
		_, _ = operationCollection.InsertOne(ctx, clearLog)
	}

	return map[string]interface{}{
		"message":       "日志清空成功",
		"database":      collectionName,
		"cleared_count": result.DeletedCount,
		"total_before":  totalBefore,
		"clear_time":    time.Now().Format("2006-01-02 15:04:05"),
		"operation_id":  fmt.Sprintf("clear_%s_%d", collectionName, time.Now().Unix()),
	}, nil
}

// GetUserLog 获取用户端操作日志
func (s *SystemService) GetUserLog(req inout.GetUserLogReq) (interface{}, error) {
	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// 获取MongoDB集合
	collectionName := "app_log_db"
	collection := mongodb.GetCollection(collectionName, "logs")
	if collection == nil {
		return nil, fmt.Errorf("MongoDB collection 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 构建查询条件
	filter := bson.M{}

	// 关键词搜索
	if req.Keyword != "" {
		filter["$or"] = []bson.M{
			{"username": bson.M{"$regex": req.Keyword, "$options": "i"}},
			{"path": bson.M{"$regex": req.Keyword, "$options": "i"}},
			{"method": bson.M{"$regex": req.Keyword, "$options": "i"}},
			{"client_ip": bson.M{"$regex": req.Keyword, "$options": "i"}},
			{"user_agent": bson.M{"$regex": req.Keyword, "$options": "i"}},
			{"response_message": bson.M{"$regex": req.Keyword, "$options": "i"}},
		}
	}

	// 用户名搜索
	if req.Username != "" {
		filter["username"] = bson.M{"$regex": req.Username, "$options": "i"}
	}

	// 用户ID搜索
	if req.UserID > 0 {
		filter["user_id"] = req.UserID
	}

	// 日期时间范围搜索
	if req.StartDate != "" || req.EndDate != "" {
		dateFilter := bson.M{}
		if req.StartDate != "" {
			dateFilter["$gte"] = req.StartDate
		}
		if req.EndDate != "" {
			dateFilter["$lte"] = req.EndDate
		}
		filter["timestamp"] = dateFilter
	}

	// HTTP方法过滤
	if req.Method != "" {
		filter["method"] = bson.M{"$regex": req.Method, "$options": "i"}
	}

	// 状态码过滤
	if req.StatusCode > 0 {
		filter["status_code"] = req.StatusCode
	}

	// 客户端IP过滤
	if req.ClientIP != "" {
		filter["client_ip"] = bson.M{"$regex": req.ClientIP, "$options": "i"}
	}

	// 请求路径过滤
	if req.Path != "" {
		filter["path"] = bson.M{"$regex": req.Path, "$options": "i"}
	}

	// 设备类型过滤
	if req.DeviceType != "" {
		filter["device_type"] = bson.M{"$regex": req.DeviceType, "$options": "i"}
	}

	// 应用版本过滤
	if req.AppVersion != "" {
		filter["app_version"] = bson.M{"$regex": req.AppVersion, "$options": "i"}
	}

	// 操作类型过滤
	if req.Action != "" {
		filter["action"] = bson.M{"$regex": req.Action, "$options": "i"}
	}

	// 计算总数
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("获取日志总数失败: %w", err)
	}

	// 设置排序和分页
	opts := options.Find().
		SetSort(bson.D{{"timestamp", -1}}).
		SetSkip(int64((req.Page - 1) * req.PageSize)).
		SetLimit(int64(req.PageSize))

	// 执行查询
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("查询日志失败: %w", err)
	}
	defer cursor.Close(ctx)

	// 解析结果
	var logs []bson.M
	if err = cursor.All(ctx, &logs); err != nil {
		return nil, fmt.Errorf("解析日志数据失败: %w", err)
	}

	// 格式化日志数据
	formattedLogs := s.formatUserLogData(logs)

	// 计算统计信息
	stats := s.calculateUserLogStats(logs)

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
		"filters": map[string]interface{}{
			"keyword":     req.Keyword,
			"username":    req.Username,
			"user_id":     req.UserID,
			"start_date":  req.StartDate,
			"end_date":    req.EndDate,
			"method":      req.Method,
			"status_code": req.StatusCode,
			"client_ip":   req.ClientIP,
			"path":        req.Path,
			"device_type": req.DeviceType,
			"app_version": req.AppVersion,
			"action":      req.Action,
		},
	}

	return response, nil
}

// formatUserLogData 格式化用户日志数据
func (s *SystemService) formatUserLogData(logs []bson.M) []map[string]interface{} {
	var formattedLogs []map[string]interface{}
	for _, log := range logs {
		formattedLog := map[string]interface{}{
			"timestamp":        log["timestamp"],
			"method":           log["method"],
			"path":             log["path"],
			"client_ip":        log["client_ip"],
			"user_id":          log["user_id"],
			"username":         log["username"],
			"latency_ms":       log["latency_ms"],
			"status_code":      log["status_code"],
			"response_code":    log["response_code"],
			"response_message": log["response_message"],
			"request_size":     log["request_size"],
			"response_size":    log["response_size"],
			"user_agent":       log["user_agent"],
			"referer":          log["referer"],
			"log_type":         log["log_type"],
			"local_timestamp":  log["local_timestamp"],
		}

		// 添加用户端特有的字段
		if deviceType, exists := log["device_type"]; exists {
			formattedLog["device_type"] = deviceType
		}
		if appVersion, exists := log["app_version"]; exists {
			formattedLog["app_version"] = appVersion
		}
		if action, exists := log["action"]; exists {
			formattedLog["action"] = action
		}
		if platform, exists := log["platform"]; exists {
			formattedLog["platform"] = platform
		}

		formattedLogs = append(formattedLogs, formattedLog)
	}
	return formattedLogs
}

// calculateUserLogStats 计算用户日志统计信息
func (s *SystemService) calculateUserLogStats(logs []bson.M) map[string]interface{} {
	if len(logs) == 0 {
		return map[string]interface{}{
			"total_requests":    0,
			"success_rate":      0.0,
			"avg_response_time": 0.0,
			"top_methods":       []map[string]interface{}{},
			"top_paths":         []map[string]interface{}{},
			"top_users":         []map[string]interface{}{},
			"device_types":      []map[string]interface{}{},
			"status_codes":      []map[string]interface{}{},
		}
	}

	// 统计变量
	totalRequests := len(logs)
	successCount := 0
	totalResponseTime := 0.0
	methods := make(map[string]int)
	paths := make(map[string]int)
	users := make(map[string]int)
	deviceTypes := make(map[string]int)
	statusCodes := make(map[string]int)

	// 遍历日志计算统计
	for _, log := range logs {
		// 成功请求统计
		if statusCode, exists := log["status_code"]; exists {
			if status, ok := statusCode.(int32); ok && status >= 200 && status < 300 {
				successCount++
			}
			statusCodes[fmt.Sprintf("%d", statusCode)]++
		}

		// 响应时间统计
		if latency, exists := log["latency_ms"]; exists {
			if lat, ok := latency.(int32); ok {
				totalResponseTime += float64(lat)
			}
		}

		// HTTP方法统计
		if method, exists := log["method"]; exists {
			if m, ok := method.(string); ok {
				methods[m]++
			}
		}

		// 路径统计
		if path, exists := log["path"]; exists {
			if p, ok := path.(string); ok {
				paths[p]++
			}
		}

		// 用户统计
		if username, exists := log["username"]; exists {
			if u, ok := username.(string); ok && u != "" {
				users[u]++
			}
		}

		// 设备类型统计
		if deviceType, exists := log["device_type"]; exists {
			if dt, ok := deviceType.(string); ok {
				deviceTypes[dt]++
			}
		}
	}

	// 计算成功率
	successRate := 0.0
	if totalRequests > 0 {
		successRate = float64(successCount) / float64(totalRequests) * 100
	}

	// 计算平均响应时间
	avgResponseTime := 0.0
	if totalRequests > 0 {
		avgResponseTime = totalResponseTime / float64(totalRequests)
	}

	// 获取前5个最常用的HTTP方法
	topMethods := s.getTopItems(methods, 5)

	// 获取前5个最常用的路径
	topPaths := s.getTopItems(paths, 5)

	// 获取前5个最活跃的用户
	topUsers := s.getTopItems(users, 5)

	// 获取设备类型分布
	topDeviceTypes := s.getTopItems(deviceTypes, 10)

	// 获取状态码分布
	topStatusCodes := s.getTopItems(statusCodes, 10)

	return map[string]interface{}{
		"total_requests":    totalRequests,
		"success_rate":      successRate,
		"avg_response_time": avgResponseTime,
		"top_methods":       topMethods,
		"top_paths":         topPaths,
		"top_users":         topUsers,
		"device_types":      topDeviceTypes,
		"status_codes":      topStatusCodes,
	}
}

// getTopItems 获取前N个最常用的项目
func (s *SystemService) getTopItems(items map[string]int, limit int) []map[string]interface{} {
	// 转换为切片进行排序
	type itemCount struct {
		Item  string
		Count int
	}

	var itemCounts []itemCount
	for item, count := range items {
		itemCounts = append(itemCounts, itemCount{Item: item, Count: count})
	}

	// 按计数降序排序
	sort.Slice(itemCounts, func(i, j int) bool {
		return itemCounts[i].Count > itemCounts[j].Count
	})

	// 返回前N个
	var result []map[string]interface{}
	for i, item := range itemCounts {
		if i >= limit {
			break
		}
		result = append(result, map[string]interface{}{
			"item":  item.Item,
			"count": item.Count,
		})
	}

	return result
}
