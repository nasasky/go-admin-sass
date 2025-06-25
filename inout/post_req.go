package inout

// CreatePostReq 创建帖子请求
type CreatePostReq struct {
	Title   string   `json:"title" binding:"required,max=200"` // 标题
	Content string   `json:"content" binding:"required"`       // 内容
	Images  []string `json:"images" binding:"omitempty,max=9"` // 图片URL数组，最多9张图片
}

// UpdatePostReq 更新帖子请求
type UpdatePostReq struct {
	ID      uint     `json:"id" binding:"required"`            // 帖子ID
	Title   string   `json:"title" binding:"required,max=200"` // 标题
	Content string   `json:"content" binding:"required"`       // 内容
	Images  []string `json:"images" binding:"omitempty,max=9"` // 图片URL数组，最多9张图片
}

// PostListReq 获取帖子列表请求
type PostListReq struct {
	Page     int  `form:"page" binding:"required,min=1"`             // 页码
	PageSize int  `form:"page_size" binding:"required,min=1,max=50"` // 每页数量
	UserID   uint `form:"user_id" binding:"omitempty"`               // 可选的用户ID过滤
}
