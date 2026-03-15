package supplier

import (
	"strings"
	"sync"

	"go-api/internal/database"
)

// PlatformConfig 平台接口差异配置。
type PlatformConfig struct {
	AuthType           string
	QueryAct           string
	QueryPath          string
	QueryMethod        string
	QueryBodyType      string
	QueryParamMap      string
	OrderPath          string
	OrderMethod        string
	OrderBodyType      string
	OrderParamMap      string
	ProgressPath       string
	ProgressMethod     string
	ProgressBodyType   string
	ProgressParamMap   string
	SuccessCode        string
	ReturnsYID         bool
	ExtraParams        bool
	YIDInDataArray     bool
	PauseMethod        string
	PauseBodyType      string
	PauseParamMap      string
	PauseIDParam       string
	ChangePassMethod   string
	ChangePassBodyType string
	ChangePassParamMap string
	PausePath          string
	ChangePassPath     string
	UseJSON            bool
	ResumePath         string
	ResumeMethod       string
	ResumeBodyType     string
	ResumeParamMap     string
	LogPath            string
	LogMethod          string
	LogBodyType        string
	LogParamMap        string
	LogIDParam         string
	ChangePassParam    string
	ChangePassIDParam  string
	ResubmitPath       string
	ResubmitMethod     string
	ResubmitBodyType   string
	ResubmitParamMap   string
	ResubmitIDParam    string
	BalancePath        string
	BalanceMoneyField  string
	BalanceMethod      string
	BalanceBodyType    string
	BalanceParamMap    string
	BalanceAuthType    string
	ReportPath         string
	ReportMethod       string
	ReportBodyType     string
	ReportParamMap     string
	GetReportPath      string
	GetReportMethod    string
	GetReportBodyType  string
	GetReportParamMap  string
	ReportSuccessCode  string
	ReportParamStyle   string
	ReportAuthType     string
	RefreshPath        string
	CategoryPath       string
	CategoryMethod     string
	CategoryBodyType   string
	CategoryParamMap   string
	ClassListPath      string
	ClassListMethod    string
	ClassListBodyType  string
	ClassListParamMap  string
}

var (
	dbConfigCache  = map[string]PlatformConfig{}
	dbNameCache    = map[string]string{}
	dbConfigMu     sync.RWMutex
	dbConfigLoaded bool
)

func isCustomQueryDriver(driver string) bool {
	switch strings.TrimSpace(driver) {
	case "local_time", "local_script", "xxt_query", "KUN_custom", "simple_custom", "yyy_custom", "tuboshu_custom", "nx_custom", "lgwk_custom":
		return true
	default:
		return false
	}
}

func loadDBPlatformConfigs() {
	dbConfigMu.Lock()
	defer dbConfigMu.Unlock()

	rows, err := database.DB.Query(`SELECT pt, name, auth_type, success_codes,
		query_act, COALESCE(query_path,''), COALESCE(query_method,'POST'), COALESCE(query_body_type,''), COALESCE(query_param_map,''),
		COALESCE(order_path,''), COALESCE(order_method,'POST'), COALESCE(order_body_type,''), COALESCE(order_param_map,''),
		extra_params, returns_yid,
		progress_path, progress_method,
		COALESCE(progress_body_type,''), COALESCE(progress_param_map,''), yid_in_data_array,
		COALESCE(category_path,''), COALESCE(category_method,'POST'), COALESCE(category_body_type,''), COALESCE(category_param_map,''),
		COALESCE(class_list_path,''), COALESCE(class_list_method,'POST'), COALESCE(class_list_body_type,''), COALESCE(class_list_param_map,''),
		COALESCE(pause_path,''), COALESCE(pause_method,'POST'), COALESCE(pause_body_type,''), COALESCE(pause_param_map,''), COALESCE(pause_id_param,'id'),
		COALESCE(resume_path,''), COALESCE(resume_method,'POST'), COALESCE(resume_body_type,''), COALESCE(resume_param_map,''),
		change_pass_param, change_pass_id_param,
		COALESCE(change_pass_path,''), COALESCE(change_pass_method,'POST'), COALESCE(change_pass_body_type,''), COALESCE(change_pass_param_map,''),
		COALESCE(resubmit_path,''), COALESCE(resubmit_method,'POST'), COALESCE(resubmit_body_type,''), COALESCE(resubmit_param_map,''), COALESCE(resubmit_id_param,'id'),
		COALESCE(log_path,''), COALESCE(log_method,'POST'), COALESCE(log_body_type,''), COALESCE(log_param_map,''), log_id_param, use_json,
		COALESCE(balance_path,''), COALESCE(balance_money_field,'money'),
		COALESCE(balance_method,'POST'), COALESCE(balance_body_type,''), COALESCE(balance_param_map,''), COALESCE(balance_auth_type,''),
		COALESCE(report_param_style,''), COALESCE(report_auth_type,''),
		COALESCE(report_path,''), COALESCE(report_method,'POST'), COALESCE(report_body_type,''), COALESCE(report_param_map,''),
		COALESCE(get_report_path,''), COALESCE(get_report_method,'POST'), COALESCE(get_report_body_type,''), COALESCE(get_report_param_map,''),
		COALESCE(refresh_path,'')
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
			&pt, &name, &cfg.AuthType, &successCodes,
			&cfg.QueryAct, &cfg.QueryPath, &cfg.QueryMethod, &cfg.QueryBodyType, &cfg.QueryParamMap,
			&cfg.OrderPath, &cfg.OrderMethod, &cfg.OrderBodyType, &cfg.OrderParamMap,
			&cfg.ExtraParams, &cfg.ReturnsYID,
			&cfg.ProgressPath, &cfg.ProgressMethod, &cfg.ProgressBodyType, &cfg.ProgressParamMap, &cfg.YIDInDataArray,
			&cfg.CategoryPath, &cfg.CategoryMethod, &cfg.CategoryBodyType, &cfg.CategoryParamMap,
			&cfg.ClassListPath, &cfg.ClassListMethod, &cfg.ClassListBodyType, &cfg.ClassListParamMap,
			&cfg.PausePath, &cfg.PauseMethod, &cfg.PauseBodyType, &cfg.PauseParamMap, &cfg.PauseIDParam,
			&cfg.ResumePath, &cfg.ResumeMethod, &cfg.ResumeBodyType, &cfg.ResumeParamMap,
			&cfg.ChangePassParam, &cfg.ChangePassIDParam,
			&cfg.ChangePassPath, &cfg.ChangePassMethod, &cfg.ChangePassBodyType, &cfg.ChangePassParamMap,
			&cfg.ResubmitPath, &cfg.ResubmitMethod, &cfg.ResubmitBodyType, &cfg.ResubmitParamMap, &cfg.ResubmitIDParam,
			&cfg.LogPath, &cfg.LogMethod, &cfg.LogBodyType, &cfg.LogParamMap, &cfg.LogIDParam, &cfg.UseJSON,
			&cfg.BalancePath, &cfg.BalanceMoneyField, &cfg.BalanceMethod, &cfg.BalanceBodyType, &cfg.BalanceParamMap, &cfg.BalanceAuthType,
			&cfg.ReportParamStyle, &cfg.ReportAuthType,
			&cfg.ReportPath, &cfg.ReportMethod, &cfg.ReportBodyType, &cfg.ReportParamMap,
			&cfg.GetReportPath, &cfg.GetReportMethod, &cfg.GetReportBodyType, &cfg.GetReportParamMap, &cfg.RefreshPath,
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

	return fillDefaults(PlatformConfig{})
}

func fillDefaults(cfg PlatformConfig) PlatformConfig {
	if cfg.AuthType == "" {
		cfg.AuthType = "uid_key"
	}
	if !isCustomQueryDriver(cfg.QueryAct) {
		if cfg.QueryPath == "" {
			cfg.QueryPath = "/api.php?act=get"
		}
	}
	if cfg.QueryMethod == "" {
		cfg.QueryMethod = "POST"
	}
	if cfg.OrderPath == "" {
		cfg.OrderPath = "/api.php?act=add"
	}
	if cfg.OrderMethod == "" {
		cfg.OrderMethod = "POST"
	}
	if cfg.ProgressPath == "" {
		cfg.ProgressPath = "/api.php?act=chadan2"
	}
	if cfg.ProgressMethod == "" {
		cfg.ProgressMethod = "POST"
	}
	if cfg.SuccessCode == "" {
		cfg.SuccessCode = "0"
	}
	if cfg.PauseMethod == "" {
		cfg.PauseMethod = "POST"
	}
	if cfg.PausePath == "" {
		cfg.PausePath = "/api.php?act=zt"
	}
	if cfg.ResumeMethod == "" {
		cfg.ResumeMethod = "POST"
	}
	if cfg.ChangePassMethod == "" {
		cfg.ChangePassMethod = "POST"
	}
	if cfg.ChangePassPath == "" {
		cfg.ChangePassPath = "/api.php?act=gaimi"
	}
	if cfg.ResubmitMethod == "" {
		cfg.ResubmitMethod = "POST"
	}
	if cfg.ResubmitPath == "" {
		cfg.ResubmitPath = "/api.php?act=budan"
	}
	if cfg.LogMethod == "" {
		cfg.LogMethod = "POST"
	}
	if cfg.LogPath == "" {
		cfg.LogPath = "/api.php?act=xq"
	}
	if cfg.CategoryPath == "" {
		cfg.CategoryPath = "/api.php?act=getcate"
	}
	if cfg.CategoryMethod == "" {
		cfg.CategoryMethod = "POST"
	}
	if cfg.ClassListPath == "" {
		cfg.ClassListPath = "/api.php?act=getclass"
	}
	if cfg.ClassListMethod == "" {
		cfg.ClassListMethod = "POST"
	}
	if cfg.BalancePath == "" {
		cfg.BalancePath = "/api.php?act=getmoney"
	}
	if cfg.BalanceMoneyField == "" {
		cfg.BalanceMoneyField = "money"
	}
	if cfg.BalanceMethod == "" {
		cfg.BalanceMethod = "POST"
	}
	if cfg.ReportPath == "" {
		cfg.ReportPath = "/api.php?act=report"
	}
	if cfg.ReportMethod == "" {
		cfg.ReportMethod = "POST"
	}
	if cfg.GetReportPath == "" {
		cfg.GetReportPath = "/api.php?act=getReport"
	}
	if cfg.GetReportMethod == "" {
		cfg.GetReportMethod = "POST"
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
	return cfg
}

func GetPlatformNames() map[string]string {
	dbConfigMu.RLock()
	loaded := dbConfigLoaded
	dbConfigMu.RUnlock()
	if !loaded {
		loadDBPlatformConfigs()
	}

	result := map[string]string{}

	dbConfigMu.RLock()
	for pt, name := range dbNameCache {
		result[pt] = name
	}
	dbConfigMu.RUnlock()

	return result
}
