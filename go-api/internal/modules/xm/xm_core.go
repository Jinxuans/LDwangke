package xm

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

// EnsureTable 确保小米运动相关表存在
func (s *XMService) EnsureTable() {
	_, err := database.DB.Exec(`CREATE TABLE IF NOT EXISTS xm_project (
		id BIGINT NOT NULL AUTO_INCREMENT,
		name VARCHAR(255) NOT NULL COMMENT '项目名称',
		p_id INT DEFAULT 0 COMMENT '源项目ID',
		status TINYINT DEFAULT 0 COMMENT '0上架 1下架',
		description TEXT NULL COMMENT '项目说明',
		price DECIMAL(18,2) NOT NULL DEFAULT 0 COMMENT '单价',
		url VARCHAR(255) DEFAULT NULL COMMENT '对接URL',
		` + "`key`" + ` VARCHAR(255) DEFAULT NULL COMMENT '对接密钥',
		uid VARCHAR(255) DEFAULT NULL COMMENT '对接UID',
		token VARCHAR(1024) DEFAULT NULL COMMENT '对接JWT token',
		type VARCHAR(50) DEFAULT NULL COMMENT '项目类型',
		` + "`query`" + ` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否支持查询',
		password TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否需要密码',
		is_deleted TINYINT DEFAULT 0 COMMENT '软删除标记',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		PRIMARY KEY (id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='小米运动对接项目表'`)
	if err != nil {
		fmt.Printf("[XM] 建 xm_project 表失败: %v\n", err)
	}

	_, err = database.DB.Exec(`CREATE TABLE IF NOT EXISTS xm_order (
		id BIGINT NOT NULL AUTO_INCREMENT,
		y_oid BIGINT DEFAULT NULL COMMENT '源订单ID',
		user_id BIGINT NOT NULL COMMENT '用户ID',
		school VARCHAR(255) NOT NULL COMMENT '学校名称',
		account VARCHAR(255) NOT NULL COMMENT '账号',
		password VARCHAR(255) NOT NULL COMMENT '密码',
		type INT DEFAULT NULL COMMENT '跑步类型',
		pace DECIMAL(5,2) DEFAULT NULL COMMENT '配速（分/公里）',
		distance DECIMAL(5,2) DEFAULT NULL COMMENT '单次距离（公里）',
		project_id BIGINT NOT NULL COMMENT '项目ID',
		status VARCHAR(50) NOT NULL COMMENT '订单状态',
		total_km INT NOT NULL COMMENT '下单总公里数',
		run_km FLOAT DEFAULT NULL COMMENT '已跑公里',
		run_date JSON NOT NULL COMMENT '跑步日期',
		start_day DATE NOT NULL COMMENT '开始日期',
		start_time VARCHAR(5) NOT NULL COMMENT '每日开始时间',
		end_time VARCHAR(5) NOT NULL COMMENT '每日结束时间',
		deduction DECIMAL(18,2) DEFAULT 0 COMMENT '扣费金额',
		is_deleted TINYINT(1) DEFAULT 0 COMMENT '软删除标记',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		PRIMARY KEY (id),
		KEY idx_user_id (user_id),
		KEY idx_is_deleted (is_deleted)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='小米运动跑步订单表'`)
	if err != nil {
		fmt.Printf("[XM] 建 xm_order 表失败: %v\n", err)
	}

	database.DB.Exec("ALTER TABLE xm_order ADD COLUMN `pace` DECIMAL(5,2) DEFAULT NULL COMMENT '配速（分/公里）' AFTER `type`")
	database.DB.Exec("ALTER TABLE xm_order ADD COLUMN `distance` DECIMAL(5,2) DEFAULT NULL COMMENT '单次距离（公里）' AFTER `pace`")
}

func (s *XMService) httpRequest(method, reqURL string, body interface{}, headers map[string]string) (map[string]interface{}, error) {
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

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		code := -1
		if c, ok := result["code"].(float64); ok {
			code = int(c)
		}
		msg := fmt.Sprintf("外部接口 HTTP 状态异常: %d", resp.StatusCode)
		if m, ok := result["msg"].(string); ok && m != "" {
			msg = m
		}
		return map[string]interface{}{"code": float64(code), "msg": msg, "data": result["data"]}, nil
	}

	return result, nil
}

func (s *XMService) projectRequest(project map[string]interface{}, act string, body interface{}, method string) (map[string]interface{}, error) {
	pURL := strings.TrimSpace(fmt.Sprintf("%v", project["url"]))
	pType := 0
	if t, ok := project["type"].(int64); ok {
		pType = int(t)
	} else if t, ok := project["type"].([]uint8); ok {
		pType = int(t[0]) - '0'
	} else if t, ok := project["type"].(float64); ok {
		pType = int(t)
	}
	token := fmt.Sprintf("%v", project["token"])
	key := fmt.Sprintf("%v", project["key"])
	pUID := fmt.Sprintf("%v", project["uid"])

	headers := map[string]string{}
	var reqURL string

	if pType == 0 {
		params := url.Values{}
		params.Set("act", act)
		params.Set("key", key)
		params.Set("uid", pUID)
		reqURL = pURL + "?" + params.Encode()
	} else {
		reqURL = strings.TrimRight(pURL, "/") + "/" + act
		headers["token"] = token
	}

	if method == "" {
		method = "POST"
	}

	return s.httpRequest(method, reqURL, body, headers)
}

func xmLog(uid int, logType, text string, money float64) {
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

func (s *XMService) getProjectRow(projectID int) (map[string]interface{}, error) {
	rows, err := database.DB.Query("SELECT * FROM xm_project WHERE id = ? LIMIT 1", projectID)
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

func extractDomain(rawURL string) string {
	pathToRemove := "/api/v1/runorder"
	if strings.HasSuffix(rawURL, pathToRemove) {
		return strings.TrimSuffix(rawURL, pathToRemove)
	}
	if !strings.HasPrefix(rawURL, "http") {
		rawURL = "http://" + rawURL
	}
	parts := strings.SplitN(rawURL, "://", 2)
	if len(parts) == 2 {
		hostPart := strings.SplitN(parts[1], "/", 2)
		return parts[0] + "://" + hostPart[0]
	}
	return strings.TrimRight(rawURL, "/")
}
