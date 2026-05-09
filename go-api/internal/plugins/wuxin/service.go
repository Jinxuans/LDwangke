package wuxin

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-api/internal/database"
)

type WuxinOrder struct {
	ID                int     `json:"id"`
	UserID            int     `json:"user_id"`
	AuthCode          string  `json:"auth_code"`
	StartDate         string  `json:"start_date"`
	ResidueNum        int     `json:"residue_num"`
	Quantity          int     `json:"quantity"`
	CompletedQuantity int     `json:"completed_quantity"`
	RunMeter          float64 `json:"run_meter"`
	RunType           int     `json:"run_type"`
	ZoneName          string  `json:"zone_name"`
	ZoneCode          string  `json:"zone_code"`
	ZoneID            int     `json:"zone_id"`
	RunTime           string  `json:"run_time"`
	RunWeek           string  `json:"run_week"`
	RunSpeed          string  `json:"run_speed"`
	Status            int     `json:"status"`
	OrderStatus       int     `json:"order_status"`
	RunStatus         int     `json:"run_status"`
	Mark              string  `json:"mark"`
	Remarks           string  `json:"remarks"`
	Phone             string  `json:"phone"`
	AccountFlag       int     `json:"account_flag"`
	CreateTime        string  `json:"create_time"`
	UpdateTime        string  `json:"update_time"`
	OrderNumber       string  `json:"order_number"`
	RunPlanCode       string  `json:"run_plan_code"`
	FenceCode         string  `json:"fence_code"`
	ScheduleConfig    string  `json:"schedule_config"`
	NextExecuteDate   string  `json:"next_execute_date"`
	Fees              float64 `json:"fees"`
	Source            string  `json:"source"`
	AgentUID          int     `json:"agent_uid"`
}

type WuxinOrderRequest struct {
	AuthCode    string  `json:"auth_code"`
	StartDate   string  `json:"start_date"`
	RunPlanCode string  `json:"run_plan_code"`
	FenceCode   string  `json:"fence_code"`
	ZoneName    string  `json:"zone_name"`
	RunType     int     `json:"run_type"`
	RunTime     string  `json:"run_time"`
	RunMeter    float64 `json:"run_meter"`
	RunWeek     string  `json:"run_week"`
	RunSpeed    string  `json:"run_speed"`
	OrderNum    int     `json:"order_num"`
	Mark        string  `json:"mark"`
}

type WuxinService struct {
	client *http.Client
}

var wuxinService = &WuxinService{client: &http.Client{Timeout: 30 * time.Second}}

func Wuxin() *WuxinService {
	return wuxinService
}

func (s *WuxinService) loadConfig() (WuxinConfig, error) {
	var raw string
	err := database.DB.QueryRow("SELECT k FROM qingka_wangke_config WHERE v = ? LIMIT 1", "wuxin_config").Scan(&raw)
	if err != nil {
		return defaultWuxinConfig(), nil
	}
	cfg, err := parseWuxinConfig(raw)
	if err != nil {
		return defaultWuxinConfig(), nil
	}
	return cfg, nil
}

func (s *WuxinService) saveConfig(cfg WuxinConfig) error {
	raw, err := cfg.Marshal()
	if err != nil {
		return err
	}
	_, err = database.DB.Exec(
		"INSERT INTO qingka_wangke_config (v, k) VALUES (?, ?) ON DUPLICATE KEY UPDATE k = VALUES(k)",
		"wuxin_config", raw,
	)
	return err
}

func (s *WuxinService) getUserRate(uid int) float64 {
	rate := 1.0
	_ = database.DB.QueryRow("SELECT COALESCE(addprice, 1) FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&rate)
	if rate <= 0 {
		rate = 1
	}
	return rate
}

func (s *WuxinService) GetPrice(uid int) (map[string]any, error) {
	cfg, err := s.loadConfig()
	if err != nil {
		return nil, err
	}
	rate := s.getUserRate(uid)
	price := math.Round(cfg.Price*rate*100) / 100
	return map[string]any{
		"price":      price,
		"base_price": cfg.Price,
		"user_rate":  rate,
		"price_type": "per_order",
	}, nil
}

func (s *WuxinService) SchoolInfo(ctx context.Context, authCode string) (map[string]any, error) {
	cfg, err := s.loadConfig()
	if err != nil {
		return nil, err
	}
	return s.upstreamSchoolInfo(ctx, cfg, authCode)
}

func (s *WuxinService) CreateOrder(ctx context.Context, uid int, req WuxinOrderRequest, source string, agentUID int) (map[string]any, error) {
	if err := validateWuxinOrderRequest(req); err != nil {
		return nil, err
	}
	cfg, err := s.loadConfig()
	if err != nil {
		return nil, err
	}
	priceInfo, _ := s.GetPrice(uid)
	unitPrice := asFloat(priceInfo["price"])
	totalPrice := math.Round(unitPrice*float64(req.OrderNum)*100) / 100
	if totalPrice < 0 {
		return nil, fmt.Errorf("价格异常")
	}

	var balance float64
	_ = database.DB.QueryRow("SELECT COALESCE(money,0) FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&balance)
	if balance < totalPrice {
		return nil, fmt.Errorf("余额不足，请先充值")
	}

	upstream, err := s.upstreamCreateOrder(ctx, cfg, req)
	if err != nil {
		return nil, err
	}
	orderNumber := strings.TrimSpace(asString(nestedValue(upstream, "data", "order_number")))
	if orderNumber == "" {
		orderNumber = strings.TrimSpace(asString(upstream["order_number"]))
	}
	if orderNumber == "" {
		return nil, fmt.Errorf("获取订单号失败")
	}
	remarks := strings.TrimSpace(asString(nestedValue(upstream, "data", "remarks")))
	schedule, _ := buildScheduleConfig(req)
	scheduleRaw, _ := json.Marshal(schedule)
	now := time.Now().Format("2006-01-02 15:04:05")

	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	result, err := tx.Exec(`INSERT INTO wuxin_sdxy (
		user_id, auth_code, order_number, run_type, run_meter, run_week, run_time, run_speed,
		zone_id, zone_name, quantity, residue_num, status, run_status, account_flag,
		mark, remarks, create_time, update_time, order_status, completed_quantity,
		start_date, run_plan_code, fence_code, schedule_config, fees, source, agent_uid
	) VALUES (?,?,?,?,?,?,?,?,0,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		uid, req.AuthCode, orderNumber, req.RunType, req.RunMeter, req.RunWeek, req.RunTime, req.RunSpeed,
		req.ZoneName, req.OrderNum, req.OrderNum, 0, 1, 1,
		req.Mark, remarks, now, now, 1, 0,
		req.StartDate, req.RunPlanCode, req.FenceCode, string(scheduleRaw), totalPrice, source, agentUID,
	)
	if err != nil {
		return nil, fmt.Errorf("保存订单失败: %w", err)
	}
	orderID, _ := result.LastInsertId()
	res, err := tx.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ? AND money >= ?", totalPrice, uid, totalPrice)
	if err != nil {
		return nil, fmt.Errorf("扣除余额失败: %w", err)
	}
	if affected, _ := res.RowsAffected(); affected <= 0 {
		return nil, fmt.Errorf("余额不足，请先充值")
	}
	_, _ = tx.Exec("INSERT INTO qingka_wangke_moneylog (uid, type, money, mark, addtime) VALUES (?, ?, ?, ?, ?)",
		uid, "wuxin_add", -totalPrice, fmt.Sprintf("无心闪动下单，订单号:%s，扣除%.2f", orderNumber, totalPrice), now)
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return map[string]any{
		"id":           orderID,
		"order_number": orderNumber,
		"total_price":  totalPrice,
		"message":      "下单成功",
	}, nil
}

func (s *WuxinService) ListOrders(uid int, isAdmin bool, page, limit int, searchType, keyword string, status *int) ([]WuxinOrder, int, error) {
	where := "WHERE 1=1"
	args := []any{}
	if !isAdmin {
		where += " AND user_id = ?"
		args = append(args, uid)
	}
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		switch searchType {
		case "1":
			where += " AND id = ?"
			args = append(args, keyword)
		case "2":
			where += " AND auth_code LIKE ?"
			args = append(args, "%"+keyword+"%")
		case "3":
			if isAdmin {
				where += " AND user_id = ?"
				args = append(args, keyword)
			}
		case "4":
			where += " AND mark LIKE ?"
			args = append(args, "%"+keyword+"%")
		case "5":
			where += " AND phone LIKE ?"
			args = append(args, "%"+keyword+"%")
		default:
			where += " AND (auth_code LIKE ? OR order_number LIKE ? OR phone LIKE ? OR mark LIKE ?)"
			args = append(args, "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
		}
	}
	if status != nil && *status >= 0 {
		where += " AND order_status = ?"
		args = append(args, *status)
	}

	var total int
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM wuxin_sdxy "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	query := `SELECT id, user_id, auth_code, start_date, residue_num, quantity, completed_quantity,
		run_meter, run_type, zone_name, zone_code, zone_id, run_time, run_week, run_speed,
		status, order_status, run_status, mark, remarks, phone, account_flag, create_time,
		update_time, order_number, run_plan_code, fence_code, COALESCE(schedule_config,''),
		next_execute_date, COALESCE(fees,0), COALESCE(source,''), COALESCE(agent_uid,0)
		FROM wuxin_sdxy ` + where + " ORDER BY id DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)
	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	orders := []WuxinOrder{}
	for rows.Next() {
		var o WuxinOrder
		if err := rows.Scan(&o.ID, &o.UserID, &o.AuthCode, &o.StartDate, &o.ResidueNum, &o.Quantity,
			&o.CompletedQuantity, &o.RunMeter, &o.RunType, &o.ZoneName, &o.ZoneCode, &o.ZoneID,
			&o.RunTime, &o.RunWeek, &o.RunSpeed, &o.Status, &o.OrderStatus, &o.RunStatus,
			&o.Mark, &o.Remarks, &o.Phone, &o.AccountFlag, &o.CreateTime, &o.UpdateTime,
			&o.OrderNumber, &o.RunPlanCode, &o.FenceCode, &o.ScheduleConfig, &o.NextExecuteDate,
			&o.Fees, &o.Source, &o.AgentUID); err != nil {
			return nil, 0, err
		}
		orders = append(orders, o)
	}
	return orders, total, nil
}

func (s *WuxinService) RefundOrder(ctx context.Context, uid, id int, orderNumber string, isAdmin bool) (map[string]any, error) {
	order, err := s.findOrder(uid, id, orderNumber, isAdmin)
	if err != nil {
		return nil, err
	}
	if order.Status == 2 || order.OrderStatus == 4 {
		return nil, fmt.Errorf("订单已退款")
	}
	cfg, err := s.loadConfig()
	if err != nil {
		return nil, err
	}
	if _, err := s.upstreamRefund(ctx, cfg, order.OrderNumber); err != nil {
		return nil, err
	}
	priceInfo, _ := s.GetPrice(order.UserID)
	refundCount := order.ResidueNum
	if refundCount < 0 {
		refundCount = 0
	}
	refundAmount := math.Round(asFloat(priceInfo["price"])*float64(refundCount)*100) / 100
	now := time.Now().Format("2006-01-02 15:04:05")
	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if refundAmount > 0 {
		if _, err := tx.Exec("UPDATE qingka_wangke_user SET money = money + ? WHERE uid = ?", refundAmount, order.UserID); err != nil {
			return nil, err
		}
	}
	if _, err := tx.Exec("UPDATE wuxin_sdxy SET status=2, order_status=4, residue_num=0, update_time=? WHERE id=?", now, order.ID); err != nil {
		return nil, err
	}
	_, _ = tx.Exec("INSERT INTO qingka_wangke_moneylog (uid, type, money, mark, addtime) VALUES (?, ?, ?, ?, ?)",
		order.UserID, "wuxin_refund", refundAmount, fmt.Sprintf("无心闪动退款，订单号:%s，退款%.2f", order.OrderNumber, refundAmount), now)
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return map[string]any{"refund_amount": refundAmount, "refund_count": refundCount, "message": "申请退款成功"}, nil
}

func (s *WuxinService) IncreaseOrder(ctx context.Context, uid, id int, orderNumber string, quantity int, isAdmin bool) (map[string]any, error) {
	if quantity <= 0 {
		return nil, fmt.Errorf("追加次数必须大于0")
	}
	order, err := s.findOrder(uid, id, orderNumber, isAdmin)
	if err != nil {
		return nil, err
	}
	if order.Status != 0 {
		return nil, fmt.Errorf("只能追加未完成的订单")
	}
	cfg, err := s.loadConfig()
	if err != nil {
		return nil, err
	}
	priceInfo, _ := s.GetPrice(order.UserID)
	totalPrice := math.Round(asFloat(priceInfo["price"])*float64(quantity)*100) / 100
	var balance float64
	_ = database.DB.QueryRow("SELECT COALESCE(money,0) FROM qingka_wangke_user WHERE uid = ?", order.UserID).Scan(&balance)
	if balance < totalPrice {
		return nil, fmt.Errorf("余额不足，请先充值")
	}
	if _, err := s.upstreamIncrease(ctx, cfg, order.OrderNumber, quantity); err != nil {
		return nil, err
	}
	now := time.Now().Format("2006-01-02 15:04:05")
	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if _, err := tx.Exec("UPDATE wuxin_sdxy SET quantity=quantity+?, residue_num=residue_num+?, fees=fees+?, update_time=? WHERE id=?",
		quantity, quantity, totalPrice, now, order.ID); err != nil {
		return nil, err
	}
	res, err := tx.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ? AND money >= ?", totalPrice, order.UserID, totalPrice)
	if err != nil {
		return nil, err
	}
	if affected, _ := res.RowsAffected(); affected <= 0 {
		return nil, fmt.Errorf("余额不足，请先充值")
	}
	_, _ = tx.Exec("INSERT INTO qingka_wangke_moneylog (uid, type, money, mark, addtime) VALUES (?, ?, ?, ?, ?)",
		order.UserID, "wuxin_increase", -totalPrice, fmt.Sprintf("无心闪动追加，订单号:%s，追加%d次", order.OrderNumber, quantity), now)
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return map[string]any{"message": "追加成功", "total_price": totalPrice}, nil
}

func (s *WuxinService) ReassignOrder(ctx context.Context, uid, id int, orderNumber string, isAdmin bool) error {
	order, err := s.findOrder(uid, id, orderNumber, isAdmin)
	if err != nil {
		return err
	}
	cfg, err := s.loadConfig()
	if err != nil {
		return err
	}
	if _, err := s.upstreamReassign(ctx, cfg, order.OrderNumber); err != nil {
		return err
	}
	_, err = database.DB.Exec("UPDATE wuxin_sdxy SET order_status=1, status=0, update_time=NOW() WHERE id=?", order.ID)
	return err
}

func (s *WuxinService) OrderRecords(ctx context.Context, uid, id int, orderNumber string, page, limit int, isAdmin bool) (map[string]any, error) {
	order, err := s.findOrder(uid, id, orderNumber, isAdmin)
	if err != nil {
		return nil, err
	}
	cfg, err := s.loadConfig()
	if err != nil {
		return nil, err
	}
	return s.upstreamRecords(ctx, cfg, order.OrderNumber, page, limit)
}

func (s *WuxinService) OrderConfig(ctx context.Context, uid, id int, orderNumber string, isAdmin bool) (map[string]any, error) {
	order, err := s.findOrder(uid, id, orderNumber, isAdmin)
	if err != nil {
		return nil, err
	}
	cfg, err := s.loadConfig()
	if err != nil {
		return nil, err
	}
	schoolInfo, _ := s.upstreamSchoolInfo(ctx, cfg, order.AuthCode)
	var schedule any
	_ = json.Unmarshal([]byte(order.ScheduleConfig), &schedule)
	return map[string]any{"order": order, "school_info": schoolInfo, "schedule_config": schedule}, nil
}

func (s *WuxinService) EditOrder(ctx context.Context, uid int, id int, req WuxinOrderRequest, isAdmin bool) error {
	order, err := s.findOrder(uid, id, "", isAdmin)
	if err != nil {
		return err
	}
	if err := validateWuxinEditRequest(req); err != nil {
		return err
	}
	cfg, err := s.loadConfig()
	if err != nil {
		return err
	}
	if _, err := s.upstreamEdit(ctx, cfg, order.OrderNumber, req); err != nil {
		return err
	}
	schedule, _ := buildScheduleConfig(req)
	scheduleRaw, _ := json.Marshal(schedule)
	_, err = database.DB.Exec(`UPDATE wuxin_sdxy SET run_plan_code=?, fence_code=?, zone_name=?, run_type=?,
		run_time=?, run_meter=?, run_week=?, run_speed=?, mark=?, schedule_config=?, update_time=NOW() WHERE id=?`,
		req.RunPlanCode, req.FenceCode, req.ZoneName, req.RunType, req.RunTime, req.RunMeter, req.RunWeek,
		req.RunSpeed, req.Mark, string(scheduleRaw), order.ID)
	return err
}

func (s *WuxinService) UnsupportedTaskAction() error {
	return fmt.Errorf("当前上游协议暂不支持该任务级操作")
}

func (s *WuxinService) SyncOrders(ctx context.Context) (int, error) {
	cfg, err := s.loadConfig()
	if err != nil {
		return 0, err
	}
	pageSize := 50
	page := 1
	totalUpdated := 0
	for {
		list, total, err := s.upstreamListOrders(ctx, cfg, page, pageSize)
		if err != nil {
			return totalUpdated, err
		}
		for _, item := range list {
			if s.syncOneOrder(item) {
				totalUpdated++
			}
		}
		if total <= page*pageSize || len(list) == 0 {
			break
		}
		page++
	}
	return totalUpdated, nil
}

func (s *WuxinService) syncOneOrder(order map[string]any) bool {
	orderNumber := strings.TrimSpace(asString(order["order_number"]))
	if orderNumber == "" {
		return false
	}
	quantity := asInt(order["quantity"])
	completed := asInt(order["completed_quantity"])
	orderStatus := asInt(order["order_status"])
	residue := quantity - completed
	if orderStatus == 4 || residue < 0 {
		residue = 0
	}
	status := 0
	switch orderStatus {
	case 3:
		status = 1
	case 4:
		status = 2
	case 5:
		status = 3
	}
	res, err := database.DB.Exec(`UPDATE wuxin_sdxy SET residue_num=?, quantity=?, completed_quantity=?,
		status=?, order_status=?, start_date=?, next_execute_date=?, remarks=IF(?='', remarks, ?),
		phone=IF(?='', phone, ?), auth_code=IF(?='', auth_code, ?), account_flag=IF(?='', account_flag, 1),
		update_time=NOW() WHERE order_number=?`,
		residue, quantity, completed, status, orderStatus, asString(order["start_date"]), asString(order["next_execute_date"]),
		asString(order["remarks"]), asString(order["remarks"]), asString(order["phone"]), asString(order["phone"]),
		asString(order["auth_code"]), asString(order["auth_code"]), asString(order["auth_code"]), orderNumber)
	if err != nil {
		return false
	}
	affected, _ := res.RowsAffected()
	return affected > 0
}

func (s *WuxinService) findOrder(uid, id int, orderNumber string, isAdmin bool) (WuxinOrder, error) {
	where := "WHERE 1=1"
	args := []any{}
	if id > 0 {
		where += " AND id = ?"
		args = append(args, id)
	} else if strings.TrimSpace(orderNumber) != "" {
		where += " AND order_number = ?"
		args = append(args, strings.TrimSpace(orderNumber))
	} else {
		return WuxinOrder{}, fmt.Errorf("订单ID或订单号无效")
	}
	if !isAdmin {
		where += " AND user_id = ?"
		args = append(args, uid)
	}
	query := `SELECT id, user_id, auth_code, start_date, residue_num, quantity, completed_quantity,
		run_meter, run_type, zone_name, zone_code, zone_id, run_time, run_week, run_speed,
		status, order_status, run_status, mark, remarks, phone, account_flag, create_time,
		update_time, order_number, run_plan_code, fence_code, COALESCE(schedule_config,''),
		next_execute_date, COALESCE(fees,0), COALESCE(source,''), COALESCE(agent_uid,0)
		FROM wuxin_sdxy ` + where + " LIMIT 1"
	var o WuxinOrder
	err := database.DB.QueryRow(query, args...).Scan(&o.ID, &o.UserID, &o.AuthCode, &o.StartDate, &o.ResidueNum, &o.Quantity,
		&o.CompletedQuantity, &o.RunMeter, &o.RunType, &o.ZoneName, &o.ZoneCode, &o.ZoneID,
		&o.RunTime, &o.RunWeek, &o.RunSpeed, &o.Status, &o.OrderStatus, &o.RunStatus,
		&o.Mark, &o.Remarks, &o.Phone, &o.AccountFlag, &o.CreateTime, &o.UpdateTime,
		&o.OrderNumber, &o.RunPlanCode, &o.FenceCode, &o.ScheduleConfig, &o.NextExecuteDate,
		&o.Fees, &o.Source, &o.AgentUID)
	if err == sql.ErrNoRows {
		return WuxinOrder{}, fmt.Errorf("订单不存在或无权操作")
	}
	return o, err
}

func validateWuxinOrderRequest(req WuxinOrderRequest) error {
	if strings.TrimSpace(req.AuthCode) == "" {
		return fmt.Errorf("授权码不能为空")
	}
	if strings.TrimSpace(req.StartDate) == "" {
		return fmt.Errorf("开始日期不能为空")
	}
	if req.RunMeter <= 0 {
		return fmt.Errorf("跑步距离必须大于0")
	}
	if strings.TrimSpace(req.RunTime) == "" {
		return fmt.Errorf("跑步时间段不能为空")
	}
	if strings.TrimSpace(req.RunWeek) == "" {
		return fmt.Errorf("跑步周期不能为空")
	}
	if req.OrderNum <= 0 {
		return fmt.Errorf("下单次数必须大于0")
	}
	return validateWuxinEditRequest(req)
}

func validateWuxinEditRequest(req WuxinOrderRequest) error {
	if req.RunType <= 0 {
		req.RunType = 1
	}
	if strings.TrimSpace(req.RunPlanCode) == "" {
		return fmt.Errorf("跑步计划不能为空")
	}
	if strings.TrimSpace(req.FenceCode) == "" {
		return fmt.Errorf("跑步区域不能为空")
	}
	if strings.TrimSpace(req.RunTime) == "" {
		return fmt.Errorf("跑步时间段不能为空")
	}
	if req.RunMeter <= 0 {
		return fmt.Errorf("跑步距离必须大于0")
	}
	if strings.TrimSpace(req.RunWeek) == "" {
		return fmt.Errorf("跑步周期不能为空")
	}
	if strings.TrimSpace(req.RunSpeed) == "" {
		return fmt.Errorf("跑步配速不能为空")
	}
	return nil
}

func buildScheduleConfig(req WuxinOrderRequest) (map[string]any, error) {
	weekday := normalizeWeek(req.RunWeek)
	start, end := splitRunTime(req.RunTime)
	pace := parsePace(req.RunSpeed)
	return map[string]any{
		"run_plan_code": req.RunPlanCode,
		"fence_code":    req.FenceCode,
		"weekday":       weekday,
		"distance":      int(req.RunMeter * 1000),
		"pace":          pace,
		"start_time":    start,
		"end_time":      end,
	}, nil
}

func normalizeWeek(raw string) string {
	var arr []int
	if err := json.Unmarshal([]byte(raw), &arr); err != nil {
		return raw
	}
	names := map[int]string{1: "mon", 2: "tue", 3: "wed", 4: "thu", 5: "fri", 6: "sat", 7: "sun"}
	out := []string{}
	for _, n := range arr {
		if v := names[n]; v != "" {
			out = append(out, v)
		}
	}
	return strings.Join(out, ",")
}

func splitRunTime(raw string) (string, string) {
	parts := strings.SplitN(raw, "-", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return strings.TrimSpace(raw), ""
}

func parsePace(raw string) int {
	raw = strings.Trim(raw, "[]\" ")
	parts := strings.Split(raw, ",")
	n, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
	if n <= 0 {
		n = 6
	}
	return n * 60
}
