package service

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"go-api/internal/cache"
	"go-api/internal/database"
)

func (s *TuboshuService) getToken() (string, error) {
	ctx := context.Background()
	tokenKey := tbsPlatformName + "_token"
	cached, err := cache.RDB.Get(ctx, tokenKey).Result()
	if err == nil && cached != "" {
		return cached, nil
	}

	var token string
	err = database.DB.QueryRow("SELECT token FROM qingka_wangke_huoyuan WHERE pt = ? LIMIT 1", tbsPlatformName).Scan(&token)
	if err != nil || token == "" {
		return "", fmt.Errorf("土拨鼠未配置token，请在货源中心配置")
	}

	tokenWithBearer := "Bearer " + token
	cache.RDB.Set(ctx, tokenKey, tokenWithBearer, tbsTokenCacheTTL)
	return tokenWithBearer, nil
}

func (s *TuboshuService) clearToken() {
	ctx := context.Background()
	cache.RDB.Del(ctx, tbsPlatformName+"_token")
}

func (s *TuboshuService) getBestAPIURL() string {
	s.mu.RLock()
	if s.bestURL != "" {
		url := s.bestURL
		s.mu.RUnlock()
		return url
	}
	s.mu.RUnlock()

	ctx := context.Background()
	cached, err := cache.RDB.Get(ctx, "tuboshu:best_route").Result()
	if err == nil && cached != "" {
		s.mu.Lock()
		s.bestURL = cached
		s.mu.Unlock()
		return cached
	}

	best := s.testRoutes()
	if best != "" {
		s.mu.Lock()
		s.bestURL = best
		s.mu.Unlock()
		cache.RDB.Set(ctx, "tuboshu:best_route", best, tbsRouteTestTTL)
		return best
	}

	return tbsAPIURLs[0]
}

func (s *TuboshuService) testRoutes() string {
	type result struct {
		url     string
		latency time.Duration
	}
	ch := make(chan result, len(tbsAPIURLs))

	for _, u := range tbsAPIURLs {
		go func(testURL string) {
			client := &http.Client{Timeout: tbsRouteTestTimeout}
			start := time.Now()
			resp, err := client.Head(testURL)
			if err != nil {
				return
			}
			resp.Body.Close()
			ch <- result{url: testURL, latency: time.Since(start)}
		}(u)
	}

	var best result
	timer := time.After(tbsRouteTestTimeout + time.Second)
	for i := 0; i < len(tbsAPIURLs); i++ {
		select {
		case r := <-ch:
			if best.url == "" || r.latency < best.latency {
				best = r
			}
		case <-timer:
			break
		}
	}

	if best.url != "" {
		log.Printf("[Tuboshu] 最优线路: %s (延迟: %v)", best.url, best.latency)
	}
	return best.url
}

func (s *TuboshuService) upstreamRequest(endpoint string, data interface{}, method string, isBlob bool) ([]byte, error) {
	apiURL := s.getBestAPIURL() + strings.TrimLeft(endpoint, "/")

	var lastErr error
	for attempt := 0; attempt < tbsMaxRetry; attempt++ {
		result, err := s.doUpstreamRequest(apiURL, data, method, isBlob)
		if err != nil {
			lastErr = err
			log.Printf("[Tuboshu] 请求失败 (尝试 %d): %v", attempt+1, err)
			time.Sleep(tbsRetryDelay)
			continue
		}

		if !isBlob {
			var resp map[string]interface{}
			shouldRetry := false
			if json.Unmarshal(result, &resp) == nil {
				if code, ok := resp["code"]; ok && fmt.Sprintf("%v", code) == "401" {
					log.Printf("[Tuboshu] 收到401，清除token重试")
					s.clearToken()
					shouldRetry = true
				}
				if !shouldRetry {
					if msg, ok := resp["message"].(string); ok {
						notLoggedIn := []string{"当前未登录", "API令牌无效或已过期", "未授权访问", "token无效或已过期", "用户未登录或登录已过期"}
						for _, m := range notLoggedIn {
							if msg == m {
								log.Printf("[Tuboshu] 未登录响应: %s", msg)
								s.clearToken()
								shouldRetry = true
								break
							}
						}
					}
				}
			}
			if shouldRetry {
				time.Sleep(tbsRetryDelay)
				continue
			}
		}

		return result, nil
	}

	return nil, fmt.Errorf("多次请求失败: %v", lastErr)
}

func (s *TuboshuService) doUpstreamRequest(apiURL string, data interface{}, method string, isBlob bool) ([]byte, error) {
	token, err := s.getToken()
	if err != nil {
		return nil, err
	}

	var req *http.Request

	switch strings.ToUpper(method) {
	case "GET":
		if data != nil {
			params, ok := data.(map[string]interface{})
			if ok && len(params) > 0 {
				sep := "?"
				if strings.Contains(apiURL, "?") {
					sep = "&"
				}
				for k, v := range params {
					apiURL += sep + k + "=" + fmt.Sprintf("%v", v)
					sep = "&"
				}
			}
		}
		req, err = http.NewRequest("GET", apiURL, nil)
	case "POST", "PUT", "DELETE":
		var body io.Reader
		if data != nil {
			jsonData, _ := json.Marshal(data)
			body = bytes.NewReader(jsonData)
		}
		req, err = http.NewRequest(strings.ToUpper(method), apiURL, body)
		if req != nil {
			req.Header.Set("Content-Type", "application/json")
		}
	default:
		req, err = http.NewRequest(method, apiURL, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Authorization", token)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode == 401 {
		return []byte(`{"code":401,"message":"用户未登录或登录已过期"}`), nil
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP错误: %d", resp.StatusCode)
	}

	return body, nil
}

func (s *TuboshuService) upstreamRequestJSON(endpoint string, data interface{}, method string) (map[string]interface{}, error) {
	body, err := s.upstreamRequest(endpoint, data, method, false)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("JSON解析失败: %v", err)
	}
	return result, nil
}

func (s *TuboshuService) cacheKey(route string, params map[string]interface{}) string {
	paramStr := ""
	if len(params) > 0 {
		data, _ := json.Marshal(params)
		paramStr = fmt.Sprintf(":%x", md5.Sum(data))
	}
	return fmt.Sprintf("tuboshu:%s%s", route, paramStr)
}

func (s *TuboshuService) handleCacheableRoute(method, path string, params map[string]interface{}, ttl time.Duration) (map[string]interface{}, error) {
	ctx := context.Background()
	key := s.cacheKey(method+"-"+path, params)

	cached, err := cache.RDB.Get(ctx, key).Result()
	if err == nil && cached != "" {
		var result map[string]interface{}
		if json.Unmarshal([]byte(cached), &result) == nil {
			return result, nil
		}
	}

	result, err := s.upstreamRequestJSON(path, params, method)
	if err != nil {
		return nil, err
	}

	if success, ok := result["success"].(bool); ok && success {
		data, _ := json.Marshal(result)
		cache.RDB.Set(ctx, key, string(data), ttl)
	}

	return result, nil
}
