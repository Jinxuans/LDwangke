package model

import "time"

// PlatformConfigDB 平台接口配置（数据库模型）
type PlatformConfigDB struct {
	ID                int       `json:"id" db:"id"`
	PT                string    `json:"pt" db:"pt"`
	Name              string    `json:"name" db:"name"`
	AuthType          string    `json:"auth_type" db:"auth_type"`
	APIPathStyle      string    `json:"api_path_style" db:"api_path_style"`
	SuccessCodes      string    `json:"success_codes" db:"success_codes"`
	UseJSON           bool      `json:"use_json" db:"use_json"`
	NeedProxy         bool      `json:"need_proxy" db:"need_proxy"`
	ReturnsYID        bool      `json:"returns_yid" db:"returns_yid"`
	ExtraParams       bool      `json:"extra_params" db:"extra_params"`
	QueryAct          string    `json:"query_act" db:"query_act"`
	QueryPath         string    `json:"query_path" db:"query_path"`
	QueryParamStyle   string    `json:"query_param_style" db:"query_param_style"`
	QueryPolling      bool      `json:"query_polling" db:"query_polling"`
	QueryMaxAttempts  int       `json:"query_max_attempts" db:"query_max_attempts"`
	QueryInterval     int       `json:"query_interval" db:"query_interval"`
	QueryResponseMap  string    `json:"query_response_map" db:"query_response_map"`
	OrderAct          string    `json:"order_act" db:"order_act"`
	OrderPath         string    `json:"order_path" db:"order_path"`
	YIDInDataArray    bool      `json:"yid_in_data_array" db:"yid_in_data_array"`
	ProgressAct       string    `json:"progress_act" db:"progress_act"`
	ProgressNoYID     string    `json:"progress_no_yid" db:"progress_no_yid"`
	ProgressPath      string    `json:"progress_path" db:"progress_path"`
	ProgressMethod    string    `json:"progress_method" db:"progress_method"`
	ProgressNeedsAuth bool      `json:"progress_needs_auth" db:"progress_needs_auth"`
	UseIDParam        bool      `json:"use_id_param" db:"use_id_param"`
	UseUUIDParam      bool      `json:"use_uuid_param" db:"use_uuid_param"`
	AlwaysUsername    bool      `json:"always_username" db:"always_username"`
	PauseAct          string    `json:"pause_act" db:"pause_act"`
	PausePath         string    `json:"pause_path" db:"pause_path"`
	PauseIDParam      string    `json:"pause_id_param" db:"pause_id_param"`
	ResumeAct         string    `json:"resume_act" db:"resume_act"`
	ResumePath        string    `json:"resume_path" db:"resume_path"`
	ChangePassAct     string    `json:"change_pass_act" db:"change_pass_act"`
	ChangePassPath    string    `json:"change_pass_path" db:"change_pass_path"`
	ChangePassParam   string    `json:"change_pass_param" db:"change_pass_param"`
	ChangePassIDParam string    `json:"change_pass_id_param" db:"change_pass_id_param"`
	ResubmitPath      string    `json:"resubmit_path" db:"resubmit_path"`
	ResubmitIDParam   string    `json:"resubmit_id_param" db:"resubmit_id_param"`
	LogAct            string    `json:"log_act" db:"log_act"`
	LogPath           string    `json:"log_path" db:"log_path"`
	LogMethod         string    `json:"log_method" db:"log_method"`
	LogIDParam        string    `json:"log_id_param" db:"log_id_param"`
	BalanceAct        string    `json:"balance_act" db:"balance_act"`
	BalancePath       string    `json:"balance_path" db:"balance_path"`
	BalanceMoneyField string    `json:"balance_money_field" db:"balance_money_field"`
	BalanceMethod     string    `json:"balance_method" db:"balance_method"`
	BalanceAuthType   string    `json:"balance_auth_type" db:"balance_auth_type"`
	ReportParamStyle  string    `json:"report_param_style" db:"report_param_style"`
	ReportAuthType    string    `json:"report_auth_type" db:"report_auth_type"`
	ReportPath        string    `json:"report_path" db:"report_path"`
	GetReportPath     string    `json:"get_report_path" db:"get_report_path"`
	RefreshPath       string    `json:"refresh_path" db:"refresh_path"`
	SourceCode        string    `json:"source_code" db:"source_code"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// PlatformConfigSaveRequest 保存平台配置请求
type PlatformConfigSaveRequest struct {
	ID                int    `json:"id"`
	PT                string `json:"pt" binding:"required"`
	Name              string `json:"name"`
	AuthType          string `json:"auth_type"`
	APIPathStyle      string `json:"api_path_style"`
	SuccessCodes      string `json:"success_codes"`
	UseJSON           bool   `json:"use_json"`
	NeedProxy         bool   `json:"need_proxy"`
	ReturnsYID        bool   `json:"returns_yid"`
	ExtraParams       bool   `json:"extra_params"`
	QueryAct          string `json:"query_act"`
	QueryPath         string `json:"query_path"`
	QueryParamStyle   string `json:"query_param_style"`
	QueryPolling      bool   `json:"query_polling"`
	QueryMaxAttempts  int    `json:"query_max_attempts"`
	QueryInterval     int    `json:"query_interval"`
	QueryResponseMap  string `json:"query_response_map"`
	OrderAct          string `json:"order_act"`
	OrderPath         string `json:"order_path"`
	YIDInDataArray    bool   `json:"yid_in_data_array"`
	ProgressAct       string `json:"progress_act"`
	ProgressNoYID     string `json:"progress_no_yid"`
	ProgressPath      string `json:"progress_path"`
	ProgressMethod    string `json:"progress_method"`
	ProgressNeedsAuth bool   `json:"progress_needs_auth"`
	UseIDParam        bool   `json:"use_id_param"`
	UseUUIDParam      bool   `json:"use_uuid_param"`
	AlwaysUsername    bool   `json:"always_username"`
	PauseAct          string `json:"pause_act"`
	PausePath         string `json:"pause_path"`
	PauseIDParam      string `json:"pause_id_param"`
	ResumeAct         string `json:"resume_act"`
	ResumePath        string `json:"resume_path"`
	ChangePassAct     string `json:"change_pass_act"`
	ChangePassPath    string `json:"change_pass_path"`
	ChangePassParam   string `json:"change_pass_param"`
	ChangePassIDParam string `json:"change_pass_id_param"`
	ResubmitPath      string `json:"resubmit_path"`
	ResubmitIDParam   string `json:"resubmit_id_param"`
	LogAct            string `json:"log_act"`
	LogPath           string `json:"log_path"`
	LogMethod         string `json:"log_method"`
	LogIDParam        string `json:"log_id_param"`
	BalanceAct        string `json:"balance_act"`
	BalancePath       string `json:"balance_path"`
	BalanceMoneyField string `json:"balance_money_field"`
	BalanceMethod     string `json:"balance_method"`
	BalanceAuthType   string `json:"balance_auth_type"`
	ReportParamStyle  string `json:"report_param_style"`
	ReportAuthType    string `json:"report_auth_type"`
	ReportPath        string `json:"report_path"`
	GetReportPath     string `json:"get_report_path"`
	RefreshPath       string `json:"refresh_path"`
	SourceCode        string `json:"source_code"`
}
