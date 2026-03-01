package handler

import (
	"fmt"
	"strconv"

	"go-api/internal/database"
	"go-api/internal/model"
	"go-api/internal/response"
	"go-api/internal/service"
	"go-api/internal/ws"

	"github.com/gin-gonic/gin"
)

var userCenterService = service.NewUserCenterService()

// 用户资料 (按 PHP info case)
func UserProfile(c *gin.Context) {
	uid := c.GetInt("uid")
	grade := c.GetString("grade")
	profile, err := userCenterService.Profile(uid, grade)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, profile)
}

// 修改密码
func UserChangePassword(c *gin.Context) {
	uid := c.GetInt("uid")
	var req model.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写旧密码和新密码")
		return
	}
	if err := userCenterService.ChangePassword(uid, req.OldPass, req.NewPass); err != nil {
		response.BusinessError(c, 1002, err.Error())
		return
	}
	response.SuccessMsg(c, "密码修改成功")
}

// 修改二级密码（管理员专用）
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
	if err := userCenterService.ChangePass2(uid, body.OldPass2, body.NewPass2); err != nil {
		response.BusinessError(c, 1002, err.Error())
		return
	}
	response.SuccessMsg(c, "二级密码修改成功")
}

// 发送邮箱变更验证码
func SendChangeEmailCode(c *gin.Context) {
	uid := c.GetInt("uid")
	var req model.ChangeEmailCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请输入有效的邮箱地址")
		return
	}
	if err := userCenterService.SendChangeEmailCode(uid, req.NewEmail); err != nil {
		response.BusinessError(c, 1006, err.Error())
		return
	}
	response.SuccessMsg(c, "验证码已发送到新邮箱")
}

// 确认变更邮箱
func ChangeEmail(c *gin.Context) {
	uid := c.GetInt("uid")
	var req model.ChangeEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := userCenterService.ChangeEmail(uid, req.NewEmail, req.Code); err != nil {
		response.BusinessError(c, 1007, err.Error())
		return
	}
	response.SuccessMsg(c, "邮箱变更成功")
}

// 支付渠道列表
func UserPayChannels(c *gin.Context) {
	uid := c.GetInt("uid")
	channels, err := userCenterService.GetPayChannels(uid)
	if err != nil {
		response.ServerError(c, "获取支付渠道失败")
		return
	}
	response.Success(c, channels)
}

// 创建充值订单
func UserCreatePay(c *gin.Context) {
	uid := c.GetInt("uid")
	var req model.PayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请输入充值金额")
		return
	}
	domain := c.Request.Host
	order, err := userCenterService.CreatePayOrder(uid, req.Money, req.Type, domain)
	if err != nil {
		response.BusinessError(c, 1003, err.Error())
		return
	}
	response.Success(c, order)
}

// 充值记录
func UserPayOrders(c *gin.Context) {
	uid := c.GetInt("uid")
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	orders, total, err := userCenterService.PayOrders(uid, page, limit)
	if err != nil {
		response.ServerError(c, "查询充值记录失败")
		return
	}
	response.Success(c, gin.H{
		"list":       orders,
		"pagination": gin.H{"page": page, "limit": limit, "total": total},
	})
}

// 余额流水
func UserMoneyLog(c *gin.Context) {
	uid := c.GetInt("uid")
	var req model.MoneyLogListRequest
	_ = c.ShouldBindQuery(&req)
	logs, total, err := userCenterService.MoneyLogList(uid, req)
	if err != nil {
		response.ServerError(c, "查询流水失败")
		return
	}
	response.Success(c, gin.H{
		"list":       logs,
		"pagination": gin.H{"page": req.Page, "limit": req.Limit, "total": total},
	})
}

// 工单列表
func UserTicketList(c *gin.Context) {
	uid := c.GetInt("uid")
	grade := c.GetString("grade")
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	tickets, total, err := userCenterService.TicketList(uid, grade, page, limit)
	if err != nil {
		response.ServerError(c, "查询工单失败")
		return
	}
	response.Success(c, gin.H{
		"list":       tickets,
		"pagination": gin.H{"page": page, "limit": limit, "total": total},
	})
}

// 创建工单
func UserTicketCreate(c *gin.Context) {
	uid := c.GetInt("uid")
	var req model.TicketCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写工单内容")
		return
	}

	// 检查分类工单开关
	if req.OID > 0 {
		_, ticketSwitch, _, _, _, _, _ := adminService.CategorySwitchesByOID(req.OID)
		if ticketSwitch == 0 {
			response.BadRequest(c, "该分类未开启工单功能")
			return
		}
	}

	id, err := userCenterService.TicketCreate(uid, req)
	if err != nil {
		response.ServerError(c, "创建工单失败")
		return
	}
	// 推送通知给管理员（uid=1）
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

// 回复工单（管理员）
func UserTicketReply(c *gin.Context) {
	uid := c.GetInt("uid")
	grade := c.GetString("grade")
	var req model.TicketReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写回复内容")
		return
	}
	_ = uid
	if err := userCenterService.TicketReply(uid, grade, req); err != nil {
		response.BusinessError(c, 1004, err.Error())
		return
	}
	response.SuccessMsg(c, "回复成功")
}

// ===== 课程收藏 (按 PHP favorites case) =====

func UserGetFavorites(c *gin.Context) {
	uid := c.GetInt("uid")
	favorites, err := userCenterService.GetFavorites(uid)
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
	if err := userCenterService.AddFavorite(uid, body.CID); err != nil {
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
	if err := userCenterService.RemoveFavorite(uid, body.CID); err != nil {
		response.ServerError(c, "移除收藏失败")
		return
	}
	response.SuccessMsg(c, "移除成功")
}

// ===== 支付状态检测 (按 PHP check_pay_status case) =====

func UserCheckPayStatus(c *gin.Context) {
	uid := c.GetInt("uid")
	var body struct {
		OutTradeNo string `json:"out_trade_no"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	paid, msg, err := userCenterService.CheckPayStatus(uid, body.OutTradeNo)
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

// ===== 设置邀请码 =====

func UserSetInviteCode(c *gin.Context) {
	uid := c.GetInt("uid")
	var body struct {
		YQM string `json:"yqm"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.YQM == "" {
		response.BadRequest(c, "请输入邀请码")
		return
	}
	if err := userCenterService.SetInviteCode(uid, body.YQM); err != nil {
		response.BusinessError(c, 1012, err.Error())
		return
	}
	response.SuccessMsg(c, "邀请码设置成功")
}

// ===== 等级列表（用户可见） =====

func UserGradeList(c *gin.Context) {
	uid := c.GetInt("uid")
	grade := c.GetString("grade")

	list, err := adminService.GradeList()
	if err != nil {
		response.ServerError(c, "查询等级列表失败")
		return
	}

	// 管理员看全部，普通用户只看 rate >= 自己的 addprice
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

// ===== 设置自己的等级（仅管理员） =====

func UserSetMyGrade(c *gin.Context) {
	uid := c.GetInt("uid")
	grade := c.GetString("grade")
	if grade != "2" && grade != "3" {
		response.BadRequest(c, "仅管理员可自行设置等级")
		return
	}
	var body struct {
		AddPrice float64 `json:"addprice"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.AddPrice < 0.01 {
		response.BadRequest(c, "请选择有效的等级")
		return
	}
	_, err := database.DB.Exec("UPDATE qingka_wangke_user SET addprice = ? WHERE uid = ?", body.AddPrice, uid)
	if err != nil {
		response.ServerError(c, "设置失败")
		return
	}
	response.SuccessMsg(c, "等级已更新")
}

// ===== 邀请费率 (按 PHP set_invite_rate case) =====

func UserSetInviteRate(c *gin.Context) {
	uid := c.GetInt("uid")
	var body struct {
		YQPrice float64 `json:"yqprice"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	// 获取用户当前 addprice
	var addprice float64
	database.DB.QueryRow("SELECT COALESCE(addprice,1) FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice)

	if err := userCenterService.SetInviteRate(uid, body.YQPrice, addprice); err != nil {
		response.BusinessError(c, 1010, err.Error())
		return
	}
	response.SuccessMsg(c, "设置成功")
}

// ===== API密钥管理 (按 PHP change_secret_key case) =====

func UserChangeSecretKey(c *gin.Context) {
	uid := c.GetInt("uid")
	var body struct {
		Type int `json:"type"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	// 获取用户余额
	var money float64
	database.DB.QueryRow("SELECT COALESCE(money,0) FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&money)

	key, err := userCenterService.ChangeSecretKey(uid, body.Type, money)
	if err != nil {
		response.BusinessError(c, 1011, err.Error())
		return
	}
	response.Success(c, gin.H{"key": key})
}

// ===== 推送Token设置 (按 PHP set_push_token case) =====

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

// ===== 操作日志 (按 PHP log_list case) =====

func UserLogList(c *gin.Context) {
	uid := c.GetInt("uid")
	grade := c.GetString("grade")
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	logType := c.Query("type")
	keywords := c.Query("keywords")

	list, total, err := userCenterService.LogList(uid, grade, page, limit, logType, keywords)
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

// 公告列表（用户端）
func AnnouncementListPublic(c *gin.Context) {
	uid := c.GetInt("uid")
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	list, total, err := adminService.AnnouncementListPublic(uid, page, limit)
	if err != nil {
		response.ServerError(c, "查询公告失败")
		return
	}
	response.Success(c, gin.H{
		"list":  list,
		"total": total,
	})
}

// 关闭工单
func UserTicketClose(c *gin.Context) {
	uid := c.GetInt("uid")
	grade := c.GetString("grade")
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		response.BadRequest(c, "无效的工单ID")
		return
	}
	if err := userCenterService.TicketClose(uid, grade, id); err != nil {
		response.ServerError(c, "关闭工单失败")
		return
	}
	response.SuccessMsg(c, "工单已关闭")
}
