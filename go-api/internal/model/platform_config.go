package model

import "time"

// PlatformConfigDB 平台接口配置（数据库模型）
type PlatformConfigDB struct {
	ID                 int       `json:"id" db:"id"`
	PT                 string    `json:"pt" db:"pt"`
	Name               string    `json:"name" db:"name"`
	AuthType           string    `json:"auth_type" db:"auth_type"`
	SuccessCodes       string    `json:"success_codes" db:"success_codes"`
	UseJSON            bool      `json:"use_json" db:"use_json"`
	NeedProxy          bool      `json:"need_proxy" db:"need_proxy"`
	ReturnsYID         bool      `json:"returns_yid" db:"returns_yid"`
	ExtraParams        bool      `json:"extra_params" db:"extra_params"`
	QueryAct           string    `json:"query_act" db:"query_act"`
	QueryPath          string    `json:"query_path" db:"query_path"`
	QueryMethod        string    `json:"query_method" db:"query_method"`
	QueryBodyType      string    `json:"query_body_type" db:"query_body_type"`
	QueryParamStyle    string    `json:"query_param_style" db:"query_param_style"`
	QueryParamMap      string    `json:"query_param_map" db:"query_param_map"`
	QueryPolling       bool      `json:"query_polling" db:"query_polling"`
	QueryMaxAttempts   int       `json:"query_max_attempts" db:"query_max_attempts"`
	QueryInterval      int       `json:"query_interval" db:"query_interval"`
	QueryResponseMap   string    `json:"query_response_map" db:"query_response_map"`
	OrderPath          string    `json:"order_path" db:"order_path"`
	OrderMethod        string    `json:"order_method" db:"order_method"`
	OrderBodyType      string    `json:"order_body_type" db:"order_body_type"`
	OrderParamMap      string    `json:"order_param_map" db:"order_param_map"`
	YIDInDataArray     bool      `json:"yid_in_data_array" db:"yid_in_data_array"`
	ProgressPath       string    `json:"progress_path" db:"progress_path"`
	ProgressMethod     string    `json:"progress_method" db:"progress_method"`
	ProgressBodyType   string    `json:"progress_body_type" db:"progress_body_type"`
	ProgressParamMap   string    `json:"progress_param_map" db:"progress_param_map"`
	CategoryPath       string    `json:"category_path" db:"category_path"`
	CategoryMethod     string    `json:"category_method" db:"category_method"`
	CategoryBodyType   string    `json:"category_body_type" db:"category_body_type"`
	CategoryParamMap   string    `json:"category_param_map" db:"category_param_map"`
	ClassListPath      string    `json:"class_list_path" db:"class_list_path"`
	ClassListMethod    string    `json:"class_list_method" db:"class_list_method"`
	ClassListBodyType  string    `json:"class_list_body_type" db:"class_list_body_type"`
	ClassListParamMap  string    `json:"class_list_param_map" db:"class_list_param_map"`
	PausePath          string    `json:"pause_path" db:"pause_path"`
	PauseMethod        string    `json:"pause_method" db:"pause_method"`
	PauseBodyType      string    `json:"pause_body_type" db:"pause_body_type"`
	PauseParamMap      string    `json:"pause_param_map" db:"pause_param_map"`
	PauseIDParam       string    `json:"pause_id_param" db:"pause_id_param"`
	ResumePath         string    `json:"resume_path" db:"resume_path"`
	ResumeMethod       string    `json:"resume_method" db:"resume_method"`
	ResumeBodyType     string    `json:"resume_body_type" db:"resume_body_type"`
	ResumeParamMap     string    `json:"resume_param_map" db:"resume_param_map"`
	ChangePassPath     string    `json:"change_pass_path" db:"change_pass_path"`
	ChangePassMethod   string    `json:"change_pass_method" db:"change_pass_method"`
	ChangePassBodyType string    `json:"change_pass_body_type" db:"change_pass_body_type"`
	ChangePassParamMap string    `json:"change_pass_param_map" db:"change_pass_param_map"`
	ChangePassParam    string    `json:"change_pass_param" db:"change_pass_param"`
	ChangePassIDParam  string    `json:"change_pass_id_param" db:"change_pass_id_param"`
	ResubmitPath       string    `json:"resubmit_path" db:"resubmit_path"`
	ResubmitMethod     string    `json:"resubmit_method" db:"resubmit_method"`
	ResubmitBodyType   string    `json:"resubmit_body_type" db:"resubmit_body_type"`
	ResubmitParamMap   string    `json:"resubmit_param_map" db:"resubmit_param_map"`
	ResubmitIDParam    string    `json:"resubmit_id_param" db:"resubmit_id_param"`
	LogPath            string    `json:"log_path" db:"log_path"`
	LogMethod          string    `json:"log_method" db:"log_method"`
	LogBodyType        string    `json:"log_body_type" db:"log_body_type"`
	LogParamMap        string    `json:"log_param_map" db:"log_param_map"`
	LogIDParam         string    `json:"log_id_param" db:"log_id_param"`
	BalancePath        string    `json:"balance_path" db:"balance_path"`
	BalanceMoneyField  string    `json:"balance_money_field" db:"balance_money_field"`
	BalanceMethod      string    `json:"balance_method" db:"balance_method"`
	BalanceBodyType    string    `json:"balance_body_type" db:"balance_body_type"`
	BalanceParamMap    string    `json:"balance_param_map" db:"balance_param_map"`
	BalanceAuthType    string    `json:"balance_auth_type" db:"balance_auth_type"`
	ReportParamStyle   string    `json:"report_param_style" db:"report_param_style"`
	ReportAuthType     string    `json:"report_auth_type" db:"report_auth_type"`
	ReportMethod       string    `json:"report_method" db:"report_method"`
	ReportBodyType     string    `json:"report_body_type" db:"report_body_type"`
	ReportParamMap     string    `json:"report_param_map" db:"report_param_map"`
	ReportPath         string    `json:"report_path" db:"report_path"`
	GetReportMethod    string    `json:"get_report_method" db:"get_report_method"`
	GetReportBodyType  string    `json:"get_report_body_type" db:"get_report_body_type"`
	GetReportParamMap  string    `json:"get_report_param_map" db:"get_report_param_map"`
	GetReportPath      string    `json:"get_report_path" db:"get_report_path"`
	RefreshPath        string    `json:"refresh_path" db:"refresh_path"`
	SourceCode         string    `json:"source_code" db:"source_code"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

// PlatformConfigSaveRequest 保存平台配置请求
type PlatformConfigSaveRequest struct {
	ID                 int    `json:"id"`
	PT                 string `json:"pt" binding:"required"`
	Name               string `json:"name"`
	AuthType           string `json:"auth_type"`
	SuccessCodes       string `json:"success_codes"`
	UseJSON            bool   `json:"use_json"`
	NeedProxy          bool   `json:"need_proxy"`
	ReturnsYID         bool   `json:"returns_yid"`
	ExtraParams        bool   `json:"extra_params"`
	QueryAct           string `json:"query_act"`
	QueryPath          string `json:"query_path"`
	QueryMethod        string `json:"query_method"`
	QueryBodyType      string `json:"query_body_type"`
	QueryParamStyle    string `json:"query_param_style"`
	QueryParamMap      string `json:"query_param_map"`
	QueryPolling       bool   `json:"query_polling"`
	QueryMaxAttempts   int    `json:"query_max_attempts"`
	QueryInterval      int    `json:"query_interval"`
	QueryResponseMap   string `json:"query_response_map"`
	OrderPath          string `json:"order_path"`
	OrderMethod        string `json:"order_method"`
	OrderBodyType      string `json:"order_body_type"`
	OrderParamMap      string `json:"order_param_map"`
	YIDInDataArray     bool   `json:"yid_in_data_array"`
	ProgressPath       string `json:"progress_path"`
	ProgressMethod     string `json:"progress_method"`
	ProgressBodyType   string `json:"progress_body_type"`
	ProgressParamMap   string `json:"progress_param_map"`
	CategoryPath       string `json:"category_path"`
	CategoryMethod     string `json:"category_method"`
	CategoryBodyType   string `json:"category_body_type"`
	CategoryParamMap   string `json:"category_param_map"`
	ClassListPath      string `json:"class_list_path"`
	ClassListMethod    string `json:"class_list_method"`
	ClassListBodyType  string `json:"class_list_body_type"`
	ClassListParamMap  string `json:"class_list_param_map"`
	PausePath          string `json:"pause_path"`
	PauseMethod        string `json:"pause_method"`
	PauseBodyType      string `json:"pause_body_type"`
	PauseParamMap      string `json:"pause_param_map"`
	PauseIDParam       string `json:"pause_id_param"`
	ResumePath         string `json:"resume_path"`
	ResumeMethod       string `json:"resume_method"`
	ResumeBodyType     string `json:"resume_body_type"`
	ResumeParamMap     string `json:"resume_param_map"`
	ChangePassPath     string `json:"change_pass_path"`
	ChangePassMethod   string `json:"change_pass_method"`
	ChangePassBodyType string `json:"change_pass_body_type"`
	ChangePassParamMap string `json:"change_pass_param_map"`
	ChangePassParam    string `json:"change_pass_param"`
	ChangePassIDParam  string `json:"change_pass_id_param"`
	ResubmitPath       string `json:"resubmit_path"`
	ResubmitMethod     string `json:"resubmit_method"`
	ResubmitBodyType   string `json:"resubmit_body_type"`
	ResubmitParamMap   string `json:"resubmit_param_map"`
	ResubmitIDParam    string `json:"resubmit_id_param"`
	LogPath            string `json:"log_path"`
	LogMethod          string `json:"log_method"`
	LogBodyType        string `json:"log_body_type"`
	LogParamMap        string `json:"log_param_map"`
	LogIDParam         string `json:"log_id_param"`
	BalancePath        string `json:"balance_path"`
	BalanceMoneyField  string `json:"balance_money_field"`
	BalanceMethod      string `json:"balance_method"`
	BalanceBodyType    string `json:"balance_body_type"`
	BalanceParamMap    string `json:"balance_param_map"`
	BalanceAuthType    string `json:"balance_auth_type"`
	ReportParamStyle   string `json:"report_param_style"`
	ReportAuthType     string `json:"report_auth_type"`
	ReportMethod       string `json:"report_method"`
	ReportBodyType     string `json:"report_body_type"`
	ReportParamMap     string `json:"report_param_map"`
	ReportPath         string `json:"report_path"`
	GetReportMethod    string `json:"get_report_method"`
	GetReportBodyType  string `json:"get_report_body_type"`
	GetReportParamMap  string `json:"get_report_param_map"`
	GetReportPath      string `json:"get_report_path"`
	RefreshPath        string `json:"refresh_path"`
	SourceCode         string `json:"source_code"`
}
