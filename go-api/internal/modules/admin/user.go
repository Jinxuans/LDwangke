package admin

import (
	"log"
	"strconv"

	"go-api/internal/model"
	authmodule "go-api/internal/modules/auth"
	classmodule "go-api/internal/modules/class"
	usermodule "go-api/internal/modules/user"
	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

func registerUserRoutes(admin *gin.RouterGroup) {
	admin.POST("/impersonate", authmodule.Impersonate)
	admin.GET("/users", AdminUserList)
	admin.POST("/user/reset-pass", AdminUserResetPass)
	admin.POST("/user/balance", AdminUserSetBalance)
	admin.POST("/user/grade", AdminUserSetGrade)
	admin.GET("/grades", AdminGradeList)
	admin.POST("/grade/save", AdminGradeSave)
	admin.DELETE("/grade/:id", AdminGradeDelete)
}

func AdminUserList(c *gin.Context) {
	var req model.UserListRequest
	_ = c.ShouldBindQuery(&req)

	list, total, err := usermodule.User().UserList(req)
	if err != nil {
		log.Printf("[AdminUserList] 查询失败: %v", err)
		response.ServerError(c, "查询用户失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":       list,
		"pagination": gin.H{"page": req.Page, "limit": req.Limit, "total": total},
	})
}

func AdminUserResetPass(c *gin.Context) {
	var body struct {
		UID     int    `json:"uid" binding:"required"`
		NewPass string `json:"new_pass"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := usermodule.User().ResetPassword(body.UID, body.NewPass); err != nil {
		response.ServerError(c, "重置密码失败")
		return
	}

	response.SuccessMsg(c, "密码已重置")
}

func AdminUserSetBalance(c *gin.Context) {
	var body struct {
		UID     int     `json:"uid" binding:"required"`
		Balance float64 `json:"balance"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := usermodule.User().SetBalance(body.UID, body.Balance); err != nil {
		response.ServerError(c, "设置余额失败")
		return
	}

	response.SuccessMsg(c, "余额已更新")
}

func AdminUserSetGrade(c *gin.Context) {
	var body struct {
		UID      int     `json:"uid" binding:"required"`
		AddPrice float64 `json:"addprice"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if body.AddPrice < 0.01 {
		response.BadRequest(c, "费率不能小于0.01")
		return
	}

	if err := usermodule.User().SetGrade(body.UID, body.AddPrice); err != nil {
		response.ServerError(c, "设置等级失败")
		return
	}

	response.SuccessMsg(c, "等级已更新")
}

func AdminGradeList(c *gin.Context) {
	list, err := classmodule.Classes().GradeList()
	if err != nil {
		response.ServerError(c, "查询等级列表失败")
		return
	}
	response.Success(c, list)
}

func AdminGradeSave(c *gin.Context) {
	var req model.GradeSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := classmodule.Classes().GradeSave(req); err != nil {
		log.Printf("[AdminGradeSave] 保存失败: %v", err)
		response.ServerError(c, "保存失败: "+err.Error())
		return
	}
	response.SuccessMsg(c, "保存成功")
}

func AdminGradeDelete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	if err := classmodule.Classes().GradeDelete(id); err != nil {
		response.ServerError(c, "删除失败")
		return
	}
	response.SuccessMsg(c, "删除成功")
}
