package jiguang

import (
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
	r.Any("/jiguang/jiguang.api.php", CompatAPI)
}

func CompatAPI(c *gin.Context) {
	act := strings.TrimSpace(firstCompat(c.Query("act"), c.PostForm("act")))
	user, ok := compatAuth(c)
	if !ok {
		return
	}
	switch act {
	case "get_price":
		prices, err := Jiguang().ProductPrices(user.UID)
		compatResult(c, prices, err)
	case "products":
		list, err := Jiguang().Products(user.UID)
		compatResult(c, list, err)
	case "schools":
		page, _ := strconv.Atoi(firstCompat(c.PostForm("page"), "1"))
		pageSize, _ := strconv.Atoi(firstCompat(c.PostForm("pageSize"), "500"))
		result, err := Jiguang().Schools(c.Request.Context(), page, pageSize, c.PostForm("keyword"))
		compatResult(c, result, err)
	case "add":
		req := compatOrderRequest(c)
		result, err := Jiguang().CreateOrder(c.Request.Context(), user.UID, req, "agent", user.UID)
		compatResult(c, result, err)
	case "orders":
		page, _ := strconv.Atoi(firstCompat(c.PostForm("page"), "1"))
		limit, _ := strconv.Atoi(firstCompat(c.PostForm("limit"), "20"))
		list, total, err := Jiguang().ListOrders(user.UID, user.UID == 1, page, limit, c.PostForm("type"), c.PostForm("keywords"), c.PostForm("status"), c.PostForm("school"), 0)
		if err != nil {
			compatError(c, -1, err.Error())
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "success", "data": list, "pagination": gin.H{"page": page, "limit": limit, "total": total}})
	case "refund_preview":
		result, err := Jiguang().RefundOrder(c.Request.Context(), user.UID, c.PostForm("order_no"), user.UID == 1, false)
		compatResult(c, result, err)
	case "refund_confirm":
		result, err := Jiguang().RefundOrder(c.Request.Context(), user.UID, c.PostForm("order_no"), user.UID == 1, true)
		compatResult(c, result, err)
	case "addtimes_preview":
		delta, _ := strconv.Atoi(c.PostForm("delta"))
		result, err := Jiguang().AddTimes(c.Request.Context(), user.UID, c.PostForm("order_no"), delta, user.UID == 1, false)
		compatResult(c, result, err)
	case "addtimes_confirm":
		delta, _ := strconv.Atoi(c.PostForm("delta"))
		result, err := Jiguang().AddTimes(c.Request.Context(), user.UID, c.PostForm("order_no"), delta, user.UID == 1, true)
		compatResult(c, result, err)
	case "order_logs":
		result, err := Jiguang().OrderLogs(c.Request.Context(), user.UID, c.PostForm("order_no"), user.UID == 1)
		compatResult(c, result, err)
	default:
		compatError(c, -1, "未知操作")
	}
}

func compatAuth(c *gin.Context) (compatUser, bool) {
	uid, _ := strconv.Atoi(firstCompat(c.PostForm("login_uid"), c.PostForm("uid"), c.Query("login_uid"), c.Query("uid")))
	key := firstCompat(c.PostForm("login_key"), c.PostForm("key"), c.Query("login_key"), c.Query("key"))
	if uid <= 0 || key == "" {
		compatError(c, -4, "您还没有登录")
		return compatUser{}, false
	}
	var user compatUser
	err := database.DB.QueryRow("SELECT uid, COALESCE(`key`,''), COALESCE(money,0), COALESCE(addprice,1) FROM qingka_wangke_user WHERE uid=?", uid).
		Scan(&user.UID, &user.Key, &user.Money, &user.AddPrice)
	if err != nil {
		compatError(c, -4, "您还没有登录")
		return compatUser{}, false
	}
	if user.Key == "" || user.Key == "0" {
		compatError(c, -1, "你还没有开通接口哦")
		return compatUser{}, false
	}
	if user.Key != key {
		compatError(c, -2, "秘钥错误")
		return compatUser{}, false
	}
	if user.Money < 30 {
		compatError(c, -3, "您的余额不足30元，无法使用API对接")
		return compatUser{}, false
	}
	return user, true
}

func compatOrderRequest(c *gin.Context) OrderRequest {
	productID, _ := strconv.Atoi(c.PostForm("product_id"))
	times, _ := strconv.Atoi(c.PostForm("times"))
	km, _ := strconv.ParseFloat(c.PostForm("km_per_day"), 64)
	return OrderRequest{
		ProductID:       productID,
		SchoolName:      c.PostForm("school_name"),
		StudentName:     c.PostForm("student_name"),
		StudentAccount:  c.PostForm("student_account"),
		Times:           times,
		KMPerDay:        km,
		CustomerMessage: c.PostForm("customer_message"),
	}
}

func compatResult(c *gin.Context, data any, err error) {
	if err != nil {
		compatError(c, -1, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "success", "data": data})
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
