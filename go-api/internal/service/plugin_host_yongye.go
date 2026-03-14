package service

import (
	"net/http"
	"time"
)

type YongyeConfig struct {
	ApiURL  string  `json:"api_url"`
	Token   string  `json:"token"`
	Dj      float64 `json:"dj"`
	Zs      float64 `json:"zs"`
	Beis    float64 `json:"beis"`
	Xzdj    float64 `json:"xzdj"`
	Xzmo    float64 `json:"xzmo"`
	Tk      float64 `json:"tk"`
	Content string  `json:"content"`
	Tcgg    string  `json:"tcgg"`
}

type YongyeOrder struct {
	ID         int     `json:"id"`
	Pol        int     `json:"pol"`
	UID        int     `json:"uid"`
	User       string  `json:"user"`
	Pass       string  `json:"pass"`
	School     string  `json:"school"`
	Type       int     `json:"type"`
	Zkm        float64 `json:"zkm"`
	KsH        int     `json:"ks_h"`
	KsM        int     `json:"ks_m"`
	JsH        int     `json:"js_h"`
	JsM        int     `json:"js_m"`
	Weeks      string  `json:"weeks"`
	DockStatus int     `json:"dockstatus"`
	Yfees      float64 `json:"yfees"`
	Fees       float64 `json:"fees"`
	YID        string  `json:"yid"`
	Yaddtime   string  `json:"yaddtime"`
	Addtime    string  `json:"addtime"`
	Tktext     string  `json:"tktext"`
}

type YongyeStudent struct {
	ID       int     `json:"id"`
	UID      int     `json:"uid"`
	User     string  `json:"user"`
	Pass     string  `json:"pass"`
	Type     int     `json:"type"`
	Zkm      float64 `json:"zkm"`
	Weeks    string  `json:"weeks"`
	Status   int     `json:"status"`
	Tdkm     float64 `json:"tdkm"`
	Tdmoney  float64 `json:"tdmoney"`
	Stulog   string  `json:"stulog"`
	LastTime string  `json:"last_time"`
}

type YongyeSchool struct {
	Name   string  `json:"name"`
	Cpmuch float64 `json:"cpmuch"`
	Zcmuch float64 `json:"zcmuch"`
}

type YongyeService struct {
	client *http.Client
}

var yongyeService = &YongyeService{client: &http.Client{Timeout: 15 * time.Second}}

func Yongye() *YongyeService {
	return yongyeService
}
