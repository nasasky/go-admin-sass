package app_service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"nasa-go-admin/db"
	"nasa-go-admin/model/admin_model"
	"nasa-go-admin/model/app_model"
	"nasa-go-admin/redis"
	"net/http"
	"strconv"
	"sync"
	"time"

	"gorm.io/gorm"
)

type WeChatService struct{}

// WXConfig 微信小程序配置
type WXConfig struct {
	AppID     string
	AppSecret string
}

// 微信小程序配置，实际应用中应从配置文件读取
const (
	AccessTokenKey      = "wx:miniapp:access_token"
	RefreshBeforeExpire = 300 // 提前5分钟刷新
	WXConfigCacheKey    = "wx:miniapp:config"
	WXConfigCacheExpire = time.Hour * 24 // 缓存24小时
)

var (
	wxConfig     *WXConfig
	wxConfigOnce sync.Once
)

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

// Code2SessionResponse 微信登录凭证校验响应
type Code2SessionResponse struct {
	OpenID     string `json:"openid"`      // 用户唯一标识
	SessionKey string `json:"session_key"` // 会话密钥
	UnionID    string `json:"unionid"`     // 用户在开放平台的唯一标识符
	ErrCode    int    `json:"errcode"`     // 错误码
	ErrMsg     string `json:"errmsg"`      // 错误信息
}

// WxLogin 微信小程序登录
func (w *WeChatService) WxLogin(params interface{}) (map[string]interface{}, error) {
	// 解析参数
	var reqParams struct {
		Code string `json:"code"`
	}

	// 将 params 转换为结构体
	paramsBytes, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("参数格式错误: %v", err)
	}

	if err := json.Unmarshal(paramsBytes, &reqParams); err != nil {
		return nil, fmt.Errorf("解析参数失败: %v", err)
	}

	if reqParams.Code == "" {
		return nil, fmt.Errorf("缺少必要参数: code")
	}

	// 获取微信小程序配置
	config, err := GetWXConfig()
	if err != nil {
		return nil, fmt.Errorf("获取微信配置失败: %v", err)
	}

	// 构建请求URL
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		config.AppID, config.AppSecret, reqParams.Code,
	)

	// 发送HTTP请求
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求微信API失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析响应JSON
	var result Code2SessionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查错误码
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("微信登录失败: %s (错误码: %d)", result.ErrMsg, result.ErrCode)
	}

	// 记录用户登录信息，并获取用户完整信息
	userInfo, err := w.saveUserLoginInfo(result.OpenID, result.UnionID)
	if err != nil {
		log.Printf("保存用户登录信息失败: %v", err)
		// 继续执行，至少返回openid
	}

	// 构建返回数据
	responseData := map[string]interface{}{
		"openid": result.OpenID,
	}

	// 如果有unionid，也一并返回
	if result.UnionID != "" {
		responseData["unionid"] = result.UnionID
	}

	// 如果成功获取到用户信息，添加到返回数据中
	if userInfo != nil {
		responseData["user"] = userInfo["user"]
		responseData["token"] = userInfo["token"]
	}

	// 直接返回 map，无需序列化
	return responseData, nil
}

// saveUserLoginInfo 保存用户登录信息到数据库并返回用户信息
func (w *WeChatService) saveUserLoginInfo(openID, unionID string) (map[string]interface{}, error) {
	// 检查用户是否已存在
	var user app_model.AppProfile
	result := db.Dao.Where("openid = ?", openID).First(&user)

	// 生成用户token
	token := generateToken(openID)

	if result.Error != nil {
		// 用户不存在，创建新用户
		if result.Error == gorm.ErrRecordNotFound {
			// 创建新用户
			newUser := app_model.AppProfile{
				Openid:     openID,
				UnionID:    unionID,
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
				//取openid后面四位作为用户名，然后前面加入用户两个字
				UserName: fmt.Sprintf("用户%s", openID[len(openID)-4:]),
				// 注意：Token字段应该在模型中定义，如果没有需要添加gorm:"-"标记
			}

			// 仅选择数据库中存在的字段创建
			if err := db.Dao.Select("openid", "union_id", "create_time", "update_time", "username").Create(&newUser).Error; err != nil {
				return nil, fmt.Errorf("创建用户失败: %v", err)
			}

			log.Printf("新用户创建成功: %s", openID)

			// 用ID获取创建的用户
			db.Dao.Where("openid = ?", openID).First(&user)
		} else {
			// 其他数据库错误
			return nil, fmt.Errorf("查询用户失败: %v", result.Error)
		}
	}

	// 更新用户的登录时间
	if err := db.Dao.Model(&app_model.AppProfile{}).
		Where("openid = ?", openID).
		Updates(map[string]interface{}{
			"update_time": time.Now(),
			"union_id":    unionID,
		}).Error; err != nil {
		return nil, fmt.Errorf("更新用户信息失败: %v", err)
	}

	// 存储用户Token到Redis
	expiration := time.Hour * 24 // 过期时间为24小时
	userID := strconv.Itoa(int(user.ID))

	// 存储token
	if err := redis.GetClient().Set(context.Background(),
		fmt.Sprintf("user:token:%s", userID),
		token,
		expiration).Err(); err != nil {
		log.Printf("存储用户token失败: %v", err)
	}

	// 存储用户信息
	userInfo := map[string]interface{}{
		"id":      user.ID,
		"openid":  openID,
		"unionid": unionID,
		"token":   token,
	}

	// 如果有额外字段，可以从数据库中的用户对象获取
	if user.UserName != "" {
		userInfo["username"] = user.UserName
	}

	if user.Phone != "" {
		userInfo["phone"] = user.Phone
	}

	// 序列化用户信息
	userInfoJSON, _ := json.Marshal(userInfo)
	if err := redis.GetClient().Set(context.Background(),
		fmt.Sprintf("user:info:%s", userID),
		string(userInfoJSON),
		expiration).Err(); err != nil {
		log.Printf("存储用户信息失败: %v", err)
	}

	// 返回用户信息和token
	return map[string]interface{}{
		"user":  userInfo,
		"token": token,
	}, nil
}

// generateToken 生成用户令牌
func generateToken(openID string) string {
	// 简单实现：时间戳 + openID 的 MD5
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	data := timestamp + openID

	h := md5.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
