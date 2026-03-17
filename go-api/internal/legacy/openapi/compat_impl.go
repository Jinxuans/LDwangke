package openapi

import (
	"fmt"
	"go-api/internal/database"
	"go-api/internal/model"
	classmodule "go-api/internal/modules/class"
	ordermodule "go-api/internal/modules/order"
	suppliermodule "go-api/internal/modules/supplier"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// OpenAPICompat PHP兼容路由
// 让下游PHP系统通过 /api.php?act=xxx 或 /api/index.php?act=xxx 无缝对接Go系统
// 参数和返回格式与PHP系统完全一致
func Compat(c *gin.Context) {
	act := c.Query("act")
	if act == "" {
		act = c.PostForm("act")
	}

	switch act {
	case "getmoney":
		compatGetMoney(c)
	case "class":
		compatClass(c)
	case "get", "chake":
		compatQuery(c)
	case "add", "sxadd", "getadd":
		compatAdd(c)
	case "chadan":
		compatChadan(c)
	case "cd":
		compatCd(c)
	case "budan", "bd":
		compatBudan(c)
	case "up":
		compatUp(c)
	case "bindpushuid":
		compatBindPushUID(c)
	case "bindpushemail":
		compatBindPushEmail(c)
	case "bindshowdocpush":
		compatBindShowDocPush(c)
	case "gaimi":
		compatGaimi(c)
	case "stop":
		compatStop(c)
	case "getclass":
		compatGetClass(c)
	case "getcate":
		compatGetCate(c)
	case "orders":
		compatOrders(c)
	case "cha_logwk":
		compatChaLogWk(c)
	default:
		c.JSON(200, gin.H{"code": -1, "msg": "未知的act参数: " + act})
	}
}

// compatAuth 兼容认证：读取uid+key，返回用户信息
func compatAuth(c *gin.Context) (uid int, money float64, addprice float64, ok bool) {
	uidStr := c.Query("uid")
	if uidStr == "" {
		uidStr = c.PostForm("uid")
	}
	key := c.Query("key")
	if key == "" {
		key = c.PostForm("key")
	}

	if uidStr == "" || key == "" {
		c.JSON(200, gin.H{"code": 0, "msg": "uid或者key为空"})
		return 0, 0, 0, false
	}

	var dbUID int
	var dbKey, dbUser string
	var dbMoney, dbAddPrice float64
	err := database.DB.QueryRow(
		"SELECT uid, COALESCE(`key`,''), COALESCE(user,''), COALESCE(money,0), COALESCE(addprice,1) FROM qingka_wangke_user WHERE uid = ?",
		uidStr,
	).Scan(&dbUID, &dbKey, &dbUser, &dbMoney, &dbAddPrice)
	if err != nil {
		c.JSON(200, gin.H{"code": -1, "msg": "用户不存在"})
		return 0, 0, 0, false
	}
	if dbKey == "" || dbKey == "0" {
		c.JSON(200, gin.H{"code": -1, "msg": "你还没有开通接口哦"})
		return 0, 0, 0, false
	}
	if dbKey != key {
		c.JSON(200, gin.H{"code": -2, "msg": "密匙错误"})
		return 0, 0, 0, false
	}

	return dbUID, dbMoney, dbAddPrice, true
}

// getParam 从 query 或 postform 获取参数
func getParam(c *gin.Context, key string) string {
	v := c.Query(key)
	if v == "" {
		v = c.PostForm(key)
	}
	return strings.TrimSpace(v)
}

// compatGetMoney act=getmoney 查询余额
func compatGetMoney(c *gin.Context) {
	uidStr := getParam(c, "uid")
	key := getParam(c, "key")
	if uidStr == "" || key == "" {
		c.JSON(200, gin.H{"code": 0, "msg": "uid或者key为空"})
		return
	}

	var dbKey, user, name string
	var money float64
	err := database.DB.QueryRow(
		"SELECT COALESCE(`key`,''), COALESCE(user,''), COALESCE(name,''), COALESCE(money,0) FROM qingka_wangke_user WHERE uid = ?",
		uidStr,
	).Scan(&dbKey, &user, &name, &money)
	if err != nil {
		c.JSON(200, gin.H{"code": -1, "msg": "用户不存在"})
		return
	}
	if dbKey == "" || dbKey == "0" {
		c.JSON(200, gin.H{"code": -1, "msg": "你还没有开通接口哦"})
		return
	}
	if dbKey != key {
		c.JSON(200, gin.H{"code": -2, "msg": "密匙错误"})
		return
	}

	c.JSON(200, gin.H{
		"code":  1,
		"msg":   "查询成功",
		"user":  user,
		"name":  name,
		"money": fmt.Sprintf("%.2f", money),
	})
}

// compatClass act=class 获取商品列表
func compatClass(c *gin.Context) {
	uid, _, addprice, ok := compatAuth(c)
	if !ok {
		return
	}

	rows, err := database.DB.Query("SELECT cid, COALESCE(sort,0), COALESCE(name,''), COALESCE(content,''), COALESCE(status,0), COALESCE(price,'0'), COALESCE(yunsuan,'*') FROM qingka_wangke_class WHERE status = 1 ORDER BY sort ASC, cid DESC")
	if err != nil {
		c.JSON(200, gin.H{"code": -1, "msg": "查询失败"})
		return
	}
	defer rows.Close()

	type compatClassItem struct {
		cid       int
		sort      int
		status    int
		name      string
		content   string
		yunsuan   string
		basePrice float64
	}

	var items []compatClassItem
	var cids []int
	for rows.Next() {
		var item compatClassItem
		var priceStr string
		rows.Scan(&item.cid, &item.sort, &item.name, &item.content, &item.status, &priceStr, &item.yunsuan)
		item.basePrice, _ = strconv.ParseFloat(priceStr, 64)
		items = append(items, item)
		cids = append(cids, item.cid)
	}

	mijiaMap := map[int]classmodule.MiJiaRule{}
	if loaded, err := classmodule.LoadMiJiaMap(uid, cids); err == nil {
		mijiaMap = loaded
	}

	var data []gin.H
	for _, item := range items {
		userPrice := classmodule.ComputeClassBasePrice(item.basePrice, addprice, item.yunsuan, 4)
		if mj, ok := mijiaMap[item.cid]; ok {
			userPrice, _, _ = classmodule.ApplyMiJia(item.basePrice, addprice, item.yunsuan, mj.Mode, mj.Price, 4)
		}

		data = append(data, gin.H{
			"sort":    item.sort,
			"cid":     item.cid,
			"name":    item.name,
			"content": item.content,
			"status":  item.status,
			"price":   fmt.Sprintf("%.2f", item.basePrice),
			"price5":  fmt.Sprintf("%.2f", item.basePrice+0.5),
			"jiage":   fmt.Sprintf("%.2f", userPrice),
		})
	}
	if data == nil {
		data = []gin.H{}
	}

	c.JSON(200, gin.H{"code": 1, "data": data})
}

// compatQuery act=get/chake 查课
func compatQuery(c *gin.Context) {
	uid, money, _, ok := compatAuth(c)
	if !ok {
		return
	}

	platform := getParam(c, "platform")
	if platform == "" {
		platform = getParam(c, "cid")
	}
	school := getParam(c, "school")
	user := getParam(c, "user")
	pass := getParam(c, "pass")

	if platform == "" || school == "" || user == "" || pass == "" {
		c.JSON(200, gin.H{"code": 0, "msg": "所有项目不能为空"})
		return
	}

	cid, _ := strconv.Atoi(platform)
	userinfo := school + " " + user + " " + pass

	if err := checkOpenAPIQueryBalance(money); err != nil {
		c.JSON(200, gin.H{"code": -2, "msg": err.Error()})
		return
	}

	result, err := openAPIQueryCourse(uid, money, cid, userinfo)
	if err != nil {
		c.JSON(200, gin.H{"code": -1, "msg": err.Error()})
		return
	}

	// 返回与PHP一致的格式
	c.JSON(200, result)
}

// compatAdd act=add/sxadd/getadd 下单
func compatAdd(c *gin.Context) {
	uid, money, _, ok := compatAuth(c)
	if !ok {
		return
	}

	platform := getParam(c, "platform")
	if platform == "" {
		platform = getParam(c, "cid")
	}
	school := getParam(c, "school")
	user := getParam(c, "user")
	pass := getParam(c, "pass")
	kcname := getParam(c, "kcname")
	kcid := getParam(c, "kcid")

	if platform == "" || school == "" || user == "" || pass == "" || kcname == "" {
		c.JSON(200, gin.H{"code": 0, "msg": "所有项目不能为空"})
		return
	}

	cid, _ := strconv.Atoi(platform)

	if err := checkOpenAPIOrderBalance(money); err != nil {
		c.JSON(200, gin.H{"code": -2, "msg": err.Error()})
		return
	}

	userinfo := school + " " + user + " " + pass
	if kcid != "" {
		userinfo += " " + kcid
	}
	userinfo += " " + kcname

	req := model.OrderAddRequest{
		CID: cid,
		Data: []model.OrderAddItem{
			{UserInfo: userinfo},
		},
	}

	result, err := openAPIAddOrder(uid, money, req)
	if err != nil {
		c.JSON(200, gin.H{"code": -1, "msg": err.Error(), "status": -1, "message": err.Error()})
		return
	}

	// PHP格式返回
	if result != nil && result.SuccessCount > 0 {
		oid := ""
		if len(result.OIDs) > 0 {
			oid = fmt.Sprintf("%d", result.OIDs[0])
		}
		c.JSON(200, gin.H{"code": 0, "msg": "提交成功", "status": 0, "message": "提交成功", "id": oid})
	} else {
		msg := "下单失败"
		if result != nil && len(result.SkippedItems) > 0 {
			msg = result.SkippedItems[0]
		}
		c.JSON(200, gin.H{"code": -1, "msg": msg, "status": -1, "message": msg})
	}
}

// compatChadan act=chadan 查单
func compatChadan(c *gin.Context) {
	username := getParam(c, "username")
	oid := getParam(c, "oid")
	if oid == "" {
		oid = getParam(c, "id")
	}

	if username == "" && oid == "" {
		c.JSON(200, gin.H{"code": -1, "msg": "账号或订单ID不能为空"})
		return
	}

	list, err := openAPIChadan(username, oid)
	if err != nil {
		c.JSON(200, gin.H{"code": -1, "msg": err.Error()})
		return
	}
	if len(list) == 0 {
		c.JSON(200, gin.H{"code": -1, "msg": "未查到该账号的下单信息"})
		return
	}

	c.JSON(200, gin.H{"code": 1, "data": list})
}

// compatCd act=cd 按账号查单（无需uid/key认证）
func compatCd(c *gin.Context) {
	username := getParam(c, "username")
	if username == "" {
		c.JSON(200, gin.H{"code": -1, "msg": "账号不能为空"})
		return
	}

	rows, err := database.DB.Query(`SELECT oid, COALESCE(ptname,''), COALESCE(school,''), COALESCE(name,''), 
		COALESCE(user,''), COALESCE(kcname,''), COALESCE(addtime,''), 
		COALESCE(courseStartTime,''), COALESCE(courseEndTime,''), 
		COALESCE(examStartTime,''), COALESCE(examEndTime,''), 
		COALESCE(status,''), COALESCE(process,''), COALESCE(remarks,'')
		FROM qingka_wangke_order WHERE user=? ORDER BY oid DESC`, username)
	if err != nil {
		c.JSON(200, gin.H{"code": -1, "msg": "查询失败"})
		return
	}
	defer rows.Close()

	var data []gin.H
	for rows.Next() {
		var oid int
		var ptname, school, name, user, kcname, addtime string
		var courseStartTime, courseEndTime, examStartTime, examEndTime string
		var status, process, remarks string
		rows.Scan(&oid, &ptname, &school, &name, &user, &kcname, &addtime,
			&courseStartTime, &courseEndTime, &examStartTime, &examEndTime,
			&status, &process, &remarks)
		data = append(data, gin.H{
			"id":              oid,
			"ptname":          ptname,
			"school":          school,
			"name":            name,
			"user":            user,
			"kcname":          kcname,
			"addtime":         addtime,
			"courseStartTime": courseStartTime,
			"courseEndTime":   courseEndTime,
			"examStartTime":   examStartTime,
			"examEndTime":     examEndTime,
			"status":          status,
			"process":         process,
			"remarks":         remarks,
		})
	}
	if data == nil {
		c.JSON(200, gin.H{"code": -1, "msg": "未查到该账号的下单信息"})
		return
	}
	c.JSON(200, gin.H{"code": 1, "data": data})
}

// compatBudan act=budan/bd 补刷
func compatBudan(c *gin.Context) {
	oidStr := getParam(c, "id")
	if oidStr == "" {
		oidStr = getParam(c, "oid")
	}
	oid, _ := strconv.Atoi(oidStr)
	if oid <= 0 {
		c.JSON(200, gin.H{"code": -1, "msg": "订单ID不能为空"})
		return
	}

	// 查询订单信息
	var hid, bsnum int
	var yid, status string
	err := database.DB.QueryRow(
		"SELECT hid, COALESCE(yid,''), COALESCE(status,''), COALESCE(bsnum,0) FROM qingka_wangke_order WHERE oid=?", oid,
	).Scan(&hid, &yid, &status, &bsnum)
	if err != nil {
		c.JSON(200, gin.H{"code": -1, "msg": "订单不存在"})
		return
	}
	if bsnum > 20 {
		c.JSON(200, gin.H{"code": -1, "msg": "该订单补刷已超过20次"})
		return
	}
	if yid == "" || yid == "0" {
		c.JSON(200, gin.H{"code": -1, "msg": "订单未对接，无法补单"})
		return
	}

	supSvc := suppliermodule.SharedService()
	sup, err := supSvc.GetSupplierByHID(hid)
	if err != nil {
		c.JSON(200, gin.H{"code": -1, "msg": "未找到货源信息"})
		return
	}

	code, msg, err := supSvc.ResubmitOrder(sup, yid)
	if err != nil {
		c.JSON(200, gin.H{"code": -1, "msg": err.Error()})
		return
	}
	if code == 1 || code == 0 {
		database.DB.Exec("UPDATE qingka_wangke_order SET status='补刷中', dockstatus=1, bsnum=bsnum+1 WHERE oid=?", oid)
		c.JSON(200, gin.H{"code": 1, "msg": msg})
	} else {
		c.JSON(200, gin.H{"code": -1, "msg": msg})
	}
}

// compatUp act=up 同步进度
func compatUp(c *gin.Context) {
	oidStr := getParam(c, "id")
	if oidStr == "" {
		oidStr = getParam(c, "oid")
	}
	oid, _ := strconv.Atoi(oidStr)
	if oid <= 0 {
		c.JSON(200, gin.H{"code": -1, "msg": "订单ID不能为空"})
		return
	}

	_, err := ordermodule.NewServices().Sync.SyncProgress([]int{oid})
	if err != nil {
		c.JSON(200, gin.H{"code": -1, "msg": err.Error()})
		return
	}

	c.JSON(200, gin.H{"code": 1, "msg": "同步成功，请重新查询信息"})
}

// compatBindPushUID act=bindpushuid 绑定微信推送
func compatBindPushUID(c *gin.Context) {
	orderIDStr := getParam(c, "orderid")
	pushUID := getParam(c, "pushuid")
	orderID, _ := strconv.Atoi(orderIDStr)
	if orderID <= 0 {
		c.JSON(200, gin.H{"code": 0, "msg": "参数不全"})
		return
	}
	if err := openAPIBindPushUID(orderID, pushUID); err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 1, "msg": "操作成功"})
}

// compatBindPushEmail act=bindpushemail 绑定邮箱推送
func compatBindPushEmail(c *gin.Context) {
	orderIDStr := getParam(c, "orderid")
	account := getParam(c, "account")
	pushEmail := getParam(c, "pushEmail")
	orderID, _ := strconv.Atoi(orderIDStr)
	if orderID <= 0 && account == "" {
		c.JSON(200, gin.H{"code": 0, "msg": "参数不全"})
		return
	}
	if err := openAPIBindPushEmail(orderID, account, pushEmail); err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 1, "msg": "操作成功"})
}

// compatBindShowDocPush act=bindshowdocpush 绑定ShowDoc推送
func compatBindShowDocPush(c *gin.Context) {
	orderIDStr := getParam(c, "orderid")
	account := getParam(c, "account")
	showdocURL := getParam(c, "showdoc_url")
	orderID, _ := strconv.Atoi(orderIDStr)
	if orderID <= 0 && account == "" {
		c.JSON(200, gin.H{"code": 0, "msg": "参数不全"})
		return
	}
	if err := openAPIBindShowDocPush(orderID, account, showdocURL); err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 1, "msg": "操作成功"})
}

// compatGaimi act=gaimi 改密
func compatGaimi(c *gin.Context) {
	_, _, _, ok := compatAuth(c)
	if !ok {
		return
	}

	oidStr := getParam(c, "id")
	if oidStr == "" {
		oidStr = getParam(c, "oid")
	}
	newPwd := getParam(c, "newPwd")
	if newPwd == "" {
		newPwd = getParam(c, "newpwd")
	}
	oid, _ := strconv.Atoi(oidStr)
	if oid <= 0 || newPwd == "" {
		c.JSON(200, gin.H{"code": 0, "msg": "参数不全"})
		return
	}

	var hid int
	var yid, status string
	err := database.DB.QueryRow(
		"SELECT hid, COALESCE(yid,''), COALESCE(status,'') FROM qingka_wangke_order WHERE oid=?", oid,
	).Scan(&hid, &yid, &status)
	if err != nil {
		c.JSON(200, gin.H{"code": -1, "msg": "订单不存在"})
		return
	}
	if yid == "" || yid == "0" {
		c.JSON(200, gin.H{"code": -1, "msg": "订单未对接，无法改密"})
		return
	}

	supSvc := suppliermodule.SharedService()
	sup, err := supSvc.GetSupplierByHID(hid)
	if err != nil {
		c.JSON(200, gin.H{"code": -1, "msg": "未找到货源信息"})
		return
	}

	code, msg, err := supSvc.ChangePassword(sup, yid, newPwd)
	if err != nil {
		c.JSON(200, gin.H{"code": -1, "msg": err.Error()})
		return
	}
	if code == 1 || code == 0 {
		database.DB.Exec("UPDATE qingka_wangke_order SET pass=? WHERE oid=?", newPwd, oid)
		c.JSON(200, gin.H{"code": 1, "msg": msg})
	} else {
		c.JSON(200, gin.H{"code": -1, "msg": msg})
	}
}

// compatStop act=stop 暂停订单
func compatStop(c *gin.Context) {
	_, _, _, ok := compatAuth(c)
	if !ok {
		return
	}

	oidStr := getParam(c, "id")
	if oidStr == "" {
		oidStr = getParam(c, "oid")
	}
	oid, _ := strconv.Atoi(oidStr)
	if oid <= 0 {
		c.JSON(200, gin.H{"code": 0, "msg": "参数不全"})
		return
	}

	var hid int
	var yid string
	err := database.DB.QueryRow(
		"SELECT hid, COALESCE(yid,'') FROM qingka_wangke_order WHERE oid=?", oid,
	).Scan(&hid, &yid)
	if err != nil {
		c.JSON(200, gin.H{"code": -1, "msg": "订单不存在"})
		return
	}
	if yid == "" || yid == "0" {
		c.JSON(200, gin.H{"code": -1, "msg": "订单未对接，无法暂停"})
		return
	}

	supSvc := suppliermodule.SharedService()
	sup, err := supSvc.GetSupplierByHID(hid)
	if err != nil {
		c.JSON(200, gin.H{"code": -1, "msg": "未找到货源信息"})
		return
	}

	code, msg, err := supSvc.PauseOrder(sup, yid)
	if err != nil {
		c.JSON(200, gin.H{"code": -1, "msg": err.Error()})
		return
	}
	if code == 1 || code == 0 {
		c.JSON(200, gin.H{"code": 1, "msg": msg})
	} else {
		c.JSON(200, gin.H{"code": -1, "msg": msg})
	}
}

// compatGetClass act=getclass 获取课程列表（含价格）
func compatGetClass(c *gin.Context) {
	uid, _, addprice, ok := compatAuth(c)
	if !ok {
		return
	}

	fenlei := getParam(c, "fenlei")

	query := "SELECT cid, COALESCE(name,''), COALESCE(price,'0'), COALESCE(yunsuan,'*'), COALESCE(fenlei,''), COALESCE(status,0) FROM qingka_wangke_class WHERE status = 1"
	var args []interface{}
	if fenlei != "" {
		query += " AND fenlei = ?"
		args = append(args, fenlei)
	}
	query += " ORDER BY sort ASC, cid DESC"

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(200, gin.H{"code": -1, "msg": "查询失败"})
		return
	}
	defer rows.Close()

	type compatGetClassItem struct {
		cid       int
		status    int
		name      string
		yunsuan   string
		fenlei    string
		basePrice float64
	}

	var items []compatGetClassItem
	var cids []int
	for rows.Next() {
		var item compatGetClassItem
		var priceStr string
		rows.Scan(&item.cid, &item.name, &priceStr, &item.yunsuan, &item.fenlei, &item.status)
		item.basePrice, _ = strconv.ParseFloat(priceStr, 64)
		items = append(items, item)
		cids = append(cids, item.cid)
	}

	mijiaMap := map[int]classmodule.MiJiaRule{}
	if loaded, err := classmodule.LoadMiJiaMap(uid, cids); err == nil {
		mijiaMap = loaded
	}

	var data []gin.H
	for _, item := range items {
		userPrice := classmodule.ComputeClassBasePrice(item.basePrice, addprice, item.yunsuan, 4)
		if mj, ok := mijiaMap[item.cid]; ok {
			userPrice, _, _ = classmodule.ApplyMiJia(item.basePrice, addprice, item.yunsuan, mj.Mode, mj.Price, 4)
		}

		data = append(data, gin.H{
			"cid":    item.cid,
			"name":   item.name,
			"price":  fmt.Sprintf("%.2f", userPrice),
			"fenlei": item.fenlei,
		})
	}
	if data == nil {
		data = []gin.H{}
	}
	c.JSON(200, gin.H{"code": 1, "data": data})
}

// compatGetCate act=getcate 获取分类列表
func compatGetCate(c *gin.Context) {
	_, _, _, ok := compatAuth(c)
	if !ok {
		return
	}

	rows, err := database.DB.Query("SELECT id, COALESCE(name,'') FROM qingka_wangke_fenlei ORDER BY sort ASC, id ASC")
	if err != nil {
		c.JSON(200, gin.H{"code": -1, "msg": "查询失败"})
		return
	}
	defer rows.Close()

	var data []gin.H
	for rows.Next() {
		var id int
		var name string
		rows.Scan(&id, &name)
		data = append(data, gin.H{"id": id, "name": name})
	}
	if data == nil {
		data = []gin.H{}
	}
	c.JSON(200, gin.H{"code": 1, "data": data})
}

// compatOrders act=orders 获取订单列表
func compatOrders(c *gin.Context) {
	uid, _, _, ok := compatAuth(c)
	if !ok {
		return
	}

	pageStr := getParam(c, "page")
	limitStr := getParam(c, "limit")
	if pageStr == "" {
		pageStr = "1"
	}
	if limitStr == "" {
		limitStr = "100"
	}
	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 500 {
		limit = 100
	}
	offset := (page - 1) * limit

	var total int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE uid=?", uid).Scan(&total)

	rows, err := database.DB.Query(
		fmt.Sprintf(`SELECT oid, COALESCE(cid,0), COALESCE(ptname,''), COALESCE(school,''), COALESCE(user,''), COALESCE(pass,''),
			COALESCE(kcid,''), COALESCE(kcname,''), COALESCE(fees,0), COALESCE(status,''), COALESCE(process,''),
			COALESCE(remarks,''), COALESCE(addtime,'')
			FROM qingka_wangke_order WHERE uid=? ORDER BY oid DESC LIMIT %d OFFSET %d`, limit, offset), uid)
	if err != nil {
		c.JSON(200, gin.H{"code": -1, "msg": "查询失败"})
		return
	}
	defer rows.Close()

	var data []gin.H
	for rows.Next() {
		var oid, cid int
		var ptname, school, user, pass, kcid, kcname, status, process, remarks, addtime string
		var fees float64
		rows.Scan(&oid, &cid, &ptname, &school, &user, &pass, &kcid, &kcname, &fees, &status, &process, &remarks, &addtime)
		data = append(data, gin.H{
			"oid":     oid,
			"cid":     cid,
			"ptname":  ptname,
			"school":  school,
			"user":    user,
			"pass":    pass,
			"kcid":    kcid,
			"kcname":  kcname,
			"fees":    fees,
			"status":  status,
			"process": process,
			"remarks": remarks,
			"addtime": addtime,
		})
	}
	if data == nil {
		data = []gin.H{}
	}
	c.JSON(200, gin.H{"code": 1, "data": data, "total": total, "page": page, "limit": limit})
}

// compatChaLogWk act=cha_logwk 查询订单日志
func compatChaLogWk(c *gin.Context) {
	_, _, _, ok := compatAuth(c)
	if !ok {
		return
	}

	oidStr := getParam(c, "id")
	if oidStr == "" {
		oidStr = getParam(c, "oid")
	}
	oid, _ := strconv.Atoi(oidStr)
	if oid <= 0 {
		c.JSON(200, gin.H{"code": 0, "msg": "参数不全"})
		return
	}

	// 查询订单信息用于向上游获取日志
	var hid int
	var yid string
	err := database.DB.QueryRow(
		"SELECT hid, COALESCE(yid,'') FROM qingka_wangke_order WHERE oid=?", oid,
	).Scan(&hid, &yid)
	if err != nil {
		c.JSON(200, gin.H{"code": -1, "msg": "订单不存在"})
		return
	}
	if yid == "" || yid == "0" {
		c.JSON(200, gin.H{"code": -1, "msg": "订单未对接，无法查询日志"})
		return
	}

	supSvc := suppliermodule.SharedService()
	sup, err := supSvc.GetSupplierByHID(hid)
	if err != nil {
		c.JSON(200, gin.H{"code": -1, "msg": "未找到货源信息"})
		return
	}

	logs, err := supSvc.QueryOrderLogs(sup, yid)
	if err != nil {
		c.JSON(200, gin.H{"code": -1, "msg": err.Error()})
		return
	}

	c.JSON(200, gin.H{"code": 1, "data": logs})
}
