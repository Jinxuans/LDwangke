package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"go-api/internal/cache"
)

type VerificationService struct{}

func NewVerificationService() *VerificationService {
	return &VerificationService{}
}

// GenerateCode 生成6位数字验证码并存入 Redis
// purpose: register / reset_pwd / change_email
// identifier: 邮箱地址或用户标识
func (s *VerificationService) GenerateCode(purpose, identifier string, ttl time.Duration) (string, error) {
	ctx := context.Background()
	key := fmt.Sprintf("verify:%s:%s", purpose, identifier)

	code := s.randomDigits(6)

	err := cache.RDB.Set(ctx, key, code, ttl).Err()
	if err != nil {
		return "", fmt.Errorf("存储验证码失败: %v", err)
	}

	// 设置频率限制键，60秒内不能重复发送
	rlKey := fmt.Sprintf("verify_rl:%s:%s", purpose, identifier)
	cache.RDB.Set(ctx, rlKey, "1", 60*time.Second)

	return code, nil
}

// VerifyCode 校验验证码，成功后删除（一次性）
func (s *VerificationService) VerifyCode(purpose, identifier, code string) bool {
	ctx := context.Background()
	key := fmt.Sprintf("verify:%s:%s", purpose, identifier)

	stored, err := cache.RDB.Get(ctx, key).Result()
	if err != nil || stored != code {
		return false
	}

	// 验证成功，删除验证码
	cache.RDB.Del(ctx, key)
	return true
}

// RateLimitCheck 检查是否在冷却期内（60秒）
func (s *VerificationService) RateLimitCheck(purpose, identifier string) error {
	ctx := context.Background()
	rlKey := fmt.Sprintf("verify_rl:%s:%s", purpose, identifier)

	exists, _ := cache.RDB.Exists(ctx, rlKey).Result()
	if exists > 0 {
		ttl, _ := cache.RDB.TTL(ctx, rlKey).Result()
		return fmt.Errorf("发送过于频繁，请 %d 秒后重试", int(ttl.Seconds()))
	}
	return nil
}

// randomDigits 生成指定位数的随机数字字符串
func (s *VerificationService) randomDigits(n int) string {
	digits := make([]byte, n)
	for i := range digits {
		num, _ := rand.Int(rand.Reader, big.NewInt(10))
		digits[i] = byte('0' + num.Int64())
	}
	return string(digits)
}
