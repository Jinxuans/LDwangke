package user

import (
	"fmt"
	"strconv"

	"go-api/internal/database"
	"go-api/internal/model"
	classmodule "go-api/internal/modules/class"
	"go-api/internal/response"
	"go-api/internal/ws"

	"github.com/gin-gonic/gin"
)

func loadGradeList() ([]model.Grade, error) {
	rows, err := database.DB.Query("SELECT id, COALESCE(sort,'0'), COALESCE(name,''), COALESCE(rate,'1'), COALESCE(money,'0'), COALESCE(addkf,'1'), COALESCE(gjkf,'1'), COALESCE(status,'1'), CASE WHEN time IS NOT NULL AND time != '' AND time != '0' THEN FROM_UNIXTIME(CAST(time AS UNSIGNED), '%Y-%m-%d %H:%i') ELSE '' END FROM qingka_wangke_dengji WHERE status = '1' ORDER BY sort ASC, id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Grade
	for rows.Next() {
		var g model.Grade
		rows.Scan(&g.ID, &g.Sort, &g.Name, &g.Rate, &g.Money, &g.AddKF, &g.GJKF, &g.Status, &g.Time)
		list = append(list, g)
	}
	if list == nil {
		list = []model.Grade{}
	}
	return list, nil
}

func ticketEnabledForOrder(oid int) bool {
	var ticket int
	err := database.DB.QueryRow(
		`SELECT COALESCE(f.ticket,0)
		 FROM qingka_wangke_order o
		 JOIN qingka_wangke_class c ON c.cid = o.cid
		 JOIN qingka_wangke_fenlei f ON f.id = CAST(c.fenlei AS UNSIGNED)
		 WHERE o.oid = ?`, oid,
	).Scan(&ticket)
	return err == nil && ticket != 0
}

func UserProfile(c *gin.Context) {
	uid := c.GetInt("uid")
	grade := c.GetString("grade")
	profile, err := userService.Profile(uid, grade)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, profile)
}

func UserChangePassword(c *gin.Context) {
	uid := c.GetInt("uid")
	var req model.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写旧密码和新密码")
		return
	}
	if err := userService.ChangePassword(uid, req.OldPass, req.NewPass); err != nil {
		response.BusinessError(c, 1002, err.Error())
		return
	}
	response.SuccessMsg(c, "密码修改成功")
}

func UserChangePass2(c *gin.Context) {
	uid := c.GetInt("uid")
	var body struct {
		OldPass2 string `json:"old_pass2"`
		NewPass2 string `json:"new_pass2" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "请填写新二级密码")
		return
	}
	if err := userService.ChangePass2(uid, body.OldPass2, body.NewPass2); err != nil {
		response.BusinessError(c, 1002, err.Error())
		return
	}
	response.SuccessMsg(c, "二级密码修改成功")
}

func SendChangeEmailCode(c *gin.Context) {
	uid := c.GetInt("uid")
	var req model.ChangeEmailCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请输入有效的邮箱地址")
		return
	}
	if err := userService.SendChangeEmailCode(uid, req.NewEmail); err != nil {
		response.BusinessError(c, 1006, err.Error())
		return
	}
	response.SuccessMsg(c, "验证码已发送到新邮箱")
}

func ChangeEmail(c *gin.Context) {
	uid := c.GetInt("uid")
	var req model.ChangeEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := userService.ChangeEmail(uid, req.NewEmail, req.Code); err != nil {
		response.BusinessError(c, 1007, err.Error())
		return
	}
	response.SuccessMsg(c, "邮箱变更成功")
}

func UserPayChannels(c *gin.Context) {
	uid := c.GetInt("uid")
	channels, err := userService.GetPayChannels(uid)
	if err != nil {
		response.ServerError(c, "获取支付渠道失败")
		return
	}
	response.Success(c, channels)
}

func UserCreatePay(c *gin.Context) {
	uid := c.GetInt("uid")
	var req model.PayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请输入充值金额")
		return
	}
	domain := c.Request.Host
	order, err := userService.CreatePayOrder(uid, req.Money, req.Type, domain)
	if err != nil {
		response.BusinessError(c, 1003, err.Error())
		return
	}
	response.Success(c, order)
}

func UserPayOrders(c *gin.Context) {
	uid := c.GetInt("uid")
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	orders, total, err := userService.PayOrders(uid, page, limit)
	if err != nil {
		response.ServerError(c, "查询充值记录失败")
		return
	}
	response.Success(c, gin.H{
		"list":       orders,
		"pagination": gin.H{"page": page, "limit": limit, "total": total},
	})
}

func UserMoneyLog(c *gin.Context) {
	uid := c.GetInt("uid")
	var req model.MoneyLogListRequest
	_ = c.ShouldBindQuery(&req)
	logs, total, err := userService.MoneyLogList(uid, req)
	if err != nil {
		response.ServerError(c, "查询流水失败")
		return
	}
	response.Success(c, gin.H{
		"list":       logs,
		"pagination": gin.H{"page": req.Page, "limit": req.Limit, "total": total},
	})
}

func UserTicketList(c *gin.Context) {
	uid := c.GetInt("uid")
	grade := c.GetString("grade")
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	tickets, total, err := userService.TicketList(uid, grade, page, limit)
	if err != nil {
		response.ServerError(c, "查询工单失败")
		return
	}
	response.Success(c, gin.H{
		"list":       tickets,
		"pagination": gin.H{"page": page, "limit": limit, "total": total},
	})
}

func UserTicketCreate(c *gin.Context) {
	uid := c.GetInt("uid")
	var req model.TicketCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写工单内容")
		return
	}

	if req.OID > 0 {
		if !ticketEnabledForOrder(req.OID) {
			response.BadRequest(c, "该分类未开启工单功能")
			return
		}
	}

	id, err := userService.TicketCreate(uid, req)
	if err != nil {
		response.ServerError(c, "创建工单失败")
		return
	}
	if ws.GlobalHub != nil {
		ws.GlobalHub.PushToUser(1, ws.PushMessage{
			Type:    "new_ticket",
			Title:   "新工单",
			Content: fmt.Sprintf("用户 %d 提交了新工单", uid),
			Data: map[string]interface{}{
				"ticket_id": id,
				"uid":       uid,
				"oid":       req.OID,
			},
		})
	}
	response.Success(c, gin.H{"id": id})
}

func UserTicketReply(c *gin.Context) {
	uid := c.GetInt("uid")
	grade := c.GetString("grade")
	var req model.TicketReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写回复内容")
		return
	}
	if err := userService.TicketReply(uid, grade, req); err != nil {
		response.BusinessError(c, 1004, err.Error())
		return
	}
	response.SuccessMsg(c, "回复成功")
}

func UserTicketClose(c *gin.Context) {
	uid := c.GetInt("uid")
	grade := c.GetString("grade")
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		response.BadRequest(c, "无效的工单ID")
		return
	}
	if err := userService.TicketClose(uid, grade, id); err != nil {
		response.ServerError(c, "关闭工单失败")
		return
	}
	response.SuccessMsg(c, "工单已关闭")
}

func UserGetFavorites(c *gin.Context) {
	uid := c.GetInt("uid")
	favorites, err := userService.GetFavorites(uid)
	if err != nil {
		response.ServerError(c, "查询收藏失败")
		return
	}
	response.Success(c, favorites)
}

func UserAddFavorite(c *gin.Context) {
	uid := c.GetInt("uid")
	var body struct {
		CID int `json:"cid"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.CID <= 0 {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := userService.AddFavorite(uid, body.CID); err != nil {
		response.BusinessError(c, 1008, err.Error())
		return
	}
	response.SuccessMsg(c, "收藏成功")
}

func UserRemoveFavorite(c *gin.Context) {
	uid := c.GetInt("uid")
	var body struct {
		CID int `json:"cid"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.CID <= 0 {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := userService.RemoveFavorite(uid, body.CID); err != nil {
		response.ServerError(c, "移除收藏失败")
		return
	}
	response.SuccessMsg(c, "移除成功")
}

func UserCheckPayStatus(c *gin.Context) {
	uid := c.GetInt("uid")
	var body struct {
		OutTradeNo string `json:"out_trade_no"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	paid, msg, err := userService.CheckPayStatus(uid, body.OutTradeNo)
	if err != nil {
		response.BusinessError(c, 1009, err.Error())
		return
	}
	status := 0
	if paid {
		status = 1
	}
	response.Success(c, gin.H{"status": status, "msg": msg})
}

func UserSetInviteCode(c *gin.Context) {
	uid := c.GetInt("uid")
	var body struct {
		YQM string `json:"yqm"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.YQM == "" {
		response.BadRequest(c, "请输入邀请码")
		return
	}
	if err := userService.SetInviteCode(uid, body.YQM); err != nil {
		response.BusinessError(c, 1012, err.Error())
		return
	}
	response.SuccessMsg(c, "邀请码设置成功")
}

func UserGradeList(c *gin.Context) {
	uid := c.GetInt("uid")
	grade := c.GetString("grade")

	list, err := loadGradeList()
	if err != nil {
		response.ServerError(c, "查询等级列表失败")
		return
	}

	if grade != "2" && grade != "3" {
		var addprice float64
		database.DB.QueryRow("SELECT COALESCE(addprice,1) FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice)
		var filtered []model.Grade
		for _, g := range list {
			r, _ := strconv.ParseFloat(g.Rate, 64)
			if r >= addprice {
				filtered = append(filtered, g)
			}
		}
		if filtered == nil {
			filtered = []model.Grade{}
		}
		response.Success(c, filtered)
		return
	}
	response.Success(c, list)
}

func UserSetMyGrade(c *gin.Context) {
	uid := c.GetInt("uid")
	grade := c.GetString("grade")
	if grade != "2" && grade != "3" {
		response.BadRequest(c, "仅管理员可自行设置等级")
		return
	}
	var body struct {
		GradeID int `json:"gradeId"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "请选择有效的等级")
		return
	}
	record, err := classmodule.Classes().ResolveSelectedGrade(body.GradeID, true)
	if err != nil {
		response.BadRequest(c, "请选择有效的等级")
		return
	}
	_, err = database.DB.Exec("UPDATE qingka_wangke_user SET grade_id = ?, addprice = ? WHERE uid = ?", record.ID, record.Rate, uid)
	if err != nil {
		response.ServerError(c, "设置失败")
		return
	}
	response.SuccessMsg(c, "等级已更新")
}

func UserSetInviteRate(c *gin.Context) {
	uid := c.GetInt("uid")
	var body struct {
		GradeID int `json:"gradeId"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	var addprice float64
	database.DB.QueryRow("SELECT COALESCE(addprice,1) FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice)

	if err := userService.SetInviteGrade(uid, body.GradeID, addprice); err != nil {
		response.BusinessError(c, 1010, err.Error())
		return
	}
	response.SuccessMsg(c, "设置成功")
}

func UserChangeSecretKey(c *gin.Context) {
	uid := c.GetInt("uid")
	var body struct {
		Type int `json:"type"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	var money float64
	database.DB.QueryRow("SELECT COALESCE(money,0) FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&money)

	key, err := userService.ChangeSecretKey(uid, body.Type, money)
	if err != nil {
		response.BusinessError(c, 1011, err.Error())
		return
	}
	response.Success(c, gin.H{"key": key})
}

func UserSetPushToken(c *gin.Context) {
	uid := c.GetInt("uid")
	var body struct {
		Token string `json:"token"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	_, err := database.DB.Exec("UPDATE qingka_wangke_user SET tuisongtoken = ? WHERE uid = ?", body.Token, uid)
	if err != nil {
		response.ServerError(c, "设置失败")
		return
	}
	response.SuccessMsg(c, "设置成功")
}

func UserLogList(c *gin.Context) {
	uid := c.GetInt("uid")
	grade := c.GetString("grade")
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	logType := c.Query("type")
	keywords := c.Query("keywords")

	list, total, err := userService.LogList(uid, grade, page, limit, logType, keywords)
	if err != nil {
		response.ServerError(c, "查询日志失败")
		return
	}
	response.Success(c, gin.H{
		"list": list,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}
