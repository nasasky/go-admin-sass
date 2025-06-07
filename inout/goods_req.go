package inout

type AddGoodsReq struct {
	// 商品名称
	GoodsName string `form:"goods_name" binding:"required"`
	// 商品描述
	Content string `form:"content" binding:"required"`
	// 商品价格
	Price float64 `form:"price" binding:"required"`
	// 商品库存
	Stock int `form:"stock" binding:"required"`
	// 商品图片
	//Cover string `form:"cover" binding:"required"`
	//// 商品状态
	Status string `form:"status" binding:"required"`
	//// 商品分类
	//CategoryId int `form:"category_id" binding:"required"`
}

type UpdateGoodsReq struct {
	Id         int     `json:"id" binding:"required"`
	UserId     int     `json:"user_id"`
	GoodsName  string  `json:"goods_name" binding:"required"`
	Price      float64 `json:"price"`
	Content    string  `json:"content"`
	Cover      string  `json:"cover"`
	Status     string  `json:"status"`
	CategoryId int     `json:"category_id"`
	TenantsId  int     `json:"tenants_id"`
	Stock      int     `json:"stock"`
	IsDelete   int     `json:"isdelete"`
	CreateTime string  `json:"create_time"`
	UpdateTime string  `json:"update_time"`
}

type GetGoodsListReq struct {
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
	GoodsName  string `form:"goods_name"`
	Status     string `form:"status"`
	CategoryId int    `form:"category_id"`
}

type GoodsItem struct {
	// 商品id
	Id int `json:"id"`
	// 商品名称
	GoodsName string `json:"goods_name"`
	// 商品描述
	Content string `json:"content"`
	// 商品价格
	Price float64 `json:"price"`
	// 商品库存
	Stock int `json:"stock"`
	// 商品图片
	Cover string `json:"cover"`
	// 商品状态
	Status string `json:"status"`
	// 商品分类
	CategoryId int `json:"category_id"`
	// 创建时间
	CreateTime string `json:"create_time"`
	// 更新时间
	UpdateTime string `json:"update_time"`
}

type GetGoodsListResp struct {
	// 总数
	Total int64 `json:"total"`
	// 页码
	Page int `json:"page"`
	// 每页数量
	PageSize int `json:"page_size"`
	// 数据
	Items []GoodsItem `json:"items"`
}

// AddGoodsCategoryReq
type AddGoodsCategoryReq struct {
	// 分类名称
	Name string `form:"name" binding:"required"`
	// 分类描述
	Description string `form:"description"`
	// 分类状态
	Status string `form:"status" binding:"required"`
}

type GetGoodsCategoryListReq struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Name     string `form:"name"`
	Status   string `form:"status"`
}

type GetGoodsCategoryListResp struct {
	// 总数
	Total int64 `json:"total"`
	// 页码
	Page int `json:"page"`
	// 每页数量
	PageSize int `json:"page_size"`
	// 数据
	Items []GoodsCategoryItem `json:"items"`
}

type GoodsCategoryItem struct {
	// 分类id
	Id int `json:"id"`
	// 分类名称
	Name string `json:"name"`
	// 分类描述
	Description string `json:"description"`
	// 分类状态
	Status string `json:"status"`
	// 创建时间
	CreateTime string `json:"create_time"`
	// 更新时间
	UpdateTime string `json:"update_time"`
	// 是否删除
	IsDelete int `json:"is_delete"`
	// 租户id
	TenantsId int `json:"tenants_id"`
}

type UpdateGoodsCategoryReq struct {
	Id          int    `json:"id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Status      string `json:"status" binding:"required"`
	TenantsId   int    `json:"tenants_id"`
	IsDelete    int    `json:"is_delete"`
	CreateTime  string `json:"create_time"`
	UpdateTime  string `json:"update_time"`
}
