package service

import (
	"go-api/internal/model"
	"net/url"
	"strings"
	"sync"
	"time"
)

// hostRateLimiter 每个上游主机的请求速率限制器，避免并发请求打爆同一个上游。
type hostRateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rateBucket
}

type rateBucket struct {
	lastTime time.Time
	mu       sync.Mutex
}

var globalRateLimiter = &hostRateLimiter{
	limiters: make(map[string]*rateBucket),
}

// wait 对指定 host 限速：同一 host 两次请求之间至少间隔 interval。
func (rl *hostRateLimiter) wait(host string, interval time.Duration) {
	rl.mu.Lock()
	bucket, ok := rl.limiters[host]
	if !ok {
		bucket = &rateBucket{}
		rl.limiters[host] = bucket
	}
	rl.mu.Unlock()

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(bucket.lastTime)
	if elapsed < interval {
		time.Sleep(interval - elapsed)
	}
	bucket.lastTime = time.Now()
}

// extractHost 从 URL 中提取主机名（用于限速 key）。
func extractHost(rawURL string) string {
	rawURL = strings.TrimRight(rawURL, "/")
	if !strings.HasPrefix(rawURL, "http") {
		rawURL = "http://" + rawURL
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return u.Host
}

func simpleBuildBaseURL(sup *model.SupplierFull) string {
	baseURL := strings.TrimRight(sup.URL, "/")
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "http://" + baseURL
	}
	return baseURL
}

func simpleGetToken(sup *model.SupplierFull) string {
	if sup.Token != "" {
		return sup.Token
	}
	return sup.Pass
}
