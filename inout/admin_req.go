package inout

type AddTenantsReq struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
	Phone    string `form:"phone" binding:"required"`
	RoleId   int    `form:"role_id" binding:"required"`
}

type LoginAdminReq struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
	Captcha  string `form:"captcha"`
}

type LoginTenantsReq struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
}
