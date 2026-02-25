package service

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"time"

	"go-api/internal/database"
	"go-api/internal/model"
)

// OpenAPIService 外部API服务（对应 PHP apisub.php 密钥调用部分）
// 受系统设置"查课配置"6项控制：
//
//	settings       - API调用总开关
//	api_proportion - API调用加价比例(%)
//	api_ck         - API查课最低余额
//	api_xd         - API下单最低余额
//	api_tongb      - API同步随机时间最小(分钟)
//	api_tongbc     - API同步随机时间最大(分钟)
type OpenAPIService struct{}

func NewOpenAPIService() *OpenAPIService {
	return &OpenAPIService{}
}

// getConf 读取系统配置
func (s *OpenAPIService) getConf() map[string]string {
	conf, _ := NewAdminService().GetConfig()
	return conf
}

// CheckEnabled 检查API调用是否开启
func (s *OpenAPIService) CheckEnabled() error {
	conf := s.getConf()
	if conf["settings"] == "0" {
		return errors.New("API调用功能已关闭")
	}
	return nil
}

// CheckQueryBalance 检查查课余额限制
func (s *OpenAPIService) CheckQueryBalance(money float64) error {
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

// CheckOrderBalance 检查下单余额限制
func (s *OpenAPIService) CheckOrderBalance(money float64) error {
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

// GetProportion 获取API加价比例（百分比），默认0
func (s *OpenAPIService) GetProportion() float64 {
	conf := s.getConf()
	if v := conf["api_proportion"]; v != "" {
		var p float64
		fmt.Sscanf(v, "%f", &p)
		return p
	}
	return 0
}

// GetSyncDelay 获取API同步随机延迟时间（分钟）
func (s *OpenAPIService) GetSyncDelay() int {
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

// QueryCourse API查课
func (s *OpenAPIService) QueryCourse(uid int, money float64, cid int, userinfo string) (*model.CourseQueryResponse, error) {
	if err := s.CheckEnabled(); err != nil {
		return nil, err
	}
	if err := s.CheckQueryBalance(money); err != nil {
		return nil, err
	}

	supService := NewSupplierService()
	return supService.QueryCourse(cid, userinfo)
}

// AddOrder API下单（对应 PHP sxadd）
func (s *OpenAPIService) AddOrder(uid int, money float64, req model.OrderAddRequest) (*model.OrderAddResult, error) {
	if err := s.CheckEnabled(); err != nil {
		return nil, err
	}
	if err := s.CheckOrderBalance(money); err != nil {
		return nil, err
	}

	// API加价: 在下单前临时调整用户加价系数
	proportion := s.GetProportion()
	if proportion > 0 {
		// 读取原始addprice，加上比例
		var addprice float64
		database.DB.QueryRow("SELECT COALESCE(addprice,1) FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice)
		newAddprice := math.Round(addprice*(1+proportion/100)*10000) / 10000
		// 临时更新（下单完成后恢复）
		database.DB.Exec("UPDATE qingka_wangke_user SET addprice = ? WHERE uid = ?", newAddprice, uid)
		defer database.DB.Exec("UPDATE qingka_wangke_user SET addprice = ? WHERE uid = ?", addprice, uid)
	}

	// 记录下单前时间，用于定位新插入的订单
	beforeTime := time.Now().Add(-1 * time.Second).Format("2006-01-02 15:04:05")

	result, err := NewOrderService().AddOrders(uid, req)
	if err != nil {
		return nil, err
	}

	// API同步延迟：设置随机延迟时间到本次新插入的订单
	delay := s.GetSyncDelay()
	if delay > 0 && result != nil && result.SuccessCount > 0 {
		syncTime := time.Now().Add(time.Duration(delay) * time.Minute).Format("2006-01-02 15:04:05")
		database.DB.Exec("UPDATE qingka_wangke_order SET tongbtime = ? WHERE uid = ? AND addtime >= ?", syncTime, uid, beforeTime)
	}

	return result, nil
}

// OrderList API订单列表
func (s *OpenAPIService) OrderList(uid int, page, limit int, status string) ([]map[string]interface{}, int, error) {
	if err := s.CheckEnabled(); err != nil {
		return nil, 0, err
	}

	where := "uid = ?"
	args := []interface{}{uid}

	if status != "" {
		where += " AND status = ?"
		args = append(args, status)
	}

	// count
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

// GetClassList API获取课程列表
func (s *OpenAPIService) GetClassList(uid int, money float64) ([]map[string]interface{}, error) {
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

	// 读用户加价系数
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
