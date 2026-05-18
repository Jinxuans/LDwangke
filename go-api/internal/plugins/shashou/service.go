package shashou

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-api/internal/database"
)

const (
	ProjectTypeSource     = 0
	ProjectTypeSource29   = 1
	ProjectTypeSameSystem = 2

	OrderTypeNormal       = 1
	OrderTypeMorning      = 2
	OrderTypeQueryNormal  = 3
	OrderTypeRefund       = 4
	OrderTypeQueryMorning = 5
)

type Project struct {
	ID              int     `json:"id"`
	Name            string  `json:"name"`
	Type            int     `json:"type"`
	RemoteProjectID int     `json:"remote_project_id"`
	APIURL          string  `json:"api_url"`
	APIKey          string  `json:"api_key"`
	UserID          string  `json:"user_id"`
	PriceNormal     float64 `json:"price_normal"`
	PriceMorning    float64 `json:"price_morning"`
	ActualRate      float64 `json:"actual_rate"`
	RushFee         float64 `json:"rush_fee"`
	QueryFee        float64 `json:"query_fee"`
	MinBalance      float64 `json:"min_balance"`
	Status          int     `json:"status"`
	AutoSync        int     `json:"auto_sync"`
	SyncInterval    int     `json:"sync_interval"`
	Timeout         int     `json:"timeout"`
	Remark          string  `json:"remark"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
	ProjectCount    int     `json:"project_count,omitempty"`
}

type AccountForm struct {
	Account     string  `json:"account"`
	Password    string  `json:"password"`
	Distance    float64 `json:"distance"`
	StartHour   int     `json:"start_hour"`
	StartMinute int     `json:"start_minute"`
	EndHour     int     `json:"end_hour"`
	EndMinute   int     `json:"end_minute"`
	RunDays     string  `json:"run_days"`
}

type CreateOrderRequest struct {
	ProjectID   int           `json:"project_id"`
	OrderType   int           `json:"order_type"`
	IsRushOrder bool          `json:"is_rush_order"`
	Accounts    []AccountForm `json:"accounts"`
}

type QueryOrderRequest struct {
	ProjectID int    `json:"project_id"`
	QueryType int    `json:"query_type"`
	Account   string `json:"account"`
}

type RefundOrderRequest struct {
	AccountID int64  `json:"account_id"`
	Account   string `json:"account"`
	ProjectID int    `json:"project_id"`
}

type Order struct {
	ID             int             `json:"id"`
	OrderNo        string          `json:"order_no"`
	UserID         int             `json:"user_id"`
	Username       string          `json:"username"`
	ProjectID      int             `json:"project_id"`
	ProjectName    string          `json:"project_name"`
	OrderType      int             `json:"order_type"`
	IsRushOrder    int             `json:"is_rush_order"`
	TotalDistance  float64         `json:"total_distance"`
	AccountCount   int             `json:"account_count"`
	PreDeduct      float64         `json:"pre_deduct"`
	ActualCost     *float64        `json:"actual_cost"`
	FinalCharge    *float64        `json:"final_charge"`
	Difference     *float64        `json:"difference"`
	RushOrderFee   float64         `json:"rush_order_fee"`
	Status         string          `json:"status"`
	PaymentStatus  string          `json:"payment_status"`
	Accounts       json.RawMessage `json:"accounts,omitempty"`
	QueryAccount   string          `json:"query_account"`
	RefundAccount  string          `json:"refund_account"`
	ResultData     json.RawMessage `json:"result_data,omitempty"`
	CreatedAt      string          `json:"created_at"`
	CompletedAt    string          `json:"completed_at"`
	UpdatedAt      string          `json:"updated_at"`
	ErrorMessage   string          `json:"error_message"`
	RefundKM       *float64        `json:"refund_km"`
	Source         string          `json:"source"`
	AgentUID       int             `json:"agent_uid"`
	AccountDetails []Account       `json:"account_details,omitempty"`
}

type Account struct {
	ID           int64           `json:"id"`
	OrderID      int64           `json:"order_id"`
	OrderNo      string          `json:"order_no"`
	UserID       int             `json:"user_id"`
	Username     string          `json:"username"`
	ProjectID    int             `json:"project_id"`
	Account      string          `json:"account"`
	Password     string          `json:"password"`
	Distance     float64         `json:"distance"`
	StartHour    int             `json:"start_hour"`
	StartMinute  int             `json:"start_minute"`
	EndHour      int             `json:"end_hour"`
	EndMinute    int             `json:"end_minute"`
	RunDays      string          `json:"run_days"`
	OrderType    int             `json:"order_type"`
	IsRushOrder  int             `json:"is_rush_order"`
	Status       string          `json:"status"`
	ErrorMessage string          `json:"error_message"`
	ProcessedAt  string          `json:"processed_at"`
	QueryResult  json.RawMessage `json:"query_result,omitempty"`
	CreatedAt    string          `json:"created_at"`
	UpdatedAt    string          `json:"updated_at"`
}

type Service struct {
	client *http.Client
}

var service = &Service{client: newHTTPClient()}

func newHTTPClient() *http.Client {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           fallbackDialContext(dialer),
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}

func fallbackDialContext(dialer *net.Dialer) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		conn, err := dialer.DialContext(ctx, network, address)
		if err == nil {
			return conn, nil
		}
		host, port, splitErr := net.SplitHostPort(address)
		if splitErr != nil {
			return nil, err
		}
		for _, ip := range fallbackUpstreamIPs(strings.ToLower(host)) {
			conn, fallbackErr := dialer.DialContext(ctx, network, net.JoinHostPort(ip, port))
			if fallbackErr == nil {
				return conn, nil
			}
		}
		return nil, err
	}
}

func fallbackUpstreamIPs(host string) []string {
	switch host {
	case "spiderman.sbs":
		return []string{"104.21.49.181", "172.67.165.134"}
	default:
		return nil
	}
}

func ShaShou() *Service {
	return service
}

func roundMoney(v float64) float64 {
	return math.Round(v*100) / 100
}

func (s *Service) userRate(uid int) float64 {
	rate := 1.0
	_ = database.DB.QueryRow("SELECT COALESCE(addprice,1) FROM qingka_wangke_user WHERE uid=?", uid).Scan(&rate)
	if rate <= 0 {
		rate = 1
	}
	return rate
}

func (s *Service) ListProjects(admin bool) ([]Project, error) {
	where := ""
	if !admin {
		where = " WHERE status=1"
	}
	rows, err := database.DB.Query(`SELECT id, COALESCE(name,'鲨兽运动世界'), type, COALESCE(remote_project_id,0), api_url, api_key, user_id,
		price_normal, price_morning, actual_rate, rush_fee, query_fee, min_balance, status,
		COALESCE(auto_sync,1), COALESCE(sync_interval,300), COALESCE(timeout,30), COALESCE(remark,''),
		COALESCE(DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s'),'')
		FROM ss_project` + where + ` ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []Project{}
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Type, &p.RemoteProjectID, &p.APIURL, &p.APIKey, &p.UserID, &p.PriceNormal, &p.PriceMorning, &p.ActualRate, &p.RushFee, &p.QueryFee, &p.MinBalance, &p.Status, &p.AutoSync, &p.SyncInterval, &p.Timeout, &p.Remark, &p.CreatedAt, &p.UpdatedAt); err == nil {
			if !admin {
				p = redactProject(p)
			}
			list = append(list, p)
		}
	}
	return list, nil
}

func (s *Service) SaveProject(p Project) (int, error) {
	p.Name = strings.TrimSpace(p.Name)
	p.APIURL = strings.TrimRight(strings.TrimSpace(p.APIURL), "/")
	p.APIKey = strings.TrimSpace(p.APIKey)
	p.UserID = strings.TrimSpace(p.UserID)
	if p.Name == "" {
		p.Name = "鲨兽运动世界"
	}
	if p.APIURL == "" || p.APIKey == "" || p.UserID == "" {
		return 0, fmt.Errorf("API地址、密钥和用户ID不能为空")
	}
	if p.Type != ProjectTypeSource && p.Type != ProjectTypeSource29 && p.Type != ProjectTypeSameSystem {
		return 0, fmt.Errorf("对接类型无效")
	}
	if p.RemoteProjectID < 0 {
		p.RemoteProjectID = 0
	}
	if p.PriceNormal <= 0 {
		p.PriceNormal = 9
	}
	if p.PriceMorning <= 0 {
		p.PriceMorning = 10
	}
	if p.ActualRate <= 0 {
		p.ActualRate = 1
	}
	if p.Timeout <= 0 {
		p.Timeout = 30
	}
	if p.SyncInterval <= 0 {
		p.SyncInterval = 300
	}
	if p.ID > 0 {
		_, err := database.DB.Exec(`UPDATE ss_project SET name=?, type=?, remote_project_id=?, api_url=?, api_key=?, user_id=?,
			price_normal=?, price_morning=?, actual_rate=?, rush_fee=?, query_fee=?, min_balance=?, status=?,
			auto_sync=?, sync_interval=?, timeout=?, remark=?, updated_at=NOW() WHERE id=?`,
			p.Name, p.Type, p.RemoteProjectID, p.APIURL, p.APIKey, p.UserID, p.PriceNormal, p.PriceMorning, p.ActualRate, p.RushFee, p.QueryFee,
			p.MinBalance, p.Status, p.AutoSync, p.SyncInterval, p.Timeout, p.Remark, p.ID)
		return p.ID, err
	}
	res, err := database.DB.Exec(`INSERT INTO ss_project
		(name,type,remote_project_id,api_url,api_key,user_id,price_normal,price_morning,actual_rate,rush_fee,query_fee,min_balance,status,auto_sync,sync_interval,timeout,remark)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		p.Name, p.Type, p.RemoteProjectID, p.APIURL, p.APIKey, p.UserID, p.PriceNormal, p.PriceMorning, p.ActualRate, p.RushFee, p.QueryFee,
		p.MinBalance, p.Status, p.AutoSync, p.SyncInterval, p.Timeout, p.Remark)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return int(id), nil
}

func (s *Service) DeleteProject(id int) error {
	if id <= 0 {
		return fmt.Errorf("项目ID无效")
	}
	_, err := database.DB.Exec("DELETE FROM ss_project WHERE id=?", id)
	return err
}

func redactProject(p Project) Project {
	p.APIURL = ""
	p.APIKey = ""
	p.UserID = ""
	p.RemoteProjectID = 0
	p.AutoSync = 0
	p.SyncInterval = 0
	p.Timeout = 0
	p.Remark = ""
	return p
}

func (s *Service) getProject(id int) (Project, error) {
	where := "WHERE status=1"
	args := []any{}
	if id > 0 {
		where += " AND id=?"
		args = append(args, id)
	}
	row := database.DB.QueryRow(`SELECT id, COALESCE(name,'鲨兽运动世界'), type, COALESCE(remote_project_id,0), api_url, api_key, user_id,
		price_normal, price_morning, actual_rate, rush_fee, query_fee, min_balance, status,
		COALESCE(auto_sync,1), COALESCE(sync_interval,300), COALESCE(timeout,30), COALESCE(remark,''),
		COALESCE(DATE_FORMAT(created_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(updated_at,'%Y-%m-%d %H:%i:%s'),'')
		FROM ss_project `+where+` ORDER BY id ASC LIMIT 1`, args...)
	var p Project
	err := row.Scan(&p.ID, &p.Name, &p.Type, &p.RemoteProjectID, &p.APIURL, &p.APIKey, &p.UserID, &p.PriceNormal, &p.PriceMorning, &p.ActualRate, &p.RushFee, &p.QueryFee, &p.MinBalance, &p.Status, &p.AutoSync, &p.SyncInterval, &p.Timeout, &p.Remark, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return Project{}, fmt.Errorf("项目不存在或已禁用")
	}
	return p, err
}

func (s *Service) PricePreview(uid int, projectID, orderType int, rush bool, accounts []AccountForm) (map[string]any, error) {
	project, err := s.getProject(projectID)
	if err != nil {
		return nil, err
	}
	totalDistance := 0.0
	for _, a := range accounts {
		totalDistance += a.Distance
	}
	if totalDistance <= 0 && (orderType == OrderTypeNormal || orderType == OrderTypeMorning) {
		totalDistance = 1
	}
	rate := s.userRate(uid)
	unit := project.PriceNormal
	if orderType == OrderTypeMorning {
		unit = project.PriceMorning
	}
	base := roundMoney(totalDistance * unit * rate)
	rushFee := 0.0
	if rush {
		rushFee = project.RushFee
	}
	if orderType == OrderTypeQueryNormal || orderType == OrderTypeQueryMorning {
		base = roundMoney(project.QueryFee * rate)
	}
	return map[string]any{
		"project":        project,
		"user_rate":      rate,
		"unit_price":     roundMoney(unit * rate),
		"total_distance": roundMoney(totalDistance),
		"rush_fee":       rushFee,
		"amount":         roundMoney(base + rushFee),
	}, nil
}

func (s *Service) CreateOrder(ctx context.Context, uid int, req CreateOrderRequest, source string, agentUID int) (map[string]any, error) {
	if err := validateCreateOrder(req); err != nil {
		return nil, err
	}
	project, err := s.getProject(req.ProjectID)
	if err != nil {
		return nil, err
	}
	totalDistance := 0.0
	for _, a := range req.Accounts {
		totalDistance += a.Distance
	}
	preview, _ := s.PricePreview(uid, req.ProjectID, req.OrderType, req.IsRushOrder, req.Accounts)
	preDeduct := asFloat(preview["amount"])
	if err := s.requireBalance(uid, preDeduct, project.MinBalance); err != nil {
		return nil, err
	}
	upstream, err := s.upstreamCreate(ctx, project, req)
	if err != nil {
		return nil, err
	}
	upstream = ensureUpstreamOrderID(upstream)
	orderNo := extractUpstreamOrderNo(upstream)
	if orderNo == "" {
		orderNo = fmt.Sprintf("SS%d%d", time.Now().Unix(), uid)
	}
	accountsRaw, _ := json.Marshal(req.Accounts)
	resultRaw, _ := json.Marshal(upstream)
	now := time.Now().Format("2006-01-02 15:04:05")
	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	res, err := tx.Exec(`INSERT INTO ss_order
		(order_no,user_id,project_id,order_type,is_rush_order,total_distance,account_count,pre_deduct,rush_order_fee,status,payment_status,accounts,result_data,created_at,source,agent_uid)
		VALUES (?,?,?,?,?,?,?,?,?,'pending','pre_deducted',?,?,?, ?, ?)`,
		orderNo, uid, project.ID, req.OrderType, boolInt(req.IsRushOrder), roundMoney(totalDistance), len(req.Accounts), preDeduct,
		func() float64 {
			if req.IsRushOrder {
				return project.RushFee
			}
			return 0
		}(), string(accountsRaw), string(resultRaw), now, source, agentUID)
	if err != nil {
		return nil, fmt.Errorf("保存订单失败: %w", err)
	}
	orderID, _ := res.LastInsertId()
	for _, a := range req.Accounts {
		_, err = tx.Exec(`INSERT INTO ss_accounts
			(order_id,order_no,user_id,project_id,account,password,distance,start_hour,start_minute,end_hour,end_minute,run_days,order_type,is_rush_order,status,created_at)
			VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,'pending',?)`,
			orderID, orderNo, uid, project.ID, a.Account, a.Password, a.Distance, a.StartHour, a.StartMinute, a.EndHour, a.EndMinute, a.RunDays, req.OrderType, boolInt(req.IsRushOrder), now)
		if err != nil {
			return nil, fmt.Errorf("保存账号明细失败: %w", err)
		}
	}
	if preDeduct > 0 {
		res, err = tx.Exec("UPDATE qingka_wangke_user SET money=money-? WHERE uid=? AND money>=?", preDeduct, uid, preDeduct)
		if err != nil {
			return nil, err
		}
		if affected, _ := res.RowsAffected(); affected <= 0 {
			return nil, fmt.Errorf("余额不足")
		}
		_, _ = tx.Exec("INSERT INTO qingka_wangke_moneylog (uid,type,money,mark,addtime) VALUES (?,?,?,?,?)",
			uid, "鲨兽下单", -preDeduct, fmt.Sprintf("鲨兽运动下单，订单号:%s，预扣%.2f", orderNo, preDeduct), now)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return map[string]any{"id": orderID, "order_no": orderNo, "pre_deduct": preDeduct, "message": "下单成功"}, nil
}

func (s *Service) QueryAccount(ctx context.Context, uid int, req QueryOrderRequest, source string, agentUID int) (map[string]any, error) {
	req.Account = strings.TrimSpace(req.Account)
	if req.Account == "" {
		return nil, fmt.Errorf("账号不能为空")
	}
	if req.QueryType != OrderTypeQueryMorning {
		req.QueryType = OrderTypeQueryNormal
	}
	project, err := s.getProject(req.ProjectID)
	if err != nil {
		return nil, err
	}
	fee := roundMoney(project.QueryFee * s.userRate(uid))
	if err := s.requireBalance(uid, fee, project.MinBalance); err != nil {
		return nil, err
	}
	upstream, err := s.upstreamQuery(ctx, project, req.Account, req.QueryType)
	if err != nil {
		return nil, err
	}
	orderNo := extractUpstreamOrderNo(upstream)
	if orderNo == "" {
		orderNo = fmt.Sprintf("SSQ%d%d", time.Now().Unix(), uid)
	}
	resultRaw, _ := json.Marshal(upstream)
	now := time.Now().Format("2006-01-02 15:04:05")
	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	res, err := tx.Exec(`INSERT INTO ss_order
		(order_no,user_id,project_id,order_type,pre_deduct,status,payment_status,query_account,result_data,created_at,source,agent_uid)
		VALUES (?,?,?,?,?,'pending','paid',?,?,?,?,?)`,
		orderNo, uid, project.ID, req.QueryType, fee, req.Account, string(resultRaw), now, source, agentUID)
	if err != nil {
		return nil, err
	}
	orderID, _ := res.LastInsertId()
	if fee > 0 {
		res, err = tx.Exec("UPDATE qingka_wangke_user SET money=money-? WHERE uid=? AND money>=?", fee, uid, fee)
		if err != nil {
			return nil, err
		}
		if affected, _ := res.RowsAffected(); affected <= 0 {
			return nil, fmt.Errorf("余额不足")
		}
		_, _ = tx.Exec("INSERT INTO qingka_wangke_moneylog (uid,type,money,mark,addtime) VALUES (?,?,?,?,?)",
			uid, "鲨兽查单", -fee, fmt.Sprintf("鲨兽运动查单，账号:%s，扣费%.2f", req.Account, fee), now)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return map[string]any{"id": orderID, "order_no": orderNo, "fee": fee, "result": upstream}, nil
}

func (s *Service) RefundAccount(ctx context.Context, uid int, req RefundOrderRequest, isAdmin bool, source string, agentUID int) (map[string]any, error) {
	acc, err := s.findAccount(uid, req.AccountID, req.Account, req.ProjectID, isAdmin)
	if err != nil {
		return nil, err
	}
	if acc.Status != "success" && acc.Status != "completed" {
		return nil, fmt.Errorf("只能退款成功的账号")
	}
	project, err := s.getProject(acc.ProjectID)
	if err != nil {
		return nil, err
	}
	if isRefundExpired(acc.CreatedAt, 90) {
		return nil, fmt.Errorf("该订单已超过90天退款期限，无法申请退款")
	}
	upstream, err := s.upstreamRefund(ctx, project, acc)
	if err != nil {
		return nil, err
	}
	resultRaw, _ := json.Marshal(upstream)
	now := time.Now().Format("2006-01-02 15:04:05")
	res, err := database.DB.Exec(`INSERT INTO ss_order
		(order_no,user_id,project_id,order_type,status,payment_status,refund_account,pre_deduct,result_data,refund_km,created_at,source,agent_uid)
		VALUES (?,?,?,?, 'pending','pending',?,0,?,?,?, ?, ?)`,
		acc.OrderNo, acc.UserID, acc.ProjectID, OrderTypeRefund, acc.Account, string(resultRaw), acc.Distance, now, source, agentUID)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	_, _ = database.DB.Exec("UPDATE ss_accounts SET status='refunding', updated_at=NOW() WHERE id=?", acc.ID)
	return map[string]any{"id": id, "message": "退款请求已提交", "result": upstream}, nil
}

func (s *Service) QueryStoredAccount(ctx context.Context, uid int, accountID int64, force bool, isAdmin bool, source string, agentUID int) (map[string]any, error) {
	acc, err := s.findAccount(uid, accountID, "", 0, isAdmin)
	if err != nil {
		return nil, err
	}
	if !force && jsonHasValue(acc.QueryResult) {
		return map[string]any{"from_cache": true, "account_id": acc.ID, "query_result": json.RawMessage(acc.QueryResult)}, nil
	}
	queryType := OrderTypeQueryNormal
	if acc.OrderType == OrderTypeMorning {
		queryType = OrderTypeQueryMorning
	}
	result, err := s.QueryAccount(ctx, acc.UserID, QueryOrderRequest{
		ProjectID: acc.ProjectID,
		QueryType: queryType,
		Account:   acc.Account,
	}, source, agentUID)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) CheckQueryStatus(uid int, accountID int64, isAdmin bool) (map[string]any, error) {
	acc, err := s.findAccount(uid, accountID, "", 0, isAdmin)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"account_id":   acc.ID,
		"has_result":   jsonHasValue(acc.QueryResult),
		"query_result": json.RawMessage(acc.QueryResult),
	}, nil
}

func (s *Service) ClearQueryResult(uid int, accountID int64, isAdmin bool) error {
	acc, err := s.findAccount(uid, accountID, "", 0, isAdmin)
	if err != nil {
		return err
	}
	_, err = database.DB.Exec("UPDATE ss_accounts SET query_result=NULL, updated_at=NOW() WHERE id=?", acc.ID)
	return err
}

func (s *Service) ListOrders(uid int, isAdmin bool, page, limit int, status, orderNo, account string, filterUserID int) ([]Order, int, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 200 {
		limit = 20
	}
	where := "WHERE 1=1"
	args := []any{}
	if !isAdmin {
		where += " AND o.user_id=?"
		args = append(args, uid)
	} else if filterUserID > 0 {
		where += " AND o.user_id=?"
		args = append(args, filterUserID)
	}
	if strings.TrimSpace(status) != "" {
		where += " AND o.status=?"
		args = append(args, strings.TrimSpace(status))
	}
	if strings.TrimSpace(orderNo) != "" {
		where += " AND o.order_no LIKE ?"
		args = append(args, "%"+strings.TrimSpace(orderNo)+"%")
	}
	if strings.TrimSpace(account) != "" {
		where += " AND (o.query_account LIKE ? OR o.refund_account LIKE ? OR o.id IN (SELECT DISTINCT order_id FROM ss_accounts WHERE account LIKE ?))"
		kw := "%" + strings.TrimSpace(account) + "%"
		args = append(args, kw, kw, kw)
	}
	var total int
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM ss_order o "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	query := `SELECT o.id, COALESCE(o.order_no,''), o.user_id, COALESCE(u.user,''), o.project_id, COALESCE(p.name,'鲨兽运动世界'),
		o.order_type, COALESCE(o.is_rush_order,0), COALESCE(o.total_distance,0), COALESCE(o.account_count,0), COALESCE(o.pre_deduct,0),
		o.actual_cost, o.final_charge, o.difference, COALESCE(o.rush_order_fee,0), COALESCE(o.status,''), COALESCE(o.payment_status,''),
		COALESCE(o.accounts, JSON_ARRAY()), COALESCE(o.query_account,''), COALESCE(o.refund_account,''), COALESCE(o.result_data, JSON_OBJECT()),
		COALESCE(DATE_FORMAT(o.created_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(o.completed_at,'%Y-%m-%d %H:%i:%s'),''),
		COALESCE(DATE_FORMAT(o.updated_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(o.error_message,''), o.refund_km, COALESCE(o.source,''), COALESCE(o.agent_uid,0)
		FROM ss_order o
		LEFT JOIN qingka_wangke_user u ON u.uid=o.user_id
		LEFT JOIN ss_project p ON p.id=o.project_id ` + where + ` ORDER BY o.id DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)
	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := []Order{}
	for rows.Next() {
		var o Order
		var actual, final, diff, refund sql.NullFloat64
		if err := rows.Scan(&o.ID, &o.OrderNo, &o.UserID, &o.Username, &o.ProjectID, &o.ProjectName, &o.OrderType, &o.IsRushOrder, &o.TotalDistance, &o.AccountCount, &o.PreDeduct, &actual, &final, &diff, &o.RushOrderFee, &o.Status, &o.PaymentStatus, &o.Accounts, &o.QueryAccount, &o.RefundAccount, &o.ResultData, &o.CreatedAt, &o.CompletedAt, &o.UpdatedAt, &o.ErrorMessage, &refund, &o.Source, &o.AgentUID); err == nil {
			if actual.Valid {
				v := actual.Float64
				o.ActualCost = &v
			}
			if final.Valid {
				v := final.Float64
				o.FinalCharge = &v
			}
			if diff.Valid {
				v := diff.Float64
				o.Difference = &v
			}
			if refund.Valid {
				v := refund.Float64
				o.RefundKM = &v
			}
			o.AccountDetails, _, _ = s.ListAccounts(uid, isAdmin, 1, 200, "", o.OrderNo, "", 0, 0)
			o = sanitizeOrderResponse(o, false)
			list = append(list, o)
		}
	}
	return list, total, nil
}

func (s *Service) ListAccounts(uid int, isAdmin bool, page, limit int, status, orderNo, account string, orderType, filterUserID int) ([]Account, int, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 500 {
		limit = 20
	}
	where := "WHERE 1=1"
	args := []any{}
	if !isAdmin {
		where += " AND a.user_id=?"
		args = append(args, uid)
	} else if filterUserID > 0 {
		where += " AND a.user_id=?"
		args = append(args, filterUserID)
	}
	if status != "" {
		where += " AND a.status=?"
		args = append(args, status)
	}
	if orderNo != "" {
		where += " AND a.order_no=?"
		args = append(args, orderNo)
	}
	if account != "" {
		where += " AND a.account LIKE ?"
		args = append(args, "%"+account+"%")
	}
	if orderType > 0 {
		where += " AND a.order_type=?"
		args = append(args, orderType)
	}
	var total int
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM ss_accounts a "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := database.DB.Query(`SELECT a.id, a.order_id, COALESCE(a.order_no,''), a.user_id, COALESCE(u.user,''), a.project_id,
		a.account, COALESCE(a.password,''), COALESCE(a.distance,0), COALESCE(a.start_hour,0), COALESCE(a.start_minute,0),
		COALESCE(a.end_hour,0), COALESCE(a.end_minute,0), COALESCE(a.run_days,''), a.order_type, COALESCE(a.is_rush_order,0),
		COALESCE(a.status,''), COALESCE(a.error_message,''), COALESCE(DATE_FORMAT(a.processed_at,'%Y-%m-%d %H:%i:%s'),''),
		COALESCE(a.query_result, JSON_OBJECT()), COALESCE(DATE_FORMAT(a.created_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(a.updated_at,'%Y-%m-%d %H:%i:%s'),'')
		FROM ss_accounts a LEFT JOIN qingka_wangke_user u ON u.uid=a.user_id `+where+` ORDER BY a.id DESC LIMIT ? OFFSET ?`,
		append(args, limit, (page-1)*limit)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := []Account{}
	for rows.Next() {
		var a Account
		if err := rows.Scan(&a.ID, &a.OrderID, &a.OrderNo, &a.UserID, &a.Username, &a.ProjectID, &a.Account, &a.Password, &a.Distance, &a.StartHour, &a.StartMinute, &a.EndHour, &a.EndMinute, &a.RunDays, &a.OrderType, &a.IsRushOrder, &a.Status, &a.ErrorMessage, &a.ProcessedAt, &a.QueryResult, &a.CreatedAt, &a.UpdatedAt); err == nil {
			a = sanitizeAccountResponse(a, false)
			list = append(list, a)
		}
	}
	return list, total, nil
}

func sanitizeOrderResponse(o Order, includeRaw bool) Order {
	if !includeRaw {
		o.Accounts = nil
		o.ResultData = nil
	}
	for i := range o.AccountDetails {
		o.AccountDetails[i] = sanitizeAccountResponse(o.AccountDetails[i], false)
	}
	return o
}

func sanitizeAccountResponse(a Account, includeQueryResult bool) Account {
	if !includeQueryResult {
		a.QueryResult = nil
	}
	return a
}

func (s *Service) SyncOrder(ctx context.Context, uid, id int, isAdmin bool) (map[string]any, error) {
	order, err := s.findOrder(uid, id, isAdmin)
	if err != nil {
		return nil, err
	}
	project, err := s.getProject(order.ProjectID)
	if err != nil {
		return nil, err
	}
	upstream, err := s.upstreamStatus(ctx, project, order)
	if err != nil {
		_, _ = database.DB.Exec("UPDATE ss_order SET error_message=?, updated_at=NOW() WHERE id=?", err.Error(), order.ID)
		return nil, err
	}
	if err := s.applySync(order, project, upstream); err != nil {
		return nil, err
	}
	return map[string]any{"message": "同步成功", "order_id": order.ID, "status": localSyncStatus(order, upstream)}, nil
}

func (s *Service) SyncPending(ctx context.Context, limit int) (int, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := database.DB.Query("SELECT id FROM ss_order WHERE status IN ('pending','processing','refunding') ORDER BY id ASC LIMIT ?", limit)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	ids := []int{}
	for rows.Next() {
		var id int
		if rows.Scan(&id) == nil {
			ids = append(ids, id)
		}
	}
	updated := 0
	for _, id := range ids {
		if _, err := s.SyncOrder(ctx, 0, id, true); err == nil {
			updated++
		}
	}
	return updated, nil
}

func (s *Service) applySync(order Order, project Project, payload map[string]any) error {
	localStatus := localSyncStatus(order, payload)
	actualRaw := firstNonNil(payload["actual_cost"], payload["actual"], nestedAny(payload, "data", "actual_cost"), nestedAny(payload, "data", "actual"), nestedAny(payload, "amounts", "actual"), nestedAny(payload, "data", "amounts", "actual"))
	hasActual := actualRaw != nil
	actual := asFloat(actualRaw)
	refundKM := asFloat(firstNonNil(payload["refund_km"], nestedAny(payload, "data", "refund_km"), nestedAny(payload, "refund_result", "refund_km"), nestedAny(payload, "data", "refund_result", "refund_km")))
	resultRaw, _ := json.Marshal(payload)
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	completeSQL := ""
	args := []any{localStatus, string(resultRaw), time.Now().Format("2006-01-02 15:04:05"), order.ID}
	if localStatus == "completed" || localStatus == "refunded" || localStatus == "failed" {
		completeSQL = ", completed_at=NOW()"
	}
	errMsg := orderSyncError(payload, localStatus)
	_, err = tx.Exec("UPDATE ss_order SET status=?, result_data=?, updated_at=?, error_message=NULLIF(?, '')"+completeSQL+" WHERE id=?", append(args[:3], errMsg, order.ID)...)
	if err != nil {
		return err
	}
	s.applyAccountSync(tx, order, payload, localStatus)
	if localStatus == "completed" {
		_, _ = tx.Exec("UPDATE ss_accounts SET status='success', processed_at=NOW(), updated_at=NOW() WHERE order_id=? AND status IN ('pending','processing')", order.ID)
	}
	if localStatus == "failed" {
		msg := firstString(payload, "message", "msg", "error")
		_, _ = tx.Exec("UPDATE ss_accounts SET status='failed', error_message=?, processed_at=NOW(), updated_at=NOW() WHERE order_id=? AND status IN ('pending','processing')", msg, order.ID)
	}
	if localStatus == "refunded" {
		if order.OrderType == OrderTypeRefund && strings.TrimSpace(order.RefundAccount) != "" {
			_, _ = tx.Exec("UPDATE ss_accounts SET status='refunded', processed_at=NOW(), updated_at=NOW() WHERE order_no=? AND account=? AND status IN ('success','completed','refunding')", order.OrderNo, order.RefundAccount)
		} else {
			_, _ = tx.Exec("UPDATE ss_accounts SET status='refunded', processed_at=NOW(), updated_at=NOW() WHERE order_id=?", order.ID)
		}
	}
	if noOrderRefund(payload) && order.OrderType == OrderTypeRefund {
		_, _ = tx.Exec("UPDATE ss_order SET status='completed', payment_status='no_refund', error_message=?, completed_at=NOW() WHERE id=?", upstreamMsg(payload, nil), order.ID)
		if strings.TrimSpace(order.RefundAccount) != "" {
			_, _ = tx.Exec("UPDATE ss_accounts SET status='success', error_message=?, updated_at=NOW() WHERE order_no=? AND account=?", upstreamMsg(payload, nil), order.OrderNo, order.RefundAccount)
		}
	}
	if order.OrderType == OrderTypeRefund && !noOrderRefund(payload) && (localStatus == "refunded" || refundSucceeded(payload)) {
		if err := s.settleRefund(tx, order, payload, refundKM); err != nil {
			return err
		}
	}
	if !canApplyActualSettlement(localStatus) {
		if err := s.repairPrematureSettlement(tx, order); err != nil {
			return err
		}
		if refundKM > 0 {
			_, _ = tx.Exec("UPDATE ss_order SET refund_km=? WHERE id=?", refundKM, order.ID)
		}
		return tx.Commit()
	}
	if hasActual && shouldSettleOrder(order) {
		var currentPay string
		var preDeduct float64
		var currentFinal sql.NullFloat64
		if err := tx.QueryRow("SELECT COALESCE(payment_status,''), COALESCE(pre_deduct,0), final_charge FROM ss_order WHERE id=? FOR UPDATE", order.ID).Scan(&currentPay, &preDeduct, &currentFinal); err != nil {
			return err
		}
		finalCharge := s.finalCharge(order.UserID, project, actual)
		if !canSettlePayment(currentPay) {
			if strings.EqualFold(currentPay, "settled") && currentFinal.Valid && roundMoney(currentFinal.Float64) != finalCharge {
				if err := s.correctSettledCharge(tx, order, preDeduct, roundMoney(currentFinal.Float64), actual, finalCharge); err != nil {
					return err
				}
			}
			if refundKM > 0 {
				_, _ = tx.Exec("UPDATE ss_order SET refund_km=? WHERE id=?", refundKM, order.ID)
			}
			return tx.Commit()
		}
		diff := roundMoney(finalCharge - preDeduct)
		payStatus := "settled"
		if diff > 0 {
			payStatus = "insufficient"
			res, err := tx.Exec("UPDATE qingka_wangke_user SET money=money-? WHERE uid=? AND money>=?", diff, order.UserID, diff)
			if err != nil {
				return err
			}
			if affected, _ := res.RowsAffected(); affected <= 0 {
				payStatus = "insufficient"
			} else {
				_, _ = tx.Exec("INSERT INTO qingka_wangke_moneylog (uid,type,money,mark,addtime) VALUES (?,?,?,?,NOW())",
					order.UserID, "鲨兽补扣", -diff, fmt.Sprintf("鲨兽运动补扣，订单号:%s，补扣%.2f", order.OrderNo, diff))
				payStatus = "settled"
			}
		}
		if diff < 0 {
			refund := math.Abs(diff)
			_, _ = tx.Exec("UPDATE qingka_wangke_user SET money=money+? WHERE uid=?", refund, order.UserID)
			_, _ = tx.Exec("INSERT INTO qingka_wangke_moneylog (uid,type,money,mark,addtime) VALUES (?,?,?,?,NOW())",
				order.UserID, "鲨兽退差", refund, fmt.Sprintf("鲨兽运动退差，订单号:%s，退还%.2f", order.OrderNo, refund))
		}
		_, _ = tx.Exec("UPDATE ss_order SET actual_cost=?, final_charge=?, difference=?, payment_status=? WHERE id=?", actual, finalCharge, diff, payStatus, order.ID)
	}
	if !hasActual && localStatus == "completed" && shouldSettleOrder(order) {
		var currentPay string
		var preDeduct float64
		if err := tx.QueryRow("SELECT COALESCE(payment_status,''), COALESCE(pre_deduct,0) FROM ss_order WHERE id=? FOR UPDATE", order.ID).Scan(&currentPay, &preDeduct); err != nil {
			return err
		}
		if canSettlePayment(currentPay) {
			_, _ = tx.Exec("UPDATE ss_order SET actual_cost=?, final_charge=?, difference=0, payment_status='settled' WHERE id=?", preDeduct, preDeduct, order.ID)
		}
	}
	if refundKM > 0 {
		_, _ = tx.Exec("UPDATE ss_order SET refund_km=? WHERE id=?", refundKM, order.ID)
	}
	return tx.Commit()
}

func localSyncStatus(order Order, payload map[string]any) string {
	status := strings.TrimSpace(firstString(payload, "status", "order_status"))
	if status == "" {
		status = strings.TrimSpace(asString(firstNonNil(
			nestedAny(payload, "data", "status"),
			nestedAny(payload, "data", "order_status"),
		)))
	}
	localStatus := mapStatus(status)
	if localStatus == "" {
		localStatus = order.Status
	}
	if detailStatus := aggregateDetailStatus(payload); detailStatus != "" {
		localStatus = detailStatus
	}
	return localStatus
}

func (s *Service) correctSettledCharge(tx *sql.Tx, order Order, preDeduct, currentFinal, actualCost, desiredFinal float64) error {
	delta := roundMoney(desiredFinal - currentFinal)
	if delta > 0 {
		res, err := tx.Exec("UPDATE qingka_wangke_user SET money=money-? WHERE uid=? AND money>=?", delta, order.UserID, delta)
		if err != nil {
			return err
		}
		if affected, _ := res.RowsAffected(); affected <= 0 {
			_, _ = tx.Exec("UPDATE ss_order SET payment_status='insufficient' WHERE id=?", order.ID)
			return nil
		}
		_, _ = tx.Exec("INSERT INTO qingka_wangke_moneylog (uid,type,money,mark,addtime) VALUES (?,?,?,?,NOW())",
			order.UserID, "鲨兽补扣修正", -delta, fmt.Sprintf("鲨兽运动结算修正，订单号:%s，补扣%.2f", order.OrderNo, delta))
	}
	if delta < 0 {
		refund := math.Abs(delta)
		_, _ = tx.Exec("UPDATE qingka_wangke_user SET money=money+? WHERE uid=?", refund, order.UserID)
		_, _ = tx.Exec("INSERT INTO qingka_wangke_moneylog (uid,type,money,mark,addtime) VALUES (?,?,?,?,NOW())",
			order.UserID, "鲨兽退差修正", refund, fmt.Sprintf("鲨兽运动结算修正，订单号:%s，退还%.2f", order.OrderNo, refund))
	}
	diff := roundMoney(desiredFinal - preDeduct)
	_, _ = tx.Exec("UPDATE ss_order SET actual_cost=?, final_charge=?, difference=?, payment_status='settled' WHERE id=?", actualCost, desiredFinal, diff, order.ID)
	return nil
}

func (s *Service) settleRefund(tx *sql.Tx, order Order, payload map[string]any, refundKM float64) error {
	var currentPay string
	if err := tx.QueryRow("SELECT COALESCE(payment_status,'') FROM ss_order WHERE id=? FOR UPDATE", order.ID).Scan(&currentPay); err != nil {
		return err
	}
	if strings.EqualFold(currentPay, "refunded") {
		return nil
	}
	amount := refundAmount(payload)
	if amount > 0 {
		_, _ = tx.Exec("UPDATE qingka_wangke_user SET money=money+? WHERE uid=?", amount, order.UserID)
		_, _ = tx.Exec("INSERT INTO qingka_wangke_moneylog (uid,type,money,mark,addtime) VALUES (?,?,?,?,NOW())",
			order.UserID, "鲨兽退款", amount, fmt.Sprintf("鲨兽运动退款，账号:%s，退还%.2f", order.RefundAccount, amount))
	}
	_, _ = tx.Exec(`UPDATE ss_order SET status='refunded', actual_cost=?, final_charge=?, difference=0,
		refund_km=IF(? > 0, ?, refund_km), payment_status='refunded', completed_at=NOW() WHERE id=?`,
		amount, amount, refundKM, refundKM, order.ID)
	if strings.TrimSpace(order.RefundAccount) != "" {
		_, _ = tx.Exec("UPDATE ss_accounts SET status='refunded', processed_at=NOW(), updated_at=NOW() WHERE order_no=? AND account=?", order.OrderNo, order.RefundAccount)
	}
	return nil
}

func (s *Service) applyAccountSync(tx *sql.Tx, order Order, payload map[string]any, orderStatus string) {
	for _, row := range accountDetailRows(payload) {
		account := firstString(row, "account", "username", "user")
		if account == "" {
			continue
		}
		status := mapStatus(firstString(row, "status", "order_status"))
		if status == "" {
			status = orderStatus
		}
		msg := firstString(row, "error_message", "message", "msg", "error", "detail")
		processedAt := firstString(row, "processed_at", "complete_time", "updated_at")
		if processedAt == "" {
			processedAt = time.Now().Format("2006-01-02 15:04:05")
		}
		if status != "" {
			_, _ = tx.Exec(`UPDATE ss_accounts SET status=?, error_message=NULLIF(?,''), processed_at=?, updated_at=NOW()
				WHERE order_no=? AND account=?`, status, msg, processedAt, order.OrderNo, account)
		}
		queryResult := firstNonNil(row["query_result"], row["result"], row["query"], row["data"])
		if queryResult != nil {
			raw, _ := json.Marshal(queryResult)
			if len(raw) > 0 && string(raw) != "null" {
				_, _ = tx.Exec("UPDATE ss_accounts SET query_result=?, updated_at=NOW() WHERE account=? ORDER BY id DESC LIMIT 1", string(raw), account)
			}
		}
	}
	if order.QueryAccount != "" {
		queryResult := firstNonNil(payload["query_result"], payload["result"], nestedAny(payload, "data", "query_result"), nestedAny(payload, "data", "result"))
		if queryResult != nil {
			raw, _ := json.Marshal(queryResult)
			if len(raw) > 0 && string(raw) != "null" {
				_, _ = tx.Exec("UPDATE ss_accounts SET query_result=?, updated_at=NOW() WHERE account=? ORDER BY id DESC LIMIT 1", string(raw), order.QueryAccount)
			}
		}
	}
}

func accountDetailRows(payload map[string]any) []map[string]any {
	rows := payloadRows(payload, "detail", "details", "accounts", "account_details")
	if len(rows) > 0 {
		return rows
	}
	if data, ok := payload["data"].(map[string]any); ok {
		return payloadRows(data, "detail", "details", "accounts", "account_details", "list")
	}
	return nil
}

func aggregateDetailStatus(payload map[string]any) string {
	rows := accountDetailRows(payload)
	if len(rows) == 0 {
		return ""
	}
	counts := map[string]int{}
	for _, row := range rows {
		status := mapStatus(firstString(row, "status", "order_status"))
		if status == "" {
			continue
		}
		counts[status]++
	}
	if counts["processing"] > 0 {
		return "processing"
	}
	if counts["pending"] == len(rows) {
		return "pending"
	}
	if counts["success"]+counts["completed"] == len(rows) {
		return "completed"
	}
	if counts["failed"] == len(rows) {
		return "failed"
	}
	if counts["refunded"] == len(rows) {
		return "refunded"
	}
	if counts["pending"] > 0 {
		return "processing"
	}
	return ""
}

func orderSyncError(payload map[string]any, status string) string {
	if status != "failed" {
		return ""
	}
	for _, row := range accountDetailRows(payload) {
		if mapStatus(firstString(row, "status", "order_status")) == "failed" {
			if msg := firstString(row, "error_message", "message", "msg", "error", "detail"); msg != "" {
				return msg
			}
		}
	}
	if msg := firstString(payload, "error_message", "message", "error", "detail"); msg != "" {
		return msg
	}
	if msg := firstString(mapFromAny(payload["data"]), "error_message", "message", "error", "detail"); msg != "" {
		return msg
	}
	return "订单失败"
}

func noOrderRefund(payload map[string]any) bool {
	for _, value := range []any{
		payload["refund_result"],
		nestedAny(payload, "data", "refund_result"),
		nestedAny(payload, "refund_result", "status"),
		nestedAny(payload, "data", "refund_result", "status"),
	} {
		if strings.EqualFold(strings.TrimSpace(asString(value)), "no_order") {
			return true
		}
	}
	return false
}

func refundSucceeded(payload map[string]any) bool {
	for _, value := range []any{
		payload["refund_result"],
		nestedAny(payload, "data", "refund_result"),
		nestedAny(payload, "refund_result", "status"),
		nestedAny(payload, "data", "refund_result", "status"),
	} {
		raw := strings.ToLower(strings.TrimSpace(asString(value)))
		if raw == "success" || raw == "completed" || raw == "complete" || raw == "refunded" || raw == "refund" {
			return true
		}
	}
	return refundAmount(payload) > 0
}

func refundAmount(payload map[string]any) float64 {
	return roundMoney(asFloat(firstNonNil(
		payload["refund_amount"],
		payload["refund"],
		nestedAny(payload, "amounts", "refund"),
		nestedAny(payload, "data", "amounts", "refund"),
		nestedAny(payload, "refund_result", "refund_amount"),
		nestedAny(payload, "data", "refund_result", "refund_amount"),
		nestedAny(payload, "refund_result", "amount"),
		nestedAny(payload, "data", "refund_result", "amount"),
	)))
}

func shouldSettleOrder(order Order) bool {
	return order.OrderType == OrderTypeNormal || order.OrderType == OrderTypeMorning
}

func canApplyActualSettlement(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "completed", "failed":
		return true
	default:
		return false
	}
}

func canSettlePayment(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "", "pre_deducted", "partial_refund", "insufficient":
		return true
	default:
		return false
	}
}

func (s *Service) finalCharge(uid int, project Project, actual float64) float64 {
	rate := project.ActualRate
	if rate <= 0 {
		rate = 1
	}
	return roundMoney(actual * rate * s.userRate(uid))
}

func (s *Service) repairPrematureSettlement(tx *sql.Tx, order Order) error {
	if !shouldSettleOrder(order) || order.PreDeduct <= 0 {
		return nil
	}
	var payment string
	var preDeduct float64
	var actual, finalCharge, difference sql.NullFloat64
	if err := tx.QueryRow(`SELECT COALESCE(payment_status,''), COALESCE(pre_deduct,0), actual_cost, final_charge, difference
		FROM ss_order WHERE id=? FOR UPDATE`, order.ID).Scan(&payment, &preDeduct, &actual, &finalCharge, &difference); err != nil {
		return err
	}
	if !strings.EqualFold(payment, "settled") || preDeduct <= 0 {
		return nil
	}
	if !actual.Valid || !finalCharge.Valid || finalCharge.Float64 > 0 || (difference.Valid && difference.Float64 >= 0) {
		return nil
	}
	res, err := tx.Exec("UPDATE qingka_wangke_user SET money=money-? WHERE uid=? AND money>=?", preDeduct, order.UserID, preDeduct)
	if err != nil {
		return err
	}
	if affected, _ := res.RowsAffected(); affected <= 0 {
		_, _ = tx.Exec("UPDATE ss_order SET payment_status='insufficient' WHERE id=?", order.ID)
		return nil
	}
	_, _ = tx.Exec("INSERT INTO qingka_wangke_moneylog (uid,type,money,mark,addtime) VALUES (?,?,?,?,NOW())",
		order.UserID, "鲨兽恢复预扣", -preDeduct, fmt.Sprintf("鲨兽运动同步修正，订单号:%s，恢复预扣%.2f", order.OrderNo, preDeduct))
	_, _ = tx.Exec("UPDATE ss_order SET actual_cost=NULL, final_charge=NULL, difference=NULL, payment_status='pre_deducted' WHERE id=?", order.ID)
	return nil
}

func isRefundExpired(createdAt string, days int) bool {
	if days <= 0 || strings.TrimSpace(createdAt) == "" {
		return false
	}
	t, err := time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(createdAt), time.Local)
	if err != nil {
		return false
	}
	return time.Since(t) > time.Duration(days)*24*time.Hour
}

func jsonHasValue(raw json.RawMessage) bool {
	text := strings.TrimSpace(string(raw))
	return text != "" && text != "null" && text != "{}" && text != "[]"
}

func (s *Service) requireBalance(uid int, amount, minBalance float64) error {
	var balance float64
	if err := database.DB.QueryRow("SELECT COALESCE(money,0) FROM qingka_wangke_user WHERE uid=?", uid).Scan(&balance); err != nil {
		return fmt.Errorf("用户不存在")
	}
	if minBalance > 0 && balance < minBalance {
		return fmt.Errorf("余额不足，当前余额 ¥%.2f，最低余额要求 ¥%.2f", balance, minBalance)
	}
	if amount > 0 && balance < amount {
		return fmt.Errorf("余额不足，需要 %.2f 元", amount)
	}
	return nil
}

func (s *Service) findOrder(uid, id int, isAdmin bool) (Order, error) {
	if id <= 0 {
		return Order{}, fmt.Errorf("订单不存在或无权操作")
	}
	where := "WHERE o.id=?"
	args := []any{id}
	if !isAdmin {
		where += " AND o.user_id=?"
		args = append(args, uid)
	}
	rows, err := database.DB.Query(`SELECT o.id, COALESCE(o.order_no,''), o.user_id, COALESCE(u.user,''), o.project_id, COALESCE(p.name,'鲨兽运动世界'),
			o.order_type, COALESCE(o.is_rush_order,0), COALESCE(o.total_distance,0), COALESCE(o.account_count,0), COALESCE(o.pre_deduct,0),
			o.actual_cost, o.final_charge, o.difference, COALESCE(o.rush_order_fee,0), COALESCE(o.status,''), COALESCE(o.payment_status,''),
			COALESCE(o.accounts, JSON_ARRAY()), COALESCE(o.query_account,''), COALESCE(o.refund_account,''), COALESCE(o.result_data, JSON_OBJECT()),
			COALESCE(DATE_FORMAT(o.created_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(o.completed_at,'%Y-%m-%d %H:%i:%s'),''),
			COALESCE(DATE_FORMAT(o.updated_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(o.error_message,''), o.refund_km, COALESCE(o.source,''), COALESCE(o.agent_uid,0)
			FROM ss_order o LEFT JOIN qingka_wangke_user u ON u.uid=o.user_id LEFT JOIN ss_project p ON p.id=o.project_id `+where+` LIMIT 1`, args...)
	if err != nil {
		return Order{}, err
	}
	defer rows.Close()
	if rows.Next() {
		var o Order
		var actual, final, diff, refund sql.NullFloat64
		err = rows.Scan(&o.ID, &o.OrderNo, &o.UserID, &o.Username, &o.ProjectID, &o.ProjectName, &o.OrderType, &o.IsRushOrder, &o.TotalDistance, &o.AccountCount, &o.PreDeduct, &actual, &final, &diff, &o.RushOrderFee, &o.Status, &o.PaymentStatus, &o.Accounts, &o.QueryAccount, &o.RefundAccount, &o.ResultData, &o.CreatedAt, &o.CompletedAt, &o.UpdatedAt, &o.ErrorMessage, &refund, &o.Source, &o.AgentUID)
		if actual.Valid {
			v := actual.Float64
			o.ActualCost = &v
		}
		if final.Valid {
			v := final.Float64
			o.FinalCharge = &v
		}
		if diff.Valid {
			v := diff.Float64
			o.Difference = &v
		}
		if refund.Valid {
			v := refund.Float64
			o.RefundKM = &v
		}
		return o, err
	}
	return Order{}, fmt.Errorf("订单不存在或无权操作")
}

func (s *Service) findAccount(uid int, id int64, account string, projectID int, isAdmin bool) (Account, error) {
	where := "WHERE 1=1"
	args := []any{}
	if id > 0 {
		where += " AND a.id=?"
		args = append(args, id)
	} else if strings.TrimSpace(account) != "" {
		where += " AND a.account=?"
		args = append(args, strings.TrimSpace(account))
		if projectID > 0 {
			where += " AND a.project_id=?"
			args = append(args, projectID)
		}
	} else {
		return Account{}, fmt.Errorf("账号ID或账号不能为空")
	}
	if !isAdmin {
		where += " AND a.user_id=?"
		args = append(args, uid)
	}
	rows, err := database.DB.Query(`SELECT a.id, a.order_id, COALESCE(a.order_no,''), a.user_id, COALESCE(u.user,''), a.project_id,
		a.account, COALESCE(a.password,''), COALESCE(a.distance,0), COALESCE(a.start_hour,0), COALESCE(a.start_minute,0),
		COALESCE(a.end_hour,0), COALESCE(a.end_minute,0), COALESCE(a.run_days,''), a.order_type, COALESCE(a.is_rush_order,0),
		COALESCE(a.status,''), COALESCE(a.error_message,''), COALESCE(DATE_FORMAT(a.processed_at,'%Y-%m-%d %H:%i:%s'),''),
		COALESCE(a.query_result, JSON_OBJECT()), COALESCE(DATE_FORMAT(a.created_at,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(a.updated_at,'%Y-%m-%d %H:%i:%s'),'')
		FROM ss_accounts a LEFT JOIN qingka_wangke_user u ON u.uid=a.user_id `+where+` ORDER BY a.id DESC LIMIT 1`, args...)
	if err != nil {
		return Account{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return Account{}, fmt.Errorf("账号不存在或无权操作")
	}
	var a Account
	err = rows.Scan(&a.ID, &a.OrderID, &a.OrderNo, &a.UserID, &a.Username, &a.ProjectID, &a.Account, &a.Password, &a.Distance, &a.StartHour, &a.StartMinute, &a.EndHour, &a.EndMinute, &a.RunDays, &a.OrderType, &a.IsRushOrder, &a.Status, &a.ErrorMessage, &a.ProcessedAt, &a.QueryResult, &a.CreatedAt, &a.UpdatedAt)
	return a, err
}

func validateCreateOrder(req CreateOrderRequest) error {
	if req.ProjectID <= 0 {
		return fmt.Errorf("请选择项目")
	}
	if req.OrderType != OrderTypeNormal && req.OrderType != OrderTypeMorning {
		return fmt.Errorf("订单类型无效")
	}
	if len(req.Accounts) == 0 {
		return fmt.Errorf("请至少填写一个账号")
	}
	for i, a := range req.Accounts {
		if strings.TrimSpace(a.Account) == "" {
			return fmt.Errorf("第%d个账号不能为空", i+1)
		}
		if strings.TrimSpace(a.Password) == "" {
			return fmt.Errorf("第%d个密码不能为空", i+1)
		}
		if a.Distance <= 0 {
			return fmt.Errorf("第%d个公里数必须大于0", i+1)
		}
		if a.StartHour < 0 || a.StartHour > 23 || a.EndHour < 0 || a.EndHour > 23 || a.StartMinute < 0 || a.StartMinute > 59 || a.EndMinute < 0 || a.EndMinute > 59 {
			return fmt.Errorf("第%d个时间段无效", i+1)
		}
		if strings.TrimSpace(a.RunDays) == "" {
			return fmt.Errorf("第%d个跑步周期不能为空", i+1)
		}
	}
	return nil
}

func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func mapStatus(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "pending", "wait", "waiting", "queued", "queue", "待处理", "等待处理", "待提交", "排队中", "0", "1":
		return "pending"
	case "processing", "running", "doing", "处理中", "运行中", "执行中", "2":
		return "processing"
	case "completed", "complete", "success", "已完成", "完成", "成功", "3":
		return "completed"
	case "refunded", "refund", "已退款", "退款成功", "4":
		return "refunded"
	case "failed", "fail", "error", "失败", "错误", "-1", "5":
		return "failed"
	default:
		return raw
	}
}

func asString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case json.Number:
		return t.String()
	case nil:
		return ""
	default:
		return fmt.Sprint(t)
	}
}

func asFloat(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case json.Number:
		n, _ := t.Float64()
		return n
	case string:
		n, _ := strconv.ParseFloat(strings.TrimSpace(t), 64)
		return n
	default:
		return 0
	}
}

func nestedAny(m map[string]any, keys ...string) any {
	var cur any = m
	for _, key := range keys {
		next, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur = next[key]
	}
	return cur
}

func mapFromAny(v any) map[string]any {
	if row, ok := v.(map[string]any); ok {
		return row
	}
	return nil
}

func payloadRows(payload map[string]any, keys ...string) []map[string]any {
	for _, key := range keys {
		value, ok := payload[key]
		if !ok {
			continue
		}
		rows := rowsFromAny(value)
		if len(rows) > 0 {
			return rows
		}
	}
	return nil
}

func rowsFromAny(value any) []map[string]any {
	switch typed := value.(type) {
	case []any:
		rows := make([]map[string]any, 0, len(typed))
		for _, item := range typed {
			if row, ok := item.(map[string]any); ok {
				rows = append(rows, row)
			}
		}
		return rows
	case map[string]any:
		return payloadRows(typed, "detail", "details", "accounts", "account_details", "data", "list", "orders")
	default:
		return nil
	}
}

func firstString(m map[string]any, keys ...string) string {
	for _, key := range keys {
		if v := strings.TrimSpace(asString(m[key])); v != "" && v != "<nil>" {
			return v
		}
	}
	return ""
}

func firstNonNil(values ...any) any {
	for _, v := range values {
		if v != nil {
			return v
		}
	}
	return nil
}

func extractUpstreamOrderNo(payload map[string]any) string {
	for _, value := range []any{
		payload["order_no"],
		payload["orderNo"],
		nestedAny(payload, "data", "order_no"),
		nestedAny(payload, "data", "orderNo"),
		nestedAny(payload, "data", "order", "order_no"),
		nestedAny(payload, "data", "order", "orderNo"),
		payload["order_id"],
		payload["id"],
		nestedAny(payload, "data", "order_id"),
		nestedAny(payload, "data", "id"),
	} {
		if text := strings.TrimSpace(asString(value)); text != "" && text != "<nil>" {
			return text
		}
	}
	return ""
}

func upstreamOrderID(order Order) string {
	var payload map[string]any
	_ = json.Unmarshal(order.ResultData, &payload)
	if id := extractUpstreamOrderID(payload); isPositiveIntString(id) {
		return id
	}
	if id := strings.TrimSpace(asString(nestedAny(payload, "data", "order", "id"))); isPositiveIntString(id) {
		return id
	}
	if id := strings.TrimSpace(asString(nestedAny(payload, "data", "order", "order_id"))); isPositiveIntString(id) {
		return id
	}
	return ""
}

func extractUpstreamOrderID(payload map[string]any) string {
	for _, value := range []any{
		payload["upstream_order_id"],
		nestedAny(payload, "data", "order_id"),
		nestedAny(payload, "data", "id"),
		payload["order_id"],
		payload["id"],
	} {
		if text := strings.TrimSpace(asString(value)); isPositiveIntString(text) {
			return text
		}
	}
	return ""
}

func isPositiveIntString(raw string) bool {
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	return err == nil && n > 0
}
