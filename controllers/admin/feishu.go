package admin

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"nasa-go-admin/inout"
	"nasa-go-admin/model/admin_model"
	"nasa-go-admin/services/admin_service"
	"net/http"

	"github.com/gin-gonic/gin"
)

var feishuService = &admin_service.FeishuService{}

//huoqu

// func GetFeishu(c *gin.Context) {

// 	// 创建请求体
// 	requestBody := admin_model.FeishuRequest{
// 		AppID:     "cli_a79e4b283bfad00c",
// 		AppSecret: "7y2O5s1vbw2NZXZJzCLR5C8AsTwskJXG",
// 	}

// 	// 将请求体转换为 JSON
// 	jsonData, err := json.Marshal(requestBody)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// 发送 POST 请求
// 	resp, err := http.Post("https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal", "application/json", bytes.NewBuffer(jsonData))
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	defer resp.Body.Close()

// 	// 读取响应
// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// 解析响应体
// 	var feishuResp admin_model.FeishuResponse
// 	err = json.Unmarshal(body, &feishuResp)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	// 返回响应
// 	//Resp.Succ(c, feishuResp)

// 	GetFeishuGroupList(c, feishuResp.TenantAccessToken)

// }

// 获取飞书群列表
func GetFeishuGroupList(c *gin.Context, Token string) {

	req, err := http.NewRequest("GET", "https://open.feishu.cn/open-apis/im/v1/chats", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 设置请求头
	token := Token // 替换为你的实际 token
	req.Header.Set("Authorization", "Bearer "+token)

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 解析响应体

	var feishuResp admin_model.FeishuGroupListResponse
	err = json.Unmarshal(body, &feishuResp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 把群列表存入数据库
	data, err := feishuService.AddFeishuGroupList(c, feishuResp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回响应
	Resp.Succ(c, data)

}

// GetFeishuGroupList 获取飞书群列表
func GetFeishuGroupListdata(c *gin.Context) {
	// 获取 token
	token := admin_service.GetToken()
	if token == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get token"})
		return
	}

	// 使用获取到的 token 获取飞书群列表
	req, err := http.NewRequest("GET", "https://open.feishu.cn/open-apis/im/v1/chats", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 设置请求头
	req.Header.Set("Authorization", "Bearer "+token)

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 解析响应体

	var feishuGroupResp admin_model.FeishuGroupListResponse
	err = json.Unmarshal(body, &feishuGroupResp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	data, err := feishuService.AddFeishuGroupList(c, feishuGroupResp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回响应
	Resp.Succ(c, data)

}

// SendFeishuMessage 推送飞书机器人消息
func SendFeishuMessage(c *gin.Context) {
	var params inout.FeishuSendReq

	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	fmt.Println(params)
	// 创建请求体
	requestBody := admin_model.FeishuMessageRequest{
		MsgType:   params.MsgType,
		Content:   params.Content,
		ReceiveId: params.ReceiveId,
	}

	Id, err := feishuService.SendFeishuMessage(c, requestBody)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, Id)
}
