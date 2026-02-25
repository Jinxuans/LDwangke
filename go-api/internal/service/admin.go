package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-api/internal/database"
	"go-api/internal/model"
)

type AdminService struct{}

func NewAdminService() *AdminService {
	return &AdminService{}
}

// ===== 用户管理 =====

func (s *AdminService) UserList(req model.UserListRequest) ([]model.UserManage, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}

	where := "1=1"
	args := []interface{}{}
	if req.Keywords != "" {
		where += " AND (user LIKE ? OR name LIKE ? OR uid = ?)"
		kw := "%" + req.Keywords + "%"
		args = append(args, kw, kw, req.Keywords)
	}

	var total int64
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_user WHERE "+where, args...).Scan(&total)

	// 加载等级名称映射 (rate -> name)
	gradeNameMap := s.loadGradeNameMap()

	offset := (req.Page - 1) * req.Limit
	args2 := append(args, req.Limit, offset)
	rows, err := database.DB.Query(
		fmt.Sprintf("SELECT uid, user, COALESCE(name,''), COALESCE(grade,''), COALESCE(addprice,1), COALESCE(money,0), COALESCE(yqm,''), COALESCE(addtime,''), COALESCE(active,1) FROM qingka_wangke_user WHERE %s ORDER BY uid DESC LIMIT ? OFFSET ?", where),
		args2...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []model.UserManage
	for rows.Next() {
		var u model.UserManage
		rows.Scan(&u.UID, &u.User, &u.Name, &u.Grade, &u.AddPrice, &u.Balance, &u.YQM, &u.AddTime, &u.Status)
		u.GradeName = gradeNameMap[fmt.Sprintf("%.2f", u.AddPrice)]
		if u.GradeName == "" {
			u.GradeName = fmt.Sprintf("费率%.2f", u.AddPrice)
		}
		users = append(users, u)
	}
	if users == nil {
		users = []model.UserManage{}
	}
	return users, total, nil
}

// loadGradeNameMap 加载 dengji 表的 rate→name 映射
func (s *AdminService) loadGradeNameMap() map[string]string {
	m := map[string]string{}
	rows, err := database.DB.Query("SELECT COALESCE(rate,''), COALESCE(name,'') FROM qingka_wangke_dengji WHERE status = '1' ORDER BY sort ASC")
	if err != nil {
		return m
	}
	defer rows.Close()
	for rows.Next() {
		var rate, name string
		rows.Scan(&rate, &name)
		m[rate] = name
		// 同时存储标准化格式，确保 "1" 和 "1.00" 都能匹配
		if f, err := strconv.ParseFloat(rate, 64); err == nil {
			m[fmt.Sprintf("%.2f", f)] = name
			m[fmt.Sprintf("%g", f)] = name
		}
	}
	return m
}

func (s *AdminService) UserResetPassword(uid int, newPass string) error {
	if newPass == "" {
		newPass = "123456"
	}
	_, err := database.DB.Exec("UPDATE qingka_wangke_user SET pass = ? WHERE uid = ?", newPass, uid)
	return err
}

func (s *AdminService) UserSetBalance(uid int, balance float64) error {
	// 查询旧余额
	var oldMoney float64
	database.DB.QueryRow("SELECT COALESCE(money,0) FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&oldMoney)

	_, err := database.DB.Exec("UPDATE qingka_wangke_user SET money = ? WHERE uid = ?", balance, uid)
	if err != nil {
		return err
	}

	// 记录余额流水
	diff := balance - oldMoney
	now := time.Now().Format("2006-01-02 15:04:05")
	remark := fmt.Sprintf("管理员调整余额 %.2f → %.2f", oldMoney, balance)
	database.DB.Exec(
		"INSERT INTO qingka_wangke_moneylog (uid, type, money, balance, remark, addtime) VALUES (?, '调整', ?, ?, ?, ?)",
		uid, diff, balance, remark, now,
	)
	return nil
}

func (s *AdminService) UserSetGrade(uid int, addprice float64) error {
	_, err := database.DB.Exec("UPDATE qingka_wangke_user SET addprice = ? WHERE uid = ?", addprice, uid)
	return err
}

// CategorySwitchesByCID 通过课程ID查询所属分类的开关配置
// 返回: log, ticket, changepass, allowpause, supplierReport, supplierReportHID, error
func (s *AdminService) CategorySwitchesByCID(cid int) (log, ticket, changepass, allowpause, supplierReport, supplierReportHID int, err error) {
	// 默认值：changepass=1 其余=0（与 DB DEFAULT 一致）
	changepass = 1
	err = database.DB.QueryRow(
		`SELECT COALESCE(f.log,0), COALESCE(f.ticket,0), COALESCE(f.changepass,1), COALESCE(f.allowpause,0), COALESCE(f.supplier_report,0), COALESCE(f.supplier_report_hid,0)
		 FROM qingka_wangke_class c
		 JOIN qingka_wangke_fenlei f ON f.id = CAST(c.fenlei AS UNSIGNED)
		 WHERE c.cid = ?`, cid,
	).Scan(&log, &ticket, &changepass, &allowpause, &supplierReport, &supplierReportHID)
	if err != nil {
		// 查不到分类时使用默认值，不阻塞操作
		return 0, 0, 1, 0, 0, 0, nil
	}
	return
}

// CategorySwitchesByOID 通过订单ID查询所属分类的开关配置
func (s *AdminService) CategorySwitchesByOID(oid int) (log, ticket, changepass, allowpause, supplierReport, supplierReportHID int, err error) {
	changepass = 1
	err = database.DB.QueryRow(
		`SELECT COALESCE(f.log,0), COALESCE(f.ticket,0), COALESCE(f.changepass,1), COALESCE(f.allowpause,0), COALESCE(f.supplier_report,0), COALESCE(f.supplier_report_hid,0)
		 FROM qingka_wangke_order o
		 JOIN qingka_wangke_class c ON c.cid = o.cid
		 JOIN qingka_wangke_fenlei f ON f.id = CAST(c.fenlei AS UNSIGNED)
		 WHERE o.oid = ?`, oid,
	).Scan(&log, &ticket, &changepass, &allowpause, &supplierReport, &supplierReportHID)
	if err != nil {
		return 0, 0, 1, 0, 0, 0, nil
	}
	return
}

// ===== 分类管理 =====

func (s *AdminService) CategoryListAll() ([]model.Category, error) {
	rows, err := database.DB.Query("SELECT id, COALESCE(name,''), COALESCE(status,'1'), COALESCE(sort,0), COALESCE(time,''), COALESCE(recommend,0), COALESCE(log,0), COALESCE(ticket,0), COALESCE(changepass,1), COALESCE(allowpause,0), COALESCE(supplier_report,0), COALESCE(supplier_report_hid,0) FROM qingka_wangke_fenlei ORDER BY sort ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cats []model.Category
	for rows.Next() {
		var c model.Category
		rows.Scan(&c.ID, &c.Name, &c.Status, &c.Sort, &c.Time, &c.Recommend, &c.Log, &c.Ticket, &c.ChangePass, &c.AllowPause, &c.SupplierReport, &c.SupplierReportHID)
		cats = append(cats, c)
	}
	if cats == nil {
		cats = []model.Category{}
	}
	return cats, nil
}

func (s *AdminService) CategoryListPaged(req model.CategoryListRequest) ([]model.Category, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}
	where := "1=1"
	args := []interface{}{}
	if req.Keyword != "" {
		where += " AND name LIKE ?"
		args = append(args, "%"+req.Keyword+"%")
	}
	if req.Status != "" {
		where += " AND status = ?"
		args = append(args, req.Status)
	}
	var total int64
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_fenlei WHERE "+where, args...).Scan(&total)

	offset := (req.Page - 1) * req.Limit
	args2 := append(args, req.Limit, offset)
	rows, err := database.DB.Query(
		fmt.Sprintf("SELECT id, COALESCE(name,''), COALESCE(status,'1'), COALESCE(sort,0), CASE WHEN time IS NOT NULL AND time != '' AND time != '0' THEN FROM_UNIXTIME(CAST(time AS UNSIGNED), '%%Y-%%m-%%d %%H:%%i') ELSE '' END, COALESCE(recommend,0), COALESCE(log,0), COALESCE(ticket,0), COALESCE(changepass,1), COALESCE(allowpause,0), COALESCE(supplier_report,0), COALESCE(supplier_report_hid,0) FROM qingka_wangke_fenlei WHERE %s ORDER BY sort ASC, id ASC LIMIT ? OFFSET ?", where),
		args2...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var cats []model.Category
	for rows.Next() {
		var c model.Category
		rows.Scan(&c.ID, &c.Name, &c.Status, &c.Sort, &c.Time, &c.Recommend, &c.Log, &c.Ticket, &c.ChangePass, &c.AllowPause, &c.SupplierReport, &c.SupplierReportHID)
		cats = append(cats, c)
	}
	if cats == nil {
		cats = []model.Category{}
	}
	return cats, total, nil
}

func (s *AdminService) CategorySave(cat model.Category) error {
	if cat.ID > 0 {
		_, err := database.DB.Exec("UPDATE qingka_wangke_fenlei SET name=?, status=?, sort=?, recommend=?, log=?, ticket=?, changepass=?, allowpause=?, supplier_report=?, supplier_report_hid=? WHERE id=?",
			cat.Name, cat.Status, cat.Sort, cat.Recommend, cat.Log, cat.Ticket, cat.ChangePass, cat.AllowPause, cat.SupplierReport, cat.SupplierReportHID, cat.ID)
		return err
	}
	_, err := database.DB.Exec("INSERT INTO qingka_wangke_fenlei (name, status, sort, time, recommend, log, ticket, changepass, allowpause, supplier_report, supplier_report_hid) VALUES (?, ?, ?, UNIX_TIMESTAMP(), ?, ?, ?, ?, ?, ?, ?)",
		cat.Name, cat.Status, cat.Sort, cat.Recommend, cat.Log, cat.Ticket, cat.ChangePass, cat.AllowPause, cat.SupplierReport, cat.SupplierReportHID)
	return err
}

func (s *AdminService) CategoryDelete(id int) error {
	// 同时删除分类内的商品
	database.DB.Exec("DELETE FROM qingka_wangke_class WHERE fenlei = ?", id)
	_, err := database.DB.Exec("DELETE FROM qingka_wangke_fenlei WHERE id = ?", id)
	return err
}

func (s *AdminService) CategoryQuickModify(keyword string, categoryID int) (int64, error) {
	result, err := database.DB.Exec("UPDATE qingka_wangke_class SET fenlei = ? WHERE name LIKE ?",
		categoryID, "%"+keyword+"%")
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// ===== 课程管理 =====

func (s *AdminService) ClassList(cateID int, keywords string, page, limit int) ([]model.ClassManage, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	where := []string{"1=1"}
	args := []interface{}{}
	if cateID > 0 {
		where = append(where, "fenlei = ?")
		args = append(args, cateID)
	}
	if keywords != "" {
		where = append(where, "(name LIKE ? OR cid = ?)")
		args = append(args, "%"+keywords+"%", keywords)
	}

	whereStr := strings.Join(where, " AND ")

	var total int64
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_class WHERE "+whereStr, args...).Scan(&total)

	offset := (page - 1) * limit
	args2 := append(args, limit, offset)
	rows, err := database.DB.Query(
		fmt.Sprintf("SELECT cid, COALESCE(name,''), COALESCE(price,'0'), COALESCE(content,''), COALESCE(fenlei,'0'), COALESCE(status,0), COALESCE(docking,'0'), COALESCE(sort,0), COALESCE(noun,''), COALESCE(yunsuan,'*') FROM qingka_wangke_class WHERE %s ORDER BY sort ASC, cid DESC LIMIT ? OFFSET ?", whereStr),
		args2...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var classes []model.ClassManage
	for rows.Next() {
		var c model.ClassManage
		rows.Scan(&c.CID, &c.Name, &c.Price, &c.Content, &c.CateID, &c.Status, &c.HID, &c.Sort, &c.Noun, &c.Yunsuan)
		classes = append(classes, c)
	}
	if classes == nil {
		classes = []model.ClassManage{}
	}
	return classes, total, nil
}

func (s *AdminService) ClassSave(req model.ClassEditRequest) error {
	if req.Yunsuan == "" {
		req.Yunsuan = "*"
	}
	if req.CID > 0 {
		_, err := database.DB.Exec(
			"UPDATE qingka_wangke_class SET name=?, price=?, content=?, fenlei=?, status=?, docking=?, sort=?, noun=?, yunsuan=? WHERE cid=?",
			req.Name, req.Price, req.Content, req.CateID, req.Status, req.HID, req.Sort, req.Noun, req.Yunsuan, req.CID)
		return err
	}
	_, err := database.DB.Exec(
		"INSERT INTO qingka_wangke_class (name, price, content, fenlei, status, docking, noun, getnoun, sort, yunsuan, addtime) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		req.Name, req.Price, req.Content, req.CateID, req.Status, req.HID, req.Noun, req.Noun, req.Sort, req.Yunsuan, time.Now().Format("2006-01-02 15:04:05"))
	return err
}

func (s *AdminService) ClassToggleStatus(cid, status int) error {
	_, err := database.DB.Exec("UPDATE qingka_wangke_class SET status = ? WHERE cid = ?", status, cid)
	return err
}

func (s *AdminService) ClassBatchDelete(cids []int) (int64, error) {
	if len(cids) == 0 {
		return 0, nil
	}
	placeholders := ""
	args := make([]interface{}, len(cids))
	for i, id := range cids {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args[i] = id
	}
	result, err := database.DB.Exec("DELETE FROM qingka_wangke_class WHERE cid IN ("+placeholders+")", args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (s *AdminService) ClassBatchCategory(cids []int, cateID string) (int64, error) {
	if len(cids) == 0 {
		return 0, nil
	}
	placeholders := ""
	args := []interface{}{cateID}
	for i, id := range cids {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args = append(args, id)
	}
	result, err := database.DB.Exec("UPDATE qingka_wangke_class SET fenlei = ? WHERE cid IN ("+placeholders+")", args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (s *AdminService) ClassBatchPrice(cids []int, rate float64, yunsuan string) (int64, error) {
	if len(cids) == 0 {
		return 0, nil
	}
	placeholders := ""
	args := []interface{}{}
	for i, id := range cids {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args = append(args, id)
	}
	var sql string
	if yunsuan == "+" {
		sql = fmt.Sprintf("UPDATE qingka_wangke_class SET price = CAST(CAST(price AS DECIMAL(10,2)) + %.2f AS CHAR) WHERE cid IN ("+placeholders+")", rate)
	} else {
		sql = fmt.Sprintf("UPDATE qingka_wangke_class SET price = CAST(ROUND(CAST(price AS DECIMAL(10,2)) * %.4f, 2) AS CHAR) WHERE cid IN ("+placeholders+")", rate)
	}
	result, err := database.DB.Exec(sql, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// ===== 货源管理 =====

func (s *AdminService) SupplierList() ([]model.Supplier, error) {
	rows, err := database.DB.Query(
		"SELECT hid, COALESCE(pt,''), COALESCE(name,''), COALESCE(url,''), COALESCE(user,''), COALESCE(pass,''), COALESCE(token,''), COALESCE(money,'0'), COALESCE(status,'1'), COALESCE(addtime,'') FROM qingka_wangke_huoyuan ORDER BY hid ASC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Supplier
	for rows.Next() {
		var sup model.Supplier
		rows.Scan(&sup.HID, &sup.PT, &sup.Name, &sup.URL, &sup.User, &sup.Pass, &sup.Token, &sup.Money, &sup.Status, &sup.AddTime)
		list = append(list, sup)
	}
	if list == nil {
		list = []model.Supplier{}
	}
	return list, nil
}

func (s *AdminService) SupplierSave(sup model.Supplier) error {
	if sup.HID > 0 {
		_, err := database.DB.Exec(
			"UPDATE qingka_wangke_huoyuan SET name=?, url=?, user=?, pass=?, token=?, pt=?, status=? WHERE hid=?",
			sup.Name, sup.URL, sup.User, sup.Pass, sup.Token, sup.PT, sup.Status, sup.HID)
		return err
	}
	_, err := database.DB.Exec(
		"INSERT INTO qingka_wangke_huoyuan (name, url, user, pass, token, pt, ip, cookie, money, status, addtime, endtime) VALUES (?, ?, ?, ?, ?, ?, '', '', '0', ?, NOW(), '')",
		sup.Name, sup.URL, sup.User, sup.Pass, sup.Token, sup.PT, sup.Status)
	return err
}

func (s *AdminService) SupplierDelete(hid int) error {
	_, err := database.DB.Exec("DELETE FROM qingka_wangke_huoyuan WHERE hid = ?", hid)
	return err
}

// ===== 公开站点配置（无需管理员权限） =====

// 允许公开访问的配置键（不含敏感信息）
var publicConfigKeys = map[string]bool{
	"sitename": true, "logo": true, "hlogo": true,
	"sykg": true, "bz": true, "notice": true, "tcgonggao": true,
	"flkg": true, "fllx": true, "fontsZDY": true, "fontsFamily": true,
	"qd_notice_open": true, "xdsmopen": true, "anti_debug": true,
	"version": true, "onlineStore_trdltz": true, "sjqykg": true,
	"user_yqzc": true, "login_slider_verify": true, "login_email_verify": true,
	"webVfx_open": true, "webVfx": true,
	"keywords": true, "description": true,
}

func (s *AdminService) GetPublicConfig() (map[string]string, error) {
	all, err := s.GetConfig()
	if err != nil {
		return map[string]string{}, nil
	}
	result := make(map[string]string)
	for k, v := range all {
		if publicConfigKeys[k] {
			result[k] = v
		}
	}
	return result, nil
}

// ===== 系统设置（管理员） =====

func (s *AdminService) GetConfig() (map[string]string, error) {
	rows, err := database.DB.Query("SELECT `v`, `k` FROM qingka_wangke_config")
	if err != nil {
		// 如果表不存在，返回空 map
		return map[string]string{}, nil
	}
	defer rows.Close()

	config := make(map[string]string)
	for rows.Next() {
		var k, v string
		rows.Scan(&k, &v)
		config[k] = v
	}
	return config, nil
}

func (s *AdminService) SaveConfig(configs map[string]string) error {
	for k, v := range configs {
		_, err := database.DB.Exec(
			"INSERT INTO qingka_wangke_config (`v`, `k`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `k` = ?",
			k, v, v,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// ===== 管理员支付配置 (paydata) =====

func (s *AdminService) GetPayData() (map[string]string, error) {
	var paydata string
	err := database.DB.QueryRow("SELECT COALESCE(paydata,'') FROM qingka_wangke_user WHERE uid = 1").Scan(&paydata)
	if err != nil {
		return map[string]string{}, nil
	}
	result := make(map[string]string)
	if paydata != "" {
		json.Unmarshal([]byte(paydata), &result)
	}
	return result, nil
}

func (s *AdminService) SavePayData(data map[string]string) error {
	// 先读取现有的 paydata，合并更新
	existing, _ := s.GetPayData()
	for k, v := range data {
		existing[k] = v
	}
	jsonBytes, err := json.Marshal(existing)
	if err != nil {
		return err
	}
	_, err = database.DB.Exec("UPDATE qingka_wangke_user SET paydata = ? WHERE uid = 1", string(jsonBytes))
	return err
}

// ===== 管理员余额流水 =====

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

// ===== 公告管理 =====

func (s *AdminService) AnnouncementList(req model.AnnouncementListRequest) ([]model.Announcement, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}

	where := "1=1"
	args := []interface{}{}
	if req.Keyword != "" {
		where += " AND (title LIKE ? OR content LIKE ?)"
		kw := "%" + req.Keyword + "%"
		args = append(args, kw, kw)
	}

	var total int64
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_gonggao WHERE "+where, args...).Scan(&total)

	offset := (req.Page - 1) * req.Limit
	args2 := append(args, req.Limit, offset)
	rows, err := database.DB.Query(
		fmt.Sprintf("SELECT id, COALESCE(title,''), COALESCE(content,''), COALESCE(time,''), COALESCE(uid,0), COALESCE(status,'1'), COALESCE(zhiding,'0'), COALESCE(author,'') FROM qingka_wangke_gonggao WHERE %s ORDER BY zhiding DESC, id DESC LIMIT ? OFFSET ?", where),
		args2...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []model.Announcement
	for rows.Next() {
		var a model.Announcement
		rows.Scan(&a.ID, &a.Title, &a.Content, &a.Time, &a.UID, &a.Status, &a.Zhiding, &a.Author)
		list = append(list, a)
	}
	if list == nil {
		list = []model.Announcement{}
	}
	return list, total, nil
}

func (s *AdminService) AnnouncementSave(uid int, username string, req model.AnnouncementSaveRequest) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	if req.Status == "" {
		req.Status = "1"
	}
	if req.Zhiding == "" {
		req.Zhiding = "0"
	}

	if req.ID > 0 {
		_, err := database.DB.Exec(
			"UPDATE qingka_wangke_gonggao SET title = ?, content = ?, status = ?, zhiding = ?, uptime = ? WHERE id = ?",
			req.Title, req.Content, req.Status, req.Zhiding, now, req.ID,
		)
		return err
	}

	_, err := database.DB.Exec(
		"INSERT INTO qingka_wangke_gonggao (title, content, time, uid, status, zhiding, uptime, author) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		req.Title, req.Content, now, uid, req.Status, req.Zhiding, now, username,
	)
	return err
}

func (s *AdminService) AnnouncementDelete(id int) error {
	_, err := database.DB.Exec("DELETE FROM qingka_wangke_gonggao WHERE id = ?", id)
	return err
}

// AnnouncementListPublic 获取已发布的公告（用户端）
func (s *AdminService) AnnouncementListPublic(page, limit int) ([]model.Announcement, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	var total int64
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_gonggao WHERE status = '1'").Scan(&total)

	offset := (page - 1) * limit
	rows, err := database.DB.Query(
		"SELECT id, COALESCE(title,''), COALESCE(content,''), COALESCE(time,''), COALESCE(uid,0), COALESCE(status,'1'), COALESCE(zhiding,'0'), COALESCE(author,'') FROM qingka_wangke_gonggao WHERE status = '1' ORDER BY zhiding DESC, id DESC LIMIT ? OFFSET ?",
		limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []model.Announcement
	for rows.Next() {
		var a model.Announcement
		rows.Scan(&a.ID, &a.Title, &a.Content, &a.Time, &a.UID, &a.Status, &a.Zhiding, &a.Author)
		list = append(list, a)
	}
	if list == nil {
		list = []model.Announcement{}
	}
	return list, total, nil
}

// ===== 仪表盘统计 =====

func (s *AdminService) DashboardStats() (map[string]interface{}, error) {
	stats := map[string]interface{}{}

	// 用户总数
	var userCount int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_user").Scan(&userCount)
	stats["user_count"] = userCount

	// 今日新增用户
	var todayNewUsers int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_user WHERE DATE(addtime) = CURDATE()").Scan(&todayNewUsers)
	stats["today_new_users"] = todayNewUsers

	// 今日订单数
	var todayOrders int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE DATE(addtime) = CURDATE()").Scan(&todayOrders)
	stats["today_orders"] = todayOrders

	// 昨日订单数
	var yesterdayOrders int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE DATE(addtime) = CURDATE() - INTERVAL 1 DAY").Scan(&yesterdayOrders)
	stats["yesterday_orders"] = yesterdayOrders

	// 今日收入
	var todayIncome float64
	database.DB.QueryRow("SELECT COALESCE(SUM(fees), 0) FROM qingka_wangke_order WHERE DATE(addtime) = CURDATE()").Scan(&todayIncome)
	stats["today_income"] = todayIncome

	// 昨日收入
	var yesterdayIncome float64
	database.DB.QueryRow("SELECT COALESCE(SUM(fees), 0) FROM qingka_wangke_order WHERE DATE(addtime) = CURDATE() - INTERVAL 1 DAY").Scan(&yesterdayIncome)
	stats["yesterday_income"] = yesterdayIncome

	// 总订单数
	var totalOrders int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order").Scan(&totalOrders)
	stats["total_orders"] = totalOrders

	// 进行中订单
	var processingOrders int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE status NOT IN ('已完成','已退款','已取消','失败')").Scan(&processingOrders)
	stats["processing_orders"] = processingOrders

	// 已完成订单
	var completedOrders int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE status = '已完成'").Scan(&completedOrders)
	stats["completed_orders"] = completedOrders

	// 异常订单
	var failedOrders int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE status = '异常'").Scan(&failedOrders)
	stats["failed_orders"] = failedOrders

	// 总余额
	var totalBalance float64
	database.DB.QueryRow("SELECT COALESCE(SUM(money), 0) FROM qingka_wangke_user").Scan(&totalBalance)
	stats["total_balance"] = totalBalance

	// 近7天趋势
	trend := s.getWeekTrend()
	stats["trend"] = trend

	// 近期订单（最新10条）
	recentOrders := s.getRecentOrders()
	stats["recent_orders"] = recentOrders

	// 订单状态分布
	statusDist := s.getOrderStatusDistribution()
	stats["status_distribution"] = statusDist

	return stats, nil
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

// ===== 统计报表 =====

func (s *AdminService) StatsReport(days int) (map[string]interface{}, error) {
	if days <= 0 {
		days = 30
	}

	result := map[string]interface{}{}

	// 每日订单量和收入
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

	// 按课程分类统计
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

	// 按状态统计
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

	// 用户排行
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

// ===== 等级管理 =====

func (s *AdminService) GradeList() ([]model.Grade, error) {
	rows, err := database.DB.Query("SELECT id, COALESCE(sort,'0'), COALESCE(name,''), COALESCE(rate,'1'), COALESCE(money,'0'), COALESCE(addkf,'1'), COALESCE(gjkf,'1'), COALESCE(status,'1'), CASE WHEN time IS NOT NULL AND time != '' AND time != '0' THEN FROM_UNIXTIME(CAST(time AS UNSIGNED), '%Y-%m-%d %H:%i') ELSE '' END FROM qingka_wangke_dengji ORDER BY sort ASC, id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.Grade
	for rows.Next() {
		var g model.Grade
		rows.Scan(&g.ID, &g.Sort, &g.Name, &g.Rate, &g.Money, &g.AddKF, &g.GJKF, &g.Status, &g.Time)
		list = append(list, g)
	}
	if list == nil {
		list = []model.Grade{}
	}
	return list, nil
}

func (s *AdminService) GradeSave(req model.GradeSaveRequest) error {
	if req.Status == "" {
		req.Status = "1"
	}
	if req.Rate == "" {
		req.Rate = "1"
	}
	if req.Money == "" {
		req.Money = "0"
	}
	if req.AddKF == "" {
		req.AddKF = "1"
	}
	if req.GJKF == "" {
		req.GJKF = "1"
	}
	if req.ID > 0 {
		_, err := database.DB.Exec(
			"UPDATE qingka_wangke_dengji SET sort=?, name=?, rate=?, money=?, addkf=?, gjkf=?, status=? WHERE id=?",
			req.Sort, req.Name, req.Rate, req.Money, req.AddKF, req.GJKF, req.Status, req.ID,
		)
		return err
	}
	_, err := database.DB.Exec(
		"INSERT INTO qingka_wangke_dengji (sort, name, rate, money, addkf, gjkf, status, time) VALUES (?, ?, ?, ?, ?, ?, ?, UNIX_TIMESTAMP())",
		req.Sort, req.Name, req.Rate, req.Money, req.AddKF, req.GJKF, req.Status,
	)
	return err
}

func (s *AdminService) GradeDelete(id int) error {
	_, err := database.DB.Exec("DELETE FROM qingka_wangke_dengji WHERE id=?", id)
	return err
}

// ===== 密价设置 =====

func (s *AdminService) MiJiaList(req model.MiJiaListRequest) ([]model.MiJia, int64, []int, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}

	where := "1=1"
	args := []interface{}{}
	if req.UID > 0 {
		where += " AND m.uid = ?"
		args = append(args, req.UID)
	}
	if req.CID > 0 {
		where += " AND m.cid = ?"
		args = append(args, req.CID)
	}
	if req.Keyword != "" {
		where += " AND c.name LIKE ?"
		args = append(args, "%"+req.Keyword+"%")
	}

	var total int64
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_mijia m LEFT JOIN qingka_wangke_class c ON m.cid=c.cid WHERE "+where, countArgs...).Scan(&total)

	offset := (req.Page - 1) * req.Limit
	queryArgs := make([]interface{}, len(args))
	copy(queryArgs, args)
	queryArgs = append(queryArgs, req.Limit, offset)

	rows, err := database.DB.Query(
		"SELECT m.mid, m.uid, m.cid, COALESCE(m.mode,'0'), COALESCE(m.price,'0'), COALESCE(DATE_FORMAT(m.addtime,'%Y-%m-%d %H:%i:%s'),''), COALESCE(u.user,''), COALESCE(c.name,'') FROM qingka_wangke_mijia m LEFT JOIN qingka_wangke_user u ON m.uid=u.uid LEFT JOIN qingka_wangke_class c ON m.cid=c.cid WHERE "+where+" ORDER BY m.mid DESC LIMIT ? OFFSET ?",
		queryArgs...,
	)
	if err != nil {
		return nil, 0, nil, err
	}
	defer rows.Close()
	var list []model.MiJia
	for rows.Next() {
		var mj model.MiJia
		rows.Scan(&mj.MID, &mj.UID, &mj.CID, &mj.Mode, &mj.Price, &mj.AddTime, &mj.UserName, &mj.ClassName)
		list = append(list, mj)
	}
	if list == nil {
		list = []model.MiJia{}
	}

	// 获取所有已设密价的 UID 列表
	var uids []int
	uidRows, uidErr := database.DB.Query("SELECT DISTINCT uid FROM qingka_wangke_mijia ORDER BY uid")
	if uidErr == nil {
		defer uidRows.Close()
		for uidRows.Next() {
			var uid int
			uidRows.Scan(&uid)
			uids = append(uids, uid)
		}
	}
	if uids == nil {
		uids = []int{}
	}

	return list, total, uids, nil
}

func (s *AdminService) MiJiaSave(req model.MiJiaSaveRequest) error {
	if req.Mode == "" {
		req.Mode = "0"
	}
	if req.MID > 0 {
		_, err := database.DB.Exec(
			"UPDATE qingka_wangke_mijia SET uid=?, cid=?, mode=?, price=? WHERE mid=?",
			req.UID, req.CID, req.Mode, req.Price, req.MID,
		)
		return err
	}
	// 检查是否已存在
	var cnt int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_mijia WHERE uid=? AND cid=?", req.UID, req.CID).Scan(&cnt)
	if cnt > 0 {
		_, err := database.DB.Exec(
			"UPDATE qingka_wangke_mijia SET mode=?, price=? WHERE uid=? AND cid=?",
			req.Mode, req.Price, req.UID, req.CID,
		)
		return err
	}
	_, err := database.DB.Exec(
		"INSERT INTO qingka_wangke_mijia (uid, cid, mode, price, addtime) VALUES (?, ?, ?, ?, NOW())",
		req.UID, req.CID, req.Mode, req.Price,
	)
	return err
}

func (s *AdminService) MiJiaDelete(mids []int) error {
	if len(mids) == 0 {
		return errors.New("未指定要删除的ID")
	}
	placeholders := make([]string, len(mids))
	args := make([]interface{}, len(mids))
	for i, id := range mids {
		placeholders[i] = "?"
		args[i] = id
	}
	_, err := database.DB.Exec("DELETE FROM qingka_wangke_mijia WHERE mid IN ("+strings.Join(placeholders, ",")+")", args...)
	return err
}

func (s *AdminService) MiJiaBatch(req model.MiJiaBatchRequest) (int, error) {
	if req.Mode == "" {
		req.Mode = "0"
	}
	// 查询分类下的所有课程
	rows, err := database.DB.Query("SELECT cid, price FROM qingka_wangke_class WHERE fenlei = ?", req.Fenlei)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var cid int
		var origPrice string
		rows.Scan(&cid, &origPrice)

		finalPrice := req.Price
		finalMode := req.Mode

		// mode=4 倍率定价: 原价 × 倍率, 转为固定价(mode=2)存储
		if req.Mode == "4" {
			var op, pp float64
			fmt.Sscanf(origPrice, "%f", &op)
			fmt.Sscanf(req.Price, "%f", &pp)
			finalPrice = fmt.Sprintf("%.2f", op*pp)
			finalMode = "2"
		}

		var cnt int
		database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_mijia WHERE uid=? AND cid=?", req.UID, cid).Scan(&cnt)
		if cnt > 0 {
			database.DB.Exec("UPDATE qingka_wangke_mijia SET price=?, mode=? WHERE uid=? AND cid=?", finalPrice, finalMode, req.UID, cid)
		} else {
			database.DB.Exec("INSERT INTO qingka_wangke_mijia (uid, cid, mode, price, addtime) VALUES (?, ?, ?, ?, NOW())", req.UID, cid, finalMode, finalPrice)
		}
		count++
	}
	return count, nil
}

func requireAdmin(grade string) error {
	if grade != "2" && grade != "3" {
		return errors.New("需要管理员权限")
	}
	return nil
}

// ===== 货源排行 (对齐旧 ddtj.php) =====

type SupplierRankItem struct {
	HID            int    `json:"hid"`
	Name           string `json:"name"`
	TodayCount     int    `json:"today_count"`
	YesterdayCount int    `json:"yesterday_count"`
	TotalCount     int    `json:"total_count"`
}

func (s *AdminService) SupplierRanking() ([]SupplierRankItem, error) {
	// 获取所有启用的货源
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

	// 构建 IN 子句
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

	// 条件聚合一次性查询（对齐 PHP ddtj.php）
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

	// 按总销量降序排列
	result := make([]SupplierRankItem, 0, len(hidMap))
	for _, item := range hidMap {
		result = append(result, *item)
	}
	// 排序
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].TotalCount > result[i].TotalCount {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result, nil
}

// ===== 代理商品排行 (对齐旧 dl.php) =====

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
		// 本周一
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		monday := now.AddDate(0, 0, -(weekday - 1))
		startTime = monday.Format("2006-01-02") + " 00:00:00"
		endTime = now.Format("2006-01-02") + " 23:59:59"
	case "month":
		startTime = now.Format("2006-01") + "-01 00:00:00"
		// 本月最后一天
		lastDay := time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 0, now.Location())
		endTime = lastDay.Format("2006-01-02") + " 23:59:59"
	default: // today
		startTime = now.Format("2006-01-02") + " 00:00:00"
		endTime = now.Format("2006-01-02") + " 23:59:59"
	}

	// 对齐 PHP dl.php: GROUP BY ptname, ORDER BY count DESC, LIMIT 20
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
