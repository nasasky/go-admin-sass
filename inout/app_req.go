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
