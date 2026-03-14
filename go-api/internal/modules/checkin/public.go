package checkin

import (
	"fmt"
	"math/rand"
	"time"

	"go-api/internal/database"
	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

func getCheckinConfig() (enabled bool, orderRequired bool, minBalance float64, maxUsers int, minReward, maxReward float64) {
	enabled = false
	orderRequired = true
	minBalance = 10
	maxUsers = 10
	minReward = 0.1
	maxReward = 0.2

	cfg := map[string]string{}
	rows, err := database.DB.Query("SELECT `v`, `k` FROM qingka_wangke_config WHERE `v` IN ('checkin_enabled','checkin_order_required','checkin_min_balance','checkin_max_users','checkin_min_reward','checkin_max_reward')")
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var k, v string
		rows.Scan(&k, &v)
		cfg[k] = v
	}

	if cfg["checkin_enabled"] == "1" {
		enabled = true
	}
	if cfg["checkin_order_required"] == "0" {
		orderRequired = false
	}
	if v := cfg["checkin_min_balance"]; v != "" {
		fmt.Sscanf(v, "%f", &minBalance)
	}
	if v := cfg["checkin_max_users"]; v != "" {
		fmt.Sscanf(v, "%d", &maxUsers)
	}
	if v := cfg["checkin_min_reward"]; v != "" {
		fmt.Sscanf(v, "%f", &minReward)
	}
	if v := cfg["checkin_max_reward"]; v != "" {
		fmt.Sscanf(v, "%f", &maxReward)
	}
	return
}

func doCheckin(uid int, username string) (float64, error) {
	enabled, orderRequired, minBalance, maxUsers, minReward, maxReward := getCheckinConfig()

	if !enabled {
		return 0, fmt.Errorf("签到功能暂未开放")
	}

	today := time.Now().Format("2006-01-02")

	if orderRequired {
		var orderCnt int
		database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE uid = ?", uid).Scan(&orderCnt)
		if orderCnt == 0 {
			return 0, fmt.Errorf("需要有历史订单才能签到")
		}
	}

	var balance float64
	database.DB.QueryRow("SELECT COALESCE(money,0) FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&balance)
	if balance < minBalance {
		return 0, fmt.Errorf("余额不足%.0f元，无法签到", minBalance)
	}

	var todayUsers int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_checkin WHERE checkin_date = ?", today).Scan(&todayUsers)
	if todayUsers >= maxUsers {
		return 0, fmt.Errorf("今日签到名额已满")
	}

	var already int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_checkin WHERE uid = ? AND checkin_date = ?", uid, today).Scan(&already)
	if already > 0 {
		return 0, fmt.Errorf("今日已签到")
	}

	reward := minReward + rand.Float64()*(maxReward-minReward)
	reward = float64(int(reward*100)) / 100

	now := time.Now().Format("2006-01-02 15:04:05")

	_, err := database.DB.Exec("UPDATE qingka_wangke_user SET money = money + ? WHERE uid = ?", reward, uid)
	if err != nil {
		return 0, fmt.Errorf("签到失败，请稍后重试")
	}

	database.DB.Exec(
		"INSERT INTO qingka_wangke_checkin (uid, username, reward_money, checkin_date, addtime) VALUES (?, ?, ?, ?, ?)",
		uid, username, reward, today, now,
	)

	var newBalance float64
	database.DB.QueryRow("SELECT COALESCE(money,0) FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&newBalance)
	database.DB.Exec(
		"INSERT INTO qingka_wangke_moneylog (uid, type, money, balance, remark, addtime) VALUES (?, '签到', ?, ?, ?, ?)",
		uid, reward, newBalance, fmt.Sprintf("签到奖励 +%.2f元", reward), now,
	)

	return reward, nil
}

func getCheckinStatus(uid int) (bool, float64) {
	today := time.Now().Format("2006-01-02")
	var reward float64
	err := database.DB.QueryRow("SELECT COALESCE(reward_money,0) FROM qingka_wangke_checkin WHERE uid = ? AND checkin_date = ?", uid, today).Scan(&reward)
	if err != nil {
		return false, 0
	}
	return true, reward
}

func UserCheckin(c *gin.Context) {
	uid := c.GetInt("uid")
	username := c.GetString("username")
	if uid <= 0 {
		response.Unauthorized(c, "请先登录")
		return
	}
	reward, err := doCheckin(uid, username)
	if err != nil {
		response.BusinessError(c, 1, err.Error())
		return
	}
	response.Success(c, gin.H{"reward_money": reward})
}

func UserCheckinStatus(c *gin.Context) {
	uid := c.GetInt("uid")
	checked, reward := getCheckinStatus(uid)
	response.Success(c, gin.H{"checked_in": checked, "reward_money": reward})
}
