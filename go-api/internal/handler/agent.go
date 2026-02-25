package handler

import (
	"fmt"

	"go-api/internal/database"
	"go-api/internal/response"
	"go-api/internal/service"

	"github.com/gin-gonic/gin"
)

var agentService = service.NewAgentService()

// 获取操作者信息（uid, grade, money, addprice）
func getOperatorInfo(c *gin.Context) (int, string, float64, float64) {
	uid := c.GetInt("uid")
	grade := c.GetString("grade")
	var money, addprice float64
	database.DB.QueryRow("SELECT COALESCE(money,0), COALESCE(addprice,1) FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&money, &addprice)
	return uid, grade, money, addprice
}

// ===== 代理列表 =====

func AgentList(c *gin.Context) {
	uid, grade, _, _ := getOperatorInfo(c)

	var body struct {
		Page     int    `json:"page" form:"page"`
		Limit    int    `json:"limit" form:"limit"`
		Type     string `json:"type" form:"type"`
		Keywords string `json:"keywords" form:"keywords"`
	}
	c.ShouldBind(&body)

	list, total, err := agentService.AgentList(uid, grade, body.Page, body.Limit, body.Type, body.Keywords)
	if err != nil {
		response.ServerError(c, "查询失败")
		return
	}

	lastPage := 1
	if body.Limit > 0 && total > 0 {
		lastPage = int((total + int64(body.Limit) - 1) / int64(body.Limit))
	}

	response.Success(c, gin.H{
		"list": list,
		"pagination": gin.H{
			"current_page": body.Page,
			"last_page":    lastPage,
			"total":        total,
			"limit":        body.Limit,
		},
	})
}

// ===== 添加代理 =====

func AgentCreate(c *gin.Context) {
	uid, grade, money, addprice := getOperatorInfo(c)

	var req service.AgentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	msg, err := agentService.AgentCreate(uid, grade, money, addprice, req)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

// ===== 给下级充值 =====

func AgentRecharge(c *gin.Context) {
	uid, grade, money, addprice := getOperatorInfo(c)

	var body struct {
		UID   int     `json:"uid"`
		Money float64 `json:"money"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	err := agentService.AgentRecharge(uid, grade, money, addprice, body.UID, body.Money)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.SuccessMsg(c, "充值成功")
}

// ===== 扣除下级余额 =====

func AgentDeduct(c *gin.Context) {
	uid, grade, _, _ := getOperatorInfo(c)

	var body struct {
		UID   int     `json:"uid"`
		Money float64 `json:"money"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	err := agentService.AgentDeduct(uid, grade, body.UID, body.Money)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.SuccessMsg(c, "扣除成功")
}

// ===== 修改下级等级 =====

func AgentChangeGrade(c *gin.Context) {
	uid, grade, money, addprice := getOperatorInfo(c)

	var req service.AgentChangeGradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	msg, err := agentService.AgentChangeGrade(uid, grade, money, addprice, req)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

// ===== 封禁/解封 =====

func AgentChangeStatus(c *gin.Context) {
	uid := c.GetInt("uid")

	var body struct {
		UID    int `json:"uid"`
		Active int `json:"active"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	err := agentService.AgentChangeStatus(uid, body.UID, body.Active)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.SuccessMsg(c, "操作成功")
}

// ===== 重置下级密码 =====

func AgentResetPassword(c *gin.Context) {
	uid := c.GetInt("uid")

	var body struct {
		UID int `json:"uid"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	msg, err := agentService.AgentResetPassword(uid, body.UID)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

// ===== 给下级开通密钥 =====

func AgentOpenSecretKey(c *gin.Context) {
	uid, _, money, _ := getOperatorInfo(c)

	var body struct {
		UID int `json:"uid"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	err := agentService.AgentOpenSecretKey(uid, money, body.UID)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.SuccessMsg(c, "开通成功")
}

// ===== 上级迁移 =====

func AgentMigrateSuperior(c *gin.Context) {
	uid := c.GetInt("uid")

	var body struct {
		UID int    `json:"uid"`
		YQM string `json:"yqm"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.YQM == "" {
		response.BadRequest(c, "所有项目不能为空")
		return
	}

	err := agentService.AgentMigrateSuperior(uid, body.UID, body.YQM)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.SuccessMsg(c, fmt.Sprintf("迁移成功,您已迁移至[UID%d]的名下", body.UID))
}

// ===== 给下级设置邀请码 =====

func AgentSetInviteCode(c *gin.Context) {
	uid := c.GetInt("uid")

	var body struct {
		UID int    `json:"uid"`
		YQM string `json:"yqm"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.YQM == "" {
		response.BadRequest(c, "请输入邀请码")
		return
	}

	err := agentService.AgentSetInviteCode(uid, body.UID, body.YQM)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.SuccessMsg(c, "设置成功")
}
