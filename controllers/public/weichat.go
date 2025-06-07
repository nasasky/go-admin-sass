package public

import (
	"crypto/sha1"
	"encoding/hex"
	"log"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// 替换为你在微信公众平台/小程序后台配置的 Token
	WECHAT_TOKEN = "VBuQMcmPB5vm0IMxN56uIGKoTapZZyoj"
)

// WechatVerify 处理微信服务器的 Token 验证
func WechatVerify(c *gin.Context) {
	// 获取请求参数
	signature := c.Query("signature")
	timestamp := c.Query("timestamp")
	nonce := c.Query("nonce")
	echostr := c.Query("echostr")

	log.Printf("微信验证请求参数: signature=%s, timestamp=%s, nonce=%s, echostr=%s", signature, timestamp, nonce, echostr)

	// 按照微信的验证逻辑进行处理
	// 1. 将 token、timestamp、nonce 三个参数进行字典序排序
	arrays := []string{WECHAT_TOKEN, timestamp, nonce}
	sort.Strings(arrays)

	// 2. 将三个参数字符串拼接成一个字符串进行 sha1 加密
	str := strings.Join(arrays, "")
	hasher := sha1.New()
	hasher.Write([]byte(str))
	sha1Str := hex.EncodeToString(hasher.Sum(nil))

	// 3. 将加密后的字符串与 signature 进行对比
	if sha1Str == signature {
		// 如果匹配成功，返回 echostr 给微信服务器
		c.String(200, echostr)
	} else {
		// 验证失败
		log.Printf("微信验证失败: 计算得到的签名=%s", sha1Str)
		c.String(403, "验证失败")
	}
}
