package inout

type AddNewsReq struct {
	Title       string `form:"title" binding:"required"`
	Description string `form:"description"`
	Content     string `form:"content" binding:"required"`
	CoverImage  string `form:"cover_image"`
	Sort        int    `form:"sort"`
	Status      int    `form:"status"`
}

type UpdateNewsReq struct {
	Id          int    `json:"id" binding:"required"`
	Title       string `json:"title" binding:"omitempty"`
	Description string `json:"description" binding:"omitempty"`
	Content     string `json:"content" binding:"omitempty"`
	CoverImage  string `json:"cover_image" binding:"omitempty"`
	Sort        int    `json:"sort" binding:"omitempty"`
	Status      int    `json:"status" binding:"omitempty"`
}

type NewsItem struct {
	Id          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Content     string `json:"content"`
	CoverImage  string `json:"cover_image"`
	Sort        int    `json:"sort"`
	Status      int    `json:"status"`
	CreateTime  string `json:"create_time"`
	UpdateTime  string `json:"update_time"`
}

type NewsListResponse struct {
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
	Items    []NewsItem `json:"items"`
}

type GetNewsListReq struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Search   string `form:"search"`
}

type GetNewsDetailReq struct {
	Id int `form:"id" binding:"required"`
}
