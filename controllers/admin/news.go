package admin

import (
	"nasa-go-admin/inout"
	"nasa-go-admin/services/admin_service"
	"strconv"

	"github.com/gin-gonic/gin"
)

var newsService = &admin_service.NewsService{}

func AddNews(c *gin.Context) {
	var req inout.AddNewsReq
	if err := c.ShouldBind(&req); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	id, err := newsService.AddNews(c, req)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, id)
}

func UpdateNews(c *gin.Context) {
	var req inout.UpdateNewsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	if err := newsService.UpdateNews(c, req); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, nil)
}

func GetNewsList(c *gin.Context) {
	// 修改：现在返回百度热搜数据
	var req inout.GetBaiduHotSearchReq
	if err := c.ShouldBind(&req); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	// 设置默认获取20条数据
	count := req.Count
	if count <= 0 {
		count = 20
	}

	// 调用百度热搜服务获取数据
	data, err := admin_service.BaiduHotSearchSvc.GetBaiduHotSearch(c, count)
	if err != nil {
		Resp.Err(c, 20001, "获取百度热搜失败: "+err.Error())
		return
	}

	Resp.Succ(c, data)
}

func GetNewsDetail(c *gin.Context) {
	var req inout.GetNewsDetailReq
	if err := c.ShouldBind(&req); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	data, err := newsService.GetNewsDetail(c, req)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, data)
}

func DeleteNews(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		Resp.Err(c, 20001, "id参数错误")
		return
	}
	if err := newsService.DeleteNews(c, id); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, nil)
}
