package handler

import (
	"go-api/internal/database"
	"go-api/internal/model"
	"go-api/internal/response"
	"go-api/internal/service"

	"github.com/gin-gonic/gin"
)

// AdminPlatformConfigList 获取所有平台配置
func AdminPlatformConfigList(c *gin.Context) {
	rows, err := database.DB.Query(`SELECT id, pt, name, auth_type, api_path_style, success_codes,
		use_json, need_proxy, returns_yid, extra_params,
		query_act, query_path, query_param_style, query_polling, query_max_attempts, query_interval, COALESCE(query_response_map,''),
		order_act, order_path, yid_in_data_array,
		progress_act, progress_no_yid, progress_path, progress_method, progress_needs_auth,
		use_id_param, use_uuid_param, always_username,
		pause_act, pause_path, COALESCE(pause_id_param,'id'), resume_act, resume_path,
		change_pass_act, change_pass_path, change_pass_param, change_pass_id_param,
		resubmit_path, COALESCE(resubmit_id_param,'id'),
		log_act, log_path, log_method, log_id_param,
		COALESCE(balance_act,'getmoney'), COALESCE(balance_path,''), COALESCE(balance_money_field,'money'),
		COALESCE(balance_method,'POST'), COALESCE(balance_auth_type,''),
		COALESCE(report_param_style,''), COALESCE(report_auth_type,''), COALESCE(report_path,''), COALESCE(get_report_path,''), COALESCE(refresh_path,''),
		COALESCE(source_code,''), created_at, updated_at
		FROM qingka_platform_config ORDER BY pt`)
	if err != nil {
		response.ServerError(c, "查询失败: "+err.Error())
		return
	}
	defer rows.Close()

	var list []model.PlatformConfigDB
	for rows.Next() {
		var cfg model.PlatformConfigDB
		err := rows.Scan(
			&cfg.ID, &cfg.PT, &cfg.Name, &cfg.AuthType, &cfg.APIPathStyle, &cfg.SuccessCodes,
			&cfg.UseJSON, &cfg.NeedProxy, &cfg.ReturnsYID, &cfg.ExtraParams,
			&cfg.QueryAct, &cfg.QueryPath, &cfg.QueryParamStyle, &cfg.QueryPolling, &cfg.QueryMaxAttempts, &cfg.QueryInterval, &cfg.QueryResponseMap,
			&cfg.OrderAct, &cfg.OrderPath, &cfg.YIDInDataArray,
			&cfg.ProgressAct, &cfg.ProgressNoYID, &cfg.ProgressPath, &cfg.ProgressMethod, &cfg.ProgressNeedsAuth,
			&cfg.UseIDParam, &cfg.UseUUIDParam, &cfg.AlwaysUsername,
			&cfg.PauseAct, &cfg.PausePath, &cfg.PauseIDParam, &cfg.ResumeAct, &cfg.ResumePath,
			&cfg.ChangePassAct, &cfg.ChangePassPath, &cfg.ChangePassParam, &cfg.ChangePassIDParam,
			&cfg.ResubmitPath, &cfg.ResubmitIDParam,
			&cfg.LogAct, &cfg.LogPath, &cfg.LogMethod, &cfg.LogIDParam,
			&cfg.BalanceAct, &cfg.BalancePath, &cfg.BalanceMoneyField, &cfg.BalanceMethod, &cfg.BalanceAuthType,
			&cfg.ReportParamStyle, &cfg.ReportAuthType, &cfg.ReportPath, &cfg.GetReportPath, &cfg.RefreshPath,
			&cfg.SourceCode, &cfg.CreatedAt, &cfg.UpdatedAt,
		)
		if err != nil {
			response.ServerError(c, "解析数据失败: "+err.Error())
			return
		}
		list = append(list, cfg)
	}
	if list == nil {
		list = []model.PlatformConfigDB{}
	}
	response.Success(c, list)
}

// AdminPlatformConfigSave 保存平台配置（新增/更新）
func AdminPlatformConfigSave(c *gin.Context) {
	var req model.PlatformConfigSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 填充默认值
	if req.AuthType == "" {
		req.AuthType = "uid_key"
	}
	if req.APIPathStyle == "" {
		req.APIPathStyle = "standard"
	}
	if req.SuccessCodes == "" {
		req.SuccessCodes = "0"
	}
	if req.QueryAct == "" {
		req.QueryAct = "get"
	}
	if req.OrderAct == "" {
		req.OrderAct = "add"
	}
	if req.ProgressAct == "" {
		req.ProgressAct = "chadan2"
	}
	if req.ProgressNoYID == "" {
		req.ProgressNoYID = "chadan"
	}
	if req.ProgressMethod == "" {
		req.ProgressMethod = "POST"
	}
	if req.PauseAct == "" {
		req.PauseAct = "zt"
	}
	if req.PauseIDParam == "" {
		req.PauseIDParam = "id"
	}
	if req.ChangePassAct == "" {
		req.ChangePassAct = "gaimi"
	}
	if req.ChangePassParam == "" {
		req.ChangePassParam = "newPwd"
	}
	if req.ChangePassIDParam == "" {
		req.ChangePassIDParam = "id"
	}
	if req.ResubmitIDParam == "" {
		req.ResubmitIDParam = "id"
	}
	if req.LogAct == "" {
		req.LogAct = "xq"
	}
	if req.LogMethod == "" {
		req.LogMethod = "POST"
	}
	if req.LogIDParam == "" {
		req.LogIDParam = "id"
	}
	if req.BalanceAct == "" {
		req.BalanceAct = "getmoney"
	}
	if req.BalanceMoneyField == "" {
		req.BalanceMoneyField = "money"
	}
	if req.BalanceMethod == "" {
		req.BalanceMethod = "POST"
	}

	query := `INSERT INTO qingka_platform_config (
		pt, name, auth_type, api_path_style, success_codes,
		use_json, need_proxy, returns_yid, extra_params,
		query_act, query_path, query_param_style, query_polling, query_max_attempts, query_interval, query_response_map,
		order_act, order_path, yid_in_data_array,
		progress_act, progress_no_yid, progress_path, progress_method, progress_needs_auth,
		use_id_param, use_uuid_param, always_username,
		pause_act, pause_path, pause_id_param, resume_act, resume_path,
		change_pass_act, change_pass_path, change_pass_param, change_pass_id_param,
		resubmit_path, resubmit_id_param,
		log_act, log_path, log_method, log_id_param,
		balance_act, balance_path, balance_money_field, balance_method, balance_auth_type,
		report_param_style, report_auth_type, report_path, get_report_path, refresh_path,
		source_code
	) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
	ON DUPLICATE KEY UPDATE
		name=VALUES(name), auth_type=VALUES(auth_type), api_path_style=VALUES(api_path_style),
		success_codes=VALUES(success_codes), use_json=VALUES(use_json), need_proxy=VALUES(need_proxy),
		returns_yid=VALUES(returns_yid), extra_params=VALUES(extra_params),
		query_act=VALUES(query_act), query_path=VALUES(query_path), query_param_style=VALUES(query_param_style),
		query_polling=VALUES(query_polling), query_max_attempts=VALUES(query_max_attempts),
		query_interval=VALUES(query_interval), query_response_map=VALUES(query_response_map),
		order_act=VALUES(order_act), order_path=VALUES(order_path), yid_in_data_array=VALUES(yid_in_data_array),
		progress_act=VALUES(progress_act), progress_no_yid=VALUES(progress_no_yid),
		progress_path=VALUES(progress_path), progress_method=VALUES(progress_method),
		progress_needs_auth=VALUES(progress_needs_auth),
		use_id_param=VALUES(use_id_param), use_uuid_param=VALUES(use_uuid_param),
		always_username=VALUES(always_username),
		pause_act=VALUES(pause_act), pause_path=VALUES(pause_path), pause_id_param=VALUES(pause_id_param),
		resume_act=VALUES(resume_act), resume_path=VALUES(resume_path),
		change_pass_act=VALUES(change_pass_act), change_pass_path=VALUES(change_pass_path),
		change_pass_param=VALUES(change_pass_param), change_pass_id_param=VALUES(change_pass_id_param),
		resubmit_path=VALUES(resubmit_path), resubmit_id_param=VALUES(resubmit_id_param),
		log_act=VALUES(log_act), log_path=VALUES(log_path), log_method=VALUES(log_method),
		log_id_param=VALUES(log_id_param),
		balance_act=VALUES(balance_act), balance_path=VALUES(balance_path),
		balance_money_field=VALUES(balance_money_field), balance_method=VALUES(balance_method),
		balance_auth_type=VALUES(balance_auth_type),
		report_param_style=VALUES(report_param_style), report_auth_type=VALUES(report_auth_type), report_path=VALUES(report_path), get_report_path=VALUES(get_report_path), refresh_path=VALUES(refresh_path),
		source_code=VALUES(source_code)`

	_, err := database.DB.Exec(query,
		req.PT, req.Name, req.AuthType, req.APIPathStyle, req.SuccessCodes,
		req.UseJSON, req.NeedProxy, req.ReturnsYID, req.ExtraParams,
		req.QueryAct, req.QueryPath, req.QueryParamStyle, req.QueryPolling, req.QueryMaxAttempts, req.QueryInterval, req.QueryResponseMap,
		req.OrderAct, req.OrderPath, req.YIDInDataArray,
		req.ProgressAct, req.ProgressNoYID, req.ProgressPath, req.ProgressMethod, req.ProgressNeedsAuth,
		req.UseIDParam, req.UseUUIDParam, req.AlwaysUsername,
		req.PauseAct, req.PausePath, req.PauseIDParam, req.ResumeAct, req.ResumePath,
		req.ChangePassAct, req.ChangePassPath, req.ChangePassParam, req.ChangePassIDParam,
		req.ResubmitPath, req.ResubmitIDParam,
		req.LogAct, req.LogPath, req.LogMethod, req.LogIDParam,
		req.BalanceAct, req.BalancePath, req.BalanceMoneyField, req.BalanceMethod, req.BalanceAuthType,
		req.ReportParamStyle, req.ReportAuthType, req.ReportPath, req.GetReportPath, req.RefreshPath,
		req.SourceCode,
	)
	if err != nil {
		response.ServerError(c, "保存失败: "+err.Error())
		return
	}

	service.ReloadPlatformConfigs()
	response.SuccessMsg(c, "保存成功")
}

// AdminPlatformConfigDelete 删除平台配置
func AdminPlatformConfigDelete(c *gin.Context) {
	pt := c.Param("pt")
	if pt == "" {
		response.BadRequest(c, "缺少平台标识")
		return
	}

	result, err := database.DB.Exec("DELETE FROM qingka_platform_config WHERE pt = ?", pt)
	if err != nil {
		response.ServerError(c, "删除失败: "+err.Error())
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		response.BadRequest(c, "平台不存在")
		return
	}
	service.ReloadPlatformConfigs()
	response.SuccessMsg(c, "删除成功")
}

// AdminDetectPlatform 自动检测上游平台API能力
func AdminDetectPlatform(c *gin.Context) {
	var req service.DetectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请提供平台 URL")
		return
	}

	result := service.DetectPlatform(req)
	response.Success(c, result)
}

// AdminAutoDetectSave 自动检测平台并保存配置（一键完成）
func AdminAutoDetectSave(c *gin.Context) {
	var req service.AutoDetectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请提供平台 URL 和标识")
		return
	}

	// 1. 执行检测
	result := service.DetectPlatform(req.DetectRequest)
	if !result.Success {
		response.Success(c, gin.H{
			"success": false,
			"msg":     "未检测到可用接口，请检查 URL 和凭证",
			"detect":  result,
		})
		return
	}

	// 2. 转换为可保存的配置
	saveReq := service.BuildConfigFromDetection(result, req.PT, req.Name)
	if saveReq == nil {
		response.Success(c, gin.H{
			"success": false,
			"msg":     "检测结果无法转换为配置",
			"detect":  result,
		})
		return
	}

	// 3. 保存到数据库（与 AdminPlatformConfigSave 相同的 upsert 逻辑）
	query := `INSERT INTO qingka_platform_config (
		pt, name, auth_type, api_path_style, success_codes,
		use_json, need_proxy, returns_yid, extra_params,
		query_act, query_path, query_param_style, query_polling, query_max_attempts, query_interval, query_response_map,
		order_act, order_path, yid_in_data_array,
		progress_act, progress_no_yid, progress_path, progress_method, progress_needs_auth,
		use_id_param, use_uuid_param, always_username,
		pause_act, pause_path, pause_id_param, resume_act, resume_path,
		change_pass_act, change_pass_path, change_pass_param, change_pass_id_param,
		resubmit_path, resubmit_id_param,
		log_act, log_path, log_method, log_id_param,
		balance_act, balance_path, balance_money_field, balance_method, balance_auth_type,
		report_param_style, report_auth_type, report_path, get_report_path, refresh_path,
		source_code
	) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
	ON DUPLICATE KEY UPDATE
		name=VALUES(name), auth_type=VALUES(auth_type), api_path_style=VALUES(api_path_style),
		success_codes=VALUES(success_codes), use_json=VALUES(use_json),
		returns_yid=VALUES(returns_yid),
		query_act=VALUES(query_act), query_path=VALUES(query_path),
		order_act=VALUES(order_act), order_path=VALUES(order_path),
		progress_act=VALUES(progress_act), progress_no_yid=VALUES(progress_no_yid),
		progress_path=VALUES(progress_path), progress_method=VALUES(progress_method),
		pause_act=VALUES(pause_act), pause_path=VALUES(pause_path), pause_id_param=VALUES(pause_id_param),
		change_pass_act=VALUES(change_pass_act), change_pass_path=VALUES(change_pass_path),
		change_pass_param=VALUES(change_pass_param), change_pass_id_param=VALUES(change_pass_id_param),
		resubmit_path=VALUES(resubmit_path), resubmit_id_param=VALUES(resubmit_id_param),
		log_act=VALUES(log_act), log_path=VALUES(log_path), log_method=VALUES(log_method),
		log_id_param=VALUES(log_id_param),
		balance_act=VALUES(balance_act), balance_path=VALUES(balance_path),
		balance_money_field=VALUES(balance_money_field), balance_method=VALUES(balance_method),
		balance_auth_type=VALUES(balance_auth_type),
		report_param_style=VALUES(report_param_style), report_auth_type=VALUES(report_auth_type),
		report_path=VALUES(report_path), get_report_path=VALUES(get_report_path), refresh_path=VALUES(refresh_path)`

	_, err := database.DB.Exec(query,
		saveReq.PT, saveReq.Name, saveReq.AuthType, saveReq.APIPathStyle, saveReq.SuccessCodes,
		saveReq.UseJSON, saveReq.NeedProxy, saveReq.ReturnsYID, saveReq.ExtraParams,
		saveReq.QueryAct, saveReq.QueryPath, saveReq.QueryParamStyle, saveReq.QueryPolling, saveReq.QueryMaxAttempts, saveReq.QueryInterval, saveReq.QueryResponseMap,
		saveReq.OrderAct, saveReq.OrderPath, saveReq.YIDInDataArray,
		saveReq.ProgressAct, saveReq.ProgressNoYID, saveReq.ProgressPath, saveReq.ProgressMethod, saveReq.ProgressNeedsAuth,
		saveReq.UseIDParam, saveReq.UseUUIDParam, saveReq.AlwaysUsername,
		saveReq.PauseAct, saveReq.PausePath, saveReq.PauseIDParam, saveReq.ResumeAct, saveReq.ResumePath,
		saveReq.ChangePassAct, saveReq.ChangePassPath, saveReq.ChangePassParam, saveReq.ChangePassIDParam,
		saveReq.ResubmitPath, saveReq.ResubmitIDParam,
		saveReq.LogAct, saveReq.LogPath, saveReq.LogMethod, saveReq.LogIDParam,
		saveReq.BalanceAct, saveReq.BalancePath, saveReq.BalanceMoneyField, saveReq.BalanceMethod, saveReq.BalanceAuthType,
		saveReq.ReportParamStyle, saveReq.ReportAuthType, saveReq.ReportPath, saveReq.GetReportPath, saveReq.RefreshPath,
		"",
	)
	if err != nil {
		response.Success(c, gin.H{
			"success": false,
			"msg":     "检测成功但保存失败: " + err.Error(),
			"detect":  result,
			"config":  saveReq,
		})
		return
	}

	service.ReloadPlatformConfigs()
	response.Success(c, gin.H{
		"success": true,
		"msg":     "检测成功并已保存配置",
		"detect":  result,
		"config":  saveReq,
	})
}

// AdminParsePHPCode 解析 PHP 代码，提取平台配置
func AdminParsePHPCode(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请提供 PHP 代码")
		return
	}

	result := service.ParsePHPCode(req.Code)
	response.Success(c, result)
}
