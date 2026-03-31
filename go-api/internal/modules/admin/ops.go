package admin

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"go-api/internal/database"
	"go-api/internal/license"
	"go-api/internal/model"
	auxmodule "go-api/internal/modules/auxiliary"
	supplier "go-api/internal/modules/supplier"
	tenant "go-api/internal/modules/tenant"
	papermodule "go-api/internal/plugins/paper"
	sxdkmodule "go-api/internal/plugins/sxdk"
	tuboshumodule "go-api/internal/plugins/tuboshu"
	tutuqgmodule "go-api/internal/plugins/tutuqg"
	tuzhimodule "go-api/internal/plugins/tuzhi"
	yfdkmodule "go-api/internal/plugins/yfdk"
	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

func getAdminConfig() (map[string]string, error) {
	rows, err := database.DB.Query("SELECT `v`, `k` FROM qingka_wangke_config")
	if err != nil {
		return map[string]string{}, nil
	}
	defer rows.Close()

	config := make(map[string]string)
	for rows.Next() {
		var k, v string
		rows.Scan(&k, &v)
		config[k] = v
	}
	return config, nil
}

func saveAdminConfig(configs map[string]string) error {
	if len(configs) == 0 {
		return nil
	}

	stmt := "REPLACE INTO qingka_wangke_config (`v`, `k`) VALUES "
	placeholders := make([]string, 0, len(configs))
	params := make([]interface{}, 0, len(configs)*2)
	for k, v := range configs {
		placeholders = append(placeholders, "(?, ?)")
		params = append(params, k, v)
	}
	stmt += strings.Join(placeholders, ", ")

	_, err := database.DB.Exec(stmt, params...)
	return err
}

func getAdminPayData() (map[string]string, error) {
	var paydata string
	err := database.DB.QueryRow("SELECT COALESCE(paydata,'') FROM qingka_wangke_user WHERE uid = 1").Scan(&paydata)
	if err != nil {
		return map[string]string{}, nil
	}
	result := make(map[string]string)
	if paydata != "" {
		json.Unmarshal([]byte(paydata), &result)
	}
	return result, nil
}

func saveAdminPayData(data map[string]string) error {
	existing, _ := getAdminPayData()
	for k, v := range data {
		existing[k] = v
	}
	jsonBytes, err := json.Marshal(existing)
	if err != nil {
		return err
	}
	_, err = database.DB.Exec("UPDATE qingka_wangke_user SET paydata = ? WHERE uid = 1", string(jsonBytes))
	return err
}

func listAdminTenants(page, limit int) ([]model.Tenant, int, error) {
	offset := (page - 1) * limit
	var total int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_tenant").Scan(&total)
	rows, err := database.DB.Query(
		"SELECT tid,uid,shop_name,COALESCE(shop_logo,''),COALESCE(shop_desc,''),COALESCE(domain,''),status,addtime FROM qingka_tenant ORDER BY tid DESC LIMIT ? OFFSET ?",
		limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []model.Tenant
	for rows.Next() {
		var t model.Tenant
		rows.Scan(&t.TID, &t.UID, &t.ShopName, &t.ShopLogo, &t.ShopDesc, &t.Domain, &t.Status, &t.AddTime)
		list = append(list, t)
	}
	return list, total, nil
}

func setAdminTenantStatus(tid, status int) error {
	_, err := database.DB.Exec("UPDATE qingka_tenant SET status=? WHERE tid=?", status, tid)
	return err
}

func listAdminCheckinStats(date string, page, limit int) ([]model.CheckinRecord, int64, model.CheckinDayStat, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	var stat model.CheckinDayStat
	stat.CheckinDate = date
	database.DB.QueryRow(
		"SELECT COUNT(*), COALESCE(SUM(reward_money),0) FROM qingka_wangke_checkin WHERE checkin_date = ?", date,
	).Scan(&stat.TotalUsers, &stat.TotalReward)

	var total int64
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_checkin WHERE checkin_date = ?", date).Scan(&total)

	offset := (page - 1) * limit
	rows, err := database.DB.Query(
		"SELECT id, uid, COALESCE(username,''), COALESCE(reward_money,0), checkin_date, COALESCE(DATE_FORMAT(addtime,'%Y-%m-%d %H:%i:%s'),'') FROM qingka_wangke_checkin WHERE checkin_date = ? ORDER BY id DESC LIMIT ? OFFSET ?",
		date, limit, offset,
	)
	if err != nil {
		return nil, 0, stat, err
	}
	defer rows.Close()

	var list []model.CheckinRecord
	for rows.Next() {
		var r model.CheckinRecord
		rows.Scan(&r.ID, &r.UID, &r.Username, &r.RewardMoney, &r.CheckinDate, &r.AddTime)
		list = append(list, r)
	}
	if list == nil {
		list = []model.CheckinRecord{}
	}
	return list, total, stat, nil
}

type LonglongToolConfig struct {
	LongHost      string `json:"long_host"`
	AccessKey     string `json:"access_key"`
	MysqlHost     string `json:"mysql_host"`
	MysqlPort     string `json:"mysql_port"`
	MysqlUser     string `json:"mysql_user"`
	MysqlPassword string `json:"mysql_password"`
	MysqlDatabase string `json:"mysql_database"`
	ClassTable    string `json:"class_table"`
	OrderTable    string `json:"order_table"`
	Docking       string `json:"docking"`
	Rate          string `json:"rate"`
	NamePrefix    string `json:"name_prefix"`
	Category      string `json:"category"`
	CoverPrice    bool   `json:"cover_price"`
	CoverDesc     bool   `json:"cover_desc"`
	CoverName     bool   `json:"cover_name"`
	Sort          string `json:"sort"`
	CronValue     string `json:"cron_value"`
	CronUnit      string `json:"cron_unit"`
}

func registerOpsRoutes(admin *gin.RouterGroup) {
	admin.GET("/config", AdminConfigGet)
	admin.POST("/config", AdminConfigSave)
	admin.GET("/paydata", AdminPayDataGet)
	admin.POST("/paydata", AdminPayDataSave)

	admin.GET("/platform-configs", AdminPlatformConfigList)
	admin.POST("/platform-config/save", AdminPlatformConfigSave)
	admin.DELETE("/platform-config/:pt", AdminPlatformConfigDelete)
	admin.POST("/platform-config/parse-php", AdminParsePHPCode)
	admin.POST("/platform-config/detect", AdminDetectPlatform)
	admin.POST("/platform-config/auto-detect-save", AdminAutoDetectSave)

	admin.GET("/sync/config", SyncGetConfig)
	admin.POST("/sync/config", SyncSaveConfig)
	admin.GET("/sync/preview", SyncPreview)
	admin.POST("/sync/execute", SyncExecute)
	admin.GET("/sync/logs", SyncLogs)
	admin.GET("/sync/suppliers", SyncMonitoredSuppliers)
	admin.GET("/sync/auto-status", AutoSyncStatusHandler)

	admin.GET("/longlong-tool/config", LonglongToolGetConfig)
	admin.POST("/longlong-tool/config", LonglongToolSaveConfig)
	admin.POST("/longlong-tool/sync", LonglongToolSync)
	admin.GET("/longlong-tool/status", LonglongToolStatus)
	admin.GET("/longlong-tool/cli-check", LonglongToolCheckCLI)
	admin.POST("/longlong-tool/cli-install", LonglongToolInstallCLI)

	admin.GET("/tutuqg/config", TutuQGConfigGet)
	admin.POST("/tutuqg/config", TutuQGConfigSave)
	admin.GET("/tuboshu/config", TuboshuConfigGet)
	admin.POST("/tuboshu/config", TuboshuConfigSave)
	admin.POST("/tuboshu/price-config", TuboshuSavePriceConfig)
	admin.GET("/paper/config", PaperConfigGet)
	admin.POST("/paper/config", PaperConfigSave)
	admin.GET("/yfdk/config", YFDKConfigGet)
	admin.POST("/yfdk/config", YFDKConfigSave)
	admin.GET("/yfdk/projects", YFDKAdminProjectList)
	admin.POST("/yfdk/projects/sync", YFDKAdminProjectSync)
	admin.PUT("/yfdk/projects", YFDKAdminProjectUpdate)
	admin.DELETE("/yfdk/projects/:id", YFDKAdminProjectDelete)
	admin.GET("/tuzhi/config", TuZhiConfigGet)
	admin.POST("/tuzhi/config", TuZhiConfigSave)
	admin.GET("/tuzhi/goods", TuZhiAdminGetGoods)
	admin.GET("/tuzhi/goods-overrides", TuZhiGoodsOverridesGet)
	admin.POST("/tuzhi/goods-overrides", TuZhiGoodsOverridesSave)
	admin.GET("/sxdk/config", SXDKConfigGet)
	admin.POST("/sxdk/config", SXDKConfigSave)
	admin.GET("/hzw-socket/config", HZWSocketConfigGet)
	admin.POST("/hzw-socket/config", HZWSocketConfigSave)

	admin.GET("/license/status", AdminLicenseStatus)
	admin.GET("/ops/dashboard", AdminOpsDashboard)
	admin.GET("/ops/probe-suppliers", AdminOpsProbeSuppliers)
	admin.GET("/ops/table-sizes", AdminOpsTableSizes)
	admin.GET("/ops/turbo", AdminGetTurbo)
	admin.POST("/ops/turbo", AdminSetTurbo)

	admin.GET("/tenants", AdminTenantList)
	admin.POST("/tenant/create", AdminTenantCreate)
	admin.POST("/tenant/:tid/status", AdminTenantSetStatus)
	admin.GET("/checkin/stats", AdminCheckinStats)
	admin.GET("/cardkeys", AdminCardKeyList)
	admin.POST("/cardkey/generate", AdminCardKeyGenerate)
	admin.POST("/cardkey/delete", AdminCardKeyDelete)
	admin.GET("/activities", AdminActivityList)
	admin.POST("/activity/save", AdminActivitySave)
	admin.DELETE("/activity/:hid", AdminActivityDelete)

	admin.GET("/db-compat/check", AdminDBCompatCheck)
	admin.POST("/db-compat/fix", AdminDBCompatFix)
	admin.POST("/db-sync/test", AdminDBSyncTest)
	admin.POST("/db-sync/execute", AdminDBSyncExecute)

	admin.GET("/pledge/configs", AdminPledgeConfigList)
	admin.POST("/pledge/config/save", AdminPledgeConfigSave)
	admin.DELETE("/pledge/config/:id", AdminPledgeConfigDelete)
	admin.POST("/pledge/config/toggle", AdminPledgeConfigToggle)
	admin.GET("/pledge/records", AdminPledgeRecordList)
}

func AdminLicenseStatus(c *gin.Context) {
	lm := license.Global
	if lm == nil {
		response.Success(c, gin.H{"status": "未初始化", "status_code": -1})
		return
	}
	response.Success(c, lm.GetStatusInfo())
}

func AdminOpsDashboard(c *gin.Context) {
	dash := getAdminOpsDashboard()
	response.Success(c, dash)
}

func AdminOpsProbeSuppliers(c *gin.Context) {
	probes := probeAdminSuppliers()
	response.Success(c, probes)
}

func AdminOpsTableSizes(c *gin.Context) {
	tables := getAdminTableSizes()
	response.Success(c, tables)
}

func AdminGetTurbo(c *gin.Context) {
	status := getAdminTurboStatus()
	response.Success(c, status)
}

func AdminSetTurbo(c *gin.Context) {
	var req struct {
		Mode string `json:"mode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	valid := map[string]bool{"eco": true, "normal": true, "turbo": true, "insane": true, "auto": true}
	if !valid[req.Mode] {
		response.BadRequest(c, "无效模式，可选: eco/normal/turbo/insane/auto")
		return
	}

	status := applyAdminTurbo(req.Mode)
	response.Success(c, status)
}

func AdminConfigGet(c *gin.Context) {
	config, err := getAdminConfig()
	if err != nil {
		response.ServerErrorf(c, err, "查询设置失败")
		return
	}
	response.Success(c, config)
}

func AdminConfigSave(c *gin.Context) {
	var configs map[string]string
	if err := c.ShouldBindJSON(&configs); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := saveAdminConfig(configs); err != nil {
		response.ServerErrorf(c, err, "保存设置失败")
		return
	}
	response.SuccessMsg(c, "保存成功")
}

func AdminPayDataGet(c *gin.Context) {
	data, err := getAdminPayData()
	if err != nil {
		response.ServerErrorf(c, err, "查询支付配置失败")
		return
	}
	response.Success(c, data)
}

func AdminPayDataSave(c *gin.Context) {
	var data map[string]string
	if err := c.ShouldBindJSON(&data); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := saveAdminPayData(data); err != nil {
		response.ServerErrorf(c, err, "保存支付配置失败")
		return
	}
	response.SuccessMsg(c, "保存成功")
}

func AdminPlatformConfigList(c *gin.Context) {
	rows, err := database.DB.Query(`SELECT id, pt, name, auth_type, success_codes,
		use_json, need_proxy, returns_yid, extra_params,
		query_act, COALESCE(query_path,''), COALESCE(query_method,''), COALESCE(query_body_type,''), COALESCE(query_param_style,''), COALESCE(query_param_map,''), query_polling, query_max_attempts, query_interval, COALESCE(query_response_map,''),
		COALESCE(order_path,''), COALESCE(order_method,''), COALESCE(order_body_type,''), COALESCE(order_param_map,''), yid_in_data_array,
		COALESCE(progress_path,''), COALESCE(progress_method,''),
		COALESCE(progress_body_type,''), COALESCE(progress_param_map,''),
		COALESCE(batch_progress_path,''), COALESCE(batch_progress_method,''),
		COALESCE(batch_progress_body_type,''), COALESCE(batch_progress_param_map,''),
		COALESCE(category_path,''), COALESCE(category_method,''), COALESCE(category_body_type,''), COALESCE(category_param_map,''),
		COALESCE(class_list_path,''), COALESCE(class_list_method,''), COALESCE(class_list_body_type,''), COALESCE(class_list_param_map,''),
		COALESCE(pause_path,''), COALESCE(pause_method,''), COALESCE(pause_body_type,''), COALESCE(pause_param_map,''), COALESCE(pause_id_param,''),
		COALESCE(resume_path,''), COALESCE(resume_method,''), COALESCE(resume_body_type,''), COALESCE(resume_param_map,''),
		COALESCE(change_pass_path,''), COALESCE(change_pass_method,''), COALESCE(change_pass_body_type,''), COALESCE(change_pass_param_map,''), COALESCE(change_pass_param,''), COALESCE(change_pass_id_param,''),
		COALESCE(resubmit_path,''), COALESCE(resubmit_method,''), COALESCE(resubmit_body_type,''), COALESCE(resubmit_param_map,''), COALESCE(resubmit_id_param,''),
		COALESCE(log_path,''), COALESCE(log_method,''), COALESCE(log_body_type,''), COALESCE(log_param_map,''), COALESCE(log_id_param,''),
		COALESCE(balance_path,''), COALESCE(balance_money_field,''),
		COALESCE(balance_method,''), COALESCE(balance_body_type,''), COALESCE(balance_param_map,''), COALESCE(balance_auth_type,''),
		COALESCE(report_param_style,''), COALESCE(report_auth_type,''), COALESCE(report_path,''), COALESCE(report_method,''), COALESCE(report_body_type,''), COALESCE(report_param_map,''),
		COALESCE(get_report_path,''), COALESCE(get_report_method,''), COALESCE(get_report_body_type,''), COALESCE(get_report_param_map,''), COALESCE(refresh_path,''),
		COALESCE(source_code,''), created_at, updated_at
		FROM qingka_platform_config ORDER BY pt`)
	if err != nil {
		response.ServerErrorf(c, err, "查询失败: "+err.Error())
		return
	}
	defer rows.Close()

	var list []model.PlatformConfigDB
	for rows.Next() {
		var cfg model.PlatformConfigDB
		err := rows.Scan(
			&cfg.ID, &cfg.PT, &cfg.Name, &cfg.AuthType, &cfg.SuccessCodes,
			&cfg.UseJSON, &cfg.NeedProxy, &cfg.ReturnsYID, &cfg.ExtraParams,
			&cfg.QueryAct, &cfg.QueryPath, &cfg.QueryMethod, &cfg.QueryBodyType, &cfg.QueryParamStyle, &cfg.QueryParamMap, &cfg.QueryPolling, &cfg.QueryMaxAttempts, &cfg.QueryInterval, &cfg.QueryResponseMap,
			&cfg.OrderPath, &cfg.OrderMethod, &cfg.OrderBodyType, &cfg.OrderParamMap, &cfg.YIDInDataArray,
			&cfg.ProgressPath, &cfg.ProgressMethod, &cfg.ProgressBodyType, &cfg.ProgressParamMap,
			&cfg.BatchProgressPath, &cfg.BatchProgressMethod, &cfg.BatchProgressBodyType, &cfg.BatchProgressParamMap,
			&cfg.CategoryPath, &cfg.CategoryMethod, &cfg.CategoryBodyType, &cfg.CategoryParamMap,
			&cfg.ClassListPath, &cfg.ClassListMethod, &cfg.ClassListBodyType, &cfg.ClassListParamMap,
			&cfg.PausePath, &cfg.PauseMethod, &cfg.PauseBodyType, &cfg.PauseParamMap, &cfg.PauseIDParam,
			&cfg.ResumePath, &cfg.ResumeMethod, &cfg.ResumeBodyType, &cfg.ResumeParamMap,
			&cfg.ChangePassPath, &cfg.ChangePassMethod, &cfg.ChangePassBodyType, &cfg.ChangePassParamMap, &cfg.ChangePassParam, &cfg.ChangePassIDParam,
			&cfg.ResubmitPath, &cfg.ResubmitMethod, &cfg.ResubmitBodyType, &cfg.ResubmitParamMap, &cfg.ResubmitIDParam,
			&cfg.LogPath, &cfg.LogMethod, &cfg.LogBodyType, &cfg.LogParamMap, &cfg.LogIDParam,
			&cfg.BalancePath, &cfg.BalanceMoneyField, &cfg.BalanceMethod, &cfg.BalanceBodyType, &cfg.BalanceParamMap, &cfg.BalanceAuthType,
			&cfg.ReportParamStyle, &cfg.ReportAuthType, &cfg.ReportPath, &cfg.ReportMethod, &cfg.ReportBodyType, &cfg.ReportParamMap, &cfg.GetReportPath, &cfg.GetReportMethod, &cfg.GetReportBodyType, &cfg.GetReportParamMap, &cfg.RefreshPath,
			&cfg.SourceCode, &cfg.CreatedAt, &cfg.UpdatedAt,
		)
		if err != nil {
			response.ServerErrorf(c, err, "解析数据失败: "+err.Error())
			return
		}
		canonicalizePlatformConfigDB(&cfg)
		list = append(list, cfg)
	}
	if list == nil {
		list = []model.PlatformConfigDB{}
	}
	response.Success(c, list)
}

func AdminPlatformConfigSave(c *gin.Context) {
	var req model.PlatformConfigSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	normalizePlatformConfigSaveRequest(&req)
	if msg := validatePlatformConfigSaveRequest(&req); msg != "" {
		response.BadRequest(c, msg)
		return
	}
	err := upsertAdminPlatformConfig(req)
	if err != nil {
		response.ServerErrorf(c, err, "保存失败: "+err.Error())
		return
	}

	supplier.ReloadPlatformConfigs()
	response.SuccessMsg(c, "保存成功")
}

func isCustomQueryDriverValue(driver string) bool {
	switch strings.TrimSpace(driver) {
	case "local_time", "local_script", "xxt_query", "KUN_custom", "simple_custom", "yyy_custom", "tuboshu_custom", "nx_custom", "lgwk_custom":
		return true
	default:
		return false
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func canonicalizePlatformConfigDB(cfg *model.PlatformConfigDB) {
	if !isCustomQueryDriverValue(cfg.QueryAct) {
		cfg.QueryPath = firstNonEmpty(cfg.QueryPath)
		cfg.QueryAct = ""
	}
	cfg.OrderPath = firstNonEmpty(cfg.OrderPath)
	cfg.ProgressPath = firstNonEmpty(cfg.ProgressPath)
	cfg.BatchProgressPath = firstNonEmpty(cfg.BatchProgressPath)
	cfg.CategoryPath = firstNonEmpty(cfg.CategoryPath)
	cfg.ClassListPath = firstNonEmpty(cfg.ClassListPath)
	cfg.PausePath = firstNonEmpty(cfg.PausePath)
	cfg.ResumePath = firstNonEmpty(cfg.ResumePath)
	cfg.ChangePassPath = firstNonEmpty(cfg.ChangePassPath)
	cfg.LogPath = firstNonEmpty(cfg.LogPath)
	cfg.BalancePath = firstNonEmpty(cfg.BalancePath)
}

func normalizePlatformConfigSaveRequest(req *model.PlatformConfigSaveRequest) {
	if req.AuthType == "" {
		req.AuthType = "uid_key"
	}
	if req.SuccessCodes == "" {
		req.SuccessCodes = "0"
	}
	if !isCustomQueryDriverValue(req.QueryAct) {
		req.QueryPath = firstNonEmpty(req.QueryPath)
		req.QueryAct = ""
	}
	req.QueryMethod = strings.ToUpper(strings.TrimSpace(req.QueryMethod))
	req.OrderPath = firstNonEmpty(req.OrderPath)
	req.OrderMethod = strings.ToUpper(strings.TrimSpace(req.OrderMethod))
	req.ProgressPath = firstNonEmpty(req.ProgressPath)
	req.ProgressMethod = strings.ToUpper(strings.TrimSpace(req.ProgressMethod))
	req.ProgressParamMap = strings.TrimSpace(req.ProgressParamMap)
	req.BatchProgressPath = firstNonEmpty(req.BatchProgressPath)
	req.BatchProgressMethod = strings.ToUpper(strings.TrimSpace(req.BatchProgressMethod))
	req.BatchProgressParamMap = strings.TrimSpace(req.BatchProgressParamMap)
	req.CategoryPath = firstNonEmpty(req.CategoryPath)
	req.CategoryMethod = strings.ToUpper(strings.TrimSpace(req.CategoryMethod))
	req.ClassListPath = firstNonEmpty(req.ClassListPath)
	req.ClassListMethod = strings.ToUpper(strings.TrimSpace(req.ClassListMethod))
	req.PausePath = firstNonEmpty(req.PausePath)
	req.PauseMethod = strings.ToUpper(strings.TrimSpace(req.PauseMethod))
	req.PauseIDParam = strings.TrimSpace(req.PauseIDParam)
	req.ResumePath = firstNonEmpty(req.ResumePath)
	req.ResumeMethod = strings.ToUpper(strings.TrimSpace(req.ResumeMethod))
	req.ChangePassPath = firstNonEmpty(req.ChangePassPath)
	req.ChangePassMethod = strings.ToUpper(strings.TrimSpace(req.ChangePassMethod))
	req.ChangePassParam = strings.TrimSpace(req.ChangePassParam)
	req.ChangePassIDParam = strings.TrimSpace(req.ChangePassIDParam)
	req.ResubmitMethod = strings.ToUpper(strings.TrimSpace(req.ResubmitMethod))
	req.ResubmitIDParam = strings.TrimSpace(req.ResubmitIDParam)
	req.ResubmitPath = firstNonEmpty(req.ResubmitPath)
	req.LogPath = firstNonEmpty(req.LogPath)
	req.LogMethod = strings.ToUpper(strings.TrimSpace(req.LogMethod))
	req.LogIDParam = strings.TrimSpace(req.LogIDParam)
	req.BalancePath = firstNonEmpty(req.BalancePath)
	req.BalanceMoneyField = strings.TrimSpace(req.BalanceMoneyField)
	req.BalanceMethod = strings.ToUpper(strings.TrimSpace(req.BalanceMethod))
	req.ReportParamStyle = strings.TrimSpace(req.ReportParamStyle)
	req.ReportAuthType = strings.TrimSpace(req.ReportAuthType)
	req.ReportPath = firstNonEmpty(req.ReportPath)
	req.ReportMethod = strings.ToUpper(strings.TrimSpace(req.ReportMethod))
	req.GetReportPath = firstNonEmpty(req.GetReportPath)
	req.GetReportMethod = strings.ToUpper(strings.TrimSpace(req.GetReportMethod))
}

func validatePlatformConfigSaveRequest(req *model.PlatformConfigSaveRequest) string {
	validateActionJSON := func(label, path, method, paramMap string, required bool) string {
		configured := strings.TrimSpace(path) != "" || strings.TrimSpace(method) != "" || strings.TrimSpace(paramMap) != ""
		if !configured && !required {
			return ""
		}
		if strings.TrimSpace(path) == "" {
			return "请填写" + label + "路径"
		}
		if strings.TrimSpace(method) == "" {
			return "请填写" + label + "请求方式"
		}
		if strings.TrimSpace(paramMap) == "" {
			return "请填写" + label + "参数映射JSON"
		}

		var mapped map[string]string
		if err := json.Unmarshal([]byte(paramMap), &mapped); err != nil {
			return label + "参数映射JSON格式错误: " + err.Error()
		}
		return ""
	}

	if !isCustomQueryDriverValue(req.QueryAct) {
		if msg := validateActionJSON("查课", req.QueryPath, req.QueryMethod, req.QueryParamMap, false); msg != "" {
			return msg
		}
	}

	if msg := validateActionJSON("下单", req.OrderPath, req.OrderMethod, req.OrderParamMap, false); msg != "" {
		return msg
	}
	if msg := validateActionJSON("进度", req.ProgressPath, req.ProgressMethod, req.ProgressParamMap, false); msg != "" {
		return msg
	}
	if msg := validateActionJSON("批量进度", req.BatchProgressPath, req.BatchProgressMethod, req.BatchProgressParamMap, false); msg != "" {
		return msg
	}
	if msg := validateActionJSON("分类", req.CategoryPath, req.CategoryMethod, req.CategoryParamMap, false); msg != "" {
		return msg
	}
	if msg := validateActionJSON("课程列表", req.ClassListPath, req.ClassListMethod, req.ClassListParamMap, false); msg != "" {
		return msg
	}
	if msg := validateActionJSON("暂停", req.PausePath, req.PauseMethod, req.PauseParamMap, false); msg != "" {
		return msg
	}
	if msg := validateActionJSON("恢复", req.ResumePath, req.ResumeMethod, req.ResumeParamMap, false); msg != "" {
		return msg
	}
	if msg := validateActionJSON("改密", req.ChangePassPath, req.ChangePassMethod, req.ChangePassParamMap, false); msg != "" {
		return msg
	}
	if msg := validateActionJSON("补单", req.ResubmitPath, req.ResubmitMethod, req.ResubmitParamMap, false); msg != "" {
		return msg
	}
	if msg := validateActionJSON("日志", req.LogPath, req.LogMethod, req.LogParamMap, false); msg != "" {
		return msg
	}
	if msg := validateActionJSON("余额", req.BalancePath, req.BalanceMethod, req.BalanceParamMap, false); msg != "" {
		return msg
	}
	if strings.TrimSpace(req.BalancePath) != "" && strings.TrimSpace(req.BalanceMoneyField) == "" {
		return "请填写余额金额字段路径"
	}
	if msg := validateActionJSON("提交工单", req.ReportPath, req.ReportMethod, req.ReportParamMap, false); msg != "" {
		return msg
	}
	if msg := validateActionJSON("查询工单", req.GetReportPath, req.GetReportMethod, req.GetReportParamMap, false); msg != "" {
		return msg
	}
	return ""
}

func upsertAdminPlatformConfig(req model.PlatformConfigSaveRequest) error {
	args := []interface{}{
		req.PT, req.Name, req.AuthType, req.SuccessCodes,
		req.UseJSON, req.NeedProxy, req.ReturnsYID, req.ExtraParams,
		req.QueryAct, req.QueryPath, req.QueryMethod, req.QueryBodyType, req.QueryParamStyle, req.QueryParamMap, req.QueryPolling, req.QueryMaxAttempts, req.QueryInterval, req.QueryResponseMap,
		req.OrderPath, req.OrderMethod, req.OrderBodyType, req.OrderParamMap, req.YIDInDataArray,
		req.ProgressPath, req.ProgressMethod, req.ProgressBodyType, req.ProgressParamMap,
		req.BatchProgressPath, req.BatchProgressMethod, req.BatchProgressBodyType, req.BatchProgressParamMap,
		req.CategoryPath, req.CategoryMethod, req.CategoryBodyType, req.CategoryParamMap,
		req.ClassListPath, req.ClassListMethod, req.ClassListBodyType, req.ClassListParamMap,
		req.PausePath, req.PauseMethod, req.PauseBodyType, req.PauseParamMap, req.PauseIDParam,
		req.ResumePath, req.ResumeMethod, req.ResumeBodyType, req.ResumeParamMap,
		req.ChangePassPath, req.ChangePassMethod, req.ChangePassBodyType, req.ChangePassParamMap, req.ChangePassParam, req.ChangePassIDParam,
		req.ResubmitPath, req.ResubmitMethod, req.ResubmitBodyType, req.ResubmitParamMap, req.ResubmitIDParam,
		req.LogPath, req.LogMethod, req.LogBodyType, req.LogParamMap, req.LogIDParam,
		req.BalancePath, req.BalanceMoneyField, req.BalanceMethod, req.BalanceBodyType, req.BalanceParamMap, req.BalanceAuthType,
		req.ReportParamStyle, req.ReportAuthType, req.ReportPath, req.ReportMethod, req.ReportBodyType, req.ReportParamMap,
		req.GetReportPath, req.GetReportMethod, req.GetReportBodyType, req.GetReportParamMap, req.RefreshPath,
		req.SourceCode,
	}

	query := `INSERT INTO qingka_platform_config (
		pt, name, auth_type, success_codes,
		use_json, need_proxy, returns_yid, extra_params,
		query_act, query_path, query_method, query_body_type, query_param_style, query_param_map, query_polling, query_max_attempts, query_interval, query_response_map,
		order_path, order_method, order_body_type, order_param_map, yid_in_data_array,
		progress_path, progress_method, progress_body_type, progress_param_map,
		batch_progress_path, batch_progress_method, batch_progress_body_type, batch_progress_param_map,
		category_path, category_method, category_body_type, category_param_map,
		class_list_path, class_list_method, class_list_body_type, class_list_param_map,
		pause_path, pause_method, pause_body_type, pause_param_map, pause_id_param,
		resume_path, resume_method, resume_body_type, resume_param_map,
		change_pass_path, change_pass_method, change_pass_body_type, change_pass_param_map, change_pass_param, change_pass_id_param,
		resubmit_path, resubmit_method, resubmit_body_type, resubmit_param_map, resubmit_id_param,
		log_path, log_method, log_body_type, log_param_map, log_id_param,
		balance_path, balance_money_field, balance_method, balance_body_type, balance_param_map, balance_auth_type,
		report_param_style, report_auth_type, report_path, report_method, report_body_type, report_param_map,
		get_report_path, get_report_method, get_report_body_type, get_report_param_map, refresh_path,
		source_code
	) VALUES (` + strings.TrimRight(strings.Repeat("?,", len(args)), ",") + `)
	ON DUPLICATE KEY UPDATE
		name=VALUES(name), auth_type=VALUES(auth_type),
		success_codes=VALUES(success_codes), use_json=VALUES(use_json), need_proxy=VALUES(need_proxy),
		returns_yid=VALUES(returns_yid), extra_params=VALUES(extra_params),
		query_act=VALUES(query_act), query_path=VALUES(query_path), query_method=VALUES(query_method),
		query_body_type=VALUES(query_body_type), query_param_style=VALUES(query_param_style), query_param_map=VALUES(query_param_map),
		query_polling=VALUES(query_polling), query_max_attempts=VALUES(query_max_attempts),
		query_interval=VALUES(query_interval), query_response_map=VALUES(query_response_map),
		order_path=VALUES(order_path), order_method=VALUES(order_method),
		order_body_type=VALUES(order_body_type), order_param_map=VALUES(order_param_map), yid_in_data_array=VALUES(yid_in_data_array),
		progress_path=VALUES(progress_path), progress_method=VALUES(progress_method),
		progress_body_type=VALUES(progress_body_type), progress_param_map=VALUES(progress_param_map),
		batch_progress_path=VALUES(batch_progress_path), batch_progress_method=VALUES(batch_progress_method),
		batch_progress_body_type=VALUES(batch_progress_body_type), batch_progress_param_map=VALUES(batch_progress_param_map),
		category_path=VALUES(category_path), category_method=VALUES(category_method),
		category_body_type=VALUES(category_body_type), category_param_map=VALUES(category_param_map),
		class_list_path=VALUES(class_list_path), class_list_method=VALUES(class_list_method),
		class_list_body_type=VALUES(class_list_body_type), class_list_param_map=VALUES(class_list_param_map),
		pause_path=VALUES(pause_path), pause_method=VALUES(pause_method),
		pause_body_type=VALUES(pause_body_type), pause_param_map=VALUES(pause_param_map), pause_id_param=VALUES(pause_id_param),
		resume_path=VALUES(resume_path), resume_method=VALUES(resume_method),
		resume_body_type=VALUES(resume_body_type), resume_param_map=VALUES(resume_param_map),
		change_pass_path=VALUES(change_pass_path), change_pass_method=VALUES(change_pass_method),
		change_pass_body_type=VALUES(change_pass_body_type), change_pass_param_map=VALUES(change_pass_param_map),
		change_pass_param=VALUES(change_pass_param), change_pass_id_param=VALUES(change_pass_id_param),
		resubmit_path=VALUES(resubmit_path), resubmit_method=VALUES(resubmit_method), resubmit_body_type=VALUES(resubmit_body_type),
		resubmit_param_map=VALUES(resubmit_param_map), resubmit_id_param=VALUES(resubmit_id_param),
		log_path=VALUES(log_path), log_method=VALUES(log_method),
		log_body_type=VALUES(log_body_type), log_param_map=VALUES(log_param_map), log_id_param=VALUES(log_id_param),
		balance_path=VALUES(balance_path), balance_money_field=VALUES(balance_money_field), balance_method=VALUES(balance_method),
		balance_body_type=VALUES(balance_body_type), balance_param_map=VALUES(balance_param_map),
		balance_auth_type=VALUES(balance_auth_type),
		report_param_style=VALUES(report_param_style), report_auth_type=VALUES(report_auth_type),
		report_path=VALUES(report_path), report_method=VALUES(report_method), report_body_type=VALUES(report_body_type),
		report_param_map=VALUES(report_param_map),
		get_report_path=VALUES(get_report_path), get_report_method=VALUES(get_report_method),
		get_report_body_type=VALUES(get_report_body_type), get_report_param_map=VALUES(get_report_param_map),
		refresh_path=VALUES(refresh_path), source_code=VALUES(source_code)`

	_, err := database.DB.Exec(query, args...)
	return err
}

func AdminPlatformConfigDelete(c *gin.Context) {
	pt := c.Param("pt")
	if pt == "" {
		response.BadRequest(c, "缺少平台标识")
		return
	}

	result, err := database.DB.Exec("DELETE FROM qingka_platform_config WHERE pt = ?", pt)
	if err != nil {
		response.ServerErrorf(c, err, "删除失败: "+err.Error())
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		response.BadRequest(c, "平台不存在")
		return
	}
	supplier.ReloadPlatformConfigs()
	response.SuccessMsg(c, "删除成功")
}

func AdminDetectPlatform(c *gin.Context) {
	var req DetectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请提供平台 URL")
		return
	}

	result := adminDetectPlatform(req)
	response.Success(c, result)
}

func AdminAutoDetectSave(c *gin.Context) {
	var req AutoDetectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请提供平台 URL 和标识")
		return
	}

	result := adminDetectPlatform(req.DetectRequest)
	if !result.Success {
		response.Success(c, gin.H{
			"success": false,
			"msg":     "未检测到可用接口，请检查 URL 和凭证",
			"detect":  result,
		})
		return
	}

	saveReq := buildAdminConfigFromDetection(result, req.PT, req.Name)
	if saveReq == nil {
		response.Success(c, gin.H{
			"success": false,
			"msg":     "检测结果无法转换为配置",
			"detect":  result,
		})
		return
	}

	saveReq.SourceCode = ""
	normalizePlatformConfigSaveRequest(saveReq)
	if msg := validatePlatformConfigSaveRequest(saveReq); msg != "" {
		response.Success(c, gin.H{
			"success": false,
			"msg":     "检测成功但还需补充配置: " + msg,
			"detect":  result,
			"config":  saveReq,
		})
		return
	}
	err := upsertAdminPlatformConfig(*saveReq)
	if err != nil {
		response.Success(c, gin.H{
			"success": false,
			"msg":     "检测成功但保存失败: " + err.Error(),
			"detect":  result,
			"config":  saveReq,
		})
		return
	}

	supplier.ReloadPlatformConfigs()
	response.Success(c, gin.H{
		"success": true,
		"msg":     "检测成功并已保存配置",
		"detect":  result,
		"config":  saveReq,
	})
}

func AdminParsePHPCode(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请提供 PHP 代码")
		return
	}

	result := parseAdminPHPCode(req.Code)
	response.Success(c, result)
}

func SyncGetConfig(c *gin.Context) {
	cfg, err := getAdminSyncConfig()
	if err != nil {
		response.ServerErrorf(c, err, "获取配置失败")
		return
	}
	response.Success(c, cfg)
}

func SyncSaveConfig(c *gin.Context) {
	var cfg SyncConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := saveAdminSyncConfig(&cfg); err != nil {
		response.ServerErrorf(c, err, "保存失败: "+err.Error())
		return
	}
	response.SuccessMsg(c, "保存成功")
}

func SyncPreview(c *gin.Context) {
	hid, _ := strconv.Atoi(c.Query("hid"))
	if hid <= 0 {
		response.BadRequest(c, "请指定货源")
		return
	}
	result, err := adminSyncPreview(hid)
	if err != nil {
		response.ServerErrorf(c, err, err.Error())
		return
	}
	response.Success(c, result)
}

func SyncExecute(c *gin.Context) {
	var req struct {
		HID int `json:"hid"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.HID <= 0 {
		response.BadRequest(c, "请指定货源")
		return
	}
	result, err := adminSyncExecute(req.HID)
	if err != nil {
		response.ServerErrorf(c, err, err.Error())
		return
	}
	response.Success(c, result)
}

func SyncLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	supplierID, _ := strconv.Atoi(c.Query("supplier_id"))
	action := c.Query("action")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 50
	}

	list, total, err := getAdminSyncLogs(page, pageSize, supplierID, action)
	if err != nil {
		response.ServerErrorf(c, err, "查询日志失败")
		return
	}
	response.Success(c, gin.H{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func SyncMonitoredSuppliers(c *gin.Context) {
	cfg, _ := getAdminSyncConfig()
	if cfg.SupplierIDs == "" {
		response.Success(c, []interface{}{})
		return
	}
	list, err := getAdminMonitoredSuppliers(cfg.SupplierIDs)
	if err != nil {
		response.ServerErrorf(c, err, "查询失败")
		return
	}
	response.Success(c, list)
}

func AutoSyncStatusHandler(c *gin.Context) {
	response.Success(c, getAdminAutoSyncStatus())
}

func LonglongToolGetConfig(c *gin.Context) {
	var raw string
	err := database.DB.QueryRow("SELECT `k` FROM qingka_wangke_config WHERE `v` = 'longlong_tool_config'").Scan(&raw)
	if err != nil || raw == "" {
		response.Success(c, LonglongToolConfig{
			MysqlHost: "127.0.0.1",
			MysqlPort: "3306",
			Rate:      "1.5",
			Sort:      "0",
			CronValue: "30",
			CronUnit:  "minute",
		})
		return
	}

	var cfg LonglongToolConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		response.Success(c, LonglongToolConfig{
			MysqlHost: "127.0.0.1",
			MysqlPort: "3306",
			Rate:      "1.5",
			Sort:      "0",
			CronValue: "30",
			CronUnit:  "minute",
		})
		return
	}

	if cfg.CronValue == "" {
		cfg.CronValue = "30"
	}
	if cfg.CronUnit == "" {
		cfg.CronUnit = "minute"
	}

	response.Success(c, cfg)
}

func LonglongToolSaveConfig(c *gin.Context) {
	var cfg LonglongToolConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	data, _ := json.Marshal(cfg)
	_, err := database.DB.Exec(
		"INSERT INTO qingka_wangke_config (`v`, `k`) VALUES ('longlong_tool_config', ?) ON DUPLICATE KEY UPDATE `k` = ?",
		string(data), string(data),
	)
	if err != nil {
		response.ServerErrorf(c, err, "保存失败: "+err.Error())
		return
	}
	response.SuccessMsg(c, "保存成功")
}

func LonglongToolSync(c *gin.Context) {
	msg, err := runAdminLonglongSyncOnce()
	if err != nil {
		response.BusinessError(c, -1, "同步失败: "+err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

func LonglongToolStatus(c *gin.Context) {
	response.Success(c, getAdminLonglongStatus())
}

func LonglongToolCheckCLI(c *gin.Context) {
	response.Success(c, getAdminLonglongCLIStatus())
}

func LonglongToolInstallCLI(c *gin.Context) {
	msg, err := installAdminLonglongCLI()
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

func HZWSocketConfigGet(c *gin.Context) {
	response.Success(c, gin.H{
		"socket_url": getAdminHZWSocketURL(),
	})
}

func HZWSocketConfigSave(c *gin.Context) {
	var req struct {
		SocketURL string `json:"socket_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := setAdminHZWSocketURL(req.SocketURL); err != nil {
		response.ServerErrorf(c, err, "保存失败")
		return
	}
	restartAdminHZWSocket()
	response.SuccessMsg(c, "HZW Socket 配置已保存，客户端已重启")
}

func TutuQGConfigGet(c *gin.Context) {
	cfg, err := tutuqgmodule.TutuQG().GetConfig()
	if err != nil {
		response.ServerErrorf(c, err, "获取配置失败")
		return
	}
	response.Success(c, cfg)
}

func TutuQGConfigSave(c *gin.Context) {
	var cfg tutuqgmodule.TutuQGConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := tutuqgmodule.TutuQG().SaveConfig(&cfg); err != nil {
		response.ServerErrorf(c, err, "保存配置失败")
		return
	}
	response.SuccessMsg(c, "保存成功")
}

func TuboshuConfigGet(c *gin.Context) {
	cfg, err := tuboshumodule.Tuboshu().GetConfig()
	if err != nil {
		response.ServerErrorf(c, err, "获取配置失败")
		return
	}
	response.Success(c, cfg)
}

func TuboshuConfigSave(c *gin.Context) {
	var cfg tuboshumodule.TuboshuConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := tuboshumodule.Tuboshu().SaveConfig(&cfg); err != nil {
		response.ServerErrorf(c, err, "保存配置失败")
		return
	}
	response.SuccessMsg(c, "保存成功")
}

func TuboshuSavePriceConfig(c *gin.Context) {
	raw, err := c.GetRawData()
	if err != nil {
		response.BadRequest(c, "读取请求体失败")
		return
	}

	var body map[string]interface{}
	if err := json.Unmarshal(raw, &body); err != nil {
		response.BadRequest(c, "参数格式错误")
		return
	}

	priceConfig := body
	if config, ok := body["priceConfig"].(map[string]interface{}); ok {
		priceConfig = config
	}

	if err := tuboshumodule.Tuboshu().SavePriceConfig(priceConfig); err != nil {
		response.ServerErrorf(c, err, "保存失败: "+err.Error())
		return
	}
	response.SuccessMsg(c, "保存成功")
}

func PaperConfigGet(c *gin.Context) {
	conf, err := papermodule.Paper().GetConfig()
	if err != nil {
		response.ServerErrorf(c, err, "获取配置失败")
		return
	}
	response.Success(c, conf)
}

func PaperConfigSave(c *gin.Context) {
	var data map[string]string
	if err := c.ShouldBindJSON(&data); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := papermodule.Paper().SaveConfig(data); err != nil {
		response.ServerErrorf(c, err, "保存失败")
		return
	}
	response.SuccessMsg(c, "保存成功")
}

func YFDKConfigGet(c *gin.Context) {
	cfg, err := yfdkmodule.YFDK().GetConfig()
	if err != nil {
		response.ServerErrorf(c, err, "获取配置失败")
		return
	}
	response.Success(c, cfg)
}

func YFDKConfigSave(c *gin.Context) {
	var cfg yfdkmodule.YFDKConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := yfdkmodule.YFDK().SaveConfig(&cfg); err != nil {
		response.ServerErrorf(c, err, "保存配置失败")
		return
	}
	response.SuccessMsg(c, "保存成功")
}

func YFDKAdminProjectList(c *gin.Context) {
	projects, err := yfdkmodule.YFDK().GetAdminProjects()
	if err != nil {
		response.ServerErrorf(c, err, "获取项目列表失败")
		return
	}
	response.Success(c, projects)
}

func YFDKAdminProjectSync(c *gin.Context) {
	count, err := yfdkmodule.YFDK().SyncProjectsFromUpstream()
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, gin.H{"count": count, "msg": "成功同步 " + strconv.Itoa(count) + " 个项目"})
}

func YFDKAdminProjectUpdate(c *gin.Context) {
	var req struct {
		ID        int     `json:"id" binding:"required"`
		SellPrice float64 `json:"sell_price" binding:"required"`
		Enabled   int     `json:"enabled" binding:"required"`
		Sort      int     `json:"sort" binding:"required"`
		Content   string  `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := yfdkmodule.YFDK().UpdateProject(req.ID, req.SellPrice, req.Enabled, req.Sort, req.Content); err != nil {
		response.ServerErrorf(c, err, "更新项目失败")
		return
	}
	response.SuccessMsg(c, "更新成功")
}

func YFDKAdminProjectDelete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := yfdkmodule.YFDK().DeleteProject(id); err != nil {
		response.ServerErrorf(c, err, "删除项目失败")
		return
	}
	response.SuccessMsg(c, "删除成功")
}

func TuZhiConfigGet(c *gin.Context) {
	cfg, err := tuzhimodule.TuZhi().GetConfig()
	if err != nil {
		response.ServerErrorf(c, err, "获取配置失败")
		return
	}
	response.Success(c, cfg)
}

func TuZhiConfigSave(c *gin.Context) {
	var cfg tuzhimodule.TuZhiConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := tuzhimodule.TuZhi().SaveConfig(&cfg); err != nil {
		response.ServerErrorf(c, err, "保存配置失败")
		return
	}
	response.SuccessMsg(c, "保存成功")
}

func TuZhiAdminGetGoods(c *gin.Context) {
	goods, err := tuzhimodule.TuZhi().GetGoods()
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, goods)
}

func TuZhiGoodsOverridesGet(c *gin.Context) {
	list, err := tuzhimodule.TuZhi().GetGoodsOverrides()
	if err != nil {
		response.ServerErrorf(c, err, "获取失败")
		return
	}
	response.Success(c, list)
}

func TuZhiGoodsOverridesSave(c *gin.Context) {
	var req struct {
		Items []tuzhimodule.TuZhiGoodsOverride `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := tuzhimodule.TuZhi().SaveGoodsOverrides(req.Items); err != nil {
		response.ServerErrorf(c, err, "保存失败")
		return
	}
	response.SuccessMsg(c, "保存成功")
}

func SXDKConfigGet(c *gin.Context) {
	cfg, err := sxdkmodule.SXDK().GetConfig()
	if err != nil {
		response.ServerErrorf(c, err, "获取配置失败")
		return
	}
	response.Success(c, cfg)
}

func SXDKConfigSave(c *gin.Context) {
	var cfg sxdkmodule.SXDKConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := sxdkmodule.SXDK().SaveConfig(&cfg); err != nil {
		response.ServerErrorf(c, err, "保存配置失败")
		return
	}
	response.SuccessMsg(c, "保存成功")
}

func AdminTenantList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	list, total, err := listAdminTenants(page, limit)
	if err != nil {
		response.ServerErrorf(c, err, "查询失败")
		return
	}
	response.Success(c, gin.H{"list": list, "total": total})
}

func AdminTenantCreate(c *gin.Context) {
	var req model.TenantSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	tid, err := tenant.CreateTenant(req.UID, &req)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, gin.H{"tid": tid})
}

func AdminTenantSetStatus(c *gin.Context) {
	tid, _ := strconv.Atoi(c.Param("tid"))
	var req struct {
		Status int `json:"status"`
	}
	c.ShouldBindJSON(&req)
	if err := setAdminTenantStatus(tid, req.Status); err != nil {
		response.ServerErrorf(c, err, "操作失败")
		return
	}
	response.SuccessMsg(c, "ok")
}

func AdminCheckinStats(c *gin.Context) {
	date := c.Query("date")
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	list, total, stat, err := listAdminCheckinStats(date, page, limit)
	if err != nil {
		response.ServerErrorf(c, err, "查询签到记录失败")
		return
	}
	response.Success(c, gin.H{
		"list":         list,
		"total":        total,
		"total_users":  stat.TotalUsers,
		"total_reward": stat.TotalReward,
	})
}

func AdminDBCompatCheck(c *gin.Context) {
	result, err := checkAdminDBCompat()
	if err != nil {
		response.ServerErrorf(c, err, "检查失败: "+err.Error())
		return
	}
	response.Success(c, result)
}

func AdminDBCompatFix(c *gin.Context) {
	result, err := fixAdminDBCompat()
	if err != nil {
		response.ServerErrorf(c, err, "修复失败: "+err.Error())
		return
	}
	response.Success(c, result)
}

func AdminDBSyncTest(c *gin.Context) {
	var req SyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	// 后台入口只负责请求绑定和统一响应，真正的同步探测逻辑交给 dbtools。
	result, err := testAdminDBSyncConnection(req)
	if err != nil {
		response.ServerErrorf(c, err, "测试失败: "+err.Error())
		return
	}
	response.Success(c, result)
}

func AdminDBSyncExecute(c *gin.Context) {
	var req SyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	if strings.TrimSpace(req.ConfirmationToken) == "" {
		response.BadRequest(c, "请先完成预检查，再执行导入")
		return
	}
	// 后台入口只做协议层处理，真正的数据同步执行由 dbtools 负责。
	result, err := executeAdminDBSync(req)
	if err != nil {
		response.ServerErrorf(c, err, "同步失败: "+err.Error())
		return
	}
	response.Success(c, result)
}

func AdminCardKeyList(c *gin.Context) {
	var req model.CardKeyListRequest
	_ = c.ShouldBindQuery(&req)
	list, total, err := auxmodule.Auxiliary().CardKeyList(req)
	if err != nil {
		response.ServerErrorf(c, err, "查询卡密失败")
		return
	}
	response.Success(c, gin.H{
		"list":       list,
		"pagination": gin.H{"page": req.Page, "limit": req.Limit, "total": total},
	})
}

func AdminCardKeyGenerate(c *gin.Context) {
	var req model.CardKeyGenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写金额和数量")
		return
	}
	codes, err := auxmodule.Auxiliary().CardKeyGenerate(req.Money, req.Count)
	if err != nil {
		response.ServerErrorf(c, err, "生成卡密失败")
		return
	}
	response.Success(c, gin.H{"codes": codes, "count": len(codes)})
}

func AdminCardKeyDelete(c *gin.Context) {
	var body struct {
		IDs []int `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || len(body.IDs) == 0 {
		response.BadRequest(c, "请选择要删除的卡密")
		return
	}
	deleted, err := auxmodule.Auxiliary().CardKeyDelete(body.IDs)
	if err != nil {
		response.ServerErrorf(c, err, "删除失败")
		return
	}
	response.Success(c, gin.H{"deleted": deleted})
}

func AdminActivityList(c *gin.Context) {
	var req model.ActivityListRequest
	_ = c.ShouldBindQuery(&req)
	list, total, err := auxmodule.Auxiliary().ActivityList(req)
	if err != nil {
		response.ServerErrorf(c, err, "查询活动失败")
		return
	}
	response.Success(c, gin.H{
		"list":       list,
		"pagination": gin.H{"page": req.Page, "limit": req.Limit, "total": total},
	})
}

func AdminActivitySave(c *gin.Context) {
	var req model.ActivitySaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写完整的活动信息")
		return
	}
	if err := auxmodule.Auxiliary().ActivitySave(req); err != nil {
		response.ServerErrorf(c, err, "保存活动失败")
		return
	}
	response.SuccessMsg(c, "保存成功")
}

func AdminActivityDelete(c *gin.Context) {
	hid, _ := strconv.Atoi(c.Param("hid"))
	if hid <= 0 {
		response.BadRequest(c, "无效的活动ID")
		return
	}
	if err := auxmodule.Auxiliary().ActivityDelete(hid); err != nil {
		response.ServerErrorf(c, err, "删除失败")
		return
	}
	response.SuccessMsg(c, "删除成功")
}

func AdminPledgeConfigList(c *gin.Context) {
	list, err := auxmodule.Auxiliary().PledgeConfigList()
	if err != nil {
		response.ServerErrorf(c, err, "查询质押配置失败")
		return
	}
	response.Success(c, list)
}

func AdminPledgeConfigSave(c *gin.Context) {
	var req model.PledgeConfigSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写完整的质押配置")
		return
	}
	if err := auxmodule.Auxiliary().PledgeConfigSave(req); err != nil {
		response.ServerErrorf(c, err, "保存失败")
		return
	}
	response.SuccessMsg(c, "保存成功")
}

func AdminPledgeConfigDelete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		response.BadRequest(c, "无效的配置ID")
		return
	}
	if err := auxmodule.Auxiliary().PledgeConfigDelete(id); err != nil {
		response.ServerErrorf(c, err, "删除失败")
		return
	}
	response.SuccessMsg(c, "删除成功")
}

func AdminPledgeConfigToggle(c *gin.Context) {
	var body struct {
		ID     int `json:"id" binding:"required"`
		Status int `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := auxmodule.Auxiliary().PledgeConfigToggle(body.ID, body.Status); err != nil {
		response.ServerErrorf(c, err, "更新失败")
		return
	}
	response.SuccessMsg(c, "更新成功")
}

func AdminPledgeRecordList(c *gin.Context) {
	var req model.PledgeListRequest
	_ = c.ShouldBindQuery(&req)
	list, total, err := auxmodule.Auxiliary().PledgeRecordList(req)
	if err != nil {
		response.ServerErrorf(c, err, "查询质押记录失败")
		return
	}
	response.Success(c, gin.H{
		"list":       list,
		"pagination": gin.H{"page": req.Page, "limit": req.Limit, "total": total},
	})
}
