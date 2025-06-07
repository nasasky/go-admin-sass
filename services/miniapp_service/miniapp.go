package miniapp_service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"nasa-go-admin/db"
	"nasa-go-admin/inout"
	"nasa-go-admin/middleware"
	"nasa-go-admin/model/admin_model"
	"nasa-go-admin/model/miniapp_model"
	"nasa-go-admin/redis"
	"net/http"
	"sync"
	"time"
)

type MiniappService struct {
}

// 微信小程序配置，实际应用中应从配置文件读取
const (
	AccessTokenKey      = "wx:miniapp:access_token"
	RefreshBeforeExpire = 300 // 提前5分钟刷新
	WXConfigCacheKey    = "wx:miniapp:config"
	WXConfigCacheExpire = time.Hour * 24 // 缓存24小时
	WechatNotice        = "下单通知"
)

// WXConfig 微信小程序配置
type WXConfig struct {
	AppID     string
	AppSecret string
}

var (
	wxConfig     *WXConfig
	wxConfigOnce sync.Once
)

// AccessTokenResponse 微信接口返回的数据结构
type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

// SaveUserSubscription 保存用户订阅信息
func SaveUserSubscription(openID, templateID string) error {
	subscription := miniapp_model.UserSubscription{
		OpenID:      openID,
		TemplateID:  templateID,
		SubscribeAt: time.Now(),
		Status:      1, // 1表示已订阅
	}

	// 如果记录已存在，则更新
	var existingSubscription miniapp_model.UserSubscription
	result := db.Dao.Where("open_id = ? AND template_id = ?", openID, templateID).First(&existingSubscription)
	if result.RowsAffected > 0 {
		return db.Dao.Model(&existingSubscription).Updates(map[string]interface{}{
			"subscribe_at": time.Now(),
			"status":       1,
		}).Error
	}

	// 否则创建新记录
	return db.Dao.Create(&subscription).Error
}

// GetWXConfig 从数据库获取微信小程序配置
func GetWXConfig() (*WXConfig, error) {
	var err error
	// 使用Once确保只初始化一次
	wxConfigOnce.Do(func() {
		// 1. 先尝试从Redis缓存获取
		ctx := context.Background()
		var configJSON string
		configJSON, err = redis.GetClient().Get(ctx, WXConfigCacheKey).Result()
		if err == nil && configJSON != "" {
			// 缓存命中，解析JSON
			var config WXConfig
			if err = json.Unmarshal([]byte(configJSON), &config); err == nil {
				wxConfig = &config
				log.Println("从Redis缓存获取微信配置成功")
				return
			}
		}

		// 2. 缓存未命中，从数据库查询
		var settings []admin_model.SettingList
		err = db.Dao.Where("type = ?", "wechat_id").Find(&settings).Error
		if err != nil {
			log.Printf("查询微信小程序配置失败: %v", err)
			return
		}

		// 处理查询结果
		config := &WXConfig{}
		fmt.Println(settings)
		for i, setting := range settings {
			fmt.Printf("Setting %d: Name=%s, Appid=%s, Secret=%s\n",
				i, setting.Name, setting.Appid, setting.Secret)
		}
		// for _, setting := range settings {
		// 	switch setting.Name {
		// 	case "appid":
		// 		config.AppID = setting.Appid
		// 	case "secret":
		// 		config.AppSecret = setting.Secret
		// 	}
		// }
		config.AppID = settings[0].Appid
		config.AppSecret = settings[0].Secret

		fmt.Println("<UNK>", config)
		// 检查配置是否完整
		if config.AppID == "" || config.AppSecret == "" {
			log.Println("警告: 微信小程序配置不完整")
			return
		}

		// 3. 将配置存入Redis缓存
		var jsonBytes []byte
		jsonBytes, err = json.Marshal(config)
		if err != nil {
			log.Printf("序列化微信配置失败: %v", err)
			return
		}

		// 将[]byte显式转换为string
		err = redis.GetClient().Set(ctx, WXConfigCacheKey, string(jsonBytes), WXConfigCacheExpire).Err()
		if err != nil {
			log.Printf("缓存微信配置失败: %v", err)
			return
		}

		wxConfig = config
		log.Println("从数据库获取微信配置成功")
	})

	if wxConfig == nil || wxConfig.AppID == "" || wxConfig.AppSecret == "" {
		return nil, fmt.Errorf("未找到有效的微信小程序配置")
	}

	return wxConfig, nil
}

// refreshWXConfig 刷新微信配置缓存（可在配置更新后调用）
func RefreshWXConfig() {
	// 删除Redis缓存
	redis.GetClient().Del(context.Background(), WXConfigCacheKey)
	// 重置单例
	wxConfig = nil
	wxConfigOnce = sync.Once{}
	log.Println("微信配置缓存已刷新")
}

// GetAccessToken 获取微信小程序AccessToken（带缓存）
func GetAccessToken() (string, error) {
	ctx := context.Background()

	// 1. 尝试从Redis获取缓存的AccessToken
	token, err := redis.GetClient().Get(ctx, AccessTokenKey).Result()
	if err == nil && token != "" {
		// 找到了有效的token，直接返回
		log.Println("使用缓存的AccessToken")
		return token, nil
	}

	// 2. 获取微信配置
	config, err := GetWXConfig()
	if err != nil {
		return "", fmt.Errorf("获取微信配置失败: %w", err)
	}

	// 3. 从微信服务器获取新的AccessToken
	log.Println("从微信服务器获取新的AccessToken")
	accessToken, expiresIn, err := fetchAccessTokenFromWX(config.AppID, config.AppSecret)
	if err != nil {
		return "", fmt.Errorf("获取AccessToken失败: %w", err)
	}

	// 4. 将新获取的AccessToken存入Redis
	expireDuration := time.Duration(expiresIn-RefreshBeforeExpire) * time.Second
	err = redis.GetClient().Set(ctx, AccessTokenKey, accessToken, expireDuration).Err()
	if err != nil {
		log.Printf("警告: AccessToken缓存到Redis失败: %v", err)
		// 缓存失败不影响返回，但记录日志
	}

	return accessToken, nil
}

// fetchAccessTokenFromWX 从微信服务器获取新的AccessToken
func fetchAccessTokenFromWX(appID, appSecret string) (string, int, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", appID, appSecret)

	// 设置超时
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", 0, fmt.Errorf("请求微信服务器失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("读取响应失败: %w", err)
	}

	var result AccessTokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", 0, fmt.Errorf("解析响应失败: %w", err)
	}

	if result.ErrCode != 0 {
		return "", 0, fmt.Errorf("微信服务器返回错误: %s", result.ErrMsg)
	}

	// 日志记录获取了新的token
	log.Printf("成功从微信服务器获取新的AccessToken，有效期: %d秒", result.ExpiresIn)

	return result.AccessToken, result.ExpiresIn, nil
}

// SendSubscribeMsg 发送订阅消息
func SendSubscribeMsg(openID, templateID string, id string) error {
	accessToken, err := GetAccessToken()
	if err != nil {
		return err
	}

	// 组装推送数据，根据模板ID区分不同类型的推送内容
	data := generateMsgData(templateID)

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/subscribe/send?access_token=%s", accessToken)
	payload := map[string]interface{}{
		"touser":      openID,
		"template_id": templateID,
		"page":        "pages/index/index", // 用户点击消息后跳转的页面
		"data":        data,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result inout.WxMsgResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	if result.ErrCode != 0 {
		return fmt.Errorf("发送订阅消息失败: %s", result.ErrMsg)
	}

	// 记录推送历史
	err = recordPushHistory(openID, templateID, data)
	if err != nil {
		log.Printf("记录推送历史失败: %v", err)
		// 不影响主流程，继续执行
	}

	return nil
}

// generateMsgData 根据模板ID生成不同的推送内容
func generateMsgData(templateID string) map[string]map[string]string {
	// 根据不同的模板ID生成不同的内容
	switch templateID {
	//新活动发布提醒
	case "订单通知模板ID":
		return map[string]map[string]string{
			"character_string1": {"value": "OD" + fmt.Sprintf("%d", time.Now().Unix())},
			"thing2":            {"value": "您的订单已发货"},
			"amount3":           {"value": "99.00"},
			"date4":             {"value": time.Now().Format("2006-01-02 15:04:05")},
			"thing5":            {"value": "感谢您的购买！"},
		}
	case "FL4Qq5zBk5zpXs1Jkd7F8D_STgGm9PcdSqOkZnegm2g":
		return map[string]map[string]string{
			"thing6":  {"value": "限时优惠活动"},
			"date2":   {"value": time.Now().Format("2006-01-02 15:04:05")},
			"thing4":  {"value": "全场商品8折起"},
			"amount3": {"value": "线上商城"},
			"date8":   {"value": time.Now().Format("2006-01-02 15:04:05")},
		}
	default:
		// 默认推送内容
		return map[string]map[string]string{
			"thing1": {"value": "系统通知"},
			"time2":  {"value": time.Now().Format("2006-01-02 15:04:05")},
			"thing3": {"value": "有新的系统消息，请查看"},
		}
	}
}

// recordPushHistory 记录推送历史
func recordPushHistory(openID, templateID string, data map[string]map[string]string) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

	history := miniapp_model.PushHistory{
		OpenID:     openID,
		TemplateID: templateID,
		Content:    string(dataJSON),
		PushTime:   time.Now(),
		Status:     1, // 1表示推送成功
	}
	go middleware.LogWechatEvent("wechat_plus", 0, openID, map[string]interface{}{
		"type":        WechatNotice,
		"content":     string(dataJSON),
		"time":        time.Now(),
		"status":      1,
		"openid":      openID,
		"template_id": templateID,
	})
	fmt.Println("PushHistory:", history)
	return db.Dao.Create(&history).Error
}
