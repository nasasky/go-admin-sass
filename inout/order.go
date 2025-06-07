package inout

type OrderListReq struct {
	Page     int    `form:"page" binding:"required"` // 页码
	PageSize int    `form:"page_size" `              // 每页数量
	Status   string `form:"status"`                  // 订单状态
	Search   string `form:"search"`
	No       string `form:"no"`       // 订单号
	UserId   int    `form:"user_id"`  // 用户ID
	Category int    `form:"category"` // 商品分类
}

type OrderListResp struct {
	Items    []OrderListItem `json:"items"`     // 订单列表
	Total    int64           `json:"total"`     // 总记录数
	Page     int             `json:"page"`      // 当前页码
	PageSize int             `json:"page_size"` // 每页数量

}

type OrderListItem struct {
	Id         int     `json:"id"`          // 订单ID
	No         string  `json:"no"`          // 订单号
	GoodsId    int     `json:"goods_id"`    // 商品ID
	GoodsName  string  `json:"goods_name"`  // 商品名称 (新增)
	GoodsPrice float64 `json:"goods_price"` // 商品价格 (新增)
	GoodsCover string  `json:"goods_cover"` // 商品封面图 (新增)
	Amount     string  `json:"amount"`      // 订单金额
	Status     string  `json:"status"`      // 订单状态
	Num        int     `json:"num"`         // 商品数量
	UserId     int     `json:"user_id"`     // 用户ID
	UserName   string  `json:"user_name"`   // 用户名
	UserAvatar string  `json:"user_avatar"` // 用户头像
	UserPhone  string  `json:"user_phone"`  // 用户手机号
	CreateTime string  `json:"create_time"` // 创建时间
	UpdateTime string  `json:"update_time"` // 更新时间
	CouponId   int     `json:"coupon_id"`   // 优惠券ID
}
