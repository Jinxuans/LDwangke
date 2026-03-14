package ydsj

import (
	_ "embed"
	"net/http"
	"time"
)

//go:embed ydsj_schools.json
var ydsjSchoolsJSON []byte

type YDSJConfig struct {
	BaseURL          string  `json:"base_url"`           // 上游API地址 如 http://103.236.75.71:7799/LearnExp
	UID              string  `json:"uid"`                // 对接站用户UID
	Key              string  `json:"key"`                // 对接站密钥
	Token            string  `json:"token"`              // (旧字段，保留兼容)
	PriceMultiple    float64 `json:"price_multiple"`     // 运动世界价格倍率
	XbdMorningPrice  float64 `json:"xbd_morning_price"`  // 小步点晨跑价格
	XbdExercisePrice float64 `json:"xbd_exercise_price"` // 小步点课外跑价格
	RealCostMultiple float64 `json:"real_cost_multiple"` // 实际扣费倍率
}

type YDSJOrder struct {
	ID          int    `json:"id"`
	YID         string `json:"yid"`
	UID         int    `json:"uid"`
	School      string `json:"school"`
	User        string `json:"user"`
	Pass        string `json:"pass"`
	Distance    string `json:"distance"`
	IsRun       int    `json:"is_run"`
	RunType     int    `json:"run_type"`
	StartHour   string `json:"start_hour"`
	StartMinute string `json:"start_minute"`
	EndHour     string `json:"end_hour"`
	EndMinute   string `json:"end_minute"`
	RunWeek     string `json:"run_week"`
	Status      int    `json:"status"`
	Remarks     string `json:"remarks"`
	Info        string `json:"info"`
	TmpInfo     string `json:"tmp_info"`
	Fees        string `json:"fees"`
	RealFees    string `json:"real_fees"`
	RefundMoney string `json:"refund_money"`
	Addtime     string `json:"addtime"`
}

type YDSJService struct {
	client *http.Client
}

var ydsjService = &YDSJService{client: &http.Client{Timeout: 30 * time.Second}}

func YDSJ() *YDSJService {
	return ydsjService
}
