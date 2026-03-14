package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"go-api/internal/cache"
)

func (s *PaperService) getToken() (string, error) {
	ctx := context.Background()

	cached, err := cache.RDB.Get(ctx, zwTokenCacheKey).Result()
	if err == nil && cached != "" {
		return cached, nil
	}

	conf, err := s.GetConfig()
	if err != nil {
		return "", fmt.Errorf("获取论文配置失败: %v", err)
	}
	username := conf["lunwen_api_username"]
	password := conf["lunwen_api_password"]
	if username == "" || password == "" {
		return "", fmt.Errorf("论文API未配置账号密码")
	}

	loginData, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	resp, err := s.client.Post(zwBaseURL+"prod-api/login", "application/json", bytes.NewReader(loginData))
	if err != nil {
		return "", fmt.Errorf("论文API登录请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("论文API登录响应解析失败: %v", err)
	}

	code, _ := result["code"].(float64)
	if int(code) != 200 {
		msg, _ := result["msg"].(string)
		return "", fmt.Errorf("论文API登录失败: %s", msg)
	}

	token, ok := result["token"].(string)
	if !ok || token == "" {
		return "", fmt.Errorf("论文API登录返回空token")
	}

	cache.RDB.Set(ctx, zwTokenCacheKey, token, zwTokenTTL)
	return token, nil
}

func (s *PaperService) clearToken() {
	ctx := context.Background()
	cache.RDB.Del(ctx, zwTokenCacheKey)
}

func (s *PaperService) apiGet(path string, params map[string]string) (map[string]interface{}, error) {
	return s.apiRequest("GET", path, nil, params)
}

func (s *PaperService) apiPost(path string, data interface{}) (map[string]interface{}, error) {
	return s.apiRequest("POST", path, data, nil)
}

func (s *PaperService) apiRequest(method, path string, data interface{}, queryParams map[string]string) (map[string]interface{}, error) {
	var lastErr error
	for attempt := 0; attempt < zwMaxRetry; attempt++ {
		result, err := s.doAPIRequest(method, path, data, queryParams)
		if err != nil {
			lastErr = err
			time.Sleep(500 * time.Millisecond)
			continue
		}

		code, _ := result["code"].(float64)
		if int(code) == 401 {
			s.clearToken()
			time.Sleep(500 * time.Millisecond)
			continue
		}

		return result, nil
	}
	return nil, fmt.Errorf("论文API多次请求失败: %v", lastErr)
}

func (s *PaperService) doAPIRequest(method, path string, data interface{}, queryParams map[string]string) (map[string]interface{}, error) {
	token, err := s.getToken()
	if err != nil {
		return nil, err
	}

	url := zwBaseURL + path
	if len(queryParams) > 0 {
		sep := "?"
		if strings.Contains(url, "?") {
			sep = "&"
		}
		for k, v := range queryParams {
			url += sep + k + "=" + v
			sep = "&"
		}
	}

	var req *http.Request
	if method == "GET" {
		req, err = http.NewRequest("GET", url, nil)
	} else {
		var body io.Reader
		if data != nil {
			jsonData, _ := json.Marshal(data)
			body = bytes.NewReader(jsonData)
		}
		req, err = http.NewRequest(method, url, body)
		if req != nil {
			req.Header.Set("Content-Type", "application/json")
		}
	}
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("JSON解析失败: %v, body: %s", err, string(respBody[:min(200, len(respBody))]))
	}

	return result, nil
}

func (s *PaperService) apiStreamRequest(path string, data interface{}, w http.ResponseWriter) error {
	token, err := s.getToken()
	if err != nil {
		return err
	}

	url := zwBaseURL + path
	jsonData, _ := json.Marshal(data)
	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		body, _ := io.ReadAll(resp.Body)
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
		return nil
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming unsupported")
	}

	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			w.Write(buf[:n])
			flusher.Flush()
		}
		if err != nil {
			break
		}
	}
	return nil
}

func (s *PaperService) apiFileUpload(path string, file multipart.File, fileHeader *multipart.FileHeader, extraFields map[string]string) (map[string]interface{}, error) {
	token, err := s.getToken()
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", fileHeader.Filename)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, err
	}

	for k, v := range extraFields {
		writer.WriteField(k, v)
	}
	writer.Close()

	url := zwBaseURL + path
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("JSON解析失败: %v", err)
	}

	return result, nil
}
