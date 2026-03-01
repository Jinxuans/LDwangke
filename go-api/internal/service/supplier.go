package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-api/internal/database"
	"go-api/internal/model"
)

// hostRateLimiter 每个上游主机的请求速率限制器
// 防止并发请求打爆同一个上游服务器
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

// wait 对指定 host 限速：同一 host 两次请求之间至少间隔 interval
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

type SupplierService struct {
	client *http.Client
}

// 全局共享的 HTTP 客户端，复用连接池
var sharedHTTPClient = &http.Client{
	Timeout: 60 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        50,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	},
}

func NewSupplierService() *SupplierService {
	return &SupplierService{
		client: sharedHTTPClient,
	}
}

// PlatformConfig 平台接口差异配置（按 PHP Checkorder 各平台 if/else 提取）
type PlatformConfig struct {
	QueryAct          string // 查课 act，默认 "get"；"local_time" 表示本地生成时长列表；"local_script" 表示不支持
	OrderAct          string // 下单 act，默认 "add"
	ProgressAct       string // 有 yid 时的进度 act，默认 "chadan2"
	ProgressNoYID     string // 无 yid 时的进度 act，默认 "chadan"
	ProgressPath      string // 非标准进度 URL 路径（覆盖 act 拼接），如 "/api/chadan1"、"/api/search"
	ProgressMethod    string // 进度请求方式，默认 "POST"；部分平台用 "GET"
	SuccessCode       string // 上游下单成功码，默认 "0"（大部分平台）；hzw 为 "1"
	ReturnsYID        bool   // 下单响应是否包含 yid
	ExtraParams       bool   // 下单是否传 score/shichang 额外参数
	UseIDParam        bool   // 进度查询用 "id" 参数代替 "yid"（hzw 平台）
	AlwaysUsername    bool   // 进度查询始终传 username（hzw 平台）
	YIDInDataArray    bool   // 下单响应 yid 在 data 数组里（龙龙：data:["yid"]）
	UseUUIDParam      bool   // 进度查询用 "uuid" 参数代替 "yid"（龙龙平台）
	PauseAct          string // 暂停/恢复 act，默认 "zt"；龙龙用 "zanting"；hzw 用 "stop"
	PauseIDParam      string // 暂停订单 ID 参数名，默认 "id"；Benz 用 "oid"
	ChangePassAct     string // 改密 act，默认 "gaimi"
	PausePath         string // 非标准暂停路径（覆盖 act 拼接），如 "/api/stop" (2xx)
	ChangePassPath    string // 非标准改密路径，如 "/api/update" (2xx)
	UseJSON           bool   // 暂停/改密是否用 JSON body（2xx）
	LogPath           string // 非标准日志路径，如 "/log/" (KUN/kunba)
	LogMethod         string // 日志查询方式，默认用 act=xq POST；KUN 用 GET
	LogAct            string // 日志查询 act，默认 "xq"；hzw 用 "cha_logwk"；pup 用 "orderlog"
	LogIDParam        string // 日志 ID 参数名，默认 "id"；pup 用 "oid"
	ChangePassParam   string // 改密密码参数名，默认 "newPwd"；29/spi 用 "xgmm"；pup 用 "newpwd"
	ChangePassIDParam string // 改密订单 ID 参数名，默认 "id"；pup 用 "oid"
	ResubmitPath      string // 非标准补单路径，如 "/api/reset" (2xx)
	ResubmitIDParam   string // 补单订单 ID 参数名，默认 "id"；pup 用 "oid"
	BalanceAct        string // 余额查询 act，默认 "getmoney"；longlong 用 "money"
	BalancePath       string // 余额 REST 路径（覆盖 act），如 "/api/getinfo" (2xx)、"/api/getuserinfo/" (nx)
	BalanceMoneyField string // 余额字段路径："money"(根级) / "data.money" / "data" / "data.remainscore"
	BalanceMethod     string // 余额请求方式，默认 "POST"
	BalanceAuthType   string // 余额认证覆盖（空=跟随全局）；nx 用 "bearer_token"
	ReportAct         string // 反馈提交 act，默认 "report"
	ReportPath        string // 非标准反馈路径（覆盖 act 拼接），如 "/api/submitWork" (2xx)
	GetReportAct      string // 反馈查询 act，默认 "getReport"
	GetReportPath     string // 非标准反馈查询路径，如 "/api/queryWork" (2xx)
	ReportSuccessCode string // 反馈成功码，默认 "1"
	ReportParamStyle  string // 反馈参数风格："standard"(uid/key/id/question) / "token"(token/type/id/content)
	ReportAuthType    string // 反馈认证方式：空=跟随全局(uid_key) / "token_only"
	RefreshPath       string // 非标准刷新路径，如 "/api/refresh" (2xx)
	CategoryAct       string // 获取分类 act，默认 "getcate"；天河(skyriver) 用 "getfl"
}

// platformRegistry 所有已知平台配置（按 PHP Checkorder/ckjk.php + xdjk.php + jdjk.php 提取）
var platformRegistry = map[string]PlatformConfig{
	// 春秋 - 查课返回本地时长列表，下单成功码"0"
	"27": {QueryAct: "local_time", SuccessCode: "0", ExtraParams: true},
	// 志塬 - 同春秋，本地时长列表
	"zy": {QueryAct: "local_time", SuccessCode: "0"},
	// 乐学 - 标准查课，进度用 /api/search GET
	"haha": {SuccessCode: "0", ReturnsYID: true, ExtraParams: true, ProgressPath: "/api/search", ProgressMethod: "GET"},
	// hzw - 成功码"1"，返回 yid，进度用 chadan+id 参数 +username，暂停用 stop，日志用 cha_logwk
	"hzw": {SuccessCode: "1", ReturnsYID: true, ProgressAct: "chadan", ProgressNoYID: "chadan", UseIDParam: true, AlwaysUsername: true, PauseAct: "stop", LogAct: "cha_logwk", BalanceMoneyField: "money"},
	// 龙龙 - 下单 code=0, yid 在 data 数组，查单用 uuid 参数
	"longlong": {SuccessCode: "0", ReturnsYID: true, YIDInDataArray: true, UseUUIDParam: true, ProgressAct: "chadan", ProgressNoYID: "chadan", PauseAct: "zanting", BalanceAct: "money", BalanceMoneyField: "data"},
	// 流年 - 进度用 /api/chadan1，改密用 xgmm
	"liunian": {SuccessCode: "0", ReturnsYID: true, ProgressPath: "/api/chadan1", ChangePassAct: "xgmm", ChangePassParam: "xgmm"},
	// 学习通官方 - 本地脚本，新系统暂不支持
	"xxtgf": {QueryAct: "local_script"},
	// 毛豆 mooc - 本地脚本，新系统暂不支持
	"moocmd": {QueryAct: "local_script"},
	// yyy - 完全自定义 API（JSON+Bearer token），独立适配器处理
	"yyy": {QueryAct: "yyy_custom", SuccessCode: "200", ReturnsYID: true, BalanceMoneyField: "money"},
	// 爱学习 (2xx) - JSON API，暂停用/api/stop，改密用/api/update，补单用/api/reset
	"2xx": {SuccessCode: "0", ReturnsYID: true, PausePath: "/api/stop", ChangePassPath: "/api/update", ResubmitPath: "/api/reset", UseJSON: true, BalancePath: "/api/getinfo", BalanceMoneyField: "data.money", ReportPath: "/api/submitWork", GetReportPath: "/api/queryWork", ReportParamStyle: "token", ReportAuthType: "token_only"},
	// KUN - 自定义接口：GET /query/ 查课、GET /getorder/ 下单、GET /upPwd/ 改密，日志用 /log/ GET
	"KUN": {QueryAct: "KUN_custom", SuccessCode: "0", LogPath: "/log/", LogMethod: "GET"},
	// kunba - 同 KUN
	"kunba": {QueryAct: "KUN_custom", SuccessCode: "0", LogPath: "/log/", LogMethod: "GET"},
	// Benz(奔驰) - 标准接口，改密用 xgmm(oid/pwd)，日志用 getOrderLogs(oid)，暂停用 ztdd(oid)
	"Benz": {SuccessCode: "0", ReturnsYID: true, ProgressAct: "chadan", ProgressNoYID: "chadan", ChangePassAct: "xgmm", ChangePassIDParam: "oid", ChangePassParam: "pwd", LogAct: "getOrderLogs", LogIDParam: "oid", PauseAct: "ztdd", PauseIDParam: "oid"},
	// 土拨鼠 - 完全自定义（论文/AI 服务，JSON+satoken），独立适配
	"tuboshu": {QueryAct: "tuboshu_custom", SuccessCode: "0", BalanceMoneyField: "data.money"},
	// 29 - 标准接口，改密用 xgmm
	"29": {SuccessCode: "0", ChangePassAct: "xgmm", ChangePassParam: "xgmm", BalanceMoneyField: "money"},
	// spi(spiderman) - 进度用/api/search GET，改密用 xgmm
	"spi": {SuccessCode: "0", ReturnsYID: true, ProgressPath: "/api/search", ProgressMethod: "GET", ChangePassAct: "xgmm", ChangePassParam: "newPwd", BalanceMoneyField: "money"},
	// lg(学习平台) - 标准接口，返回 yid
	"lg": {SuccessCode: "0", ReturnsYID: true, BalanceMoneyField: "data.money"},
	// nx(奶昔) - 完全自定义 REST API（token+proxy），独立适配
	"nx": {QueryAct: "nx_custom", SuccessCode: "0", ReturnsYID: true, BalancePath: "/api/getuserinfo/", BalanceMoneyField: "data.remainscore", BalanceAuthType: "bearer_token"},
	// pup - 标准接口，支持 score/duration/period 额外参数，进度只有 chadan（无 chadan2），改密用 updateorderpwd(oid/newpwd)，日志用 orderlog(oid)，补单用 budan(oid)
	"pup": {SuccessCode: "0", ReturnsYID: true, ExtraParams: true, ProgressAct: "chadan", ProgressNoYID: "chadan", ChangePassAct: "updateorderpwd", ChangePassIDParam: "oid", ChangePassParam: "newpwd", LogAct: "orderlog", LogIDParam: "oid", ResubmitIDParam: "oid"},
	// wanzi(丸子) - 成功码"1"，进度用chadan，日志用getOrderLogs，改密用xgmm(oid/pwd)，暂停用pause
	"wanzi": {SuccessCode: "1", ReturnsYID: true, ProgressAct: "chadan", ProgressNoYID: "chadan", LogAct: "getOrderLogs", ChangePassAct: "xgmm", ChangePassIDParam: "oid", ChangePassParam: "pwd", PauseAct: "pause"},
	// lgwk - 查课用/login 端点，自定义查课
	"lgwk": {QueryAct: "lgwk_custom", SuccessCode: "0"},
	// 天河(skyriver) - 成功码"1"，返回yid，进度用/api/chadan1，改密用xgmm(oid/newpass)，分类用getfl，工单用submitWorkOrder/queryWorkOrder
	"skyriver": {SuccessCode: "1", ReturnsYID: true, ProgressPath: "/api/chadan1", ProgressNoYID: "chadan", ChangePassAct: "xgmm", ChangePassParam: "newpass", ChangePassIDParam: "oid", CategoryAct: "getfl", ReportAct: "submitWorkOrder", GetReportAct: "queryWorkOrder", BalanceMoneyField: "money"},
	// 学妹 - 成功码"0"，返回yid，进度用chadan(username)，暂停用ikunStop，改密用gaimi(oid/pwd)，日志用cha_logwk(oid)，售后用shouhou
	"xuemei": {SuccessCode: "0", ReturnsYID: true, ProgressAct: "chadan", ProgressNoYID: "chadan", AlwaysUsername: true, PauseAct: "ikunStop", ChangePassAct: "gaimi", ChangePassIDParam: "oid", ChangePassParam: "pwd", LogAct: "cha_logwk", LogIDParam: "oid", BalanceMoneyField: "money"},
}

// dbConfigCache 数据库平台配置缓存
var (
	dbConfigCache  = map[string]PlatformConfig{}
	dbNameCache    = map[string]string{}
	dbConfigMu     sync.RWMutex
	dbConfigLoaded bool
)

// loadDBPlatformConfigs 从数据库加载平台配置到缓存
func loadDBPlatformConfigs() {
	dbConfigMu.Lock()
	defer dbConfigMu.Unlock()

	rows, err := database.DB.Query(`SELECT pt, name, success_codes,
		query_act, order_act, extra_params, returns_yid,
		progress_act, progress_no_yid, progress_path, progress_method,
		use_id_param, use_uuid_param, always_username, yid_in_data_array,
		pause_act, pause_path, COALESCE(pause_id_param,'id'),
		change_pass_act, change_pass_param, change_pass_id_param,
		change_pass_path, resubmit_path, COALESCE(resubmit_id_param,'id'), log_act, log_path, log_method, log_id_param, use_json,
		COALESCE(balance_act,'getmoney'), COALESCE(balance_path,''), COALESCE(balance_money_field,'money'),
		COALESCE(balance_method,'POST'), COALESCE(balance_auth_type,''),
		COALESCE(report_param_style,''), COALESCE(report_auth_type,''),
		COALESCE(report_path,''), COALESCE(get_report_path,''), COALESCE(refresh_path,'')
		FROM qingka_platform_config`)
	if err != nil {
		fmt.Printf("[platform] 从数据库加载配置失败：%v，使用硬编码配置\n", err)
		dbConfigLoaded = true
		return
	}
	defer rows.Close()

	newConfigs := map[string]PlatformConfig{}
	newNames := map[string]string{}

	for rows.Next() {
		var pt, name, successCodes string
		var cfg PlatformConfig
		err := rows.Scan(
			&pt, &name, &successCodes,
			&cfg.QueryAct, &cfg.OrderAct, &cfg.ExtraParams, &cfg.ReturnsYID,
			&cfg.ProgressAct, &cfg.ProgressNoYID, &cfg.ProgressPath, &cfg.ProgressMethod,
			&cfg.UseIDParam, &cfg.UseUUIDParam, &cfg.AlwaysUsername, &cfg.YIDInDataArray,
			&cfg.PauseAct, &cfg.PausePath, &cfg.PauseIDParam,
			&cfg.ChangePassAct, &cfg.ChangePassParam, &cfg.ChangePassIDParam,
			&cfg.ChangePassPath, &cfg.ResubmitPath, &cfg.ResubmitIDParam, &cfg.LogAct, &cfg.LogPath, &cfg.LogMethod, &cfg.LogIDParam, &cfg.UseJSON,
			&cfg.BalanceAct, &cfg.BalancePath, &cfg.BalanceMoneyField, &cfg.BalanceMethod, &cfg.BalanceAuthType,
			&cfg.ReportParamStyle, &cfg.ReportAuthType,
			&cfg.ReportPath, &cfg.GetReportPath, &cfg.RefreshPath,
		)
		if err != nil {
			fmt.Printf("[platform] 解析配置 %s 失败：%v\n", pt, err)
			continue
		}
		cfg.SuccessCode = successCodes
		newConfigs[pt] = cfg
		newNames[pt] = name
	}

	if len(newConfigs) > 0 {
		dbConfigCache = newConfigs
		dbNameCache = newNames
		fmt.Printf("[platform] 从数据库加载了 %d 个平台配置\n", len(newConfigs))
	}
	dbConfigLoaded = true
}

// ReloadPlatformConfigs 重新加载平台配置（供 handler 调用）
func ReloadPlatformConfigs() {
	dbConfigMu.Lock()
	dbConfigLoaded = false
	dbConfigMu.Unlock()
	loadDBPlatformConfigs()
}

// GetPlatformConfig 获取平台配置：优先从 DB 缓存读，fallback 到硬编码 registry
func GetPlatformConfig(pt string) PlatformConfig {
	// 确保 DB 配置已加载
	dbConfigMu.RLock()
	loaded := dbConfigLoaded
	dbConfigMu.RUnlock()
	if !loaded {
		loadDBPlatformConfigs()
	}

	// 优先从 DB 缓存读取
	dbConfigMu.RLock()
	if cfg, ok := dbConfigCache[pt]; ok {
		dbConfigMu.RUnlock()
		return fillDefaults(cfg)
	}
	dbConfigMu.RUnlock()

	// fallback 到硬编码 registry
	if cfg, ok := platformRegistry[pt]; ok {
		return fillDefaults(cfg)
	}

	// 默认配置：标准 API 接口
	return PlatformConfig{
		QueryAct:       "get",
		OrderAct:       "add",
		ProgressAct:    "chadan2",
		ProgressNoYID:  "chadan",
		ProgressMethod: "POST",
		SuccessCode:    "0",
	}
}

// fillDefaults 填充平台配置的默认值
func fillDefaults(cfg PlatformConfig) PlatformConfig {
	if cfg.QueryAct == "" {
		cfg.QueryAct = "get"
	}
	if cfg.OrderAct == "" {
		cfg.OrderAct = "add"
	}
	if cfg.ProgressAct == "" {
		cfg.ProgressAct = "chadan2"
	}
	if cfg.ProgressNoYID == "" {
		cfg.ProgressNoYID = "chadan"
	}
	if cfg.ProgressMethod == "" {
		cfg.ProgressMethod = "POST"
	}
	if cfg.SuccessCode == "" {
		cfg.SuccessCode = "0"
	}
	if cfg.BalanceAct == "" {
		cfg.BalanceAct = "getmoney"
	}
	if cfg.BalanceMoneyField == "" {
		cfg.BalanceMoneyField = "money"
	}
	if cfg.BalanceMethod == "" {
		cfg.BalanceMethod = "POST"
	}
	if cfg.ReportAct == "" {
		cfg.ReportAct = "report"
	}
	if cfg.GetReportAct == "" {
		cfg.GetReportAct = "getReport"
	}
	if cfg.ReportSuccessCode == "" {
		cfg.ReportSuccessCode = "1"
	}
	if cfg.ResubmitIDParam == "" {
		cfg.ResubmitIDParam = "id"
	}
	if cfg.ReportParamStyle == "" {
		cfg.ReportParamStyle = "standard"
	}
	if cfg.CategoryAct == "" {
		cfg.CategoryAct = "getcate"
	}
	return cfg
}

// GetPlatformNames 返回所有已注册平台名称映射（DB 优先，fallback 硬编码）
func GetPlatformNames() map[string]string {
	// 确保 DB 配置已加载
	dbConfigMu.RLock()
	loaded := dbConfigLoaded
	dbConfigMu.RUnlock()
	if !loaded {
		loadDBPlatformConfigs()
	}

	// 合并：硬编码兜底 + DB 覆盖
	result := map[string]string{
		"27":       "春秋",
		"haha":     "乐学",
		"hzw":      "hzw",
		"zy":       "志塬查课",
		"longlong": "龙龙平台",
		"liunian":  "流年",
		"xxtgf":    "学习通普通",
		"moocmd":   "毛豆 mooc",
		"yyy":      "yyy 平台",
		"2xx":      "爱学习",
		"KUN":      "KUN",
		"kunba":    "kunba",
		"tuboshu":  "土拨鼠",
		"29":       "29",
		"spi":      "spiderman",
		"lg":       "lg 学习平台",
		"nx":       "奶昔",
		"pup":      "pup",
		"wanzi":    "丸子",
		"lgwk":     "lgwk",
		"Benz":     "奔驰",
		"skyriver": "天河",
		"xuemei":   "学妹",
	}

	dbConfigMu.RLock()
	for pt, name := range dbNameCache {
		result[pt] = name
	}
	dbConfigMu.RUnlock()

	return result
}

// GetSupplierByHID 根据 hid 获取供应商完整信息
func (s *SupplierService) GetSupplierByHID(hid int) (*model.SupplierFull, error) {
	var sup model.SupplierFull
	err := database.DB.QueryRow(
		"SELECT hid, COALESCE(pt,''), COALESCE(name,''), COALESCE(url,''), COALESCE(user,''), COALESCE(pass,''), COALESCE(token,''), COALESCE(ip,''), COALESCE(cookie,''), COALESCE(money,'0'), COALESCE(status,'1') FROM qingka_wangke_huoyuan WHERE hid = ?",
		hid,
	).Scan(&sup.HID, &sup.PT, &sup.Name, &sup.URL, &sup.User, &sup.Pass, &sup.Token, &sup.IP, &sup.Cookie, &sup.Money, &sup.Status)
	if err != nil {
		return nil, fmt.Errorf("供应商不存在：%v", err)
	}
	return &sup, nil
}

// GetClassFull 获取课程完整信息
func (s *SupplierService) GetClassFull(cid int) (*model.ClassFull, error) {
	var cls model.ClassFull
	err := database.DB.QueryRow(
		"SELECT cid, COALESCE(name,''), COALESCE(noun,''), COALESCE(price,'0'), COALESCE(docking,'0'), COALESCE(fenlei,''), COALESCE(status,0), COALESCE(yunsuan,'*'), COALESCE(content,'') FROM qingka_wangke_class WHERE cid = ?",
		cid,
	).Scan(&cls.CID, &cls.Name, &cls.Noun, &cls.Price, &cls.Docking, &cls.Fenlei, &cls.Status, &cls.Yunsuan, &cls.Content)
	if err != nil {
		return nil, fmt.Errorf("课程不存在：%v", err)
	}
	return &cls, nil
}

// QueryCourse 查课：代理到上游供应商
func (s *SupplierService) QueryCourse(cid int, userinfo string) (*model.CourseQueryResponse, error) {
	cls, err := s.GetClassFull(cid)
	if err != nil {
		return nil, err
	}
	if cls.Status != 1 {
		return nil, errors.New("课程已下架")
	}

	// 解析 userinfo: "学校 账号 密码"
	parts := strings.Fields(userinfo)
	var school, user, pass string
	if len(parts) >= 3 {
		school = parts[0]
		user = parts[1]
		pass = parts[2]
	} else if len(parts) == 2 {
		school = "自动识别"
		user = parts[0]
		pass = parts[1]
	} else {
		return nil, errors.New("下单信息格式错误，请输入：学校 账号 密码")
	}

	// 如果没有对接供应商（docking=0），返回空结果提示手动处理
	dockingID, _ := strconv.Atoi(cls.Docking)
	if dockingID == 0 {
		return &model.CourseQueryResponse{
			UserInfo: userinfo,
			UserName: user,
			Msg:      "此课程无需查课，直接下单即可",
			Data:     []model.CourseItem{},
		}, nil
	}

	// 获取供应商信息
	sup, err := s.GetSupplierByHID(dockingID)
	if err != nil {
		return nil, err
	}

	// 根据平台配置决定查课方式
	cfg := GetPlatformConfig(sup.PT)

	switch cfg.QueryAct {
	case "local_time":
		// 按 PHP 27/zy: 本地生成时长选择列表
		data := s.generateLocalTimeList(cls.Noun)
		return &model.CourseQueryResponse{
			UserInfo: userinfo,
			UserName: user,
			Msg:      "查询成功",
			Data:     data,
		}, nil

	case "local_script":
		// moocmd 等本地脚本平台，新系统暂不支持
		return &model.CourseQueryResponse{
			UserInfo: userinfo,
			UserName: user,
			Msg:      fmt.Sprintf("平台 %s 暂不支持自动查课，请直接下单", sup.PT),
			Data:     []model.CourseItem{},
		}, nil

	case "xxt_query":
		// 学习通直接查课（登录超星API获取课程列表）
		result, err := xxtCallQuery(user, pass, school)
		if err != nil {
			return &model.CourseQueryResponse{
				UserInfo: userinfo,
				UserName: user,
				Msg:      fmt.Sprintf("学习通查课失败：%s", err.Error()),
				Data:     []model.CourseItem{},
			}, nil
		}
		codeVal, _ := result["code"].(int)
		if codeVal == -1 {
			msg, _ := result["msg"].(string)
			return &model.CourseQueryResponse{
				UserInfo: userinfo,
				UserName: user,
				Msg:      msg,
				Data:     []model.CourseItem{},
			}, nil
		}
		// 转换课程数据
		var items []model.CourseItem
		if data, ok := result["data"].([]map[string]interface{}); ok {
			for _, d := range data {
				items = append(items, model.CourseItem{
					ID:   toString(d["id"]),
					Name: toString(d["name"]),
				})
			}
		}
		if userInfoStr, ok := result["userinfo"].(string); ok {
			userinfo = userInfoStr
		}
		return &model.CourseQueryResponse{
			UserInfo: userinfo,
			UserName: user,
			Msg:      "查询成功",
			Data:     items,
		}, nil

	case "yyy_custom":
		// yyy 平台：部分商品无需查课，响应成功即可直接下单
		return &model.CourseQueryResponse{
			UserInfo: userinfo,
			UserName: user,
			Msg:      "查询成功",
			Data:     []model.CourseItem{},
		}, nil

	case "KUN_custom":
		// KUN/kunba 平台：GET /query/?platform=&school=&account=&password=&dtoken=
		result, err := kunCallQuery(sup, cls.Noun, school, user, pass)
		if err != nil {
			return &model.CourseQueryResponse{
				UserInfo: userinfo,
				UserName: user,
				Msg:      fmt.Sprintf("查课失败：%s", err.Error()),
				Data:     []model.CourseItem{},
			}, nil
		}
		return &model.CourseQueryResponse{
			UserInfo: userinfo,
			UserName: user,
			Msg:      result.Msg,
			Data:     result.Data,
		}, nil

	default:
		// 标准 API 查课
		result, err := s.callSupplierQuery(sup, cls, school, user, pass)
		if err != nil {
			return &model.CourseQueryResponse{
				UserInfo: userinfo,
				UserName: user,
				Msg:      fmt.Sprintf("查课失败：%s", err.Error()),
				Data:     []model.CourseItem{},
			}, nil
		}
		return &model.CourseQueryResponse{
			UserInfo: userinfo,
			UserName: result.UserName,
			Msg:      result.Msg,
			Data:     result.Data,
		}, nil
	}
}

// generateLocalTimeList 按 PHP 27/zy 平台：本地生成时长选择列表
func (s *SupplierService) generateLocalTimeList(noun string) []model.CourseItem {
	var hoursPerUnit int
	switch noun {
	case "1":
		hoursPerUnit = 5
	default:
		hoursPerUnit = 6
	}
	items := make([]model.CourseItem, 0, 20)
	for i := 1; i <= 20; i++ {
		total := i * hoursPerUnit
		items = append(items, model.CourseItem{
			ID:   fmt.Sprintf("%d", i),
			Name: fmt.Sprintf("从第一个开始选择，每选中一个加%d小时，选到此处总时长为 %d 小时", hoursPerUnit, total),
		})
	}
	return items
}

// callSupplierQuery 调用上游供应商查课 API
func (s *SupplierService) callSupplierQuery(sup *model.SupplierFull, cls *model.ClassFull, school, user, pass string) (*model.SupplierQueryResult, error) {
	apiURL := s.buildSupplierURL(sup.URL, "get")

	formData := url.Values{}
	formData.Set("uid", sup.User)
	formData.Set("key", sup.Pass)
	formData.Set("school", school)
	formData.Set("user", user)
	formData.Set("pass", pass)
	formData.Set("platform", cls.Noun)

	resp, err := s.client.PostForm(apiURL, formData)
	if err != nil {
		return nil, fmt.Errorf("请求上游失败：%v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败：%v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("解析响应失败：%s", string(body))
	}

	msg, _ := raw["msg"].(string)
	userName, _ := raw["userName"].(string)

	var items []model.CourseItem
	if dataArr, ok := raw["data"].([]interface{}); ok {
		for _, item := range dataArr {
			if m, ok := item.(map[string]interface{}); ok {
				items = append(items, model.CourseItem{
					ID:             toString(m["id"]),
					Name:           toString(m["name"]),
					KCJS:           toString(m["kcjs"]),
					StudyStartTime: toString(m["studyStartTime"]),
					StudyEndTime:   toString(m["studyEndTime"]),
					ExamStartTime:  toString(m["examStartTime"]),
					ExamEndTime:    toString(m["examEndTime"]),
					Complete:       toString(m["complete"]),
				})
			}
		}
	}

	return &model.SupplierQueryResult{
		Msg:      msg,
		UserName: userName,
		Data:     items,
	}, nil
}

// CallSupplierOrder 调用上游供应商下单 API（按平台配置处理差异）
func (s *SupplierService) CallSupplierOrder(sup *model.SupplierFull, cls *model.ClassFull, school, user, pass, kcid, kcname string, extraFields map[string]string) (*model.SupplierOrderResult, error) {
	// yyy 平台走独立适配器
	if sup.PT == "yyy" {
		return yyyCallOrder(sup, user, pass, kcname, cls.Noun)
	}

	// KUN/kunba 平台走独立适配器
	if sup.PT == "KUN" || sup.PT == "kunba" {
		return kunCallOrder(sup, cls.Noun, school, user, pass, kcname, kcid)
	}

	cfg := GetPlatformConfig(sup.PT)
	apiURL := s.buildSupplierURL(sup.URL, cfg.OrderAct)

	formData := url.Values{}
	formData.Set("uid", sup.User)
	formData.Set("key", sup.Pass)
	formData.Set("platform", cls.Noun)
	formData.Set("school", school)
	formData.Set("user", user)
	formData.Set("pass", pass)
	formData.Set("kcid", kcid)
	formData.Set("kcname", kcname)

	// 按平台配置传递额外参数 (score/shichang 等)
	if cfg.ExtraParams && extraFields != nil {
		for k, v := range extraFields {
			if v != "" {
				formData.Set(k, v)
			}
		}
	}

	// 对同一上游主机限速：每个 host 至少间隔 500ms，防止打爆上游
	if host := extractHost(sup.URL); host != "" {
		globalRateLimiter.wait(host, 500*time.Millisecond)
	}

	resp, err := s.client.PostForm(apiURL, formData)
	if err != nil {
		return nil, fmt.Errorf("请求上游失败：%v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败：%v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("解析响应失败：%s", string(body))
	}

	// 按平台配置判断成功码
	codeVal := fmt.Sprintf("%v", raw["code"])
	result := &model.SupplierOrderResult{
		Msg: fmt.Sprintf("%v", raw["msg"]),
	}

	// 调试日志：记录上游响应
	fmt.Printf("[CallSupplierOrder] pt=%s codeVal=%s cfg.SuccessCode=%s raw=%v\n", sup.PT, codeVal, cfg.SuccessCode, raw)

	if codeVal == cfg.SuccessCode {
		result.Code = 1 // 统一内部成功码为 1
		result.Msg = fmt.Sprintf("%v", raw["msg"])
		if result.Msg == "" {
			result.Msg = "下单成功"
		}
		// 提取 yid（如果平台支持）
		if cfg.ReturnsYID {
			// 优先从根级 id 字段提取（支持 string/float64/interface 多种类型）
			if idVal, ok := raw["id"]; ok && idVal != nil {
				switch v := idVal.(type) {
				case string:
					result.YID = v
				case float64:
					result.YID = fmt.Sprintf("%.0f", v)
				case int:
					result.YID = fmt.Sprintf("%d", v)
				default:
					result.YID = fmt.Sprintf("%v", v)
				}
			}
			// 如果根级 id 为空，尝试从 ids 数组提取（如 wanzi 平台）
			if result.YID == "" {
				if idsArr, ok := raw["ids"].([]interface{}); ok && len(idsArr) > 0 {
					switch v := idsArr[0].(type) {
					case string:
						result.YID = v
					case float64:
						result.YID = fmt.Sprintf("%.0f", v)
					default:
						result.YID = fmt.Sprintf("%v", v)
					}
				}
			}
			// 如果仍为空，尝试从 data 数组提取（兼容 wanzi 等平台 data 数组直接返回 yid 列表）
			if result.YID == "" {
				if dataArr, ok := raw["data"].([]interface{}); ok && len(dataArr) > 0 {
					// 检查是否是字符串数组（如 wanzi 平台）
					if str, ok := dataArr[0].(string); ok {
						result.YID = str
					} else {
						switch v := dataArr[0].(type) {
						case string:
							result.YID = v
						case float64:
							result.YID = fmt.Sprintf("%.0f", v)
						default:
							result.YID = fmt.Sprintf("%v", v)
						}
					}
				}
			}
			// 如果仍为空，尝试从 data.id 提取
			if result.YID == "" {
				if dataMap, ok := raw["data"].(map[string]interface{}); ok {
					if id, ok := dataMap["id"]; ok && id != nil {
						switch v := id.(type) {
						case string:
							result.YID = v
						case float64:
							result.YID = fmt.Sprintf("%.0f", v)
						default:
							result.YID = fmt.Sprintf("%v", v)
						}
					}
				}
			}
			// 如果仍为空，尝试从 data 数组提取（如龙龙平台）
			if result.YID == "" && cfg.YIDInDataArray {
				if dataArr, ok := raw["data"].([]interface{}); ok && len(dataArr) > 0 {
					switch v := dataArr[0].(type) {
					case string:
						result.YID = v
					case float64:
						result.YID = fmt.Sprintf("%.0f", v)
					default:
						result.YID = fmt.Sprintf("%v", v)
					}
				}
			}
			// 调试日志：记录提取的 yid
			fmt.Printf("[CallSupplierOrder] pt=%s extracted YID=%s\n", sup.PT, result.YID)
		}
	} else {
		result.Code = -1
		result.Msg = fmt.Sprintf("%v", raw["msg"])
		if result.Msg == "" {
			result.Msg = fmt.Sprintf("上游返回错误码：%s", codeVal)
		}
		fmt.Printf("[CallSupplierOrder] pt=%s 上游返回错误码 codeVal=%s expected=%s msg=%s\n", sup.PT, codeVal, cfg.SuccessCode, result.Msg)
	}

	return result, nil
}

// SupplierClassItem 上游供应商课程列表项 (按 PHP yjdj case 的 getclass 返回)
type SupplierClassItem struct {
	CID          string  `json:"cid"`
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	Fenlei       string  `json:"fenlei"`
	Content      string  `json:"content"`
	CategoryName string  `json:"category_name"`
}

// GetSupplierCategories 从上游获取分类名称映射 (按 PHP: api.php?act=getcate)
func (s *SupplierService) GetSupplierCategories(sup *model.SupplierFull) map[string]string {
	cfg := GetPlatformConfig(sup.PT)

	baseURL := strings.TrimRight(sup.URL, "/")
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "http://" + baseURL
	}

	client := &http.Client{Timeout: 8 * time.Second}

	// 确定要尝试的分类 act 列表
	tryActs := []string{}
	if cfg.CategoryAct != "" {
		tryActs = append(tryActs, cfg.CategoryAct)
	}
	// 常见分类 act 作为 fallback
	for _, fallback := range []string{"getfl", "getcate", "getfenlei"} {
		found := false
		for _, a := range tryActs {
			if a == fallback {
				found = true
				break
			}
		}
		if !found {
			tryActs = append(tryActs, fallback)
		}
	}

	for _, act := range tryActs {
		var resp *http.Response
		var err error

		if cfg.ReportAuthType == "token_only" && cfg.UseJSON {
			apiURL := baseURL + "/api/" + act
			jsonData, _ := json.Marshal(map[string]string{"token": sup.Pass})
			req, _ := http.NewRequest("POST", apiURL, strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			resp, err = client.Do(req)
		} else {
			cateURL := s.buildSupplierURL(sup.URL, act)
			formData := url.Values{}
			formData.Set("uid", sup.User)
			formData.Set("key", sup.Pass)
			resp, err = client.PostForm(cateURL, formData)
		}
		if err != nil {
			log.Printf("[GetSupplierCategories] act=%s 请求失败: %v", act, err)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		m := parseCategoryResponse(body)
		if len(m) > 0 {
			log.Printf("[GetSupplierCategories] pt=%s act=%s 成功获取 %d 个分类", sup.PT, act, len(m))
			return m
		}
	}

	log.Printf("[GetSupplierCategories] pt=%s 所有分类 act 均未获取到数据", sup.PT)
	return nil
}

// parseCategoryResponse 解析分类响应，兼容多种字段名格式
func parseCategoryResponse(body []byte) map[string]string {
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil
	}

	// 提取 data 数组（兼容 data / list / fenlei 字段）
	var dataArr []interface{}
	for _, key := range []string{"data", "list", "fenlei", "category", "categories"} {
		if v, ok := raw[key]; ok && v != nil {
			dataBytes, _ := json.Marshal(v)
			if err := json.Unmarshal(dataBytes, &dataArr); err == nil && len(dataArr) > 0 {
				break
			}
		}
	}

	if len(dataArr) == 0 {
		return nil
	}

	m := map[string]string{}
	for _, item := range dataArr {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		// 提取 ID：兼容 id / fid / fenlei_id / category_id
		fid := ""
		for _, idKey := range []string{"id", "fid", "fenlei_id", "category_id", "cate_id", "typeId", "type_id"} {
			if v, ok := itemMap[idKey]; ok && v != nil {
				fid = fmt.Sprintf("%v", v)
				break
			}
		}
		// 提取名称：兼容 name / fname / fenleiName / fenlei_name / category_name / catname / typeName / type_name / title
		fname := ""
		for _, nameKey := range []string{"name", "fname", "fenleiName", "fenlei_name", "category_name", "catname", "typeName", "type_name", "title", "label"} {
			if v, ok := itemMap[nameKey]; ok && v != nil {
				s := fmt.Sprintf("%v", v)
				if s != "" && s != "<nil>" {
					fname = s
					break
				}
			}
		}
		if fid != "" && fid != "<nil>" && fid != "0" && fname != "" {
			m[fid] = fname
		}
	}
	return m
}

// GetSupplierClasses 从上游供应商获取课程列表 (按 PHP: api.php?act=getclass)
func (s *SupplierService) GetSupplierClasses(sup *model.SupplierFull) ([]SupplierClassItem, error) {
	// yyy 平台走独立适配器
	if sup.PT == "yyy" {
		return yyyGetClasses(sup)
	}

	cfg := GetPlatformConfig(sup.PT)

	var resp *http.Response
	var err error

	if cfg.ReportAuthType == "token_only" && cfg.UseJSON {
		// 2xx 等 token_only 平台：JSON POST 到 /api/getclass
		baseURL := strings.TrimRight(sup.URL, "/")
		if !strings.HasPrefix(baseURL, "http") {
			baseURL = "http://" + baseURL
		}
		apiURL := baseURL + "/api/getclass"
		jsonData, _ := json.Marshal(map[string]string{"token": sup.Pass})
		req, _ := http.NewRequest("POST", apiURL, strings.NewReader(string(jsonData)))
		req.Header.Set("Content-Type", "application/json")
		resp, err = s.client.Do(req)
	} else {
		apiURL := s.buildSupplierURL(sup.URL, "getclass")
		formData := url.Values{}
		formData.Set("uid", sup.User)
		formData.Set("key", sup.Pass)
		resp, err = s.client.PostForm(apiURL, formData)
	}
	if err != nil {
		return nil, fmt.Errorf("请求上游失败：%v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败：%v", err)
	}

	// 使用 map 解析，兼容 code 为字符串或整数两种格式
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("解析响应失败：%s", string(body))
	}

	// 提取 code 值（兼容字符串和整数）
	codeVal := ""
	if codeRaw, ok := raw["code"]; ok {
		switch v := codeRaw.(type) {
		case string:
			codeVal = v
		case float64:
			codeVal = fmt.Sprintf("%.0f", v)
		case int:
			codeVal = fmt.Sprintf("%d", v)
		default:
			codeVal = fmt.Sprintf("%v", v)
		}
	}

	// 判断是否成功（code=0 或 code=1 都表示成功）
	if codeVal != "0" && codeVal != "1" {
		msg := ""
		if msgRaw, ok := raw["msg"]; ok {
			msg = fmt.Sprintf("%v", msgRaw)
		}
		if msg == "" {
			msg = "查询上游进度失败"
		}
		return nil, fmt.Errorf("%s", msg)
	}

	// 提取 data 数组并转换为 SupplierClassItem
	var rawItems []map[string]interface{}
	if dataRaw, ok := raw["data"]; ok && dataRaw != nil {
		dataBytes, _ := json.Marshal(dataRaw)
		json.Unmarshal(dataBytes, &rawItems)
	}

	items := make([]SupplierClassItem, 0, len(rawItems))
	for _, d := range rawItems {
		var price float64
		if priceVal, ok := d["price"]; ok {
			switch v := priceVal.(type) {
			case float64:
				price = v
			case string:
				price, _ = strconv.ParseFloat(v, 64)
			}
		}
		catName := ""
		for _, ck := range []string{"category_name", "fenleiName", "fenlei_name", "catname", "typeName", "type_name", "catName"} {
			if cn, ok := d[ck]; ok {
				s := fmt.Sprintf("%v", cn)
				if s != "" && s != "<nil>" {
					catName = s
					break
				}
			}
		}
		items = append(items, SupplierClassItem{
			CID:          fmt.Sprintf("%v", d["cid"]),
			Name:         fmt.Sprintf("%v", d["name"]),
			Price:        price,
			Fenlei:       fmt.Sprintf("%v", d["fenlei"]),
			Content:      fmt.Sprintf("%v", d["content"]),
			CategoryName: catName,
		})
	}

	// 获取上游分类名称映射，注入 category_name (按 PHP getclassdata)
	cateMap := s.GetSupplierCategories(sup)
	if len(cateMap) > 0 {
		for i := range items {
			fid := items[i].Fenlei
			if name, ok := cateMap[fid]; ok {
				items[i].CategoryName = name
			}
		}
	}

	return items, nil
}

// ImportSupplierClasses 一键对接：从上游批量导入课程 (按 PHP yjdj case)
func (s *SupplierService) ImportSupplierClasses(hid int, pricee float64, category string, name string, fd int) (int, int, string, error) {
	sup, err := s.GetSupplierByHID(hid)
	if err != nil {
		return 0, 0, "", err
	}

	classList, err := s.GetSupplierClasses(sup)
	if err != nil {
		return 0, 0, "", err
	}

	if len(classList) == 0 {
		return 0, 0, "", fmt.Errorf("接口返回的数据为空")
	}

	// 确定分类 ID (按 PHP yjdj case)
	var fenleiID int
	if category != "999999" {
		// 查找或创建分类
		catName := name
		if catName == "" {
			catName = sup.Name
		}
		err := database.DB.QueryRow(
			"SELECT id FROM qingka_wangke_fenlei WHERE name = ? AND status != 3 ORDER BY id DESC LIMIT 1", catName,
		).Scan(&fenleiID)
		if err != nil && fd == 0 {
			// 创建新分类 (按 PHP: INSERT INTO qingka_wangke_fenlei)
			result, err2 := database.DB.Exec(
				"INSERT INTO qingka_wangke_fenlei (sort, name, status, time) VALUES (10, ?, '1', NOW())", catName,
			)
			if err2 == nil {
				id, _ := result.LastInsertId()
				fenleiID = int(id)
			}
		}
	}

	inserted, updated := 0, 0

	for _, item := range classList {
		// 分类过滤 (按 PHP: if ($value['fenlei'] == $category))
		if category != "999999" && item.Fenlei != category {
			continue
		}

		price := item.Price * pricee

		// 检查是否已存在 (按 PHP: SELECT COUNT(*) FROM qingka_wangke_class WHERE docking = '{$hid}' AND noun = '{$value['cid']}')
		var existCount int
		database.DB.QueryRow(
			"SELECT COUNT(*) FROM qingka_wangke_class WHERE docking = ? AND noun = ?", hid, item.CID,
		).Scan(&existCount)

		if existCount > 0 {
			// 更新现有记录 (按 PHP: UPDATE qingka_wangke_class SET price=..., content=..., status='1')
			database.DB.Exec(
				"UPDATE qingka_wangke_class SET price = ?, content = ?, status = 1 WHERE docking = ? AND noun = ?",
				price, item.Content, hid, item.CID,
			)
			updated++
		} else if fd == 0 {
			// 全量模式且 category=999999 时需要按 category_name 自动建分类
			thisFenlei := fenleiID
			if category == "999999" && item.CategoryName != "" {
				var catID int
				err := database.DB.QueryRow(
					"SELECT id FROM qingka_wangke_fenlei WHERE name = ? AND status != 3 ORDER BY id DESC LIMIT 1",
					item.CategoryName,
				).Scan(&catID)
				if err != nil {
					r, e := database.DB.Exec(
						"INSERT INTO qingka_wangke_fenlei (sort, name, status, time) VALUES (10, ?, '1', NOW())", item.CategoryName,
					)
					if e == nil {
						id, _ := r.LastInsertId()
						catID = int(id)
					}
				}
				thisFenlei = catID
			}

			// 获取最大 sort (按 PHP: SELECT MAX(sort) as max_sort FROM qingka_wangke_class)
			var maxSort int
			database.DB.QueryRow("SELECT COALESCE(MAX(sort),10) FROM qingka_wangke_class").Scan(&maxSort)

			// 插入新记录 (按 PHP: INSERT INTO qingka_wangke_class)
			database.DB.Exec(
				"INSERT INTO qingka_wangke_class (name, getnoun, noun, fenlei, queryplat, docking, price, sort, content, addtime) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())",
				item.Name, item.CID, item.CID, thisFenlei, hid, hid, price, maxSort+1, item.Content,
			)
			inserted++
		}
	}

	total := inserted + updated
	msg := fmt.Sprintf("共计%d个，新上架%d个，更新%d个", total, inserted, updated)
	return inserted, updated, msg, nil
}

// SyncSupplierStatus 从供应商同步课程状态，自动下架已删除课程 (按 PHP updateStatus case)
func (s *SupplierService) SyncSupplierStatus(hid int) (int, string, error) {
	sup, err := s.GetSupplierByHID(hid)
	if err != nil {
		return 0, "", err
	}

	classList, err := s.GetSupplierClasses(sup)
	if err != nil {
		return 0, "", err
	}

	// 收集上游 CID 列表
	apiCIDs := map[string]bool{}
	for _, item := range classList {
		apiCIDs[item.CID] = true
	}

	// 查询本地该供应商的所有课程
	rows, err := database.DB.Query(
		"SELECT cid, noun, status FROM qingka_wangke_class WHERE docking = ?", hid,
	)
	if err != nil {
		return 0, "", err
	}
	defer rows.Close()

	var downIDs []int
	for rows.Next() {
		var cid int
		var noun string
		var status int
		rows.Scan(&cid, &noun, &status)
		if !apiCIDs[noun] && status != 0 {
			downIDs = append(downIDs, cid)
		}
	}

	if len(downIDs) == 0 {
		return 0, "没有需要更新的商品状态", nil
	}

	// 批量下架 (按 PHP: UPDATE qingka_wangke_class SET status = '0' WHERE cid IN (...))
	for _, cid := range downIDs {
		database.DB.Exec("UPDATE qingka_wangke_class SET status = 0 WHERE cid = ?", cid)
	}

	msg := fmt.Sprintf("共下架%d个商品", len(downIDs))
	return len(downIDs), msg, nil
}

// QueryOrderProgress 查询上游订单进度（按平台配置处理差异）
// 按 PHP jdjk.php processCx: 不同平台用不同 URL/方法/参数
func (s *SupplierService) QueryOrderProgress(sup *model.SupplierFull, yid string, username string, orderExtra map[string]string) ([]model.SupplierProgressItem, error) {
	// yyy 平台走独立适配器
	if sup.PT == "yyy" {
		return yyyQueryProgress(sup, username)
	}

	cfg := GetPlatformConfig(sup.PT)

	params := url.Values{}
	params.Set("uid", sup.User)
	params.Set("key", sup.Pass)

	var apiURL string

	if cfg.ProgressPath != "" {
		// 非标准路径：/api/search, /api/chadan1 等
		baseURL := strings.TrimRight(sup.URL, "/")
		if !strings.HasPrefix(baseURL, "http") {
			baseURL = "http://" + baseURL
		}
		apiURL = baseURL + cfg.ProgressPath

		// /api/search 类型需要额外参数 (按 PHP haha/aw 平台)
		if strings.Contains(cfg.ProgressPath, "/api/search") {
			params.Set("username", username)
			if orderExtra != nil {
				if v, ok := orderExtra["kcname"]; ok {
					params.Set("kcname", v)
				}
				if v, ok := orderExtra["noun"]; ok {
					params.Set("cid", v)
				}
			}
		} else {
			// /api/chadan1 类型 (按 PHP 流年)
			params.Set("username", username)
			if yid != "" && yid != "0" {
				params.Set("yid", yid)
			}
			if orderExtra != nil {
				if v, ok := orderExtra["kcname"]; ok {
					params.Set("kcname", v)
				}
				if v, ok := orderExtra["noun"]; ok {
					params.Set("cid", v)
				}
			}
		}
	} else if yid != "" && yid != "0" {
		// 标准路径：有 yid 用 chadan2/chadanoid
		apiURL = s.buildSupplierURL(sup.URL, cfg.ProgressAct)
		if cfg.UseUUIDParam {
			params.Set("uuid", yid)
		} else if cfg.UseIDParam {
			params.Set("id", yid)
		} else {
			params.Set("yid", yid)
		}
		if cfg.AlwaysUsername {
			params.Set("username", username)
		}
	} else {
		// 标准路径：无 yid 用 chadan
		apiURL = s.buildSupplierURL(sup.URL, cfg.ProgressNoYID)
		params.Set("username", username)
	}

	// 对同一上游主机限速：每个 host 至少间隔 500ms，防止打爆上游
	if host := extractHost(sup.URL); host != "" {
		globalRateLimiter.wait(host, 500*time.Millisecond)
	}

	var resp *http.Response
	var err error

	if cfg.ProgressMethod == "GET" {
		apiURL = apiURL + "?" + params.Encode()
		resp, err = s.client.Get(apiURL)
	} else {
		resp, err = s.client.PostForm(apiURL, params)
	}
	if err != nil {
		return nil, fmt.Errorf("请求上游失败：%v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败：%v", err)
	}

	// 使用 map 解析，兼容 code 为字符串或整数两种格式
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("解析响应失败：%s", string(body))
	}

	// 提取 code 值（兼容字符串和整数）
	codeVal := ""
	if codeRaw, ok := raw["code"]; ok {
		switch v := codeRaw.(type) {
		case string:
			codeVal = v
		case float64:
			codeVal = fmt.Sprintf("%.0f", v)
		case int:
			codeVal = fmt.Sprintf("%d", v)
		default:
			codeVal = fmt.Sprintf("%v", v)
		}
	}

	// 判断是否成功（code=0 或 code=1 都表示成功）
	if codeVal != "0" && codeVal != "1" {
		msg := ""
		if msgRaw, ok := raw["msg"]; ok {
			msg = fmt.Sprintf("%v", msgRaw)
		}
		if msg == "" {
			msg = "查询上游进度失败"
		}
		return nil, fmt.Errorf("%s", msg)
	}

	// 提取 data 数组（手动转换，兼容上游返回 string/number 混合类型）
	var items []model.SupplierProgressItem
	if dataArr, ok := raw["data"].([]interface{}); ok {
		for _, item := range dataArr {
			if m, ok := item.(map[string]interface{}); ok {
				items = append(items, model.SupplierProgressItem{
					YID:             toString(m["id"]),
					KCName:          toString(m["kcname"]),
					User:            toString(m["user"]),
					Status:          toString(m["status"]),
					StatusText:      toString(m["status_text"]),
					Process:         toString(m["process"]),
					Remarks:         toString(m["remarks"]),
					CourseStartTime: toString(m["courseStartTime"]),
					CourseEndTime:   toString(m["courseEndTime"]),
					ExamStartTime:   toString(m["examStartTime"]),
					ExamEndTime:     toString(m["examEndTime"]),
				})
			}
		}
	}

	return items, nil
}

// buildSupplierURL 构建供应商 API URL
func (s *SupplierService) buildSupplierURL(baseURL, act string) string {
	baseURL = strings.TrimRight(baseURL, "/")
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "http://" + baseURL
	}
	return fmt.Sprintf("%s/api.php?act=%s", baseURL, act)
}

// PauseOrder 暂停/恢复上游订单（按 PHP js.php stopOrder / ztjk.php stopWk）
func (s *SupplierService) PauseOrder(sup *model.SupplierFull, yid string) (int, string, error) {
	if sup.PT == "yyy" || sup.PT == "tuboshu" {
		return -1, "当前平台暂不支持暂停操作", nil
	}

	// KUN/kunba 平台走独立适配器
	if sup.PT == "KUN" || sup.PT == "kunba" {
		return kunPauseOrder(sup, yid)
	}

	cfg := GetPlatformConfig(sup.PT)

	if host := extractHost(sup.URL); host != "" {
		globalRateLimiter.wait(host, 500*time.Millisecond)
	}

	var resp *http.Response
	var err error

	if cfg.PausePath != "" {
		// 2xx 等非标准路径：JSON POST 到自定义路径
		baseURL := strings.TrimRight(sup.URL, "/")
		if !strings.HasPrefix(baseURL, "http") {
			baseURL = "http://" + baseURL
		}
		apiURL := baseURL + cfg.PausePath
		jsonData, _ := json.Marshal(map[string]string{"id": yid, "username": sup.User})
		resp, err = s.client.Post(apiURL, "application/json", strings.NewReader(string(jsonData)))
	} else {
		pauseAct := cfg.PauseAct
		if pauseAct == "" {
			pauseAct = "zt"
		}
		apiURL := s.buildSupplierURL(sup.URL, pauseAct)
		formData := url.Values{}
		formData.Set("uid", sup.User)
		formData.Set("key", sup.Pass)
		idParam := cfg.PauseIDParam
		if idParam == "" {
			idParam = "id"
		}
		formData.Set(idParam, yid)
		resp, err = s.client.PostForm(apiURL, formData)
	}

	if err != nil {
		return -1, "", fmt.Errorf("请求上游失败：%v", err)
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

// ChangePassword 修改上游订单密码（按 PHP js.php gaimi / xgjk.php updateWk）
func (s *SupplierService) ChangePassword(sup *model.SupplierFull, yid, newPwd string) (int, string, error) {
	if sup.PT == "yyy" || sup.PT == "tuboshu" {
		return -1, "当前平台暂不支持改密操作", nil
	}

	// KUN/kunba 平台走独立适配器
	if sup.PT == "KUN" || sup.PT == "kunba" {
		return kunChangePassword(sup, yid, newPwd)
	}

	cfg := GetPlatformConfig(sup.PT)

	if host := extractHost(sup.URL); host != "" {
		globalRateLimiter.wait(host, 500*time.Millisecond)
	}

	var resp *http.Response
	var err error

	if cfg.ChangePassPath != "" {
		// 2xx 等非标准路径：JSON POST 到自定义路径
		baseURL := strings.TrimRight(sup.URL, "/")
		if !strings.HasPrefix(baseURL, "http") {
			baseURL = "http://" + baseURL
		}
		apiURL := baseURL + cfg.ChangePassPath
		jsonData, _ := json.Marshal(map[string]string{"id": yid, "username": sup.User, "pass": newPwd})
		resp, err = s.client.Post(apiURL, "application/json", strings.NewReader(string(jsonData)))
	} else {
		changeAct := cfg.ChangePassAct
		if changeAct == "" {
			changeAct = "gaimi"
		}
		apiURL := s.buildSupplierURL(sup.URL, changeAct)
		formData := url.Values{}
		formData.Set("uid", sup.User)
		formData.Set("key", sup.Pass)
		// 使用平台特定的参数名
		idParam := cfg.ChangePassIDParam
		if idParam == "" {
			idParam = "id"
		}
		pwdParam := cfg.ChangePassParam
		if pwdParam == "" {
			pwdParam = "newPwd"
		}
		formData.Set(idParam, yid)
		formData.Set(pwdParam, newPwd)
		resp, err = s.client.PostForm(apiURL, formData)
	}

	if err != nil {
		return -1, "", fmt.Errorf("请求上游失败：%v", err)
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

// ResubmitOrder 补单/补刷（按 PHP bsjk.php budanWk）
func (s *SupplierService) ResubmitOrder(sup *model.SupplierFull, yid string) (int, string, error) {
	if sup.PT == "tuboshu" || sup.PT == "nx" {
		return -1, "当前平台暂不支持补单操作", nil
	}

	cfg := GetPlatformConfig(sup.PT)

	if host := extractHost(sup.URL); host != "" {
		globalRateLimiter.wait(host, 500*time.Millisecond)
	}

	var resp *http.Response
	var err error

	if cfg.ResubmitPath != "" {
		// 2xx 等非标准路径：JSON POST
		baseURL := strings.TrimRight(sup.URL, "/")
		if !strings.HasPrefix(baseURL, "http") {
			baseURL = "http://" + baseURL
		}
		apiURL := baseURL + cfg.ResubmitPath
		jsonData, _ := json.Marshal(map[string]string{"id": yid, "username": sup.User})
		resp, err = s.client.Post(apiURL, "application/json", strings.NewReader(string(jsonData)))
	} else {
		apiURL := s.buildSupplierURL(sup.URL, "budan")
		formData := url.Values{}
		formData.Set("uid", sup.User)
		formData.Set("key", sup.Pass)
		formData.Set(cfg.ResubmitIDParam, yid)
		resp, err = s.client.PostForm(apiURL, formData)
	}

	if err != nil {
		return -1, "", fmt.Errorf("请求上游失败：%v", err)
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

// ResetOrderScore 重置订单目标分数（PUP 平台：act=resetscore，newscore 70-100）
func (s *SupplierService) ResetOrderScore(sup *model.SupplierFull, yid string, newscore int) (int, string, error) {
	if sup.PT != "pup" {
		return -1, "当前平台不支持重置分数", nil
	}
	if newscore < 70 || newscore > 100 {
		return -1, "分数范围 70-100", nil
	}

	if host := extractHost(sup.URL); host != "" {
		globalRateLimiter.wait(host, 500*time.Millisecond)
	}

	apiURL := s.buildSupplierURL(sup.URL, "resetscore")
	formData := url.Values{}
	formData.Set("uid", sup.User)
	formData.Set("key", sup.Pass)
	formData.Set("oid", yid)
	formData.Set("newscore", fmt.Sprintf("%d", newscore))

	resp, err := s.client.PostForm(apiURL, formData)
	if err != nil {
		return -1, "", fmt.Errorf("请求上游失败：%v", err)
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

// ResetOrderDuration 重置订单目标时长（PUP 平台：act=resetsc，newsc 0-50 小时）
func (s *SupplierService) ResetOrderDuration(sup *model.SupplierFull, yid string, newsc int) (int, string, error) {
	if sup.PT != "pup" {
		return -1, "当前平台不支持重置时长", nil
	}
	if newsc < 0 || newsc > 50 {
		return -1, "时长范围 0-50 小时", nil
	}

	if host := extractHost(sup.URL); host != "" {
		globalRateLimiter.wait(host, 500*time.Millisecond)
	}

	apiURL := s.buildSupplierURL(sup.URL, "resetsc")
	formData := url.Values{}
	formData.Set("uid", sup.User)
	formData.Set("key", sup.Pass)
	formData.Set("oid", yid)
	formData.Set("newsc", fmt.Sprintf("%d", newsc))

	resp, err := s.client.PostForm(apiURL, formData)
	if err != nil {
		return -1, "", fmt.Errorf("请求上游失败：%v", err)
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

// ResetOrderPeriod 重置订单刷课周期（PUP 平台：act=resetcron，newcron 1-20 天）
func (s *SupplierService) ResetOrderPeriod(sup *model.SupplierFull, yid string, newcron int) (int, string, error) {
	if sup.PT != "pup" {
		return -1, "当前平台不支持重置周期", nil
	}
	if newcron < 1 || newcron > 20 {
		return -1, "周期范围 1-20 天", nil
	}

	if host := extractHost(sup.URL); host != "" {
		globalRateLimiter.wait(host, 500*time.Millisecond)
	}

	apiURL := s.buildSupplierURL(sup.URL, "resetcron")
	formData := url.Values{}
	formData.Set("uid", sup.User)
	formData.Set("key", sup.Pass)
	formData.Set("oid", yid)
	formData.Set("newcron", fmt.Sprintf("%d", newcron))

	resp, err := s.client.PostForm(apiURL, formData)
	if err != nil {
		return -1, "", fmt.Errorf("请求上游失败：%v", err)
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

// OrderLogEntry 订单日志条目（统一格式）
type OrderLogEntry struct {
	Time    string `json:"time"`
	Course  string `json:"course"`
	Status  string `json:"status"`
	Process string `json:"process"`
	Remarks string `json:"remarks"`
}

// QueryOrderLogs 查询订单实时日志（按 PHP Checkorder/logjk.php logWk）
func (s *SupplierService) QueryOrderLogs(sup *model.SupplierFull, yid string, orderExtra ...map[string]string) ([]OrderLogEntry, error) {
	if sup.PT == "yyy" || sup.PT == "tuboshu" {
		return nil, fmt.Errorf("当前平台暂不支持查看日志")
	}

	// 龙龙平台：SSE 流式日志 GET /api/streamLogs?id=&key=
	if sup.PT == "longlong" {
		return s.queryLonglongLogs(sup, yid)
	}

	cfg := GetPlatformConfig(sup.PT)

	if host := extractHost(sup.URL); host != "" {
		globalRateLimiter.wait(host, 500*time.Millisecond)
	}

	logClient := &http.Client{Timeout: 8 * time.Second}
	var resp *http.Response
	var err error

	if cfg.LogPath != "" {
		// KUN/kunba: GET /log/?account=&password=&course=&courseId=&dtoken=
		// wanzi: GET /business/order/{id}/logs?pageSize=50
		baseURL := strings.TrimRight(sup.URL, "/")
		if !strings.HasPrefix(baseURL, "http") {
			baseURL = "http://" + baseURL
		}

		// wanzi 平台特殊处理：/business/order/{id}/logs?uid=&key=
		if sup.PT == "wanzi" {
			apiURL := fmt.Sprintf("%s%s%s/logs?pageSize=50&uid=%s&key=%s", baseURL, cfg.LogPath, yid, sup.User, sup.Pass)
			resp, err = logClient.Get(apiURL)
		} else {
			// KUN/kunba 标准处理
			params := url.Values{}
			if len(orderExtra) > 0 {
				if v, ok := orderExtra[0]["user"]; ok {
					params.Set("account", v)
				}
				if v, ok := orderExtra[0]["pass"]; ok {
					params.Set("password", v)
				}
				if v, ok := orderExtra[0]["kcname"]; ok {
					params.Set("course", v)
				}
				if v, ok := orderExtra[0]["kcid"]; ok {
					params.Set("courseId", v)
				}
			}
			params.Set("dtoken", sup.Token)
			apiURL := baseURL + cfg.LogPath + "?" + params.Encode()
			resp, err = logClient.Get(apiURL)
		}
	} else {
		// 标准：POST api.php?act=xq with uid/key/id
		logAct := cfg.LogAct
		if logAct == "" {
			logAct = "xq"
		}
		logIDParam := cfg.LogIDParam
		if logIDParam == "" {
			logIDParam = "id"
		}
		apiURL := s.buildSupplierURL(sup.URL, logAct)
		formData := url.Values{}
		formData.Set("uid", sup.User)
		formData.Set("key", sup.Pass)
		formData.Set(logIDParam, yid)
		resp, err = logClient.PostForm(apiURL, formData)
	}

	if err != nil {
		return nil, fmt.Errorf("请求上游超时或失败")
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	// 尝试解析 {code, logs/data} 两种格式
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("解析响应失败：%s", string(body))
	}

	// 检查 code
	var code int
	if c, ok := raw["code"]; ok {
		json.Unmarshal(c, &code)
	}
	if code != 1 && code != 0 {
		var msg string
		if m, ok := raw["msg"]; ok {
			json.Unmarshal(m, &msg)
		}
		if msg == "" {
			msg = "查询日志失败"
		}
		return nil, fmt.Errorf("%s", msg)
	}

	// 优先取 logs 字段，其次 data 字段
	var logsRaw json.RawMessage
	if l, ok := raw["logs"]; ok {
		logsRaw = l
	} else if d, ok := raw["data"]; ok {
		logsRaw = d
	}
	if logsRaw == nil {
		return []OrderLogEntry{}, nil
	}

	// 尝试解析为 []OrderLogEntry（对象数组）
	var entries []OrderLogEntry
	if err := json.Unmarshal(logsRaw, &entries); err == nil {
		return entries, nil
	}

	// 尝试解析为 []string（字符串数组，如 simple 平台格式）
	var strLogs []string
	if err := json.Unmarshal(logsRaw, &strLogs); err == nil {
		for _, s := range strLogs {
			entries = append(entries, OrderLogEntry{Remarks: s})
		}
		return entries, nil
	}

	// 尝试解析为 [][]interface{}（KUN 二维数组格式）
	var arrLogs [][]interface{}
	if err := json.Unmarshal(logsRaw, &arrLogs); err == nil {
		for _, row := range arrLogs {
			entry := OrderLogEntry{}
			if len(row) > 8 {
				entry.Time = fmt.Sprintf("%v", row[8])
			}
			if len(row) > 4 {
				entry.Course = fmt.Sprintf("%v", row[4])
			}
			if len(row) > 9 {
				entry.Status = fmt.Sprintf("%v", row[9])
			}
			if len(row) > 10 {
				entry.Process = fmt.Sprintf("%v", row[10])
			}
			if len(row) > 11 {
				entry.Remarks = fmt.Sprintf("%v", row[11])
			}
			entries = append(entries, entry)
		}
		return entries, nil
	}

	return []OrderLogEntry{}, nil
}

// QueryBalance 查询供应商余额（按平台配置驱动）
func (s *SupplierService) QueryBalance(hid int) (map[string]interface{}, error) {
	sup, err := s.GetSupplierByHID(hid)
	if err != nil {
		return nil, err
	}

	// yyy 平台走独立适配器
	if sup.PT == "yyy" {
		result, err := yyyQueryBalance(sup)
		if err != nil {
			return nil, err
		}
		if money, ok := result["money"].(string); ok && money != "" && money != "<nil>" {
			database.DB.Exec("UPDATE qingka_wangke_huoyuan SET money = ? WHERE hid = ?", money, hid)
		}
		return result, nil
	}

	// 土拨鼠平台：GET /userInfo，satoken 认证
	if sup.PT == "tuboshu" {
		baseURL := strings.TrimRight(sup.URL, "/")
		if !strings.HasPrefix(baseURL, "http") {
			baseURL = "https://" + baseURL
		}
		// API 需要 /api/ 前缀
		if !strings.HasSuffix(baseURL, "/api") {
			baseURL = baseURL + "/api"
		}
		apiURL := baseURL + "/userInfo"
		log.Printf("[Tuboshu Balance] URL=%s Token=%s", apiURL, sup.Token[:min(len(sup.Token), 20)]+"...")
		req, reqErr := http.NewRequest("GET", apiURL, nil)
		if reqErr != nil {
			return nil, fmt.Errorf("构建请求失败：%v", reqErr)
		}
		req.Header.Set("Authorization", "Bearer "+sup.Token)
		req.Header.Set("Accept", "application/json")
		resp, err := s.client.Do(req)
		if err != nil {
			log.Printf("[Tuboshu Balance] 请求失败: %v", err)
			return nil, fmt.Errorf("请求土拨鼠余额接口失败：%v", err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[Tuboshu Balance] HTTP %d, Body=%s", resp.StatusCode, string(body[:min(len(body), 500)]))
		var raw map[string]interface{}
		if err := json.Unmarshal(body, &raw); err != nil {
			return nil, fmt.Errorf("解析土拨鼠响应失败：%s", string(body))
		}
		// 提取 point 字段作为余额
		money := "0"
		if data, ok := raw["data"].(map[string]interface{}); ok {
			if pt, ok := data["point"]; ok {
				money = fmt.Sprintf("%v", pt)
			}
		}
		database.DB.Exec("UPDATE qingka_wangke_huoyuan SET money = ? WHERE hid = ?", money, hid)
		return map[string]interface{}{
			"code":  200,
			"money": money,
			"pt":    sup.PT,
			"name":  sup.Name,
			"hid":   hid,
			"raw":   raw,
		}, nil
	}

	cfg := GetPlatformConfig(sup.PT)

	// 构建请求 URL
	baseURL := strings.TrimRight(sup.URL, "/")
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "http://" + baseURL
	}

	var apiURL string
	if cfg.BalancePath != "" {
		// REST 路径模式（如 2xx: /api/getinfo, nx: /api/getuserinfo/）
		apiURL = baseURL + cfg.BalancePath
	} else {
		// 标准 act 模式
		apiURL = fmt.Sprintf("%s/api.php?act=%s", baseURL, cfg.BalanceAct)
	}

	// 构建请求参数
	formData := url.Values{}

	// 根据认证方式设置参数
	authType := cfg.BalanceAuthType
	if authType == "" {
		// 默认 uid+key
		authType = "uid_key"
	}

	var resp *http.Response

	switch authType {
	case "bearer_token":
		// nx 平台：Bearer token + cookie
		req, reqErr := http.NewRequest("GET", apiURL, nil)
		if reqErr != nil {
			return nil, fmt.Errorf("构建请求失败：%v", reqErr)
		}
		req.Header.Set("Authorization", "Bearer "+sup.Token)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "*/*")
		if sup.Cookie != "" {
			cookieStr := sup.Cookie
			if !strings.Contains(cookieStr, "=") {
				cookieStr = "session_id=" + cookieStr
			}
			req.Header.Set("Cookie", cookieStr)
		}
		resp, err = s.client.Do(req)
	case "token_only":
		// token 在 pass 字段 (如 2xx)
		if cfg.UseJSON {
			jsonData, _ := json.Marshal(map[string]string{"token": sup.Pass})
			req, _ := http.NewRequest("POST", apiURL, strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			resp, err = s.client.Do(req)
		} else {
			formData.Set("token", sup.Pass)
			resp, err = s.client.PostForm(apiURL, formData)
		}
	default:
		// uid + key 标准认证
		formData.Set("uid", sup.User)
		formData.Set("key", sup.Pass)
		resp, err = s.client.PostForm(apiURL, formData)
	}

	if err != nil {
		return nil, fmt.Errorf("请求上游余额接口失败：%v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败：%v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("解析响应失败：%s", string(body))
	}

	// 提取余额值
	money := extractMoneyField(raw, cfg.BalanceMoneyField)

	// 统一返回格式
	result := map[string]interface{}{
		"code":  200,
		"money": money,
		"pt":    sup.PT,
		"name":  sup.Name,
		"hid":   hid,
		"raw":   raw,
	}

	// 同步更新数据库里的余额缓存
	if money != "" && money != "<nil>" {
		database.DB.Exec("UPDATE qingka_wangke_huoyuan SET money = ? WHERE hid = ?", money, hid)
	}

	return result, nil
}

// extractMoneyField 从响应 JSON 中按路径提取余额值
// 支持："money" (根级), "data.money", "data" (data 本身是值), "data.remainscore"
func extractMoneyField(raw map[string]interface{}, fieldPath string) string {
	parts := strings.Split(fieldPath, ".")
	var current interface{} = raw

	for _, part := range parts {
		if m, ok := current.(map[string]interface{}); ok {
			current = m[part]
		} else {
			return fmt.Sprintf("%v", current)
		}
	}
	return fmt.Sprintf("%v", current)
}

// queryLonglongLogs 读取龙龙平台 SSE 流式日志（GET /api/streamLogs?id=&key=）
// SSE 每行格式：data: <json> 或 data: <text>，读完整个流后解析
func (s *SupplierService) queryLonglongLogs(sup *model.SupplierFull, yid string) ([]OrderLogEntry, error) {
	baseURL := strings.TrimRight(sup.URL, "/")
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "http://" + baseURL
	}
	apiURL := fmt.Sprintf("%s/api/streamLogs?id=%s&key=%s", baseURL, url.QueryEscape(yid), url.QueryEscape(sup.Pass))

	logClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := logClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("请求龙龙日志失败：%v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var entries []OrderLogEntry
	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" || data == "[DONE]" {
			continue
		}
		// 尝试解析为 JSON 对象
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(data), &obj); err == nil {
			entry := OrderLogEntry{}
			if v, ok := obj["time"]; ok {
				entry.Time = fmt.Sprintf("%v", v)
			}
			if v, ok := obj["course"]; ok {
				entry.Course = fmt.Sprintf("%v", v)
			}
			if v, ok := obj["status"]; ok {
				entry.Status = fmt.Sprintf("%v", v)
			}
			if v, ok := obj["process"]; ok {
				entry.Process = fmt.Sprintf("%v", v)
			}
			if v, ok := obj["remarks"]; ok {
				entry.Remarks = fmt.Sprintf("%v", v)
			}
			// 兼容 message/msg 字段
			if entry.Remarks == "" {
				if v, ok := obj["message"]; ok {
					entry.Remarks = fmt.Sprintf("%v", v)
				} else if v, ok := obj["msg"]; ok {
					entry.Remarks = fmt.Sprintf("%v", v)
				}
			}
			entries = append(entries, entry)
		} else {
			// 纯文本行
			entries = append(entries, OrderLogEntry{Remarks: data})
		}
	}
	return entries, nil
}

// SubmitReport 向上游供应商提交工单反馈（按平台配置处理差异）
// 返回 (code, workId, msg, err)
func (s *SupplierService) SubmitReport(sup *model.SupplierFull, yid, ticketType, content string) (int, int, string, error) {
	cfg := GetPlatformConfig(sup.PT)
	baseURL := strings.TrimRight(sup.URL, "/")
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "http://" + baseURL
	}
	if host := extractHost(sup.URL); host != "" {
		globalRateLimiter.wait(host, 500*time.Millisecond)
	}
	var resp *http.Response
	var err error
	if cfg.ReportParamStyle == "token" {
		apiURL := baseURL + cfg.ReportPath
		jsonData, _ := json.Marshal(map[string]string{
			"token":   sup.Pass,
			"type":    ticketType,
			"id":      yid,
			"content": content,
		})
		req, _ := http.NewRequest("POST", apiURL, strings.NewReader(string(jsonData)))
		req.Header.Set("Content-Type", "application/json")
		resp, err = s.client.Do(req)
	} else {
		var apiURL string
		if cfg.ReportPath != "" {
			apiURL = baseURL + cfg.ReportPath
		} else {
			apiURL = fmt.Sprintf("%s/api.php?act=%s", baseURL, cfg.ReportAct)
		}
		formData := url.Values{
			"uid":      {sup.User},
			"key":      {sup.Pass},
			"id":       {yid},
			"question": {content},
		}
		resp, err = s.client.PostForm(apiURL, formData)
	}
	if err != nil {
		return 0, 0, "", fmt.Errorf("请求上游失败：%v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return 0, 0, "", fmt.Errorf("上游返回解析失败：%s", string(body))
	}
	code := 0
	if c, ok := raw["code"].(float64); ok {
		code = int(c)
	}
	msg, _ := raw["msg"].(string)
	workId := 0
	if wid, ok := raw["workId"].(float64); ok {
		workId = int(wid)
	} else if dataMap, ok := raw["data"].(map[string]interface{}); ok {
		if rid, ok := dataMap["reportId"].(float64); ok {
			workId = int(rid)
		}
	}
	return code, workId, msg, nil
}

// QueryReport 查询上游供应商工单反馈状态（按平台配置处理差异）
// 返回 (code, answer, state, err)
func (s *SupplierService) QueryReport(sup *model.SupplierFull, reportID string) (int, string, string, error) {
	cfg := GetPlatformConfig(sup.PT)
	baseURL := strings.TrimRight(sup.URL, "/")
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "http://" + baseURL
	}
	if host := extractHost(sup.URL); host != "" {
		globalRateLimiter.wait(host, 500*time.Millisecond)
	}
	var resp *http.Response
	var err error
	if cfg.ReportParamStyle == "token" {
		apiURL := baseURL + cfg.GetReportPath
		jsonData, _ := json.Marshal(map[string]string{
			"token":  sup.Pass,
			"workId": reportID,
		})
		req, _ := http.NewRequest("POST", apiURL, strings.NewReader(string(jsonData)))
		req.Header.Set("Content-Type", "application/json")
		resp, err = s.client.Do(req)
	} else {
		var apiURL string
		if cfg.GetReportPath != "" {
			apiURL = baseURL + cfg.GetReportPath
		} else {
			apiURL = fmt.Sprintf("%s/api.php?act=%s", baseURL, cfg.GetReportAct)
		}
		formData := url.Values{
			"uid":      {sup.User},
			"key":      {sup.Pass},
			"reportId": {reportID},
		}
		resp, err = s.client.PostForm(apiURL, formData)
	}
	if err != nil {
		return 0, "", "", fmt.Errorf("请求上游失败：%v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return 0, "", "", fmt.Errorf("上游返回解析失败：%s", string(body))
	}
	code := 0
	if c, ok := raw["code"].(float64); ok {
		code = int(c)
	}
	answer := ""
	state := ""
	if dataMap, ok := raw["data"].(map[string]interface{}); ok {
		if a, ok := dataMap["answer"].(string); ok {
			answer = a
		}
		if st, ok := dataMap["state"].(string); ok {
			state = st
		}
		if state == "" {
			if s, ok := dataMap["status"].(float64); ok {
				state = fmt.Sprintf("%d", int(s))
			}
		}
	}
	return code, answer, state, nil
}

// extractHost 从 URL 中提取主机名（用于限速 key）
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
