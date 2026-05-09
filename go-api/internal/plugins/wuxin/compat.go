package wuxin

import (
	"net/http"
	"strconv"
	"strings"

	"go-api/internal/database"

	"github.com/gin-gonic/gin"
)

type wuxinCompatUser struct {
	UID      int
	Money    float64
	AddPrice float64
}

func RegisterCompatRoutes(r *gin.Engine) {
	r.Any("/wuxin/api.php", CompatAPI)
}

func CompatAPI(c *gin.Context) {
	action := strings.TrimSpace(c.Query("act"))
	if action == "" {
		action = strings.TrimSpace(c.PostForm("act"))
	}
	if action == "getWuxinSdxyPrice" {
		user := wuxinCompatUser{UID: 0, AddPrice: 1}
		if u, ok := compatAuth(c); ok {
			user = u
		}
		compatPrice(c, user)
		return
	}
	user, ok := compatAuth(c)
	if !ok {
		return
	}

	switch action {
	case "getWuxinSdxySchoolInfo":
		result, err := Wuxin().SchoolInfo(c.Request.Context(), c.PostForm("auth_code"))
		compatResult(c, result, err)
	case "addWuxinSdxyOrder":
		req := compatOrderRequest(c)
		result, err := Wuxin().CreateOrder(c.Request.Context(), user.UID, req, "agent", user.UID)
		compatResult(c, result, err)
	case "getWuxinSdxyOrders", "getWuxinSdxyOrdersList":
		page, _ := strconv.Atoi(firstNonEmpty(c.PostForm("page"), c.Query("page"), "1"))
		limit, _ := strconv.Atoi(firstNonEmpty(c.PostForm("limit"), c.PostForm("pageSize"), c.Query("limit"), c.Query("pageSize"), "20"))
		var status *int
		if raw := firstNonEmpty(c.PostForm("status"), c.Query("status")); raw != "" {
			n, _ := strconv.Atoi(raw)
			status = &n
		}
		list, total, err := Wuxin().ListOrders(user.UID, user.UID == 1, page, limit, firstNonEmpty(c.PostForm("type"), c.Query("type")), firstNonEmpty(c.PostForm("keywords"), c.Query("keywords")), status)
		if err != nil {
			compatError(c, -1, err.Error())
			return
		}
		compatSuccess(c, gin.H{"list": list, "total": total, "page": page, "limit": limit})
	case "deleteWuxinSdxyOrder":
		id, _ := strconv.Atoi(c.PostForm("id"))
		result, err := Wuxin().RefundOrder(c.Request.Context(), user.UID, id, c.PostForm("order_number"), user.UID == 1)
		compatResult(c, result, err)
	case "getWuxinSdxyOrderRecords":
		id, _ := strconv.Atoi(c.PostForm("id"))
		page, _ := strconv.Atoi(firstNonEmpty(c.PostForm("page"), "1"))
		limit, _ := strconv.Atoi(firstNonEmpty(c.PostForm("limit"), "20"))
		result, err := Wuxin().OrderRecords(c.Request.Context(), user.UID, id, c.PostForm("order_number"), page, limit, user.UID == 1)
		compatResult(c, result, err)
	case "getWuxinSdxyOrderConfig":
		id, _ := strconv.Atoi(c.PostForm("id"))
		result, err := Wuxin().OrderConfig(c.Request.Context(), user.UID, id, c.PostForm("order_number"), user.UID == 1)
		compatResult(c, result, err)
	case "editWuxinSdxyOrder":
		id, _ := strconv.Atoi(c.PostForm("id"))
		err := Wuxin().EditOrder(c.Request.Context(), user.UID, id, compatOrderRequest(c), user.UID == 1)
		compatResult(c, gin.H{"message": "编辑订单成功"}, err)
	case "increaseWuxinSdxyOrder":
		id, _ := strconv.Atoi(c.PostForm("id"))
		quantity, _ := strconv.Atoi(c.PostForm("quantity"))
		result, err := Wuxin().IncreaseOrder(c.Request.Context(), user.UID, id, c.PostForm("order_number"), quantity, user.UID == 1)
		compatResult(c, result, err)
	case "reassignWuxinSdxyOrder":
		id, _ := strconv.Atoi(c.PostForm("id"))
		err := Wuxin().ReassignOrder(c.Request.Context(), user.UID, id, c.PostForm("order_number"), user.UID == 1)
		compatResult(c, gin.H{"message": "重新分配成功"}, err)
	case "editWuxinSdxyRunTime", "rerunWuxinSdxyOrder":
		compatError(c, -1, Wuxin().UnsupportedTaskAction().Error())
	default:
		compatError(c, -1, "未知操作")
	}
}

func compatAuth(c *gin.Context) (wuxinCompatUser, bool) {
	uid, _ := strconv.Atoi(firstNonEmpty(c.Query("u_uid"), c.Query("uid"), c.PostForm("u_uid"), c.PostForm("uid")))
	key := firstNonEmpty(c.Query("key"), c.PostForm("key"))
	if uid <= 0 || key == "" {
		compatError(c, -1, "API参数不完整")
		return wuxinCompatUser{}, false
	}
	var user wuxinCompatUser
	var dbKey string
	err := database.DB.QueryRow(
		"SELECT uid, COALESCE(`key`,''), COALESCE(money,0), COALESCE(addprice,1) FROM qingka_wangke_user WHERE uid = ?",
		uid,
	).Scan(&user.UID, &dbKey, &user.Money, &user.AddPrice)
	if err != nil || dbKey != key {
		compatError(c, -1, "API密钥验证失败")
		return wuxinCompatUser{}, false
	}
	if user.AddPrice <= 0 {
		user.AddPrice = 1
	}
	return user, true
}

func compatPrice(c *gin.Context, user wuxinCompatUser) {
	result, err := Wuxin().GetPrice(user.UID)
	compatResult(c, result, err)
}

func compatOrderRequest(c *gin.Context) WuxinOrderRequest {
	runType, _ := strconv.Atoi(firstNonEmpty(c.PostForm("run_type"), "1"))
	orderNum, _ := strconv.Atoi(firstNonEmpty(c.PostForm("order_num"), "1"))
	runMeter, _ := strconv.ParseFloat(firstNonEmpty(c.PostForm("run_meter"), "0"), 64)
	return WuxinOrderRequest{
		AuthCode:    strings.TrimSpace(c.PostForm("auth_code")),
		StartDate:   strings.TrimSpace(c.PostForm("start_date")),
		RunPlanCode: strings.TrimSpace(c.PostForm("run_plan_code")),
		FenceCode:   strings.TrimSpace(c.PostForm("fence_code")),
		ZoneName:    strings.TrimSpace(c.PostForm("zone_name")),
		RunType:     runType,
		RunTime:     strings.TrimSpace(c.PostForm("run_time")),
		RunMeter:    runMeter,
		RunWeek:     strings.TrimSpace(c.PostForm("run_week")),
		RunSpeed:    strings.TrimSpace(c.PostForm("run_speed")),
		OrderNum:    orderNum,
		Mark:        strings.TrimSpace(c.PostForm("mark")),
	}
}

func compatResult(c *gin.Context, data map[string]any, err error) {
	if err != nil {
		compatError(c, -1, err.Error())
		return
	}
	compatSuccess(c, data)
}

func compatSuccess(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "success", "data": data})
}

func compatError(c *gin.Context, code int, msg string) {
	if code == 0 {
		code = -1
	}
	c.JSON(http.StatusOK, gin.H{"code": code, "msg": msg})
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
