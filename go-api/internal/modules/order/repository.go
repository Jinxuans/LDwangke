package order

import (
	"errors"
	"fmt"
	"go-api/internal/config"
	"go-api/internal/database"
	"go-api/internal/model"
	classmodule "go-api/internal/modules/class"
	suppliermodule "go-api/internal/modules/supplier"
	shared "go-api/internal/shared/db"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func autoSyncVerboseLogging() bool {
	if config.Global == nil {
		return true
	}
	return strings.ToLower(strings.TrimSpace(config.Global.Server.Mode)) != "release"
}

func autoSyncProgressLogStep() int64 {
	return 100
}

func parseOrderTime(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	// 兼容旧系统带时区噪音的格式，如 "2024-12-17 13:40:53--506400"，截取前19字符
	if len(value) > 19 && value[10] == ' ' {
		value = value[:19]
	}
	if t, err := time.ParseInLocation("2006-01-02 15:04:05", value, time.Local); err == nil {
		return t, true
	}
	// 兼容 ISO 8601 格式，如 "2025-10-08T18:39:57Z"
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t.In(time.Local), true
	}
	// 兼容旧系统 Unix 时间戳格式
	if ts, err := strconv.ParseInt(value, 10, 64); err == nil && ts > 0 {
		return time.Unix(ts, 0), true
	}
	return time.Time{}, false
}

func matchAutoSyncRule(ageHours float64, rules []AutoSyncRule) (AutoSyncRule, bool) {
	for _, rule := range rules {
		if !rule.Enabled || rule.IntervalMinutes <= 0 {
			continue
		}
		if ageHours < float64(rule.MinAgeHours) {
			continue
		}
		if rule.MaxAgeHours > 0 && ageHours >= float64(rule.MaxAgeHours) {
			continue
		}
		return rule, true
	}
	return AutoSyncRule{}, false
}

// Repository 为订单模块提供最小的存取边界，当前先复用旧 service 实现。
type Repository interface {
	List(uid int, grade string, req model.OrderListRequest) ([]model.Order, int64, error)
	Detail(uid int, grade string, oid int) (*model.Order, error)
	Stats(uid int, grade string) (*model.OrderStats, error)
	AddOrders(uid int, req model.OrderAddRequest) (*model.OrderAddResult, error)
	AddOrdersForMall(bUID, tid, cUID int, retailPrice float64, outTradeNo string, req model.OrderAddRequest) (*model.OrderAddResult, error)
	ChangeStatus(uid int, grade string, req model.OrderStatusRequest) error
	CancelOrder(uid int, grade string, oid int) error
	RefundOrders(uid int, grade string, oids []int) error
	ModifyRemarks(oids []int, remarks string) error
	ManualDockOrders(oids []int) (int, int, error)
	SyncOrderProgress(oids []int) (int, error)
	AutoSyncAllProgress(opts AutoSyncOptions) (int, int, error)
	BatchSyncOrders(oids []int) (int, error)
	BatchResendOrders(oids []int) (int, int, error)
}

type supplierGateway interface {
	GetSupplierByHID(hid int) (*model.SupplierFull, error)
	GetClassFull(cid int) (*model.ClassFull, error)
	CallSupplierOrder(sup *model.SupplierFull, cls *model.ClassFull, school, user, pass, kcid, kcname string, extraFields map[string]string) (*model.SupplierOrderResult, error)
	HasBatchProgressForPT(pt string) bool
	HasBatchProgressConfig(sup *model.SupplierFull) bool
	QueryBatchOrderProgress(sup *model.SupplierFull, refs []model.SupplierBatchProgressRef) ([]model.SupplierProgressItem, error)
	QueryOrderProgress(sup *model.SupplierFull, yid string, username string, orderExtra map[string]string) ([]model.SupplierProgressItem, error)
	ResubmitOrder(sup *model.SupplierFull, yid string) (int, string, error)
}

type legacyRepository struct {
	orders *shared.OrderRepo
	sup    supplierGateway
}

type autoSyncItem struct {
	OID        int
	YID        string
	HID        int
	PT         string
	User       string
	KCName     string
	Noun       string
	KCID       string
	Status     string
	AddTime    string
	UpdateTime string

	MatchedRule AutoSyncRule
}

func NewRepository() Repository {
	return &legacyRepository{
		orders: shared.NewOrderRepo(),
		sup:    suppliermodule.SharedService(),
	}
}

func autoSyncUserCourseKey(user string, kcname string) string {
	user = strings.TrimSpace(user)
	kcname = strings.TrimSpace(kcname)
	if user == "" || kcname == "" {
		return ""
	}
	return user + "\x00" + kcname
}

func autoSyncNounUserCourseKey(noun string, user string, kcname string) string {
	noun = strings.TrimSpace(noun)
	user = strings.TrimSpace(user)
	kcname = strings.TrimSpace(kcname)
	if noun == "" || user == "" || kcname == "" {
		return ""
	}
	return noun + "\x00" + user + "\x00" + kcname
}

func buildBatchProgressRefs(items []autoSyncItem) []model.SupplierBatchProgressRef {
	refs := make([]model.SupplierBatchProgressRef, 0, len(items))
	for _, item := range items {
		refs = append(refs, model.SupplierBatchProgressRef{
			YID:    item.YID,
			User:   item.User,
			KCName: item.KCName,
			KCID:   item.KCID,
			Noun:   item.Noun,
		})
	}
	return refs
}

func indexAutoSyncItems(items []autoSyncItem) (map[string][]autoSyncItem, map[string][]autoSyncItem) {
	byYID := make(map[string][]autoSyncItem, len(items))
	byNounUserCourse := make(map[string][]autoSyncItem, len(items))
	for _, item := range items {
		if yid := strings.TrimSpace(item.YID); yid != "" {
			byYID[yid] = append(byYID[yid], item)
		}
		if key := autoSyncNounUserCourseKey(item.Noun, item.User, item.KCName); key != "" {
			byNounUserCourse[key] = append(byNounUserCourse[key], item)
		}
	}
	return byYID, byNounUserCourse
}

func matchAutoSyncItems(progress model.SupplierProgressItem, byYID map[string][]autoSyncItem, byNounUserCourse map[string][]autoSyncItem) []autoSyncItem {
	if yid := strings.TrimSpace(progress.YID); yid != "" {
		if items := byYID[yid]; len(items) > 0 {
			return items
		}
	}
	if key := autoSyncNounUserCourseKey(progress.Noun, progress.User, progress.KCName); key != "" {
		if items := byNounUserCourse[key]; len(items) > 0 {
			return items
		}
	}
	return nil
}

func mergeAutoSyncProgress(item autoSyncItem, progress model.SupplierProgressItem) model.SupplierProgressItem {
	if strings.TrimSpace(progress.YID) == "" {
		progress.YID = item.YID
	}
	if strings.TrimSpace(progress.KCName) == "" {
		progress.KCName = item.KCName
	}
	if strings.TrimSpace(progress.User) == "" {
		progress.User = item.User
	}
	if strings.TrimSpace(progress.Status) == "" {
		progress.Status = item.Status
	}
	if strings.TrimSpace(progress.StatusText) == "" {
		progress.StatusText = progress.Status
	}
	return progress
}

func touchAutoSyncOrder(oid int, syncTime string) {
	database.DB.Exec("UPDATE qingka_wangke_order SET updatetime = ? WHERE oid = ?", syncTime, oid)
}

func applyAutoSyncProgressUpdate(item autoSyncItem, progress model.SupplierProgressItem, syncTime string) error {
	progress = mergeAutoSyncProgress(item, progress)
	statusText := progress.Status
	if progress.StatusText != "" {
		statusText = progress.StatusText
	}
	if _, err := database.DB.Exec(
		"UPDATE qingka_wangke_order SET name = ?, yid = ?, status = ?, process = ?, remarks = ?, courseStartTime = ?, courseEndTime = ?, examStartTime = ?, examEndTime = ?, updatetime = ? WHERE oid = ?",
		progress.KCName, progress.YID, statusText, progress.Process, progress.Remarks,
		progress.CourseStartTime, progress.CourseEndTime, progress.ExamStartTime, progress.ExamEndTime, syncTime,
		item.OID,
	); err != nil {
		return err
	}
	orderStatusNotifier(item.OID, statusText, progress.Process, progress.Remarks)
	return nil
}

func (r *legacyRepository) List(uid int, grade string, req model.OrderListRequest) ([]model.Order, int64, error) {
	return r.orders.List(uid, grade, req)
}

func (r *legacyRepository) Detail(uid int, grade string, oid int) (*model.Order, error) {
	return r.orders.Detail(uid, grade, oid)
}

func (r *legacyRepository) Stats(uid int, grade string) (*model.OrderStats, error) {
	return r.orders.Stats(uid, grade)
}

func (r *legacyRepository) AddOrders(uid int, req model.OrderAddRequest) (*model.OrderAddResult, error) {
	cls, err := r.sup.GetClassFull(req.CID)
	if err != nil {
		return nil, err
	}
	if cls.Status != 1 {
		return nil, errors.New("课程已下架")
	}

	clsPrice, _ := strconv.ParseFloat(cls.Price, 64)
	dockingID, _ := strconv.Atoi(cls.Docking)

	var money float64
	var addprice float64
	err = database.DB.QueryRow("SELECT COALESCE(money,0), COALESCE(addprice,1) FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&money, &addprice)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	var unitPrice float64
	if cls.Yunsuan == "+" {
		unitPrice = math.Round((clsPrice+addprice)*10000) / 10000
	} else {
		unitPrice = math.Round((clsPrice*addprice)*10000) / 10000
	}

	if mijia, ok, err := classmodule.LoadMiJia(uid, req.CID); err == nil && ok {
		unitPrice, _, _ = classmodule.ApplyMiJia(clsPrice, addprice, cls.Yunsuan, mijia.Mode, mijia.Price, 4)
	}

	var pledgeDiscount float64
	err = database.DB.QueryRow(`
		SELECT c.discount_rate FROM qingka_wangke_zhiya_records r
		JOIN qingka_wangke_zhiya_config c ON r.config_id = c.id
		JOIN qingka_wangke_class cl ON cl.cid = ?
		WHERE r.uid = ? AND r.status = 1 AND c.status = 1 AND c.category_id = cl.fenlei
		ORDER BY c.discount_rate ASC LIMIT 1`, req.CID, uid).Scan(&pledgeDiscount)
	if err == nil && pledgeDiscount > 0 && pledgeDiscount < 1 {
		unitPrice = math.Round(unitPrice*pledgeDiscount*10000) / 10000
	}

	if unitPrice == 0 || addprice < 0.1 {
		return nil, errors.New("价格异常，请联系管理员")
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return nil, errors.New("系统繁忙，请稍后重试")
	}
	defer tx.Rollback()

	err = tx.QueryRow("SELECT COALESCE(money,0) FROM qingka_wangke_user WHERE uid = ? FOR UPDATE", uid).Scan(&money)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	totalCost := float64(len(req.Data)) * unitPrice
	if money < totalCost {
		return nil, fmt.Errorf("余额不足，需要 %.2f 元，当前余额 %.2f 元", totalCost, money)
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	successCount := 0
	var totalDeducted float64
	var skippedCount int
	var skippedDetails []string

	for _, item := range req.Data {
		parts := strings.Fields(item.UserInfo)
		var school, user, pass string
		if len(parts) >= 3 {
			school = parts[0]
			user = parts[1]
			pass = parts[2]
		} else if len(parts) == 2 {
			school = "自动识别"
			user = parts[0]
			pass = parts[1]
		} else {
			continue
		}

		kcid := item.Data.ID
		kcname := item.Data.Name
		kcjs := item.Data.KCJS

		var dupCount int
		err := database.DB.QueryRow(
			"SELECT COUNT(*) FROM qingka_wangke_order WHERE uid = ? AND ptname = ? AND school = ? AND user = ? AND pass = ? AND kcid = ? AND kcname = ?",
			uid, cls.Name, school, user, pass, kcid, kcname,
		).Scan(&dupCount)
		if err == nil && dupCount > 0 {
			skippedCount++
			skippedDetails = append(skippedDetails, fmt.Sprintf("%s-%s", user, kcname))
			continue
		}

		dockStatus := 0
		if dockingID == 0 {
			dockStatus = 99
		}

		_, err = tx.Exec(
			"INSERT INTO qingka_wangke_order (uid, cid, hid, ptname, school, name, user, pass, kcid, kcname, courseEndTime, fees, noun, addtime, ip, dockstatus) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			uid, cls.CID, dockingID, cls.Name, school, item.UserName, user, pass, kcid, kcname, kcjs, fmt.Sprintf("%.4f", unitPrice), cls.Noun, now, "", dockStatus,
		)
		if err != nil {
			continue
		}

		tx.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ?", unitPrice, uid)
		tx.Exec(
			"INSERT INTO qingka_wangke_moneylog (uid, type, money, balance, remark, addtime) VALUES (?, '扣费', ?, (SELECT money FROM qingka_wangke_user WHERE uid = ?), ?, ?)",
			uid, -unitPrice, uid, fmt.Sprintf("%s %s %s %s 扣除%.2f 元", cls.Name, user, pass, kcname, unitPrice), now,
		)

		successCount++
		totalDeducted += unitPrice
		// 已停用逐级返利：
		// 现有代理充值按费率折算扣费，若继续在下级下单后给上级返利，
		// 会与充值链路叠加形成套利，导致平台出现账务亏损。
	}

	if successCount == 0 {
		return nil, errors.New("提交失败，请检查下单信息")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.New("提交订单失败，请重试")
	}

	database.DB.Exec("UPDATE qingka_wangke_user SET order_count = order_count + ? WHERE uid = ?", successCount, uid)

	return &model.OrderAddResult{
		SuccessCount: successCount,
		SkippedCount: skippedCount,
		TotalCost:    totalDeducted,
		SkippedItems: skippedDetails,
	}, nil
}

func (r *legacyRepository) AddOrdersForMall(bUID, tid, cUID int, retailPrice float64, outTradeNo string, req model.OrderAddRequest) (*model.OrderAddResult, error) {
	cls, err := r.sup.GetClassFull(req.CID)
	if err != nil {
		return nil, err
	}
	if cls.Status != 1 {
		return nil, errors.New("课程已下架")
	}

	clsPrice, _ := strconv.ParseFloat(cls.Price, 64)
	dockingID, _ := strconv.Atoi(cls.Docking)

	var money, mallMoney, addprice float64
	err = database.DB.QueryRow("SELECT COALESCE(money,0), COALESCE(mall_money,0), COALESCE(addprice,1) FROM qingka_wangke_user WHERE uid=?", bUID).Scan(&money, &mallMoney, &addprice)
	if err != nil {
		return nil, errors.New("商家账户异常")
	}

	var supplyPrice float64
	if cls.Yunsuan == "+" {
		supplyPrice = math.Round((clsPrice+addprice)*10000) / 10000
	} else {
		supplyPrice = math.Round((clsPrice*addprice)*10000) / 10000
	}
	if supplyPrice <= 0 {
		return nil, errors.New("供货价异常，请联系平台")
	}

	var pledgeDiscount float64
	err = database.DB.QueryRow(`
		SELECT c.discount_rate FROM qingka_wangke_zhiya_records r
		JOIN qingka_wangke_zhiya_config c ON r.config_id = c.id
		JOIN qingka_wangke_class cl ON cl.cid = ?
		WHERE r.uid = ? AND r.status = 1 AND c.status = 1 AND c.category_id = cl.fenlei
		ORDER BY c.discount_rate ASC LIMIT 1`, req.CID, bUID).Scan(&pledgeDiscount)
	if err == nil && pledgeDiscount > 0 && pledgeDiscount < 1 {
		supplyPrice = math.Round(supplyPrice*pledgeDiscount*10000) / 10000
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return nil, errors.New("系统繁忙，请稍后重试")
	}
	defer tx.Rollback()

	err = tx.QueryRow("SELECT COALESCE(money,0), COALESCE(mall_money,0) FROM qingka_wangke_user WHERE uid=? FOR UPDATE", bUID).Scan(&money, &mallMoney)
	if err != nil {
		return nil, errors.New("商家账户异常")
	}

	totalCost := float64(len(req.Data)) * supplyPrice
	if money+mallMoney < totalCost {
		return nil, fmt.Errorf("余额不足，主余额 %.2f 元，商城钱包 %.2f 元，需扣 %.2f 元", money, mallMoney, totalCost)
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	successCount := 0
	var totalDeducted float64
	var allOIDs []int64
	var insertErrors []string

	for _, item := range req.Data {
		parts := strings.Fields(item.UserInfo)
		var school, user, pass string
		if len(parts) >= 3 {
			school, user, pass = parts[0], parts[1], parts[2]
		} else if len(parts) == 2 {
			school, user, pass = "自动识别", parts[0], parts[1]
		} else {
			continue
		}

		kcid := item.Data.ID
		kcname := item.Data.Name
		kcjs := item.Data.KCJS

		dockStatus := 0
		if dockingID == 0 {
			dockStatus = 99
		}

		result, err := tx.Exec(
			`INSERT INTO qingka_wangke_order
			 (uid, cid, hid, ptname, school, name, user, pass, kcid, kcname, courseEndTime, fees, noun, addtime, ip, dockstatus, tid, c_uid, retail_fees, out_trade_no)
			 VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			bUID, cls.CID, dockingID, cls.Name, school, item.UserName, user, pass,
			kcid, kcname, kcjs,
			fmt.Sprintf("%.4f", supplyPrice), cls.Noun, now, "", dockStatus,
			tid, cUID, fmt.Sprintf("%.2f", retailPrice), outTradeNo,
		)
		if err != nil {
			insertErrors = append(insertErrors, fmt.Sprintf("%s-%s: %s", user, kcname, err.Error()))
			continue
		}

		remaining := supplyPrice
		if money > 0 {
			deductMain := math.Min(money, remaining)
			if deductMain > 0 {
				if _, err := tx.Exec("UPDATE qingka_wangke_user SET money=money-? WHERE uid=?", deductMain, bUID); err == nil {
					money -= deductMain
					remaining -= deductMain
					tx.Exec(
						"INSERT INTO qingka_wangke_moneylog (uid,type,money,balance,remark,addtime) VALUES (?,'商城扣费',?,(SELECT money FROM qingka_wangke_user WHERE uid=?),?,?)",
						bUID, -deductMain, bUID, fmt.Sprintf("商城订单 %s %s %s 优先扣主余额 %.2f 元", cls.Name, user, kcname, deductMain), now,
					)
				}
			}
		}
		if remaining > 0 {
			if _, err := tx.Exec("UPDATE qingka_wangke_user SET mall_money=mall_money-? WHERE uid=?", remaining, bUID); err == nil {
				mallMoney -= remaining
				tx.Exec(
					"INSERT INTO qingka_wangke_moneylog (uid,type,money,balance,remark,addtime) VALUES (?,'商城扣费',?,(SELECT mall_money FROM qingka_wangke_user WHERE uid=?),?,?)",
					bUID, -remaining, bUID, fmt.Sprintf("商城订单 %s %s %s 补扣商城钱包 %.2f 元", cls.Name, user, kcname, remaining), now,
				)
			}
		}

		var insertedOID int64
		if oid, e := result.LastInsertId(); e == nil {
			insertedOID = oid
		}
		successCount++
		totalDeducted += supplyPrice
		if insertedOID > 0 {
			allOIDs = append(allOIDs, insertedOID)
		}
	}

	if successCount == 0 {
		if len(insertErrors) > 0 {
			return nil, fmt.Errorf("提交失败: %s", strings.Join(insertErrors, "; "))
		}
		return nil, errors.New("提交失败，请检查下单信息")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.New("提交订单失败，请重试")
	}

	return &model.OrderAddResult{
		SuccessCount: successCount,
		TotalCost:    totalDeducted,
		OIDs:         allOIDs,
	}, nil
}

func (r *legacyRepository) ChangeStatus(uid int, grade string, req model.OrderStatusRequest) error {
	if len(req.OIDs) == 0 {
		return errors.New("请选择订单")
	}

	placeholders := make([]string, len(req.OIDs))
	args := make([]interface{}, 0, len(req.OIDs)+1)

	if req.Type == 1 || req.Type == 2 {
		args = append(args, req.Status)
	} else {
		return errors.New("无效操作类型")
	}

	for i, oid := range req.OIDs {
		placeholders[i] = "?"
		args = append(args, oid)
	}
	oidIn := strings.Join(placeholders, ",")

	sqlStr := ""
	if req.Type == 1 {
		sqlStr = fmt.Sprintf("UPDATE qingka_wangke_order SET status = ? WHERE oid IN (%s)", oidIn)
	} else {
		sqlStr = fmt.Sprintf("UPDATE qingka_wangke_order SET dockstatus = ? WHERE oid IN (%s)", oidIn)
	}

	if grade != "2" && grade != "3" {
		sqlStr += " AND uid = ?"
		args = append(args, uid)
	}

	_, err := database.DB.Exec(sqlStr, args...)
	if err == nil && req.Type == 1 {
		for _, oid := range req.OIDs {
			orderStatusNotifier(oid, req.Status, "", "")
		}
	}
	return err
}

func (r *legacyRepository) CancelOrder(uid int, grade string, oid int) error {
	if grade != "2" && grade != "3" {
		var orderUID int
		err := database.DB.QueryRow("SELECT uid FROM qingka_wangke_order WHERE oid = ?", oid).Scan(&orderUID)
		if err != nil {
			return errors.New("订单不存在")
		}
		if orderUID != uid {
			return errors.New("无权限")
		}
	}
	_, err := database.DB.Exec("UPDATE qingka_wangke_order SET status = '已取消', dockstatus = '4' WHERE oid = ?", oid)
	return err
}

func (r *legacyRepository) RefundOrders(uid int, grade string, oids []int) error {
	if len(oids) == 0 {
		return errors.New("请选择订单")
	}
	if grade != "2" && grade != "3" {
		return errors.New("无权限")
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	for _, oid := range oids {
		var feesStr string
		var orderUID int
		var kcname string
		err := database.DB.QueryRow("SELECT uid, COALESCE(fees,'0'), COALESCE(kcname,'') FROM qingka_wangke_order WHERE oid = ?", oid).Scan(&orderUID, &feesStr, &kcname)
		if err != nil {
			continue
		}
		fees, _ := strconv.ParseFloat(feesStr, 64)

		tx, err := database.DB.Begin()
		if err != nil {
			continue
		}
		tx.Exec("UPDATE qingka_wangke_user SET money = money + ? WHERE uid = ?", fees, orderUID)
		tx.Exec("UPDATE qingka_wangke_order SET dockstatus = '6', status = '已退款' WHERE oid = ?", oid)
		tx.Exec(
			"INSERT INTO qingka_wangke_moneylog (uid, type, money, balance, remark, addtime) VALUES (?, '退款', ?, (SELECT money FROM qingka_wangke_user WHERE uid = ?), ?, ?)",
			orderUID, fees, orderUID, fmt.Sprintf("订单%d %s 退款%.2f元", oid, kcname, fees), now,
		)
		tx.Commit()
	}
	return nil
}

func (r *legacyRepository) ModifyRemarks(oids []int, remarks string) error {
	if len(oids) == 0 {
		return errors.New("请选择订单")
	}
	for _, oid := range oids {
		database.DB.Exec("UPDATE qingka_wangke_order SET remarks = ? WHERE oid = ?", remarks, oid)
	}
	return nil
}

func (r *legacyRepository) ManualDockOrders(oids []int) (int, int, error) {
	if len(oids) == 0 {
		return 0, 0, errors.New("请选择订单")
	}
	success, fail := 0, 0

	for _, oid := range oids {
		var cid, hid int
		var school, user, pass, kcid, kcname string
		err := database.DB.QueryRow(
			"SELECT cid, COALESCE(hid,0), COALESCE(school,''), COALESCE(user,''), COALESCE(pass,''), COALESCE(kcid,''), COALESCE(kcname,'') FROM qingka_wangke_order WHERE oid = ?",
			oid,
		).Scan(&cid, &hid, &school, &user, &pass, &kcid, &kcname)
		if err != nil {
			fail++
			continue
		}

		cls, err := r.sup.GetClassFull(cid)
		if err != nil {
			database.DB.Exec("UPDATE qingka_wangke_order SET dockstatus = 2, remarks = ? WHERE oid = ?", fmt.Sprintf("课程不存在: %v", err), oid)
			fail++
			continue
		}

		docking, _ := strconv.Atoi(cls.Docking)
		if docking == 0 {
			database.DB.Exec("UPDATE qingka_wangke_order SET dockstatus = 99 WHERE oid = ?", oid)
			fail++
			continue
		}

		sup, err := r.sup.GetSupplierByHID(docking)
		if err != nil {
			database.DB.Exec("UPDATE qingka_wangke_order SET dockstatus = 2, remarks = ? WHERE oid = ?", fmt.Sprintf("供应商不存在: %v", err), oid)
			fail++
			continue
		}

		result, err := r.sup.CallSupplierOrder(sup, cls, school, user, pass, kcid, kcname, nil)
		if err != nil {
			database.DB.Exec("UPDATE qingka_wangke_order SET dockstatus = 2, remarks = ? WHERE oid = ?", fmt.Sprintf("对接失败: %s", err.Error()), oid)
			fail++
			continue
		}

		if result.Code == 1 {
			database.DB.Exec("UPDATE qingka_wangke_order SET dockstatus = 1, yid = ?, hid = ?, status = '进行中' WHERE oid = ?", result.YID, docking, oid)
			success++
		} else {
			database.DB.Exec("UPDATE qingka_wangke_order SET dockstatus = 2, remarks = ? WHERE oid = ?", fmt.Sprintf("对接失败: %s", result.Msg), oid)
			fail++
		}
	}

	return success, fail, nil
}

// SyncOrderProgress 处理“手动同步订单进度”。
// 这条链路的输入是明确的 oid 列表，特点是：
// 1. 只处理用户/管理员选中的订单；
// 2. 每条订单单独查询上游，不做跨供应商聚合；
// 3. 成功后立即回写主订单表并触发状态通知。
func (r *legacyRepository) SyncOrderProgress(oids []int) (int, error) {
	if len(oids) == 0 {
		return 0, errors.New("请选择订单")
	}
	updated := 0

	for _, oid := range oids {
		syncTime := time.Now().Format("2006-01-02 15:04:05")
		// 先把本地主订单的“查询上游”所需最小字段取出来：
		// yid/hid 用于定位上游订单和供应商，
		// user/kcname/noun 用于那些要求附带账号、课程名、上游商品ID 的平台做回退匹配。
		var yidStr, hidStr string
		var user, kcname, kcid, noun, status string
		err := database.DB.QueryRow(
			"SELECT COALESCE(yid,''), COALESCE(hid,'0'), COALESCE(user,''), COALESCE(kcname,''), COALESCE(kcid,''), COALESCE(noun,''), COALESCE(status,'') FROM qingka_wangke_order WHERE oid = ?",
			oid,
		).Scan(&yidStr, &hidStr, &user, &kcname, &kcid, &noun, &status)
		hid, _ := strconv.Atoi(hidStr)
		if err != nil || hid == 0 {
			continue
		}
		// 已退款/已退单的订单不再主动向上游查进度，
		// 避免把本地已经终态的业务重新改回“进行中”等中间态。
		if status == "已退款" || status == "已退单" {
			continue
		}

		// 通过 hid 找到本地保存的供应商配置。
		// 这里的供应商对象会携带平台类型、域名、密钥等信息，供 QueryOrderProgress 组装上游请求。
		sup, err := r.sup.GetSupplierByHID(hid)
		if err != nil {
			continue
		}

		// kcname 作为额外上下文传下去，主要是兼容某些平台只支持“用户名 + 课程名”联合查询。
		orderExtra := map[string]string{
			"kcname":       kcname,
			"kcid":         kcid,
			"noun":         noun,
			"__debug_http": "1",
			"__debug_oid":  strconv.Itoa(oid),
		}
		items, err := r.sup.QueryOrderProgress(sup, yidStr, user, orderExtra)
		if err != nil {
			database.DB.Exec("UPDATE qingka_wangke_order SET updatetime = ? WHERE oid = ?", syncTime, oid)
			continue
		}
		if len(items) == 0 {
			database.DB.Exec("UPDATE qingka_wangke_order SET updatetime = ? WHERE oid = ?", syncTime, oid)
		}

		for _, item := range items {
			// 不同平台返回的状态字段并不完全一致。
			// 优先使用 status_text 这种更接近中文业务态的字段，没有时再退回原始 status。
			statusText := item.Status
			if item.StatusText != "" {
				statusText = item.StatusText
			}
			// 更新时保留了旧系统的匹配条件：oid 是主键定位，
			// user + kcname 则作为额外保护，避免极端情况下把别的课程结果错写到当前订单上。
			database.DB.Exec(
				"UPDATE qingka_wangke_order SET name = ?, yid = ?, status = ?, process = ?, remarks = ?, courseStartTime = ?, courseEndTime = ?, examStartTime = ?, examEndTime = ?, updatetime = ? WHERE user = ? AND kcname = ? AND oid = ?",
				item.KCName, item.YID, statusText, item.Process, item.Remarks,
				item.CourseStartTime, item.CourseEndTime, item.ExamStartTime, item.ExamEndTime, syncTime,
				item.User, item.KCName, oid,
			)
			// 回写后立即通知站内推送/前端状态感知逻辑，
			// 这样手动点“同步上游进度”后，订单通知链路也能保持一致。
			orderStatusNotifier(oid, statusText, item.Process, item.Remarks)
			updated++
		}
	}

	return updated, nil
}

// AutoSyncAllProgress 处理“主订单表全局自动轮询”。
// 它和手动同步最大的区别是：
// 1. 不接收外部传入 oid，而是自己扫描所有已对接订单；
// 2. 为了降低上游接口压力，先按供应商 hid 分组，再并发轮询；
// 3. 会输出较详细的运行日志，便于观察几万单的大批量同步进度。
func (r *legacyRepository) AutoSyncAllProgress(opts AutoSyncOptions) (int, int, error) {
	logf := func(format string, args ...interface{}) {
		msg := fmt.Sprintf(format, args...)
		log.Printf("[AutoSync] %s", msg)
		if opts.LogCollector != nil {
			opts.LogCollector(msg)
		}
	}
	query := `
		SELECT oid, COALESCE(yid,''), COALESCE(hid,'0'),
			COALESCE((SELECT pt FROM qingka_wangke_huoyuan WHERE hid = qingka_wangke_order.hid LIMIT 1),''),
			COALESCE(user,''), COALESCE(kcname,''), COALESCE(noun,''), COALESCE(kcid,''), COALESCE(status,''), COALESCE(addtime,''), COALESCE(updatetime,'')
		FROM qingka_wangke_order
		WHERE dockstatus = 1`
	args := make([]interface{}, 0, 8)
	if opts.RecentHours > 0 {
		query += " AND addtime >= DATE_SUB(NOW(), INTERVAL ? HOUR)"
		args = append(args, opts.RecentHours)
	}
	if len(opts.SupplierHIDs) > 0 {
		placeholders := make([]string, 0, len(opts.SupplierHIDs))
		for _, hid := range opts.SupplierHIDs {
			if hid <= 0 {
				continue
			}
			placeholders = append(placeholders, "?")
			args = append(args, hid)
		}
		if len(placeholders) > 0 {
			query += " AND hid IN (" + strings.Join(placeholders, ",") + ")"
		}
	}
	if len(opts.ExcludedStatuses) > 0 {
		placeholders := make([]string, 0, len(opts.ExcludedStatuses))
		for _, status := range opts.ExcludedStatuses {
			status = strings.TrimSpace(status)
			if status == "" {
				continue
			}
			placeholders = append(placeholders, "?")
			args = append(args, status)
		}
		if len(placeholders) > 0 {
			query += " AND COALESCE(status,'') NOT IN (" + strings.Join(placeholders, ",") + ")"
		}
	}
	query += " ORDER BY oid DESC"

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()

	// 先把所有待轮询订单按 hid 分组。
	// 这样后面每个 goroutine 处理一个供应商分组，日志和上游限流都更容易控制。
	hidGroups := map[int][]autoSyncItem{}
	totalCount := 0
	now := time.Now()
	for rows.Next() {
		var item autoSyncItem
		var hidStr string
		if err := rows.Scan(&item.OID, &item.YID, &hidStr, &item.PT, &item.User, &item.KCName, &item.Noun, &item.KCID, &item.Status, &item.AddTime, &item.UpdateTime); err != nil {
			continue
		}
		item.HID, _ = strconv.Atoi(hidStr)
		if item.HID <= 0 {
			continue
		}
		hasBatchProgress := r.sup.HasBatchProgressForPT(item.PT)
		if opts.OnlyBatchSuppliers && !hasBatchProgress {
			continue
		}
		if opts.SkipBatchSuppliers && hasBatchProgress {
			continue
		}

		if !opts.IgnoreRules && len(opts.Rules) > 0 {
			addTime, ok := parseOrderTime(item.AddTime)
			if !ok {
				continue
			}
			ageHours := now.Sub(addTime).Hours()
			rule, matched := matchAutoSyncRule(ageHours, opts.Rules)
			if !matched {
				continue
			}
			if updateTime, ok := parseOrderTime(item.UpdateTime); ok {
				interval := time.Duration(rule.IntervalMinutes) * time.Minute
				if now.Sub(updateTime) < interval {
					continue
				}
			}
			item.MatchedRule = rule
		}

		hidGroups[item.HID] = append(hidGroups[item.HID], item)
		totalCount++
	}
	if err := rows.Err(); err != nil {
		setLastAutoSyncReport(AutoSyncReport{})
		return 0, 0, err
	}
	if totalCount == 0 {
		logf("当前没有可同步的已对接订单")
		setLastAutoSyncReport(AutoSyncReport{})
		return 0, 0, nil
	}

	logf("开始同步 %d 个已对接订单（%d 个供应商）", totalCount, len(hidGroups))

	verboseLogging := autoSyncVerboseLogging()
	var updatedCount int64
	var errorCount int64
	var processedCount int64
	var sampleErrorMu sync.Mutex
	var sampleErrors []string
	var supplierNameMu sync.Mutex
	supplierNames := map[string]bool{}
	sem := make(chan struct{}, 5)
	var wg sync.WaitGroup
	progressLogStep := autoSyncProgressLogStep()
	const sampleErrorLimit = 8

	// 只保留少量失败样例，避免几万单场景把日志刷爆。
	recordSampleError := func(message string) {
		sampleErrorMu.Lock()
		defer sampleErrorMu.Unlock()
		if len(sampleErrors) < sampleErrorLimit {
			sampleErrors = append(sampleErrors, message)
		}
	}

	recordSupplierName := func(name string) {
		name = strings.TrimSpace(name)
		if name == "" {
			return
		}
		supplierNameMu.Lock()
		defer supplierNameMu.Unlock()
		supplierNames[name] = true
	}

	logProgress := func(delta int64) {
		if delta <= 0 {
			return
		}
		// processedCount 统计的是“已经处理完的订单数”，
		// 不区分成功还是失败，只用来给运维一个可见的推进刻度。
		current := atomic.AddInt64(&processedCount, delta)
		previous := current - delta
		if current == int64(totalCount) || current/progressLogStep != previous/progressLogStep {
			logf(
				"进度 %d/%d，已更新 %d，失败 %d",
				current,
				totalCount,
				atomic.LoadInt64(&updatedCount),
				atomic.LoadInt64(&errorCount),
			)
		}
	}

	for hid, items := range hidGroups {
		sem <- struct{}{}
		wg.Add(1)

		go func(hid int, items []autoSyncItem) {
			// 使用固定大小的信号量限制供应商并发，避免同时打爆太多上游。
			defer func() {
				<-sem
				wg.Done()
			}()

			// 同一个 hid 分组内的订单都共用同一份供应商配置。
			sup, err := r.sup.GetSupplierByHID(hid)
			if err != nil {
				message := fmt.Sprintf("供应商 hid=%d 查询失败: %v（影响 %d 个订单）", hid, err, len(items))
				recordSupplierName(fmt.Sprintf("hid=%d", hid))
				if verboseLogging {
					log.Printf("[AutoSync] %s", message)
				}
				recordSampleError(message)
				atomic.AddInt64(&errorCount, int64(len(items)))
				logProgress(int64(len(items)))
				return
			}
			recordSupplierName(sup.Name)

			if r.sup.HasBatchProgressConfig(sup) {
				syncTime := time.Now().Format("2006-01-02 15:04:05")
				progressItems, err := r.sup.QueryBatchOrderProgress(sup, buildBatchProgressRefs(items))
				if err != nil {
					message := fmt.Sprintf("hid=%d 批量进度接口失败，回退逐单同步: %v", hid, err)
					if opts.OnlyBatchSuppliers {
						message = fmt.Sprintf("hid=%d 批量进度接口失败: %v", hid, err)
					}
					if verboseLogging {
						log.Printf("[AutoSync] %s", message)
					}
					recordSampleError(message)
					if opts.OnlyBatchSuppliers {
						atomic.AddInt64(&errorCount, int64(len(items)))
						logProgress(int64(len(items)))
						return
					}
				} else {
					byYID, byNounUserCourse := indexAutoSyncItems(items)
					updatedOIDs := map[int]bool{}

					for _, progress := range progressItems {
						matchedItems := matchAutoSyncItems(progress, byYID, byNounUserCourse)
						if len(matchedItems) == 0 {
							if verboseLogging {
								log.Printf("[AutoSync] hid=%d 批量进度项未匹配本地订单: yid=%s noun=%s user=%s kcname=%s", hid, progress.YID, progress.Noun, progress.User, progress.KCName)
							}
							continue
						}
						for _, item := range matchedItems {
							if err := applyAutoSyncProgressUpdate(item, progress, syncTime); err != nil {
								message := fmt.Sprintf("oid=%d 批量更新进度失败(hid=%d yid=%s): %v", item.OID, hid, item.YID, err)
								if verboseLogging {
									log.Printf("[AutoSync] %s", message)
								}
								recordSampleError(message)
								atomic.AddInt64(&errorCount, 1)
								continue
							}
							updatedOIDs[item.OID] = true
							atomic.AddInt64(&updatedCount, 1)
						}
					}

					if !opts.OnlyBatchSuppliers {
						for _, item := range items {
							if updatedOIDs[item.OID] {
								continue
							}
							touchAutoSyncOrder(item.OID, syncTime)
						}
					}
					if verboseLogging {
						log.Printf("[AutoSync] hid=%d 批量进度同步完成，返回 %d 条变更，命中 %d 个订单", hid, len(progressItems), len(updatedOIDs))
					}
					logProgress(int64(len(items)))
					return
				}
			}

			for _, item := range items {
				syncTime := time.Now().Format("2006-01-02 15:04:05")
				// 自动轮询仍然把 kcname 透传给供应商查询层，
				// 这样与手动同步保持相同的平台兼容行为。
				orderExtra := map[string]string{
					"kcname": item.KCName,
					"noun":   item.Noun,
				}
				if item.KCID != "" {
					orderExtra["kcid"] = item.KCID
				}
				progressItems, err := r.sup.QueryOrderProgress(sup, item.YID, item.User, orderExtra)
				if err != nil {
					touchAutoSyncOrder(item.OID, syncTime)
					message := fmt.Sprintf("oid=%d 查询进度失败(hid=%d yid=%s): %v", item.OID, hid, item.YID, err)
					if verboseLogging {
						log.Printf("[AutoSync] %s", message)
					}
					recordSampleError(message)
					atomic.AddInt64(&errorCount, 1)
					logProgress(1)
					continue
				}
				if len(progressItems) == 0 {
					touchAutoSyncOrder(item.OID, syncTime)
					// “接口成功但 data 为空”在业务上也视为失败，
					// 因为这说明当前订单没有拿到任何可回写的上游进度。
					message := fmt.Sprintf("oid=%d 未查询到上游进度(hid=%d yid=%s)", item.OID, hid, item.YID)
					if verboseLogging {
						log.Printf("[AutoSync] %s", message)
					}
					recordSampleError(message)
					atomic.AddInt64(&errorCount, 1)
					logProgress(1)
					continue
				}

				for _, progress := range progressItems {
					if err := applyAutoSyncProgressUpdate(item, progress, syncTime); err != nil {
						message := fmt.Sprintf("oid=%d 更新进度失败(hid=%d yid=%s): %v", item.OID, hid, item.YID, err)
						if verboseLogging {
							log.Printf("[AutoSync] %s", message)
						}
						recordSampleError(message)
						atomic.AddInt64(&errorCount, 1)
						continue
					}
					atomic.AddInt64(&updatedCount, 1)
				}
				logProgress(1)
			}
		}(hid, items)
	}

	wg.Wait()
	if len(sampleErrors) > 0 {
		logf("失败样例: %s", strings.Join(sampleErrors, "；"))
	}
	names := make([]string, 0, len(supplierNames))
	for name := range supplierNames {
		names = append(names, name)
	}
	sort.Strings(names)
	setLastAutoSyncReport(AutoSyncReport{
		SupplierNames: names,
		SampleErrors:  append([]string(nil), sampleErrors...),
		Processed:     totalCount,
	})
	return int(updatedCount), int(errorCount), nil
}

func (r *legacyRepository) BatchSyncOrders(oids []int) (int, error) {
	if len(oids) == 0 {
		return 0, errors.New("请选择要同步的订单")
	}
	return r.SyncOrderProgress(oids)
}

func (r *legacyRepository) BatchResendOrders(oids []int) (int, int, error) {
	if len(oids) == 0 {
		return 0, 0, errors.New("请选择要补单的订单")
	}

	success, fail := 0, 0
	for _, oid := range oids {
		var hid int
		var yid, status string
		err := database.DB.QueryRow(
			"SELECT COALESCE(hid,0), COALESCE(yid,''), COALESCE(status,'') FROM qingka_wangke_order WHERE oid = ?",
			oid,
		).Scan(&hid, &yid, &status)
		if err != nil || hid == 0 || yid == "" || yid == "0" {
			fail++
			continue
		}
		if status == "已退款" || status == "已取消" {
			fail++
			continue
		}

		sup, err := r.sup.GetSupplierByHID(hid)
		if err != nil {
			fail++
			continue
		}

		code, msg, err := r.sup.ResubmitOrder(sup, yid)
		if err != nil {
			database.DB.Exec("UPDATE qingka_wangke_order SET remarks = ? WHERE oid = ?", fmt.Sprintf("补单失败: %s", err.Error()), oid)
			fail++
			continue
		}
		if code == 1 || code == 0 {
			now := time.Now().Format("2006-01-02 15:04:05")
			database.DB.Exec(
				"UPDATE qingka_wangke_order SET status = '补刷中', dockstatus = 1, remarks = ?, bsnum = bsnum + 1 WHERE oid = ?",
				fmt.Sprintf("补刷成功，等待进度更新。补刷时间：%s", now), oid,
			)
			success++
		} else {
			database.DB.Exec("UPDATE qingka_wangke_order SET remarks = ? WHERE oid = ?", fmt.Sprintf("补单失败: %s", msg), oid)
			fail++
		}
	}
	return success, fail, nil
}

func (r *legacyRepository) distributeCommission(buyerUID int, costPrice float64, yunsuan string, buyerUnitPrice float64, count int) {
	currentUID := buyerUID
	currentChildUnitPrice := buyerUnitPrice

	for i := 0; i < 10; i++ {
		var parentUID int
		err := database.DB.QueryRow("SELECT COALESCE(uuid, 0) FROM qingka_wangke_user WHERE uid = ?", currentUID).Scan(&parentUID)
		if err != nil || parentUID <= 1 {
			break
		}

		var parentAddPrice float64
		err = database.DB.QueryRow("SELECT COALESCE(addprice, 1) FROM qingka_wangke_user WHERE uid = ?", parentUID).Scan(&parentAddPrice)
		if err != nil {
			break
		}

		var parentUnitPrice float64
		if yunsuan == "+" {
			parentUnitPrice = math.Round((costPrice+parentAddPrice)*10000) / 10000
		} else {
			parentUnitPrice = math.Round((costPrice*parentAddPrice)*10000) / 10000
		}

		profit := (currentChildUnitPrice - parentUnitPrice) * float64(count)
		profit = math.Round(profit*10000) / 10000

		if profit > 0 {
			_, err = database.DB.Exec("UPDATE qingka_wangke_user SET money = money + ? WHERE uid = ?", profit, parentUID)
			if err == nil {
				now := time.Now().Format("2006-01-02 15:04:05")
				remark := fmt.Sprintf("分销提成: 下级UID %d 下单, 数量 %d, 提成 %.4f 元", buyerUID, count, profit)
				database.DB.Exec(
					"INSERT INTO qingka_wangke_moneylog (uid, type, money, balance, remark, addtime) VALUES (?, '提成', ?, (SELECT money FROM qingka_wangke_user WHERE uid = ?), ?, ?)",
					parentUID, profit, parentUID, remark, now,
				)
			}
		}

		currentUID = parentUID
		currentChildUnitPrice = parentUnitPrice
	}
}
