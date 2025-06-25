package app_service

import (
	"nasa-go-admin/db"
	"nasa-go-admin/model/app_model"
	"strings"
	"sync"
)

var (
	sensitiveWordsCache []string
	cacheMutex          sync.RWMutex
	cacheInitialized    bool
)

// initSensitiveWordsCache 初始化敏感词缓存
func initSensitiveWordsCache() error {
	if cacheInitialized {
		return nil
	}

	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if cacheInitialized {
		return nil
	}

	var words []app_model.SensitiveWord
	if err := db.Dao.Where("is_enabled = ?", true).Find(&words).Error; err != nil {
		return err
	}

	sensitiveWordsCache = make([]string, len(words))
	for i, word := range words {
		sensitiveWordsCache[i] = word.Word
	}

	cacheInitialized = true
	return nil
}

// RefreshSensitiveWordsCache 刷新敏感词缓存
func RefreshSensitiveWordsCache() error {
	cacheMutex.Lock()
	cacheInitialized = false
	cacheMutex.Unlock()
	return initSensitiveWordsCache()
}

// ContentFilter 内容过滤器
type ContentFilter struct {
	Title   string
	Content string
}

// FilterResult 过滤结果
type FilterResult struct {
	HasSensitiveWords bool
	MatchedWords      []string
	RejectReason      string
}

// Filter 过滤内容
func (cf *ContentFilter) Filter() (*FilterResult, error) {
	if err := initSensitiveWordsCache(); err != nil {
		return nil, err
	}

	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	result := &FilterResult{
		HasSensitiveWords: false,
		MatchedWords:      make([]string, 0),
	}

	// 合并标题和内容进行检查
	content := strings.ToLower(cf.Title + " " + cf.Content)

	for _, word := range sensitiveWordsCache {
		if strings.Contains(content, strings.ToLower(word)) {
			result.HasSensitiveWords = true
			result.MatchedWords = append(result.MatchedWords, word)
		}
	}

	if result.HasSensitiveWords {
		result.RejectReason = "内容包含敏感词：" + strings.Join(result.MatchedWords, "、")
	}

	return result, nil
}
