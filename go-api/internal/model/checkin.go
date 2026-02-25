package model

type CheckinRecord struct {
	ID          int     `json:"id"`
	UID         int     `json:"uid"`
	Username    string  `json:"username"`
	RewardMoney float64 `json:"reward_money"`
	CheckinDate string  `json:"checkin_date"`
	AddTime     string  `json:"addtime"`
}

type CheckinDayStat struct {
	CheckinDate string  `json:"checkin_date"`
	TotalUsers  int     `json:"total_users"`
	TotalReward float64 `json:"total_reward"`
}
