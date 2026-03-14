package openapi

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"time"

	"go-api/internal/database"
	"go-api/internal/model"
	commonmodule "go-api/internal/modules/common"
	ordermodule "go-api/internal/modules/order"
	suppliermodule "go-api/internal/modules/supplier"
)

type openAPIService struct{}

var legacyOpenAPI = &openAPIService{}

func checkOpenAPIQueryBalance(money float64) error {
	return legacyOpenAPI.CheckQueryBalance(money)
}

func checkOpenAPIOrderBalance(money float64) error {
	return legacyOpenAPI.CheckOrderBalance(money)
}

func openAPIQueryCourse(uid int, money float64, cid int, userinfo string) (*model.CourseQueryResponse, error) {
	return legacyOpenAPI.QueryCourse(uid, money, cid, userinfo)
}

func openAPIAddOrder(uid int, money float64, req model.OrderAddRequest) (*model.OrderAddResult, error) {
	return legacyOpenAPI.AddOrder(uid, money, req)
}

func openAPIOrderList(uid int, page, limit int, status string) ([]map[string]interface{}, int, error) {
	return legacyOpenAPI.OrderList(uid, page, limit, status)
}

func openAPIGetClassList(uid int, money float64) ([]map[string]interface{}, error) {
	return legacyOpenAPI.GetClassList(uid, money)
}

func openAPIChadan(username string, orderID string) ([]map[string]interface{}, error) {
	return legacyOpenAPI.Chadan(username, orderID)
}

func openAPIBindPushUID(orderID int, pushUID string) error {
	return bindPushUID(orderID, pushUID)
}

func openAPIBindPushEmail(orderID int, account string, pushEmail string) error {
	_, err := bindPushEmail(orderID, account, pushEmail)
	return err
}

func openAPIBindShowDocPush(orderID int, account string, showdocURL string) error {
	_, err := bindShowDocPush(orderID, account, showdocURL)
	return err
}

func (s *openAPIService) getConf() map[string]string {
	conf, _ := commonmodule.GetAdminConfigMap()
	return conf
}

func (s *openAPIService) CheckEnabled() error {
	conf := s.getConf()
	if conf["settings"] == "0" {
		return errors.New("API调用功能已关闭")
	}
	return nil
}

func (s *openAPIService) CheckQueryBalance(money float64) error {
	conf := s.getConf()
	if apiCk := conf["api_ck"]; apiCk != "" {
		var minBalance float64
		fmt.Sscanf(apiCk, "%f", &minBalance)
		if minBalance > 0 && money < minBalance {
			return fmt.Errorf("API查课余额不足，最低需要%.0f元", minBalance)
		}
	}
	return nil
}

func (s *openAPIService) CheckOrderBalance(money float64) error {
	conf := s.getConf()
	if apiXd := conf["api_xd"]; apiXd != "" {
		var minBalance float64
		fmt.Sscanf(apiXd, "%f", &minBalance)
		if minBalance > 0 && money < minBalance {
			return fmt.Errorf("API下单余额不足，最低需要%.0f元", minBalance)
		}
	}
	return nil
}

func (s *openAPIService) GetProportion() float64 {
	conf := s.getConf()
	if v := conf["api_proportion"]; v != "" {
		var p float64
		fmt.Sscanf(v, "%f", &p)
		return p
	}
	return 0
}

func (s *openAPIService) GetSyncDelay() int {
	conf := s.getConf()
	minStr := conf["api_tongb"]
	maxStr := conf["api_tongbc"]
	var minVal, maxVal int
	fmt.Sscanf(minStr, "%d", &minVal)
	fmt.Sscanf(maxStr, "%d", &maxVal)
	if maxVal <= minVal {
		return minVal
	}
	return minVal + rand.Intn(maxVal-minVal+1)
}

func (s *openAPIService) QueryCourse(uid int, money float64, cid int, userinfo string) (*model.CourseQueryResponse, error) {
	if err := s.CheckEnabled(); err != nil {
		return nil, err
	}
	if err := s.CheckQueryBalance(money); err != nil {
		return nil, err
	}
	return suppliermodule.SharedService().QueryCourse(cid, userinfo)
}

func (s *openAPIService) AddOrder(uid int, money float64, req model.OrderAddRequest) (*model.OrderAddResult, error) {
	if err := s.CheckEnabled(); err != nil {
		return nil, err
	}
	if err := s.CheckOrderBalance(money); err != nil {
		return nil, err
	}

	proportion := s.GetProportion()
	if proportion > 0 {
		var addprice float64
		database.DB.QueryRow("SELECT COALESCE(addprice,1) FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice)
		newAddprice := math.Round(addprice*(1+proportion/100)*10000) / 10000
		database.DB.Exec("UPDATE qingka_wangke_user SET addprice = ? WHERE uid = ?", newAddprice, uid)
		defer database.DB.Exec("UPDATE qingka_wangke_user SET addprice = ? WHERE uid = ?", addprice, uid)
	}

	beforeTime := time.Now().Add(-1 * time.Second).Format("2006-01-02 15:04:05")
	result, err := ordermodule.NewServices().Command.Add(uid, req)
	if err != nil {
		return nil, err
	}

	delay := s.GetSyncDelay()
	if delay > 0 && result != nil && result.SuccessCount > 0 {
		syncTime := time.Now().Add(time.Duration(delay) * time.Minute).Format("2006-01-02 15:04:05")
		database.DB.Exec("UPDATE qingka_wangke_order SET tongbtime = ? WHERE uid = ? AND addtime >= ?", syncTime, uid, beforeTime)
	}

	return result, nil
}

func (s *openAPIService) OrderList(uid int, page, limit int, status string) ([]map[string]interface{}, int, error) {
	if err := s.CheckEnabled(); err != nil {
		return nil, 0, err
	}

	where := "uid = ?"
	args := []interface{}{uid}
	if status != "" {
		where += " AND status = ?"
		args = append(args, status)
	}

	var total int
	countSQL := "SELECT COUNT(*) FROM qingka_wangke_order WHERE " + where
	database.DB.QueryRow(countSQL, args...).Scan(&total)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	querySQL := fmt.Sprintf("SELECT oid, COALESCE(cid,0), COALESCE(kcname,''), COALESCE(user,''), COALESCE(status,''), COALESCE(fees,0), COALESCE(addtime,''), COALESCE(yid,''), COALESCE(progress,'') FROM qingka_wangke_order WHERE %s ORDER BY oid DESC LIMIT %d OFFSET %d", where, limit, offset)
	rows, err := database.DB.Query(querySQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []map[string]interface{}
	for rows.Next() {
		var oid, cid int
		var kcname, user, status, addtime, yid, progress string
		var fees float64
		rows.Scan(&oid, &cid, &kcname, &user, &status, &fees, &addtime, &yid, &progress)
		list = append(list, map[string]interface{}{
			"oid":      oid,
			"cid":      cid,
			"kcname":   kcname,
			"user":     user,
			"status":   status,
			"fees":     fees,
			"addtime":  addtime,
			"yid":      yid,
			"progress": progress,
		})
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	return list, total, nil
}

func (s *openAPIService) GetClassList(uid int, money float64) ([]map[string]interface{}, error) {
	if err := s.CheckEnabled(); err != nil {
		return nil, err
	}
	if err := s.CheckQueryBalance(money); err != nil {
		return nil, err
	}

	rows, err := database.DB.Query("SELECT cid, COALESCE(name,''), COALESCE(price,'0'), COALESCE(fenlei,''), COALESCE(status,0) FROM qingka_wangke_class WHERE status = 1 ORDER BY cid")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var addprice float64
	database.DB.QueryRow("SELECT COALESCE(addprice,1) FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice)
	proportion := s.GetProportion()

	var list []map[string]interface{}
	for rows.Next() {
		var cid, status int
		var name, priceStr, fenlei string
		rows.Scan(&cid, &name, &priceStr, &fenlei, &status)

		basePrice, _ := strconv.ParseFloat(priceStr, 64)
		userPrice := math.Round(basePrice*addprice*100) / 100
		if proportion > 0 {
			userPrice = math.Round(userPrice*(1+proportion/100)*100) / 100
		}

		list = append(list, map[string]interface{}{
			"cid":    cid,
			"name":   name,
			"price":  userPrice,
			"fenlei": fenlei,
		})
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	return list, nil
}

func (s *openAPIService) Chadan(username string, orderID string) ([]map[string]interface{}, error) {
	if username == "" && orderID == "" {
		return nil, errors.New("账号或订单ID不能为空")
	}

	var rows *sql.Rows
	var err error
	if username != "" {
		rows, err = database.DB.Query(`SELECT oid, COALESCE(ptname,''), COALESCE(school,''), COALESCE(user,''), COALESCE(pass,''),
			COALESCE(kcname,''), COALESCE(addtime,''), COALESCE(status,''), COALESCE(process,''), COALESCE(remarks,''),
			COALESCE(pushUid,''), COALESCE(pushStatus,''), COALESCE(pushEmail,''),
			COALESCE(pushEmailStatus,'0'), COALESCE(showdoc_push_url,''), COALESCE(pushShowdocStatus,'0')
			FROM qingka_wangke_order WHERE user=? ORDER BY oid ASC`, username)
	} else {
		rows, err = database.DB.Query(`SELECT oid, COALESCE(ptname,''), COALESCE(school,''), COALESCE(user,''), COALESCE(pass,''),
			COALESCE(kcname,''), COALESCE(addtime,''), COALESCE(status,''), COALESCE(process,''), COALESCE(remarks,''),
			COALESCE(pushUid,''), COALESCE(pushStatus,''), COALESCE(pushEmail,''),
			COALESCE(pushEmailStatus,'0'), COALESCE(showdoc_push_url,''), COALESCE(pushShowdocStatus,'0')
			FROM qingka_wangke_order WHERE oid=?`, orderID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []map[string]interface{}
	for rows.Next() {
		var oid int
		var ptname, school, user, pass, kcname, addtime, status, process, remarks string
		var pushUID, pushStatus, pushEmail, pushEmailStatus, showdocPushURL, pushShowdocStatus string
		rows.Scan(&oid, &ptname, &school, &user, &pass, &kcname, &addtime, &status, &process, &remarks,
			&pushUID, &pushStatus, &pushEmail, &pushEmailStatus, &showdocPushURL, &pushShowdocStatus)

		schoolVal := school
		if schoolVal == "" {
			schoolVal = "自动识别"
		}
		account := schoolVal + " " + user + " " + pass

		list = append(list, map[string]interface{}{
			"id":                oid,
			"ptname":            ptname,
			"school":            school,
			"user":              user,
			"pass":              pass,
			"account":           account,
			"kcname":            kcname,
			"addtime":           addtime,
			"status":            status,
			"process":           process,
			"remarks":           remarks,
			"pushUid":           pushUID,
			"pushStatus":        pushStatus,
			"pushEmail":         pushEmail,
			"pushEmailStatus":   pushEmailStatus,
			"showdoc_push_url":  showdocPushURL,
			"pushShowdocStatus": pushShowdocStatus,
		})
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	return list, nil
}

func bindPushUID(orderID int, pushUID string) error {
	_, err := database.DB.Exec(
		"UPDATE qingka_wangke_order SET pushUid=?, pushStatus='0' WHERE oid=?",
		pushUID, orderID,
	)
	return err
}

func bindPushEmail(orderID int, account string, email string) (int64, error) {
	if orderID > 0 {
		result, err := database.DB.Exec(
			"UPDATE qingka_wangke_order SET pushEmail=?, pushEmailStatus='0' WHERE oid=?",
			email, orderID,
		)
		if err != nil {
			return 0, err
		}
		return result.RowsAffected()
	}
	if account != "" {
		result, err := database.DB.Exec(
			"UPDATE qingka_wangke_order SET pushEmail=?, pushEmailStatus='0' WHERE user=?",
			email, account,
		)
		if err != nil {
			return 0, err
		}
		return result.RowsAffected()
	}
	return 0, fmt.Errorf("需要订单ID或账号")
}

func bindShowDocPush(orderID int, account string, showdocURL string) (int64, error) {
	if orderID > 0 {
		result, err := database.DB.Exec(
			"UPDATE qingka_wangke_order SET showdoc_push_url=?, pushShowdocStatus='0' WHERE oid=?",
			showdocURL, orderID,
		)
		if err != nil {
			return 0, err
		}
		return result.RowsAffected()
	}
	if account != "" {
		result, err := database.DB.Exec(
			"UPDATE qingka_wangke_order SET showdoc_push_url=?, pushShowdocStatus='0' WHERE user=?",
			showdocURL, account,
		)
		if err != nil {
			return 0, err
		}
		return result.RowsAffected()
	}
	return 0, fmt.Errorf("需要订单ID或账号")
}
