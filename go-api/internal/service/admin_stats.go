package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"go-api/internal/database"
)

func (s *AdminService) AdminMoneyLogList(page, limit int, uid string, logType string) ([]map[string]interface{}, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	where := "1=1"
	args := []interface{}{}
	if uid != "" {
		where += " AND m.uid = ?"
		args = append(args, uid)
	}
	if logType != "" {
		where += " AND m.type = ?"
		args = append(args, logType)
	}

	var total int64
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_moneylog m WHERE "+where, args...).Scan(&total)

	offset := (page - 1) * limit
	args2 := append(args, limit, offset)
	rows, err := database.DB.Query(
		fmt.Sprintf("SELECT m.id, m.uid, COALESCE(u.user,''), COALESCE(m.type,''), COALESCE(m.money,0), COALESCE(m.balance,0), COALESCE(m.remark,''), COALESCE(DATE_FORMAT(m.addtime,'%%Y-%%m-%%d %%H:%%i:%%s'),'') FROM qingka_wangke_moneylog m LEFT JOIN qingka_wangke_user u ON m.uid=u.uid WHERE %s ORDER BY m.id DESC LIMIT ? OFFSET ?", where),
		args2...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []map[string]interface{}
	for rows.Next() {
		var id, uid2 int
		var username, logType2, remark, addtime string
		var money, balance float64
		rows.Scan(&id, &uid2, &username, &logType2, &money, &balance, &remark, &addtime)
		list = append(list, map[string]interface{}{
			"id":       id,
			"uid":      uid2,
			"username": username,
			"type":     logType2,
			"money":    money,
			"balance":  balance,
			"remark":   remark,
			"addtime":  addtime,
		})
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	return list, total, nil
}

func (s *AdminService) DashboardStats() (map[string]interface{}, error) {
	stats := map[string]interface{}{}

	var userCount int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_user").Scan(&userCount)
	stats["user_count"] = userCount

	var todayNewUsers int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_user WHERE DATE(addtime) = CURDATE()").Scan(&todayNewUsers)
	stats["today_new_users"] = todayNewUsers

	var todayOrders int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE DATE(addtime) = CURDATE()").Scan(&todayOrders)
	stats["today_orders"] = todayOrders

	var yesterdayOrders int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE DATE(addtime) = CURDATE() - INTERVAL 1 DAY").Scan(&yesterdayOrders)
	stats["yesterday_orders"] = yesterdayOrders

	var todayIncome float64
	database.DB.QueryRow("SELECT COALESCE(SUM(fees), 0) FROM qingka_wangke_order WHERE DATE(addtime) = CURDATE()").Scan(&todayIncome)
	stats["today_income"] = todayIncome

	var yesterdayIncome float64
	database.DB.QueryRow("SELECT COALESCE(SUM(fees), 0) FROM qingka_wangke_order WHERE DATE(addtime) = CURDATE() - INTERVAL 1 DAY").Scan(&yesterdayIncome)
	stats["yesterday_income"] = yesterdayIncome

	var totalOrders int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order").Scan(&totalOrders)
	stats["total_orders"] = totalOrders

	var processingOrders int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE status NOT IN ('已完成','已退款','已取消','失败')").Scan(&processingOrders)
	stats["processing_orders"] = processingOrders

	var completedOrders int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE status = '已完成'").Scan(&completedOrders)
	stats["completed_orders"] = completedOrders

	var failedOrders int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE status = '异常'").Scan(&failedOrders)
	stats["failed_orders"] = failedOrders

	var totalBalance float64
	database.DB.QueryRow("SELECT COALESCE(SUM(money), 0) FROM qingka_wangke_user").Scan(&totalBalance)
	stats["total_balance"] = totalBalance

	stats["trend"] = s.getWeekTrend()
	stats["recent_orders"] = s.getRecentOrders()
	stats["status_distribution"] = s.getOrderStatusDistribution()
	stats["top_users"] = s.getTopUsers(7)

	return stats, nil
}

func (s *AdminService) getTopUsers(days int) []map[string]interface{} {
	rows, err := database.DB.Query(
		"SELECT o.uid, COALESCE(u.user,''), COUNT(*), COALESCE(SUM(o.fees),0) FROM qingka_wangke_order o LEFT JOIN qingka_wangke_user u ON o.uid=u.uid WHERE o.addtime >= CURDATE() - INTERVAL ? DAY GROUP BY o.uid ORDER BY SUM(o.fees) DESC LIMIT 5",
		days-1,
	)
	if err != nil {
		return []map[string]interface{}{}
	}
	defer rows.Close()

	var list []map[string]interface{}
	for rows.Next() {
		var uid, cnt int
		var username string
		var total float64
		rows.Scan(&uid, &username, &cnt, &total)
		list = append(list, map[string]interface{}{"uid": uid, "username": username, "orders": cnt, "total": total})
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	return list
}

func (s *AdminService) getWeekTrend() []map[string]interface{} {
	rows, err := database.DB.Query(
		"SELECT DATE(addtime) AS day, COUNT(*) AS cnt, COALESCE(SUM(fees),0) AS income FROM qingka_wangke_order WHERE addtime >= CURDATE() - INTERVAL 6 DAY GROUP BY DATE(addtime) ORDER BY day ASC",
	)
	if err != nil {
		return []map[string]interface{}{}
	}
	defer rows.Close()

	var trend []map[string]interface{}
	for rows.Next() {
		var day string
		var cnt int
		var income float64
		rows.Scan(&day, &cnt, &income)
		trend = append(trend, map[string]interface{}{
			"date":   day,
			"orders": cnt,
			"income": income,
		})
	}
	if trend == nil {
		trend = []map[string]interface{}{}
	}
	return trend
}

func (s *AdminService) getRecentOrders() []map[string]interface{} {
	rows, err := database.DB.Query(
		"SELECT oid, COALESCE(ptname,''), COALESCE(user,''), COALESCE(kcname,''), COALESCE(status,''), COALESCE(fees,0), COALESCE(DATE_FORMAT(addtime,'%Y-%m-%d %H:%i:%s'),'') FROM qingka_wangke_order ORDER BY oid DESC LIMIT 10",
	)
	if err != nil {
		return []map[string]interface{}{}
	}
	defer rows.Close()

	var list []map[string]interface{}
	for rows.Next() {
		var oid int
		var ptname, user, kcname, status, addtime string
		var fees float64
		rows.Scan(&oid, &ptname, &user, &kcname, &status, &fees, &addtime)
		list = append(list, map[string]interface{}{
			"oid":     oid,
			"ptname":  ptname,
			"user":    user,
			"kcname":  kcname,
			"status":  status,
			"fees":    fees,
			"addtime": addtime,
		})
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	return list
}

func (s *AdminService) getOrderStatusDistribution() []map[string]interface{} {
	rows, err := database.DB.Query(
		"SELECT COALESCE(status,'未知'), COUNT(*) FROM qingka_wangke_order GROUP BY status ORDER BY COUNT(*) DESC",
	)
	if err != nil {
		return []map[string]interface{}{}
	}
	defer rows.Close()

	var dist []map[string]interface{}
	for rows.Next() {
		var status string
		var cnt int
		rows.Scan(&status, &cnt)
		dist = append(dist, map[string]interface{}{
			"status": status,
			"count":  cnt,
		})
	}
	if dist == nil {
		dist = []map[string]interface{}{}
	}
	return dist
}

func (s *AdminService) StatsReport(days int) (map[string]interface{}, error) {
	if days <= 0 {
		days = 30
	}

	result := map[string]interface{}{}

	dailyRows, err := database.DB.Query(
		"SELECT DATE(addtime) AS day, COUNT(*) AS cnt, COALESCE(SUM(fees),0) AS income FROM qingka_wangke_order WHERE addtime >= CURDATE() - INTERVAL ? DAY GROUP BY DATE(addtime) ORDER BY day ASC",
		days-1,
	)
	if err == nil {
		defer dailyRows.Close()
		var daily []map[string]interface{}
		for dailyRows.Next() {
			var day string
			var cnt int
			var income float64
			dailyRows.Scan(&day, &cnt, &income)
			daily = append(daily, map[string]interface{}{"date": day, "orders": cnt, "income": income})
		}
		if daily == nil {
			daily = []map[string]interface{}{}
		}
		result["daily"] = daily
	}

	cateRows, err := database.DB.Query(
		"SELECT COALESCE(ptname,'未知'), COUNT(*), COALESCE(SUM(fees),0) FROM qingka_wangke_order WHERE addtime >= CURDATE() - INTERVAL ? DAY GROUP BY ptname ORDER BY COUNT(*) DESC LIMIT 20",
		days-1,
	)
	if err == nil {
		defer cateRows.Close()
		var byClass []map[string]interface{}
		for cateRows.Next() {
			var name string
			var cnt int
			var income float64
			cateRows.Scan(&name, &cnt, &income)
			byClass = append(byClass, map[string]interface{}{"name": name, "count": cnt, "income": income})
		}
		if byClass == nil {
			byClass = []map[string]interface{}{}
		}
		result["by_class"] = byClass
	}

	statusRows, err := database.DB.Query(
		"SELECT COALESCE(status,'未知'), COUNT(*) FROM qingka_wangke_order WHERE addtime >= CURDATE() - INTERVAL ? DAY GROUP BY status",
		days-1,
	)
	if err == nil {
		defer statusRows.Close()
		var byStatus []map[string]interface{}
		for statusRows.Next() {
			var status string
			var cnt int
			statusRows.Scan(&status, &cnt)
			byStatus = append(byStatus, map[string]interface{}{"status": status, "count": cnt})
		}
		if byStatus == nil {
			byStatus = []map[string]interface{}{}
		}
		result["by_status"] = byStatus
	}

	userRows, err := database.DB.Query(
		"SELECT o.uid, COALESCE(u.user,''), COUNT(*), COALESCE(SUM(o.fees),0) FROM qingka_wangke_order o LEFT JOIN qingka_wangke_user u ON o.uid=u.uid WHERE o.addtime >= CURDATE() - INTERVAL ? DAY GROUP BY o.uid ORDER BY SUM(o.fees) DESC LIMIT 10",
		days-1,
	)
	if err == nil {
		defer userRows.Close()
		var topUsers []map[string]interface{}
		for userRows.Next() {
			var uid int
			var username string
			var cnt int
			var total float64
			userRows.Scan(&uid, &username, &cnt, &total)
			topUsers = append(topUsers, map[string]interface{}{"uid": uid, "username": username, "orders": cnt, "total": total})
		}
		if topUsers == nil {
			topUsers = []map[string]interface{}{}
		}
		result["top_users"] = topUsers
	}

	return result, nil
}

func requireAdmin(grade string) error {
	if grade != "2" && grade != "3" {
		return errors.New("需要管理员权限")
	}
	return nil
}

type SupplierRankItem struct {
	HID            int    `json:"hid"`
	Name           string `json:"name"`
	TodayCount     int    `json:"today_count"`
	YesterdayCount int    `json:"yesterday_count"`
	TotalCount     int    `json:"total_count"`
}

func (s *AdminService) SupplierRanking() ([]SupplierRankItem, error) {
	rows, err := database.DB.Query("SELECT hid, COALESCE(name,'') FROM qingka_wangke_huoyuan WHERE status = 1")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hidMap := map[int]*SupplierRankItem{}
	var hidList []int
	for rows.Next() {
		var item SupplierRankItem
		rows.Scan(&item.HID, &item.Name)
		hidMap[item.HID] = &item
		hidList = append(hidList, item.HID)
	}

	if len(hidList) == 0 {
		return []SupplierRankItem{}, nil
	}

	placeholders := make([]string, len(hidList))
	args := make([]interface{}, len(hidList))
	for i, hid := range hidList {
		placeholders[i] = "?"
		args[i] = hid
	}
	inClause := strings.Join(placeholders, ",")

	now := time.Now()
	todayStart := now.Format("2006-01-02") + " 00:00:00"
	todayEnd := now.Format("2006-01-02") + " 23:59:59"
	yesterday := now.AddDate(0, 0, -1)
	yesterdayStart := yesterday.Format("2006-01-02") + " 00:00:00"
	yesterdayEnd := yesterday.Format("2006-01-02") + " 23:59:59"

	statsSQL := fmt.Sprintf(`
		SELECT hid, COUNT(*) as total_count,
			SUM(CASE WHEN addtime >= ? AND addtime <= ? THEN 1 ELSE 0 END) as today_count,
			SUM(CASE WHEN addtime >= ? AND addtime <= ? THEN 1 ELSE 0 END) as yesterday_count
		FROM qingka_wangke_order
		WHERE hid IN (%s)
		GROUP BY hid
	`, inClause)

	statsArgs := []interface{}{todayStart, todayEnd, yesterdayStart, yesterdayEnd}
	statsArgs = append(statsArgs, args...)

	statsRows, err := database.DB.Query(statsSQL, statsArgs...)
	if err != nil {
		return nil, err
	}
	defer statsRows.Close()

	for statsRows.Next() {
		var hid, totalCount, todayCount, yesterdayCount int
		statsRows.Scan(&hid, &totalCount, &todayCount, &yesterdayCount)
		if item, ok := hidMap[hid]; ok {
			item.TotalCount = totalCount
			item.TodayCount = todayCount
			item.YesterdayCount = yesterdayCount
		}
	}

	result := make([]SupplierRankItem, 0, len(hidMap))
	for _, item := range hidMap {
		result = append(result, *item)
	}
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].TotalCount > result[i].TotalCount {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result, nil
}

type AgentProductRankItem struct {
	Rank   int    `json:"rank"`
	PtName string `json:"ptname"`
	Count  int    `json:"count"`
	Latest string `json:"latest"`
}

func (s *AdminService) AgentProductRanking(uid int, timeType string, limit int) ([]AgentProductRankItem, error) {
	if uid <= 0 {
		return []AgentProductRankItem{}, nil
	}
	if limit <= 0 {
		limit = 20
	}

	now := time.Now()
	var startTime, endTime string

	switch timeType {
	case "yesterday":
		yesterday := now.AddDate(0, 0, -1)
		startTime = yesterday.Format("2006-01-02") + " 00:00:00"
		endTime = yesterday.Format("2006-01-02") + " 23:59:59"
	case "week":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		monday := now.AddDate(0, 0, -(weekday - 1))
		startTime = monday.Format("2006-01-02") + " 00:00:00"
		endTime = now.Format("2006-01-02") + " 23:59:59"
	case "month":
		startTime = now.Format("2006-01") + "-01 00:00:00"
		lastDay := time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 0, now.Location())
		endTime = lastDay.Format("2006-01-02") + " 23:59:59"
	default:
		startTime = now.Format("2006-01-02") + " 00:00:00"
		endTime = now.Format("2006-01-02") + " 23:59:59"
	}

	rows, err := database.DB.Query(
		"SELECT ptname, COUNT(*) AS cnt, MAX(addtime) as latest FROM qingka_wangke_order WHERE uid = ? AND addtime >= ? AND addtime <= ? GROUP BY ptname ORDER BY cnt DESC LIMIT ?",
		uid, startTime, endTime, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []AgentProductRankItem
	rank := 1
	for rows.Next() {
		var item AgentProductRankItem
		rows.Scan(&item.PtName, &item.Count, &item.Latest)
		item.Rank = rank
		rank++
		result = append(result, item)
	}
	if result == nil {
		result = []AgentProductRankItem{}
	}
	return result, nil
}
