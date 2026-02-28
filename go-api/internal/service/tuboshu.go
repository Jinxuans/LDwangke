package service

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-api/internal/cache"
	"go-api/internal/database"
)

// ==================== 配置 ====================

const (
	tbsPlatformName     = "tuboshu"
	tbsTokenCacheTTL    = 86400 * time.Second
	tbsDefaultTTL       = 3600 * time.Second
	tbsMaxRetry         = 3
	tbsRetryDelay       = 500 * time.Millisecond
	tbsRouteTestTTL     = 86400 * time.Second
	tbsRouteTestTimeout = 5 * time.Second
)

var tbsAPIURLs = []string{
	"https://sd.polars.cc/api/",
	"http://sdapi.polars.cc/",
	"https://api.aiwriting.icu/",
}

// ==================== 类型定义 ====================

// TuboshuConfig 土拨鼠配置（存在 qingka_wangke_config 中）
type TuboshuConfig struct {
	PriceRatio     float64                `json:"price_ratio"`     // 价格倍率，默认5
	PriceConfig    map[string]interface{} `json:"price_config"`    // 价格配置
	PageVisibility map[string]bool        `json:"page_visibility"` // 页面显示配置
}

// TuboshuDialogue 论文订单
type TuboshuDialogue struct {
	ID          int     `json:"id"`
	UID         int     `json:"uid"`
	Title       string  `json:"title"`
	State       string  `json:"state"`
	DownloadURL string  `json:"download_url"`
	AddTime     string  `json:"addtime"`
	IP          string  `json:"ip"`
	SourceID    int64   `json:"source_id"`
	DialogueID  string  `json:"dialogue_id"`
	Point       float64 `json:"point"`
	Type        string  `json:"type"`
}

// TuboshuRouteRequest 前端请求体
type TuboshuRouteRequest struct {
	Method string                 `json:"method"`
	Path   string                 `json:"path"`
	Params map[string]interface{} `json:"params"`
	IsBlob bool                   `json:"isBlob"`
}

// TuboshuService 土拨鼠论文服务
type TuboshuService struct {
	client  *http.Client
	mu      sync.RWMutex
	bestURL string
}

var tuboshuServiceInstance *TuboshuService
var tuboshuServiceOnce sync.Once

// NewTuboshuService 获取单例服务
func NewTuboshuService() *TuboshuService {
	tuboshuServiceOnce.Do(func() {
		tuboshuServiceInstance = &TuboshuService{
			client: &http.Client{Timeout: 120 * time.Second},
		}
	})
	return tuboshuServiceInstance
}

// ==================== 建表 ====================

func (s *TuboshuService) EnsureTable() {
	_, err := database.DB.Exec(`CREATE TABLE IF NOT EXISTS qingka_wangke_dialogue (
		id int(11) NOT NULL AUTO_INCREMENT,
		uid int(11) NOT NULL,
		title varchar(255) NOT NULL DEFAULT '',
		state varchar(255) NOT NULL DEFAULT 'PENDING',
		download_url varchar(255) NOT NULL DEFAULT '',
		addtime varchar(255) NOT NULL DEFAULT '',
		ip varchar(255) NOT NULL DEFAULT '',
		source_id bigint(17) NOT NULL DEFAULT 0,
		dialogue_id varchar(32) NOT NULL DEFAULT '0',
		point decimal(11,2) NOT NULL DEFAULT 0.00,
		type varchar(32) NOT NULL DEFAULT '',
		PRIMARY KEY (id),
		KEY idx_uid (uid),
		KEY idx_state (state)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`)
	if err != nil {
		log.Printf("[Tuboshu] 创建 dialogue 表失败: %v", err)
	}

	_, err = database.DB.Exec(`CREATE TABLE IF NOT EXISTS points_product (
		id INT(11) NOT NULL AUTO_INCREMENT,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		image_url VARCHAR(500),
		price DECIMAL(10,2) NOT NULL,
		status ENUM('ENABLED','DISABLED') NOT NULL DEFAULT 'ENABLED',
		sort_order INT(11) NOT NULL DEFAULT 0,
		create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		PRIMARY KEY (id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`)
	if err != nil {
		log.Printf("[Tuboshu] 创建 points_product 表失败: %v", err)
	}

	_, err = database.DB.Exec(`CREATE TABLE IF NOT EXISTS points_product_code (
		id INT(11) NOT NULL AUTO_INCREMENT,
		product_id INT(11) NOT NULL,
		code VARCHAR(500) NOT NULL,
		status ENUM('AVAILABLE','EXCHANGED') NOT NULL DEFAULT 'AVAILABLE',
		exchanged_by INT(11),
		exchanged_at DATETIME,
		create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (id),
		KEY idx_product_status (product_id, status)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`)
	if err != nil {
		log.Printf("[Tuboshu] 创建 points_product_code 表失败: %v", err)
	}

	_, err = database.DB.Exec(`CREATE TABLE IF NOT EXISTS points_exchange_record (
		id INT(11) NOT NULL AUTO_INCREMENT,
		uid INT(11) NOT NULL,
		product_id INT(11) NOT NULL,
		product_name VARCHAR(255) NOT NULL,
		code_id INT(11) NOT NULL,
		code VARCHAR(500) NOT NULL,
		points_cost DECIMAL(10,2) NOT NULL,
		create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (id),
		KEY idx_uid (uid)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`)
	if err != nil {
		log.Printf("[Tuboshu] 创建 points_exchange_record 表失败: %v", err)
	}

	// 初始化默认配置（如果不存在）
	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_config WHERE skey = 'tuboshu_config'").Scan(&count)
	if count == 0 {
		defaultCfg := TuboshuConfig{
			PriceRatio:  5,
			PriceConfig: defaultPriceConfig(),
			PageVisibility: map[string]bool{
				"ComponentStagePage": true,
				"ChatPage":           true,
				"ChartPage":          true,
				"TemplatePage":       true,
				"ReductionPage":      true,
				"AccountTable":       true,
				"TicketPage":         true,
			},
		}
		data, _ := json.Marshal(defaultCfg)
		database.DB.Exec("INSERT INTO qingka_wangke_config (skey, svalue) VALUES ('tuboshu_config', ?)", string(data))
	}
}

func defaultPriceConfig() map[string]interface{} {
	return map[string]interface{}{
		"SIMPLE_DIALOGUE": map[string]interface{}{
			"type": "id_based", "enabled": true,
			"prices": map[string]interface{}{
				"1": 44.8, "2": 12, "3": 12, "4": 12, "5": 18,
				"6": 0.6, "7": 0.6, "8": 0.6, "9": 24, "10": 12,
				"11": 12, "12": 12, "13": 3, "14": 12, "15": 12,
				"16": 12, "17": 12, "19": 0.6, "21": 0.6, "22": 0.6,
				"23": 0.6, "24": 0.6, "25": 0.6, "26": 0.6, "31": 0.6,
				"36": 4, "37": 4, "40": 0.2, "41": 0.2, "42": 0.2,
				"43": 4, "44": 4, "45": 0.1,
			},
		},
		"STAGE_DIALOGUE": map[string]interface{}{
			"type": "id_based", "enabled": true,
			"prices": map[string]interface{}{
				"1": 44.8, "2": 12, "3": 12, "4": 12, "5": 18,
				"9": 24, "10": 12, "11": 12, "12": 12, "13": 3,
				"14": 12, "15": 12, "16": 12, "17": 12,
			},
		},
		"PAPER_WRITING": map[string]interface{}{
			"type": "paper_writing",
			"config": map[string]interface{}{
				"enabled": true, "sectionBasePrice": 1.6,
				"pointBasePrice": 1.0, "v3ModelExtraPrice": 0.0,
				"reductionExtraPrice": 10.0, "enableV3Model": false,
			},
		},
		"PAPER_REDUCTION": map[string]interface{}{
			"type": "id_based", "enabled": true,
			"prices": map[string]interface{}{
				"NORMAL": 2, "ENGLISH": 4, "WEIPU": 3, "REPEAT": 5,
			},
		},
		"PAPER_OUTLINE": map[string]interface{}{
			"type": "fixed", "enabled": true, "price": 0.2,
		},
		"CHART_GENERATE": map[string]interface{}{
			"type": "fixed", "enabled": true, "price": 2.0,
		},
		"PPT_GENERATE": map[string]interface{}{
			"type": "fixed", "enabled": true, "price": 2.0,
		},
		"WORD_TEMPLATE": map[string]interface{}{
			"type": "fixed", "enabled": true, "price": 2.0,
		},
	}
}

// ==================== 配置管理 ====================

func (s *TuboshuService) GetConfig() (*TuboshuConfig, error) {
	var val string
	err := database.DB.QueryRow("SELECT svalue FROM qingka_wangke_config WHERE skey = 'tuboshu_config' LIMIT 1").Scan(&val)
	if err != nil {
		cfg := &TuboshuConfig{PriceRatio: 5, PriceConfig: defaultPriceConfig()}
		return cfg, nil
	}
	var cfg TuboshuConfig
	json.Unmarshal([]byte(val), &cfg)
	if cfg.PriceRatio == 0 {
		cfg.PriceRatio = 5
	}
	if cfg.PriceConfig == nil {
		cfg.PriceConfig = defaultPriceConfig()
	}
	return &cfg, nil
}

func (s *TuboshuService) SaveConfig(cfg *TuboshuConfig) error {
	data, _ := json.Marshal(cfg)
	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_config WHERE skey = 'tuboshu_config'").Scan(&count)
	if count > 0 {
		_, err := database.DB.Exec("UPDATE qingka_wangke_config SET svalue = ? WHERE skey = 'tuboshu_config'", string(data))
		return err
	}
	_, err := database.DB.Exec("INSERT INTO qingka_wangke_config (skey, svalue) VALUES ('tuboshu_config', ?)", string(data))
	return err
}

// SavePriceConfig 管理员保存价格配置
func (s *TuboshuService) SavePriceConfig(priceConfig map[string]interface{}) error {
	cfg, err := s.GetConfig()
	if err != nil {
		return err
	}

	// 合并配置
	for k, v := range priceConfig {
		cfg.PriceConfig[k] = v
	}

	// 确保 PAGE_VISIBILITY 存在
	if cfg.PageVisibility == nil {
		cfg.PageVisibility = map[string]bool{
			"ComponentStagePage": true, "ChatPage": true, "ChartPage": true,
			"TemplatePage": true, "ReductionPage": true, "AccountTable": true, "TicketPage": true,
		}
	}

	return s.SaveConfig(cfg)
}

// ==================== Token 管理 ====================

func (s *TuboshuService) getToken() (string, error) {
	ctx := context.Background()

	// 先从 Redis 缓存获取
	tokenKey := tbsPlatformName + "_token"
	cached, err := cache.RDB.Get(ctx, tokenKey).Result()
	if err == nil && cached != "" {
		return cached, nil
	}

	// 从数据库获取 token
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

// ==================== API 线路管理 ====================

func (s *TuboshuService) getBestAPIURL() string {
	s.mu.RLock()
	if s.bestURL != "" {
		url := s.bestURL
		s.mu.RUnlock()
		return url
	}
	s.mu.RUnlock()

	// 尝试从 Redis 获取缓存
	ctx := context.Background()
	cached, err := cache.RDB.Get(ctx, "tuboshu:best_route").Result()
	if err == nil && cached != "" {
		s.mu.Lock()
		s.bestURL = cached
		s.mu.Unlock()
		return cached
	}

	// 测试线路
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

// ==================== 上游 API 请求 ====================

// upstreamRequest 向土拨鼠上游 API 发请求
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

		// 检测 401
		if !isBlob {
			var resp map[string]interface{}
			shouldRetry := false
			if json.Unmarshal(result, &resp) == nil {
				if code, ok := resp["code"]; ok {
					if fmt.Sprintf("%v", code) == "401" {
						log.Printf("[Tuboshu] 收到401，清除token重试")
						s.clearToken()
						shouldRetry = true
					}
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
		// GET 请求：参数拼接到 URL
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

// upstreamRequestJSON 请求并解析 JSON 结果
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

// ==================== 缓存路由 ====================

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

	// 尝试从缓存读取
	cached, err := cache.RDB.Get(ctx, key).Result()
	if err == nil && cached != "" {
		var result map[string]interface{}
		if json.Unmarshal([]byte(cached), &result) == nil {
			return result, nil
		}
	}

	// 请求上游
	result, err := s.upstreamRequestJSON(path, params, method)
	if err != nil {
		return nil, err
	}

	// 缓存成功响应
	if success, ok := result["success"].(bool); ok && success {
		data, _ := json.Marshal(result)
		cache.RDB.Set(ctx, key, string(data), ttl)
	}

	return result, nil
}

// ==================== 路由定义 ====================

type routeConfig struct {
	pattern *regexp.Regexp
	handler string
	ttl     time.Duration
	replace string // "dialogueId" 表示需要替换路径中的 ID
	isBlob  bool
}

var tuboshuRoutes []routeConfig

func init() {
	type routeDef struct {
		pattern string
		handler string
		ttl     time.Duration
		replace string
		isBlob  bool
	}

	defs := []routeDef{
		// 缓存路由
		{`^GET-/dialogue/stage/list$`, "cacheable", 3600 * time.Second, "", false},
		{`^GET-/dialogue/template$`, "cacheable", 3600 * time.Second, "", false},
		{`^GET-/dialogue/list$`, "cacheable", 3600 * time.Second, "", false},
		{`^GET-/paint/templates$`, "cacheable", 3600 * time.Second, "", false},
		{`^GET-/dialogue/outlineInsertBtn$`, "cacheable", 3600 * time.Second, "", false},
		{`^GET-/dialogue/stage/\d+$`, "cacheable", 3600 * time.Second, "", false},
		{`^GET-/dialogue/paperCategory$`, "cacheable", 3600 * time.Second, "", false},
		{`^GET-/dialogue/csl-styles$`, "cacheable", 3600 * time.Second, "", false},
		{`^GET-/dialogue/paper-outline-types$`, "cacheable", 3600 * time.Second, "", false},
		{`^GET-/dialogue/\d+$`, "cacheable", 3600 * time.Second, "", false},
		{`^GET-/dialogue/wordTemplate$`, "cacheable", 3600 * time.Second, "", false},
		{`^GET-/userInfo$`, "cacheable", 300 * time.Second, "", false},
		{`^GET-/task/reduction/types$`, "cacheable", 3600 * time.Second, "", false},

		// 直接转发路由
		{`^POST-/dialogue/parse-proposal$`, "forward", 0, "", false},
		{`^GET-/dialogue/reference/search`, "forward", 0, "", false},
		{`^GET-/dialogue/tool`, "forward", 0, "", false},
		{`^GET-/task/reduction/document-paragraphs$`, "forward", 0, "", false},

		// 需要 ID 替换的路由
		{`^POST-/task/\d+/export`, "forward", 0, "dialogueId", true},
		{`^GET-/task/\d+`, "forward", 0, "dialogueId", false},
		{`^POST-/task/\d+`, "forward", 0, "dialogueId", false},

		// 特殊处理路由
		{`^GET-/task/list$`, "taskList", 0, "", false},
		{`^POST-/task/reduction$`, "reductionSubmit", 0, "", false},
		{`^POST-/task/reduction/check$`, "reductionCheck", 0, "", false},
		{`^POST-/dialogue/stage`, "dialogueStageSubmit", 0, "", false},
		{`^GET-/dialogue/chat`, "simpleChat", 0, "", false},
		{`^POST-/dialogue/chat`, "simpleChat", 0, "", false},
		{`^POST-/dialogue/outline$`, "outlineSubmit", 0, "", false},
		{`^GET-/dialogue/generateChart`, "chartGenerate", 0, "", false},
		{`^POST-/dialogue/templateGenerate$`, "templateGenerate", 0, "", false},
		{`^PUT-/subsite/`, "savePriceConfig", 0, "", false},

		// 工单路由
		{`^GET-/tickets`, "ticketRoute", 0, "", false},
		{`^POST-/tickets`, "ticketRoute", 0, "", false},
		{`^PUT-/tickets`, "ticketRoute", 0, "", false},

		// 知识库路由
		{`^GET-/knowledge-bases`, "knowledgeRoute", 0, "", false},
		{`^POST-/knowledge-bases`, "knowledgeRoute", 0, "", false},
		{`^PUT-/knowledge-bases`, "knowledgeRoute", 0, "", false},
		{`^DELETE-/knowledge-bases`, "knowledgeRoute", 0, "", false},

		// 点数兑换路由 - 用户端
		{`^GET-/points-exchange/products$`, "pointsExchange", 0, "", false},
		{`^POST-/points-exchange/exchange$`, "pointsExchange", 0, "", false},
		{`^GET-/points-exchange/records$`, "pointsExchange", 0, "", false},

		// 点数兑换路由 - 管理端
		{`^GET-/admin/points-exchange/products$`, "pointsExchange", 0, "", false},
		{`^POST-/admin/points-exchange/products$`, "pointsExchange", 0, "", false},
		{`^DELETE-/admin/points-exchange/products/\d+$`, "pointsExchange", 0, "", false},
		{`^GET-/admin/points-exchange/products/\d+/codes$`, "pointsExchange", 0, "", false},
		{`^POST-/admin/points-exchange/codes$`, "pointsExchange", 0, "", false},
		{`^DELETE-/admin/points-exchange/codes/\d+$`, "pointsExchange", 0, "", false},
		{`^GET-/admin/points-exchange/records$`, "pointsExchange", 0, "", false},
	}

	for _, d := range defs {
		tuboshuRoutes = append(tuboshuRoutes, routeConfig{
			pattern: regexp.MustCompile(d.pattern),
			handler: d.handler,
			ttl:     d.ttl,
			replace: d.replace,
			isBlob:  d.isBlob,
		})
	}
}

// ==================== 主路由处理 ====================

// HandleRoute 处理 tuboshu_route 请求
func (s *TuboshuService) HandleRoute(uid int, isAdmin bool, req TuboshuRouteRequest, clientIP string) (interface{}, bool, error) {
	method := strings.ToUpper(req.Method)
	path := req.Path
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	route := method + "-" + path

	for _, rc := range tuboshuRoutes {
		if rc.pattern.MatchString(route) {
			switch rc.handler {
			case "cacheable":
				result, err := s.handleCacheableRoute(method, path, req.Params, rc.ttl)
				return result, false, err

			case "forward":
				fwdPath := path
				if rc.replace == "dialogueId" {
					var err error
					fwdPath, err = s.replaceDialogueID(path, uid, isAdmin)
					if err != nil {
						return nil, false, err
					}
				}
				if rc.isBlob {
					body, err := s.upstreamRequest(fwdPath, req.Params, method, true)
					return body, true, err
				}
				result, err := s.upstreamRequestJSON(fwdPath, req.Params, method)
				return result, false, err

			case "taskList":
				return s.handleTaskList(uid, isAdmin, req.Params)

			case "reductionSubmit":
				return s.handleReductionSubmit(uid, req.Params, clientIP)

			case "reductionCheck":
				return s.handleReductionCheck(uid, req.Params)

			case "dialogueStageSubmit":
				return s.handleDialogueStageSubmit(uid, route, req.Params, clientIP)

			case "simpleChat":
				return s.handleSimpleChat(uid, method, route, req.Params, clientIP)

			case "outlineSubmit":
				return s.handleOutlineSubmit(uid, req.Params, clientIP)

			case "chartGenerate":
				return s.handleChartGenerate(uid, route, req.Params, clientIP)

			case "templateGenerate":
				return s.handleTemplateGenerate(uid, req.Params, clientIP)

			case "savePriceConfig":
				if !isAdmin {
					return nil, false, fmt.Errorf("非管理员禁止修改")
				}
				priceConfig, ok := req.Params["priceConfig"].(map[string]interface{})
				if !ok {
					return nil, false, fmt.Errorf("参数错误")
				}
				err := s.SavePriceConfig(priceConfig)
				if err != nil {
					return nil, false, err
				}
				return map[string]interface{}{"success": true}, false, nil

			case "ticketRoute":
				return s.handleTicketRoute(uid, method, path, req.Params)

			case "knowledgeRoute":
				return s.handleKnowledgeRoute(uid, method, path, req.Params)

			case "pointsExchange":
				return s.handlePointsExchange(uid, isAdmin, method, path, req.Params)
			}
		}
	}

	return nil, false, fmt.Errorf("未知的请求路由: %s", route)
}

// HandleFormDataRoute 处理文件上传请求
func (s *TuboshuService) HandleFormDataRoute(path, method string, file multipart.File, fileHeader *multipart.FileHeader) (map[string]interface{}, error) {
	token, err := s.getToken()
	if err != nil {
		return nil, err
	}

	apiURL := s.getBestAPIURL() + strings.TrimLeft(path, "/")

	// 构建 multipart body
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", fileHeader.Filename)
	if err != nil {
		return nil, fmt.Errorf("创建表单失败: %v", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("复制文件失败: %v", err)
	}
	writer.Close()

	req, err := http.NewRequest(strings.ToUpper(method), apiURL, &buf)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", token)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("响应解析失败: %v", err)
	}
	return result, nil
}

// ==================== replaceDialogueID ====================

func (s *TuboshuService) replaceDialogueID(path string, uid int, isAdmin bool) (string, error) {
	re := regexp.MustCompile(`(\d+)`)
	matches := re.FindStringSubmatch(path)
	if len(matches) < 2 {
		return path, nil
	}
	dialogueID := matches[1]

	query := "SELECT source_id, uid FROM qingka_wangke_dialogue WHERE id = ? LIMIT 1"
	var sourceID int64
	var orderUID int
	err := database.DB.QueryRow(query, dialogueID).Scan(&sourceID, &orderUID)
	if err != nil || sourceID == 0 {
		return "", fmt.Errorf("订单不存在或未完成")
	}
	if !isAdmin && orderUID != uid {
		return "", fmt.Errorf("订单不属于当前用户")
	}

	return strings.Replace(path, dialogueID, strconv.FormatInt(sourceID, 10), 1), nil
}

// ==================== 业务处理方法 ====================

// handleTaskList 获取订单列表
func (s *TuboshuService) handleTaskList(uid int, isAdmin bool, params map[string]interface{}) (interface{}, bool, error) {
	page := getIntParam(params, "page", 1)
	size := getIntParam(params, "size", 10)
	offset := (page - 1) * size

	whereClause := "1=1"
	var args []interface{}
	if !isAdmin {
		whereClause = "uid = ?"
		args = append(args, uid)
	}

	var total int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_dialogue WHERE "+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, false, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(size)))
	queryArgs := append(args, offset, size)
	rows, err := database.DB.Query(
		"SELECT id, title, state, point, addtime, download_url, type, source_id, dialogue_id FROM qingka_wangke_dialogue WHERE "+whereClause+" ORDER BY id DESC LIMIT ?, ?",
		queryArgs...,
	)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	var content []map[string]interface{}
	for rows.Next() {
		var id, sourceID int64
		var title, state, addtime, downloadURL, dtype, dialogueID string
		var point float64
		if err := rows.Scan(&id, &title, &state, &point, &addtime, &downloadURL, &dtype, &sourceID, &dialogueID); err != nil {
			continue
		}

		// 如果订单未完成且有 source_id，同步状态
		disableRefresh := map[string]bool{"PAPER_OUTLINE": true, "PAPER_WRITING": true}
		canRestart := false
		if sourceID > 0 && state != "FINISHED" && state != "REFUNDED" && !disableRefresh[dtype] {
			sourceOrder, err := s.upstreamRequestJSON(fmt.Sprintf("task/%d", sourceID), nil, "GET")
			if err == nil {
				if data, ok := sourceOrder["data"].(map[string]interface{}); ok {
					if newState, ok := data["status"].(string); ok && newState != state {
						database.DB.Exec("UPDATE qingka_wangke_dialogue SET state = ? WHERE id = ?", newState, id)
						state = newState
					}
					if dl, ok := data["downloadUrl"].(string); ok && dl != "" && downloadURL == "" {
						database.DB.Exec("UPDATE qingka_wangke_dialogue SET download_url = ? WHERE id = ?", dl, id)
						downloadURL = dl
					}
					if cr, ok := data["canRestart"].(bool); ok && cr {
						canRestart = true
					}
				}
			}
		}

		t, _ := time.Parse("2006-01-02 15:04:05", addtime)
		createTimeMs := t.UnixMilli()

		content = append(content, map[string]interface{}{
			"id":           id,
			"originPrompt": title,
			"status":       state,
			"point":        point,
			"createTime":   createTimeMs,
			"updateTime":   createTimeMs,
			"download_url": downloadURL,
			"type":         dtype,
			"sourceId":     sourceID,
			"dialogueId":   dialogueID,
			"owner":        uid,
			"canRestart":   canRestart,
		})
	}
	if content == nil {
		content = []map[string]interface{}{}
	}

	return map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"content":          content,
			"totalElements":    total,
			"totalPages":       totalPages,
			"last":             page >= totalPages,
			"first":            page <= 1,
			"empty":            len(content) == 0,
			"number":           page,
			"numberOfElements": len(content),
			"size":             size,
		},
	}, false, nil
}

// handleDialogueStageSubmit 提交对话阶段
func (s *TuboshuService) handleDialogueStageSubmit(uid int, route string, params map[string]interface{}, clientIP string) (interface{}, bool, error) {
	// 从 route 中提取 dialogueId
	re := regexp.MustCompile(`id=(\d+)`)
	matches := re.FindStringSubmatch(route)
	var dialogueID string
	if len(matches) > 1 {
		dialogueID = matches[1]
	}

	// 获取对话信息并计算价格
	price, err := s.calculateStagePrice(uid, dialogueID, params)
	if err != nil {
		return nil, false, err
	}

	// 检查余额
	if err := s.checkBalance(uid, price); err != nil {
		return nil, false, err
	}

	// 提交到上游
	path := "dialogue/stage"
	if dialogueID != "" {
		path += "?id=" + dialogueID
	}
	result, err := s.upstreamRequestJSON(path, params, "POST")
	if err != nil {
		return nil, false, err
	}

	success, _ := result["success"].(bool)
	if !success {
		msg, _ := result["message"].(string)
		return nil, false, fmt.Errorf("提交失败: %s", msg)
	}

	// 保存订单并扣费
	data, _ := result["data"].(map[string]interface{})
	sourceID := getInt64FromInterface(data["id"])
	dtype, _ := data["type"].(string)
	prompt, _ := params["prompt"].(string)

	s.saveOrderAndDeduct(uid, sourceID, prompt, dialogueID, price, dtype, clientIP)

	return map[string]interface{}{
		"success": true,
		"message": "提交成功",
		"data":    data,
	}, false, nil
}

// handleSimpleChat 简单对话
func (s *TuboshuService) handleSimpleChat(uid int, method, route string, params map[string]interface{}, clientIP string) (interface{}, bool, error) {
	// 提取 dialogueId
	re := regexp.MustCompile(`id=(\d+)`)
	matches := re.FindStringSubmatch(route)
	dialogueID := ""
	if len(matches) > 1 {
		dialogueID = matches[1]
	}
	if dialogueID == "" {
		if id, ok := params["id"]; ok {
			dialogueID = fmt.Sprintf("%v", id)
		}
	}
	if dialogueID == "" {
		return nil, false, fmt.Errorf("Missing dialogue ID")
	}

	// 获取价格
	price, err := s.getPriceFromConfig("SIMPLE_DIALOGUE", dialogueID, uid)
	if err != nil {
		return nil, false, err
	}

	// 检查余额
	if err := s.checkBalance(uid, price); err != nil {
		return nil, false, err
	}

	// 请求上游
	var result map[string]interface{}
	if method == "GET" {
		prompt, _ := params["prompt"].(string)
		result, err = s.upstreamRequestJSON("dialogue/chat", map[string]interface{}{"id": dialogueID, "prompt": prompt}, "GET")
	} else {
		result, err = s.upstreamRequestJSON("dialogue/chat?id="+dialogueID, params, "POST")
	}
	if err != nil {
		return nil, false, err
	}

	success, _ := result["success"].(bool)
	if !success {
		msg, _ := result["message"].(string)
		return nil, false, fmt.Errorf("生成失败: %s", msg)
	}

	data := result["data"]
	// 如果 data 是字符串，直接扣费返回
	if dataStr, ok := data.(string); ok {
		s.deductFee(uid, price)
		return map[string]interface{}{"success": true, "data": dataStr}, false, nil
	}

	// 保存订单
	if dataMap, ok := data.(map[string]interface{}); ok {
		sourceID := getInt64FromInterface(dataMap["id"])
		dtype, _ := dataMap["type"].(string)
		prompt := ""
		if p, ok := params["prompt"].(string); ok {
			prompt = p
		} else if p, ok := params["content"].(string); ok {
			prompt = p
		}
		orderID := s.saveOrderAndDeduct(uid, sourceID, prompt, dialogueID, price, dtype, clientIP)
		dataMap["id"] = orderID
		return map[string]interface{}{"success": true, "data": dataMap}, false, nil
	}

	return result, false, nil
}

// handleOutlineSubmit 生成大纲
func (s *TuboshuService) handleOutlineSubmit(uid int, params map[string]interface{}, clientIP string) (interface{}, bool, error) {
	price, err := s.getPriceFromConfig("PAPER_OUTLINE", "", uid)
	if err != nil {
		return nil, false, err
	}
	if err := s.checkBalance(uid, price); err != nil {
		return nil, false, err
	}

	prompt, _ := params["prompt"].(string)
	result, err := s.upstreamRequestJSON("dialogue/outline", map[string]interface{}{"prompt": prompt}, "POST")
	if err != nil {
		return nil, false, err
	}

	success, _ := result["success"].(bool)
	if !success {
		msg, _ := result["message"].(string)
		return nil, false, fmt.Errorf("生成失败: %s", msg)
	}

	data, _ := result["data"].(map[string]interface{})
	sourceID := getInt64FromInterface(data["id"])
	s.saveOrderAndDeduct(uid, sourceID, prompt, "outline", price, "PAPER_OUTLINE", clientIP)

	return map[string]interface{}{"success": true, "data": data}, false, nil
}

// handleChartGenerate 图表生成
func (s *TuboshuService) handleChartGenerate(uid int, route string, params map[string]interface{}, clientIP string) (interface{}, bool, error) {
	price, err := s.getPriceFromConfig("CHART_GENERATE", "", uid)
	if err != nil {
		return nil, false, err
	}
	if err := s.checkBalance(uid, price); err != nil {
		return nil, false, err
	}

	result, err := s.upstreamRequestJSON("dialogue/generateChart", params, "GET")
	if err != nil {
		return nil, false, err
	}

	success, _ := result["success"].(bool)
	if !success {
		msg, _ := result["message"].(string)
		return nil, false, fmt.Errorf("生成失败: %s", msg)
	}

	data, _ := result["data"].(map[string]interface{})
	sourceID := getInt64FromInterface(data["id"])
	prompt, _ := params["prompt"].(string)
	if prompt == "" {
		prompt = "图表生成"
	}
	s.saveOrderAndDeduct(uid, sourceID, prompt, "chart", price, "CHART_GENERATE", clientIP)

	return result, false, nil
}

// handleTemplateGenerate 模板生成
func (s *TuboshuService) handleTemplateGenerate(uid int, params map[string]interface{}, clientIP string) (interface{}, bool, error) {
	price, err := s.getPriceFromConfig("PAPER_OUTLINE", "", uid)
	if err != nil {
		return nil, false, err
	}
	if err := s.checkBalance(uid, price); err != nil {
		return nil, false, err
	}

	result, err := s.upstreamRequestJSON("dialogue/templateGenerate", params, "POST")
	if err != nil {
		return nil, false, err
	}

	success, _ := result["success"].(bool)
	if !success {
		msg, _ := result["message"].(string)
		return nil, false, fmt.Errorf("生成失败: %s", msg)
	}

	data, _ := result["data"].(map[string]interface{})
	sourceID := getInt64FromInterface(data["id"])
	s.saveOrderAndDeduct(uid, sourceID, "Word模板智能填充", "templateGenerate", price, "TEMPLATE", clientIP)

	return result, false, nil
}

// handleReductionSubmit 文本降重提交
func (s *TuboshuService) handleReductionSubmit(uid int, params map[string]interface{}, clientIP string) (interface{}, bool, error) {
	text, _ := params["text"].(string)
	if text == "" {
		return nil, false, fmt.Errorf("文本不能为空")
	}
	reductionType, _ := params["type"].(string)
	if reductionType == "" {
		reductionType = "NORMAL"
	}

	// 先检查
	checkResult, _, err := s.handleReductionCheck(uid, params)
	if err != nil {
		return nil, false, err
	}
	checkMap, ok := checkResult.(map[string]interface{})
	if !ok {
		return nil, false, fmt.Errorf("检查失败")
	}
	success, _ := checkMap["success"].(bool)
	if !success {
		return nil, false, fmt.Errorf("检查失败")
	}

	checkData, _ := checkMap["data"].(map[string]interface{})
	price := getFloat64FromInterface(checkData["price"])
	charCount := getIntFromInterface(checkData["totalCharCount"])
	balanceEnough, _ := checkData["balanceEnough"].(bool)

	if !balanceEnough {
		return nil, false, fmt.Errorf("余额不足")
	}

	// 调用上游 API
	result, err := s.upstreamRequestJSON("task/reduction", map[string]interface{}{"text": text, "type": reductionType}, "POST")
	if err != nil {
		return nil, false, err
	}

	resultSuccess, _ := result["success"].(bool)
	if !resultSuccess {
		msg, _ := result["message"].(string)
		return nil, false, fmt.Errorf("生成失败: %s", msg)
	}

	data, _ := result["data"].(map[string]interface{})
	sourceID := getInt64FromInterface(data["id"])
	title := fmt.Sprintf("文本降重 (%d字) - %s", charCount, reductionType)
	orderID := s.saveOrderAndDeduct(uid, sourceID, title, "reduction", price, "PAPER_REDUCTION", clientIP)
	data["id"] = orderID

	return result, false, nil
}

// handleReductionCheck 文本降重检查
func (s *TuboshuService) handleReductionCheck(uid int, params map[string]interface{}) (interface{}, bool, error) {
	text, _ := params["text"].(string)
	if text == "" {
		return nil, false, fmt.Errorf("文本不能为空")
	}
	reductionType, _ := params["type"].(string)
	if reductionType == "" {
		reductionType = "NORMAL"
	}

	pricePerThousand, err := s.getPriceFromConfig("PAPER_REDUCTION", reductionType, uid)
	if err != nil {
		return nil, false, err
	}

	result, err := s.upstreamRequestJSON("task/reduction/check", map[string]interface{}{"text": text, "type": reductionType}, "POST")
	if err != nil {
		return nil, false, err
	}

	success, _ := result["success"].(bool)
	if !success {
		msg, _ := result["message"].(string)
		return nil, false, fmt.Errorf("检查失败: %s", msg)
	}

	data, _ := result["data"].(map[string]interface{})
	charCount := getFloat64FromInterface(data["totalCharCount"])
	calculatedPrice := math.Round((charCount/1000)*pricePerThousand*1000) / 1000
	data["price"] = calculatedPrice

	// 检查余额
	var money float64
	database.DB.QueryRow("SELECT money FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&money)
	data["balanceEnough"] = money >= calculatedPrice
	data["balance"] = money

	return result, false, nil
}

// handleTicketRoute 工单路由转发
func (s *TuboshuService) handleTicketRoute(uid int, method, path string, params map[string]interface{}) (interface{}, bool, error) {
	cleanPath := strings.TrimLeft(path, "/")

	// 工单列表 - 注入 subUid
	if method == "GET" && cleanPath == "tickets" {
		if _, ok := params["subUid"]; !ok {
			params["subUid"] = strconv.Itoa(uid)
		}
		apiPath := "tickets/by-sub-uid/" + fmt.Sprintf("%v", params["subUid"])
		result, err := s.upstreamRequestJSON(apiPath, params, "GET")
		return result, false, err
	}

	// 创建工单 - 注入 subUid
	if method == "POST" && cleanPath == "tickets" {
		if subUid, ok := params["subUid"].(string); ok && subUid != "" {
			params["subUid"] = subUid + "_" + strconv.Itoa(uid)
		} else {
			params["subUid"] = strconv.Itoa(uid)
		}
		result, err := s.upstreamRequestJSON("tickets", params, "POST")
		return result, false, err
	}

	// 其他工单请求直接转发
	result, err := s.upstreamRequestJSON(cleanPath, params, method)
	return result, false, err
}

// handleKnowledgeRoute 知识库路由转发
func (s *TuboshuService) handleKnowledgeRoute(uid int, method, path string, params map[string]interface{}) (interface{}, bool, error) {
	cleanPath := strings.TrimLeft(path, "/")

	// 知识库列表 - 通过 subUid 查询
	if method == "GET" && cleanPath == "knowledge-bases" {
		subUid := strconv.Itoa(uid)
		if su, ok := params["subUid"].(string); ok && su != "" {
			subUid = su + "_" + strconv.Itoa(uid)
		}
		apiPath := "knowledge-bases/by-sub-uid/" + subUid
		result, err := s.upstreamRequestJSON(apiPath, params, "GET")
		return result, false, err
	}

	// 创建知识库 - 注入 subUid
	if method == "POST" && cleanPath == "knowledge-bases" {
		if su, ok := params["subUid"].(string); ok && su != "" {
			params["subUid"] = su + "_" + strconv.Itoa(uid)
		} else {
			params["subUid"] = strconv.Itoa(uid)
		}
		result, err := s.upstreamRequestJSON("knowledge-bases", params, "POST")
		return result, false, err
	}

	// 其他直接转发
	result, err := s.upstreamRequestJSON(cleanPath, params, method)
	return result, false, err
}

// handlePointsExchange 点数兑换处理
func (s *TuboshuService) handlePointsExchange(uid int, isAdmin bool, method, path string, params map[string]interface{}) (interface{}, bool, error) {
	cleanPath := strings.TrimLeft(path, "/")

	// ---- 用户端 ----
	if method == "GET" && cleanPath == "points-exchange/products" {
		data, err := s.getAvailableProducts()
		return map[string]interface{}{"success": true, "data": data}, false, err
	}
	if method == "POST" && cleanPath == "points-exchange/exchange" {
		productID := getIntFromInterface(params["productId"])
		data, err := s.exchangeProduct(uid, productID)
		return map[string]interface{}{"success": true, "data": data}, false, err
	}
	if method == "GET" && cleanPath == "points-exchange/records" {
		page := getIntParam(params, "page", 1)
		pageSize := getIntParam(params, "pageSize", 10)
		data, err := s.getExchangeRecords(uid, page, pageSize)
		return map[string]interface{}{"success": true, "data": data}, false, err
	}

	// ---- 管理端 ----
	if !isAdmin {
		return nil, false, fmt.Errorf("需要管理员权限")
	}

	if method == "GET" && cleanPath == "admin/points-exchange/products" {
		page := getIntParam(params, "page", 1)
		pageSize := getIntParam(params, "pageSize", 10)
		data, err := s.getAdminProducts(page, pageSize)
		return map[string]interface{}{"success": true, "data": data}, false, err
	}
	if method == "POST" && cleanPath == "admin/points-exchange/products" {
		err := s.saveProduct(params)
		return map[string]interface{}{"success": true}, false, err
	}
	if method == "DELETE" && strings.HasPrefix(cleanPath, "admin/points-exchange/products/") {
		re := regexp.MustCompile(`products/(\d+)$`)
		m := re.FindStringSubmatch(cleanPath)
		if len(m) > 1 {
			id, _ := strconv.Atoi(m[1])
			err := s.deleteProduct(id)
			return map[string]interface{}{"success": true}, false, err
		}
	}
	if method == "GET" && strings.Contains(cleanPath, "/codes") {
		re := regexp.MustCompile(`products/(\d+)/codes`)
		m := re.FindStringSubmatch(cleanPath)
		if len(m) > 1 {
			productID, _ := strconv.Atoi(m[1])
			page := getIntParam(params, "page", 1)
			pageSize := getIntParam(params, "pageSize", 20)
			data, err := s.getProductCodes(productID, page, pageSize)
			return map[string]interface{}{"success": true, "data": data}, false, err
		}
	}
	if method == "POST" && cleanPath == "admin/points-exchange/codes" {
		productID := getIntFromInterface(params["productId"])
		codes, _ := params["codes"].([]interface{})
		err := s.addCodes(productID, codes)
		return map[string]interface{}{"success": true}, false, err
	}
	if method == "DELETE" && strings.HasPrefix(cleanPath, "admin/points-exchange/codes/") {
		re := regexp.MustCompile(`codes/(\d+)$`)
		m := re.FindStringSubmatch(cleanPath)
		if len(m) > 1 {
			id, _ := strconv.Atoi(m[1])
			err := s.deleteCode(id)
			return map[string]interface{}{"success": true}, false, err
		}
	}
	if method == "GET" && cleanPath == "admin/points-exchange/records" {
		page := getIntParam(params, "page", 1)
		pageSize := getIntParam(params, "pageSize", 20)
		data, err := s.getAdminExchangeRecords(page, pageSize)
		return map[string]interface{}{"success": true, "data": data}, false, err
	}

	return nil, false, fmt.Errorf("未知的点数兑换路由")
}

// ==================== 价格计算 ====================

func (s *TuboshuService) calculateStagePrice(uid int, dialogueID string, options map[string]interface{}) (float64, error) {
	cfg, err := s.GetConfig()
	if err != nil {
		return 0, err
	}

	// 获取对话信息判断类型
	dialogues, err := s.upstreamRequestJSON("dialogue/stage", nil, "GET")
	if err != nil {
		return 0, err
	}

	// 在对话列表中查找
	var dialogueName string
	if data, ok := dialogues["data"].([]interface{}); ok {
		for _, item := range data {
			if d, ok := item.(map[string]interface{}); ok {
				if fmt.Sprintf("%v", d["id"]) == dialogueID {
					dialogueName, _ = d["name"].(string)
					break
				}
			}
		}
	}

	if dialogueName == "论文撰写" {
		// 论文撰写特殊计算
		priceResult, err := s.upstreamRequestJSON("dialogue/stage/price", options, "POST")
		if err != nil {
			return 0, fmt.Errorf("获取价格失败")
		}
		priceSuccess, _ := priceResult["success"].(bool)
		if !priceSuccess {
			return 0, fmt.Errorf("获取价格失败")
		}

		priceData, _ := priceResult["data"].(map[string]interface{})
		outlineLength := getFloat64FromInterface(priceData["outlineLength"])
		pointLength := getFloat64FromInterface(priceData["pointLength"])
		useReduction := false
		if ur, ok := priceData["useAigcReduction"].(bool); ok {
			useReduction = ur
		}

		pwConfig := s.getPaperWritingConfig(cfg)
		price := pwConfig["sectionBasePrice"].(float64)*outlineLength +
			pwConfig["pointBasePrice"].(float64)*pointLength
		if useReduction {
			price += pwConfig["reductionExtraPrice"].(float64)
		}

		var addprice float64
		database.DB.QueryRow("SELECT addprice FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice)
		price = price * addprice * cfg.PriceRatio
		return price, nil
	}

	// 其他类型
	return s.getPriceFromConfig("STAGE_DIALOGUE", dialogueID, uid)
}

func (s *TuboshuService) getPaperWritingConfig(cfg *TuboshuConfig) map[string]interface{} {
	if pw, ok := cfg.PriceConfig["PAPER_WRITING"].(map[string]interface{}); ok {
		if c, ok := pw["config"].(map[string]interface{}); ok {
			return c
		}
	}
	return map[string]interface{}{
		"sectionBasePrice": 1.6, "pointBasePrice": 1.0,
		"reductionExtraPrice": 10.0, "v3ModelExtraPrice": 0.0,
	}
}

func (s *TuboshuService) getPriceFromConfig(priceType, key string, uid int) (float64, error) {
	cfg, err := s.GetConfig()
	if err != nil {
		return 0, err
	}

	typeConfig, ok := cfg.PriceConfig[priceType].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("价格配置无效: %s", priceType)
	}

	if priceType == "PAPER_WRITING" {
		return 0, fmt.Errorf("论文撰写使用 calculateStagePrice 计算")
	}

	enabled, _ := typeConfig["enabled"].(bool)
	if !enabled {
		return 0, fmt.Errorf("该功能未启用")
	}

	configType, _ := typeConfig["type"].(string)
	var addprice float64
	database.DB.QueryRow("SELECT addprice FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice)

	if configType == "id_based" {
		prices, ok := typeConfig["prices"].(map[string]interface{})
		if !ok {
			return 0, fmt.Errorf("价格配置缺失")
		}
		priceVal, ok := prices[key]
		if !ok {
			return 0, fmt.Errorf("该类型未配置价格: %s", key)
		}
		return getFloat64FromInterface(priceVal) * addprice * cfg.PriceRatio, nil
	}

	// fixed 类型
	priceVal, ok := typeConfig["price"]
	if !ok {
		return 0, fmt.Errorf("固定价格未配置")
	}
	return getFloat64FromInterface(priceVal) * addprice * cfg.PriceRatio, nil
}

// ==================== 余额与扣费 ====================

func (s *TuboshuService) checkBalance(uid int, price float64) error {
	var money float64
	err := database.DB.QueryRow("SELECT money FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&money)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}
	if money < price {
		return fmt.Errorf("余额不足，当前余额：%.2f，需要：%.2f", money, price)
	}
	return nil
}

func (s *TuboshuService) deductFee(uid int, price float64) {
	database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ? AND money >= ?", price, uid, price)
	// 写入资金日志
	database.DB.Exec("INSERT INTO qingka_wangke_moneylog (uid, type, money, remarks, addtime) VALUES (?, '论文下单', ?, ?, NOW())",
		uid, -price, fmt.Sprintf("扣除%.2f元", price))
}

func (s *TuboshuService) saveOrderAndDeduct(uid int, sourceID int64, title, dialogueID string, price float64, dtype, clientIP string) int64 {
	now := time.Now().Format("2006-01-02 15:04:05")
	result, err := database.DB.Exec(
		"INSERT INTO qingka_wangke_dialogue (uid, title, state, addtime, ip, source_id, dialogue_id, point, download_url, type) VALUES (?, ?, 'PENDING', ?, ?, ?, ?, ?, '', ?)",
		uid, title, now, clientIP, sourceID, dialogueID, price, dtype,
	)
	if err != nil {
		log.Printf("[Tuboshu] 保存订单失败: %v", err)
		return 0
	}
	orderID, _ := result.LastInsertId()

	s.deductFee(uid, price)
	return orderID
}

// ==================== 点数兑换 ====================

func (s *TuboshuService) getAvailableProducts() ([]map[string]interface{}, error) {
	rows, err := database.DB.Query(`SELECT id, name, description, image_url, price, sort_order
		FROM points_product WHERE status = 'ENABLED'
		AND id IN (SELECT DISTINCT product_id FROM points_product_code WHERE status = 'AVAILABLE')
		ORDER BY sort_order DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []map[string]interface{}
	for rows.Next() {
		var id, sortOrder int
		var name, desc, imageURL string
		var price float64
		rows.Scan(&id, &name, &desc, &imageURL, &price, &sortOrder)

		var stock int
		database.DB.QueryRow("SELECT COUNT(*) FROM points_product_code WHERE product_id = ? AND status = 'AVAILABLE'", id).Scan(&stock)

		products = append(products, map[string]interface{}{
			"id": id, "name": name, "description": desc,
			"image_url": imageURL, "price": price, "stock": stock,
		})
	}
	if products == nil {
		products = []map[string]interface{}{}
	}
	return products, nil
}

func (s *TuboshuService) exchangeProduct(uid, productID int) (map[string]interface{}, error) {
	// 查询商品
	var name string
	var price float64
	var status string
	err := database.DB.QueryRow("SELECT name, price, status FROM points_product WHERE id = ?", productID).Scan(&name, &price, &status)
	if err != nil {
		return nil, fmt.Errorf("商品不存在")
	}
	if status != "ENABLED" {
		return nil, fmt.Errorf("商品未上架")
	}

	// 检查余额
	if err := s.checkBalance(uid, price); err != nil {
		return nil, err
	}

	// 获取可用兑换码
	var codeID int
	var code string
	err = database.DB.QueryRow("SELECT id, code FROM points_product_code WHERE product_id = ? AND status = 'AVAILABLE' LIMIT 1", productID).Scan(&codeID, &code)
	if err != nil {
		return nil, fmt.Errorf("库存不足")
	}

	// 标记兑换码为已使用
	res, err := database.DB.Exec("UPDATE points_product_code SET status = 'EXCHANGED', exchanged_by = ?, exchanged_at = NOW() WHERE id = ? AND status = 'AVAILABLE'", uid, codeID)
	if err != nil {
		return nil, fmt.Errorf("兑换失败")
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return nil, fmt.Errorf("兑换码已被使用")
	}

	// 扣费
	s.deductFee(uid, price)

	// 记录
	database.DB.Exec("INSERT INTO points_exchange_record (uid, product_id, product_name, code_id, code, points_cost) VALUES (?, ?, ?, ?, ?, ?)",
		uid, productID, name, codeID, code, price)

	return map[string]interface{}{"code": code, "product_name": name, "cost": price}, nil
}

func (s *TuboshuService) getExchangeRecords(uid, page, pageSize int) (map[string]interface{}, error) {
	offset := (page - 1) * pageSize
	var total int
	database.DB.QueryRow("SELECT COUNT(*) FROM points_exchange_record WHERE uid = ?", uid).Scan(&total)

	rows, err := database.DB.Query("SELECT id, product_name, code, points_cost, create_time FROM points_exchange_record WHERE uid = ? ORDER BY id DESC LIMIT ?, ?", uid, offset, pageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []map[string]interface{}
	for rows.Next() {
		var id int
		var productName, code, createTime string
		var cost float64
		rows.Scan(&id, &productName, &code, &cost, &createTime)
		records = append(records, map[string]interface{}{
			"id": id, "product_name": productName, "code": code, "cost": cost, "create_time": createTime,
		})
	}
	if records == nil {
		records = []map[string]interface{}{}
	}
	return map[string]interface{}{"list": records, "total": total, "page": page, "pageSize": pageSize}, nil
}

func (s *TuboshuService) getAdminProducts(page, pageSize int) (map[string]interface{}, error) {
	offset := (page - 1) * pageSize
	var total int
	database.DB.QueryRow("SELECT COUNT(*) FROM points_product").Scan(&total)

	rows, err := database.DB.Query("SELECT id, name, description, image_url, price, status, sort_order, create_time FROM points_product ORDER BY sort_order DESC, id DESC LIMIT ?, ?", offset, pageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []map[string]interface{}
	for rows.Next() {
		var id, sortOrder int
		var name, desc, imageURL, status, createTime string
		var price float64
		rows.Scan(&id, &name, &desc, &imageURL, &price, &status, &sortOrder, &createTime)

		var stock int
		database.DB.QueryRow("SELECT COUNT(*) FROM points_product_code WHERE product_id = ? AND status = 'AVAILABLE'", id).Scan(&stock)

		products = append(products, map[string]interface{}{
			"id": id, "name": name, "description": desc, "image_url": imageURL,
			"price": price, "status": status, "sort_order": sortOrder,
			"create_time": createTime, "stock": stock,
		})
	}
	if products == nil {
		products = []map[string]interface{}{}
	}
	return map[string]interface{}{"list": products, "total": total, "page": page, "pageSize": pageSize}, nil
}

func (s *TuboshuService) saveProduct(params map[string]interface{}) error {
	name, _ := params["name"].(string)
	desc, _ := params["description"].(string)
	imageURL, _ := params["image_url"].(string)
	price := getFloat64FromInterface(params["price"])
	status, _ := params["status"].(string)
	if status == "" {
		status = "ENABLED"
	}
	sortOrder := getIntFromInterface(params["sort_order"])
	id := getIntFromInterface(params["id"])

	if id > 0 {
		_, err := database.DB.Exec("UPDATE points_product SET name=?, description=?, image_url=?, price=?, status=?, sort_order=? WHERE id=?",
			name, desc, imageURL, price, status, sortOrder, id)
		return err
	}
	_, err := database.DB.Exec("INSERT INTO points_product (name, description, image_url, price, status, sort_order) VALUES (?, ?, ?, ?, ?, ?)",
		name, desc, imageURL, price, status, sortOrder)
	return err
}

func (s *TuboshuService) deleteProduct(id int) error {
	_, err := database.DB.Exec("DELETE FROM points_product WHERE id = ?", id)
	return err
}

func (s *TuboshuService) getProductCodes(productID, page, pageSize int) (map[string]interface{}, error) {
	offset := (page - 1) * pageSize
	var total int
	database.DB.QueryRow("SELECT COUNT(*) FROM points_product_code WHERE product_id = ?", productID).Scan(&total)

	rows, err := database.DB.Query("SELECT id, code, status, exchanged_by, exchanged_at, create_time FROM points_product_code WHERE product_id = ? ORDER BY id DESC LIMIT ?, ?", productID, offset, pageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var codes []map[string]interface{}
	for rows.Next() {
		var id int
		var code, status string
		var exchangedBy *int
		var exchangedAt, createTime *string
		rows.Scan(&id, &code, &status, &exchangedBy, &exchangedAt, &createTime)
		item := map[string]interface{}{
			"id": id, "code": code, "status": status, "create_time": createTime,
		}
		if exchangedBy != nil {
			item["exchanged_by"] = *exchangedBy
		}
		if exchangedAt != nil {
			item["exchanged_at"] = *exchangedAt
		}
		codes = append(codes, item)
	}
	if codes == nil {
		codes = []map[string]interface{}{}
	}
	return map[string]interface{}{"list": codes, "total": total, "page": page, "pageSize": pageSize}, nil
}

func (s *TuboshuService) addCodes(productID int, codes []interface{}) error {
	if productID <= 0 {
		return fmt.Errorf("商品ID无效")
	}
	for _, c := range codes {
		code := fmt.Sprintf("%v", c)
		if code != "" {
			database.DB.Exec("INSERT INTO points_product_code (product_id, code) VALUES (?, ?)", productID, code)
		}
	}
	return nil
}

func (s *TuboshuService) deleteCode(id int) error {
	_, err := database.DB.Exec("DELETE FROM points_product_code WHERE id = ? AND status = 'AVAILABLE'", id)
	return err
}

func (s *TuboshuService) getAdminExchangeRecords(page, pageSize int) (map[string]interface{}, error) {
	offset := (page - 1) * pageSize
	var total int
	database.DB.QueryRow("SELECT COUNT(*) FROM points_exchange_record").Scan(&total)

	rows, err := database.DB.Query("SELECT id, uid, product_name, code, points_cost, create_time FROM points_exchange_record ORDER BY id DESC LIMIT ?, ?", offset, pageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []map[string]interface{}
	for rows.Next() {
		var id, uid int
		var productName, code, createTime string
		var cost float64
		rows.Scan(&id, &uid, &productName, &code, &cost, &createTime)
		records = append(records, map[string]interface{}{
			"id": id, "uid": uid, "product_name": productName, "code": code, "cost": cost, "create_time": createTime,
		})
	}
	if records == nil {
		records = []map[string]interface{}{}
	}
	return map[string]interface{}{"list": records, "total": total, "page": page, "pageSize": pageSize}, nil
}

// ==================== 辅助函数 ====================

func getIntParam(params map[string]interface{}, key string, defaultVal int) int {
	if v, ok := params[key]; ok {
		return getIntFromInterface(v)
	}
	return defaultVal
}

func getIntFromInterface(v interface{}) int {
	switch val := v.(type) {
	case float64:
		return int(val)
	case int:
		return val
	case int64:
		return int(val)
	case string:
		n, _ := strconv.Atoi(val)
		return n
	case json.Number:
		n, _ := val.Int64()
		return int(n)
	}
	return 0
}

func getInt64FromInterface(v interface{}) int64 {
	switch val := v.(type) {
	case float64:
		return int64(val)
	case int:
		return int64(val)
	case int64:
		return val
	case string:
		n, _ := strconv.ParseInt(val, 10, 64)
		return n
	case json.Number:
		n, _ := val.Int64()
		return n
	}
	return 0
}

func getFloat64FromInterface(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		n, _ := strconv.ParseFloat(val, 64)
		return n
	case json.Number:
		n, _ := val.Float64()
		return n
	}
	return 0
}
