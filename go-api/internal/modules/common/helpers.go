package common

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"go-api/internal/cache"
	"go-api/internal/database"
	mailmodule "go-api/internal/modules/mail"
)

type RechargeBonusRule struct {
	Min      float64 `json:"min"`
	Max      float64 `json:"max"`
	BonusPct float64 `json:"bonus_pct"`
}

type RechargeBonusActivity struct {
	Enabled  bool                `json:"enabled"`
	Weekdays []int               `json:"weekdays"`
	Rules    []RechargeBonusRule `json:"rules"`
	Hint     string              `json:"hint"`
}

type RechargeBonusConfig struct {
	Enabled  bool                  `json:"enabled"`
	Rules    []RechargeBonusRule   `json:"rules"`
	Activity RechargeBonusActivity `json:"activity"`
}

func GetAdminConfigMap() (map[string]string, error) {
	rows, err := database.DB.Query("SELECT `v`, `k` FROM qingka_wangke_config")
	if err != nil {
		return map[string]string{}, nil
	}
	defer rows.Close()

	configMap := make(map[string]string)
	for rows.Next() {
		var key, value string
		rows.Scan(&key, &value)
		configMap[key] = value
	}
	return configMap, nil
}

func SendEmail(to, subject, htmlBody string) error {
	return mailmodule.Mail().SendEmail(to, subject, htmlBody)
}

func SendEmailWithType(to, subject, htmlBody, mailType string) error {
	return mailmodule.Mail().SendEmailWithType(to, subject, htmlBody, mailType)
}

func GenerateVerificationCode(purpose, identifier string, ttl time.Duration) (string, error) {
	ctx := context.Background()
	key := fmt.Sprintf("verify:%s:%s", purpose, identifier)
	code := randomDigits(6)

	if err := cache.RDB.Set(ctx, key, code, ttl).Err(); err != nil {
		return "", fmt.Errorf("存储验证码失败: %v", err)
	}

	rlKey := fmt.Sprintf("verify_rl:%s:%s", purpose, identifier)
	cache.RDB.Set(ctx, rlKey, "1", 60*time.Second)
	return code, nil
}

func VerifyVerificationCode(purpose, identifier, code string) bool {
	ctx := context.Background()
	key := fmt.Sprintf("verify:%s:%s", purpose, identifier)
	stored, err := cache.RDB.Get(ctx, key).Result()
	if err != nil || stored != code {
		return false
	}
	cache.RDB.Del(ctx, key)
	return true
}

func CheckVerificationRateLimit(purpose, identifier string) error {
	ctx := context.Background()
	rlKey := fmt.Sprintf("verify_rl:%s:%s", purpose, identifier)
	exists, _ := cache.RDB.Exists(ctx, rlKey).Result()
	if exists > 0 {
		ttl, _ := cache.RDB.TTL(ctx, rlKey).Result()
		return fmt.Errorf("发送过于频繁，请 %d 秒后重试", int(ttl.Seconds()))
	}
	return nil
}

func CrossRechargeAllowed(uid int) bool {
	if uid == 1 {
		return true
	}

	configMap, _ := GetAdminConfigMap()
	uidList := configMap["cross_recharge_uids"]
	if uidList == "" {
		return false
	}

	uidStr := fmt.Sprintf("%d", uid)
	for _, item := range splitCSV(uidList) {
		if item == uidStr {
			return true
		}
	}
	return false
}

func LoadRechargeBonusConfig() *RechargeBonusConfig {
	var raw string
	database.DB.QueryRow("SELECT `k` FROM qingka_wangke_config WHERE `v` = 'recharge_bonus_rules'").Scan(&raw)
	if raw == "" {
		return &RechargeBonusConfig{}
	}

	var cfg RechargeBonusConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return &RechargeBonusConfig{}
	}
	return &cfg
}

func randomDigits(n int) string {
	digits := make([]byte, n)
	for i := range digits {
		num, _ := rand.Int(rand.Reader, big.NewInt(10))
		digits[i] = byte('0' + num.Int64())
	}
	return string(digits)
}

func splitCSV(raw string) []string {
	var result []string
	for _, item := range strings.Split(raw, ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}
