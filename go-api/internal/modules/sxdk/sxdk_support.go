package sxdk

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-api/internal/database"
)

func sxdkLog(uid int, logType, text string, money float64) {
	now := time.Now().Format("2006-01-02 15:04:05")
	var newBalance float64
	database.DB.QueryRow("SELECT money FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&newBalance)
	moneyStr := fmt.Sprintf("%.2f", money)
	if money > 0 {
		moneyStr = fmt.Sprintf("+%.2f", money)
	}
	database.DB.Exec(
		"INSERT INTO qingka_wangke_log (uid, type, text, money, smoney, ip, addtime) VALUES (?, ?, ?, ?, ?, '', ?)",
		uid, logType, text, moneyStr, fmt.Sprintf("%.2f", newBalance), now,
	)
}

func processWxpush(wxpush string) map[string]interface{} {
	if wxpush == "" {
		return map[string]interface{}{"wxpush": nil}
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(wxpush), &result); err != nil {
		return map[string]interface{}{"wxpush": wxpush}
	}
	return result
}

func getDayOfWeek(t time.Time) int {
	d := int(t.Weekday())
	if d == 0 {
		return 6
	}
	return d - 1
}

func timeCalcTrueday(now time.Time, endTimeStr string, checkWeekStr string) int {
	parts := strings.Split(checkWeekStr, ",")
	var checkWeek []int
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if v, err := strconv.Atoi(p); err == nil {
			checkWeek = append(checkWeek, v)
		}
	}
	sort.Ints(checkWeek)

	endTime, err := time.Parse("2006-01-02 15:04:05", endTimeStr+" 23:59:59")
	if err != nil {
		endTime, err = time.Parse("2006-01-02", endTimeStr)
		if err != nil {
			return 0
		}
		endTime = endTime.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	}

	if endTime.Before(now) {
		return 0
	}

	nowWeekDay := getDayOfWeek(now)
	daysToSunday := 6 - nowWeekDay
	weekEnd := time.Date(now.Year(), now.Month(), now.Day()+daysToSunday, 23, 59, 59, 0, now.Location())

	var nowWeekLast []int
	for _, d := range checkWeek {
		if d >= nowWeekDay {
			nowWeekLast = append(nowWeekLast, d)
		}
	}

	endWeekDay := getDayOfWeek(endTime)
	if endTime.Before(weekEnd) || endTime.Equal(weekEnd) {
		count := 0
		for _, d := range nowWeekLast {
			if d <= endWeekDay {
				count++
			}
		}
		return count
	}

	var endWeekLast []int
	for _, d := range checkWeek {
		if d <= endWeekDay {
			endWeekLast = append(endWeekLast, d)
		}
	}

	endWeekStart := time.Date(endTime.Year(), endTime.Month(), endTime.Day()-(endWeekDay+1), 23, 59, 59, 0, endTime.Location())
	intDuration := endWeekStart.Sub(weekEnd)
	fullWeeks := int(intDuration.Hours() / 24 / 7)

	return len(nowWeekLast) + fullWeeks*len(checkWeek) + len(endWeekLast)
}
