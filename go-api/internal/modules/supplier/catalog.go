package supplier

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-api/internal/database"
	"go-api/internal/model"
)

func (s *Service) GetSupplierCategories(sup *model.SupplierFull) map[string]string {
	cfg := GetPlatformConfig(sup.PT)

	baseURL := strings.TrimRight(sup.URL, "/")
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "http://" + baseURL
	}

	client := &http.Client{Timeout: 8 * time.Second}
	tryActs := []string{}
	if cfg.CategoryAct != "" {
		tryActs = append(tryActs, cfg.CategoryAct)
	}
	for _, fallback := range []string{"getfl", "getcate", "getfenlei"} {
		found := false
		for _, act := range tryActs {
			if act == fallback {
				found = true
				break
			}
		}
		if !found {
			tryActs = append(tryActs, fallback)
		}
	}

	for _, act := range tryActs {
		var resp *http.Response
		var err error

		if cfg.ReportAuthType == "token_only" && cfg.UseJSON {
			apiURL := baseURL + "/api/" + act
			jsonData, _ := json.Marshal(map[string]string{"token": getSupplierToken(sup)})
			req, _ := http.NewRequest("POST", apiURL, strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			resp, err = client.Do(req)
		} else {
			cateURL := buildSupplierURL(sup.URL, act)
			formData := url.Values{}
			formData.Set("uid", sup.User)
			formData.Set("key", sup.Pass)
			resp, err = client.PostForm(cateURL, formData)
		}
		if err != nil {
			log.Printf("[GetSupplierCategories] act=%s 请求失败: %v", act, err)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		result := parseCategoryResponse(body)
		if len(result) > 0 {
			log.Printf("[GetSupplierCategories] pt=%s act=%s 成功获取 %d 个分类", sup.PT, act, len(result))
			return result
		}
	}

	log.Printf("[GetSupplierCategories] pt=%s 所有分类 act 均未获取到数据", sup.PT)
	return nil
}

func (s *Service) GetSupplierClasses(sup *model.SupplierFull) ([]SupplierClassItem, error) {
	if sup.PT == "yyy" {
		return yyyGetClasses(sup)
	}

	cfg := GetPlatformConfig(sup.PT)
	var resp *http.Response
	var err error

	if cfg.ReportAuthType == "token_only" && cfg.UseJSON {
		baseURL := strings.TrimRight(sup.URL, "/")
		if !strings.HasPrefix(baseURL, "http") {
			baseURL = "http://" + baseURL
		}
		apiURL := baseURL + "/api/getclass"
		jsonData, _ := json.Marshal(map[string]string{"token": getSupplierToken(sup)})
		req, _ := http.NewRequest("POST", apiURL, strings.NewReader(string(jsonData)))
		req.Header.Set("Content-Type", "application/json")
		resp, err = s.client.Do(req)
	} else {
		apiURL := buildSupplierURL(sup.URL, "getclass")
		formData := url.Values{}
		formData.Set("uid", sup.User)
		formData.Set("key", sup.Pass)
		resp, err = s.client.PostForm(apiURL, formData)
	}
	if err != nil {
		return nil, fmt.Errorf("请求上游失败：%v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败：%v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("解析响应失败：%s", string(body))
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
	baseURL = strings.TrimRight(baseURL, "/")
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "http://" + baseURL
	}
	return fmt.Sprintf("%s/api.php?act=%s", baseURL, act)
}

func getSupplierToken(sup *model.SupplierFull) string {
	if sup.Pass != "" {
		return sup.Pass
	}
	return sup.Token
}
