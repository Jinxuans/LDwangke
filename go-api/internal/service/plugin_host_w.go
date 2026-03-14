package service

import (
	"net/http"
	"time"
)

type WApp struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Code     string  `json:"code"`
	OrgAppID string  `json:"org_app_id"`
	Status   int     `json:"status"`
	Desc     string  `json:"description"`
	Price    float64 `json:"price"`
	CacType  string  `json:"cac_type"`
	URL      string  `json:"url"`
	Key      string  `json:"key"`
	UID      string  `json:"uid"`
	Token    string  `json:"token"`
	Type     string  `json:"type"`
	Deleted  int     `json:"deleted"`
}

type WAppUser struct {
	AppID    int64   `json:"app_id"`
	OrgAppID string  `json:"org_app_id"`
	Code     string  `json:"code"`
	Name     string  `json:"name"`
	Desc     string  `json:"description"`
	CacType  string  `json:"cac_type"`
	Price    float64 `json:"price"`
}

type WOrder struct {
	ID         int64       `json:"id"`
	AggOrderID *string     `json:"agg_order_id"`
	UserID     int64       `json:"user_id"`
	School     string      `json:"school"`
	Account    string      `json:"account"`
	Password   string      `json:"password"`
	AppID      int64       `json:"app_id"`
	AppName    string      `json:"app_name"`
	Status     string      `json:"status"`
	Num        int         `json:"num"`
	Cost       float64     `json:"cost"`
	Pause      bool        `json:"pause"`
	SubOrder   interface{} `json:"sub_order"`
	Deleted    bool        `json:"deleted"`
	Created    string      `json:"created"`
	Updated    string      `json:"updated"`
}

type WService struct {
	client *http.Client
}

var wService = &WService{
	client: &http.Client{Timeout: 15 * time.Second},
}

func W() *WService {
	return wService
}
