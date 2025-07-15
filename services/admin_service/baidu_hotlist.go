package admin_service

import (
	"encoding/json"
	"fmt"
	"io"
	"nasa-go-admin/inout"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// BaiduHotSearchService 百度热搜服务
type BaiduHotSearchService struct{}

// BaiduApiResponse 百度API返回结构
type BaiduApiResponse struct {
	Data []struct {
		QueryWord string `json:"query_word"`
		ShowLink  string `json:"show_link"`
		HotScore  int    `json:"hot_score"`
		Img       string `json:"img"`
		Desc      string `json:"desc"`
	} `json:"data"`
}

// GetBaiduHotSearch 获取百度热搜数据
func (s *BaiduHotSearchService) GetBaiduHotSearch(c *gin.Context, count int) (*inout.BaiduHotSearchResp, error) {
	// 设置默认值
	if count <= 0 || count > 50 {
		count = 20
	}

	// 百度热搜API地址
	apiUrl := "http://top.baidu.com/api/boardsearch?platform=pc&domain=news"

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: time.Second * 30,
	}

	// 设置请求头
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置User-Agent模拟浏览器
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Referer", "https://top.baidu.com/board?tab=realtime")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 如果API请求失败，使用备用方法
	if resp.StatusCode != 200 {
		return s.getBaiduHotSearchFallback(count)
	}

	// 解析JSON数据
	var apiResp BaiduApiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		// 如果解析失败，使用备用方法
		return s.getBaiduHotSearchFallback(count)
	}

	// 转换数据格式
	var hotSearchItems []inout.BaiduHotSearchItem
	for i, item := range apiResp.Data {
		if i >= count {
			break
		}

		// 处理热度值
		hotValue := "0"
		if item.HotScore > 0 {
			hotValue = strconv.Itoa(item.HotScore)
		}

		// 处理标题
		title := item.QueryWord
		if title == "" {
			title = "未知标题"
		}

		hotSearchItems = append(hotSearchItems, inout.BaiduHotSearchItem{
			Rank:       i + 1,
			Title:      title,
			HotValue:   hotValue,
			Link:       item.ShowLink,
			Tag:        s.getHotTag(item.HotScore),
			Desc:       item.Desc,
			ImageUrl:   item.Img,
			UpdateTime: time.Now().Format("2006-01-02 15:04:05"),
		})
	}

	// 构建响应
	response := &inout.BaiduHotSearchResp{
		Code:       0,
		Message:    "获取成功",
		Data:       hotSearchItems,
		Total:      len(hotSearchItems),
		UpdateTime: time.Now().Format("2006-01-02 15:04:05"),
	}

	return response, nil
}

// getBaiduHotSearchFallback 备用方法，返回模拟数据
func (s *BaiduHotSearchService) getBaiduHotSearchFallback(count int) (*inout.BaiduHotSearchResp, error) {
	// 模拟热搜数据
	mockData := []struct {
		title    string
		hotValue string
		tag      string
		desc     string
	}{
		{"2024年新闻热点", "4850000", "热", "今日最新热点新闻"},
		{"科技发展前沿", "3920000", "新", "科技创新最新动态"},
		{"经济形势分析", "3150000", "热", "当前经济发展趋势"},
		{"教育改革政策", "2680000", "新", "教育政策最新变化"},
		{"健康生活方式", "2340000", "热", "健康生活新理念"},
		{"体育赛事精彩", "2180000", "热", "体育比赛最新战况"},
		{"文化艺术活动", "1950000", "新", "文化艺术最新资讯"},
		{"环保绿色发展", "1720000", "热", "环保政策新举措"},
		{"社会民生关注", "1580000", "热", "社会热点问题"},
		{"国际时事动态", "1420000", "新", "国际新闻最新报道"},
		{"金融投资理财", "1280000", "热", "金融市场最新动态"},
		{"房地产市场", "1150000", "热", "房地产行业分析"},
		{"汽车行业发展", "1030000", "新", "汽车产业最新消息"},
		{"旅游出行指南", "920000", "热", "旅游资讯和攻略"},
		{"美食文化探索", "810000", "新", "美食文化新发现"},
		{"时尚潮流趋势", "750000", "热", "时尚流行新趋势"},
		{"科学研究发现", "680000", "新", "科学技术新突破"},
		{"娱乐圈动态", "610000", "热", "娱乐新闻最新消息"},
		{"互联网科技", "540000", "新", "互联网行业动态"},
		{"生活小常识", "480000", "热", "实用生活小贴士"},
	}

	var hotSearchItems []inout.BaiduHotSearchItem
	for i, item := range mockData {
		if i >= count {
			break
		}

		hotSearchItems = append(hotSearchItems, inout.BaiduHotSearchItem{
			Rank:       i + 1,
			Title:      item.title,
			HotValue:   item.hotValue,
			Link:       fmt.Sprintf("https://www.baidu.com/s?wd=%s", item.title),
			Tag:        item.tag,
			Desc:       item.desc,
			ImageUrl:   "",
			UpdateTime: time.Now().Format("2006-01-02 15:04:05"),
		})
	}

	// 构建响应
	response := &inout.BaiduHotSearchResp{
		Code:       0,
		Message:    "获取成功（备用数据）",
		Data:       hotSearchItems,
		Total:      len(hotSearchItems),
		UpdateTime: time.Now().Format("2006-01-02 15:04:05"),
	}

	return response, nil
}

// getHotTag 根据热度值获取标签
func (s *BaiduHotSearchService) getHotTag(hotScore int) string {
	if hotScore >= 3000000 {
		return "爆"
	} else if hotScore >= 2000000 {
		return "热"
	} else if hotScore >= 1000000 {
		return "新"
	}
	return ""
}

// 实例化服务
var BaiduHotSearchSvc = &BaiduHotSearchService{}
