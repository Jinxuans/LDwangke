package shashou

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"go-api/internal/database"

	"github.com/gin-gonic/gin"
)

type compatUser struct {
	UID      int
	Key      string
	Money    float64
	AddPrice float64
}

func RegisterCompatRoutes(r *gin.Engine) {
	r.Any("/ss_apis.php", CompatAPI)
	r.Any("/shashou/api.php", CompatAPI)
}

func CompatAPI(c *gin.Context) {
	payload := compatJSONPayload(c)
	act := firstCompat(c.Query("act"), c.PostForm("act"), compatString(payload, "act"))
	if act == "getProjects" || act == "getVersionInfo" {
		user, _ := compatAuth(c, payload, true)
		switch act {
		case "getProjects":
			list, err := ShaShou().ListProjects(false)
			compatResult(c, gin.H{"projects": list, "list": list, "user": user}, err)
		default:
			c.JSON(http.StatusOK, ShaShou().VersionInfoPayload(c.Request.Context()))
		}
		return
	}
	user, ok := compatAuth(c, payload, false)
	if !ok {
		return
	}
	switch act {
	case "add_order":
		req := CreateOrderRequest{
			ProjectID:   compatInt(c, payload, "project_id", 0),
			OrderType:   compatInt(c, payload, "order_type", 1),
			IsRushOrder: compatBool(c, payload, "is_rush_order"),
			Accounts:    compatAccounts(c, payload),
		}
		result, err := ShaShou().CreateOrder(c.Request.Context(), user.UID, req, "agent", user.UID)
		compatResult(c, result, err)
	case "query_order":
		req := QueryOrderRequest{
			ProjectID: compatInt(c, payload, "project_id", 0),
			QueryType: compatInt(c, payload, "query_type", OrderTypeQueryNormal),
			Account:   compatField(c, payload, "account", "query_account"),
		}
		result, err := ShaShou().QueryAccount(c.Request.Context(), user.UID, req, "agent", user.UID)
		compatResult(c, result, err)
	case "query_account":
		accountID, _ := strconv.ParseInt(compatField(c, payload, "account_id"), 10, 64)
		if accountID > 0 {
			result, err := ShaShou().QueryStoredAccount(c.Request.Context(), user.UID, accountID, compatBool(c, payload, "force_update"), user.UID == 1, "agent", user.UID)
			compatResult(c, result, err)
			return
		}
		req := QueryOrderRequest{
			ProjectID: compatInt(c, payload, "project_id", 0),
			QueryType: compatInt(c, payload, "query_type", OrderTypeQueryNormal),
			Account:   compatField(c, payload, "account", "query_account"),
		}
		result, err := ShaShou().QueryAccount(c.Request.Context(), user.UID, req, "agent", user.UID)
		compatResult(c, result, err)
	case "refund_order", "refund_account":
		id, _ := strconv.ParseInt(compatField(c, payload, "account_id"), 10, 64)
		req := RefundOrderRequest{
			AccountID: id,
			Account:   compatField(c, payload, "account", "refund_account"),
			ProjectID: compatInt(c, payload, "project_id", 0),
		}
		result, err := ShaShou().RefundAccount(c.Request.Context(), user.UID, req, user.UID == 1, "agent", user.UID)
		compatResult(c, result, err)
	case "get_order_detail":
		id := compatInt(c, payload, "order_id", 0)
		if id <= 0 {
			id = compatInt(c, payload, "id", 0)
		}
		order, err := ShaShou().findOrder(user.UID, id, user.UID == 1)
		if err != nil {
			compatError(c, -1, err.Error())
			return
		}
		order.AccountDetails, _, _ = ShaShou().ListAccounts(user.UID, user.UID == 1, 1, 200, "", order.OrderNo, "", 0, 0)
		order = sanitizeOrderResponse(order, false)
		compatSuccess(c, gin.H{"order": order})
	case "get_orders":
		page := intFormValue(firstCompat(c.PostForm("page"), c.Query("page")), 1)
		limit := intFormValue(firstCompat(c.PostForm("page_size"), c.PostForm("limit"), c.Query("limit")), 20)
		list, total, err := ShaShou().ListOrders(user.UID, user.UID == 1, page, limit, firstCompat(c.PostForm("status"), c.Query("status")), firstCompat(c.PostForm("order_no"), c.Query("order_no")), firstCompat(c.PostForm("account"), c.Query("account")), 0)
		if err != nil {
			compatError(c, -1, err.Error())
			return
		}
		compatSuccess(c, gin.H{"list": list, "total": total, "page": page, "page_size": limit})
	case "get_accounts":
		page := intFormValue(firstCompat(c.PostForm("page"), c.Query("page")), 1)
		limit := intFormValue(firstCompat(c.PostForm("page_size"), c.PostForm("limit"), c.Query("limit")), 20)
		orderType := intFormValue(firstCompat(c.PostForm("order_type"), c.Query("order_type")), 0)
		list, total, err := ShaShou().ListAccounts(user.UID, user.UID == 1, page, limit, firstCompat(c.PostForm("status"), c.Query("status")), firstCompat(c.PostForm("order_no"), c.Query("order_no")), firstCompat(c.PostForm("account"), c.Query("account")), orderType, 0)
		if err != nil {
			compatError(c, -1, err.Error())
			return
		}
		compatSuccess(c, gin.H{"list": list, "total": total, "page": page, "page_size": limit})
	case "sync_order":
		id := compatInt(c, payload, "order_id", 0)
		if id <= 0 {
			id = compatInt(c, payload, "id", 0)
		}
		result, err := ShaShou().SyncOrder(c.Request.Context(), user.UID, id, user.UID == 1)
		compatResult(c, result, err)
	case "check_query_status":
		accountID, _ := strconv.ParseInt(compatField(c, payload, "account_id"), 10, 64)
		result, err := ShaShou().CheckQueryStatus(user.UID, accountID, user.UID == 1)
		compatResult(c, result, err)
	case "clear_query_result":
		accountID, _ := strconv.ParseInt(compatField(c, payload, "account_id"), 10, 64)
		err := ShaShou().ClearQueryResult(user.UID, accountID, user.UID == 1)
		if err != nil {
			compatError(c, -1, err.Error())
			return
		}
		compatSuccess(c, gin.H{"message": "查询结果已清除"})
	default:
		compatError(c, -1, "未知接口")
	}
}

func compatAuth(c *gin.Context, payload map[string]any, optional bool) (compatUser, bool) {
	uid, _ := strconv.Atoi(compatField(c, payload, "u_uid", "uid"))
	key := compatField(c, payload, "key")
	if uid <= 0 || key == "" {
		if optional {
			return compatUser{AddPrice: 1}, true
		}
		compatError(c, -1, "API参数不完整")
		return compatUser{}, false
	}
	var user compatUser
	err := database.DB.QueryRow("SELECT uid, COALESCE(`key`,''), COALESCE(money,0), COALESCE(addprice,1) FROM qingka_wangke_user WHERE uid=?", uid).
		Scan(&user.UID, &user.Key, &user.Money, &user.AddPrice)
	if err != nil || user.Key != key {
		compatError(c, -1, "API密钥验证失败")
		return compatUser{}, false
	}
	if user.AddPrice <= 0 {
		user.AddPrice = 1
	}
	return user, true
}

func compatResult(c *gin.Context, data map[string]any, err error) {
	if err != nil {
		compatError(c, -1, err.Error())
		return
	}
	compatSuccess(c, data)
}

func compatSuccess(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": data})
}

func compatError(c *gin.Context, code int, msg string) {
	if code == 0 {
		code = -1
	}
	c.JSON(http.StatusOK, gin.H{"code": code, "msg": msg})
}

func firstCompat(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func compatJSONPayload(c *gin.Context) map[string]any {
	if !strings.Contains(strings.ToLower(c.GetHeader("Content-Type")), "application/json") {
		return nil
	}
	var payload map[string]any
	if err := c.ShouldBindJSON(&payload); err != nil {
		return nil
	}
	return payload
}

func compatField(c *gin.Context, payload map[string]any, keys ...string) string {
	values := make([]string, 0, len(keys)*3)
	for _, key := range keys {
		values = append(values, c.Query(key), c.PostForm(key), compatString(payload, key))
	}
	return firstCompat(values...)
}

func compatString(payload map[string]any, key string) string {
	if payload == nil {
		return ""
	}
	if value, ok := payload[key]; ok {
		return strings.TrimSpace(asString(value))
	}
	return ""
}

func compatInt(c *gin.Context, payload map[string]any, key string, def int) int {
	return intFormValue(compatField(c, payload, key), def)
}

func compatBool(c *gin.Context, payload map[string]any, key string) bool {
	value := strings.ToLower(compatField(c, payload, key))
	return value == "1" || value == "true" || value == "yes"
}

func compatAccounts(c *gin.Context, payload map[string]any) []AccountForm {
	if payload != nil {
		if value, ok := payload["accounts"]; ok {
			switch typed := value.(type) {
			case string:
				return accountFormsFromString(typed)
			default:
				raw, _ := json.Marshal(typed)
				var accounts []AccountForm
				_ = json.Unmarshal(raw, &accounts)
				return accounts
			}
		}
	}
	return accountFormsFromString(firstCompat(c.PostForm("accounts"), c.Query("accounts")))
}
