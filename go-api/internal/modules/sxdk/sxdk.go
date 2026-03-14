package sxdk

import (
	"net/http"
	"time"
)

type SXDKConfig struct {
	BaseURL string `json:"base_url"`
	Token   string `json:"token"`
	Admin   string `json:"admin"`
}

type SXDKOrder struct {
	ID            int    `json:"id"`
	SxdkID        int    `json:"sxdkId"`
	UID           int    `json:"uid"`
	Platform      string `json:"platform"`
	Phone         string `json:"phone"`
	Password      string `json:"password"`
	Code          int    `json:"code"`
	Wxpush        string `json:"wxpush"`
	Name          string `json:"name"`
	Address       string `json:"address"`
	UpCheckTime   string `json:"up_check_time"`
	DownCheckTime string `json:"down_check_time"`
	CheckWeek     string `json:"check_week"`
	EndTime       string `json:"end_time"`
	DayPaper      int    `json:"day_paper"`
	WeekPaper     int    `json:"week_paper"`
	MonthPaper    int    `json:"month_paper"`
	CreateTime    string `json:"createTime"`
	UpdateTime    string `json:"updateTime"`
	WxpushURL     string `json:"wxpushUrl,omitempty"`
	RunType       int    `json:"runType,omitempty"`
}

type SXDKService struct {
	client *http.Client
}

var sxdkService = &SXDKService{
	client: &http.Client{Timeout: 30 * time.Second},
}

func SXDK() *SXDKService {
	return sxdkService
}
