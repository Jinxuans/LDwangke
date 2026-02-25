package handler

import (
	"strconv"

	"go-api/internal/response"
	"go-api/internal/service"

	"github.com/gin-gonic/gin"
)

var checkinService = service.NewCheckinService()

func UserCheckin(c *gin.Context) {
	uid := c.GetInt("uid")
	username := c.GetString("username")
	if uid <= 0 {
		response.Unauthorized(c, "请先登录")
		return
	}
	reward, err := checkinService.DoCheckin(uid, username)
	if err != nil {
		response.BusinessError(c, 1, err.Error())
		return
	}
	response.Success(c, gin.H{"reward_money": reward})
}

func UserCheckinStatus(c *gin.Context) {
	uid := c.GetInt("uid")
	checked, reward := checkinService.GetCheckinStatus(uid)
	response.Success(c, gin.H{"checked_in": checked, "reward_money": reward})
}

func AdminCheckinStats(c *gin.Context) {
	date := c.Query("date")
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	list, total, stat, err := checkinService.AdminCheckinStats(date, page, limit)
	if err != nil {
		response.ServerError(c, "查询签到记录失败")
		return
	}
	response.Success(c, gin.H{
		"list":         list,
		"total":        total,
		"total_users":  stat.TotalUsers,
		"total_reward": stat.TotalReward,
	})
}
