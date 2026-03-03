package service

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go-api/internal/database"
	"go-api/internal/model"
	"go-api/internal/queue"
)

type OrderService struct{}

func NewOrderService() *OrderService {
	return &OrderService{}
}

const orderColumns = "oid, uid, cid, hid, COALESCE(ptname,''), COALESCE(school,''), COALESCE(name,''), COALESCE(user,''), COALESCE(pass,''), COALESCE(kcname,''), COALESCE(kcid,''), COALESCE(status,'待处理'), COALESCE(fees,'0'), COALESCE(process,''), COALESCE(remarks,''), COALESCE(dockstatus,'0'), COALESCE(yid,''), COALESCE(addtime,''), COALESCE(pushUid,''), COALESCE(pushStatus,''), COALESCE(pushEmail,''), COALESCE(pushEmailStatus,'0'), COALESCE(showdoc_push_url,''), COALESCE(pushShowdocStatus,'0'), COALESCE((SELECT pt FROM qingka_wangke_huoyuan WHERE hid=qingka_wangke_order.hid LIMIT 1),'')"

func scanOrder(rows *sql.Rows) (model.Order, error) {
	var o model.Order
	err := rows.Scan(&o.OID, &o.UID, &o.CID, &o.HID, &o.PTName, &o.School, &o.Name, &o.User, &o.Pass, &o.KCName, &o.KCID, &o.Status, &o.Fees, &o.Process, &o.Remarks, &o.DockStatus, &o.YID, &o.AddTime, &o.PushUid, &o.PushStatus, &o.PushEmail, &o.PushEmailStatus, &o.ShowdocPushURL, &o.PushShowdocStatus, &o.SupplierPT)
	return o, err
}

func (s *OrderService) List(uid int, grade string, req model.OrderListRequest) ([]model.Order, int64, error) {
	where := []string{"1=1"}
	args := []interface{}{}

	// 非管理员只能看自己的订单
	if grade != "2" && grade != "3" {
		where = append(where, "uid = ?")
		args = append(args, uid)
	} else if req.UID != "" {
		where = append(where, "uid = ?")
		args = append(args, req.UID)
	}

	if req.StatusText != "" {
		where = append(where, "status = ?")
		args = append(args, req.StatusText)
	}
	if req.CID != "" {
		where = append(where, "cid = ?")
		args = append(args, req.CID)
	}
	if req.Dock != "" {
		where = append(where, "dockstatus = ?")
		args = append(args, req.Dock)
	}
	if req.OID != "" {
		where = append(where, "oid = ?")
		args = append(args, req.OID)
	}
	if req.HID != "" {
		where = append(where, "hid = ?")
		args = append(args, req.HID)
	}
	if req.User != "" {
		where = append(where, "user = ?")
		args = append(args, req.User)
	}
	if req.Pass != "" {
		where = append(where, "pass = ?")
		args = append(args, req.Pass)
	}
	if req.School != "" {
		where = append(where, "school LIKE ?")
		args = append(args, "%"+req.School+"%")
	}
	if req.KCName != "" {
		where = append(where, "kcname LIKE ?")
		args = append(args, "%"+req.KCName+"%")
	}
	if req.Search != "" {
		where = append(where, "(uid LIKE ? OR ptname LIKE ? OR school LIKE ? OR user LIKE ? OR pass LIKE ? OR kcname LIKE ? OR process LIKE ? OR remarks LIKE ?)")
		s := "%" + req.Search + "%"
		args = append(args, s, s, s, s, s, s, s, s)
	}

	// 限制最近 100 天 (addtime 是 varchar, 用字符串比较)
	cutoff := time.Now().AddDate(0, 0, -100).Format("2006-01-02")
	where = append(where, "addtime >= ?")
	args = append(args, cutoff)

	whereStr := strings.Join(where, " AND ")

	// 总数
	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM qingka_wangke_order WHERE %s", whereStr)
	if err := database.DB.QueryRow(countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 2000 {
		req.Limit = 2000
	}

	offset := (req.Page - 1) * req.Limit
	querySQL := fmt.Sprintf("SELECT %s FROM qingka_wangke_order WHERE %s ORDER BY oid DESC LIMIT ? OFFSET ?", orderColumns, whereStr)
	args = append(args, req.Limit, offset)

	rows, err := database.DB.Query(querySQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			continue
		}
		orders = append(orders, o)
	}
	if orders == nil {
		orders = []model.Order{}
	}

	return orders, total, nil
}

func (s *OrderService) Detail(uid int, grade string, oid int) (*model.Order, error) {
	querySQL := fmt.Sprintf("SELECT %s FROM qingka_wangke_order WHERE oid = ?", orderColumns)
	args := []interface{}{oid}

	// 非管理员需限制 uid
	if grade != "2" && grade != "3" {
		querySQL += " AND uid = ?"
		args = append(args, uid)
	}

	rows, err := database.DB.Query(querySQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, errors.New("订单不存在")
	}
	o, err := scanOrder(rows)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (s *OrderService) Stats(uid int, grade string) (*model.OrderStats, error) {
	stats := &model.OrderStats{}

	whereUID := "WHERE uid = ?"
	uidArg := uid
	if grade == "2" || grade == "3" {
		whereUID = "WHERE 1=1"
		uidArg = 0
	}

	var scanArgs []interface{}
	if grade == "2" || grade == "3" {
		_ = database.DB.QueryRow("SELECT COUNT(*), COALESCE(SUM(fees),0) FROM qingka_wangke_order").Scan(&stats.Total, &stats.TotalFees)
		_ = database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE status = '进行中'").Scan(&stats.Processing)
		_ = database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE status = '已完成'").Scan(&stats.Completed)
		_ = database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE status = '异常'").Scan(&stats.Failed)
	} else {
		_ = database.DB.QueryRow("SELECT COUNT(*), COALESCE(SUM(fees),0) FROM qingka_wangke_order WHERE uid = ?", uidArg).Scan(&stats.Total, &stats.TotalFees)
		_ = database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE uid = ? AND status = '进行中'", uidArg).Scan(&stats.Processing)
		_ = database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE uid = ? AND status = '已完成'", uidArg).Scan(&stats.Completed)
		_ = database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE uid = ? AND status = '异常'", uidArg).Scan(&stats.Failed)
	}
	_ = scanArgs
	_ = whereUID

	return stats, nil
}

func (s *OrderService) ChangeStatus(uid int, grade string, req model.OrderStatusRequest) error {
	if len(req.OIDs) == 0 {
		return errors.New("请选择订单")
	}

	placeholders := make([]string, len(req.OIDs))
	args := make([]interface{}, 0, len(req.OIDs)+1)

	if req.Type == 1 {
		// 修改任务状态 status
		args = append(args, req.Status)
	} else if req.Type == 2 {
		// 修改处理状态 dockstatus
		args = append(args, req.Status)
	} else {
		return errors.New("无效操作类型")
	}

	for i, oid := range req.OIDs {
		placeholders[i] = "?"
		args = append(args, oid)
	}
	oidIn := strings.Join(placeholders, ",")

	var sqlStr string
	if req.Type == 1 {
		sqlStr = fmt.Sprintf("UPDATE qingka_wangke_order SET status = ? WHERE oid IN (%s)", oidIn)
	} else {
		sqlStr = fmt.Sprintf("UPDATE qingka_wangke_order SET dockstatus = ? WHERE oid IN (%s)", oidIn)
	}

	// 非管理员只能操作自己的
	if grade != "2" && grade != "3" {
		sqlStr += " AND uid = ?"
		args = append(args, uid)
	}

	_, err := database.DB.Exec(sqlStr, args...)
	if err == nil && req.Type == 1 {
		// 手动修改任务状态时，触发推送通知
		for _, oid := range req.OIDs {
			NotifyOrderStatusChange(oid, req.Status, "", "")
		}
	}
	return err
}

// AddOrders 交单逻辑：查课后下单
func (s *OrderService) AddOrders(uid int, req model.OrderAddRequest) (*model.OrderAddResult, error) {
	supService := NewSupplierService()

	// 1. 查课程信息
	cls, err := supService.GetClassFull(req.CID)
	if err != nil {
		return nil, err
	}
	if cls.Status != 1 {
		return nil, errors.New("课程已下架")
	}

	// 解析 price/docking (DB: varchar)
	clsPrice, _ := strconv.ParseFloat(cls.Price, 64)
	dockingID, _ := strconv.Atoi(cls.Docking)

	// 2. 获取用户信息（余额、加价系数）
	var money float64
	var addprice float64
	err = database.DB.QueryRow("SELECT COALESCE(money,0), COALESCE(addprice,1) FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&money, &addprice)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	// 3. 计算单价 (参照PHP: $danjia = round($rs['price'] * $userrow['addprice'], 2) 或 +)
	var unitPrice float64
	if cls.Yunsuan == "+" {
		unitPrice = math.Round((clsPrice+addprice)*10000) / 10000
	} else {
		unitPrice = math.Round((clsPrice*addprice)*10000) / 10000
	}

	// 处理密价 (mode 0=减价, 1=底价*加价, 2=固定价, 4=倍率定价)
	var mijiaPrice float64
	var mijiaMode int
	err = database.DB.QueryRow("SELECT COALESCE(price,0), COALESCE(mode,0) FROM qingka_wangke_mijia WHERE uid = ? AND cid = ?", uid, req.CID).Scan(&mijiaPrice, &mijiaMode)
	if err == nil {
		switch mijiaMode {
		case 0:
			unitPrice = math.Round((unitPrice-mijiaPrice)*10000) / 10000
		case 1:
			unitPrice = math.Round(((clsPrice-mijiaPrice)*addprice)*10000) / 10000
		case 2:
			unitPrice = mijiaPrice
		case 4:
			unitPrice = math.Round((clsPrice*mijiaPrice)*10000) / 10000
		}
		if unitPrice < 0 {
			unitPrice = 0
		}
	}

	// 应用质押折扣：查用户是否有该课程分类的生效质押
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

	// 4. 使用事务保护余额操作
	tx, err := database.DB.Begin()
	if err != nil {
		return nil, errors.New("系统繁忙，请稍后重试")
	}
	defer tx.Rollback()

	// SELECT ... FOR UPDATE 锁定用户行，防止并发透支
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
	var pendingDockOIDs []int64

	// 6. 遍历下单
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

		// 检查重复订单（防重复下单）
		var dupCount int
		err := database.DB.QueryRow(
			"SELECT COUNT(*) FROM qingka_wangke_order WHERE uid = ? AND ptname = ? AND school = ? AND user = ? AND pass = ? AND kcid = ? AND kcname = ?",
			uid, cls.Name, school, user, pass, kcid, kcname,
		).Scan(&dupCount)

		if err == nil && dupCount > 0 {
			// 重复订单：跳过，不插入，不扣费
			skippedCount++
			skippedDetails = append(skippedDetails, fmt.Sprintf("%s-%s", user, kcname))
			continue
		}

		// 确定对接状态
		var dockStatus int
		if dockingID == 0 {
			dockStatus = 99 // 无对接
		} else {
			dockStatus = 0 // 待对接
		}

		// 插入订单 (事务内，与 PHP 一致的最小字段集)
		result, err := tx.Exec(
			"INSERT INTO qingka_wangke_order (uid, cid, hid, ptname, school, name, user, pass, kcid, kcname, courseEndTime, fees, noun, addtime, ip, dockstatus) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			uid, cls.CID, dockingID, cls.Name, school, item.UserName, user, pass, kcid, kcname, kcjs, fmt.Sprintf("%.4f", unitPrice), cls.Noun, now, "", dockStatus,
		)
		if err != nil {
			fmt.Printf("[AddOrders] INSERT failed: %v | userinfo=%s kcid=%s kcname=%s\n", err, item.UserInfo, kcid, kcname)
			continue
		}

		// 扣除余额（事务内）
		tx.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ?", unitPrice, uid)

		// 记录余额流水（事务内）
		tx.Exec(
			"INSERT INTO qingka_wangke_moneylog (uid, type, money, balance, remark, addtime) VALUES (?, '扣费', ?, (SELECT money FROM qingka_wangke_user WHERE uid = ?), ?, ?)",
			uid, -unitPrice, uid, fmt.Sprintf("%s %s %s %s 扣除%.2f 元", cls.Name, user, pass, kcname, unitPrice), now,
		)

		// 收集待对接的订单 ID
		if dockStatus == 0 {
			if oid, e := result.LastInsertId(); e == nil {
				pendingDockOIDs = append(pendingDockOIDs, oid)
			}
		}

		successCount++
		totalDeducted += unitPrice

		// 发放多级分销提成
		s.distributeCommission(uid, clsPrice, cls.Yunsuan, unitPrice, 1)
	}

	// 如果有重复订单被跳过，记录日志
	if skippedCount > 0 {
		fmt.Printf("[AddOrders] 跳过%d个重复订单：%v\n", skippedCount, skippedDetails)
	}

	if successCount == 0 {
		return nil, errors.New("提交失败，请检查下单信息")
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return nil, errors.New("提交订单失败，请重试")
	}

	// 更新用户订单计数
	database.DB.Exec("UPDATE qingka_wangke_user SET order_count = order_count + ? WHERE uid = ?", successCount, uid)

	// 通过并发队列对接上游供应商
	if len(pendingDockOIDs) > 0 && queue.GlobalDockQueue != nil {
		queue.GlobalDockQueue.PushBatch(pendingDockOIDs)
	}

	return &model.OrderAddResult{
		SuccessCount: successCount,
		SkippedCount: skippedCount,
		TotalCost:    totalDeducted,
		SkippedItems: skippedDetails,
	}, nil
}

// DockSingleOrder 对接单个订单（供队列 worker 调用）
// 带指数退避重试：网络/临时错误最多重试3次（间隔 3s/9s/27s）
func DockSingleOrder(oid int64) {
	supService := NewSupplierService()
	var cid int
	var hidStr string
	var school, user, pass, kcid, kcname, noun string
	err := database.DB.QueryRow(
		"SELECT cid, COALESCE(hid,'0'), COALESCE(school,''), COALESCE(user,''), COALESCE(pass,''), COALESCE(kcid,''), COALESCE(kcname,''), COALESCE(noun,'') FROM qingka_wangke_order WHERE oid = ? AND dockstatus = 0",
		oid,
	).Scan(&cid, &hidStr, &school, &user, &pass, &kcid, &kcname, &noun)
	if err != nil {
		fmt.Printf("[DockHandler] oid=%d 查询失败或非待对接: %v\n", oid, err)
		return
	}
	hid, _ := strconv.Atoi(hidStr)
	if hid == 0 {
		database.DB.Exec("UPDATE qingka_wangke_order SET dockstatus = 99 WHERE oid = ?", oid)
		return
	}
	cls, err := supService.GetClassFull(cid)
	if err != nil {
		database.DB.Exec("UPDATE qingka_wangke_order SET dockstatus = 2, remarks = ? WHERE oid = ?",
			fmt.Sprintf("课程不存在: %v", err), oid)
		return
	}
	sup, err := supService.GetSupplierByHID(hid)
	if err != nil {
		database.DB.Exec("UPDATE qingka_wangke_order SET dockstatus = 2, remarks = ? WHERE oid = ?",
			fmt.Sprintf("供应商不存在: %v", err), oid)
		return
	}

	// 构建 extraFields（传递订单中的扩展参数，如 noun/score 等）
	extraFields := map[string]string{}
	if noun != "" {
		extraFields["noun"] = noun
	}

	// 指数退避重试（网络/临时错误）
	const maxRetries = 3
	var lastErr error
	var result *model.SupplierOrderResult
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(math.Pow(3, float64(attempt))) * time.Second // 3s, 9s, 27s
			fmt.Printf("[DockHandler] oid=%d 第%d次重试，等待 %v\n", oid, attempt, delay)
			time.Sleep(delay)
			// 重试前检查订单是否还处于待对接状态（可能已被手动处理）
			var currentDock int
			database.DB.QueryRow("SELECT dockstatus FROM qingka_wangke_order WHERE oid = ?", oid).Scan(&currentDock)
			if currentDock != 0 {
				fmt.Printf("[DockHandler] oid=%d 已被处理(dockstatus=%d)，跳过重试\n", oid, currentDock)
				return
			}
		}
		result, lastErr = supService.CallSupplierOrder(sup, cls, school, user, pass, kcid, kcname, extraFields)
		if lastErr == nil {
			break
		}
		fmt.Printf("[DockHandler] oid=%d 第%d次尝试失败: %v\n", oid, attempt+1, lastErr)
	}

	if lastErr != nil {
		database.DB.Exec("UPDATE qingka_wangke_order SET dockstatus = 2, remarks = ? WHERE oid = ?",
			fmt.Sprintf("对接失败(重试%d次): %s", maxRetries, lastErr.Error()), oid)
		fmt.Printf("[DockHandler] oid=%d 对接最终失败: %v\n", oid, lastErr)
		return
	}
	if result.Code == 1 {
		// 对接成功：更新 dockstatus=1, yid, status=进行中
		database.DB.Exec("UPDATE qingka_wangke_order SET dockstatus = 1, yid = ?, status = '进行中' WHERE oid = ?",
			result.YID, oid)
		fmt.Printf("[DockHandler] oid=%d 对接成功，yid=%s\n", oid, result.YID)
		NotifyOrderStatusChange(int(oid), "进行中", "", "")
	} else {
		// 对接失败：更新 dockstatus=2, status=异常，并记录备注
		remarkText := fmt.Sprintf("对接失败：%s", result.Msg)
		database.DB.Exec("UPDATE qingka_wangke_order SET dockstatus = 2, status = '异常', remarks = ? WHERE oid = ?",
			remarkText, oid)
		fmt.Printf("[DockHandler] oid=%d 对接失败：%s\n", oid, result.Msg)
		NotifyOrderStatusChange(int(oid), "异常", "", remarkText)
	}
}

// autoDockOrders 异步将订单推送到上游供应商（保留兼容）
func (s *OrderService) autoDockOrders(oids []int64, sup *model.SupplierFull, cls *model.ClassFull) {
	supService := NewSupplierService()
	for _, oid := range oids {
		var school, user, pass, kcid, kcname, noun string
		err := database.DB.QueryRow(
			"SELECT COALESCE(school,''), COALESCE(user,''), COALESCE(pass,''), COALESCE(kcid,''), COALESCE(kcname,''), COALESCE(noun,'') FROM qingka_wangke_order WHERE oid = ?",
			oid,
		).Scan(&school, &user, &pass, &kcid, &kcname, &noun)
		if err != nil {
			continue
		}

		result, err := supService.CallSupplierOrder(sup, cls, school, user, pass, kcid, kcname, nil)
		if err != nil {
			// 对接失败，标记 dockstatus=2
			database.DB.Exec("UPDATE qingka_wangke_order SET dockstatus = 2, remarks = ? WHERE oid = ?",
				fmt.Sprintf("自动对接失败: %s", err.Error()), oid)
			continue
		}

		if result.Code == 1 {
			// 对接成功
			yid := result.YID
			database.DB.Exec("UPDATE qingka_wangke_order SET dockstatus = 1, yid = ?, status = '进行中' WHERE oid = ?", yid, oid)
		} else {
			// 对接返回错误
			database.DB.Exec("UPDATE qingka_wangke_order SET dockstatus = 2, remarks = ? WHERE oid = ?",
				fmt.Sprintf("对接失败: %s", result.Msg), oid)
		}
	}
}

// ManualDockOrders 管理员手动对接/重新对接订单到上游
func (s *OrderService) ManualDockOrders(oids []int) (int, int, error) {
	if len(oids) == 0 {
		return 0, 0, errors.New("请选择订单")
	}
	supService := NewSupplierService()
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

		// 获取课程和供应商
		cls, err := supService.GetClassFull(cid)
		if err != nil {
			database.DB.Exec("UPDATE qingka_wangke_order SET dockstatus = 2, remarks = ? WHERE oid = ?",
				fmt.Sprintf("课程不存在: %v", err), oid)
			fail++
			continue
		}

		docking, _ := strconv.Atoi(cls.Docking)
		if docking == 0 {
			database.DB.Exec("UPDATE qingka_wangke_order SET dockstatus = 99 WHERE oid = ?", oid)
			fail++
			continue
		}

		sup, err := supService.GetSupplierByHID(docking)
		if err != nil {
			database.DB.Exec("UPDATE qingka_wangke_order SET dockstatus = 2, remarks = ? WHERE oid = ?",
				fmt.Sprintf("供应商不存在: %v", err), oid)
			fail++
			continue
		}

		result, err := supService.CallSupplierOrder(sup, cls, school, user, pass, kcid, kcname, nil)
		if err != nil {
			database.DB.Exec("UPDATE qingka_wangke_order SET dockstatus = 2, remarks = ? WHERE oid = ?",
				fmt.Sprintf("对接失败: %s", err.Error()), oid)
			fail++
			continue
		}

		if result.Code == 1 {
			database.DB.Exec("UPDATE qingka_wangke_order SET dockstatus = 1, yid = ?, hid = ?, status = '进行中' WHERE oid = ?",
				result.YID, docking, oid)
			success++
		} else {
			database.DB.Exec("UPDATE qingka_wangke_order SET dockstatus = 2, remarks = ? WHERE oid = ?",
				fmt.Sprintf("对接失败: %s", result.Msg), oid)
			fail++
		}
	}

	return success, fail, nil
}

// SyncOrderProgress 从上游同步订单进度 (按 PHP uporder case + jdjk.php processCx)
func (s *OrderService) SyncOrderProgress(oids []int) (int, error) {
	if len(oids) == 0 {
		return 0, errors.New("请选择订单")
	}
	supService := NewSupplierService()
	updated := 0

	for _, oid := range oids {
		var yidStr, hidStr string
		var user, kcname, status string
		err := database.DB.QueryRow(
			"SELECT COALESCE(yid,''), COALESCE(hid,'0'), COALESCE(user,''), COALESCE(kcname,''), COALESCE(status,'') FROM qingka_wangke_order WHERE oid = ?",
			oid,
		).Scan(&yidStr, &hidStr, &user, &kcname, &status)
		hid, _ := strconv.Atoi(hidStr)
		if err != nil || hid == 0 {
			continue
		}

		// 按 PHP: 已退款/已退单的订单跳过
		if status == "已退款" || status == "已退单" {
			continue
		}

		sup, err := supService.GetSupplierByHID(hid)
		if err != nil {
			continue
		}

		orderExtra := map[string]string{"kcname": kcname}
		items, err := supService.QueryOrderProgress(sup, yidStr, user, orderExtra)
		if err != nil {
			continue
		}

		// 按 PHP uporder case: 遍历返回数据，匹配 user + kcname 更新
		for _, item := range items {
			statusText := item.Status
			if item.StatusText != "" {
				statusText = item.StatusText
			}
			database.DB.Exec(
				"UPDATE qingka_wangke_order SET name = ?, yid = ?, status = ?, process = ?, remarks = ?, courseStartTime = ?, courseEndTime = ?, examStartTime = ?, examEndTime = ? WHERE user = ? AND kcname = ? AND oid = ?",
				item.KCName, item.YID, statusText, item.Process, item.Remarks,
				item.CourseStartTime, item.CourseEndTime, item.ExamStartTime, item.ExamEndTime,
				item.User, item.KCName, oid,
			)
			NotifyOrderStatusChange(oid, statusText, item.Process, item.Remarks)
			updated++
		}
	}

	return updated, nil
}

// AutoSyncAllProgress 自动同步所有已对接订单的进度（定时任务调用）
// 改进：同步所有 dockstatus=1 的订单（不管 status 是什么），按供应商聚合减少重复查询，增加错误日志
func AutoSyncAllProgress() {
	supService := NewSupplierService()

	// 查询所有 dockstatus=1 的订单（无 LIMIT，全量同步）
	// 包括：待处理、进行中、补刷中、异常等所有已对接状态
	rows, err := database.DB.Query(
		"SELECT oid, COALESCE(yid,''), COALESCE(hid,'0'), COALESCE(user,''), COALESCE(kcname,'') FROM qingka_wangke_order WHERE dockstatus = 1 ORDER BY oid DESC",
	)
	if err != nil {
		fmt.Printf("[AutoSync] 查询进行中订单失败: %v\n", err)
		return
	}
	defer rows.Close()

	type syncItem struct {
		OID    int
		YID    string
		HID    int
		User   string
		KCName string
	}

	// 按 HID 分组，减少重复查询供应商
	hidGroups := map[int][]syncItem{}
	totalCount := 0
	for rows.Next() {
		var it syncItem
		var hidStr string
		rows.Scan(&it.OID, &it.YID, &hidStr, &it.User, &it.KCName)
		it.HID, _ = strconv.Atoi(hidStr)
		if it.HID > 0 {
			hidGroups[it.HID] = append(hidGroups[it.HID], it)
			totalCount++
		}
	}

	if totalCount == 0 {
		return
	}

	fmt.Printf("[AutoSync] 开始同步 %d 个进行中订单（%d 个供应商）\n", totalCount, len(hidGroups))
	var updatedCount int64
	var errorCount int64

	// 并发按供应商分组处理，最多5个供应商同时查上游
	sem := make(chan struct{}, 5)
	var wg sync.WaitGroup

	for hid, items := range hidGroups {
		sem <- struct{}{}
		wg.Add(1)
		go func(hid int, items []syncItem) {
			defer func() { <-sem; wg.Done() }()

			// 每个供应商只查一次
			sup, err := supService.GetSupplierByHID(hid)
			if err != nil {
				fmt.Printf("[AutoSync] 供应商 hid=%d 查询失败: %v（影响 %d 个订单）\n", hid, err, len(items))
				atomic.AddInt64(&errorCount, int64(len(items)))
				return
			}

			for _, it := range items {
				orderExtra := map[string]string{"kcname": it.KCName}
				progressItems, err := supService.QueryOrderProgress(sup, it.YID, it.User, orderExtra)
				if err != nil {
					fmt.Printf("[AutoSync] oid=%d 查询进度失败(hid=%d): %v\n", it.OID, hid, err)
					atomic.AddInt64(&errorCount, 1)
					continue
				}

				for _, p := range progressItems {
					statusText := p.Status
					if p.StatusText != "" {
						statusText = p.StatusText
					}
					database.DB.Exec(
						"UPDATE qingka_wangke_order SET name = ?, yid = ?, status = ?, process = ?, remarks = ?, courseStartTime = ?, courseEndTime = ?, examStartTime = ?, examEndTime = ? WHERE oid = ?",
						p.KCName, p.YID, statusText, p.Process, p.Remarks,
						p.CourseStartTime, p.CourseEndTime, p.ExamStartTime, p.ExamEndTime,
						it.OID,
					)
					NotifyOrderStatusChange(it.OID, statusText, p.Process, p.Remarks)
					atomic.AddInt64(&updatedCount, 1)
				}
			}
		}(hid, items)
	}
	wg.Wait()

	errCnt := atomic.LoadInt64(&errorCount)
	updCnt := atomic.LoadInt64(&updatedCount)
	if updCnt > 0 || errCnt > 0 {
		fmt.Printf("[AutoSync] 同步完成，更新 %d 个订单，失败 %d 个\n", updCnt, errCnt)
	}
}

// CancelOrder 取消订单 (按 PHP qx_order case)
func (s *OrderService) CancelOrder(uid int, grade string, oid int) error {
	// 检查权限：管理员或本人
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

// BatchSyncOrders 批量同步进度 (按 PHP batchCronOrder，改为同步执行)
func (s *OrderService) BatchSyncOrders(oids []int) (int, error) {
	if len(oids) == 0 {
		return 0, errors.New("请选择要同步的订单")
	}
	updated, err := s.SyncOrderProgress(oids)
	return updated, err
}

// BatchResendOrders 批量补单 (按 PHP batchResetOrder → 调用上游 act=budan)
func (s *OrderService) BatchResendOrders(oids []int) (int, int, error) {
	if len(oids) == 0 {
		return 0, 0, errors.New("请选择要补单的订单")
	}
	supService := NewSupplierService()
	success, fail := 0, 0

	for _, oid := range oids {
		var hidStr, yid, status string
		err := database.DB.QueryRow(
			"SELECT COALESCE(hid,'0'), COALESCE(yid,''), COALESCE(status,'') FROM qingka_wangke_order WHERE oid = ?",
			oid,
		).Scan(&hidStr, &yid, &status)
		if err != nil {
			fail++
			continue
		}
		hid, _ := strconv.Atoi(hidStr)
		if hid == 0 || yid == "" || yid == "0" {
			fail++
			continue
		}
		if status == "已退款" || status == "已取消" {
			fail++
			continue
		}

		sup, err := supService.GetSupplierByHID(hid)
		if err != nil {
			fail++
			continue
		}

		code, msg, err := supService.ResubmitOrder(sup, yid)
		if err != nil {
			database.DB.Exec("UPDATE qingka_wangke_order SET remarks = ? WHERE oid = ?",
				fmt.Sprintf("补单失败: %s", err.Error()), oid)
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
			database.DB.Exec("UPDATE qingka_wangke_order SET remarks = ? WHERE oid = ?",
				fmt.Sprintf("补单失败: %s", msg), oid)
			fail++
		}
	}
	return success, fail, nil
}

// ModifyRemarks 批量修改备注 (按 PHP xgbz case)
func (s *OrderService) ModifyRemarks(oids []int, remarks string) error {
	if len(oids) == 0 {
		return errors.New("请选择订单")
	}
	for _, oid := range oids {
		database.DB.Exec("UPDATE qingka_wangke_order SET remarks = ? WHERE oid = ?", remarks, oid)
	}
	return nil
}

func (s *OrderService) RefundOrders(uid int, grade string, oids []int) error {
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

		// 使用事务保护退款操作
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

// distributeCommission 递归向上发放分销提成
func (s *OrderService) distributeCommission(buyerUID int, costPrice float64, yunsuan string, buyerUnitPrice float64, count int) {
	currentUID := buyerUID
	currentChildUnitPrice := buyerUnitPrice

	// 向上追溯，最多支持10级，防止死循环
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

		// 计算上级应付单价
		var parentUnitPrice float64
		if yunsuan == "+" {
			parentUnitPrice = math.Round((costPrice+parentAddPrice)*10000) / 10000
		} else {
			parentUnitPrice = math.Round((costPrice*parentAddPrice)*10000) / 10000
		}

		// 提成 = (下级成本单价 - 本级成本单价) * 数量
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

// AddOrdersForMall C端商城下单：用B端uid扣供货价，额外记录 tid/c_uid/retail_fees
func (s *OrderService) AddOrdersForMall(bUID, tid, cUID int, retailPrice float64, req model.OrderAddRequest) (*model.OrderAddResult, error) {
	supService := NewSupplierService()

	cls, err := supService.GetClassFull(req.CID)
	if err != nil {
		return nil, err
	}
	if cls.Status != 1 {
		return nil, errors.New("课程已下架")
	}

	clsPrice, _ := strconv.ParseFloat(cls.Price, 64)
	dockingID, _ := strconv.Atoi(cls.Docking)

	// 供货价：按B端用户的 addprice 计算
	var money, addprice float64
	err = database.DB.QueryRow("SELECT COALESCE(money,0), COALESCE(addprice,1) FROM qingka_wangke_user WHERE uid=?", bUID).Scan(&money, &addprice)
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

	// 应用质押折扣：查B端用户是否有该课程分类的生效质押
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

	err = tx.QueryRow("SELECT COALESCE(money,0) FROM qingka_wangke_user WHERE uid=? FOR UPDATE", bUID).Scan(&money)
	if err != nil {
		return nil, errors.New("商家账户异常")
	}

	totalCost := float64(len(req.Data)) * supplyPrice
	if money < totalCost {
		return nil, fmt.Errorf("商家余额不足，无法完成下单")
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	successCount := 0
	var totalDeducted float64
	var pendingDockOIDs []int64
	var allOIDs []int64
	var skippedCount int
	var skippedDetails []string

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

		var dupCount int
		database.DB.QueryRow(
			"SELECT COUNT(*) FROM qingka_wangke_order WHERE uid=? AND ptname=? AND school=? AND user=? AND pass=? AND kcid=? AND kcname=?",
			bUID, cls.Name, school, user, pass, kcid, kcname,
		).Scan(&dupCount)
		if dupCount > 0 {
			skippedCount++
			skippedDetails = append(skippedDetails, fmt.Sprintf("%s-%s", user, kcname))
			continue
		}

		var dockStatus int
		if dockingID == 0 {
			dockStatus = 99
		} else {
			dockStatus = 0
		}

		result, err := tx.Exec(
			`INSERT INTO qingka_wangke_order
			 (uid, cid, hid, ptname, school, name, user, pass, kcid, kcname, courseEndTime, fees, noun, addtime, ip, dockstatus, tid, c_uid, retail_fees)
			 VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			bUID, cls.CID, dockingID, cls.Name, school, item.UserName, user, pass,
			kcid, kcname, kcjs,
			fmt.Sprintf("%.4f", supplyPrice), cls.Noun, now, "", dockStatus,
			tid, cUID, fmt.Sprintf("%.2f", retailPrice),
		)
		if err != nil {
			continue
		}

		tx.Exec("UPDATE qingka_wangke_user SET money=money-? WHERE uid=?", supplyPrice, bUID)
		tx.Exec(
			"INSERT INTO qingka_wangke_moneylog (uid,type,money,balance,remark,addtime) VALUES (?,'扣费',?,(SELECT money FROM qingka_wangke_user WHERE uid=?),?,?)",
			bUID, -supplyPrice, bUID, fmt.Sprintf("商城订单 %s %s %s 扣除%.2f元", cls.Name, user, kcname, supplyPrice), now,
		)

		var insertedOID int64
		if oid, e := result.LastInsertId(); e == nil {
			insertedOID = oid
			if dockStatus == 0 {
				pendingDockOIDs = append(pendingDockOIDs, oid)
			}
		}
		successCount++
		totalDeducted += supplyPrice
		if insertedOID > 0 {
			allOIDs = append(allOIDs, insertedOID)
		}
	}

	if successCount == 0 {
		return nil, errors.New("提交失败，请检查下单信息")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.New("提交订单失败，请重试")
	}

	if len(pendingDockOIDs) > 0 && queue.GlobalDockQueue != nil {
		queue.GlobalDockQueue.PushBatch(pendingDockOIDs)
	}

	return &model.OrderAddResult{
		SuccessCount: successCount,
		SkippedCount: skippedCount,
		TotalCost:    totalDeducted,
		SkippedItems: skippedDetails,
		OIDs:         allOIDs,
	}, nil
}
