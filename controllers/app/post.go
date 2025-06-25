package app

import (
	"nasa-go-admin/db"
	"nasa-go-admin/inout"
	"nasa-go-admin/model/app_model"
	"nasa-go-admin/pkg/response"
	"nasa-go-admin/services/app_service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreatePost 创建帖子
func CreatePost(c *gin.Context) {
	var req inout.CreatePostReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.INVALID_PARAMS, err.Error())
		return
	}

	// 获取当前用户ID
	userID := c.GetUint("user_id")
	if userID == 0 {
		response.Error(c, response.INVALID_PARAMS, "用户未登录")
		return
	}

	// 内容过滤
	filter := &app_service.ContentFilter{
		Title:   req.Title,
		Content: req.Content,
	}

	filterResult, err := filter.Filter()
	if err != nil {
		response.Error(c, response.ERROR, "内容审核失败")
		return
	}

	// 创建帖子
	post := &app_model.UserPost{
		UserID:  userID,
		Title:   req.Title,
		Content: req.Content,
		Images:  app_model.StringArray(req.Images),
	}

	// 如果包含敏感词，设置状态为已拒绝
	if filterResult.HasSensitiveWords {
		post.Status = app_model.PostStatusRejected
		post.RejectReason = filterResult.RejectReason
	} else {
		post.Status = app_model.PostStatusApproved
	}

	if err := db.Dao.Create(post).Error; err != nil {
		response.Error(c, response.ERROR, "创建帖子失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "创建成功",
		"data": post,
	})
}

// UpdatePost 更新帖子
func UpdatePost(c *gin.Context) {
	var req inout.UpdatePostReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.INVALID_PARAMS, err.Error())
		return
	}

	// 获取当前用户ID
	userID := c.GetUint("user_id")
	if userID == 0 {
		response.Error(c, response.INVALID_PARAMS, "用户未登录")
		return
	}

	// 查找并更新帖子
	post := &app_model.UserPost{}
	if err := db.Dao.Where("id = ? AND user_id = ?", req.ID, userID).First(post).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, response.ERROR, "帖子不存在或无权限")
		} else {
			response.Error(c, response.ERROR, "查询帖子失败")
		}
		return
	}

	// 内容过滤
	filter := &app_service.ContentFilter{
		Title:   req.Title,
		Content: req.Content,
	}

	filterResult, err := filter.Filter()
	if err != nil {
		response.Error(c, response.ERROR, "内容审核失败")
		return
	}

	// 更新帖子
	post.Title = req.Title
	post.Content = req.Content
	post.Images = app_model.StringArray(req.Images)

	// 如果包含敏感词，设置状态为已拒绝
	if filterResult.HasSensitiveWords {
		post.Status = app_model.PostStatusRejected
		post.RejectReason = filterResult.RejectReason
	} else {
		post.Status = app_model.PostStatusApproved
	}

	if err := db.Dao.Save(post).Error; err != nil {
		response.Error(c, response.ERROR, "更新帖子失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "更新成功",
		"data": post,
	})
}

// DeletePost 删除帖子
func DeletePost(c *gin.Context) {
	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, response.INVALID_PARAMS, "无效的帖子ID")
		return
	}

	// 获取当前用户ID
	userID := c.GetUint("user_id")
	if userID == 0 {
		response.Error(c, response.INVALID_PARAMS, "用户未登录")
		return
	}

	// 删除帖子
	result := db.Dao.Where("id = ? AND user_id = ?", postID, userID).Delete(&app_model.UserPost{})
	if result.Error != nil {
		response.Error(c, response.ERROR, "删除帖子失败")
		return
	}
	if result.RowsAffected == 0 {
		response.Error(c, response.ERROR, "帖子不存在或无权限")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "删除成功",
	})
}

// GetPostList 获取帖子列表
func GetPostList(c *gin.Context) {
	var req inout.PostListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, response.INVALID_PARAMS, err.Error())
		return
	}

	// 构建查询
	query := db.Dao.Model(&app_model.UserPost{}).Where("status = ?", app_model.PostStatusApproved)
	if req.UserID > 0 {
		query = query.Where("user_id = ?", req.UserID)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Error(c, response.ERROR, "获取帖子总数失败")
		return
	}

	// 获取分页数据
	var posts []app_model.UserPost
	offset := (req.Page - 1) * req.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(req.PageSize).Find(&posts).Error; err != nil {
		response.Error(c, response.ERROR, "获取帖子列表失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "获取成功",
		"data": gin.H{
			"total": total,
			"list":  posts,
		},
	})
}

// GetPostDetail 获取帖子详情
func GetPostDetail(c *gin.Context) {
	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, response.INVALID_PARAMS, "无效的帖子ID")
		return
	}

	var post app_model.UserPost
	if err := db.Dao.First(&post, postID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, response.ERROR, "帖子不存在")
		} else {
			response.Error(c, response.ERROR, "获取帖子详情失败")
		}
		return
	}

	// 如果帖子未通过审核，且不是作者本人，则不允许查看
	if post.Status != app_model.PostStatusApproved {
		userID := c.GetUint("user_id")
		if userID != post.UserID {
			response.Error(c, response.ERROR, "帖子不存在")
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "获取成功",
		"data": post,
	})
}
