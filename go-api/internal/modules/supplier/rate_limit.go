package supplier

import (
	"net/url"
	"strings"
	"sync"
	"time"
)

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
