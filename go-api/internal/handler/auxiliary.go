package handler

import (
	"strconv"

	"go-api/internal/model"
	"go-api/internal/response"
	"go-api/internal/service"

	"github.com/gin-gonic/gin"
)

var auxService = service.NewAuxiliaryService()

// ===== 卡密系统 =====

func AdminCardKeyList(c *gin.Context) {
	var req model.CardKeyListRequest
	_ = c.ShouldBindQuery(&req)
	list, total, err := auxService.CardKeyList(req)
	if err != nil {
		response.ServerError(c, "查询卡密失败")
		return
	}
	response.Success(c, gin.H{
		"list":       list,
		"pagination": gin.H{"page": req.Page, "limit": req.Limit, "total": total},
	})
}

func AdminCardKeyGenerate(c *gin.Context) {
	var req model.CardKeyGenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写金额和数量")
		return
	}
	codes, err := auxService.CardKeyGenerate(req.Money, req.Count)
	if err != nil {
		response.ServerError(c, "生成卡密失败")
		return
	}
	response.Success(c, gin.H{"codes": codes, "count": len(codes)})
}

func AdminCardKeyDelete(c *gin.Context) {
	var body struct {
		IDs []int `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || len(body.IDs) == 0 {
		response.BadRequest(c, "请选择要删除的卡密")
		return
	}
	deleted, err := auxService.CardKeyDelete(body.IDs)
	if err != nil {
		response.ServerError(c, "删除失败")
		return
	}
	response.Success(c, gin.H{"deleted": deleted})
}

func UserCardKeyUse(c *gin.Context) {
	uid := c.GetInt("uid")
	var req model.CardKeyUseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请输入卡密")
		return
	}
	money, err := auxService.CardKeyUse(uid, req.Content)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, gin.H{"money": money, "msg": "充值成功"})
}

// ===== 活动系统 =====

func AdminActivityList(c *gin.Context) {
	var req model.ActivityListRequest
	_ = c.ShouldBindQuery(&req)
	list, total, err := auxService.ActivityList(req)
	if err != nil {
		response.ServerError(c, "查询活动失败")
		return
	}
	response.Success(c, gin.H{
		"list":       list,
		"pagination": gin.H{"page": req.Page, "limit": req.Limit, "total": total},
	})
}

func AdminActivitySave(c *gin.Context) {
	var req model.ActivitySaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写完整的活动信息")
		return
	}
	if err := auxService.ActivitySave(req); err != nil {
		response.ServerError(c, "保存活动失败")
		return
	}
	response.SuccessMsg(c, "保存成功")
}

func AdminActivityDelete(c *gin.Context) {
	hid, _ := strconv.Atoi(c.Param("hid"))
	if hid <= 0 {
		response.BadRequest(c, "无效的活动ID")
		return
	}
	if err := auxService.ActivityDelete(hid); err != nil {
		response.ServerError(c, "删除失败")
		return
	}
	response.SuccessMsg(c, "删除成功")
}

func UserActivityList(c *gin.Context) {
	list, err := auxService.ActivityListPublic()
	if err != nil {
		response.ServerError(c, "查询活动失败")
		return
	}
	response.Success(c, list)
}

// ===== 质押系统 =====

func AdminPledgeConfigList(c *gin.Context) {
	list, err := auxService.PledgeConfigList()
	if err != nil {
		response.ServerError(c, "查询质押配置失败")
		return
	}
	response.Success(c, list)
}

func AdminPledgeConfigSave(c *gin.Context) {
	var req model.PledgeConfigSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写完整的质押配置")
		return
	}
	if err := auxService.PledgeConfigSave(req); err != nil {
		response.ServerError(c, "保存失败")
		return
	}
	response.SuccessMsg(c, "保存成功")
}

func AdminPledgeConfigDelete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		response.BadRequest(c, "无效的配置ID")
		return
	}
	if err := auxService.PledgeConfigDelete(id); err != nil {
		response.ServerError(c, "删除失败")
		return
	}
	response.SuccessMsg(c, "删除成功")
}

func AdminPledgeConfigToggle(c *gin.Context) {
	var body struct {
		ID     int `json:"id" binding:"required"`
		Status int `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := auxService.PledgeConfigToggle(body.ID, body.Status); err != nil {
		response.ServerError(c, "更新失败")
		return
	}
	response.SuccessMsg(c, "更新成功")
}

func AdminPledgeRecordList(c *gin.Context) {
	var req model.PledgeListRequest
	_ = c.ShouldBindQuery(&req)
	list, total, err := auxService.PledgeRecordList(req)
	if err != nil {
		response.ServerError(c, "查询质押记录失败")
		return
	}
	response.Success(c, gin.H{
		"list":       list,
		"pagination": gin.H{"page": req.Page, "limit": req.Limit, "total": total},
	})
}

func UserPledgeConfigList(c *gin.Context) {
	list, err := auxService.PledgeConfigList()
	if err != nil {
		response.ServerError(c, "查询质押配置失败")
		return
	}
	// 只返回生效的
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
	if err := auxService.PledgeCreate(uid, body.ConfigID); err != nil {
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
	if err := auxService.PledgeCancel(uid, id); err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.SuccessMsg(c, "取消质押成功")
}

func UserPledgeList(c *gin.Context) {
	uid := c.GetInt("uid")
	list, err := auxService.UserActivePledges(uid)
	if err != nil {
		response.ServerError(c, "查询质押记录失败")
		return
	}
	response.Success(c, list)
}

// ===== 外部查单（公开接口，无需认证） =====

func CheckOrderPublic(c *gin.Context) {
	var req model.CheckOrderRequest
	_ = c.ShouldBindQuery(&req)
	// 也支持 POST
	if req.User == "" && req.OID == "" {
		_ = c.ShouldBindJSON(&req)
	}

	list, err := auxService.CheckOrder(req)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, gin.H{"list": list, "total": len(list)})
}
