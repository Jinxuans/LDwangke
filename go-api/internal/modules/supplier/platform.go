package supplier

import (
	"sync"

	"go-api/internal/database"
)

// PlatformConfig 平台接口差异配置。
type PlatformConfig struct {
	QueryAct          string
	OrderAct          string
	ProgressAct       string
	ProgressNoYID     string
	ProgressPath      string
	ProgressMethod    string
	SuccessCode       string
	ReturnsYID        bool
	ExtraParams       bool
	UseIDParam        bool
	AlwaysUsername    bool
	YIDInDataArray    bool
	UseUUIDParam      bool
	PauseAct          string
	PauseIDParam      string
	ChangePassAct     string
	PausePath         string
	ChangePassPath    string
	UseJSON           bool
	LogPath           string
	LogMethod         string
	LogAct            string
	LogIDParam        string
	ChangePassParam   string
	ChangePassIDParam string
	ResubmitPath      string
	ResubmitIDParam   string
	BalanceAct        string
	BalancePath       string
	BalanceMoneyField string
	BalanceMethod     string
	BalanceAuthType   string
	ReportAct         string
	ReportPath        string
	GetReportAct      string
	GetReportPath     string
	ReportSuccessCode string
	ReportParamStyle  string
	ReportAuthType    string
	RefreshPath       string
	CategoryAct       string
}

var platformRegistry = map[string]PlatformConfig{
	"27":       {QueryAct: "local_time", SuccessCode: "0", ExtraParams: true},
	"zy":       {QueryAct: "local_time", SuccessCode: "0"},
	"haha":     {SuccessCode: "0", ReturnsYID: true, ExtraParams: true, ProgressPath: "/api/search", ProgressMethod: "GET"},
	"hzw":      {SuccessCode: "1", ReturnsYID: true, ProgressAct: "chadan", ProgressNoYID: "chadan", UseIDParam: true, AlwaysUsername: true, PauseAct: "stop", LogAct: "cha_logwk", BalanceMoneyField: "money"},
	"longlong": {SuccessCode: "0", ReturnsYID: true, YIDInDataArray: true, UseUUIDParam: true, ProgressAct: "chadan", ProgressNoYID: "chadan", PauseAct: "zanting", BalanceAct: "money", BalanceMoneyField: "data"},
	"liunian":  {SuccessCode: "0", ReturnsYID: true, ProgressPath: "/api/chadan1", ChangePassAct: "xgmm", ChangePassParam: "xgmm", PauseAct: "zt", LogAct: "xq", CategoryAct: "getfl", ReportAct: "submitWorkOrder", GetReportAct: "queryWorkOrder", BalanceMoneyField: "money"},
	"xxtgf":    {QueryAct: "local_script"},
	"moocmd":   {QueryAct: "local_script"},
	"yyy":      {QueryAct: "yyy_custom", SuccessCode: "200", ReturnsYID: true, BalanceMoneyField: "money"},
	"2xx":      {SuccessCode: "1", ReturnsYID: true, PausePath: "/api/stop", ChangePassPath: "/api/update", ResubmitPath: "/api/reset", RefreshPath: "/api/refresh", UseJSON: true, BalancePath: "/api/getinfo", BalanceMoneyField: "data.money", ReportPath: "/api/submitWork", GetReportPath: "/api/queryWork", ReportParamStyle: "token", ReportAuthType: "token_only"},
	"KUN":      {QueryAct: "KUN_custom", SuccessCode: "0", LogPath: "/log/", LogMethod: "GET"},
	"kunba":    {QueryAct: "KUN_custom", SuccessCode: "0", LogPath: "/log/", LogMethod: "GET"},
	"Benz":     {SuccessCode: "0", ReturnsYID: true, ProgressAct: "chadan", ProgressNoYID: "chadan", ChangePassAct: "xgmm", ChangePassIDParam: "oid", ChangePassParam: "pwd", LogAct: "getOrderLogs", LogIDParam: "oid", PauseAct: "ztdd", PauseIDParam: "oid"},
	"tuboshu":  {QueryAct: "tuboshu_custom", SuccessCode: "0", BalanceMoneyField: "data.money"},
	"29":       {SuccessCode: "0", ChangePassAct: "xgmm", ChangePassParam: "xgmm", BalanceMoneyField: "money"},
	"spi":      {SuccessCode: "0", ReturnsYID: true, ProgressPath: "/api/search", ProgressMethod: "GET", ChangePassAct: "xgmm", ChangePassParam: "newPwd", BalanceMoneyField: "money"},
	"lg":       {SuccessCode: "0", ReturnsYID: true, BalanceMoneyField: "data.money"},
	"nx":       {QueryAct: "nx_custom", SuccessCode: "0", ReturnsYID: true, BalancePath: "/api/getuserinfo/", BalanceMoneyField: "data.remainscore", BalanceAuthType: "bearer_token"},
	"pup":      {SuccessCode: "0", ReturnsYID: true, ExtraParams: true, ProgressAct: "chadan", ProgressNoYID: "chadan", ChangePassAct: "updateorderpwd", ChangePassIDParam: "oid", ChangePassParam: "newpwd", LogAct: "orderlog", LogIDParam: "oid", ResubmitIDParam: "oid"},
	"wanzi":    {SuccessCode: "1", ReturnsYID: true, ProgressAct: "chadan", ProgressNoYID: "chadan", LogAct: "getOrderLogs", ChangePassAct: "xgmm", ChangePassIDParam: "oid", ChangePassParam: "pwd", PauseAct: "pause"},
	"lgwk":     {QueryAct: "lgwk_custom", SuccessCode: "0"},
	"skyriver": {SuccessCode: "1", ReturnsYID: true, ProgressPath: "/api/chadan1", ProgressNoYID: "chadan", ChangePassAct: "xgmm", ChangePassParam: "newpass", ChangePassIDParam: "oid", CategoryAct: "getfl", ReportAct: "submitWorkOrder", GetReportAct: "queryWorkOrder", BalanceMoneyField: "money"},
	"xuemei":   {SuccessCode: "0", ReturnsYID: true, ProgressAct: "chadan", ProgressNoYID: "chadan", AlwaysUsername: true, PauseAct: "ikunStop", ChangePassAct: "gaimi", ChangePassIDParam: "oid", ChangePassParam: "pwd", LogAct: "cha_logwk", LogIDParam: "oid", BalanceMoneyField: "money"},
	"simple":   {QueryAct: "simple_custom", SuccessCode: "1", BalanceMoneyField: "money"},
}

var (
	dbConfigCache  = map[string]PlatformConfig{}
	dbNameCache    = map[string]string{}
	dbConfigMu     sync.RWMutex
	dbConfigLoaded bool
)

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
			continue
		}
		cfg.SuccessCode = successCodes
		newConfigs[pt] = cfg
		newNames[pt] = name
	}

	if len(newConfigs) > 0 {
		dbConfigCache = newConfigs
		dbNameCache = newNames
	}
	dbConfigLoaded = true
}

func ReloadPlatformConfigs() {
	dbConfigMu.Lock()
	dbConfigLoaded = false
	dbConfigMu.Unlock()
	loadDBPlatformConfigs()
}

func GetPlatformConfig(pt string) PlatformConfig {
	dbConfigMu.RLock()
	loaded := dbConfigLoaded
	dbConfigMu.RUnlock()
	if !loaded {
		loadDBPlatformConfigs()
	}

	dbConfigMu.RLock()
	if cfg, ok := dbConfigCache[pt]; ok {
		dbConfigMu.RUnlock()
		return fillDefaults(cfg)
	}
	dbConfigMu.RUnlock()

	if cfg, ok := platformRegistry[pt]; ok {
		return fillDefaults(cfg)
	}

	return fillDefaults(PlatformConfig{})
}

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

func GetPlatformNames() map[string]string {
	dbConfigMu.RLock()
	loaded := dbConfigLoaded
	dbConfigMu.RUnlock()
	if !loaded {
		loadDBPlatformConfigs()
	}

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
		"simple":   "至强",
	}

	dbConfigMu.RLock()
	for pt, name := range dbNameCache {
		result[pt] = name
	}
	dbConfigMu.RUnlock()

	return result
}
