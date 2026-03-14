package service

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"strconv"

	"go-api/internal/database"
)

func (s *PaperService) GenerateTitles(params map[string]interface{}) (map[string]interface{}, error) {
	return s.apiPost("prod-api/system/lunwen/generate-titles", params)
}

func (s *PaperService) GenerateOutline(params map[string]interface{}) (map[string]interface{}, error) {
	return s.apiPost("prod-api/system/lunwen/generate-outline", params)
}

func (s *PaperService) OutlineStatus(orderID string) (map[string]interface{}, error) {
	return s.apiGet("prod-api/system/lunwen/outline-status/"+orderID, nil)
}

func (s *PaperService) GetTemplateList(params map[string]string) (map[string]interface{}, error) {
	return s.apiGet("prod-api/system/template/list", params)
}

func (s *PaperService) SaveTemplate(params map[string]interface{}) (map[string]interface{}, error) {
	return s.apiPost("prod-api/system/template", params)
}

func (s *PaperService) PaperDownload(orderID, fileName string) (map[string]interface{}, error) {
	return s.apiGet("prod-api/system/lunwen/download/"+orderID, map[string]string{"fileName": fileName})
}

func (s *PaperService) GetUpstreamList(params map[string]string) (map[string]interface{}, error) {
	return s.apiGet("prod-api/system/lunwen/list", params)
}

func (s *PaperService) GenerateTask(params map[string]interface{}) (map[string]interface{}, error) {
	return s.apiPost("prod-api/system/lunwen/generate-task", params)
}

func (s *PaperService) GenerateProposal(params map[string]interface{}) (map[string]interface{}, error) {
	return s.apiPost("prod-api/system/lunwen/generate-proposal", params)
}

func (s *PaperService) CountWords(file multipart.File, header *multipart.FileHeader) (map[string]interface{}, error) {
	return s.apiFileUpload("prod-api/system/lunwen/countWords", file, header, nil)
}

func (s *PaperService) UploadCover(file multipart.File, header *multipart.FileHeader) (map[string]interface{}, error) {
	return s.apiFileUpload("prod-api/system/template/uploadCover", file, header, nil)
}

func (s *PaperService) PaperOrderSubmit(uid int, params map[string]interface{}) (map[string]interface{}, error) {
	shopcode, _ := params["shopcode"].(string)
	if shopcode == "" {
		return nil, fmt.Errorf("请选择商品类型")
	}

	addprice := s.GetUserAddPrice(uid)

	var totalPrice float64
	basePrice := s.GetConfigPrice("lunwen_api_" + shopcode + "_price")
	totalPrice += math.Round(basePrice*addprice*100) / 100

	if ktbg, ok := params["ktbg"]; ok {
		if v, _ := toFloat(ktbg); v == 1 {
			totalPrice += math.Round(s.GetConfigPrice("lunwen_api_ktbg_price")*addprice*100) / 100
		}
	}
	if rws, ok := params["rws"]; ok {
		if v, _ := toFloat(rws); v == 1 {
			totalPrice += math.Round(s.GetConfigPrice("lunwen_api_rws_price")*addprice*100) / 100
		}
	}
	if jc, ok := params["jiangchong"]; ok {
		if v, _ := toFloat(jc); v == 1 {
			totalPrice += math.Round(s.GetConfigPrice("lunwen_api_jdaigchj_price")*addprice*100) / 100
		}
	}

	money := s.GetUserMoney(uid)
	if money < totalPrice {
		return nil, fmt.Errorf("余额不足")
	}

	result, err := s.apiPost("prod-api/system/lunwen/xiadan", params)
	if err != nil {
		return nil, err
	}

	code, _ := result["code"].(float64)
	if int(code) != 200 {
		msg, _ := result["msg"].(string)
		return nil, fmt.Errorf("下单失败: %s", msg)
	}

	title, _ := params["title"].(string)
	listParams := map[string]string{
		"pageNum":  "1",
		"pageSize": "10",
		"shopname": "论文" + shopcode + "字",
		"title":    title,
	}
	listResult, err := s.GetUpstreamList(listParams)
	if err == nil {
		listCode, _ := listResult["code"].(float64)
		if int(listCode) == 200 {
			if rows, ok := listResult["rows"].([]interface{}); ok && len(rows) > 0 {
				if row, ok := rows[0].(map[string]interface{}); ok {
					orderID := fmt.Sprintf("%v", row["id"])
					s.insertOrder(uid, orderID, shopcode, title, totalPrice)
				}
			}
		}
	}

	s.deductMoney(uid, totalPrice, "lunwen-下单成功", title)

	return map[string]interface{}{"code": 200, "msg": "下单成功"}, nil
}

func (s *PaperService) FileDedupSubmit(uid int, file multipart.File, header *multipart.FileHeader, wordCount int, aigc, jiangchong int) (map[string]interface{}, error) {
	addprice := s.GetUserAddPrice(uid)

	var totalPrice float64
	if aigc == 1 {
		totalPrice += math.Round(float64(wordCount)/1000*s.GetConfigPrice("lunwen_api_xgdl_price")*addprice*100) / 100
	}
	if jiangchong == 1 {
		totalPrice += math.Round(float64(wordCount)/1000*s.GetConfigPrice("lunwen_api_jdaigcl_price")*addprice*100) / 100
	}

	money := s.GetUserMoney(uid)
	if money < totalPrice {
		return nil, fmt.Errorf("余额不足")
	}

	extra := map[string]string{
		"wordCount":  strconv.Itoa(wordCount),
		"aigc":       strconv.Itoa(aigc),
		"jiangchong": strconv.Itoa(jiangchong),
	}
	result, err := s.apiFileUpload("prod-api/system/lunwen/jiangchong", file, header, extra)
	if err != nil {
		return nil, err
	}

	code, _ := result["code"].(float64)
	if int(code) == 200 {
		listParams := map[string]string{
			"pageNum":  "1",
			"pageSize": "10",
			"shopname": "论文降重",
		}
		listResult, err := s.GetUpstreamList(listParams)
		if err == nil {
			listCode, _ := listResult["code"].(float64)
			if int(listCode) == 200 {
				if rows, ok := listResult["rows"].([]interface{}); ok && len(rows) > 0 {
					if row, ok := rows[0].(map[string]interface{}); ok {
						orderID := fmt.Sprintf("%v", row["id"])
						rowTitle := fmt.Sprintf("%v", row["title"])
						rowShopcode := fmt.Sprintf("%v", row["shopcode"])
						s.insertOrder(uid, orderID, rowShopcode, rowTitle, totalPrice)
					}
				}
			}
		}
		s.deductMoney(uid, totalPrice, "lunwen-文件降重成功", "文件降重")
	}

	return result, nil
}

func (s *PaperService) TextRewriteSubmit(uid int, content string, w http.ResponseWriter) error {
	charCount := len([]rune(content))
	addprice := s.GetUserAddPrice(uid)
	price := math.Round(float64(charCount)/1000*s.GetConfigPrice("lunwen_api_jcl_price")*addprice*100) / 100

	money := s.GetUserMoney(uid)
	if money < price {
		return fmt.Errorf("余额不足")
	}

	err := s.apiStreamRequest("prod-api/system/lunwen/rewrite/stream", map[string]interface{}{
		"content": content,
	}, w)
	if err != nil {
		return err
	}

	listParams := map[string]string{"pageNum": "1", "pageSize": "10"}
	listResult, err := s.GetUpstreamList(listParams)
	if err == nil {
		listCode, _ := listResult["code"].(float64)
		if int(listCode) == 200 {
			if rows, ok := listResult["rows"].([]interface{}); ok && len(rows) > 0 {
				if row, ok := rows[0].(map[string]interface{}); ok {
					orderID := fmt.Sprintf("%v", row["id"])
					rowTitle := fmt.Sprintf("%v", row["title"])
					rowShopcode := fmt.Sprintf("%v", row["shopcode"])
					s.insertOrder(uid, orderID, rowShopcode, rowTitle, price)
				}
			}
		}
	}
	s.deductMoney(uid, price, "lunwen-文本降重成功", "文本降重")

	return nil
}

func (s *PaperService) TextRewriteAIGCSubmit(uid int, content string, w http.ResponseWriter) error {
	charCount := len([]rune(content))
	addprice := s.GetUserAddPrice(uid)
	price := math.Round(float64(charCount)/1000*s.GetConfigPrice("lunwen_api_jdaigcl_price")*addprice*100) / 100

	money := s.GetUserMoney(uid)
	if money < price {
		return fmt.Errorf("余额不足")
	}

	err := s.apiStreamRequest("prod-api/system/lunwen/rewrite-aigc/stream", map[string]interface{}{
		"content": content,
	}, w)
	if err != nil {
		return err
	}

	listParams := map[string]string{"pageNum": "1", "pageSize": "10", "shopname": "降aigc"}
	listResult, err := s.GetUpstreamList(listParams)
	if err == nil {
		listCode, _ := listResult["code"].(float64)
		if int(listCode) == 200 {
			if rows, ok := listResult["rows"].([]interface{}); ok && len(rows) > 0 {
				if row, ok := rows[0].(map[string]interface{}); ok {
					orderID := fmt.Sprintf("%v", row["id"])
					rowTitle := fmt.Sprintf("%v", row["title"])
					rowShopcode := fmt.Sprintf("%v", row["shopcode"])
					s.insertOrder(uid, orderID, rowShopcode, rowTitle, price)
				}
			}
		}
	}
	s.deductMoney(uid, price, "lunwen-文本降AIGC成功", "降AIGC")

	return nil
}

func (s *PaperService) PaperParaEditSubmit(uid int, content, yijian string, w http.ResponseWriter) error {
	charCount := len([]rune(content))
	addprice := s.GetUserAddPrice(uid)
	price := math.Round(float64(charCount)/1000*s.GetConfigPrice("lunwen_api_xgdl_price")*addprice*100) / 100

	money := s.GetUserMoney(uid)
	if money < price {
		return fmt.Errorf("余额不足")
	}

	err := s.apiStreamRequest("prod-api/system/lunwen/xiugai/stream", map[string]interface{}{
		"content": content,
		"yijian":  yijian,
	}, w)
	if err != nil {
		return err
	}

	listParams := map[string]string{"pageNum": "1", "pageSize": "10", "shopname": "段落修改"}
	listResult, err := s.GetUpstreamList(listParams)
	if err == nil {
		listCode, _ := listResult["code"].(float64)
		if int(listCode) == 200 {
			if rows, ok := listResult["rows"].([]interface{}); ok && len(rows) > 0 {
				if row, ok := rows[0].(map[string]interface{}); ok {
					orderID := fmt.Sprintf("%v", row["id"])
					rowTitle := fmt.Sprintf("%v", row["title"])
					rowShopcode := fmt.Sprintf("%v", row["shopcode"])
					s.insertOrder(uid, orderID, rowShopcode, rowTitle, price)
				}
			}
		}
	}
	s.deductMoney(uid, price, "lunwen-段落修改成功", "段落修改")

	return nil
}

func (s *PaperService) GenerateTaskWithFee(uid int, orderID string) (map[string]interface{}, error) {
	addprice := s.GetUserAddPrice(uid)
	price := math.Round(s.GetConfigPrice("lunwen_api_rws_price")*addprice*100) / 100

	money := s.GetUserMoney(uid)
	if money < price {
		return nil, fmt.Errorf("余额不足")
	}

	result, err := s.GenerateTask(map[string]interface{}{"id": orderID})
	if err != nil {
		return nil, err
	}

	code, _ := result["code"].(float64)
	if int(code) == 200 {
		s.insertOrder(uid, orderID, "rws", "任务书生成", price)
		s.deductMoney(uid, price, "lunwen-生成任务书", "任务书")
	}

	return result, nil
}

func (s *PaperService) GenerateProposalWithFee(uid int, orderID string) (map[string]interface{}, error) {
	addprice := s.GetUserAddPrice(uid)
	price := math.Round(s.GetConfigPrice("lunwen_api_ktbg_price")*addprice*100) / 100

	money := s.GetUserMoney(uid)
	if money < price {
		return nil, fmt.Errorf("余额不足")
	}

	result, err := s.GenerateProposal(map[string]interface{}{"id": orderID})
	if err != nil {
		return nil, err
	}

	code, _ := result["code"].(float64)
	if int(code) == 200 {
		s.insertOrder(uid, orderID, "ktbg", "开题报告生成", price)
		s.deductMoney(uid, price, "lunwen-生成开题报告", "开题报告")
	}

	return result, nil
}

func (s *PaperService) GetOrderList(uid int, isAdmin bool, page, pageSize int, searchParams map[string]string) (map[string]interface{}, error) {
	offset := (page - 1) * pageSize

	whereClause := ""
	if !isAdmin {
		whereClause = fmt.Sprintf("WHERE uid = %d", uid)
	}

	var total int
	database.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM qingka_wangke_lunwen %s", whereClause)).Scan(&total)

	rows, err := database.DB.Query(
		fmt.Sprintf("SELECT id, uid, order_id, shopcode, title, price FROM qingka_wangke_lunwen %s ORDER BY id DESC LIMIT ?, ?", whereClause),
		offset, pageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	localOrders := make(map[string]PaperOrder)
	for rows.Next() {
		var o PaperOrder
		rows.Scan(&o.ID, &o.UID, &o.OrderID, &o.ShopCode, &o.Title, &o.Price)
		localOrders[o.OrderID] = o
	}

	if len(localOrders) == 0 {
		return map[string]interface{}{
			"code":  200,
			"msg":   "查询成功",
			"rows":  []interface{}{},
			"total": 0,
		}, nil
	}

	upstreamParams := map[string]string{
		"pageNum":  "1",
		"pageSize": "1000",
	}
	for k, v := range searchParams {
		if v != "" {
			upstreamParams[k] = v
		}
	}

	apiResult, err := s.GetUpstreamList(upstreamParams)
	if err != nil {
		return nil, err
	}

	apiCode, _ := apiResult["code"].(float64)
	if int(apiCode) != 200 {
		return map[string]interface{}{
			"code":  200,
			"msg":   "查询成功",
			"rows":  []interface{}{},
			"total": 0,
		}, nil
	}

	var mergedRows []interface{}
	if apiRows, ok := apiResult["rows"].([]interface{}); ok {
		for _, item := range apiRows {
			if row, ok := item.(map[string]interface{}); ok {
				rowID := fmt.Sprintf("%v", row["id"])
				if localOrder, exists := localOrders[rowID]; exists {
					row["price"] = localOrder.Price
					mergedRows = append(mergedRows, row)
				}
			}
		}
	}

	return map[string]interface{}{
		"code":  200,
		"msg":   "查询成功",
		"rows":  mergedRows,
		"total": total,
	}, nil
}

func (s *PaperService) GetPriceInfo(uid int) map[string]interface{} {
	addprice := s.GetUserAddPrice(uid)
	return map[string]interface{}{
		"price_6000":     math.Round(s.GetConfigPrice("lunwen_api_6000_price")*addprice*100) / 100,
		"price_8000":     math.Round(s.GetConfigPrice("lunwen_api_8000_price")*addprice*100) / 100,
		"price_10000":    math.Round(s.GetConfigPrice("lunwen_api_10000_price")*addprice*100) / 100,
		"price_12000":    math.Round(s.GetConfigPrice("lunwen_api_12000_price")*addprice*100) / 100,
		"price_15000":    math.Round(s.GetConfigPrice("lunwen_api_15000_price")*addprice*100) / 100,
		"price_rws":      math.Round(s.GetConfigPrice("lunwen_api_rws_price")*addprice*100) / 100,
		"price_ktbg":     math.Round(s.GetConfigPrice("lunwen_api_ktbg_price")*addprice*100) / 100,
		"price_jdaigchj": math.Round(s.GetConfigPrice("lunwen_api_jdaigchj_price")*addprice*100) / 100,
		"price_xgdl":     math.Round(s.GetConfigPrice("lunwen_api_xgdl_price")*addprice*100) / 100,
		"price_jcl":      math.Round(s.GetConfigPrice("lunwen_api_jcl_price")*addprice*100) / 100,
		"price_jdaigcl":  math.Round(s.GetConfigPrice("lunwen_api_jdaigcl_price")*addprice*100) / 100,
		"addprice":       addprice,
	}
}

func (s *PaperService) insertOrder(uid int, orderID, shopcode, title string, price float64) {
	_, err := database.DB.Exec(
		"INSERT INTO qingka_wangke_lunwen (uid, order_id, shopcode, title, price) VALUES (?, ?, ?, ?, ?)",
		uid, orderID, shopcode, title, price)
	if err != nil {
		log.Printf("[Paper] 插入订单记录失败: %v", err)
	}
}

func (s *PaperService) deductMoney(uid int, amount float64, logType, desc string) {
	_, err := database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ? LIMIT 1", amount, uid)
	if err != nil {
		log.Printf("[Paper] 扣费失败: %v", err)
		return
	}
	database.DB.Exec(
		"INSERT INTO qingka_wangke_moneylog (uid, type, money, note, addtime) VALUES (?, ?, ?, ?, NOW())",
		uid, logType, -amount, fmt.Sprintf("%s 扣除%.2f元", desc, amount))
}

func toFloat(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case string:
		f, err := strconv.ParseFloat(val, 64)
		return f, err == nil
	case json.Number:
		f, err := val.Float64()
		return f, err == nil
	}
	return 0, false
}
