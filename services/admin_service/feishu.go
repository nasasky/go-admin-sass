package admin_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"nasa-go-admin/db"
	"nasa-go-admin/model/admin_model"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

type FeishuService struct{}

func (s *FeishuService) AddFeishuGroupList(c *gin.Context, resp admin_model.FeishuGroupListResponse) (int, error) {
	var insertedCount int
	for _, item := range resp.Data.Items {
		// 查询数据库是否存在
		var data admin_model.AddFeishuGroup
		err := db.Dao.Where("chat_id = ?", item.ChatID).First(&data).Error
		if err == nil {
			continue
		}

		newData := admin_model.AddFeishuGroup{ // 确保类型名称正确
			ChatId:  item.ChatID,
			Name:    item.Name,
			Avatar:  item.Avatar,
			OwnerId: item.OwnerID,
		}

		// 将 data 插入数据库
		err = db.Dao.Create(&newData).Error
		if err != nil {
			return insertedCount, err
		}

		// 成功插入，计数加一
		insertedCount++

	}
	return insertedCount, nil
}

// SendFeishuMessage 发送飞书消息
func (s *FeishuService) SendFeishuMessage(c *gin.Context, req admin_model.FeishuMessageRequest) (int, error) {
	// 飞书消息推送https://open.feishu.cn/open-apis/im/v1/messages
	baseURL := "https://open.feishu.cn/open-apis/im/v1/messages"
	// 设置查询参数
	params := url.Values{}
	params.Add("receive_id_type", "chat_id")

	// 构建完整的 URL
	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// 请求头
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + GetToken(),
	}
	// 将 content 转换为 JSON 字符串
	content := map[string]string{
		"text": req.Content,
	}

	contentJSON, err := json.Marshal(content)
	if err != nil {
		log.Printf("Error marshaling content JSON: %v", err)
		return 0, err
	}
	// 请求体
	body := map[string]interface{}{
		"receive_id": req.ReceiveId,
		"msg_type":   req.MsgType,
		"content":    string(contentJSON),
	}

	// 转json
	jsonStr, err := json.Marshal(body)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return 0, err
	}
	// 创建 HTTP 请求
	httpReq, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Printf("Error creating HTTP request: %v", err)
		return 0, err
	}

	// 设置请求头
	for key, value := range headers {
		httpReq.Header.Set(key, value)
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		log.Printf("Error sending HTTP request: %v", err)
		return 0, err
	}
	defer resp.Body.Close()

	// 读取响应体
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return 0, err
	}
	// 记录响应信息
	log.Printf("Response status code: %d", resp.StatusCode)
	log.Printf("Response body: %s", string(bodyBytes))

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to send message, status code: %d", resp.StatusCode)
	}

	return 1, nil
}
