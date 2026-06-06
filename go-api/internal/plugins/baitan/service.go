package baitan

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

type Service struct {
	client *http.Client
}

var service = &Service{client: &http.Client{Timeout: 30 * time.Second}}

func Baitan() *Service { return service }

type OrderRequest struct {
	ID       int            `json:"id"`
	Type     string         `json:"type"`
	UserName string         `json:"userName"`
	PassWord string         `json:"passWord"`
	NikeName string         `json:"nikeName"`
	SID      string         `json:"sid"`
	EndDate  string         `json:"endDate"`
	Days     int            `json:"days"`
	Weeks    []string       `json:"weeks"`
	Report   []string       `json:"report"`
	Address  string         `json:"address"`
	Lon      string         `json:"lon"`
	Lat      string         `json:"lat"`
	Version  string         `json:"version"`
	WeekNum  int            `json:"weekNum"`
	MonthNum int            `json:"monthNum"`
	SchoolID string         `json:"schoolId"`
	Raw      map[string]any `json:"-"`
}

type Order struct {
	ID            int      `json:"id"`
	UID           int      `json:"uid"`
	Username      string   `json:"username,omitempty"`
	Type          string   `json:"type"`
	PlatformLabel string   `json:"platform_label"`
	UserName      string   `json:"userName"`
	PassWord      string   `json:"passWord"`
	NikeName      string   `json:"nikeName"`
	SID           string   `json:"sid"`
	UpstreamID    string   `json:"sxdkId"`
	EndDate       string   `json:"endDate"`
	Status        string   `json:"status"`
	Code          int      `json:"code"`
	Week          string   `json:"week"`
	Report        string   `json:"report"`
	Weeks         []string `json:"weeks"`
	Reports       []string `json:"reports"`
	Address       string   `json:"address"`
	Lon           string   `json:"lon"`
	Lat           string   `json:"lat"`
	Version       string   `json:"version"`
	WeekNum       int      `json:"weekNum"`
	MonthNum      int      `json:"monthNum"`
	PreDeduct     float64  `json:"pre_deduct"`
	ActualCost    *float64 `json:"actual_cost"`
	FinalCharge   *float64 `json:"final_charge"`
	Difference    *float64 `json:"difference"`
	PaymentStatus string   `json:"payment_status"`
	ErrorMessage  string   `json:"error_message"`
	Source        string   `json:"source"`
	AgentUID      int      `json:"agent_uid"`
	CreateTime    string   `json:"createTime"`
	UpdatedAt     string   `json:"updated_at"`
}

type BukaRequest struct {
	UserName     string `json:"userName"`
	PlatformType string `json:"platformType"`
	Type         string `json:"type"`
	StartDate    string `json:"startDate"`
	EndDate      string `json:"endDate"`
}

type BukaEstimate struct {
	Units     int     `json:"units"`
	Money     float64 `json:"money"`
	UnitLabel string  `json:"unitLabel"`
}

func (s *Service) loadConfig() (Config, error) {
	var raw string
	err := database.DB.QueryRow("SELECT k FROM qingka_wangke_config WHERE v = ? LIMIT 1", "baitan_config").Scan(&raw)
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
	_, err = database.DB.Exec("INSERT INTO qingka_wangke_config (v, k) VALUES (?, ?) ON DUPLICATE KEY UPDATE k = VALUES(k)", "baitan_config", raw)
	return err
}

func (s *Service) Platforms(uid int) ([]PlatformOption, error) {
	cfg, _ := s.loadConfig()
	out := platformOptions()
	rate := s.userRate(uid)
	for i := range out {
		base := cfg.PlatformPrices[out[i].Value]
		if base <= 0 {
			base = 1
		}
		out[i].Price = round2(base * rate)
	}
	return out, nil
}

func (s *Service) PlatformPrice(uid int, platform string) (float64, error) {
	cfg, _ := s.loadConfig()
	base := cfg.PlatformPrices[strings.TrimSpace(platform)]
	if base <= 0 {
		base = 1
	}
	return round2(base * s.userRate(uid)), nil
}

func (s *Service) SearchPhoneInfo(ctx context.Context, req OrderRequest) (map[string]any, error) {
	req = normalizeOrderRequest(req)
	if err := validatePhoneInfoRequest(req); err != nil {
		return nil, err
	}
	cfg, _ := s.loadConfig()
	return s.upstreamSearchPhoneInfo(ctx, cfg, req)
}

func (s *Service) CreateOrder(ctx context.Context, uid int, req OrderRequest, source string, agentUID int) (map[string]any, error) {
	req = normalizeOrderRequest(req)
	if err := validateOrderRequest(req, false); err != nil {
		return nil, err
	}
	price, _ := s.PlatformPrice(uid, req.Type)
	amount := round2(price * float64(req.Days))
	if err := s.requireBalance(uid, amount); err != nil {
		return nil, err
	}
	if exists, err := s.accountExists(uid, req.UserName, 0); err != nil {
		return nil, err
	} else if exists {
		return nil, fmt.Errorf("添加失败，您已存在该订单")
	}
	cfg, _ := s.loadConfig()
	upstream, err := s.upstreamCreate(ctx, cfg, req)
	if err != nil {
		return nil, err
	}
	upstreamID := extractUpstreamID(upstream)
	if upstreamID == "" {
		if lookup, err := s.upstreamQuerySourceOrder(ctx, cfg, req.UserName, req.Type, req.PassWord); err == nil {
			upstreamID = extractUpstreamID(lookup)
		}
	}
	weeks, reports := joinList(req.Weeks), joinList(req.Report)
	raw, _ := json.Marshal(upstream)
	now := time.Now().Format("2006-01-02 15:04:05")
	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	res, err := tx.Exec(`INSERT INTO qingka_baitan
		(uid,type,platform,userName,passWord,nikeName,sid,sxdkId,endDate,status,code,week,report,address,lon,lat,version,weekNum,monthNum,pre_deduct,payment_status,result_data,source,agent_uid,createTime,updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,'active',1,?,?,?,?,?,?,?,?,?,'paid',?,?,?, ?, ?)`,
		uid, req.Type, req.Type, req.UserName, req.PassWord, req.NikeName, req.SID, upstreamID, req.EndDate,
		weeks, reports, req.Address, req.Lon, req.Lat, req.Version, req.WeekNum, req.MonthNum, amount, string(raw), source, agentUID, now, now)
	if err != nil {
		return nil, fmt.Errorf("保存订单失败: %w", err)
	}
	id, _ := res.LastInsertId()
	if amount > 0 {
		res, err = tx.Exec("UPDATE qingka_wangke_user SET money=money-? WHERE uid=? AND money>=?", amount, uid, amount)
		if err != nil {
			return nil, err
		}
		if affected, _ := res.RowsAffected(); affected <= 0 {
			return nil, fmt.Errorf("余额不足")
		}
		_, _ = tx.Exec("INSERT INTO qingka_wangke_moneylog (uid,type,money,mark,addtime) VALUES (?,?,?,?,?)", uid, "baitan_add", -amount, fmt.Sprintf("摆摊打卡下单，账号:%s，扣除%.2f", req.UserName, amount), now)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return map[string]any{"id": id, "sxdkId": upstreamID, "total_price": amount, "message": fmt.Sprintf("订单添加成功，扣除%.2f元", amount)}, nil
}

func (s *Service) ListOrders(uid int, isAdmin bool, page, limit int, search, keyword, status string, filterUID int) ([]Order, int, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 200 {
		limit = 20
	}
	where := "WHERE 1=1"
	args := []any{}
	if !isAdmin {
		where += " AND b.uid=?"
		args = append(args, uid)
	} else if filterUID > 0 {
		where += " AND b.uid=?"
		args = append(args, filterUID)
	}
	if strings.TrimSpace(status) != "" {
		where += " AND b.status=?"
		args = append(args, strings.TrimSpace(status))
	}
	keyword = strings.TrimSpace(keyword)
	if keyword != "" {
		switch strings.TrimSpace(search) {
		case "id":
			where += " AND b.id=?"
			args = append(args, keyword)
		case "nikeName":
			where += " AND b.nikeName LIKE ?"
			args = append(args, "%"+keyword+"%")
		case "type":
			where += " AND b.type=?"
			args = append(args, keyword)
		case "uid":
			if isAdmin {
				where += " AND b.uid=?"
				args = append(args, keyword)
			}
		default:
			where += " AND (b.userName LIKE ? OR b.nikeName LIKE ? OR b.type LIKE ? OR b.sid LIKE ?)"
			args = append(args, "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
		}
	}
	var total int
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM qingka_baitan b "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	query := `SELECT b.id,b.uid,COALESCE(u.user,''),COALESCE(NULLIF(b.type,''), b.platform),b.userName,b.passWord,COALESCE(b.nikeName,''),COALESCE(b.sid,''),COALESCE(b.sxdkId,''),
		COALESCE(DATE_FORMAT(b.endDate,'%Y-%m-%d'),''),COALESCE(b.status,''),COALESCE(b.code,1),COALESCE(b.week,''),COALESCE(b.report,''),COALESCE(b.address,''),COALESCE(b.lon,''),COALESCE(b.lat,''),COALESCE(b.version,''),
		COALESCE(b.weekNum,6),COALESCE(b.monthNum,25),COALESCE(b.pre_deduct,0),b.actual_cost,b.final_charge,b.difference,COALESCE(b.payment_status,''),COALESCE(b.error_message,''),COALESCE(b.source,''),COALESCE(b.agent_uid,0),
		COALESCE(DATE_FORMAT(b.createTime,'%Y-%m-%d %H:%i:%s'),''),COALESCE(DATE_FORMAT(b.updated_at,'%Y-%m-%d %H:%i:%s'),'')
		FROM qingka_baitan b LEFT JOIN qingka_wangke_user u ON u.uid=b.uid ` + where + ` ORDER BY b.id DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)
	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := []Order{}
	for rows.Next() {
		var o Order
		var actual, final, diff sql.NullFloat64
		if err := rows.Scan(&o.ID, &o.UID, &o.Username, &o.Type, &o.UserName, &o.PassWord, &o.NikeName, &o.SID, &o.UpstreamID, &o.EndDate, &o.Status, &o.Code, &o.Week, &o.Report, &o.Address, &o.Lon, &o.Lat, &o.Version, &o.WeekNum, &o.MonthNum, &o.PreDeduct, &actual, &final, &diff, &o.PaymentStatus, &o.ErrorMessage, &o.Source, &o.AgentUID, &o.CreateTime, &o.UpdatedAt); err != nil {
			return nil, 0, err
		}
		o.PlatformLabel = platformLabel(o.Type)
		o.Weeks = splitList(o.Week)
		o.Reports = splitList(o.Report)
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
		list = append(list, o)
	}
	return list, total, nil
}

func (s *Service) AddDays(ctx context.Context, uid, id, days int, isAdmin bool) (map[string]any, error) {
	if days <= 0 || days > 3650 {
		return nil, fmt.Errorf("增加天数必须为 1~3650")
	}
	order, err := s.findOrder(uid, id, isAdmin)
	if err != nil {
		return nil, err
	}
	price, _ := s.PlatformPrice(order.UID, order.Type)
	amount := round2(price * float64(days))
	if err := s.requireBalance(order.UID, amount); err != nil {
		return nil, err
	}
	cfg, _ := s.loadConfig()
	if _, err := s.upstreamAddDays(ctx, cfg, order, days); err != nil {
		return nil, err
	}
	now := time.Now().Format("2006-01-02 15:04:05")
	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	_, err = tx.Exec("UPDATE qingka_baitan SET endDate=DATE_ADD(GREATEST(COALESCE(endDate,CURDATE()), CURDATE()), INTERVAL ? DAY), pre_deduct=pre_deduct+?, updated_at=NOW() WHERE id=?", days, amount, order.ID)
	if err != nil {
		return nil, err
	}
	if amount > 0 {
		res, err := tx.Exec("UPDATE qingka_wangke_user SET money=money-? WHERE uid=? AND money>=?", amount, order.UID, amount)
		if err != nil {
			return nil, err
		}
		if affected, _ := res.RowsAffected(); affected <= 0 {
			return nil, fmt.Errorf("余额不足")
		}
		_, _ = tx.Exec("INSERT INTO qingka_wangke_moneylog (uid,type,money,mark,addtime) VALUES (?,?,?,?,?)", order.UID, "baitan_add_days", -amount, fmt.Sprintf("摆摊打卡增加天数，账号:%s，增加%d天，扣除%.2f", order.UserName, days, amount), now)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return map[string]any{"message": fmt.Sprintf("增加天数成功，扣除%.2f元", amount), "amount": amount}, nil
}

func (s *Service) EditOrder(ctx context.Context, uid int, req OrderRequest, isAdmin bool) (map[string]any, error) {
	req = normalizeOrderRequest(req)
	if err := validateOrderRequest(req, true); err != nil {
		return nil, err
	}
	order, err := s.findOrder(uid, req.ID, isAdmin)
	if err != nil {
		return nil, err
	}
	cfg, _ := s.loadConfig()
	if _, err := s.upstreamEdit(ctx, cfg, req); err != nil {
		return nil, err
	}
	weeks, reports := joinList(req.Weeks), joinList(req.Report)
	_, err = database.DB.Exec(`UPDATE qingka_baitan SET passWord=?, nikeName=?, sid=?, week=?, report=?, address=?, lon=?, lat=?, version=?, weekNum=?, monthNum=?, updated_at=NOW() WHERE id=?`, req.PassWord, req.NikeName, req.SID, weeks, reports, req.Address, req.Lon, req.Lat, req.Version, req.WeekNum, req.MonthNum, order.ID)
	if err != nil {
		return nil, err
	}
	return map[string]any{"message": "订单修改成功"}, nil
}

func (s *Service) DeleteOrder(ctx context.Context, uid, id int, isAdmin bool) (map[string]any, error) {
	order, err := s.findOrder(uid, id, isAdmin)
	if err != nil {
		return nil, err
	}
	cfg, _ := s.loadConfig()
	if _, err := s.upstreamDelete(ctx, cfg, order); err != nil {
		return nil, err
	}
	refund := s.refundByEndDate(order)
	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if _, err := tx.Exec("DELETE FROM qingka_baitan WHERE id=?", order.ID); err != nil {
		return nil, err
	}
	if refund > 0 {
		_, _ = tx.Exec("UPDATE qingka_wangke_user SET money=money+? WHERE uid=?", refund, order.UID)
		_, _ = tx.Exec("INSERT INTO qingka_wangke_moneylog (uid,type,money,mark,addtime) VALUES (?,?,?,?,NOW())", order.UID, "baitan_refund", refund, fmt.Sprintf("摆摊打卡删除退款，账号:%s，退款%.2f", order.UserName, refund))
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return map[string]any{"message": fmt.Sprintf("删除成功，退款%.2f", refund), "refund": refund}, nil
}

func (s *Service) QuerySourceOrder(ctx context.Context, uid, id int, isAdmin bool) (map[string]any, error) {
	order, err := s.findOrder(uid, id, isAdmin)
	if err != nil {
		return nil, err
	}
	cfg, _ := s.loadConfig()
	payload, err := s.upstreamQuerySourceOrder(ctx, cfg, order.UserName, order.Type, order.PassWord)
	if err != nil {
		return nil, err
	}
	upstreamID := extractUpstreamID(payload)
	week, report := extractTaskFields(payload)
	if upstreamID != "" || week != "" || report != "" {
		_, _ = database.DB.Exec("UPDATE qingka_baitan SET sxdkId=IF(?='', sxdkId, ?), week=IF(?='', week, ?), report=IF(?='', report, ?), updated_at=NOW() WHERE id=?", upstreamID, upstreamID, week, week, report, report, order.ID)
	}
	return payload, nil
}

func (s *Service) Logs(ctx context.Context, uid, id int, isAdmin bool) (map[string]any, error) {
	order, err := s.findOrder(uid, id, isAdmin)
	if err != nil {
		return nil, err
	}
	cfg, _ := s.loadConfig()
	return s.upstreamLogs(ctx, cfg, order)
}

func (s *Service) Notice(ctx context.Context) (map[string]any, error) {
	cfg, _ := s.loadConfig()
	return s.upstreamNotice(ctx, cfg)
}

func (s *Service) Schools(ctx context.Context, platform, dictKey string) (map[string]any, error) {
	cfg, _ := s.loadConfig()
	if strings.TrimSpace(dictKey) == "" {
		dictKey = platformDictKey(platform)
	}
	if strings.TrimSpace(dictKey) == "" {
		return map[string]any{"list": []any{}}, nil
	}
	return s.upstreamSchools(ctx, cfg, dictKey)
}

func (s *Service) BukaEstimate(req BukaRequest) (BukaEstimate, error) {
	cfg, _ := s.loadConfig()
	units := bukaCountUnits(req.Type, req.StartDate, req.EndDate)
	if units < 0 {
		units = 0
	}
	label := "天"
	if req.Type == "4" {
		label = "周"
	} else if req.Type == "5" {
		label = "月"
	}
	return BukaEstimate{Units: units, Money: round2(float64(units) * cfg.BukaUnitPrice), UnitLabel: label}, nil
}

func (s *Service) SubmitBuka(ctx context.Context, uid int, req BukaRequest, isAdmin bool) (map[string]any, error) {
	req.UserName = strings.TrimSpace(req.UserName)
	req.PlatformType = strings.TrimSpace(req.PlatformType)
	if req.UserName == "" || req.PlatformType == "" || req.Type == "" || req.StartDate == "" || req.EndDate == "" {
		return nil, fmt.Errorf("请填写完整补签信息")
	}
	order, err := s.findOrderByAccount(uid, req.UserName, req.PlatformType, isAdmin)
	if err != nil {
		return nil, err
	}
	est, err := s.BukaEstimate(req)
	if err != nil {
		return nil, err
	}
	if est.Units < 1 {
		return nil, fmt.Errorf("日期范围无效")
	}
	if err := s.requireBalance(order.UID, est.Money); err != nil {
		return nil, err
	}
	cfg, _ := s.loadConfig()
	if _, err := s.upstreamBuka(ctx, cfg, req); err != nil {
		return nil, err
	}
	now := time.Now().Format("2006-01-02 15:04:05")
	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if est.Money > 0 {
		res, err := tx.Exec("UPDATE qingka_wangke_user SET money=money-? WHERE uid=? AND money>=?", est.Money, order.UID, est.Money)
		if err != nil {
			return nil, err
		}
		if affected, _ := res.RowsAffected(); affected <= 0 {
			return nil, fmt.Errorf("余额不足")
		}
		_, _ = tx.Exec("INSERT INTO qingka_wangke_moneylog (uid,type,money,mark,addtime) VALUES (?,?,?,?,?)", order.UID, "baitan_buka", -est.Money, fmt.Sprintf("摆摊打卡补签，账号:%s，%s~%s，扣除%.2f", req.UserName, req.StartDate, req.EndDate, est.Money), now)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return map[string]any{"message": fmt.Sprintf("提交成功，已扣除 %.2f 元", est.Money), "estimate": est}, nil
}

func (s *Service) SyncOrders(ctx context.Context, limit int) (int, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := database.DB.Query(`SELECT id FROM qingka_baitan WHERE status NOT IN ('deleted','refunded') ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	updated := 0
	for rows.Next() {
		var id int
		if rows.Scan(&id) == nil {
			if ok, _ := s.SyncOne(ctx, 0, id, true); ok {
				updated++
			}
		}
	}
	return updated, nil
}

func (s *Service) SyncOne(ctx context.Context, uid, id int, isAdmin bool) (bool, error) {
	order, err := s.findOrder(uid, id, isAdmin)
	if err != nil {
		return false, err
	}
	payload, err := s.QuerySourceOrder(ctx, order.UID, order.ID, true)
	if err != nil {
		return false, err
	}
	actual := round2(asFloat(firstNonNil(payload["actual_cost"], payload["actual"], nestedAny(payload, "data", "actual_cost"), nestedAny(payload, "data", "actual"))))
	status := firstString(mapFromAny(firstNonNil(payload["data"], payload)), "status", "order_status")
	if status == "" {
		status = order.Status
	}
	if actual <= 0 {
		_, _ = database.DB.Exec("UPDATE qingka_baitan SET status=?, updated_at=NOW() WHERE id=?", status, order.ID)
		return true, nil
	}
	finalCharge := actual
	diff := round2(finalCharge - order.PreDeduct)
	tx, err := database.DB.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	_, err = tx.Exec("UPDATE qingka_baitan SET status=?, actual_cost=?, final_charge=?, difference=?, payment_status='settled', updated_at=NOW() WHERE id=?", status, actual, finalCharge, diff, order.ID)
	if err != nil {
		return false, err
	}
	if diff < 0 {
		refund := math.Abs(diff)
		_, _ = tx.Exec("UPDATE qingka_wangke_user SET money=money+? WHERE uid=?", refund, order.UID)
		_, _ = tx.Exec("INSERT INTO qingka_wangke_moneylog (uid,type,money,mark,addtime) VALUES (?,?,?,?,NOW())", order.UID, "baitan_diff_refund", refund, fmt.Sprintf("摆摊打卡差价退款，账号:%s，退款%.2f", order.UserName, refund))
	} else if diff > 0 {
		res, err := tx.Exec("UPDATE qingka_wangke_user SET money=money-? WHERE uid=? AND money>=?", diff, order.UID, diff)
		if err != nil {
			return false, err
		}
		if affected, _ := res.RowsAffected(); affected <= 0 {
			_, _ = tx.Exec("UPDATE qingka_baitan SET payment_status='insufficient' WHERE id=?", order.ID)
		} else {
			_, _ = tx.Exec("INSERT INTO qingka_wangke_moneylog (uid,type,money,mark,addtime) VALUES (?,?,?,?,NOW())", order.UID, "baitan_diff_charge", -diff, fmt.Sprintf("摆摊打卡差价补扣，账号:%s，补扣%.2f", order.UserName, diff))
		}
	}
	return true, tx.Commit()
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

func (s *Service) accountExists(uid int, account string, excludeID int) (bool, error) {
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM qingka_baitan WHERE uid=? AND userName=? AND id<>?", uid, account, excludeID).Scan(&count)
	return count > 0, err
}

func (s *Service) findOrder(uid, id int, isAdmin bool) (Order, error) {
	if id <= 0 {
		return Order{}, fmt.Errorf("订单不存在或无权操作")
	}
	where := "WHERE b.id=?"
	args := []any{id}
	if !isAdmin {
		where += " AND b.uid=?"
		args = append(args, uid)
	}
	list, _, err := s.queryOrders(where, args, 1, 0)
	if err != nil {
		return Order{}, err
	}
	if len(list) == 0 {
		return Order{}, fmt.Errorf("订单不存在或无权操作")
	}
	return list[0], nil
}

func (s *Service) findOrderByAccount(uid int, account, platform string, isAdmin bool) (Order, error) {
	where := "WHERE b.userName=? AND COALESCE(NULLIF(b.type,''), b.platform)=?"
	args := []any{account, platform}
	if !isAdmin {
		where += " AND b.uid=?"
		args = append(args, uid)
	}
	list, _, err := s.queryOrders(where, args, 1, 0)
	if err != nil {
		return Order{}, err
	}
	if len(list) == 0 {
		return Order{}, fmt.Errorf("您无此订单")
	}
	return list[0], nil
}

func (s *Service) queryOrders(where string, args []any, limit, offset int) ([]Order, int, error) {
	query := `SELECT b.id,b.uid,COALESCE(u.user,''),COALESCE(NULLIF(b.type,''), b.platform),b.userName,b.passWord,COALESCE(b.nikeName,''),COALESCE(b.sid,''),COALESCE(b.sxdkId,''),
		COALESCE(DATE_FORMAT(b.endDate,'%Y-%m-%d'),''),COALESCE(b.status,''),COALESCE(b.code,1),COALESCE(b.week,''),COALESCE(b.report,''),COALESCE(b.address,''),COALESCE(b.lon,''),COALESCE(b.lat,''),COALESCE(b.version,''),
		COALESCE(b.weekNum,6),COALESCE(b.monthNum,25),COALESCE(b.pre_deduct,0),b.actual_cost,b.final_charge,b.difference,COALESCE(b.payment_status,''),COALESCE(b.error_message,''),COALESCE(b.source,''),COALESCE(b.agent_uid,0),
		COALESCE(DATE_FORMAT(b.createTime,'%Y-%m-%d %H:%i:%s'),''),COALESCE(DATE_FORMAT(b.updated_at,'%Y-%m-%d %H:%i:%s'),'')
		FROM qingka_baitan b LEFT JOIN qingka_wangke_user u ON u.uid=b.uid ` + where + ` ORDER BY b.id DESC`
	if limit > 0 {
		query += " LIMIT ? OFFSET ?"
		args = append(args, limit, offset)
	}
	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := []Order{}
	for rows.Next() {
		var o Order
		var actual, final, diff sql.NullFloat64
		if err := rows.Scan(&o.ID, &o.UID, &o.Username, &o.Type, &o.UserName, &o.PassWord, &o.NikeName, &o.SID, &o.UpstreamID, &o.EndDate, &o.Status, &o.Code, &o.Week, &o.Report, &o.Address, &o.Lon, &o.Lat, &o.Version, &o.WeekNum, &o.MonthNum, &o.PreDeduct, &actual, &final, &diff, &o.PaymentStatus, &o.ErrorMessage, &o.Source, &o.AgentUID, &o.CreateTime, &o.UpdatedAt); err != nil {
			return nil, 0, err
		}
		o.PlatformLabel = platformLabel(o.Type)
		o.Weeks = splitList(o.Week)
		o.Reports = splitList(o.Report)
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
		list = append(list, o)
	}
	return list, len(list), nil
}

func (s *Service) refundByEndDate(order Order) float64 {
	if order.EndDate == "" {
		return 0
	}
	end, err := time.ParseInLocation("2006-01-02", order.EndDate, time.Local)
	if err != nil {
		return 0
	}
	today := time.Now()
	days := int(end.Sub(time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.Local)).Hours() / 24)
	if days < 1 {
		return 0
	}
	price, _ := s.PlatformPrice(order.UID, order.Type)
	return round2(float64(days) * price)
}

func normalizeOrderRequest(req OrderRequest) OrderRequest {
	req.Type = strings.TrimSpace(req.Type)
	req.UserName = strings.TrimSpace(req.UserName)
	req.PassWord = strings.TrimSpace(req.PassWord)
	req.NikeName = strings.TrimSpace(req.NikeName)
	req.SID = strings.TrimSpace(firstNonEmpty(req.SID, req.SchoolID))
	req.EndDate = strings.TrimSpace(req.EndDate)
	req.Address = strings.TrimSpace(req.Address)
	req.Lon = strings.TrimSpace(req.Lon)
	req.Lat = strings.TrimSpace(req.Lat)
	req.Version = strings.TrimSpace(req.Version)
	if req.WeekNum <= 0 {
		req.WeekNum = 6
	}
	if req.MonthNum <= 0 {
		req.MonthNum = 25
	}
	return req
}

func validateOrderRequest(req OrderRequest, edit bool) error {
	if edit && req.ID <= 0 {
		return fmt.Errorf("订单ID不能为空")
	}
	if req.Type == "" {
		return fmt.Errorf("请选择打卡平台")
	}
	if req.UserName == "" || req.PassWord == "" {
		return fmt.Errorf("账号和密码必填")
	}
	if !edit {
		if req.EndDate == "" {
			return fmt.Errorf("到期时间必填")
		}
		if req.Days <= 0 {
			return fmt.Errorf("下单天数必须大于0")
		}
	}
	return nil
}

func validatePhoneInfoRequest(req OrderRequest) error {
	if req.UserName == "" || req.PassWord == "" {
		return fmt.Errorf("账号和密码必填")
	}
	if req.EndDate == "" {
		return fmt.Errorf("请先配置到期时间")
	}
	if _, err := time.ParseInLocation("2006-01-02", req.EndDate, time.Local); err != nil {
		return fmt.Errorf("到期时间格式错误")
	}
	return nil
}

func joinList(items []string) string {
	out := []string{}
	for _, item := range items {
		if text := strings.TrimSpace(item); text != "" {
			out = append(out, text)
		}
	}
	return strings.Join(out, ",")
}

func splitList(raw string) []string {
	parts := strings.Split(strings.TrimSpace(raw), ",")
	out := []string{}
	for _, part := range parts {
		if text := strings.TrimSpace(part); text != "" {
			out = append(out, text)
		}
	}
	return out
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func bukaCountUnits(kind, startDate, endDate string) int {
	start, err1 := time.ParseInLocation("2006-01-02", strings.TrimSpace(startDate), time.Local)
	end, err2 := time.ParseInLocation("2006-01-02", strings.TrimSpace(endDate), time.Local)
	if err1 != nil || err2 != nil || end.Before(start) {
		return 0
	}
	if kind == "1" || kind == "3" {
		return int(end.Sub(start).Hours()/24) + 1
	}
	if kind == "4" {
		count := 0
		for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
			if d.Weekday() == time.Monday {
				count++
			}
		}
		return count
	}
	if kind == "5" {
		months := 0
		cursor := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.Local)
		last := time.Date(end.Year(), end.Month(), 1, 0, 0, 0, 0, time.Local)
		for !cursor.After(last) {
			months++
			cursor = cursor.AddDate(0, 1, 0)
		}
		return months
	}
	return int(end.Sub(start).Hours()/24) + 1
}

func extractUpstreamID(payload map[string]any) string {
	for _, v := range []any{payload["id"], payload["sxdkId"], payload["order_id"], payload["upstream_id"], nestedAny(payload, "data", "id"), nestedAny(payload, "data", "sxdkId"), nestedAny(payload, "data", "order_id")} {
		if text := strings.TrimSpace(asString(v)); text != "" && text != "<nil>" {
			return text
		}
	}
	if rows := payloadRows(payload, "data", "list"); len(rows) > 0 {
		return extractUpstreamID(rows[0])
	}
	return ""
}

func extractTaskFields(payload map[string]any) (string, string) {
	row := mapFromAny(firstNonNil(payload["data"], payload))
	if rows := payloadRows(payload, "data", "list"); len(rows) > 0 {
		row = rows[0]
	}
	week := joinAnyList(firstNonNil(row["weeks"], row["week"]))
	report := joinAnyList(firstNonNil(row["report"], row["reports"]))
	return week, report
}

func joinAnyList(v any) string {
	switch t := v.(type) {
	case []any:
		items := []string{}
		for _, item := range t {
			if text := strings.TrimSpace(asString(item)); text != "" {
				items = append(items, text)
			}
		}
		return strings.Join(items, ",")
	case []string:
		return joinList(t)
	default:
		return strings.TrimSpace(asString(v))
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
func mapFromAny(v any) map[string]any {
	if row, ok := v.(map[string]any); ok {
		return row
	}
	return map[string]any{}
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

func payloadRows(payload map[string]any, keys ...string) []map[string]any {
	for _, key := range keys {
		if rows := rowsFromAny(payload[key]); len(rows) > 0 {
			return rows
		}
	}
	return rowsFromAny(payload)
}

func rowsFromAny(value any) []map[string]any {
	switch typed := value.(type) {
	case []any:
		out := []map[string]any{}
		for _, item := range typed {
			if row, ok := item.(map[string]any); ok {
				out = append(out, row)
			}
		}
		return out
	case map[string]any:
		for _, key := range []string{"list", "data", "rows", "orders"} {
			if rows := rowsFromAny(typed[key]); len(rows) > 0 {
				return rows
			}
		}
	}
	return nil
}

func asString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
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
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case json.Number:
		f, _ := t.Float64()
		return f
	case string:
		f, _ := strconv.ParseFloat(strings.TrimSpace(t), 64)
		return f
	default:
		return 0
	}
}
