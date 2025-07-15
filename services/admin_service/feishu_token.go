package admin_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"nasa-go-admin/model/admin_model"
	"net/http"
	"sync"
	"time"
)

var (
	token     string
	tokenLock sync.RWMutex
)

func init() {
	go refreshToken()
}

func refreshToken() {
	for {
		newToken, err := getToken()
		if err != nil {
			log.Printf("Failed to refresh token: %v", err)
			time.Sleep(1 * time.Minute) // 如果获取 token 失败，1 分钟后重试
			continue
		}

		tokenLock.Lock()
		token = newToken
		tokenLock.Unlock()

		time.Sleep(1 * time.Hour) // 每小时刷新一次 token
	}
}

func getToken() (string, error) {
	// 创建请求体
	requestBody := admin_model.FeishuRequest{
		AppID:     "cli_a8fa4c584eba900c",
		AppSecret: "rIaa5LzmOdorwlX8pt5rSuEE37zIM8Zn",
	}

	// 将请求体转换为 JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	// 发送 POST 请求
	resp, err := http.Post("https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// 解析响应体
	var feishuResp admin_model.FeishuResponse
	err = json.Unmarshal(body, &feishuResp)
	if err != nil {
		return "", err
	}

	// 检查获取 token 是否成功
	if feishuResp.Code != 0 {
		return "", fmt.Errorf("failed to get token: %s", feishuResp.Msg)
	}

	return feishuResp.TenantAccessToken, nil
}

func GetToken() string {
	tokenLock.RLock()
	defer tokenLock.RUnlock()
	return token
}
