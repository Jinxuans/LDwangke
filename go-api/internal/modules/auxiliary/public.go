package auxiliary

import (
	"strconv"

	"go-api/internal/model"
	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

func UserCardKeyUse(c *gin.Context) {
	uid := c.GetInt("uid")
	var req model.CardKeyUseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请输入卡密")
		return
	}
	money, err := cardKeyUse(uid, req.Content)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, gin.H{"money": money, "msg": "充值成功"})
}

func UserActivityList(c *gin.Context) {
	list, err := listPublicActivities()
	if err != nil {
		response.ServerError(c, "查询活动失败")
		return
	}
	response.Success(c, list)
}

func UserPledgeConfigList(c *gin.Context) {
	list, err := listPublicPledgeConfigs()
	if err != nil {
		response.ServerError(c, "查询质押配置失败")
		return
	}
	var active []model.PledgeConfig
	for _, p := range list {
		if p.Status == 1 {
			active = append(active, p)
		}
	}
	if active == nil {
		active = []model.PledgeConfig{}
	}
	response.Success(c, active)
}

func UserPledgeCreate(c *gin.Context) {
	uid := c.GetInt("uid")
	var body struct {
		ConfigID int `json:"config_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "请选择质押配置")
		return
	}
	if err := createPledge(uid, body.ConfigID); err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.SuccessMsg(c, "质押成功")
}

func UserPledgeCancel(c *gin.Context) {
	uid := c.GetInt("uid")
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		response.BadRequest(c, "无效的质押记录ID")
		return
	}
	if err := cancelPledge(uid, id); err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.SuccessMsg(c, "取消质押成功")
}

func UserPledgeList(c *gin.Context) {
	uid := c.GetInt("uid")
	list, err := listActivePledges(uid)
	if err != nil {
		response.ServerError(c, "查询质押记录失败")
		return
	}
	response.Success(c, list)
}

func CheckOrderPublic(c *gin.Context) {
	var req model.CheckOrderRequest
	_ = c.ShouldBindQuery(&req)
	if req.User == "" && req.OID == "" {
		_ = c.ShouldBindJSON(&req)
	}
	list, err := checkOrder(req)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, gin.H{"list": list, "total": len(list)})
}
