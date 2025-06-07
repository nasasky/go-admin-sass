package inout

type AddUserAppReq struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
	Phone    string `form:"phone" binding:"required"`
}

type LoginAppReq struct {
	Password string `form:"password" binding:"required"`
	Phone    string `form:"phone" binding:"required"`
}

type RechargeReq struct {
	Amount float64 `form:"amount" binding:"required"`
}

type CreateOrderReq struct {
	GoodsId int `form:"goods_id" binding:"required"`
	Num     int `form:"num" binding:"required"`
}

type MyOrderReq struct {
	Page int `json:"page"`
	// 每页数量
	PageSize int `json:"page_size"`
	// 数据
	Status string `json:"status"`
}

type MyOrderResp struct {
	Total    int64       `json:"total"`
	List     interface{} `json:"list"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

type OrderItem struct {
	Id         int     `json:"id"`
	UserId     int     `json:"user_id"`
	GoodsId    int     `json:"goods_id"`
	Num        int     `json:"num"`
	Amount     float64 `json:"amount"`
	GoodsName  string  `json:"goods_name"`
	GoodsPrice float64 `json:"goods_price"`
	Status     string  `json:"status"`
	CreateTime string  `json:"create_time"`
	UpdateTime string  `json:"update_time"`
}

type DetailReq struct {
	Id int `json:"id" binding:"required"`
}

type RefundReq struct {
	OrderId int    `form:"order_id" binding:"required"`
	Reason  string `form:"reason" binding:"required"`
}

type UpdateUserAppReq struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Phone    string `json:"phone"`
	NickName string `json:"nick_name"`
	Gender   int    `json:"gender"`
	Address  string `json:"address"`
	Email    string `json:"email"`
}
