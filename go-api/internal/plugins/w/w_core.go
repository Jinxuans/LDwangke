package w

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go-api/internal/database"
)

func (s *WService) EnsureTable() {
	_, err := database.DB.Exec(`CREATE TABLE IF NOT EXISTS w_app (
		id BIGINT NOT NULL AUTO_INCREMENT,
		name VARCHAR(255) NOT NULL COMMENT '项目名称',
		code VARCHAR(50) NOT NULL COMMENT '项目代码',
		org_app_id VARCHAR(10) NOT NULL COMMENT '源项目ID',
		status TINYINT DEFAULT 0 COMMENT '0上架 1下架',
		description TEXT NULL COMMENT '项目说明',
		price DECIMAL(18,2) NOT NULL DEFAULT 1 COMMENT '单价',
		cac_type VARCHAR(2) NOT NULL COMMENT 'TS按次 KM按公里',
		url VARCHAR(255) NOT NULL COMMENT '对接URL',
		` + "`key`" + ` VARCHAR(255) DEFAULT NULL COMMENT '对接密钥',
		uid VARCHAR(255) DEFAULT NULL COMMENT '对接UID',
		token VARCHAR(1024) DEFAULT NULL COMMENT '源台token',
		type VARCHAR(50) NOT NULL COMMENT '项目类型',
		deleted TINYINT DEFAULT 0 COMMENT '软删除',
		created DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		PRIMARY KEY (id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='W对接项目表'`)
	if err != nil {
		fmt.Printf("[W] 建 w_app 表失败: %v\n", err)
	}

	_, err = database.DB.Exec(`CREATE TABLE IF NOT EXISTS w_order (
		id BIGINT NOT NULL AUTO_INCREMENT,
		agg_order_id VARCHAR(10) DEFAULT NULL UNIQUE COMMENT 'W源台订单ID',
		user_id BIGINT NOT NULL COMMENT '用户ID',
		school VARCHAR(255) DEFAULT NULL COMMENT '学校名称',
		account VARCHAR(255) NOT NULL COMMENT '账号',
		password VARCHAR(255) NOT NULL COMMENT '密码',
		app_id BIGINT NOT NULL COMMENT '项目ID',
		status VARCHAR(50) NOT NULL COMMENT '订单状态',
		num INT NOT NULL COMMENT '次数',
		cost DECIMAL(18,2) DEFAULT 0 COMMENT '金额',
		pause TINYINT(1) DEFAULT 0 COMMENT '是否暂停',
		sub_order JSON DEFAULT NULL COMMENT '子订单',
		deleted TINYINT(1) DEFAULT 0 COMMENT '软删除',
		created DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		PRIMARY KEY (id),
		KEY idx_user_id (user_id),
		KEY idx_deleted (deleted)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='W跑步订单表'`)
	if err != nil {
		fmt.Printf("[W] 建 w_order 表失败: %v\n", err)
	}
}

func (s *WService) httpReq(method, reqURL string, body interface{}, headers map[string]string) (map[string]interface{}, error) {
	var reqBody io.Reader
	if body != nil {
		switch v := body.(type) {
		case string:
			reqBody = strings.NewReader(v)
		default:
			jsonData, _ := json.Marshal(body)
			reqBody = strings.NewReader(string(jsonData))
		}
	}

	req, err := http.NewRequest(method, reqURL, reqBody)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	if method == "POST" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求外部接口失败: %v", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("外部接口返回格式错误: %s", string(respBody))
	}
	return result, nil
}

func wLog(uid int, logType, text string, money float64) {
	now := time.Now().Format("2006-01-02 15:04:05")
	var newBalance float64
	database.DB.QueryRow("SELECT COALESCE(money,0) FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&newBalance)
	moneyStr := fmt.Sprintf("%.2f", money)
	if money > 0 {
		moneyStr = fmt.Sprintf("+%.2f", money)
	}
	database.DB.Exec(
		"INSERT INTO qingka_wangke_log (uid, type, text, money, smoney, ip, addtime) VALUES (?, ?, ?, ?, ?, '', ?)",
		uid, logType, text, moneyStr, fmt.Sprintf("%.2f", newBalance), now,
	)
}

func (s *WService) getAppRow(appID int64) (map[string]interface{}, error) {
	rows, err := database.DB.Query("SELECT * FROM w_app WHERE id = ? AND deleted = 0 LIMIT 1", appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	columns, _ := rows.Columns()
	if !rows.Next() {
		return nil, fmt.Errorf("项目不存在")
	}
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}
	rows.Scan(valuePtrs...)
	result := make(map[string]interface{})
	for i, col := range columns {
		val := values[i]
		if b, ok := val.([]byte); ok {
			result[col] = string(b)
		} else {
			result[col] = val
		}
	}
	return result, nil
}

func (s *WService) appRequest(app map[string]interface{}, act string, body interface{}, method string) (map[string]interface{}, error) {
	pURL := strings.TrimSpace(fmt.Sprintf("%v", app["url"]))
	pType := fmt.Sprintf("%v", app["type"])
	token := fmt.Sprintf("%v", app["token"])
	key := fmt.Sprintf("%v", app["key"])
	pUID := fmt.Sprintf("%v", app["uid"])

	headers := map[string]string{}
	var reqURL string

	if pType == "0" {
		params := url.Values{}
		params.Set("act", formatAct(act))
		params.Set("key", key)
		params.Set("uid", pUID)
		reqURL = pURL + "?" + params.Encode()
	} else {
		reqURL = strings.TrimRight(pURL, "/") + act
		headers["X-WTK"] = token
	}

	if method == "" {
		method = "POST"
	}

	return s.httpReq(method, reqURL, body, headers)
}

func formatAct(orgAct string) string {
	return strings.ReplaceAll(strings.TrimLeft(orgAct, "/"), "/", "-")
}

func unformatAct(outAct string) string {
	s := strings.ReplaceAll(outAct, "-", "/")
	if len(s) > 0 && s[0] != '/' {
		s = "/" + s
	}
	return s
}
