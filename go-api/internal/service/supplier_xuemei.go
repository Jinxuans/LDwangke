package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"go-api/internal/model"
)

// xuemeiBuildURL 构建学妹平台 API URL
func xuemeiBuildURL(sup *model.SupplierFull, act string) string {
	baseURL := strings.TrimRight(sup.URL, "/")
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "http://" + baseURL
	}
	return fmt.Sprintf("%s/api.php?act=%s", baseURL, act)
}

// xuemeiAuthParams 返回学妹平台通用认证参数
func xuemeiAuthParams(sup *model.SupplierFull) url.Values {
	params := url.Values{}
	params.Set("uid", sup.User)
	params.Set("key", sup.Pass)
	return params
}

// XueMeiShouHou 学妹售后反馈：act=shouhou, uid/key/oid/fankui
func XueMeiShouHou(sup *model.SupplierFull, oid, fankui string) (int, string, error) {
	apiURL := xuemeiBuildURL(sup, "shouhou")
	params := xuemeiAuthParams(sup)
	params.Set("oid", oid)
	params.Set("fankui", fankui)

	if host := extractHost(sup.URL); host != "" {
		globalRateLimiter.wait(host, 500*time.Millisecond)
	}

	resp, err := sharedHTTPClient.PostForm(apiURL, params)
	if err != nil {
		return -1, "", fmt.Errorf("请求学妹售后失败：%v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return -1, string(body), nil
	}
	return result.Code, result.Msg, nil
}

// XueMeiGetCity 学妹获取城市/IP节点列表：act=getcity, uid/key
func XueMeiGetCity(sup *model.SupplierFull) (interface{}, error) {
	apiURL := xuemeiBuildURL(sup, "getcity")
	params := xuemeiAuthParams(sup)

	if host := extractHost(sup.URL); host != "" {
		globalRateLimiter.wait(host, 500*time.Millisecond)
	}

	resp, err := sharedHTTPClient.PostForm(apiURL, params)
	if err != nil {
		return nil, fmt.Errorf("请求学妹城市列表失败：%v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("解析响应失败：%s", string(body))
	}
	return raw, nil
}

// XueMeiEditIP 学妹修改订单IP节点：act=editIp, uid/key/oid/id
func XueMeiEditIP(sup *model.SupplierFull, oid, nodeID string) (int, string, error) {
	apiURL := xuemeiBuildURL(sup, "editIp")
	params := xuemeiAuthParams(sup)
	params.Set("oid", oid)
	params.Set("id", nodeID)

	if host := extractHost(sup.URL); host != "" {
		globalRateLimiter.wait(host, 500*time.Millisecond)
	}

	resp, err := sharedHTTPClient.PostForm(apiURL, params)
	if err != nil {
		return -1, "", fmt.Errorf("请求学妹修改IP失败：%v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return -1, string(body), nil
	}
	return result.Code, result.Msg, nil
}

// XueMeiYouXian 学妹优先处理订单：act=youxian, uid/key/oid
func XueMeiYouXian(sup *model.SupplierFull, oid string) (int, string, error) {
	apiURL := xuemeiBuildURL(sup, "youxian")
	params := xuemeiAuthParams(sup)
	params.Set("oid", oid)

	if host := extractHost(sup.URL); host != "" {
		globalRateLimiter.wait(host, 500*time.Millisecond)
	}

	resp, err := sharedHTTPClient.PostForm(apiURL, params)
	if err != nil {
		return -1, "", fmt.Errorf("请求学妹优先处理失败：%v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return -1, string(body), nil
	}
	return result.Code, result.Msg, nil
}

// XueMeiGetNameResult 学妹获取项目名称返回结果
type XueMeiGetNameResult struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		OID          string `json:"oid"`
		RunZoneData  []struct {
			Value string `json:"value"`
			Label string `json:"label"`
		} `json:"runZoneData"`
	} `json:"data"`
}

// XueMeiGetName 学妹获取可切换项目列表：act=getname, uid/key/id
func XueMeiGetName(sup *model.SupplierFull, orderID string) (*XueMeiGetNameResult, error) {
	apiURL := xuemeiBuildURL(sup, "getname")
	params := xuemeiAuthParams(sup)
	params.Set("id", orderID)

	if host := extractHost(sup.URL); host != "" {
		globalRateLimiter.wait(host, 500*time.Millisecond)
	}

	resp, err := sharedHTTPClient.PostForm(apiURL, params)
	if err != nil {
		return nil, fmt.Errorf("请求学妹获取项目失败：%v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result XueMeiGetNameResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败：%s", string(body))
	}
	return &result, nil
}

// XueMeiEditName 学妹修改订单项目：act=editname, uid/key/oid/nameid
func XueMeiEditName(sup *model.SupplierFull, oid, nameID string) (int, string, error) {
	apiURL := xuemeiBuildURL(sup, "editname")
	params := xuemeiAuthParams(sup)
	params.Set("oid", oid)
	params.Set("nameid", nameID)

	if host := extractHost(sup.URL); host != "" {
		globalRateLimiter.wait(host, 500*time.Millisecond)
	}

	resp, err := sharedHTTPClient.PostForm(apiURL, params)
	if err != nil {
		return -1, "", fmt.Errorf("请求学妹修改项目失败：%v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return -1, string(body), nil
	}
	return result.Code, result.Msg, nil
}

// XueMeiChaZhsLog 学妹查询智慧树日志：act=cha_zhs_log, uid/key/id
func XueMeiChaZhsLog(sup *model.SupplierFull, orderID string) (interface{}, error) {
	apiURL := xuemeiBuildURL(sup, "cha_zhs_log")
	params := xuemeiAuthParams(sup)
	params.Set("id", orderID)

	if host := extractHost(sup.URL); host != "" {
		globalRateLimiter.wait(host, 500*time.Millisecond)
	}

	resp, err := sharedHTTPClient.PostForm(apiURL, params)
	if err != nil {
		return nil, fmt.Errorf("请求学妹智慧树日志失败：%v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("解析响应失败：%s", string(body))
	}
	return raw, nil
}
