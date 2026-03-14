package service

import (
	"net/http"
	"time"
)

type YFDKConfig struct {
	BaseURL string `json:"base_url"`
	Token   string `json:"token"`
}

type YFDKProject struct {
	ID         int     `json:"id"`
	CID        string  `json:"cid"`
	Name       string  `json:"name"`
	Content    string  `json:"content"`
	CostPrice  float64 `json:"cost_price"`
	SellPrice  float64 `json:"sell_price"`
	Enabled    int     `json:"enabled"`
	Sort       int     `json:"sort"`
	CreateTime string  `json:"create_time"`
	UpdateTime string  `json:"update_time"`
}

type YFDKOrder struct {
	ID           int     `json:"id"`
	UID          int     `json:"uid"`
	OID          string  `json:"oid"`
	CID          string  `json:"cid"`
	Username     string  `json:"username"`
	Password     string  `json:"password"`
	School       string  `json:"school"`
	Name         string  `json:"name"`
	Email        string  `json:"email"`
	Offer        string  `json:"offer"`
	Address      string  `json:"address"`
	Longitude    string  `json:"longitude"`
	Latitude     string  `json:"latitude"`
	Week         string  `json:"week"`
	Worktime     string  `json:"worktime"`
	Offwork      int     `json:"offwork"`
	Offtime      string  `json:"offtime"`
	Day          int     `json:"day"`
	DailyFee     float64 `json:"daily_fee"`
	TotalFee     float64 `json:"total_fee"`
	DayReport    int     `json:"day_report"`
	WeekReport   int     `json:"week_report"`
	WeekDate     int     `json:"week_date"`
	MonthReport  int     `json:"month_report"`
	MonthDate    int     `json:"month_date"`
	SkipHolidays int     `json:"skip_holidays"`
	Image        int     `json:"image"`
	Status       int     `json:"status"`
	Mark         string  `json:"mark"`
	Endtime      string  `json:"endtime"`
	CreateTime   string  `json:"create_time"`
	UpdateTime   string  `json:"update_time"`
}

type YFDKService struct {
	client *http.Client
}

var yfdkService = &YFDKService{
	client: &http.Client{Timeout: 25 * time.Second},
}

func YFDK() *YFDKService {
	return yfdkService
}
