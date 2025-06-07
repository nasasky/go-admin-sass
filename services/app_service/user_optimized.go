package app_service

import (
	"context"
	"fmt"
	"nasa-go-admin/db"
	"nasa-go-admin/model/app_model"
	"nasa-go-admin/pkg/cache"
	"nasa-go-admin/redis"
	"time"
)

// OptimizedUserService 优化的用户服务
type OptimizedUserService struct {
	cache *cache.CacheManager
}

// NewOptimizedUserService 创建优化的用户服务
func NewOptimizedUserService() *OptimizedUserService {
	return &OptimizedUserService{
		cache: cache.NewCacheManager(redis.GetClient()),
	}
}

// GetUserInfoOptimized 获取用户信息（优化版本）
func (s *OptimizedUserService) GetUserInfoOptimized(ctx context.Context, uid int) (*app_model.UserApp, error) {
	// 1. 尝试从缓存获取
	cacheKey := fmt.Sprintf("user:info:%d", uid)
	var user app_model.UserApp
	if err := s.cache.Get(ctx, cacheKey, &user); err == nil {
		return &user, nil
	}

	// 2. 从数据库查询
	err := db.Dao.WithContext(ctx).
		Select("id, username, phone, avatar, nickname, sex, openid, unionid, create_time, update_time").
		Where("id = ?", uid).
		First(&user).Error

	if err != nil {
		return nil, fmt.Errorf("查询用户信息失败: %w", err)
	}

	// 3. 缓存结果（30分钟）
	s.cache.Set(ctx, cacheKey, user, 30*time.Minute)

	return &user, nil
}

// BatchGetUsersByIDs 批量获取用户信息（优化版本）
func (s *OptimizedUserService) BatchGetUsersByIDs(ctx context.Context, userIDs []int) (map[int]app_model.UserApp, error) {
	if len(userIDs) == 0 {
		return make(map[int]app_model.UserApp), nil
	}

	// 去重
	uniqueIDs := make(map[int]struct{})
	for _, id := range userIDs {
		uniqueIDs[id] = struct{}{}
	}

	result := make(map[int]app_model.UserApp)
	var missedIDs []int

	// 1. 尝试从缓存批量获取
	for id := range uniqueIDs {
		cacheKey := fmt.Sprintf("user:info:%d", id)
		var user app_model.UserApp
		if err := s.cache.Get(ctx, cacheKey, &user); err == nil {
			result[id] = user
		} else {
			missedIDs = append(missedIDs, id)
		}
	}

	// 2. 对于缓存未命中的用户，批量查询数据库
	if len(missedIDs) > 0 {
		var userList []app_model.UserApp
		err := db.Dao.WithContext(ctx).
			Select("id, username, phone, avatar, nickname, sex, openid, unionid, create_time, update_time").
			Where("id IN ?", missedIDs).
			Find(&userList).Error

		if err != nil {
			return nil, fmt.Errorf("批量查询用户失败: %w", err)
		}

		// 添加到结果中并异步缓存
		for _, user := range userList {
			result[user.ID] = user

			// 异步缓存单个用户
			go func(u app_model.UserApp) {
				cacheKey := fmt.Sprintf("user:info:%d", u.ID)
				s.cache.Set(context.Background(), cacheKey, u, 30*time.Minute)
			}(user)
		}
	}

	return result, nil
}

// InvalidateUserCache 失效用户缓存
func (s *OptimizedUserService) InvalidateUserCache(ctx context.Context, userID int) error {
	cacheKey := fmt.Sprintf("user:info:%d", userID)
	return s.cache.Delete(ctx, cacheKey)
}

// UpdateUserInfoWithCache 更新用户信息并失效缓存
func (s *OptimizedUserService) UpdateUserInfoWithCache(ctx context.Context, userID int, updates map[string]interface{}) error {
	// 1. 更新数据库
	err := db.Dao.WithContext(ctx).
		Model(&app_model.UserApp{}).
		Where("id = ?", userID).
		Updates(updates).Error

	if err != nil {
		return fmt.Errorf("更新用户信息失败: %w", err)
	}

	// 2. 失效缓存
	if err := s.InvalidateUserCache(ctx, userID); err != nil {
		// 记录日志但不返回错误，因为数据库更新已成功
		fmt.Printf("Warning: 删除用户缓存失败: %v\n", err)
	}

	return nil
}

// GetUserProfile 获取用户详细资料（包含profile信息）
func (s *OptimizedUserService) GetUserProfile(ctx context.Context, uid int) (*UserProfileData, error) {
	// 1. 尝试从缓存获取
	cacheKey := fmt.Sprintf("user:profile:%d", uid)
	var profile UserProfileData
	if err := s.cache.Get(ctx, cacheKey, &profile); err == nil {
		return &profile, nil
	}

	// 2. 从数据库查询
	var user app_model.UserApp
	var userProfile app_model.AppProfile

	// 并发查询用户基本信息和详细资料
	userChan := make(chan error, 1)
	profileChan := make(chan error, 1)

	go func() {
		err := db.Dao.WithContext(ctx).
			Select("id, username, phone, avatar, nickname, sex, create_time, update_time").
			Where("id = ?", uid).
			First(&user).Error
		userChan <- err
	}()

	go func() {
		err := db.Dao.WithContext(ctx).
			Select("id, nick_name, address, email, openid, unionid").
			Where("id = ?", uid).
			First(&userProfile).Error
		profileChan <- err
	}()

	// 等待查询完成
	userErr := <-userChan
	profileErr := <-profileChan

	if userErr != nil {
		return nil, fmt.Errorf("查询用户基本信息失败: %w", userErr)
	}

	// Profile不存在不算错误
	if profileErr != nil && profileErr.Error() != "record not found" {
		return nil, fmt.Errorf("查询用户详细资料失败: %w", profileErr)
	}

	// 3. 组装数据
	profile = UserProfileData{
		ID:        user.ID,
		Username:  user.Username,
		Phone:     user.Phone,
		Avatar:    user.Avatar,
		Nickname:  user.Nickname,
		Sex:       user.Sex,
		NickName:  userProfile.NickName,
		Address:   userProfile.Address,
		Email:     userProfile.Email,
		Openid:    userProfile.Openid,
		UnionID:   userProfile.UnionID,
		CreatedAt: user.CreateTime,
		UpdatedAt: user.UpdateTime,
	}

	// 4. 缓存结果（20分钟）
	s.cache.Set(ctx, cacheKey, profile, 20*time.Minute)

	return &profile, nil
}

// UserProfileData 用户完整资料数据
type UserProfileData struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Phone     string    `json:"phone"`
	Avatar    string    `json:"avatar"`
	Nickname  string    `json:"nickname"`
	Sex       int       `json:"sex"`
	NickName  string    `json:"nick_name"`
	Address   string    `json:"address"`
	Email     string    `json:"email"`
	Openid    string    `json:"openid"`
	UnionID   string    `json:"unionid"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CheckUserExists 检查用户是否存在（带缓存）
func (s *OptimizedUserService) CheckUserExists(ctx context.Context, phone string) (bool, error) {
	// 1. 尝试从缓存获取
	cacheKey := fmt.Sprintf("user:exists:%s", phone)
	var exists bool
	if err := s.cache.Get(ctx, cacheKey, &exists); err == nil {
		return exists, nil
	}

	// 2. 从数据库查询
	var count int64
	err := db.Dao.WithContext(ctx).
		Model(&app_model.UserApp{}).
		Where("phone = ?", phone).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("检查用户是否存在失败: %w", err)
	}

	exists = count > 0

	// 3. 缓存结果（5分钟，相对较短因为用户状态可能变化）
	s.cache.Set(ctx, cacheKey, exists, 5*time.Minute)

	return exists, nil
}
