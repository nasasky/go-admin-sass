package inout

type GetRevenueListReq struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Start    string `form:"start"`
	End      string `form:"end"`
	Search   string `form:"search"`
}

type RevenueRepItems struct {
	Id            int     `json:"id"`
	TenantsId     int     `json:"tenants_id"`
	StatDate      string  `json:"stat_date"`
	PeriodStart   string  `json:"period_start"`
	PeriodEnd     string  `json:"period_end"`
	TotalOrders   int     `json:"total_orders"`
	TotalRevenue  float64 `json:"total_revenue"`
	ActualRevenue float64 `json:"actual_revenue"`
	PaidOrders    int     `json:"paid_orders"`
}

type RevenueRep struct {
	Total    int64             `json:"total"`
	Items    []RevenueRepItems `json:"items"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}
