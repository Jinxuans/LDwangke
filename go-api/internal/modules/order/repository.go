package order

import (
	"errors"
	"fmt"
	"go-api/internal/database"
	"go-api/internal/model"
	suppliermodule "go-api/internal/modules/supplier"
	"go-api/internal/queue"
	shared "go-api/internal/shared/db"
	"math"
	"strconv"
	"strings"
	"time"
)

// Repository 为订单模块提供最小的存取边界，当前先复用旧 service 实现。
type Repository interface {
	List(uid int, grade string, req model.OrderListRequest) ([]model.Order, int64, error)
	Detail(uid int, grade string, oid int) (*model.Order, error)
	Stats(uid int, grade string) (*model.OrderStats, error)
	AddOrders(uid int, req model.OrderAddRequest) (*model.OrderAddResult, error)
	AddOrdersForMall(bUID, tid, cUID int, retailPrice float64, req model.OrderAddRequest) (*model.OrderAddResult, error)
	ChangeStatus(uid int, grade string, req model.OrderStatusRequest) error
	CancelOrder(uid int, grade string, oid int) error
	RefundOrders(uid int, grade string, oids []int) error
	ModifyRemarks(oids []int, remarks string) error
	ManualDockOrders(oids []int) (int, int, error)
	SyncOrderProgress(oids []int) (int, error)
	BatchSyncOrders(oids []int) (int, error)
	BatchResendOrders(oids []int) (int, int, error)
}

type supplierGateway interface {
	GetSupplierByHID(hid int) (*model.SupplierFull, error)
	GetClassFull(cid int) (*model.ClassFull, error)
	CallSupplierOrder(sup *model.SupplierFull, cls *model.ClassFull, school, user, pass, kcid, kcname string, extraFields map[string]string) (*model.SupplierOrderResult, error)
	QueryOrderProgress(sup *model.SupplierFull, yid string, username string, orderExtra map[string]string) ([]model.SupplierProgressItem, error)
	ResubmitOrder(sup *model.SupplierFull, yid string) (int, string, error)
}

type legacyRepository struct {
	orders *shared.OrderRepo
	sup    supplierGateway
}

func NewRepository() Repository {
	return &legacyRepository{
		orders: shared.NewOrderRepo(),
		sup:    suppliermodule.SharedService(),
	}
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
	var pendingDockOIDs []int64
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

		result, err := tx.Exec(
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

		if dockStatus == 0 {
			if oid, e := result.LastInsertId(); e == nil {
				pendingDockOIDs = append(pendingDockOIDs, oid)
			}
		}

		successCount++
		totalDeducted += unitPrice
		r.distributeCommission(uid, clsPrice, cls.Yunsuan, unitPrice, 1)
	}

	if successCount == 0 {
		return nil, errors.New("提交失败，请检查下单信息")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.New("提交订单失败，请重试")
	}

	database.DB.Exec("UPDATE qingka_wangke_user SET order_count = order_count + ? WHERE uid = ?", successCount, uid)
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

func (r *legacyRepository) AddOrdersForMall(bUID, tid, cUID int, retailPrice float64, req model.OrderAddRequest) (*model.OrderAddResult, error) {
	cls, err := r.sup.GetClassFull(req.CID)
	if err != nil {
		return nil, err
	}
	if cls.Status != 1 {
		return nil, errors.New("课程已下架")
	}

	clsPrice, _ := strconv.ParseFloat(cls.Price, 64)
	dockingID, _ := strconv.Atoi(cls.Docking)

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

		dockStatus := 0
		if dockingID == 0 {
			dockStatus = 99
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

func (r *legacyRepository) SyncOrderProgress(oids []int) (int, error) {
	if len(oids) == 0 {
		return 0, errors.New("请选择订单")
	}
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
		if status == "已退款" || status == "已退单" {
			continue
		}

		sup, err := r.sup.GetSupplierByHID(hid)
		if err != nil {
			continue
		}

		orderExtra := map[string]string{"kcname": kcname}
		items, err := r.sup.QueryOrderProgress(sup, yidStr, user, orderExtra)
		if err != nil {
			continue
		}

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
			orderStatusNotifier(oid, statusText, item.Process, item.Remarks)
			updated++
		}
	}

	return updated, nil
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
