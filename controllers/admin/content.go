package admin

import (
	"fmt"
	"nasa-go-admin/inout"
	"nasa-go-admin/model/admin_model"
	"nasa-go-admin/services/admin_service"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
)

// ProcessContent 处理内容接口
func ProcessContent(c *gin.Context) {
	var req inout.ProcessContentReq
	if err := c.ShouldBind(&req); err != nil {
		Resp.Err(c, 20001, "参数错误: "+err.Error())
		return
	}

	// 处理内容逻辑
	content := req.Content
	if content == "" {
		Resp.Err(c, 20001, "内容不能为空")
		return
	}

	// 构建响应数据
	resp := inout.ProcessContentResp{
		Message:     "内容处理成功",
		ProcessTime: time.Now().Format("2006-01-02 15:04:05"),
	}

	// 计算内容信息
	resp.ContentInfo.Length = len(content)
	resp.ContentInfo.CharCount = utf8.RuneCountInString(content)
	resp.ContentInfo.WordCount = len(strings.Fields(content))
	resp.ContentInfo.ProcessedAt = time.Now().Format(time.RFC3339)

	// 记录处理日志
	fmt.Printf("处理内容: 长度=%d, 字符数=%d, 单词数=%d, 时间=%s\n",
		resp.ContentInfo.Length,
		resp.ContentInfo.CharCount,
		resp.ContentInfo.WordCount,
		resp.ProcessTime)

	// 调用飞书推送接口
	feishuService := &admin_service.FeishuService{}
	feishuReq := admin_model.FeishuMessageRequest{
		ReceiveId: "oc_b6bfe8a5799f0296b32a8f6a09f18311",
		// ReceiveIdType: "open_id", // 使用open_id类型
		MsgType: "text",
		Content: content,
	}

	if _, err := feishuService.SendFeishuMessage(c, feishuReq); err != nil {
		// 即使飞书推送失败，我们仍然返回内容处理成功
		fmt.Printf("飞书推送失败: %v\n", err)
	}

	Resp.Succ(c, resp)
}
