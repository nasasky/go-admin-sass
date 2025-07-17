package admin_service

import (
	"log"
	"nasa-go-admin/db"
	"nasa-go-admin/inout"
	"nasa-go-admin/model/admin_model"
	"time"

	"github.com/gin-gonic/gin"
)

type NewsService struct{}

func (s *NewsService) AddNews(c *gin.Context, req inout.AddNewsReq) (int, error) {
	data := admin_model.News{
		Title:       req.Title,
		Description: req.Description,
		Content:     req.Content,
		CoverImage:  req.CoverImage,
		Sort:        req.Sort,
		Status:      req.Status,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}
	err := db.Dao.Create(&data).Error
	if err != nil {
		return 0, err
	}
	return data.Id, nil
}

func (s *NewsService) UpdateNews(c *gin.Context, req inout.UpdateNewsReq) error {
	// 调试：输出接收到的参数
	log.Printf("UpdateNews - 接收到的参数: %+v", req)

	// 先查询现有数据
	var existingNews admin_model.News
	err := db.Dao.Where("id = ?", req.Id).First(&existingNews).Error
	if err != nil {
		log.Printf("UpdateNews - 查询现有数据失败: %v", err)
		return err
	}

	log.Printf("UpdateNews - 查询到的现有数据: %+v", existingNews)

	// 更新非空字段
	if req.Title != "" {
		existingNews.Title = req.Title
		log.Printf("UpdateNews - 更新title: %s", req.Title)
	}
	if req.Description != "" {
		existingNews.Description = req.Description
		log.Printf("UpdateNews - 更新description: %s", req.Description)
	}
	if req.Content != "" {
		existingNews.Content = req.Content
		log.Printf("UpdateNews - 更新content: %s", req.Content)
	}
	if req.CoverImage != "" {
		existingNews.CoverImage = req.CoverImage
		log.Printf("UpdateNews - 更新cover_image: %s", req.CoverImage)
	}
	if req.Sort != 0 {
		existingNews.Sort = req.Sort
		log.Printf("UpdateNews - 更新sort: %d", req.Sort)
	}
	if req.Status != 0 {
		existingNews.Status = req.Status
		log.Printf("UpdateNews - 更新status: %d", req.Status)
	}

	// 更新更新时间
	existingNews.UpdateTime = time.Now()

	log.Printf("UpdateNews - 更新后的数据: %+v", existingNews)

	// 保存更新
	err = db.Dao.Save(&existingNews).Error
	if err != nil {
		log.Printf("UpdateNews - 保存失败: %v", err)
		return err
	}

	log.Printf("UpdateNews - 保存成功")
	return nil
}

func (s *NewsService) GetNewsList(c *gin.Context, req inout.GetNewsListReq) (interface{}, error) {
	var data []admin_model.News
	var total int64
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	offset := (req.Page - 1) * req.PageSize
	query := db.Dao.Model(&admin_model.News{})
	if req.Search != "" {
		query = query.Where("title LIKE ? OR description LIKE ?", "%"+req.Search+"%", "%"+req.Search+"%")
	}
	err := query.Count(&total).Offset(offset).Limit(req.PageSize).Order("sort desc, id desc").Find(&data).Error
	if err != nil {
		return nil, err
	}
	items := make([]inout.NewsItem, len(data))
	for i, item := range data {
		items[i] = inout.NewsItem{
			Id:          item.Id,
			Title:       item.Title,
			Description: item.Description,
			Content:     item.Content,
			CoverImage:  item.CoverImage,
			Sort:        item.Sort,
			Status:      item.Status,
			CreateTime:  item.CreateTime.Format("2006-01-02 15:04:05"),
			UpdateTime:  item.UpdateTime.Format("2006-01-02 15:04:05"),
		}
	}
	return inout.NewsListResponse{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Items:    items,
	}, nil
}

func (s *NewsService) GetNewsDetail(c *gin.Context, req inout.GetNewsDetailReq) (interface{}, error) {
	var data admin_model.News
	err := db.Dao.Where("id = ?", req.Id).First(&data).Error
	if err != nil {
		return nil, err
	}
	return inout.NewsItem{
		Id:          data.Id,
		Title:       data.Title,
		Description: data.Description,
		Content:     data.Content,
		CoverImage:  data.CoverImage,
		Sort:        data.Sort,
		Status:      data.Status,
		CreateTime:  data.CreateTime.Format("2006-01-02 15:04:05"),
		UpdateTime:  data.UpdateTime.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *NewsService) DeleteNews(c *gin.Context, id int) error {
	return db.Dao.Where("id = ?", id).Delete(&admin_model.News{}).Error
}
