package app

import (
	"log"
	"nasa-go-admin/db"
	"nasa-go-admin/model/admin_model"
	"nasa-go-admin/pkg/response"
	"nasa-go-admin/utils"
	"net/http"
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

	// 从数据库获取OSS配置
	var settings []admin_model.SettingList
	if err := db.Dao.Where("type = ?", "oss_code").Find(&settings).Error; err != nil {
		log.Printf("获取OSS配置失败: %v", err)
		Resp.Err(c, response.INVALID_PARAMS, "获取OSS配置失败")
		return
	}

	// 将设置转换为map便于查找
	settingMap := make(map[string]string)
	for _, setting := range settings {
		settingMap[setting.Name] = setting.Secret
		settingMap[setting.Appid] = setting.Secret
	}

	// 初始化OSS配置
	ossConfig := utils.OSSConfig{
		Endpoint:        settings[0].Endpoint,
		AccessKeyID:     settings[0].Appid,
		AccessKeySecret: settings[0].Secret,
		BucketName:      settings[0].BucketName,
		BaseURL:         settings[0].BaseUrl,
	}

	// 验证必要的配置是否存在
	if ossConfig.Endpoint == "" || ossConfig.AccessKeyID == "" ||
		ossConfig.AccessKeySecret == "" || ossConfig.BucketName == "" {
		log.Printf("OSS配置不完整: %+v", ossConfig)
		Resp.Err(c, response.INVALID_PARAMS, "OSS配置不完整")
		return
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
