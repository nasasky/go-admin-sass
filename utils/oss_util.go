package utils

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

// 最大允许的文件大小 (10MB)
const MaxFileSize = 10 * 1024 * 1024

// 允许的文件类型
var AllowedFileTypes = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
	".pdf":  true,
	".doc":  true,
	".docx": true,
	".xls":  true,
	".xlsx": true,
}

// OSSConfig 存储OSS配置信息
type OSSConfig struct {
	Endpoint        string
	AccessKeyID     string
	AccessKeySecret string
	BucketName      string
	BaseURL         string // 访问URL前缀
	Timeout         int    // 超时时间(秒)
	Region          string // 区域
}

// OSSUtil OSS工具类
type OSSUtil struct {
	config OSSConfig
	client *tos.ClientV2
}

// NewOSSUtil 创建OSS工具实例
func NewOSSUtil(config OSSConfig) (*OSSUtil, error) {
	if config.Endpoint == "" || config.AccessKeyID == "" || config.AccessKeySecret == "" || config.BucketName == "" {
		return nil, errors.New("TOS配置参数不完整")
	}

	// 默认超时时间为30秒
	if config.Timeout <= 0 {
		config.Timeout = 30
	}

	// 创建TOS客户端凭证
	credential := tos.NewStaticCredentials(config.AccessKeyID, config.AccessKeySecret)

	// 按官方文档创建TOS客户端
	tosClient, err := tos.NewClientV2(config.Endpoint,
		tos.WithCredentials(credential),
		tos.WithRegion(config.Region))

	if err != nil {
		return nil, fmt.Errorf("初始化TOS客户端失败: %w", err)
	}

	return &OSSUtil{
		config: config,
		client: tosClient,
	}, nil
}

// Close 关闭客户端并释放资源
func (u *OSSUtil) Close() {
	if u.client != nil {
		u.client.Close()
	}
}

// UploadFile 上传文件到OSS
func (u *OSSUtil) UploadFile(file *multipart.FileHeader) (string, error) {
	return u.uploadFileToPath(file, "uploads")
}

// UploadFileWithDir 上传文件到指定目录
func (u *OSSUtil) UploadFileWithDir(file *multipart.FileHeader, directory string) (string, error) {
	if directory == "" {
		return "", errors.New("目录不能为空")
	}
	return u.uploadFileToPath(file, directory)
}

// uploadFileToPath 内部方法：上传文件到指定路径
func (u *OSSUtil) uploadFileToPath(file *multipart.FileHeader, directory string) (string, error) {
	// 检查文件大小
	if file.Size > MaxFileSize {
		return "", fmt.Errorf("文件过大，最大允许 %d MB", MaxFileSize/(1024*1024))
	}

	// 检查文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !AllowedFileTypes[ext] {
		return "", errors.New("不支持的文件类型")
	}

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer func(src multipart.File) {
		err := src.Close()
		if err != nil {

		}
	}(src)

	// 确保目录格式正确
	directory = strings.Trim(directory, "/")
	if directory != "" {
		directory += "/"
	}

	// 生成唯一文件名
	timestamp := time.Now().Format("20060102150405")
	nanoSuffix := fmt.Sprintf("%03d", time.Now().UnixNano()%1000)
	filename := timestamp + nanoSuffix + ext
	objectName := directory + filename

	// 创建上下文并设置超时
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(u.config.Timeout)*time.Second)
	defer cancel()

	// 获取MIME类型
	// contentType := getMimeType(ext)

	// 上传文件到TOS
	input := &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket: u.config.BucketName,
			Key:    objectName,
		},
		Content: src, // 使用 Content 替代 Body

	}

	_, err = u.client.PutObjectV2(ctx, input)
	if err != nil {
		return "", fmt.Errorf("上传文件到TOS失败: %w", err)
	}

	// 返回可访问的URL
	return u.config.BaseURL + objectName, nil
}

// DeleteFile 删除OSS上的文件
func (u *OSSUtil) DeleteFile(objectPath string) error {
	// 从URL提取对象路径
	if strings.HasPrefix(objectPath, u.config.BaseURL) {
		objectPath = strings.TrimPrefix(objectPath, u.config.BaseURL+"/")
	}

	if objectPath == "" {
		return errors.New("无效的对象路径")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(u.config.Timeout)*time.Second)
	defer cancel()

	input := &tos.DeleteObjectV2Input{
		Bucket: u.config.BucketName,
		Key:    objectPath,
	}

	_, err := u.client.DeleteObjectV2(ctx, input)
	if err != nil {
		return fmt.Errorf("删除TOS文件失败: %w", err)
	}

	return nil
}

// ListFiles 列出指定目录下的文件
func (u *OSSUtil) ListFiles(directory string, maxKeys int) ([]string, error) {
	if directory != "" && !strings.HasSuffix(directory, "/") {
		directory += "/"
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(u.config.Timeout)*time.Second)
	defer cancel()

	input := &tos.ListObjectsV2Input{
		Bucket: u.config.BucketName,
	}

	output, err := u.client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("列出TOS文件失败: %w", err)
	}

	files := make([]string, 0, len(output.Contents))
	for _, obj := range output.Contents {
		files = append(files, u.config.BaseURL+"/"+obj.Key)
	}

	return files, nil
}

// getMimeType 根据扩展名获取MIME类型
func getMimeType(ext string) string {
	mimeTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	}

	if mime, ok := mimeTypes[ext]; ok {
		return mime
	}
	return "application/octet-stream"
}
