package jiguang

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-api/internal/database"
)

type Service struct {
	client *http.Client
	locks  sync.Map
}

var service = &Service{client: &http.Client{Timeout: 30 * time.Second}}

func Jiguang() *Service {
	return service
}

type OrderRequest struct {
	ProductID       int     `json:"product_id"`
	SchoolName      string  `json:"school_name"`
	StudentName     string  `json:"student_name"`
	StudentAccount  string  `json:"student_account"`
	Times           int     `json:"times"`
	KMPerDay        float64 `json:"km_per_day"`
	CustomerMessage string  `json:"customer_message"`
}

type Order struct {
	ID              int      `json:"id"`
	UID             int      `json:"uid"`
	Username        string   `json:"username,omitempty"`
	OrderNo         string   `json:"order_no"`
	UpstreamID      int      `json:"upstream_id"`
	ProductID       int      `json:"product_id"`
	ProductName     string   `json:"product_name"`
	SchoolName      string   `json:"school_name"`
	StudentName     string   `json:"student_name"`
	StudentAccount  string   `json:"student_account"`
	RunTimes        int      `json:"run_times"`
	CompletedTimes  int      `json:"completed_times"`
	KMPerDay        float64  `json:"km_per_day"`
	CustomerMessage string   `json:"customer_message"`
	Status          string   `json:"status"`
	Fees            float64  `json:"fees"`
	RefundAmount    *float64 `json:"refund_amount"`
	Notes           string   `json:"notes"`
	Source          string   `json:"source"`
	AgentUID        int      `json:"agent_uid"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
	RefundedAt      string   `json:"refunded_at"`
}

type Product struct {
	ProductID int     `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	BasePrice float64 `json:"base_price"`
}

type RefundPreview struct {
	OrderNo        string  `json:"order_no"`
	ProductName    string  `json:"product_name"`
	StudentAccount string  `json:"student_account"`
	Remaining      int     `json:"remaining"`
	KMPerDay       float64 `json:"km_per_day"`
	PricePerKM     float64 `json:"price_per_km"`
	Amount         float64 `json:"amount"`
}

type AddTimesPreview struct {
	OrderNo        string  `json:"order_no"`
	ProductName    string  `json:"product_name"`
	StudentAccount string  `json:"student_account"`
	Delta          int     `json:"delta"`
	KMPerDay       float64 `json:"km_per_day"`
	PricePerKM     float64 `json:"price_per_km"`
	Cost           float64 `json:"cost"`
	BeforeRunTimes int     `json:"before_run_times"`
	AfterRunTimes  int     `json:"after_run_times"`
}

func (s *Service) loadConfig() (Config, error) {
	var raw string
	err := database.DB.QueryRow("SELECT k FROM qingka_wangke_config WHERE v = ? LIMIT 1", "jiguang_config").Scan(&raw)
	if err != nil {
		return defaultConfig(), nil
	}
	cfg, err := parseConfig(raw)
	if err != nil {
		return defaultConfig(), nil
	}
	return cfg, nil
}

func (s *Service) saveConfig(cfg Config) error {
	raw, err := cfg.Marshal()
	if err != nil {
		return err
	}
	_, err = database.DB.Exec(
		"INSERT INTO qingka_wangke_config (v, k) VALUES (?, ?) ON DUPLICATE KEY UPDATE k = VALUES(k)",
		"jiguang_config", raw,
	)
	return err
}

func (s *Service) Products(uid int) ([]Product, error) {
	cfg, err := s.loadConfig()
	if err != nil {
		return nil, err
	}
	list := []Product{}
	for _, id := range []int{ProductMorning, ProductDaily} {
		price, name, err := s.productPrice(uid, cfg, id)
		if err != nil {
			price = 0
		}
		if strings.TrimSpace(name) == "" {
			name = productName(id)
		}
		list = append(list, Product{ProductID: id, Name: name, Price: price, BasePrice: round2(productBasePrice(cfg, id))})
	}
	return list, nil
}

func (s *Service) ProductPrices(uid int) (map[int]float64, error) {
	products, err := s.Products(uid)
	if err != nil {
		return nil, err
	}
	prices := map[int]float64{}
	for _, item := range products {
		prices[item.ProductID] = item.Price
	}
	return prices, nil
}

func (s *Service) Schools(ctx context.Context, page, pageSize int, keyword string) (map[string]any, error) {
	cfg, err := s.loadConfig()
	if err != nil {
		return nil, err
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 500 {
		pageSize = 500
	}
	payload, err := s.upstreamSchools(ctx, cfg, page, pageSize, keyword)
	if err != nil {
		return nil, err
	}
	return ensureMapData(payload), nil
}

func (s *Service) CreateOrder(ctx context.Context, uid int, req OrderRequest, source string, agentUID int) (map[string]any, error) {
	req = normalizeOrderRequest(req)
	if err := validateOrderRequest(req); err != nil {
		return nil, err
	}
	cfg, err := s.loadConfig()
	if err != nil {
		return nil, err
	}
	unitPrice, productLabel, err := s.productPrice(uid, cfg, req.ProductID)
	if err != nil {
		return nil, err
	}
	totalPrice := round2(unitPrice * float64(req.Times) * req.KMPerDay)
	if totalPrice < 0 {
		return nil, fmt.Errorf("价格异常")
	}
	if err := s.requireBalance(uid, totalPrice); err != nil {
		return nil, err
	}
	upstream, err := s.upstreamCreate(ctx, cfg, req)
	if err != nil {
		return nil, err
	}
	orderNo := extractOrderNo(upstream)
	if orderNo == "" {
		return nil, fmt.Errorf("获取订单号失败")
	}
	upstreamID := extractOrderID(upstream)
	if name := extractProductName(upstream); name != "" {
		productLabel = name
	}
	status := extractStatus(upstream, "pending")
	now := time.Now().Format("2006-01-02 15:04:05")

	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	res, err := tx.Exec(`INSERT INTO qingka_wangke_jiguang
		(uid, order_no, upstream_id, product_id, product_name, school_name, student_name, student_account,
		 run_times, completed_times, km_per_day, customer_message, status, fees, source, agent_uid, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		uid, orderNo, upstreamID, req.ProductID, productLabel, req.SchoolName, req.StudentName, req.StudentAccount,
		req.Times, 0, req.KMPerDay, req.CustomerMessage, status, totalPrice, source, agentUID, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("保存订单失败: %w", err)
	}
	orderID, _ := res.LastInsertId()
	if totalPrice > 0 {
		res, err = tx.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ? AND money >= ?", totalPrice, uid, totalPrice)
		if err != nil {
			return nil, fmt.Errorf("扣除余额失败: %w", err)
		}
		if affected, _ := res.RowsAffected(); affected <= 0 {
			return nil, fmt.Errorf("余额不足，请先充值")
		}
		_, _ = tx.Exec("INSERT INTO qingka_wangke_moneylog (uid,type,money,mark,addtime) VALUES (?,?,?,?,?)",
			uid, "jiguang_add", -totalPrice, fmt.Sprintf("极光跑步下单，订单号:%s，扣除%.2f", orderNo, totalPrice), now)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return map[string]any{"id": orderID, "order_no": orderNo, "upstream_id": upstreamID, "total_price": totalPrice, "message": "下单成功"}, nil
}

func (s *Service) ListOrders(uid int, isAdmin bool, page, limit int, searchType, keyword, status, school string, filterUID int) ([]Order, int, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 200 {
		limit = 20
	}
	where := "WHERE 1=1"
	args := []any{}
	if !isAdmin {
		where += " AND j.uid=?"
		args = append(args, uid)
	} else if filterUID > 0 {
		where += " AND j.uid=?"
		args = append(args, filterUID)
	}
	if strings.TrimSpace(status) != "" {
		where += " AND j.status=?"
		args = append(args, strings.TrimSpace(status))
	}
	if strings.TrimSpace(school) != "" {
		where += " AND j.school_name LIKE ?"
		args = append(args, "%"+strings.TrimSpace(school)+"%")
	}
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		switch searchType {
		case "1", "id":
			where += " AND j.id = ?"
			args = append(args, keyword)
		case "2", "student_account":
			where += " AND j.student_account LIKE ?"
			args = append(args, "%"+keyword+"%")
		case "3", "uid":
			if isAdmin {
				where += " AND j.uid = ?"
				args = append(args, keyword)
			}
		default:
			where += " AND (j.order_no LIKE ? OR j.student_account LIKE ? OR j.student_name LIKE ? OR j.school_name LIKE ?)"
			args = append(args, "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
		}
	}
	var total int
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_jiguang j "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	query := `SELECT j.id, j.uid, COALESCE(u.user,''), j.order_no, COALESCE(j.upstream_id,0), j.product_id, j.product_name,
		j.school_name, j.student_name, j.student_account, j.run_times, j.completed_times, j.km_per_day,
		j.customer_message, j.status, COALESCE(j.fees,0), j.refund_amount, COALESCE(j.notes,''),
		COALESCE(j.source,''), COALESCE(j.agent_uid,0),
		COALESCE(DATE_FORMAT(j.created_at,'%Y-%m-%d %H:%i:%s'),''),
		COALESCE(DATE_FORMAT(j.updated_at,'%Y-%m-%d %H:%i:%s'),''),
		COALESCE(DATE_FORMAT(j.refunded_at,'%Y-%m-%d %H:%i:%s'),'')
		FROM qingka_wangke_jiguang j LEFT JOIN qingka_wangke_user u ON u.uid=j.uid ` + where + ` ORDER BY j.id DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)
	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := []Order{}
	for rows.Next() {
		var o Order
		var refund sql.NullFloat64
		if err := rows.Scan(&o.ID, &o.UID, &o.Username, &o.OrderNo, &o.UpstreamID, &o.ProductID, &o.ProductName,
			&o.SchoolName, &o.StudentName, &o.StudentAccount, &o.RunTimes, &o.CompletedTimes, &o.KMPerDay,
			&o.CustomerMessage, &o.Status, &o.Fees, &refund, &o.Notes, &o.Source, &o.AgentUID,
			&o.CreatedAt, &o.UpdatedAt, &o.RefundedAt); err != nil {
			return nil, 0, err
		}
		if refund.Valid {
			value := refund.Float64
			o.RefundAmount = &value
		}
		list = append(list, o)
	}
	return list, total, nil
}

func (s *Service) RefundOrder(ctx context.Context, uid int, orderNo string, isAdmin bool, confirm bool) (map[string]any, error) {
	order, err := s.findOrder(uid, 0, orderNo, isAdmin)
	if err != nil {
		return nil, err
	}
	if order.Status == "refunded" {
		return nil, fmt.Errorf("该订单已退款")
	}
	cfg, err := s.loadConfig()
	if err != nil {
		return nil, err
	}
	preview, err := s.refundPreview(order, cfg)
	if err != nil {
		return nil, err
	}
	lockDone, err := s.lockAction("refund:" + order.OrderNo)
	if err != nil {
		return nil, err
	}
	if confirm {
		defer lockDone()
	} else {
		lockDone()
	}
	if _, err := s.upstreamRefund(ctx, cfg, order, !confirm); err != nil {
		return nil, err
	}
	if !confirm {
		return map[string]any{"item": preview}, nil
	}
	now := time.Now().Format("2006-01-02 15:04:05")
	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	res, err := tx.Exec(`UPDATE qingka_wangke_jiguang
		SET status='refunded', refund_amount=?, refunded_at=?, updated_at=?
		WHERE id=? AND status!='refunded'`, preview.Amount, now, now, order.ID)
	if err != nil {
		return nil, err
	}
	if affected, _ := res.RowsAffected(); affected > 0 {
		if preview.Amount > 0 {
			if _, err := tx.Exec("UPDATE qingka_wangke_user SET money = money + ? WHERE uid = ?", preview.Amount, order.UID); err != nil {
				return nil, err
			}
		}
		_, _ = tx.Exec("INSERT INTO qingka_wangke_moneylog (uid,type,money,mark,addtime) VALUES (?,?,?,?,?)",
			order.UID, "jiguang_refund", preview.Amount,
			fmt.Sprintf("极光跑步退款，订单号:%s，剩余%d次，退款%.2f", order.OrderNo, preview.Remaining, preview.Amount), now)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return map[string]any{"item": preview, "message": fmt.Sprintf("已退款 %.2f 元", preview.Amount)}, nil
}

func (s *Service) AddTimes(ctx context.Context, uid int, orderNo string, delta int, isAdmin bool, confirm bool) (map[string]any, error) {
	if delta <= 0 || delta > 9999 {
		return nil, fmt.Errorf("加次数必须为 1~9999")
	}
	order, err := s.findOrder(uid, 0, orderNo, isAdmin)
	if err != nil {
		return nil, err
	}
	if order.Status == "refunded" {
		return nil, fmt.Errorf("已退款订单无法加次数")
	}
	cfg, err := s.loadConfig()
	if err != nil {
		return nil, err
	}
	preview, err := s.addTimesPreview(order, cfg, delta)
	if err != nil {
		return nil, err
	}
	if err := s.requireBalance(order.UID, preview.Cost); err != nil {
		return nil, fmt.Errorf("订单所属用户余额不足")
	}
	lockDone, err := s.lockAction("addtimes:" + order.OrderNo)
	if err != nil {
		return nil, err
	}
	if confirm {
		defer lockDone()
	} else {
		lockDone()
	}
	if _, err := s.upstreamAddTimes(ctx, cfg, order, delta, !confirm); err != nil {
		return nil, err
	}
	if !confirm {
		return map[string]any{"item": preview}, nil
	}
	now := time.Now().Format("2006-01-02 15:04:05")
	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	newStatus := order.Status
	if newStatus == "finished" {
		newStatus = "in_progress"
	}
	if _, err := tx.Exec(`UPDATE qingka_wangke_jiguang
		SET run_times=run_times+?, fees=fees+?, status=?, updated_at=?
		WHERE id=?`, delta, preview.Cost, newStatus, now, order.ID); err != nil {
		return nil, err
	}
	if preview.Cost > 0 {
		res, err := tx.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ? AND money >= ?", preview.Cost, order.UID, preview.Cost)
		if err != nil {
			return nil, err
		}
		if affected, _ := res.RowsAffected(); affected <= 0 {
			return nil, fmt.Errorf("订单所属用户余额不足")
		}
	}
	_, _ = tx.Exec("INSERT INTO qingka_wangke_moneylog (uid,type,money,mark,addtime) VALUES (?,?,?,?,?)",
		order.UID, "jiguang_add_times", -preview.Cost,
		fmt.Sprintf("极光跑步加次数，订单号:%s，加%d次，扣除%.2f", order.OrderNo, delta, preview.Cost), now)
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return map[string]any{"item": preview, "message": fmt.Sprintf("已加 %d 次，扣费 %.2f 元", delta, preview.Cost)}, nil
}

func (s *Service) OrderLogs(ctx context.Context, uid int, orderNo string, isAdmin bool) (map[string]any, error) {
	order, err := s.findOrder(uid, 0, orderNo, isAdmin)
	if err != nil {
		return nil, err
	}
	cfg, err := s.loadConfig()
	if err != nil {
		return nil, err
	}
	payload, err := s.upstreamLogs(ctx, cfg, order)
	if err != nil {
		return nil, err
	}
	rows := payloadRows(payload, "data", "list", "logs")
	list := []map[string]any{}
	for _, row := range rows {
		if strings.TrimSpace(asString(row["type"])) == "refund" {
			continue
		}
		if !isAdmin {
			delete(row, "operator")
		}
		list = append(list, row)
	}
	if order.RefundAmount != nil && strings.TrimSpace(order.RefundedAt) != "" {
		list = append(list, map[string]any{
			"id":           fmt.Sprintf("local_refund_%d", order.ID),
			"type":         "refund",
			"action":       "退款到账",
			"createdAt":    timeToRFC3339(order.RefundedAt),
			"refundAmount": round2(*order.RefundAmount),
		})
	}
	return map[string]any{"list": list}, nil
}

func (s *Service) SyncOrders(ctx context.Context) (int, error) {
	cfg, err := s.loadConfig()
	if err != nil {
		return 0, err
	}
	if !cfg.UpstreamReady() {
		return 0, fmt.Errorf("极光上游未配置完整")
	}
	pageSize := 100
	page1, total, err := s.syncPage(ctx, cfg, 1, pageSize)
	if err != nil {
		return page1, err
	}
	cursor := cfg.SyncCursorPage
	if cursor < 2 {
		cursor = 2
	}
	totalPages := 1
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(pageSize)))
	}
	if cursor > totalPages {
		cursor = 2
	}
	page2, _, err := s.syncPage(ctx, cfg, cursor, pageSize)
	if err != nil {
		return page1 + page2, err
	}
	next := cursor + 1
	if next > totalPages {
		next = 2
	}
	cfg.SyncCursorPage = next
	_ = s.saveConfig(cfg)
	return page1 + page2, nil
}

func (s *Service) syncPage(ctx context.Context, cfg Config, page, limit int) (int, int, error) {
	payload, err := s.upstreamListOrders(ctx, cfg, page, limit)
	if err != nil {
		return 0, 0, err
	}
	rows := payloadRows(payload, "data", "list", "orders")
	total := asInt(firstNonNil(
		nestedAny(payload, "data", "total"),
		nestedAny(payload, "pagination", "total"),
		payload["total"],
	))
	updated := 0
	for _, row := range rows {
		if s.syncOne(row) {
			updated++
		}
	}
	if total == 0 {
		total = len(rows)
	}
	return updated, total, nil
}

func (s *Service) syncOne(row map[string]any) bool {
	orderNo := firstString(row, "order_no", "orderNo")
	if orderNo == "" {
		return false
	}
	status := firstString(row, "status", "order_status")
	completed := asInt(firstNonNil(row["completed_times"], row["completedTimes"]))
	notes := firstString(row, "notes", "remark", "message")
	res, err := database.DB.Exec(`UPDATE qingka_wangke_jiguang
		SET status=IF(status='refunded', status, IF(?='', status, ?)),
			completed_times=IF(status='refunded', completed_times, ?),
			notes=IF(?='', notes, ?),
			updated_at=NOW()
		WHERE order_no=?`,
		status, status, completed, notes, notes, orderNo)
	if err != nil {
		return false
	}
	affected, _ := res.RowsAffected()
	return affected > 0
}

func (s *Service) productPrice(uid int, cfg Config, productID int) (float64, string, error) {
	base := productBasePrice(cfg, productID)
	if base <= 0 {
		return 0, productName(productID), fmt.Errorf("%s 单价未配置", productName(productID))
	}
	rate := s.userRate(uid)
	price := round2(base * rate)
	if price < 0 {
		price = 0
	}
	return price, productName(productID), nil
}

func (s *Service) userRate(uid int) float64 {
	rate := 1.0
	_ = database.DB.QueryRow("SELECT COALESCE(addprice,1) FROM qingka_wangke_user WHERE uid=?", uid).Scan(&rate)
	if rate <= 0 {
		rate = 1
	}
	return rate
}

func (s *Service) requireBalance(uid int, amount float64) error {
	if amount <= 0 {
		return nil
	}
	var balance float64
	if err := database.DB.QueryRow("SELECT COALESCE(money,0) FROM qingka_wangke_user WHERE uid=?", uid).Scan(&balance); err != nil {
		return fmt.Errorf("用户不存在")
	}
	if balance < amount {
		return fmt.Errorf("余额不足，需要 %.2f 元", amount)
	}
	return nil
}

func (s *Service) refundPreview(order Order, cfg Config) (RefundPreview, error) {
	unitPrice, _, err := s.productPrice(order.UID, cfg, order.ProductID)
	if err != nil {
		return RefundPreview{}, err
	}
	remaining := order.RunTimes - order.CompletedTimes
	if remaining < 0 {
		remaining = 0
	}
	amount := round2(float64(remaining) * order.KMPerDay * unitPrice)
	if amount > order.Fees {
		amount = order.Fees
	}
	return RefundPreview{
		OrderNo:        order.OrderNo,
		ProductName:    order.ProductName,
		StudentAccount: order.StudentAccount,
		Remaining:      remaining,
		KMPerDay:       order.KMPerDay,
		PricePerKM:     unitPrice,
		Amount:         amount,
	}, nil
}

func (s *Service) addTimesPreview(order Order, cfg Config, delta int) (AddTimesPreview, error) {
	unitPrice, _, err := s.productPrice(order.UID, cfg, order.ProductID)
	if err != nil {
		return AddTimesPreview{}, err
	}
	cost := round2(order.KMPerDay * float64(delta) * unitPrice)
	return AddTimesPreview{
		OrderNo:        order.OrderNo,
		ProductName:    order.ProductName,
		StudentAccount: order.StudentAccount,
		Delta:          delta,
		KMPerDay:       order.KMPerDay,
		PricePerKM:     unitPrice,
		Cost:           cost,
		BeforeRunTimes: order.RunTimes,
		AfterRunTimes:  order.RunTimes + delta,
	}, nil
}

func (s *Service) findOrder(uid, id int, orderNo string, isAdmin bool) (Order, error) {
	where := "WHERE 1=1"
	args := []any{}
	if id > 0 {
		where += " AND j.id=?"
		args = append(args, id)
	} else if strings.TrimSpace(orderNo) != "" {
		where += " AND j.order_no=?"
		args = append(args, strings.TrimSpace(orderNo))
	} else {
		return Order{}, fmt.Errorf("订单号不能为空")
	}
	if !isAdmin {
		where += " AND j.uid=?"
		args = append(args, uid)
	}
	query := `SELECT j.id, j.uid, COALESCE(u.user,''), j.order_no, COALESCE(j.upstream_id,0), j.product_id, j.product_name,
		j.school_name, j.student_name, j.student_account, j.run_times, j.completed_times, j.km_per_day,
		j.customer_message, j.status, COALESCE(j.fees,0), j.refund_amount, COALESCE(j.notes,''),
		COALESCE(j.source,''), COALESCE(j.agent_uid,0),
		COALESCE(DATE_FORMAT(j.created_at,'%Y-%m-%d %H:%i:%s'),''),
		COALESCE(DATE_FORMAT(j.updated_at,'%Y-%m-%d %H:%i:%s'),''),
		COALESCE(DATE_FORMAT(j.refunded_at,'%Y-%m-%d %H:%i:%s'),'')
		FROM qingka_wangke_jiguang j LEFT JOIN qingka_wangke_user u ON u.uid=j.uid ` + where + " LIMIT 1"
	var o Order
	var refund sql.NullFloat64
	err := database.DB.QueryRow(query, args...).Scan(&o.ID, &o.UID, &o.Username, &o.OrderNo, &o.UpstreamID, &o.ProductID, &o.ProductName,
		&o.SchoolName, &o.StudentName, &o.StudentAccount, &o.RunTimes, &o.CompletedTimes, &o.KMPerDay,
		&o.CustomerMessage, &o.Status, &o.Fees, &refund, &o.Notes, &o.Source, &o.AgentUID,
		&o.CreatedAt, &o.UpdatedAt, &o.RefundedAt)
	if err == sql.ErrNoRows {
		return Order{}, fmt.Errorf("订单不存在或无权操作")
	}
	if err != nil {
		return Order{}, err
	}
	if refund.Valid {
		value := refund.Float64
		o.RefundAmount = &value
	}
	return o, nil
}

func (s *Service) lockAction(key string) (func(), error) {
	if _, loaded := s.locks.LoadOrStore(key, struct{}{}); loaded {
		return nil, fmt.Errorf("该订单处理中，请稍后再试")
	}
	return func() { s.locks.Delete(key) }, nil
}

func normalizeOrderRequest(req OrderRequest) OrderRequest {
	req.SchoolName = strings.TrimSpace(req.SchoolName)
	req.StudentName = strings.TrimSpace(req.StudentName)
	req.StudentAccount = strings.TrimSpace(req.StudentAccount)
	req.CustomerMessage = strings.TrimSpace(req.CustomerMessage)
	return req
}

func validateOrderRequest(req OrderRequest) error {
	if req.ProductID != ProductMorning && req.ProductID != ProductDaily {
		return fmt.Errorf("请选择有效商品")
	}
	if req.SchoolName == "" || req.StudentName == "" || req.StudentAccount == "" {
		return fmt.Errorf("学校/姓名/学号必填")
	}
	if req.Times < 1 || req.Times > 9999 {
		return fmt.Errorf("次数必须为 1~9999")
	}
	if !allowedKM(req.KMPerDay) {
		return fmt.Errorf("公里数必须为 1/1.2/1.5/1.6/2.0/3.0/5/10")
	}
	if len([]rune(req.CustomerMessage)) > 500 {
		return fmt.Errorf("备注不能超过500字")
	}
	return nil
}

func allowedKM(value float64) bool {
	for _, allowed := range []float64{1, 1.2, 1.5, 1.6, 2, 3, 5, 10} {
		if math.Abs(value-allowed) < 0.0001 {
			return true
		}
	}
	return false
}

func ensureMapData(payload map[string]any) map[string]any {
	if data, ok := payload["data"].(map[string]any); ok {
		return data
	}
	return payload
}

func payloadRows(payload map[string]any, keys ...string) []map[string]any {
	for _, value := range []any{payload} {
		if rows := rowsFromAny(value); len(rows) > 0 {
			return rows
		}
	}
	for _, key := range keys {
		if value := nestedAny(payload, key); value != nil {
			if rows := rowsFromAny(value); len(rows) > 0 {
				return rows
			}
		}
		if value, ok := payload[key]; ok {
			if rows := rowsFromAny(value); len(rows) > 0 {
				return rows
			}
		}
	}
	if data, ok := payload["data"].(map[string]any); ok {
		return payloadRows(data, keys...)
	}
	return []map[string]any{}
}

func rowsFromAny(value any) []map[string]any {
	switch typed := value.(type) {
	case []any:
		out := make([]map[string]any, 0, len(typed))
		for _, item := range typed {
			if row, ok := item.(map[string]any); ok {
				out = append(out, row)
			}
		}
		return out
	case map[string]any:
		for _, key := range []string{"list", "orders", "data", "items", "logs"} {
			if rows := rowsFromAny(typed[key]); len(rows) > 0 {
				return rows
			}
		}
	}
	return nil
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

func extractOrderNo(payload map[string]any) string {
	return strings.TrimSpace(asString(firstNonNil(
		payload["order_no"], payload["orderNo"],
		nestedAny(payload, "data", "order_no"), nestedAny(payload, "data", "orderNo"),
		nestedAny(payload, "data", "order", "order_no"), nestedAny(payload, "data", "order", "orderNo"),
	)))
}

func extractOrderID(payload map[string]any) int {
	return asInt(firstNonNil(
		payload["upstream_id"], payload["order_id"], payload["id"],
		nestedAny(payload, "data", "upstream_id"), nestedAny(payload, "data", "order_id"), nestedAny(payload, "data", "id"),
		nestedAny(payload, "data", "order", "id"), nestedAny(payload, "data", "order", "order_id"),
	))
}

func extractProductName(payload map[string]any) string {
	return strings.TrimSpace(asString(firstNonNil(
		payload["product_name"], payload["productName"],
		nestedAny(payload, "data", "product_name"), nestedAny(payload, "data", "productName"),
		nestedAny(payload, "data", "order", "product_name"), nestedAny(payload, "data", "order", "productName"),
	)))
}

func extractStatus(payload map[string]any, def string) string {
	status := strings.TrimSpace(asString(firstNonNil(
		payload["status"],
		nestedAny(payload, "data", "status"),
		nestedAny(payload, "data", "order", "status"),
	)))
	if status == "" {
		return def
	}
	return status
}

func firstString(m map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(asString(m[key])); value != "" && value != "<nil>" {
			return value
		}
	}
	return ""
}

func firstNonNil(values ...any) any {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

func asString(v any) string {
	switch value := v.(type) {
	case string:
		return value
	case []byte:
		return string(value)
	case json.Number:
		return value.String()
	case fmt.Stringer:
		return value.String()
	case nil:
		return ""
	default:
		return fmt.Sprint(value)
	}
}

func asInt(v any) int {
	switch value := v.(type) {
	case int:
		return value
	case int64:
		return int(value)
	case float64:
		return int(value)
	case json.Number:
		n, _ := value.Int64()
		return int(n)
	case string:
		n, _ := strconv.Atoi(strings.TrimSpace(value))
		return n
	default:
		return 0
	}
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func timeToRFC3339(raw string) string {
	t, err := time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(raw), time.Local)
	if err != nil {
		return raw
	}
	return t.Format(time.RFC3339)
}
