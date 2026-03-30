package admin

import (
	"strconv"

	"go-api/internal/model"
	tenantmodule "go-api/internal/modules/tenant"
	usermodule "go-api/internal/modules/user"
	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

func registerWithdrawRoutes(admin *gin.RouterGroup) {
	admin.GET("/withdraw/requests", AdminWithdrawRequests)
	admin.POST("/withdraw/:id/review", AdminWithdrawReview)
	admin.GET("/mall-cuser-withdraw/requests", AdminMallCUserWithdrawRequests)
	admin.POST("/mall-cuser-withdraw/:id/review", AdminMallCUserWithdrawReview)
}

func AdminWithdrawRequests(c *gin.Context) {
	var req model.AdminWithdrawListRequest
	_ = c.ShouldBindQuery(&req)
	list, total, err := usermodule.AdminWithdrawRequests(req)
	if err != nil {
		response.ServerErrorf(c, err, "查询提现申请失败")
		return
	}
	response.Success(c, gin.H{
		"list":       list,
		"pagination": gin.H{"page": req.Page, "limit": req.Limit, "total": total},
	})
}

func AdminWithdrawReview(c *gin.Context) {
	adminUID := c.GetInt("uid")
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		response.BadRequest(c, "提现申请ID无效")
		return
	}
	var req model.WithdrawReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "审核参数错误")
		return
	}
	if err := usermodule.AdminReviewWithdrawRequest(adminUID, id, req.Status, req.Remark); err != nil {
		response.BusinessError(c, 1003, err.Error())
		return
	}
	response.SuccessMsg(c, "审核完成")
}

func AdminMallCUserWithdrawRequests(c *gin.Context) {
	var req model.AdminCUserWithdrawListRequest
	_ = c.ShouldBindQuery(&req)
	list, total, err := tenantmodule.AdminCUserWithdrawRequests(req)
	if err != nil {
		response.ServerErrorf(c, err, "查询会员提现申请失败")
		return
	}
	response.Success(c, gin.H{
		"list":       list,
		"pagination": gin.H{"page": req.Page, "limit": req.Limit, "total": total},
	})
}

func AdminMallCUserWithdrawReview(c *gin.Context) {
	adminUID := c.GetInt("uid")
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		response.BadRequest(c, "提现申请ID无效")
		return
	}
	var req model.WithdrawReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "审核参数错误")
		return
	}
	if err := tenantmodule.AdminReviewCUserWithdrawRequest(adminUID, id, req.Status, req.Remark); err != nil {
		response.BusinessError(c, 1003, err.Error())
		return
	}
	response.SuccessMsg(c, "审核完成")
}
