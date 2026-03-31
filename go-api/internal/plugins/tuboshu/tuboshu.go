package tuboshu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-api/internal/database"
)

const (
	tbsPlatformName     = "tuboshu"
	tbsTokenCacheTTL    = 86400 * time.Second
	tbsDefaultTTL       = 3600 * time.Second
	tbsMaxRetry         = 3
	tbsRetryDelay       = 500 * time.Millisecond
	tbsRouteTestTTL     = 86400 * time.Second
	tbsRouteTestTimeout = 5 * time.Second
)

var tbsAPIURLs = []string{
	"https://sd.polars.cc/api/",
	"http://sdapi.polars.cc/",
	"https://api.aiwriting.icu/",
}

type TuboshuConfig struct {
	PriceRatio     float64                `json:"price_ratio"`
	PriceConfig    map[string]interface{} `json:"price_config"`
	PageVisibility map[string]bool        `json:"page_visibility"`
}

type TuboshuDialogue struct {
	ID          int     `json:"id"`
	UID         int     `json:"uid"`
	Title       string  `json:"title"`
	State       string  `json:"state"`
	DownloadURL string  `json:"download_url"`
	AddTime     string  `json:"addtime"`
	IP          string  `json:"ip"`
	SourceID    int64   `json:"source_id"`
	DialogueID  string  `json:"dialogue_id"`
	Point       float64 `json:"point"`
	Type        string  `json:"type"`
}

type TuboshuRouteRequest struct {
	Method string                 `json:"method"`
	Path   string                 `json:"path"`
	Params map[string]interface{} `json:"params"`
	IsBlob bool                   `json:"isBlob"`
}

type TuboshuService struct {
	client  *http.Client
	mu      sync.RWMutex
	bestURL string
}

var tuboshuService *TuboshuService
var tuboshuServiceOnce sync.Once

func Tuboshu() *TuboshuService {
	tuboshuServiceOnce.Do(func() {
		tuboshuService = &TuboshuService{
			client: &http.Client{Timeout: 120 * time.Second},
		}
	})
	return tuboshuService
}

type routeConfig struct {
	pattern *regexp.Regexp
	handler string
	ttl     time.Duration
	replace string
	isBlob  bool
}

var tuboshuRoutes []routeConfig

func init() {
	type routeDef struct {
		pattern string
		handler string
		ttl     time.Duration
		replace string
		isBlob  bool
	}

	defs := []routeDef{
		{`^GET-/dialogue/stage/list$`, "cacheable", 3600 * time.Second, "", false},
		{`^GET-/dialogue/template$`, "cacheable", 3600 * time.Second, "", false},
		{`^GET-/dialogue/list$`, "cacheable", 3600 * time.Second, "", false},
		{`^GET-/paint/templates$`, "cacheable", 3600 * time.Second, "", false},
		{`^GET-/dialogue/outlineInsertBtn$`, "cacheable", 3600 * time.Second, "", false},
		{`^GET-/dialogue/stage/\d+$`, "cacheable", 3600 * time.Second, "", false},
		{`^GET-/dialogue/paperCategory$`, "cacheable", 3600 * time.Second, "", false},
		{`^GET-/dialogue/csl-styles$`, "cacheable", 3600 * time.Second, "", false},
		{`^GET-/dialogue/paper-outline-types$`, "cacheable", 3600 * time.Second, "", false},
		{`^GET-/dialogue/\d+$`, "cacheable", 3600 * time.Second, "", false},
		{`^GET-/dialogue/wordTemplate$`, "cacheable", 3600 * time.Second, "", false},
		{`^GET-/userInfo$`, "cacheable", 300 * time.Second, "", false},
		{`^GET-/task/reduction/types$`, "cacheable", 3600 * time.Second, "", false},
		{`^POST-/dialogue/parse-proposal$`, "forward", 0, "", false},
		{`^GET-/dialogue/reference/search`, "forward", 0, "", false},
		{`^GET-/dialogue/tool`, "forward", 0, "", false},
		{`^GET-/task/reduction/document-paragraphs$`, "forward", 0, "", false},
		{`^POST-/task/\d+/export`, "forward", 0, "dialogueId", true},
		{`^GET-/task/\d+`, "forward", 0, "dialogueId", false},
		{`^POST-/task/\d+`, "forward", 0, "dialogueId", false},
		{`^GET-/task/list$`, "taskList", 0, "", false},
		{`^POST-/task/reduction$`, "reductionSubmit", 0, "", false},
		{`^POST-/task/reduction/check$`, "reductionCheck", 0, "", false},
		{`^POST-/dialogue/stage`, "dialogueStageSubmit", 0, "", false},
		{`^GET-/dialogue/chat`, "simpleChat", 0, "", false},
		{`^POST-/dialogue/chat`, "simpleChat", 0, "", false},
		{`^POST-/dialogue/outline$`, "outlineSubmit", 0, "", false},
		{`^GET-/dialogue/generateChart`, "chartGenerate", 0, "", false},
		{`^POST-/dialogue/templateGenerate$`, "templateGenerate", 0, "", false},
		{`^PUT-/subsite/`, "savePriceConfig", 0, "", false},
		{`^GET-/tickets`, "ticketRoute", 0, "", false},
		{`^POST-/tickets`, "ticketRoute", 0, "", false},
		{`^PUT-/tickets`, "ticketRoute", 0, "", false},
		{`^GET-/knowledge-bases`, "knowledgeRoute", 0, "", false},
		{`^POST-/knowledge-bases`, "knowledgeRoute", 0, "", false},
		{`^PUT-/knowledge-bases`, "knowledgeRoute", 0, "", false},
		{`^DELETE-/knowledge-bases`, "knowledgeRoute", 0, "", false},
		{`^GET-/points-exchange/products$`, "pointsExchange", 0, "", false},
		{`^POST-/points-exchange/exchange$`, "pointsExchange", 0, "", false},
		{`^GET-/points-exchange/records$`, "pointsExchange", 0, "", false},
		{`^GET-/admin/points-exchange/products$`, "pointsExchange", 0, "", false},
		{`^POST-/admin/points-exchange/products$`, "pointsExchange", 0, "", false},
		{`^DELETE-/admin/points-exchange/products/\d+$`, "pointsExchange", 0, "", false},
		{`^GET-/admin/points-exchange/products/\d+/codes$`, "pointsExchange", 0, "", false},
		{`^POST-/admin/points-exchange/codes$`, "pointsExchange", 0, "", false},
		{`^DELETE-/admin/points-exchange/codes/\d+$`, "pointsExchange", 0, "", false},
		{`^GET-/admin/points-exchange/records$`, "pointsExchange", 0, "", false},
	}

	for _, d := range defs {
		tuboshuRoutes = append(tuboshuRoutes, routeConfig{
			pattern: regexp.MustCompile(d.pattern),
			handler: d.handler,
			ttl:     d.ttl,
			replace: d.replace,
			isBlob:  d.isBlob,
		})
	}
}

func (s *TuboshuService) HandleRoute(uid int, isAdmin bool, req TuboshuRouteRequest, clientIP string) (interface{}, bool, error) {
	method := strings.ToUpper(req.Method)
	path := req.Path
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	route := method + "-" + path

	for _, rc := range tuboshuRoutes {
		if rc.pattern.MatchString(route) {
			switch rc.handler {
			case "cacheable":
				result, err := s.handleCacheableRoute(method, path, req.Params, rc.ttl)
				return result, false, err
			case "forward":
				fwdPath := path
				if rc.replace == "dialogueId" {
					var err error
					fwdPath, err = s.replaceDialogueID(path, uid, isAdmin)
					if err != nil {
						return nil, false, err
					}
				}
				if rc.isBlob {
					body, err := s.upstreamRequest(fwdPath, req.Params, method, true)
					return body, true, err
				}
				result, err := s.upstreamRequestJSON(fwdPath, req.Params, method)
				return result, false, err
			case "taskList":
				return s.handleTaskList(uid, isAdmin, req.Params)
			case "reductionSubmit":
				return s.handleReductionSubmit(uid, req.Params, clientIP)
			case "reductionCheck":
				return s.handleReductionCheck(uid, req.Params)
			case "dialogueStageSubmit":
				return s.handleDialogueStageSubmit(uid, route, req.Params, clientIP)
			case "simpleChat":
				return s.handleSimpleChat(uid, method, route, req.Params, clientIP)
			case "outlineSubmit":
				return s.handleOutlineSubmit(uid, req.Params, clientIP)
			case "chartGenerate":
				return s.handleChartGenerate(uid, route, req.Params, clientIP)
			case "templateGenerate":
				return s.handleTemplateGenerate(uid, req.Params, clientIP)
			case "savePriceConfig":
				if !isAdmin {
					return nil, false, fmt.Errorf("非管理员禁止修改")
				}
				priceConfig, ok := req.Params["priceConfig"].(map[string]interface{})
				if !ok {
					return nil, false, fmt.Errorf("参数错误")
				}
				err := s.SavePriceConfig(priceConfig)
				if err != nil {
					return nil, false, err
				}
				return map[string]interface{}{"success": true}, false, nil
			case "ticketRoute":
				return s.handleTicketRoute(uid, method, path, req.Params)
			case "knowledgeRoute":
				return s.handleKnowledgeRoute(uid, method, path, req.Params)
			case "pointsExchange":
				return s.handlePointsExchange(uid, isAdmin, method, path, req.Params)
			}
		}
	}

	return nil, false, fmt.Errorf("未知的请求路由: %s", route)
}

func (s *TuboshuService) HandleFormDataRoute(path, method string, file multipart.File, fileHeader *multipart.FileHeader) (map[string]interface{}, error) {
	token, err := s.getToken()
	if err != nil {
		return nil, err
	}

	apiURL := s.getBestAPIURL() + strings.TrimLeft(path, "/")

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", fileHeader.Filename)
	if err != nil {
		return nil, fmt.Errorf("创建表单失败: %v", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("复制文件失败: %v", err)
	}
	writer.Close()

	req, err := http.NewRequest(strings.ToUpper(method), apiURL, &buf)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", token)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("响应解析失败: %v", err)
	}
	return result, nil
}

func (s *TuboshuService) replaceDialogueID(path string, uid int, isAdmin bool) (string, error) {
	re := regexp.MustCompile(`(\d+)`)
	matches := re.FindStringSubmatch(path)
	if len(matches) < 2 {
		return path, nil
	}
	dialogueID := matches[1]

	query := "SELECT source_id, uid FROM qingka_wangke_dialogue WHERE id = ? LIMIT 1"
	var sourceID int64
	var orderUID int
	err := database.DB.QueryRow(query, dialogueID).Scan(&sourceID, &orderUID)
	if err != nil || sourceID == 0 {
		return "", fmt.Errorf("订单不存在或未完成")
	}
	if !isAdmin && orderUID != uid {
		return "", fmt.Errorf("订单不属于当前用户")
	}

	return strings.Replace(path, dialogueID, strconv.FormatInt(sourceID, 10), 1), nil
}

func (s *TuboshuService) handleTaskList(uid int, isAdmin bool, params map[string]interface{}) (interface{}, bool, error) {
	page := getIntParam(params, "page", 1)
	size := getIntParam(params, "size", 10)
	offset := (page - 1) * size

	whereClause := "1=1"
	var args []interface{}
	if !isAdmin {
		whereClause = "uid = ?"
		args = append(args, uid)
	}

	var total int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_dialogue WHERE "+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, false, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(size)))
	queryArgs := append(args, offset, size)
	rows, err := database.DB.Query(
		"SELECT id, title, state, point, addtime, download_url, type, source_id, dialogue_id FROM qingka_wangke_dialogue WHERE "+whereClause+" ORDER BY id DESC LIMIT ?, ?",
		queryArgs...,
	)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	var content []map[string]interface{}
	for rows.Next() {
		var id, sourceID int64
		var title, state, addtime, downloadURL, dtype, dialogueID string
		var point float64
		if err := rows.Scan(&id, &title, &state, &point, &addtime, &downloadURL, &dtype, &sourceID, &dialogueID); err != nil {
			continue
		}

		disableRefresh := map[string]bool{"PAPER_OUTLINE": true, "PAPER_WRITING": true}
		canRestart := false
		if sourceID > 0 && state != "FINISHED" && state != "REFUNDED" && !disableRefresh[dtype] {
			sourceOrder, err := s.upstreamRequestJSON(fmt.Sprintf("task/%d", sourceID), nil, "GET")
			if err == nil {
				if data, ok := sourceOrder["data"].(map[string]interface{}); ok {
					if newState, ok := data["status"].(string); ok && newState != state {
						database.DB.Exec("UPDATE qingka_wangke_dialogue SET state = ? WHERE id = ?", newState, id)
						state = newState
					}
					if dl, ok := data["downloadUrl"].(string); ok && dl != "" && downloadURL == "" {
						database.DB.Exec("UPDATE qingka_wangke_dialogue SET download_url = ? WHERE id = ?", dl, id)
						downloadURL = dl
					}
					if cr, ok := data["canRestart"].(bool); ok && cr {
						canRestart = true
					}
				}
			}
		}

		t, _ := time.Parse("2006-01-02 15:04:05", addtime)
		createTimeMs := t.UnixMilli()

		content = append(content, map[string]interface{}{
			"id":           id,
			"originPrompt": title,
			"status":       state,
			"point":        point,
			"createTime":   createTimeMs,
			"updateTime":   createTimeMs,
			"download_url": downloadURL,
			"type":         dtype,
			"sourceId":     sourceID,
			"dialogueId":   dialogueID,
			"owner":        uid,
			"canRestart":   canRestart,
		})
	}
	if content == nil {
		content = []map[string]interface{}{}
	}

	return map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"content":          content,
			"totalElements":    total,
			"totalPages":       totalPages,
			"last":             page >= totalPages,
			"first":            page <= 1,
			"empty":            len(content) == 0,
			"number":           page,
			"numberOfElements": len(content),
			"size":             size,
		},
	}, false, nil
}

func (s *TuboshuService) handleDialogueStageSubmit(uid int, route string, params map[string]interface{}, clientIP string) (interface{}, bool, error) {
	re := regexp.MustCompile(`id=(\d+)`)
	matches := re.FindStringSubmatch(route)
	var dialogueID string
	if len(matches) > 1 {
		dialogueID = matches[1]
	}

	price, err := s.calculateStagePrice(uid, dialogueID, params)
	if err != nil {
		return nil, false, err
	}
	if err := s.checkBalance(uid, price); err != nil {
		return nil, false, err
	}

	path := "dialogue/stage"
	if dialogueID != "" {
		path += "?id=" + dialogueID
	}
	result, err := s.upstreamRequestJSON(path, params, "POST")
	if err != nil {
		return nil, false, err
	}

	success, _ := result["success"].(bool)
	if !success {
		msg, _ := result["message"].(string)
		return nil, false, fmt.Errorf("提交失败: %s", msg)
	}

	data, _ := result["data"].(map[string]interface{})
	sourceID := getInt64FromInterface(data["id"])
	dtype, _ := data["type"].(string)
	prompt, _ := params["prompt"].(string)

	s.saveOrderAndDeduct(uid, sourceID, prompt, dialogueID, price, dtype, clientIP)

	return map[string]interface{}{
		"success": true,
		"message": "提交成功",
		"data":    data,
	}, false, nil
}

func (s *TuboshuService) handleSimpleChat(uid int, method, route string, params map[string]interface{}, clientIP string) (interface{}, bool, error) {
	re := regexp.MustCompile(`id=(\d+)`)
	matches := re.FindStringSubmatch(route)
	dialogueID := ""
	if len(matches) > 1 {
		dialogueID = matches[1]
	}
	if dialogueID == "" {
		if id, ok := params["id"]; ok {
			dialogueID = fmt.Sprintf("%v", id)
		}
	}
	if dialogueID == "" {
		return nil, false, fmt.Errorf("Missing dialogue ID")
	}

	price, err := s.getPriceFromConfig("SIMPLE_DIALOGUE", dialogueID, uid)
	if err != nil {
		return nil, false, err
	}
	if err := s.checkBalance(uid, price); err != nil {
		return nil, false, err
	}

	var result map[string]interface{}
	if method == "GET" {
		prompt, _ := params["prompt"].(string)
		result, err = s.upstreamRequestJSON("dialogue/chat", map[string]interface{}{"id": dialogueID, "prompt": prompt}, "GET")
	} else {
		result, err = s.upstreamRequestJSON("dialogue/chat?id="+dialogueID, params, "POST")
	}
	if err != nil {
		return nil, false, err
	}

	success, _ := result["success"].(bool)
	if !success {
		msg, _ := result["message"].(string)
		return nil, false, fmt.Errorf("生成失败: %s", msg)
	}

	data := result["data"]
	if dataStr, ok := data.(string); ok {
		s.deductFee(uid, price)
		return map[string]interface{}{"success": true, "data": dataStr}, false, nil
	}

	if dataMap, ok := data.(map[string]interface{}); ok {
		sourceID := getInt64FromInterface(dataMap["id"])
		dtype, _ := dataMap["type"].(string)
		prompt := ""
		if p, ok := params["prompt"].(string); ok {
			prompt = p
		} else if p, ok := params["content"].(string); ok {
			prompt = p
		}
		orderID := s.saveOrderAndDeduct(uid, sourceID, prompt, dialogueID, price, dtype, clientIP)
		dataMap["id"] = orderID
		return map[string]interface{}{"success": true, "data": dataMap}, false, nil
	}

	return result, false, nil
}

func (s *TuboshuService) handleOutlineSubmit(uid int, params map[string]interface{}, clientIP string) (interface{}, bool, error) {
	price, err := s.getPriceFromConfig("PAPER_OUTLINE", "", uid)
	if err != nil {
		return nil, false, err
	}
	if err := s.checkBalance(uid, price); err != nil {
		return nil, false, err
	}

	prompt, _ := params["prompt"].(string)
	result, err := s.upstreamRequestJSON("dialogue/outline", map[string]interface{}{"prompt": prompt}, "POST")
	if err != nil {
		return nil, false, err
	}

	success, _ := result["success"].(bool)
	if !success {
		msg, _ := result["message"].(string)
		return nil, false, fmt.Errorf("生成失败: %s", msg)
	}

	data, _ := result["data"].(map[string]interface{})
	sourceID := getInt64FromInterface(data["id"])
	s.saveOrderAndDeduct(uid, sourceID, prompt, "outline", price, "PAPER_OUTLINE", clientIP)

	return map[string]interface{}{"success": true, "data": data}, false, nil
}

func (s *TuboshuService) handleChartGenerate(uid int, route string, params map[string]interface{}, clientIP string) (interface{}, bool, error) {
	price, err := s.getPriceFromConfig("CHART_GENERATE", "", uid)
	if err != nil {
		return nil, false, err
	}
	if err := s.checkBalance(uid, price); err != nil {
		return nil, false, err
	}

	result, err := s.upstreamRequestJSON("dialogue/generateChart", params, "GET")
	if err != nil {
		return nil, false, err
	}

	success, _ := result["success"].(bool)
	if !success {
		msg, _ := result["message"].(string)
		return nil, false, fmt.Errorf("生成失败: %s", msg)
	}

	data, _ := result["data"].(map[string]interface{})
	sourceID := getInt64FromInterface(data["id"])
	prompt, _ := params["prompt"].(string)
	if prompt == "" {
		prompt = "图表生成"
	}
	s.saveOrderAndDeduct(uid, sourceID, prompt, "chart", price, "CHART_GENERATE", clientIP)

	return result, false, nil
}

func (s *TuboshuService) handleTemplateGenerate(uid int, params map[string]interface{}, clientIP string) (interface{}, bool, error) {
	price, err := s.getPriceFromConfig("PAPER_OUTLINE", "", uid)
	if err != nil {
		return nil, false, err
	}
	if err := s.checkBalance(uid, price); err != nil {
		return nil, false, err
	}

	result, err := s.upstreamRequestJSON("dialogue/templateGenerate", params, "POST")
	if err != nil {
		return nil, false, err
	}

	success, _ := result["success"].(bool)
	if !success {
		msg, _ := result["message"].(string)
		return nil, false, fmt.Errorf("生成失败: %s", msg)
	}

	data, _ := result["data"].(map[string]interface{})
	sourceID := getInt64FromInterface(data["id"])
	s.saveOrderAndDeduct(uid, sourceID, "Word模板智能填充", "templateGenerate", price, "TEMPLATE", clientIP)

	return result, false, nil
}

func (s *TuboshuService) handleReductionSubmit(uid int, params map[string]interface{}, clientIP string) (interface{}, bool, error) {
	text, _ := params["text"].(string)
	if text == "" {
		return nil, false, fmt.Errorf("文本不能为空")
	}
	reductionType, _ := params["type"].(string)
	if reductionType == "" {
		reductionType = "NORMAL"
	}

	checkResult, _, err := s.handleReductionCheck(uid, params)
	if err != nil {
		return nil, false, err
	}
	checkMap, ok := checkResult.(map[string]interface{})
	if !ok {
		return nil, false, fmt.Errorf("检查失败")
	}
	success, _ := checkMap["success"].(bool)
	if !success {
		return nil, false, fmt.Errorf("检查失败")
	}

	checkData, _ := checkMap["data"].(map[string]interface{})
	price := getFloat64FromInterface(checkData["price"])
	charCount := getIntFromInterface(checkData["totalCharCount"])
	balanceEnough, _ := checkData["balanceEnough"].(bool)

	if !balanceEnough {
		return nil, false, fmt.Errorf("余额不足")
	}

	result, err := s.upstreamRequestJSON("task/reduction", map[string]interface{}{"text": text, "type": reductionType}, "POST")
	if err != nil {
		return nil, false, err
	}

	resultSuccess, _ := result["success"].(bool)
	if !resultSuccess {
		msg, _ := result["message"].(string)
		return nil, false, fmt.Errorf("生成失败: %s", msg)
	}

	data, _ := result["data"].(map[string]interface{})
	sourceID := getInt64FromInterface(data["id"])
	title := fmt.Sprintf("文本降重 (%d字) - %s", charCount, reductionType)
	orderID := s.saveOrderAndDeduct(uid, sourceID, title, "reduction", price, "PAPER_REDUCTION", clientIP)
	data["id"] = orderID

	return result, false, nil
}

func (s *TuboshuService) handleReductionCheck(uid int, params map[string]interface{}) (interface{}, bool, error) {
	text, _ := params["text"].(string)
	if text == "" {
		return nil, false, fmt.Errorf("文本不能为空")
	}
	reductionType, _ := params["type"].(string)
	if reductionType == "" {
		reductionType = "NORMAL"
	}

	pricePerThousand, err := s.getPriceFromConfig("PAPER_REDUCTION", reductionType, uid)
	if err != nil {
		return nil, false, err
	}

	result, err := s.upstreamRequestJSON("task/reduction/check", map[string]interface{}{"text": text, "type": reductionType}, "POST")
	if err != nil {
		return nil, false, err
	}

	success, _ := result["success"].(bool)
	if !success {
		msg, _ := result["message"].(string)
		return nil, false, fmt.Errorf("检查失败: %s", msg)
	}

	data, _ := result["data"].(map[string]interface{})
	charCount := getFloat64FromInterface(data["totalCharCount"])
	calculatedPrice := math.Round((charCount/1000)*pricePerThousand*1000) / 1000
	data["price"] = calculatedPrice

	var money float64
	database.DB.QueryRow("SELECT money FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&money)
	data["balanceEnough"] = money >= calculatedPrice
	data["balance"] = money

	return result, false, nil
}

func (s *TuboshuService) handleTicketRoute(uid int, method, path string, params map[string]interface{}) (interface{}, bool, error) {
	cleanPath := strings.TrimLeft(path, "/")

	if method == "GET" && cleanPath == "tickets" {
		if _, ok := params["subUid"]; !ok {
			params["subUid"] = strconv.Itoa(uid)
		}
		apiPath := "tickets/by-sub-uid/" + fmt.Sprintf("%v", params["subUid"])
		result, err := s.upstreamRequestJSON(apiPath, params, "GET")
		return result, false, err
	}

	if method == "POST" && cleanPath == "tickets" {
		if subUid, ok := params["subUid"].(string); ok && subUid != "" {
			params["subUid"] = subUid + "_" + strconv.Itoa(uid)
		} else {
			params["subUid"] = strconv.Itoa(uid)
		}
		result, err := s.upstreamRequestJSON("tickets", params, "POST")
		return result, false, err
	}

	result, err := s.upstreamRequestJSON(cleanPath, params, method)
	return result, false, err
}

func (s *TuboshuService) handleKnowledgeRoute(uid int, method, path string, params map[string]interface{}) (interface{}, bool, error) {
	cleanPath := strings.TrimLeft(path, "/")

	if method == "GET" && cleanPath == "knowledge-bases" {
		subUid := strconv.Itoa(uid)
		if su, ok := params["subUid"].(string); ok && su != "" {
			subUid = su + "_" + strconv.Itoa(uid)
		}
		apiPath := "knowledge-bases/by-sub-uid/" + subUid
		result, err := s.upstreamRequestJSON(apiPath, params, "GET")
		return result, false, err
	}

	if method == "POST" && cleanPath == "knowledge-bases" {
		if su, ok := params["subUid"].(string); ok && su != "" {
			params["subUid"] = su + "_" + strconv.Itoa(uid)
		} else {
			params["subUid"] = strconv.Itoa(uid)
		}
		result, err := s.upstreamRequestJSON("knowledge-bases", params, "POST")
		return result, false, err
	}

	result, err := s.upstreamRequestJSON(cleanPath, params, method)
	return result, false, err
}
