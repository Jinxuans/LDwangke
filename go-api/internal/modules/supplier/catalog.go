package supplier

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"go-api/internal/database"
	"go-api/internal/model"
)

func (s *Service) GetSupplierCategories(sup *model.SupplierFull) map[string]string {
	cfg := GetPlatformConfig(sup.PT)
	client := &http.Client{Timeout: 8 * time.Second}
	apiURL := resolveConfiguredActionURL(sup.URL, cfg.CategoryPath)
	execResult, err := s.executeConfiguredActionWithClient(
		client,
		sup,
		apiURL,
		cfg.CategoryMethod,
		cfg.CategoryBodyType,
		cfg.CategoryParamMap,
		http.MethodPost,
		"form",
		defaultSupplierAuthParams(sup, cfg.AuthType),
		map[string]string{},
	)
	if err != nil {
		log.Printf("[GetSupplierCategories] pt=%s endpoint=%s 请求失败: %v", sup.PT, apiURL, err)
		return nil
	}

	result := parseCategoryResponse(execResult.Body)
	if len(result) > 0 {
		log.Printf("[GetSupplierCategories] pt=%s endpoint=%s 成功获取 %d 个分类", sup.PT, apiURL, len(result))
		return result
	}

	log.Printf("[GetSupplierCategories] pt=%s endpoint=%s 未获取到分类数据", sup.PT, apiURL)
	return nil
}

func (s *Service) GetSupplierClasses(sup *model.SupplierFull) ([]SupplierClassItem, error) {
	if sup.PT == "yyy" {
		return yyyGetClasses(sup)
	}

	cfg := GetPlatformConfig(sup.PT)
	apiURL := resolveConfiguredActionURL(sup.URL, cfg.ClassListPath)
	execResult, err := s.executeConfiguredAction(
		sup,
		apiURL,
		cfg.ClassListMethod,
		cfg.ClassListBodyType,
		cfg.ClassListParamMap,
		http.MethodPost,
		"form",
		defaultSupplierAuthParams(sup, cfg.AuthType),
		map[string]string{},
	)
	if err != nil {
		return nil, fmt.Errorf("请求上游失败：%v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(execResult.Body, &raw); err != nil {
		return nil, fmt.Errorf("解析响应失败：%s", string(execResult.Body))
	}

	codeVal := ""
	if codeRaw, ok := raw["code"]; ok {
		switch v := codeRaw.(type) {
		case string:
			codeVal = v
		case float64:
			codeVal = fmt.Sprintf("%.0f", v)
		case int:
			codeVal = fmt.Sprintf("%d", v)
		default:
			codeVal = fmt.Sprintf("%v", v)
		}
	}
	if codeVal != "0" && codeVal != "1" {
		msg := ""
		if msgRaw, ok := raw["msg"]; ok {
			msg = fmt.Sprintf("%v", msgRaw)
		}
		if msg == "" {
			msg = "查询上游进度失败"
		}
		return nil, fmt.Errorf("%s", msg)
	}

	var rawItems []map[string]interface{}
	if dataRaw, ok := raw["data"]; ok && dataRaw != nil {
		dataBytes, _ := json.Marshal(dataRaw)
		json.Unmarshal(dataBytes, &rawItems)
	}

	items := make([]SupplierClassItem, 0, len(rawItems))
	for _, d := range rawItems {
		var price float64
		if priceVal, ok := d["price"]; ok {
			switch v := priceVal.(type) {
			case float64:
				price = v
			case string:
				price, _ = strconv.ParseFloat(v, 64)
			}
		}
		catName := ""
		for _, ck := range []string{"category_name", "fenleiName", "fenlei_name", "catname", "typeName", "type_name", "catName"} {
			if cn, ok := d[ck]; ok {
				str := fmt.Sprintf("%v", cn)
				if str != "" && str != "<nil>" {
					catName = str
					break
				}
			}
		}
		items = append(items, SupplierClassItem{
			CID:          fmt.Sprintf("%v", d["cid"]),
			Name:         fmt.Sprintf("%v", d["name"]),
			Price:        price,
			Fenlei:       fmt.Sprintf("%v", d["fenlei"]),
			Content:      fmt.Sprintf("%v", d["content"]),
			CategoryName: catName,
		})
	}

	cateMap := s.GetSupplierCategories(sup)
	if len(cateMap) > 0 {
		for i := range items {
			fid := items[i].Fenlei
			if name, ok := cateMap[fid]; ok {
				items[i].CategoryName = name
			}
		}
	}

	return items, nil
}

func (s *Service) ImportSupplierClasses(hid int, priceRate float64, category string, name string, fd int) (int, int, string, error) {
	sup, err := s.GetSupplierByHID(hid)
	if err != nil {
		return 0, 0, "", err
	}

	classList, err := s.GetSupplierClasses(sup)
	if err != nil {
		return 0, 0, "", err
	}
	if len(classList) == 0 {
		return 0, 0, "", fmt.Errorf("接口返回的数据为空")
	}

	var fenleiID int
	if category != "999999" {
		catName := name
		if catName == "" {
			catName = sup.Name
		}
		err := database.DB.QueryRow(
			"SELECT id FROM qingka_wangke_fenlei WHERE name = ? AND status != 3 ORDER BY id DESC LIMIT 1", catName,
		).Scan(&fenleiID)
		if err != nil && fd == 0 {
			result, err2 := database.DB.Exec(
				"INSERT INTO qingka_wangke_fenlei (sort, name, status, time) VALUES (10, ?, '1', NOW())", catName,
			)
			if err2 == nil {
				id, _ := result.LastInsertId()
				fenleiID = int(id)
			}
		}
	}

	inserted, updated := 0, 0
	for _, item := range classList {
		if category != "999999" && item.Fenlei != category {
			continue
		}

		price := item.Price * priceRate
		var existCount int
		database.DB.QueryRow(
			"SELECT COUNT(*) FROM qingka_wangke_class WHERE docking = ? AND noun = ?", hid, item.CID,
		).Scan(&existCount)

		if existCount > 0 {
			database.DB.Exec(
				"UPDATE qingka_wangke_class SET price = ?, content = ?, status = 1 WHERE docking = ? AND noun = ?",
				price, item.Content, hid, item.CID,
			)
			updated++
		} else if fd == 0 {
			thisFenlei := fenleiID
			if category == "999999" && item.CategoryName != "" {
				var catID int
				err := database.DB.QueryRow(
					"SELECT id FROM qingka_wangke_fenlei WHERE name = ? AND status != 3 ORDER BY id DESC LIMIT 1",
					item.CategoryName,
				).Scan(&catID)
				if err != nil {
					r, e := database.DB.Exec(
						"INSERT INTO qingka_wangke_fenlei (sort, name, status, time) VALUES (10, ?, '1', NOW())", item.CategoryName,
					)
					if e == nil {
						id, _ := r.LastInsertId()
						catID = int(id)
					}
				}
				thisFenlei = catID
			}

			var maxSort int
			database.DB.QueryRow("SELECT COALESCE(MAX(sort),10) FROM qingka_wangke_class").Scan(&maxSort)
			database.DB.Exec(
				"INSERT INTO qingka_wangke_class (name, getnoun, noun, fenlei, queryplat, docking, price, sort, content, addtime) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())",
				item.Name, item.CID, item.CID, thisFenlei, hid, hid, price, maxSort+1, item.Content,
			)
			inserted++
		}
	}

	total := inserted + updated
	msg := fmt.Sprintf("共计%d个，新上架%d个，更新%d个", total, inserted, updated)
	return inserted, updated, msg, nil
}

func parseCategoryResponse(body []byte) map[string]string {
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil
	}

	var dataArr []interface{}
	for _, key := range []string{"data", "list", "fenlei", "category", "categories"} {
		if v, ok := raw[key]; ok && v != nil {
			dataBytes, _ := json.Marshal(v)
			if err := json.Unmarshal(dataBytes, &dataArr); err == nil && len(dataArr) > 0 {
				break
			}
		}
	}
	if len(dataArr) == 0 {
		return nil
	}

	result := map[string]string{}
	for _, item := range dataArr {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		fid := ""
		for _, idKey := range []string{"id", "fid", "fenlei_id", "category_id", "cate_id", "typeId", "type_id"} {
			if v, ok := itemMap[idKey]; ok && v != nil {
				fid = fmt.Sprintf("%v", v)
				break
			}
		}
		fname := ""
		for _, nameKey := range []string{"name", "fname", "fenleiName", "fenlei_name", "category_name", "catname", "typeName", "type_name", "title", "label"} {
			if v, ok := itemMap[nameKey]; ok && v != nil {
				str := fmt.Sprintf("%v", v)
				if str != "" && str != "<nil>" {
					fname = str
					break
				}
			}
		}
		if fid != "" && fid != "<nil>" && fid != "0" && fname != "" {
			result[fid] = fname
		}
	}
	return result
}

func buildSupplierURL(baseURL, act string) string {
	return fmt.Sprintf("%s/api.php?act=%s", normalizeSupplierBaseURL(baseURL), act)
}

func getSupplierToken(sup *model.SupplierFull) string {
	if sup.Pass != "" {
		return sup.Pass
	}
	return sup.Token
}
