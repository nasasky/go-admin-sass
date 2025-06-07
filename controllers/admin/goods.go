package admin

import (
	"fmt"
	"nasa-go-admin/inout"
	"nasa-go-admin/model/admin_model"
	"nasa-go-admin/services/admin_service"
	"nasa-go-admin/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

var goodsService = &admin_service.GoodsService{}

// AddGoods 添加商品
func AddGoods(c *gin.Context) {
	var params inout.AddGoodsReq
	fmt.Println(params)
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	// 调用公共函数获取 parent_id
	parentId, err := utils.GetParentId(c)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	fmt.Println(params)
	var goods = admin_model.Goods{
		GoodsName: params.GoodsName,
		Content:   params.Content,
		Price:     params.Price,
		Stock:     params.Stock,
		TenantsId: parentId,
		//Cover:      params.Cover,
		Status: params.Status,
		//CategoryId: params.CategoryId,
	}

	Id, err := goodsService.AddGoods(c, goods)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, Id)

}

// GetGoodsList 获取商品列表
func GetGoodsList(c *gin.Context) {
	var params inout.GetGoodsListReq

	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	// 调用公共函数获取 parent_id
	parentId, err := utils.GetParentId(c)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	fmt.Println("parentId:", parentId)
	list, err := goodsService.GetGoodsList(c, params)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, list)
}

// UpdateGoods 更新商品
func UpdateGoods(c *gin.Context) {

	var params inout.UpdateGoodsReq
	if err := c.ShouldBindJSON(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	var goods = admin_model.Goods{
		Id:         params.Id,
		GoodsName:  params.GoodsName,
		Content:    params.Content,
		Price:      params.Price,
		Stock:      params.Stock,
		Cover:      params.Cover,
		Status:     params.Status,
		CategoryId: params.CategoryId,
	}
	id, err := goodsService.UpdateGoods(c, goods)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, id)
}

// GetGoodsDetail 获取商品详情
// GetGoodsDetail 获取商品详情
func GetGoodsDetail(c *gin.Context) {
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
	detail, err := goodsService.GetGoodsDetail(c, id)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, detail)
}

// DeleteGoods 删除商品
// DeleteGoods 删除商品
func DeleteGoods(c *gin.Context) {
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
	err := goodsService.DeleteGoods(c, params.Ids)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, nil)
}

// 添加商品分类
func AddGoodsCategory(c *gin.Context) {
	var params inout.AddGoodsCategoryReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	// 调用公共函数获取 parent_id
	parentId, err := utils.GetParentId(c)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	category := admin_model.GoodsCategory{
		Name:        params.Name,
		TenantsId:   parentId,
		Status:      params.Status,
		Description: params.Description,
	}
	id, err := goodsService.AddGoodsCategory(c, category)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, id)
}

// GetGoodsCategoryList 获取商品分类列表
func GetGoodsCategoryList(c *gin.Context) {
	var params inout.GetGoodsCategoryListReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	list, err := goodsService.GetGoodsCategoryList(c, params)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, list)
}

// GetGoodsCategoryDetail 获取商品分类详情
func GetGoodsCategoryDetail(c *gin.Context) {
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
	detail, err := goodsService.GetGoodsCategoryDetail(c, id)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, detail)
}

// UpdateGoodsCategory 更新商品分类
func UpdateGoodsCategory(c *gin.Context) {
	var params inout.UpdateGoodsCategoryReq
	if err := c.ShouldBindJSON(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	category := admin_model.GoodsCategory{
		Id:          params.Id,
		Name:        params.Name,
		Description: params.Description,
		Status:      params.Status,
	}
	id, err := goodsService.UpdateGoodsCategory(c, category)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, id)
}

// DeleteGoodsCategory 删除商品分类
func DeleteGoodsCategory(c *gin.Context) {
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
	err := goodsService.DeleteGoodsCategory(c, params.Ids)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, nil)
}
