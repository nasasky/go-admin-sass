package admin

import (
	"nasa-go-admin/utils"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// UploadFile 处理文件上传
func UploadFile(c *gin.Context) {
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "获取上传文件失败: " + err.Error(),
		})
		return
	}

	// 检查文件类型
	if !isAllowedFileType(file.Filename) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "不支持的文件类型，仅支持图片文件",
		})
		return
	}

	// 从环境变量初始化OSS配置
	ossConfig := utils.OSSConfig{
		Endpoint:        os.Getenv("OSS_ENDPOINT"),
		AccessKeyID:     os.Getenv("OSS_ACCESS_KEY_ID"),
		AccessKeySecret: os.Getenv("OSS_ACCESS_KEY_SECRET"),
		BucketName:      os.Getenv("OSS_BUCKET_NAME"),
		BaseURL:         os.Getenv("OSS_BASE_URL"),
	}

	ossUtil, err := utils.NewOSSUtil(ossConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "初始化TOS客户端失败: " + err.Error(),
		})
		return
	}

	// 上传文件
	fileURL, err := ossUtil.UploadFile(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "上传文件失败: " + err.Error(),
		})
		return
	}

	// 返回文件URL
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "上传成功",
		"data": map[string]string{
			"url": fileURL,
		},
	})
}

// isAllowedFileType 检查是否为允许的文件类型
func isAllowedFileType(filename string) bool {
	ext := filepath.Ext(filename)
	allowedTypes := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}
	return allowedTypes[strings.ToLower(ext)]
}
