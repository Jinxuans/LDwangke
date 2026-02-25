package service

import (
	"regexp"
	"strings"
)

// ParsedPHPConfig 从 PHP if 代码块解析出的平台配置
type ParsedPHPConfig struct {
	PT              string `json:"pt"`
	Name            string `json:"name"`
	AuthType        string `json:"auth_type"`
	APIPathStyle    string `json:"api_path_style"`
	SuccessCodes    string `json:"success_codes"`
	UseJSON         bool   `json:"use_json"`
	QueryAct        string `json:"query_act"`
	QueryPath       string `json:"query_path"`
	OrderAct        string `json:"order_act"`
	OrderPath       string `json:"order_path"`
	ProgressAct     string `json:"progress_act"`
	ProgressPath    string `json:"progress_path"`
	ProgressMethod  string `json:"progress_method"`
	PauseAct        string `json:"pause_act"`
	PausePath       string `json:"pause_path"`
	ChangePassAct   string `json:"change_pass_act"`
	ChangePassPath  string `json:"change_pass_path"`
	ChangePassParam string `json:"change_pass_param"`
	ChangePassID    string `json:"change_pass_id_param"`
	LogAct          string `json:"log_act"`
	LogPath         string `json:"log_path"`
	LogIDParam      string `json:"log_id_param"`
	ReturnsYID      bool   `json:"returns_yid"`
	Confidence      int    `json:"confidence"` // 解析置信度 0-100
	Warnings        []string `json:"warnings"`  // 解析警告
}

// ParsePHPCode 解析 PHP if 代码块，提取平台配置
// 支持 ckjk.php(查课)、xdjk.php(下单)、jdjk.php(进度)、xgjk.php(改密)、pausejk.php(暂停) 等格式
func ParsePHPCode(code string) *ParsedPHPConfig {
	cfg := &ParsedPHPConfig{
		AuthType:        "uid_key",
		APIPathStyle:    "standard",
		SuccessCodes:    "0",
		QueryAct:        "get",
		OrderAct:        "add",
		ProgressAct:     "chadan2",
		ProgressMethod:  "POST",
		PauseAct:        "zt",
		ChangePassAct:   "gaimi",
		ChangePassParam: "newPwd",
		ChangePassID:    "id",
		LogAct:          "xq",
		LogIDParam:      "id",
		Confidence:      0,
		Warnings:        []string{},
	}

	// 1. 提取平台标识 $type == "xxx"
	ptRe := regexp.MustCompile(`\$type\s*==\s*["']([^"']+)["']`)
	if m := ptRe.FindStringSubmatch(code); len(m) > 1 {
		cfg.PT = m[1]
		cfg.Name = m[1]
		cfg.Confidence += 20
	} else {
		cfg.Warnings = append(cfg.Warnings, "未找到 $type == \"xxx\" 平台标识")
	}

	// 2. 提取 act 参数 act=xxx 或 ?act=xxx
	actRe := regexp.MustCompile(`act=([a-zA-Z0-9_]+)`)
	actMatches := actRe.FindAllStringSubmatch(code, -1)
	if len(actMatches) > 0 {
		act := actMatches[0][1]
		cfg.Confidence += 15

		// 根据 act 名推断操作类型
		switch {
		case act == "get":
			cfg.QueryAct = "get"
		case act == "add" || strings.Contains(act, "add"):
			cfg.OrderAct = act
		case strings.Contains(act, "chadan"):
			cfg.ProgressAct = act
		case act == "zt" || act == "zanting" || act == "stop":
			cfg.PauseAct = act
		case act == "gaimi" || act == "xgmm" || strings.Contains(act, "update"):
			cfg.ChangePassAct = act
		case act == "xq" || strings.Contains(act, "log"):
			cfg.LogAct = act
		default:
			// 猜测：如果代码中有查课关键词
			if containsAny(code, "school", "user", "pass", "platform", "查课", "get") {
				cfg.QueryAct = act
			} else if containsAny(code, "kcname", "kcid", "下单") {
				cfg.OrderAct = act
			}
		}
	}

	// 3. 检测 REST 风格路径 /api/xxx
	restPathRe := regexp.MustCompile(`["'](/api/[a-zA-Z0-9_/\-]+)["']`)
	if m := restPathRe.FindStringSubmatch(code); len(m) > 1 {
		path := m[1]
		cfg.APIPathStyle = "rest"
		cfg.Confidence += 10

		// 根据路径推断操作
		pathLower := strings.ToLower(path)
		switch {
		case strings.Contains(pathLower, "query") || strings.Contains(pathLower, "course") || strings.Contains(pathLower, "login") || strings.Contains(pathLower, "get"):
			cfg.QueryPath = path
		case strings.Contains(pathLower, "add") || strings.Contains(pathLower, "create") || strings.Contains(pathLower, "order"):
			cfg.OrderPath = path
		case strings.Contains(pathLower, "search") || strings.Contains(pathLower, "chadan") || strings.Contains(pathLower, "progress"):
			cfg.ProgressPath = path
		case strings.Contains(pathLower, "freeze") || strings.Contains(pathLower, "stop") || strings.Contains(pathLower, "pause"):
			cfg.PausePath = path
		case strings.Contains(pathLower, "update") || strings.Contains(pathLower, "password") || strings.Contains(pathLower, "change"):
			cfg.ChangePassPath = path
		case strings.Contains(pathLower, "reset") || strings.Contains(pathLower, "resubmit"):
			// resubmit path
		case strings.Contains(pathLower, "log"):
			cfg.LogPath = path
		}
	}

	// 4. 检测认证方式
	if strings.Contains(code, `$a["token"]`) || strings.Contains(code, `$token`) || strings.Contains(code, `"token" =>`) {
		if !strings.Contains(code, `$a["user"]`) && !strings.Contains(code, `"uid"`) {
			cfg.AuthType = "token_field"
			cfg.Confidence += 10
		}
	}
	if strings.Contains(code, `"token" => $a["pass"]`) {
		cfg.AuthType = "token_only"
		cfg.Confidence += 10
	}

	// 5. 检测成功码
	successRe := regexp.MustCompile(`\$result\['code'\]\s*==\s*["']?(\d+)["']?`)
	successMatches := successRe.FindAllStringSubmatch(code, -1)
	if len(successMatches) > 0 {
		codes := map[string]bool{}
		for _, m := range successMatches {
			codes[m[1]] = true
		}
		codeList := []string{}
		for c := range codes {
			codeList = append(codeList, c)
		}
		cfg.SuccessCodes = strings.Join(codeList, ",")
		cfg.Confidence += 10
	}

	// 也检测 ($result['code'] == 0 || $result['code'] == 1 || $result['code'] == 200)
	multiCodeRe := regexp.MustCompile(`\$result\['code'\]\s*==\s*(\d+)\s*\|\|\s*\$result\['code'\]\s*==\s*(\d+)(?:\s*\|\|\s*\$result\['code'\]\s*==\s*(\d+))?`)
	if m := multiCodeRe.FindStringSubmatch(code); len(m) > 1 {
		codes := map[string]bool{}
		for _, c := range m[1:] {
			if c != "" {
				codes[c] = true
			}
		}
		codeList := []string{}
		for c := range codes {
			codeList = append(codeList, c)
		}
		cfg.SuccessCodes = strings.Join(codeList, ",")
	}

	// 6. 检测 JSON 请求
	if strings.Contains(code, "json_encode") && strings.Contains(code, "CURLOPT_POSTFIELDS") {
		cfg.UseJSON = true
		cfg.Confidence += 5
	}
	if strings.Contains(code, "httpRequest") && strings.Contains(code, "true)") {
		cfg.UseJSON = true
		cfg.Confidence += 5
	}
	if strings.Contains(code, "application/json") {
		cfg.UseJSON = true
		cfg.Confidence += 5
	}

	// 7. 检测改密参数名
	passParamRe := regexp.MustCompile(`["'](newPwd|xgmm|newpwd|new_password|odpwd)["']\s*=>`)
	if m := passParamRe.FindStringSubmatch(code); len(m) > 1 {
		cfg.ChangePassParam = m[1]
		cfg.Confidence += 5
	}

	// 8. 检测 ID 参数名
	idParamRe := regexp.MustCompile(`["'](oid|id|ids)["']\s*=>\s*\$(?:d\["yid"\]|yid|order)`)
	if m := idParamRe.FindStringSubmatch(code); len(m) > 1 {
		cfg.ChangePassID = m[1]
		cfg.Confidence += 5
	}

	// 9. 检测 yid 返回
	if strings.Contains(code, `$result['id']`) || strings.Contains(code, `$result['yid']`) || strings.Contains(code, `"yid"`) {
		cfg.ReturnsYID = true
		cfg.Confidence += 5
	}

	// 10. 检测代理
	if strings.Contains(code, "getProxy") || strings.Contains(code, "CURLPROXY_SOCKS5") || strings.Contains(code, "proxy") {
		cfg.Warnings = append(cfg.Warnings, "检测到代理逻辑，需要手动启用 need_proxy")
	}

	// 11. 检测请求方法
	if strings.Contains(code, "CURLOPT_CUSTOMREQUEST") && strings.Contains(code, "GET") {
		cfg.ProgressMethod = "GET"
	}
	if strings.Contains(code, "file_get_contents") {
		cfg.ProgressMethod = "GET"
	}

	// 12. 检测特殊的暂停 act
	pauseRe := regexp.MustCompile(`act=([a-zA-Z_]+).*(?:暂停|pause|stop|freeze)`)
	if m := pauseRe.FindStringSubmatch(code); len(m) > 1 {
		cfg.PauseAct = m[1]
	}
	// 反向：从注释/上下文推断
	if strings.Contains(code, "暂停") || strings.Contains(code, "pause") || strings.Contains(code, "freeze") {
		pauseActRe := regexp.MustCompile(`act=([a-zA-Z_]+)`)
		if m := pauseActRe.FindStringSubmatch(code); len(m) > 1 {
			cfg.PauseAct = m[1]
		}
	}

	// 最终置信度调整
	if cfg.PT != "" {
		cfg.Confidence += 10
	}
	if cfg.Confidence > 100 {
		cfg.Confidence = 100
	}
	if cfg.Confidence < 30 {
		cfg.Warnings = append(cfg.Warnings, "解析置信度较低，建议手动核对配置")
	}

	return cfg
}

func containsAny(s string, substrs ...string) bool {
	s = strings.ToLower(s)
	for _, sub := range substrs {
		if strings.Contains(s, strings.ToLower(sub)) {
			return true
		}
	}
	return false
}
