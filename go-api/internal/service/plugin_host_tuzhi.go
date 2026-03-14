package service

import (
	"net/http"
	"time"
)

type TuZhiConfig struct {
	Username string `json:"daka_api_username"`
	Password string `json:"daka_api_password"`
}

type TuZhiGoodsOverride struct {
	GoodsID int     `json:"goods_id"`
	Price   float64 `json:"price"`
	Enabled int     `json:"enabled"`
}

type TuZhiService struct {
	client  *http.Client
	baseURL string
}

var tuzhiService = &TuZhiService{
	client:  &http.Client{Timeout: 30 * time.Second},
	baseURL: "http://apis.bbwace.icu",
}

func TuZhi() *TuZhiService {
	return tuzhiService
}
