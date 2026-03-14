package agent

import (
	"fmt"

	"go-api/internal/database"
	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

func getOperatorInfo(c *gin.Context) (int, string, float64, float64) {
	uid := c.GetInt("uid")
	grade := c.GetString("grade")
	var money, addprice float64
	database.DB.QueryRow("SELECT COALESCE(money,0), COALESCE(addprice,1) FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&money, &addprice)
	return uid, grade, money, addprice
}

func AgentList(c *gin.Context) {
	uid, grade, _, _ := getOperatorInfo(c)

	var body struct {
		Page     int    `json:"page" form:"page"`
		Limit    int    `json:"limit" form:"limit"`
		Type     string `json:"type" form:"type"`
		Keywords string `json:"keywords" form:"keywords"`
	}
	c.ShouldBind(&body)

	list, total, err := agents.AgentList(uid, grade, body.Page, body.Limit, body.Type, body.Keywords)
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

func AgentCreate(c *gin.Context) {
	uid, grade, money, addprice := getOperatorInfo(c)

	var req AgentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	msg, err := agents.AgentCreate(uid, grade, money, addprice, req)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

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

	err := agents.AgentRecharge(uid, grade, money, addprice, body.UID, body.Money)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.SuccessMsg(c, "充值成功")
}

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

	err := agents.AgentDeduct(uid, grade, body.UID, body.Money)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.SuccessMsg(c, "扣除成功")
}

func AgentChangeGrade(c *gin.Context) {
	uid, grade, money, addprice := getOperatorInfo(c)

	var req AgentChangeGradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	msg, err := agents.AgentChangeGrade(uid, grade, money, addprice, req)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

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

	err := agents.AgentChangeStatus(uid, body.UID, body.Active)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.SuccessMsg(c, "操作成功")
}

func AgentResetPassword(c *gin.Context) {
	uid := c.GetInt("uid")

	var body struct {
		UID int `json:"uid"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	msg, err := agents.AgentResetPassword(uid, body.UID)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

func AgentOpenSecretKey(c *gin.Context) {
	uid, _, money, _ := getOperatorInfo(c)

	var body struct {
		UID int `json:"uid"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	err := agents.AgentOpenSecretKey(uid, money, body.UID)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.SuccessMsg(c, "开通成功")
}

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

	err := agents.AgentMigrateSuperior(uid, body.UID, body.YQM)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.SuccessMsg(c, fmt.Sprintf("迁移成功,您已迁移至[UID%d]的名下", body.UID))
}

func AgentCrossRechargeCheck(c *gin.Context) {
	uid := c.GetInt("uid")
	allowed := agents.CrossRechargeAllowed(uid)
	response.Success(c, gin.H{"allowed": allowed})
}

func AgentCrossRecharge(c *gin.Context) {
	uid := c.GetInt("uid")

	var body struct {
		UID   int     `json:"uid"`
		Money float64 `json:"money"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	err := agents.AgentCrossRecharge(uid, body.UID, body.Money)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.SuccessMsg(c, "跨户充值成功")
}

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

	err := agents.AgentSetInviteCode(uid, body.UID, body.YQM)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.SuccessMsg(c, "设置成功")
}
