package supplier

import (
	"strings"
	"sync"

	"go-api/internal/database"
)

// PlatformConfig 平台接口差异配置。
type PlatformConfig struct {
	AuthType              string
	QueryAct              string
	QueryPath             string
	QueryMethod           string
	QueryBodyType         string
	QueryParamMap         string
	OrderPath             string
	OrderMethod           string
	OrderBodyType         string
	OrderParamMap         string
	ProgressPath          string
	ProgressMethod        string
	ProgressBodyType      string
	ProgressParamMap      string
	BatchProgressPath     string
	BatchProgressMethod   string
	BatchProgressBodyType string
	BatchProgressParamMap string
	SuccessCode           string
	ReturnsYID            bool
	ExtraParams           bool
	YIDInDataArray        bool
	PauseMethod           string
	PauseBodyType         string
	PauseParamMap         string
	PauseIDParam          string
	ChangePassMethod      string
	ChangePassBodyType    string
	ChangePassParamMap    string
	PausePath             string
	ChangePassPath        string
	UseJSON               bool
	ResumePath            string
	ResumeMethod          string
	ResumeBodyType        string
	ResumeParamMap        string
	LogPath               string
	LogMethod             string
	LogBodyType           string
	LogParamMap           string
	LogIDParam            string
	ChangePassParam       string
	ChangePassIDParam     string
	ResubmitPath          string
	ResubmitMethod        string
	ResubmitBodyType      string
	ResubmitParamMap      string
	ResubmitIDParam       string
	BalancePath           string
	BalanceMoneyField     string
	BalanceMethod         string
	BalanceBodyType       string
	BalanceParamMap       string
	BalanceAuthType       string
	ReportPath            string
	ReportMethod          string
	ReportBodyType        string
	ReportParamMap        string
	GetReportPath         string
	GetReportMethod       string
	GetReportBodyType     string
	GetReportParamMap     string
	ReportSuccessCode     string
	ReportParamStyle      string
	ReportAuthType        string
	RefreshPath           string
	CategoryPath          string
	CategoryMethod        string
	CategoryBodyType      string
	CategoryParamMap      string
	ClassListPath         string
	ClassListMethod       string
	ClassListBodyType     string
	ClassListParamMap     string
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
		query_act, COALESCE(query_path,''), COALESCE(query_method,''), COALESCE(query_body_type,''), COALESCE(query_param_map,''),
		COALESCE(order_path,''), COALESCE(order_method,''), COALESCE(order_body_type,''), COALESCE(order_param_map,''),
		extra_params, returns_yid,
		COALESCE(progress_path,''), COALESCE(progress_method,''),
		COALESCE(progress_body_type,''), COALESCE(progress_param_map,''),
		COALESCE(batch_progress_path,''), COALESCE(batch_progress_method,''),
		COALESCE(batch_progress_body_type,''), COALESCE(batch_progress_param_map,''), yid_in_data_array,
		COALESCE(category_path,''), COALESCE(category_method,''), COALESCE(category_body_type,''), COALESCE(category_param_map,''),
		COALESCE(class_list_path,''), COALESCE(class_list_method,''), COALESCE(class_list_body_type,''), COALESCE(class_list_param_map,''),
		COALESCE(pause_path,''), COALESCE(pause_method,''), COALESCE(pause_body_type,''), COALESCE(pause_param_map,''), COALESCE(pause_id_param,''),
		COALESCE(resume_path,''), COALESCE(resume_method,''), COALESCE(resume_body_type,''), COALESCE(resume_param_map,''),
		COALESCE(change_pass_param,''), COALESCE(change_pass_id_param,''),
		COALESCE(change_pass_path,''), COALESCE(change_pass_method,''), COALESCE(change_pass_body_type,''), COALESCE(change_pass_param_map,''),
		COALESCE(resubmit_path,''), COALESCE(resubmit_method,''), COALESCE(resubmit_body_type,''), COALESCE(resubmit_param_map,''), COALESCE(resubmit_id_param,''),
		COALESCE(log_path,''), COALESCE(log_method,''), COALESCE(log_body_type,''), COALESCE(log_param_map,''), COALESCE(log_id_param,''), use_json,
		COALESCE(balance_path,''), COALESCE(balance_money_field,''),
		COALESCE(balance_method,''), COALESCE(balance_body_type,''), COALESCE(balance_param_map,''), COALESCE(balance_auth_type,''),
		COALESCE(report_param_style,''), COALESCE(report_auth_type,''),
		COALESCE(report_path,''), COALESCE(report_method,''), COALESCE(report_body_type,''), COALESCE(report_param_map,''),
		COALESCE(get_report_path,''), COALESCE(get_report_method,''), COALESCE(get_report_body_type,''), COALESCE(get_report_param_map,''),
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
			&cfg.ProgressPath, &cfg.ProgressMethod, &cfg.ProgressBodyType, &cfg.ProgressParamMap,
			&cfg.BatchProgressPath, &cfg.BatchProgressMethod, &cfg.BatchProgressBodyType, &cfg.BatchProgressParamMap, &cfg.YIDInDataArray,
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
	if cfg.SuccessCode == "" {
		cfg.SuccessCode = "0"
	}
	if cfg.ReportSuccessCode == "" {
		cfg.ReportSuccessCode = "1"
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
