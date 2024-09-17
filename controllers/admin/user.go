package admin

import (
	"github.com/gin-gonic/gin"
	"nasa-go-admin/inout"
	"nasa-go-admin/services/admin_service"
)

var tenantsService = &admin_service.TenantsService{}

// register
func TenantsRegister(c *gin.Context) {
	var params inout.AddTenantsReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	// Check if the phone number already exists
	if tenantsService.UserExists(params.Username) {
		Resp.Err(c, 20002, "名称已存在，请修改")
		return
	}
	// Create new user
	newTanants, err := tenantsService.CreateUser(params.Username, params.Password, params.Phone, params.RoleId)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, newTanants)

}

func TenantsLogin(c *gin.Context) {
	var params inout.LoginTenantsReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	// Check if the phone number already exists
	user, err := tenantsService.LoginTenants(c, params.Username, params.Password)
	if err != nil {
		//Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, user)
}

// login
func Login(c *gin.Context) {

	var params inout.LoginAdminReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	// Check if the phone number already exists
	user, err := tenantsService.Login(c, params.Username, params.Password)
	if err != nil {
		//Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, user)

}

// GetRoute
func GetRoute(c *gin.Context) {

}
