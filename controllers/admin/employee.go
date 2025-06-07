package admin

import (
	"nasa-go-admin/inout"
	"nasa-go-admin/services/admin_service"
	"strconv"

	"github.com/gin-gonic/gin"
)

var employeeService = &admin_service.EmployeeService{}

// AddEmployee
func AddEmployee(c *gin.Context) {
	var params inout.AddEmployeeReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	id, err := employeeService.AddEmployee(c, params)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, id)
}

// GetEmployeeList
func GetEmployeeList(c *gin.Context) {
	var params inout.ListpageReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return

	}
	list, err := employeeService.GetEmployeeList(c, params)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, list)

}

// UpdateEmployee
func UpdateEmployee(c *gin.Context) {
	var params inout.UpdateEmployeeReq
	if err := c.ShouldBindJSON(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	id, err := employeeService.UpdateEmployee(c, params)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, id)
}

// DeleteEmployee
func DeleteEmployee(c *gin.Context) {

	var params struct {
		Ids []int `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	if len(params.Ids) == 0 {
		Resp.Err(c, 20001, "ids不能为空")
		return
	}
	err := employeeService.DeleteEmployee(c, params.Ids)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return

	}
	Resp.Succ(c, nil)

}

// GetEmployeeDetail
func GetEmployeeDetail(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		Resp.Err(c, 20001, "id不能为空")
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	detail, err := employeeService.GetEmployeeDetail(c, id)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	Resp.Succ(c, detail)
}
